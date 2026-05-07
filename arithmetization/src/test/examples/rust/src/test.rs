#![no_std]
#![no_main]

// no_mangle so the linker can find this entry point by its exact name
core::arch::global_asm!(
    ".global _start",
    "_start:",
    "li sp, 0x087fffff", // set stack pointer to a known memory region
    "call main",
);

#[no_mangle]
fn main() -> ! {
    let r = add(1, 1);
    if r != 2 {
        exit(1); // test failed
    }
    exit(0) // test passed
}

fn add(op1: u8, op2: u8) -> u16 {
    let r = (op1 as u16).wrapping_add(op2 as u16);
    r
}

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

// required by the compiler even if unreachable — no std means no default panic handler
#[panic_handler]
fn panic(_: &core::panic::PanicInfo) -> ! {
    exit(3);
}
