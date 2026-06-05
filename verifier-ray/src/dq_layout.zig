const std = @import("std");

const field = @import("field/koalabear.zig");
const ext = @import("field/koalabear_ext.zig");

pub const DecodeError = std.mem.Allocator.Error || error{
    UnexpectedEnd,
    TrailingBytes,
    InvalidMagic,
    NonCanonicalField,
};

pub const DQLayout = struct {
    sizes: []const u32 = &.{},
    eval_points: []const []const ext.Ext = &.{},
    column_names: []const []const []const []const u8 = &.{},
    column_keys: []const []const []const []const u8 = &.{},
    air_chunks: []const []const []const u8 = &.{},
};

pub const OwnedDQLayout = struct {
    value: DQLayout,
    allocator: std.mem.Allocator,

    pub fn deinit(self: *OwnedDQLayout) void {
        self.allocator.free(self.value.sizes);

        for (self.value.eval_points) |points| {
            self.allocator.free(points);
        }
        self.allocator.free(self.value.eval_points);

        freeStringGrid(self.allocator, self.value.column_names);
        freeStringGrid(self.allocator, self.value.column_keys);

        for (self.value.air_chunks) |chunks| {
            self.allocator.free(chunks);
        }
        self.allocator.free(self.value.air_chunks);

        self.* = undefined;
    }
};

/// Decodes `DQL1` bytes. Names and keys borrow from `encoded`, so callers must
/// keep the byte buffer alive for at least as long as the returned layout.
pub fn decode(allocator: std.mem.Allocator, encoded: []const u8) DecodeError!OwnedDQLayout {
    var reader = Reader.init(encoded);
    try reader.expectMagic("DQL1");

    const level_count = try reader.readCount();
    var owned = OwnedDQLayout{
        .allocator = allocator,
        .value = .{
            .sizes = &.{},
            .eval_points = &.{},
            .column_names = &.{},
            .column_keys = &.{},
            .air_chunks = &.{},
        },
    };
    errdefer owned.deinit();

    const sizes = try allocator.alloc(u32, level_count);
    owned.value.sizes = sizes;
    const eval_point_rows = try allocator.alloc([]const ext.Ext, level_count);
    owned.value.eval_points = eval_point_rows;
    const column_name_rows = try allocator.alloc([]const []const []const u8, level_count);
    owned.value.column_names = column_name_rows;
    const column_key_rows = try allocator.alloc([]const []const []const u8, level_count);
    owned.value.column_keys = column_key_rows;
    const air_chunk_rows = try allocator.alloc([]const []const u8, level_count);
    owned.value.air_chunks = air_chunk_rows;

    for (0..level_count) |level_index| {
        eval_point_rows[level_index] = &.{};
        column_name_rows[level_index] = &.{};
        column_key_rows[level_index] = &.{};
        air_chunk_rows[level_index] = &.{};
    }

    for (0..level_count) |level_index| {
        sizes[level_index] = try reader.readU32();

        const eval_count = try reader.readCount();
        const eval_points = try allocator.alloc(ext.Ext, eval_count);
        eval_point_rows[level_index] = eval_points;
        for (eval_points) |*point| point.* = try reader.readExt();

        const level_names = try allocator.alloc([]const []const u8, eval_count);
        column_name_rows[level_index] = level_names;
        const level_keys = try allocator.alloc([]const []const u8, eval_count);
        column_key_rows[level_index] = level_keys;
        for (0..eval_count) |eval_index| {
            level_names[eval_index] = &.{};
            level_keys[eval_index] = &.{};
        }

        for (0..eval_count) |eval_index| {
            const term_count = try reader.readCount();
            const names = try allocator.alloc([]const u8, term_count);
            level_names[eval_index] = names;
            const keys = try allocator.alloc([]const u8, term_count);
            level_keys[eval_index] = keys;
            for (0..term_count) |term_index| {
                names[term_index] = try reader.readString();
                keys[term_index] = try reader.readString();
            }
        }

        const air_count = try reader.readCount();
        const chunks = try allocator.alloc([]const u8, air_count);
        air_chunk_rows[level_index] = chunks;
        for (0..air_count) |chunk_index| {
            chunks[chunk_index] = try reader.readString();
        }
    }

    if (!reader.isDone()) return DecodeError.TrailingBytes;
    return owned;
}

fn freeStringGrid(allocator: std.mem.Allocator, grid: []const []const []const []const u8) void {
    for (grid) |level| {
        for (level) |row| {
            allocator.free(row);
        }
        allocator.free(level);
    }
    allocator.free(grid);
}

const Reader = struct {
    bytes: []const u8,
    cursor: usize = 0,

    fn init(bytes: []const u8) Reader {
        return .{ .bytes = bytes };
    }

    fn isDone(self: Reader) bool {
        return self.cursor == self.bytes.len;
    }

    fn remaining(self: Reader) usize {
        return self.bytes.len - self.cursor;
    }

    fn expectMagic(self: *Reader, expected: *const [4]u8) DecodeError!void {
        const actual = try self.readBytes(4);
        if (!std.mem.eql(u8, actual, expected)) return DecodeError.InvalidMagic;
    }

    fn readBytes(self: *Reader, len: usize) DecodeError![]const u8 {
        if (len > self.remaining()) return DecodeError.UnexpectedEnd;
        const start = self.cursor;
        self.cursor += len;
        return self.bytes[start..self.cursor];
    }

    fn readU32(self: *Reader) DecodeError!u32 {
        const raw = try self.readBytes(4);
        return @as(u32, raw[0]) |
            (@as(u32, raw[1]) << 8) |
            (@as(u32, raw[2]) << 16) |
            (@as(u32, raw[3]) << 24);
    }

    fn readCount(self: *Reader) DecodeError!usize {
        return @intCast(try self.readU32());
    }

    fn readString(self: *Reader) DecodeError![]const u8 {
        const len = try self.readCount();
        return try self.readBytes(len);
    }

    fn readExt(self: *Reader) DecodeError!ext.Ext {
        return ext.Ext{
            .B0 = .{
                .a0 = try self.readField(),
                .a1 = try self.readField(),
            },
            .B1 = .{
                .a0 = try self.readField(),
                .a1 = try self.readField(),
            },
            .B2 = .{
                .a0 = try self.readField(),
                .a1 = try self.readField(),
            },
        };
    }

    fn readField(self: *Reader) DecodeError!field.Element {
        return field.Element.fromCanonical(try self.readU32()) catch return DecodeError.NonCanonicalField;
    }
};
