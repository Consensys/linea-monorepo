/// Blake2b‑F compression function (used in EIP‑152 precompile, zkVMs, etc.)
///
/// Reference: RFC 7693, plus the Java gist:
/// https://gist.github.com/DavePearce/fca4c7fcfac840dc362b1c907d672093
///
/// To run:
/// rustc src/blake_from_main.rs -o bin/blake_from_main && ./bin/blake_from_main; echo $?
use std::convert::TryInto;
use std::process;

include!("blake_core.rs");
include!("blake_test_vectors.rs");

fn main() {
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

        println!(
            "Test vector {}: {}",
            i,
            if codes[i] == 0 { "PASS" } else { "FAIL" }
        );
    }

    // Encode the 5 codes into a single exit code (e.g. 0000 for all pass, 1000 for 1st test failing, etc.)
    process::exit(codes[0] * 1000 + codes[1] * 100 + codes[2] * 10 + codes[3]);
}
