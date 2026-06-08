fn exit(code: u32) noreturn {
    // no OS to return to, signal halt via ecall
    asm volatile (
        \\li a0, code   # exit code 0
        \\li a7, 93  # syscall number for exit
        \\ecall
    );
}

fn panic() noreturn {
    exit(1);
    unreachable;
}
