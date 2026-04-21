export fn _start() noreturn {
    const a: i64 = 42;
    const b: i64 = 7;

    _ = a + b;
    _ = a - b;
    _ = a * b;
    // _ = @divTrunc(a, b);
    // _ = @rem(a, b);

    // freestanding has no OS to return to so halt the CPU until an interrupt fires, then sleep again
    while (true) {
        asm volatile ("wfi");
    }
}
