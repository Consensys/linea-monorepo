const custom_std = @import("wrappers").custom_std;

extern var _heap_start: u8;
extern var _heap_end: u8;

export fn main() noreturn {
    const heap_start = @intFromPtr(&_heap_start);
    const heap_end = @intFromPtr(&_heap_end);

    if (heap_start + 1 > heap_end) {
        custom_std.panic();
    }

    const allocation: *volatile u8 = @ptrFromInt(heap_start);
    allocation.* = 42;

    if (allocation.* != 42) {
        custom_std.panic();
    }

    custom_std.exit(0);
}
