#![no_std]
#![no_main]

include!("custom_std.rs");

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