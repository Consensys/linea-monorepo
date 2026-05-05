#![no_std]
#![no_main]

/// Blake2b‑F compression function (used in EIP‑152 precompile, zkVMs, etc.)
///
/// Reference: RFC 7693, plus the Java gist:
/// https://gist.github.com/DavePearce/fca4c7fcfac840dc362b1c907d672093
///
/// Note: this is a freestanding implementation that reads input from memory
///
/// To run using test vector 5:
/// zkc-test blake_with_in_bytes.rs IN_BYTES="0x0000000c48c9bdf267e6096a3ba7ca8485ae67bb2bf894fe72f36e3cf1361d5f3af54fa5d182e6ad7f520e511f6c3e2b8c68059b6bbd41fbabd9831f79217e1319cde05b61626300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000300000000000000000000000000000001ba80a53f981c4d0d6a2797b69f12f6e94c212f14685ac4b74b12bb6fdbffa2d17d87c5392aab792dc252d5de4533cc9518d38aa8dbf1925ab92386edd4009923"
use core::convert::TryInto;
use core::result::Result;
use core::result::Result::Err;
use core::result::Result::Ok;

include!("blake_core.rs");

core::arch::global_asm!(
    ".global _start",
    "_start:",
    "li sp, 0x087fffff", // set stack pointer to a known memory region
    "call main",
);

#[no_mangle]
fn main() -> ! {
    let (input_hex, expected_hex) = get_test_vector();
    let input = hex_to_input(input_hex);
    let expected = hex_to_expected(expected_hex);

    let code = match blake2b_f_eip152(&input) {
        Ok(result) if result == expected.as_slice() => 0, // success
        Ok(_) => 1,                                       // wrong result
        Err(_) => 2,                                      // compression failed
    };

    exit(code);
}

const TEST_VECTOR_ADDR: usize = 0x08800000; // input starts here per zkVM memory layout
const INPUT_LEN: usize = 426; // 213 bytes as hex
const OUTPUT_LEN: usize = 128; // 64 bytes as hex

fn get_test_vector() -> (&'static str, &'static str) {
    unsafe {
        let base = TEST_VECTOR_ADDR as *const u8;
        let input = core::str::from_utf8_unchecked(core::slice::from_raw_parts(base, INPUT_LEN));
        let expected = core::str::from_utf8_unchecked(core::slice::from_raw_parts(
            base.add(INPUT_LEN),
            OUTPUT_LEN,
        ));
        (input, expected)
    }
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
