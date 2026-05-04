#![no_std]
#![no_main]

/// Blake2b‑F compression function (used in EIP‑152 precompile, zkVMs, etc.)
///
/// Reference: RFC 7693, plus the Java gist:
/// https://gist.github.com/DavePearce/fca4c7fcfac840dc362b1c907d672093
///
/// Note: this is a freestanding implementation
use core::convert::TryInto;
use core::result::Result;
use core::result::Result::Err;
use core::result::Result::Ok;

include!("blake_core.rs");
include!("blake_test_vectors.rs");

core::arch::global_asm!(
    ".global _start",
    "_start:",
    "li sp, 0x087fffff", // set stack pointer to a known memory region
    "call main",
);

#[no_mangle]
fn main() -> ! {
    // Note: test vector 8 is not included for now as number of rounds is 0xffffffff
    let test_vectors = [TEST_VECTOR_4, TEST_VECTOR_5, TEST_VECTOR_6, TEST_VECTOR_7];

    let mut codes = [0, 0, 0, 0];

    for i in 0..test_vectors.len() {
        let input = hex_to_input(test_vectors[i][0]);
        let expected = hex_to_expected(test_vectors[i][1]);

        codes[i] = match blake2b_f_eip152(&input) {
            Ok(result) if result == expected.as_slice() => 0, // success
            Ok(_) => 1,                                       // wrong result
            Err(_) => 2,                                      // compression failed
        };
    }

    // Encode the 5 codes into a single exit code (e.g. 0000 for all pass, 1000 for 1st test failing, etc.)
    process::exit(codes[0] * 1000 + codes[1] * 100 + codes[2] * 10 + codes[3]);
}

fn exit(code: i32) -> ! {
    unsafe {
        core::arch::asm!(
            "mv a0, {0}",  // exit code
            "li a7, 93",   // syscall number for exit
            "ecall",
            in(reg) code,
            options(noreturn)
        );
    }
}

// required by the compiler
#[panic_handler]
fn panic(_: &core::panic::PanicInfo) -> ! {
    exit(3);
}
