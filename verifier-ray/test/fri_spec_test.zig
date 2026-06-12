const std = @import("std");
const spec = @import("fri_spec");

test "fri_spec params fields" {
    try std.testing.expectEqual(@as(u32, 32), spec.params.n);
    try std.testing.expectEqual(@as(u32, 8), spec.params.d);
    try std.testing.expectEqual(@as(u32, 4), spec.params.num_queries);
    try std.testing.expectEqual(@as(u32, 3), spec.params.num_rounds);
    try std.testing.expectEqual(@as(u32, 0), spec.params.grinding);
    try std.testing.expectEqual(@as(usize, 4), spec.params.domain_gens.len);
    try std.testing.expectEqual(@as(usize, 4), spec.params.domain_gens_inv.len);
}

test "fri_spec layout col_slots lookup" {
    const col0 = comptime spec.layout.colSlot("col0");
    try std.testing.expectEqual(@as(u32, 0), col0.tree_idx);
    try std.testing.expectEqual(@as(u32, 0), col0.poly_idx);
    try std.testing.expectEqual(.base, col0.rail);

    const col1 = comptime spec.layout.colSlot("col1");
    try std.testing.expectEqual(@as(u32, 0), col1.tree_idx);
    try std.testing.expectEqual(@as(u32, 1), col1.poly_idx);
    try std.testing.expectEqual(.ext, col1.rail);
}

test "fri_spec layout structural fields" {
    try std.testing.expectEqual(@as(u32, 1), spec.layout.num_trees);
    try std.testing.expectEqual(@as(usize, 1), spec.layout.tree_size.len);
    try std.testing.expectEqual(@as(u32, 32), spec.layout.tree_size[0]);
    try std.testing.expectEqual(@as(usize, 2), spec.layout.col_slots.len);
    try std.testing.expectEqual(@as(usize, 0), spec.layout.air_chunk_slots.len);
}

test "fri_spec dq_layout structure" {
    try std.testing.expectEqual(@as(usize, 1), spec.dq_layout.levels.len);
    const lv = spec.dq_layout.levels[0];
    try std.testing.expectEqual(@as(u32, 32), lv.size);
    try std.testing.expectEqual(@as(usize, 2), lv.shifts.len);
    try std.testing.expectEqual(@as(u32, 0), lv.shifts[0]);
    try std.testing.expectEqual(@as(u32, 1), lv.shifts[1]);
    try std.testing.expectEqual(@as(usize, 2), lv.col_groups.len);
    try std.testing.expectEqual(@as(usize, 0), lv.air_chunks.len);
}
