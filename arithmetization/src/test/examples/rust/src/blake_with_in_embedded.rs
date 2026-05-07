#![no_std]
#![no_main]

/// Blake2b‑F compression function (used in EIP‑152 precompile, zkVMs, etc.)
///
/// Reference: RFC 7693, plus the Java gist:
/// https://gist.github.com/DavePearce/fca4c7fcfac840dc362b1c907d672093
///
/// Note: this is a freestanding implementation
///
/// To run:
/// zkc-test blake_with_in_embedded.rs
use core::convert::TryInto;
use core::result::Result;
use core::result::Result::Err;
use core::result::Result::Ok;

include!("blake_core.rs");
include!("blake_test_vectors.rs");

core::arch::global_asm!(
    ".global _start",
    "_start:",
    "la sp, _stack_start", // SP start address from linker script
    "call main",
);

#[no_mangle]
fn main() -> ! {
    let input = hex_to_input(_TEST_VECTOR_5[0]);
    let expected = hex_to_expected(_TEST_VECTOR_5[1]);

    let code = match blake2b_f_eip152(&input) {
        Ok(result) if result == expected.as_slice() => 0, // success
        Ok(_) => 1,                                       // wrong result
        Err(_) => 2,                                      // compression failed
    };

    exit(code);
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

// required by the compiler
#[panic_handler]
fn panic(_: &core::panic::PanicInfo) -> ! {
    exit(3);
}
