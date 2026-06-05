const std = @import("std");

const fri_types = @import("fri/types.zig");

pub const DecodeError = std.mem.Allocator.Error || error{
    UnexpectedEnd,
    TrailingBytes,
    InvalidMagic,
    InvalidRail,
    InvalidDimensions,
};

pub const Slot = struct {
    tree_idx: u32,
    poly_idx: u32,
    rail: fri_types.Rail,
};

pub const Layout = struct {
    num_trees: u32,
    setup_begin: u32,
    setup_end: u32,
    trace_begin: []const u32 = &.{},
    trace_end: []const u32 = &.{},
    air_begin: u32,
    air_end: u32,
    tree_size: []const u32 = &.{},
    col_slot: std.StringHashMap(Slot),
    air_chunk_slot: std.StringHashMap(Slot),
};

pub const OwnedLayout = struct {
    value: Layout,
    allocator: std.mem.Allocator,

    pub fn deinit(self: *OwnedLayout) void {
        self.allocator.free(self.value.trace_begin);
        self.allocator.free(self.value.trace_end);
        self.allocator.free(self.value.tree_size);
        self.value.col_slot.deinit();
        self.value.air_chunk_slot.deinit();
        self.* = undefined;
    }
};

/// Decodes `LAY1` bytes. String keys borrow from `encoded`, so callers must
/// keep the byte buffer alive for at least as long as the returned layout.
pub fn decode(allocator: std.mem.Allocator, encoded: []const u8) DecodeError!OwnedLayout {
    var reader = Reader.init(encoded);
    try reader.expectMagic("LAY1");

    var owned = OwnedLayout{
        .allocator = allocator,
        .value = .{
            .num_trees = 0,
            .setup_begin = 0,
            .setup_end = 0,
            .air_begin = 0,
            .air_end = 0,
            .col_slot = std.StringHashMap(Slot).init(allocator),
            .air_chunk_slot = std.StringHashMap(Slot).init(allocator),
        },
    };
    errdefer owned.deinit();

    owned.value.num_trees = try reader.readU32();
    owned.value.setup_begin = try reader.readU32();
    owned.value.setup_end = try reader.readU32();

    const trace_count = try reader.readCount();
    owned.value.trace_begin = try reader.readU32Slice(allocator, trace_count);
    owned.value.trace_end = try reader.readU32Slice(allocator, trace_count);

    owned.value.air_begin = try reader.readU32();
    owned.value.air_end = try reader.readU32();

    const tree_count = try reader.readCount();
    if (tree_count != @as(usize, @intCast(owned.value.num_trees))) return DecodeError.InvalidDimensions;
    owned.value.tree_size = try reader.readU32Slice(allocator, tree_count);

    try reader.readSlotMap(&owned.value.col_slot);
    try reader.readSlotMap(&owned.value.air_chunk_slot);

    if (!reader.isDone()) return DecodeError.TrailingBytes;
    return owned;
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

    fn readU8(self: *Reader) DecodeError!u8 {
        return (try self.readBytes(1))[0];
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

    fn readU32Slice(self: *Reader, allocator: std.mem.Allocator, count: usize) DecodeError![]const u32 {
        if (count > self.remaining() / 4) return DecodeError.UnexpectedEnd;
        const out = try allocator.alloc(u32, count);
        errdefer allocator.free(out);
        for (out) |*slot| slot.* = try self.readU32();
        return out;
    }

    fn readString(self: *Reader) DecodeError![]const u8 {
        const len = try self.readCount();
        return try self.readBytes(len);
    }

    fn readRail(self: *Reader) DecodeError!fri_types.Rail {
        return switch (try self.readU8()) {
            0 => .base,
            1 => .ext,
            else => return DecodeError.InvalidRail,
        };
    }

    fn readSlotMap(self: *Reader, map: *std.StringHashMap(Slot)) DecodeError!void {
        const count = try self.readCount();
        var index: usize = 0;
        while (index < count) : (index += 1) {
            const name = try self.readString();
            const slot = Slot{
                .tree_idx = try self.readU32(),
                .poly_idx = try self.readU32(),
                .rail = try self.readRail(),
            };
            try map.put(name, slot);
        }
    }
};
