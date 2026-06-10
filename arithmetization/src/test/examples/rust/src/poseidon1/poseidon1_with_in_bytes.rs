#![no_std]
#![no_main]

const STATE_WIDTH: usize = 16;
const BYTES_PER_FELT: usize = 4;
const STATE_LEN: usize = STATE_WIDTH * BYTES_PER_FELT;

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
fn compare(output_buf: &[u8], expected: &[u8]) -> bool {
  assert!(output_buf.len() == STATE_LEN, "incorrect output size");
  assert!(expected.len() == STATE_LEN, "incorrect expected result size");
  let mut i = 0;
  while i < STATE_LEN {
    if output_buf[i] != expected[i] {
      return false;
    }
    i += 1;
  }
  return true
}

// run_inputs computes the Poseidon hash of the input vector
fn run_inputs(inputs: &[u8], expected: &[u8]) -> bool {
  let mut output_buf = [0u8; STATE_LEN];
  in_line_assembly_poseidon_call(
    inputs.as_ptr() as usize,
    inputs.len(),
    output_buf.as_mut_ptr() as usize,
  );

  compare(&output_buf, &expected)
}

#[no_mangle]
fn main() -> ! {

  let input_len = get_test_vector_input_len();
  let (input, expected) = get_test_vector(input_len);

  let result :bool = run_inputs(input, expected);

  if result {
    exit(0);
  } else {
    exit(1);
  }
}

// inputs must be structured like so
//
//    [ input_len ] || [ output_state ] || [       inputs       ]
//    <---- 4 ---->    <---- 16*4 ---->    <--- inpu_len -- … -->
//
// with input_len being a 4 byte integer, output_state being 16 KoalaBear field elements
// each occupying 4 bytes, and inputs being an arbirary _nonempty_ slice of bytes.
fn get_test_vector(input_len :usize) -> (&'static [u8], &'static [u8]) {
  unsafe {
    let base = &raw const _input_start;
    let expected = core::slice::from_raw_parts(base.add(4), STATE_LEN);
    let inputs = core::slice::from_raw_parts(base.add(4 + STATE_LEN), input_len);
    (inputs, expected)
  }
}

fn get_test_vector_input_len() -> usize {
  static mut BUF: [u8; 4] = [0u8; 4];
  unsafe {
    read_memory(&raw mut BUF as *mut u8, 4);
    u32::from_le_bytes(BUF) as usize
  }
}

