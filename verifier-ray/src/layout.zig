const std = @import("std");

const fri_types = @import("fri/types.zig");

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
