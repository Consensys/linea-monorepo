#![no_std]
#![no_main]

// no_mangle so the linker can find this entry point by its exact name
#[no_mangle]
pub extern "C" fn _start() -> ! {
    let _ = add(2, 3);
    exit() // no OS to return to, signal halt via ecall
}

fn add(op1: u8, op2: u8) -> u16 {
    (op1 as u16) + (op2 as u16)
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
#[panic_handler]
fn panic(_: &core::panic::PanicInfo) -> ! {
    exit();
}


