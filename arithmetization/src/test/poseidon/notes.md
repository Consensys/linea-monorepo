- Rust or Zig program `poseidon.rs`
- that populates memory with inputs, either as
  - hardcoded (nonempty) inputs
  - alternatively that takes (nonempty) inputs from a file
- example: in_bytes = [0x00, 0x01, ..., 0xff]
- add an in-line assembly custom-0 instruction to run poseidon with
  - special opcode
  - special type (maybe R-type)
  - and funct3/funct7 coding for the Poseidon1 precompile

Question: which registers will hold the pointer to the data + its size ? Also: where do you write the output ?

All of these could be in rs1, rs2, rd [which is ok if we go with R-type]

```rust
poseidon(io :u64, is: u64, ro: u64) {
  unsafe {
    core::arch::asm!(
      // funct3 = 42, funct7 = 69
      // r-type isntruction, 0x0b ≡ custom-0, 42 = funct3, 69 = funct7,
      // {0} = register address holding io
      // {1} = register address holding is
      // {2} = register address holding ro
      // order is decided by the declaration order of the in(reg) XXX
      ":insn r 0x0b, 42, 69, {2}, {0}, {1}",
      in(reg) io,
      in(reg) is,
      in(reg) ro,
    );
  }
}

// à la
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
```

According to Claude:

```
Custom-0 through custom-3 instructions in RISC-V don't have a fixed instruction type — the spec reserves the opcodes but leaves the encoding format up to the implementer. You can use any of the standard formats (R, I, S, etc.) or define your own, as long as bits [1:0] are 11 (indicating a 32-bit instruction).

That said, R-type is the most common choice for custom instructions since it gives you two source registers and one destination register with no implicit memory semantics.
```
