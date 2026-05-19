const std = @import("std");
const builtin = @import("builtin");

var active_allocator: ?std.mem.Allocator = null;

pub fn init(allocator: std.mem.Allocator) void {
    active_allocator = allocator;
}

pub fn get() std.mem.Allocator {
    if (builtin.cpu.arch == .riscv64) {
        return active_allocator orelse @panic("rollup guest allocator not initialized");
    }

    return active_allocator orelse std.heap.page_allocator;
}
