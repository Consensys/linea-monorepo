#![no_std]
#![no_main]

// To run:
// zkc-test test.rs IN_BYTES="0x01" (pass)
// zkc-test test.rs IN_BYTES="0x02" (fail)

include!("custom_std.rs");

const SECOND_ADDEND: u8 = 2;

#[no_mangle]
fn main() -> ! {
    let first_addend = read_first_addend();
    let r = add(first_addend, SECOND_ADDEND);
    if r != 3 {
        exit(1); // test failed
    }
    exit(0) // test passed
}

fn add(op1: u8, op2: u8) -> u16 {
    let r = (op1 as u16).wrapping_add(op2 as u16);
    r
}

fn read_first_addend() -> u8 {
    static mut BUF: [u8; 1] = [0u8; 1];
    unsafe {
        read_memory(
            (&raw mut BUF) as *mut u8,
            &raw const _input_start as usize,
            1,
        );
        BUF[0]
    }
}
