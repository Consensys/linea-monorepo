const std = @import("std");
const verifier_ray = @import("verifier_ray");

const field = verifier_ray.field.koalabear;
const ext = verifier_ray.field.koalabear_ext;
const layout_mod = verifier_ray.layout;
const dq_layout_mod = verifier_ray.dq_layout;
const vectors = @import("test_vectors");

test "layout loader decodes vectored bytes" {
    var loaded = try layout_mod.decode(std.testing.allocator, vectors.layout_loader_vector[0..]);
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
    try std.testing.expectError(error.UnexpectedEnd, layout_mod.decode(std.testing.allocator, vectors.layout_loader_vector[0 .. vectors.layout_loader_vector.len - 1]));

    var bad_magic = vectors.layout_loader_vector;
    bad_magic[0] = 0;
    try std.testing.expectError(error.InvalidMagic, layout_mod.decode(std.testing.allocator, bad_magic[0..]));

    const trailing = vectors.layout_loader_vector ++ [_]u8{0};
    try std.testing.expectError(error.TrailingBytes, layout_mod.decode(std.testing.allocator, trailing[0..]));

    var bad_rail = vectors.layout_loader_vector;
    bad_rail[77] = 9;
    try std.testing.expectError(error.InvalidRail, layout_mod.decode(std.testing.allocator, bad_rail[0..]));
}

test "DQ layout loader decodes vectored bytes" {
    var loaded = try dq_layout_mod.decode(std.testing.allocator, vectors.dq_layout_loader_vector[0..]);
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
    try std.testing.expectError(error.UnexpectedEnd, dq_layout_mod.decode(std.testing.allocator, vectors.dq_layout_loader_vector[0 .. vectors.dq_layout_loader_vector.len - 1]));

    var bad_magic = vectors.dq_layout_loader_vector;
    bad_magic[0] = 0;
    try std.testing.expectError(error.InvalidMagic, dq_layout_mod.decode(std.testing.allocator, bad_magic[0..]));

    const trailing = vectors.dq_layout_loader_vector ++ [_]u8{0};
    try std.testing.expectError(error.TrailingBytes, dq_layout_mod.decode(std.testing.allocator, trailing[0..]));

    var bad_field = vectors.dq_layout_loader_vector;
    writeU32Le(&bad_field, 16, field.modulus);
    try std.testing.expectError(error.NonCanonicalField, dq_layout_mod.decode(std.testing.allocator, bad_field[0..]));
}

test "DQ layout loader rejects impossible counts before allocation" {
    const huge_level_count = [_]u8{
        'D', 'Q', 'L', '1',
        255, 255, 255, 255,
    };
    try std.testing.expectError(error.UnexpectedEnd, dq_layout_mod.decode(std.testing.allocator, huge_level_count[0..]));

    const huge_eval_count = [_]u8{
        'D', 'Q', 'L', '1',
        1,   0,   0,   0,
        4,   0,   0,   0,
        255, 255, 255, 255,
    };
    try std.testing.expectError(error.UnexpectedEnd, dq_layout_mod.decode(std.testing.allocator, huge_eval_count[0..]));

    const huge_term_count = [_]u8{
        'D', 'Q', 'L', '1',
        1,   0,   0,   0,
        4,   0,   0,   0,
        1,   0,   0,   0,
        0,   0,   0,   0,
        0,   0,   0,   0,
        0,   0,   0,   0,
        0,   0,   0,   0,
        0,   0,   0,   0,
        0,   0,   0,   0,
        255, 255, 255, 255,
    };
    try std.testing.expectError(error.UnexpectedEnd, dq_layout_mod.decode(std.testing.allocator, huge_term_count[0..]));

    const huge_air_count = [_]u8{
        'D', 'Q', 'L', '1',
        1,   0,   0,   0,
        4,   0,   0,   0,
        0,   0,   0,   0,
        255, 255, 255, 255,
    };
    try std.testing.expectError(error.UnexpectedEnd, dq_layout_mod.decode(std.testing.allocator, huge_air_count[0..]));
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
