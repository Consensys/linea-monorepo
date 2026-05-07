core::arch::global_asm!(
    ".global _start",
    "_start:",
    "la sp, _stack_start", // SP from linker script
    "call main",
);

fn exit(code: i32) -> ! {
    unsafe {
        core::arch::asm!(
            "ecall",
            in("a0") code, // exit code
            in("a7") 93i32, // syscall number for exit (93)
            options(noreturn)
        );
    }
}

#[panic_handler]
fn panic(_: &core::panic::PanicInfo) -> ! {
    exit(i32::MAX) // use max int to indicate panic
}
