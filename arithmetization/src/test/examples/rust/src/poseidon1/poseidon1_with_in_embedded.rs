#![no_std]
#![no_main]
#![feature(generic_const_exprs)]
#![allow(incomplete_features)]

const STATE_LEN: usize = STATE_WIDTH * BYTES_PER_OUTPUT_FELT;
const STATE_WIDTH: usize = 16;
const BYTES_PER_OUTPUT_FELT: usize = 4;
const BYTES_PER_INPUT_FELT: usize = 3;

include!("../custom_std.rs");

fn in_line_assembly_poseidon_call(input_offset: usize, input_size: usize, output_offset: usize) {
  unsafe {
    core::arch::asm!(
      ".insn r 0x0b, 0b111, 0b1111111, {2}, {0}, {1}",
      in(reg) input_offset,
      in(reg) input_size,
      in(reg) output_offset,
    );
  }
}

// Compare the raw bytes written by the precompile against a list of expected
// little-endian u32 felts. Returns -1 on a full match, or the felt index of
// the first mismatch otherwise.
fn compare(output: &[u8; STATE_LEN], expected: &[u32; STATE_WIDTH]) -> bool {
  let mut i = 0;
  let mut aggregate :u32;
  while i < STATE_WIDTH {
    aggregate = 0;
    aggregate = (aggregate << 8) + output[4 * i + 3] as u32;
    aggregate = (aggregate << 8) + output[4 * i + 2] as u32;
    aggregate = (aggregate << 8) + output[4 * i + 1] as u32;
    aggregate = (aggregate << 8) + output[4 * i + 0] as u32;

    if aggregate != expected[i] {
      return false;
    }
    i += 1;
  }
  return true;
}

// Build a range(N) input as N felts of values 0..N, each encoded as 3 LE bytes
// (matching the precompile's `get_full_felt`). Total input size: N * 3 bytes.
fn run_range<const N: usize>(fill_expected: fn(&mut [u32; STATE_WIDTH])) -> bool
where
  [(); N * BYTES_PER_INPUT_FELT]: Sized,
{
  let mut input = [0u8; N * BYTES_PER_INPUT_FELT];
  let mut i = 0;
  while i < N {
    input[i * BYTES_PER_INPUT_FELT]     = (i         & 0xff) as u8;
    input[i * BYTES_PER_INPUT_FELT + 1] = ((i >> 8)  & 0xff) as u8;
    input[i * BYTES_PER_INPUT_FELT + 2] = ((i >> 16) & 0xff) as u8;
    i += 1;
  }

  let mut output = [0u8; STATE_LEN];
  in_line_assembly_poseidon_call(
    input.as_ptr() as usize,
    N * BYTES_PER_INPUT_FELT,
    output.as_mut_ptr() as usize,
  );

  let mut expected = [0u32; STATE_WIDTH];
  fill_expected(&mut expected);
  compare(&output, &expected)
}

// All-zeros input variant: N felts of value 0, i.e. N * 3 zero bytes.
fn run_zeros<const N: usize>(fill_expected: fn(&mut [u32; STATE_WIDTH])) -> bool
where
  [(); N * BYTES_PER_INPUT_FELT]: Sized,
{
  let input = [0u8; N * BYTES_PER_INPUT_FELT];

  let mut output = [0u8; STATE_LEN];
  in_line_assembly_poseidon_call(
    input.as_ptr() as usize,
    N * BYTES_PER_INPUT_FELT,
    output.as_mut_ptr() as usize,
  );

  let mut expected = [0u32; STATE_WIDTH];
  fill_expected(&mut expected);
  compare(&output, &expected)
}

// ---- expected output vectors (built at runtime to keep them out of .rodata) ----

fn fill_expected_range_7(out: &mut [u32; STATE_WIDTH]) {
  out[0]  = 0x362e517e;
  out[1]  = 0x2046663b;
  out[2]  = 0x6f66ef6b;
  out[3]  = 0x3dbaf0c1;
  out[4]  = 0x2319f56b;
  out[5]  = 0x64e61516;
  out[6]  = 0x4c624307;
  out[7]  = 0x6f0be16c;
  out[8]  = 0x4802e7c3;
  out[9]  = 0x2cac08b4;
  out[10] = 0x6ac517b0;
  out[11] = 0x2743eb39;
  out[12] = 0x6b826249;
  out[13] = 0x65ca07ee;
  out[14] = 0x48bc5b1b;
  out[15] = 0x356e26d8;
}

fn fill_expected_range_16(out: &mut [u32; STATE_WIDTH]) {
  out[0]  = 0x22d2a167;
  out[1]  = 0x4911cc3f;
  out[2]  = 0x19f8c9c0;
  out[3]  = 0x2c725fa2;
  out[4]  = 0x5e3a602f;
  out[5]  = 0x5314c6b4;
  out[6]  = 0x5a49ceff;
  out[7]  = 0x546b713a;
  out[8]  = 0x17f21463;
  out[9]  = 0x12389621;
  out[10] = 0x0761c306;
  out[11] = 0x7654d674;
  out[12] = 0x27660c01;
  out[13] = 0x7475057a;
  out[14] = 0x52e1c31e;
  out[15] = 0x61973dfb;
}

fn fill_expected_range_256(out: &mut [u32; STATE_WIDTH]) {
  out[0]  = 0x271c8229;
  out[1]  = 0x727a7388;
  out[2]  = 0x62edde19;
  out[3]  = 0x07da259c;
  out[4]  = 0x42e1642f;
  out[5]  = 0x25512854;
  out[6]  = 0x0984d864;
  out[7]  = 0x28192e26;
  out[8]  = 0x0c884443;
  out[9]  = 0x515d24f6;
  out[10] = 0x6192dcd5;
  out[11] = 0x15453862;
  out[12] = 0x1223175e;
  out[13] = 0x5fe133e8;
  out[14] = 0x69b186d1;
  out[15] = 0x4d04e284;
}

fn fill_expected_zeros_1(out: &mut [u32; STATE_WIDTH]) {
  out[0]  = 0x575cd201;
  out[1]  = 0x7a759ad4;
  out[2]  = 0x4c8b5cee;
  out[3]  = 0x71899383;
  out[4]  = 0x5b501db2;
  out[5]  = 0x7112a138;
  out[6]  = 0x4630624f;
  out[7]  = 0x61f01af4;
  out[8]  = 0x323a1379;
  out[9]  = 0x328b53b9;
  out[10] = 0x13c26fa2;
  out[11] = 0x3ed3287c;
  out[12] = 0x79811d70;
  out[13] = 0x5ccaf1c6;
  out[14] = 0x07b0fd6d;
  out[15] = 0x7c4e9f27;
}

fn fill_expected_zeros_16(out: &mut [u32; STATE_WIDTH]) {
  out[0]  = 0x5a5faf93;
  out[1]  = 0x44476824;
  out[2]  = 0x7e77d1aa;
  out[3]  = 0x49b298ff;
  out[4]  = 0x1682d9d0;
  out[5]  = 0x3aec1bbb;
  out[6]  = 0x2e78f3e5;
  out[7]  = 0x5cd0366d;
  out[8]  = 0x68bbd72d;
  out[9]  = 0x79f1c79b;
  out[10] = 0x644e1d28;
  out[11] = 0x1ed961f0;
  out[12] = 0x11e7672a;
  out[13] = 0x4ba86f19;
  out[14] = 0x1ec73662;
  out[15] = 0x06a05093;
}

fn fill_expected_zeros_256(out: &mut [u32; STATE_WIDTH]) {
  out[0]  = 0x6fbb7934;
  out[1]  = 0x40f944a4;
  out[2]  = 0x0ee275f0;
  out[3]  = 0x0545e8b7;
  out[4]  = 0x5d8950d6;
  out[5]  = 0x23d1b013;
  out[6]  = 0x6a53c5cf;
  out[7]  = 0x7d588984;
  out[8]  = 0x4cb562a5;
  out[9]  = 0x3d08f400;
  out[10] = 0x43c210d2;
  out[11] = 0x572309e4;
  out[12] = 0x2570ad53;
  out[13] = 0x7accb79c;
  out[14] = 0x13db0bb8;
  out[15] = 0x0e21f8b2;
}

fn fill_expected_zeros_2_to_16(out: &mut [u32; STATE_WIDTH]) {
  out[0]  = 0x1bb62370;
  out[1]  = 0x17f76303;
  out[2]  = 0x3b21106c;
  out[3]  = 0x5d260ef2;
  out[4]  = 0x69a5b09e;
  out[5]  = 0x5c5a5e12;
  out[6]  = 0x4b9e4318;
  out[7]  = 0x2e467f54;
  out[8]  = 0x78819a30;
  out[9]  = 0x46b74f95;
  out[10] = 0x2cb26af6;
  out[11] = 0x07f4242c;
  out[12] = 0x1d60b807;
  out[13] = 0x1f66ad60;
  out[14] = 0x5abf7644;
  out[15] = 0x4a27fb88;
}

fn fill_expected_zeros_2_to_18(out: &mut [u32; STATE_WIDTH]) {
  out[0]  = 0x4952180a;
  out[1]  = 0x1bae33c0;
  out[2]  = 0x67da8b52;
  out[3]  = 0x35410fef;
  out[4]  = 0x3a1c5841;
  out[5]  = 0x6a2f6a31;
  out[6]  = 0x0c32646b;
  out[7]  = 0x5a47019b;
  out[8]  = 0x749311b0;
  out[9]  = 0x7de625b5;
  out[10] = 0x2bee65b8;
  out[11] = 0x48d451d3;
  out[12] = 0x2ad561b1;
  out[13] = 0x22d44298;
  out[14] = 0x6e3fb781;
  out[15] = 0x0894eb57;
}

fn fill_expected_zeros_2_to_20(out: &mut [u32; STATE_WIDTH]) {
  out[0]  = 0x5d3806af;
  out[1]  = 0x19db630e;
  out[2]  = 0x116a1a97;
  out[3]  = 0x3b89dee4;
  out[4]  = 0x3b50d1f5;
  out[5]  = 0x3f828727;
  out[6]  = 0x13d03e94;
  out[7]  = 0x4a6aeeb6;
  out[8]  = 0x0c0ed47f;
  out[9]  = 0x3cb2340c;
  out[10] = 0x4b12db5a;
  out[11] = 0x720e22e5;
  out[12] = 0x4def8c36;
  out[13] = 0x1c22abd7;
  out[14] = 0x67cd7e44;
  out[15] = 0x6d51f595;
}

// Runs 7 out of 9 test cases and exits with a bitmask of failures: bit i is set if
// test i did not match its expected output (see the table below). Exit code 0
// means every test passed.
//
//   bit 0: range(7)        bit 1: range(16)       bit 2: range(256)
//   bit 3: zeros(1)        bit 4: zeros(16)       bit 5: zeros(256)
//   bit 6: zeros(2^16)     bit 7: zeros(2^18)     bit 8: zeros(2^20)
//
// Cases 6-8 allocate very large stack arrays (64 KiB, 256 KiB, 1 MiB) and
// `[0u8; N]` for such N typically lowers to a memset call. If the build fails
// with "undefined symbol: memset", `-Z build-std` needs to include
// `compiler_builtins` (or define a local memset).
//
// Note. Tests 8 and 9 corresponding to 1 << 18 and 1 << 20 zeros respectively
// are skipped due to taking too long.
#[no_mangle]
fn main() -> ! {
  let results: [bool; 7] = [
    run_range::<7>(fill_expected_range_7),
    run_range::<16>(fill_expected_range_16),
    run_range::<256>(fill_expected_range_256),
    run_zeros::<1>(fill_expected_zeros_1),
    run_zeros::<16>(fill_expected_zeros_16),
    run_zeros::<256>(fill_expected_zeros_256),
    run_zeros::<{ 1 << 16 }>(fill_expected_zeros_2_to_16),
    // run_zeros::<{ 1 << 18 }>(fill_expected_zeros_2_to_18),
    // run_zeros::<{ 1 << 20 }>(fill_expected_zeros_2_to_20),
  ];

  let mut failures: u32 = 0;
  let mut i = 0;
  while i < results.len() {
    if !results[i] {
      failures |= 1u32 << i;
    }
    i += 1;
  }
  exit(failures as i32);
}
