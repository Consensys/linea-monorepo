const std = @import("std");
const verifier_ray = @import("verifier_ray");

const field = verifier_ray.field.koalabear;
const ext = verifier_ray.field.koalabear_ext;
const layout_mod = verifier_ray.layout;
const dq_layout_mod = verifier_ray.dq_layout;

test "layout loader decodes vectored bytes" {
    var loaded = try layout_mod.decode(std.testing.allocator, layout_vector[0..]);
    defer loaded.deinit();

    const layout = loaded.value;
    try std.testing.expectEqual(@as(u32, 3), layout.num_trees);
    try std.testing.expectEqual(@as(u32, 0), layout.setup_begin);
    try std.testing.expectEqual(@as(u32, 1), layout.setup_end);
    try expectU32Slice(&.{ 1, 3 }, layout.trace_begin);
    try expectU32Slice(&.{ 3, 5 }, layout.trace_end);
    try std.testing.expectEqual(@as(u32, 5), layout.air_begin);
    try std.testing.expectEqual(@as(u32, 6), layout.air_end);
    try expectU32Slice(&.{ 4, 2, 8 }, layout.tree_size);

    try expectSlot(layout.col_slot.get("A").?, 0, 1, .base);
    try expectSlot(layout.col_slot.get("B.shift").?, 1, 0, .ext);
    try expectSlot(layout.col_slot.get("C").?, 2, 4, .base);
    try expectSlot(layout.air_chunk_slot.get("air.main.0").?, 2, 0, .ext);
    try std.testing.expect(layout.col_slot.get("missing") == null);
}

test "layout loader rejects malformed bytes" {
    try std.testing.expectError(error.UnexpectedEnd, layout_mod.decode(std.testing.allocator, layout_vector[0 .. layout_vector.len - 1]));

    var bad_magic = layout_vector;
    bad_magic[0] = 0;
    try std.testing.expectError(error.InvalidMagic, layout_mod.decode(std.testing.allocator, bad_magic[0..]));

    const trailing = layout_vector ++ [_]u8{0};
    try std.testing.expectError(error.TrailingBytes, layout_mod.decode(std.testing.allocator, trailing[0..]));

    var bad_rail = layout_vector;
    bad_rail[77] = 9;
    try std.testing.expectError(error.InvalidRail, layout_mod.decode(std.testing.allocator, bad_rail[0..]));
}

test "DQ layout loader decodes vectored bytes" {
    var loaded = try dq_layout_mod.decode(std.testing.allocator, dq_layout_vector[0..]);
    defer loaded.deinit();

    const layout = loaded.value;
    try expectU32Slice(&.{ 4, 2 }, layout.sizes);
    try expectExt(layout.eval_points[0][0], .{ 2, 3, 5, 7, 11, 13 });
    try expectExt(layout.eval_points[0][1], .{ 17, 19, 23, 29, 31, 37 });
    try expectExt(layout.eval_points[1][0], .{ 41, 43, 47, 53, 59, 61 });

    try expectStringSlice(&.{ "A", "B.shift" }, layout.column_names[0][0]);
    try expectStringSlice(&.{ "A", "B@1" }, layout.column_keys[0][0]);
    try expectStringSlice(&.{"C"}, layout.column_names[0][1]);
    try expectStringSlice(&.{"C@rot"}, layout.column_keys[0][1]);
    try expectStringSlice(&.{"air.main.0"}, layout.air_chunks[0]);
    try expectStringSlice(&.{"small"}, layout.column_names[1][0]);
    try expectStringSlice(&.{"small"}, layout.column_keys[1][0]);
    try expectStringSlice(&.{}, layout.air_chunks[1]);
}

test "DQ layout loader rejects malformed bytes" {
    try std.testing.expectError(error.UnexpectedEnd, dq_layout_mod.decode(std.testing.allocator, dq_layout_vector[0 .. dq_layout_vector.len - 1]));

    var bad_magic = dq_layout_vector;
    bad_magic[0] = 0;
    try std.testing.expectError(error.InvalidMagic, dq_layout_mod.decode(std.testing.allocator, bad_magic[0..]));

    const trailing = dq_layout_vector ++ [_]u8{0};
    try std.testing.expectError(error.TrailingBytes, dq_layout_mod.decode(std.testing.allocator, trailing[0..]));

    var bad_field = dq_layout_vector;
    writeU32Le(&bad_field, 16, field.modulus);
    try std.testing.expectError(error.NonCanonicalField, dq_layout_mod.decode(std.testing.allocator, bad_field[0..]));
}

fn expectSlot(actual: layout_mod.Slot, tree_idx: u32, poly_idx: u32, rail: verifier_ray.fri.Rail) !void {
    try std.testing.expectEqual(tree_idx, actual.tree_idx);
    try std.testing.expectEqual(poly_idx, actual.poly_idx);
    try std.testing.expectEqual(rail, actual.rail);
}

fn expectU32Slice(expected: []const u32, actual: []const u32) !void {
    try std.testing.expectEqual(expected.len, actual.len);
    for (expected, actual) |expected_value, actual_value| {
        try std.testing.expectEqual(expected_value, actual_value);
    }
}

fn expectStringSlice(expected: []const []const u8, actual: []const []const u8) !void {
    try std.testing.expectEqual(expected.len, actual.len);
    for (expected, actual) |expected_value, actual_value| {
        try std.testing.expectEqualStrings(expected_value, actual_value);
    }
}

fn expectExt(actual: ext.Ext, expected: [6]u32) !void {
    const limbs = [_]field.Element{
        actual.B0.a0,
        actual.B0.a1,
        actual.B1.a0,
        actual.B1.a1,
        actual.B2.a0,
        actual.B2.a1,
    };
    for (limbs, expected) |limb, expected_value| {
        try std.testing.expectEqual(expected_value, limb.value);
    }
}

fn writeU32Le(bytes: []u8, offset: usize, value: u32) void {
    bytes[offset + 0] = @truncate(value);
    bytes[offset + 1] = @truncate(value >> 8);
    bytes[offset + 2] = @truncate(value >> 16);
    bytes[offset + 3] = @truncate(value >> 24);
}

const layout_vector = [_]u8{
    76, 65, 89, 49, 3,   0,   0,   0,   0,   0,   0,   0,  1,   0,  0,   0,
    2,  0,  0,  0,  1,   0,   0,   0,   3,   0,   0,   0,  3,   0,  0,   0,
    5,  0,  0,  0,  5,   0,   0,   0,   6,   0,   0,   0,  3,   0,  0,   0,
    4,  0,  0,  0,  2,   0,   0,   0,   8,   0,   0,   0,  3,   0,  0,   0,
    1,  0,  0,  0,  65,  0,   0,   0,   0,   1,   0,   0,  0,   0,  7,   0,
    0,  0,  66, 46, 115, 104, 105, 102, 116, 1,   0,   0,  0,   0,  0,   0,
    0,  1,  1,  0,  0,   0,   67,  2,   0,   0,   0,   4,  0,   0,  0,   0,
    1,  0,  0,  0,  10,  0,   0,   0,   97,  105, 114, 46, 109, 97, 105, 110,
    46, 48, 2,  0,  0,   0,   0,   0,   0,   0,   1,
};

const dq_layout_vector = [_]u8{
    68,  81,  76, 49, 2,   0,   0,   0,   4,   0,   0,  0,   2,   0,   0,   0,
    2,   0,   0,  0,  3,   0,   0,   0,   5,   0,   0,  0,   7,   0,   0,   0,
    11,  0,   0,  0,  13,  0,   0,   0,   17,  0,   0,  0,   19,  0,   0,   0,
    23,  0,   0,  0,  29,  0,   0,   0,   31,  0,   0,  0,   37,  0,   0,   0,
    2,   0,   0,  0,  1,   0,   0,   0,   65,  1,   0,  0,   0,   65,  7,   0,
    0,   0,   66, 46, 115, 104, 105, 102, 116, 3,   0,  0,   0,   66,  64,  49,
    1,   0,   0,  0,  1,   0,   0,   0,   67,  5,   0,  0,   0,   67,  64,  114,
    111, 116, 1,  0,  0,   0,   10,  0,   0,   0,   97, 105, 114, 46,  109, 97,
    105, 110, 46, 48, 2,   0,   0,   0,   1,   0,   0,  0,   41,  0,   0,   0,
    43,  0,   0,  0,  47,  0,   0,   0,   53,  0,   0,  0,   59,  0,   0,   0,
    61,  0,   0,  0,  1,   0,   0,   0,   5,   0,   0,  0,   115, 109, 97,  108,
    108, 5,   0,  0,  0,   115, 109, 97,  108, 108, 0,  0,   0,   0,
};
