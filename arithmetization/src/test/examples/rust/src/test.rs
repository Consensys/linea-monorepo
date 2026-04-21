#![no_std]
#![no_main]

// required by the compiler even if unreachable — no std means no default panic handler
#[panic_handler]
fn panic(_: &core::panic::PanicInfo) -> ! {
    loop {}
}

fn add(op1: u8, op2: u8) -> u16 {
    (op1 as u16) + (op2 as u16)
}

// no_mangle so the linker can find this entry point by its exact name
#[no_mangle]
pub extern "C" fn _start() -> ! {
    let _ = add(2, 3);
    loop {} // no OS to return to, spin forever
}