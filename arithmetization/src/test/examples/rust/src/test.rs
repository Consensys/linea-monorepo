#![no_std]
#![no_main]

// no_mangle so the linker can find this entry point by its exact name
#[unsafe(naked)]
#[no_mangle]
pub unsafe extern "C" fn _start() -> ! {
    core::arch::naked_asm!(
        "li sp, 0x7FFFFFF", // set stack pointer to a known memory region
        "call main",
    );
}

#[no_mangle]
fn main() -> ! {
    let _ = add(2, 3);
    exit() // no OS to return to, signal halt via ecall
}

fn add(op1: u8, op2: u8) -> u16 {
    (op1 as u16).wrapping_add(op2 as u16)
}

fn exit() -> ! {
    unsafe {
        core::arch::asm!(
            "li a0, 0",   // exit code 0
            "li a7, 93",  // syscall number for exit
            "ecall",
            options(noreturn)
        );
    }
}

// required by the compiler even if unreachable — no std means no default panic handler
// Note: that core contains .c instructions that ends up in the ELF file even if we exluce that extension from the targer, so we use opt-level=2 to remove unused code. To actually completetely avoid .c instructions, we need to use a custom JSON configuration for the targer and a nightly compiler
#[panic_handler]
fn panic(_: &core::panic::PanicInfo) -> ! {
    exit();
}


