/// Blake2b‑F compression function (used in EIP‑152 precompile, zkVMs, etc.)
///
/// Reference: RFC 7693, plus the Java gist:
/// https://gist.github.com/DavePearce/fca4c7fcfac840dc362b1c907d672093
/// 
/// rustc src/blake_from_main.rs -o bin/blake_from_main && ./bin/blake_from_main; echo $?

use std::convert::TryInto;
use std::process;

/// Initial state vector (IV) for Blake2b
const IV: [u64; 8] = [
    0x6a09e667f3bcc908,
    0xbb67ae8584caa73b,
    0x3c6ef372fe94f82b,
    0xa54ff53a5f1d36f1,
    0x510e527fade682d1,
    0x9b05688c2b3e6c1f,
    0x1f83d9abfb41bd6b,
    0x5be0cd19137e2179,
];

/// Sigma permutation table (10 rounds × 16 entries)
const SIGMA: [[usize; 16]; 10] = [
    [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15],
    [14, 10, 4, 8, 9, 15, 13, 6, 1, 12, 0, 2, 11, 7, 5, 3],
    [11, 8, 12, 0, 5, 2, 15, 13, 10, 14, 3, 6, 7, 1, 9, 4],
    [7, 9, 3, 1, 13, 12, 11, 14, 2, 6, 5, 10, 4, 0, 15, 8],
    [9, 0, 5, 7, 2, 4, 10, 15, 14, 1, 11, 12, 6, 8, 3, 13],
    [2, 12, 6, 10, 0, 11, 8, 3, 4, 13, 7, 5, 15, 14, 1, 9],
    [12, 5, 1, 15, 14, 13, 4, 10, 0, 7, 6, 3, 9, 2, 8, 11],
    [13, 11, 7, 14, 12, 1, 3, 9, 5, 0, 15, 4, 8, 6, 2, 10],
    [6, 15, 14, 9, 11, 3, 0, 8, 12, 2, 13, 7, 1, 4, 10, 5],
    [10, 2, 8, 4, 7, 6, 1, 5, 15, 11, 9, 14, 3, 12, 13, 0],
];

/// Rotation constants for the G mixing function
const R1: u32 = 32;
const R2: u32 = 24;
const R3: u32 = 16;
const R4: u32 = 63;

/// The G mixing function: mixes two words from the message into the state.
#[inline(always)]
fn g(v: &mut [u64; 16], a: usize, b: usize, c: usize, d: usize, x: u64, y: u64) {
    v[a] = v[a].wrapping_add(v[b]).wrapping_add(x);
    v[d] = (v[d] ^ v[a]).rotate_right(R1);

    v[c] = v[c].wrapping_add(v[d]);
    v[b] = (v[b] ^ v[c]).rotate_right(R2);

    v[a] = v[a].wrapping_add(v[b]).wrapping_add(y);
    v[d] = (v[d] ^ v[a]).rotate_right(R3);

    v[c] = v[c].wrapping_add(v[d]);
    v[b] = (v[b] ^ v[c]).rotate_right(R4);
}

/// Blake2b‑F compression function.
///
/// # Arguments
///
/// * `rounds` - number of rounds (typically 12 for Blake2b)
/// * `h` - current state (8 × u64), modified in place
/// * `m` - message block (16 × u64)
/// * `t` - offset counters `[t0, t1]`
/// * `f` - if `true`, this is the final block
///
/// After the call, `h` contains the new state.
pub fn blake2b_f(rounds: u32, h: &mut [u64; 8], m: &[u64; 16], t: [u64; 2], f: bool) {
    // Initialize working vector v[0..15]
    let mut v: [u64; 16] = [0u64; 16];

    // v[0..8] = h[0..8]
    v[..8].copy_from_slice(h);

    // v[8..16] = IV[0..8]
    v[8..].copy_from_slice(&IV);

    // Mix in the counter and finalization flag
    v[12] ^= t[0];
    v[13] ^= t[1];
    if f {
        v[14] ^= u64::MAX; // finalization flag
    }

    // Run the rounds
    for i in 0..(rounds as usize) {
        let s = &SIGMA[i % 10];

        // Column step
        g(&mut v, 0, 4, 8, 12, m[s[0]], m[s[1]]);
        g(&mut v, 1, 5, 9, 13, m[s[2]], m[s[3]]);
        g(&mut v, 2, 6, 10, 14, m[s[4]], m[s[5]]);
        g(&mut v, 3, 7, 11, 15, m[s[6]], m[s[7]]);

        // Diagonal step
        g(&mut v, 0, 5, 10, 15, m[s[8]], m[s[9]]);
        g(&mut v, 1, 6, 11, 12, m[s[10]], m[s[11]]);
        g(&mut v, 2, 7, 8, 13, m[s[12]], m[s[13]]);
        g(&mut v, 3, 4, 9, 14, m[s[14]], m[s[15]]);
    }

    // Finalize: h[i] = h[i] ^ v[i] ^ v[i+8]
    for i in 0..8 {
        h[i] ^= v[i] ^ v[i + 8];
    }
}

// ─────────────────────────────────────────────────────────────────────────────
// Helper functions for parsing / encoding (matching the Java gist's test vector)
// ─────────────────────────────────────────────────────────────────────────────

/// Parse a 426-character hex string into a [u8; 213], for input.
pub fn hex_to_input(s: &str) -> [u8; 213] {
    let mut out = [0u8; 213];
    for i in 0..213 {
        out[i] = u8::from_str_radix(&s[i * 2..i * 2 + 2], 16).unwrap();
    }
    out
}

/// Parse a 128-character hex string into a [u8; 64], for expected output.
pub fn hex_to_expected(s: &str) -> [u8; 64] {
    let mut out = [0u8; 64];
    for i in 0..64 {
        out[i] = u8::from_str_radix(&s[i * 2..i * 2 + 2], 16).unwrap();
    }
    out
}

/// Read a little‑endian u64 from a byte slice at offset `i * 8`.
pub fn read_u64_le(data: &[u8], i: usize) -> u64 {
    let off = i * 8;
    u64::from_le_bytes(data[off..off + 8].try_into().unwrap())
}

/// Write a little‑endian u64 into a byte slice at offset `i * 8`.
pub fn write_u64_le(data: &mut [u8], i: usize, val: u64) {
    let off = i * 8;
    data[off..off + 8].copy_from_slice(&val.to_le_bytes());
}

// ─────────────────────────────────────────────────────────────────────────────
// EIP‑152 style interface (input = 213 bytes, output = 64 bytes)
// ─────────────────────────────────────────────────────────────────────────────

/// Runs the Blake2b‑F compression given the EIP‑152 input format:
///
/// ```text
/// input (213 bytes):
///   [0..4]     rounds      (big‑endian u32)
///   [4..68]    h           (8 × u64, little‑endian)
///   [68..196]  m           (16 × u64, little‑endian)
///   [196..212] t           (2 × u64, little‑endian)
///   [212]      f           (0 or 1)
/// ```
///
/// Returns the new state `h` as 64 bytes (little‑endian).
pub fn blake2b_f_eip152(input: &[u8]) -> Result<[u8; 64], &'static str> {
    if input.len() != 213 {
        return Err("input must be exactly 213 bytes");
    }

    // Parse rounds (big‑endian u32)
    let rounds = u32::from_be_bytes(input[0..4].try_into().unwrap());

    // Parse h (8 × u64, little‑endian)
    let mut h = [0u64; 8];
    for i in 0..8 {
        h[i] = read_u64_le(&input[4..], i);
    }

    // Parse m (16 × u64, little‑endian)
    let mut m = [0u64; 16];
    for i in 0..16 {
        m[i] = read_u64_le(&input[68..], i);
    }

    // Parse t (2 × u64, little‑endian)
    let t0 = read_u64_le(&input[196..], 0);
    let t1 = read_u64_le(&input[196..], 1);

    // Parse f (single byte, must be 0 or 1)
    let f_byte = input[212];
    if f_byte > 1 {
        return Err("f must be 0 or 1");
    }
    let f = f_byte == 1;

    // Run compression
    blake2b_f(rounds, &mut h, &m, [t0, t1], f);

    // Encode result as 64 bytes (little‑endian)
    let mut out = [0u8; 64];
    for i in 0..8 {
        write_u64_le(&mut out, i, h[i]);
    }

    Ok(out)
}

fn main() {
    let test_vector_4 = [
        "0000000048c9bdf267e6096a3ba7ca8485ae67bb2bf894fe72f36e3cf1361d5f3af54fa5d182e6ad7f520e511f6c3e2b8c68059b6bbd41fbabd9831f79217e1319cde05b61626300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000300000000000000000000000000000001",
        "08c9bcf367e6096a3ba7ca8485ae67bb2bf894fe72f36e3cf1361d5f3af54fa5d282e6ad7f520e511f6c3e2b8c68059b9442be0454267ce079217e1319cde05b" ];

    let test_vector_5 =["0000000c48c9bdf267e6096a3ba7ca8485ae67bb2bf894fe72f36e3cf1361d5f3af54fa5d182e6ad7f520e511f6c3e2b8c68059b6bbd41fbabd9831f79217e1319cde05b61626300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000300000000000000000000000000000001","ba80a53f981c4d0d6a2797b69f12f6e94c212f14685ac4b74b12bb6fdbffa2d17d87c5392aab792dc252d5de4533cc9518d38aa8dbf1925ab92386edd4009923"];

    let test_vector_6 = [
        "0000000c48c9bdf267e6096a3ba7ca8485ae67bb2bf894fe72f36e3cf1361d5f3af54fa5d182e6ad7f520e511f6c3e2b8c68059b6bbd41fbabd9831f79217e1319cde05b61626300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000300000000000000000000000000000000",
        "75ab69d3190a562c51aef8d88f1c2775876944407270c42c9844252c26d2875298743e7f6d5ea2f2d3e8d226039cd31b4e426ac4f2d3d666a610c2116fde4735" ];

    let test_vector_7 = [
        "0000000148c9bdf267e6096a3ba7ca8485ae67bb2bf894fe72f36e3cf1361d5f3af54fa5d182e6ad7f520e511f6c3e2b8c68059b6bbd41fbabd9831f79217e1319cde05b61626300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000300000000000000000000000000000001",
        "b63a380cb2897d521994a85234ee2c181b5f844d2c624c002677e9703449d2fba551b3a8333bcdf5f2f7e08993d53923de3d64fcc68c034e717b9293fed7a421" ];

    let _test_vector_8 = [
        "ffffffff48c9bdf267e6096a3ba7ca8485ae67bb2bf894fe72f36e3cf1361d5f3af54fa5d182e6ad7f520e511f6c3e2b8c68059b6bbd41fbabd9831f79217e1319cde05b61626300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000300000000000000000000000000000001",
        "fc59093aafa9ab43daae0e914c57635c5402d8e3d2130eb9b3cc181de7f0ecf9b22bf99a7815ce16419e200e01846e6b5df8cc7703041bbceb571de6631d2615" ];

    // Note: test vector 8 is not included for now as number of rounds is 0xffffffff
    let test_vectors = [test_vector_4, test_vector_5, test_vector_6, test_vector_7];

    for i in 0..test_vectors.len() {
        let input = hex_to_input(test_vectors[i][0]);
        let expected = hex_to_expected(test_vectors[i][1]);

        let code = match blake2b_f_eip152(&input) {
            Ok(result) if result == expected.as_slice() => 0, // success
            Ok(_) => 1,                                       // wrong result
            Err(_) => 2,                                      // compression failed
        };

        println!(
            "Test vector {}: {}",
            i + 4,
            if code == 0 { "PASS" } else { "FAIL" }
        );
    }

    process::exit(0)
}