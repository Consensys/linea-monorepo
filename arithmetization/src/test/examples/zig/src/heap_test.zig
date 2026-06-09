const custom_std = @import("wrappers").custom_std;

// Linker-defined symbols whose addresses mark the heap boundaries
extern var _heap_start: u8;
extern var _heap_end: u8;

export fn main() noreturn {
    // Read the heap boundary addresses
    const heap_start = @intFromPtr(&_heap_start);
    const heap_end = @intFromPtr(&_heap_end);

    // Check the heap is at least 1 byte
    if (heap_start + 1 > heap_end) {
        custom_std.panic();
    }

    // Write to the first heap byte and read it back through a volatile pointer
    const allocation: *volatile u8 = @ptrFromInt(heap_start);
    allocation.* = 42;

    if (allocation.* != 42) {
        custom_std.panic();
    }

    custom_std.exit(0);
}
