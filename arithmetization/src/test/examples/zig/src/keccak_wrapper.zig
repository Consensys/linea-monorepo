export fn main() noreturn {
    const input_offset: usize = 0x1000;
    const input_size: usize = 64;
    const output_offset: usize = 0x2000;
    in_line_assembly_keccak_call(input_offset, input_size, output_offset);

    // no OS to return to, signal halt via ecall
    asm volatile (
        \\li a0, 0   # exit code 0
        \\li a7, 93  # syscall number for exit
        \\ecall
    );
    unreachable;
}

fn in_line_assembly_keccak_call(
    input_offset: usize,
    input_size: usize,
    output_offset: usize,
) void {
    // invoke custom opcode for keccak
    // opcode format: opcode(0x0c = custom-1) | funct3(0b000) | funct7(0b0000000) | rd(output_offset) | rs1(input_offset) | rs2(input_size)
    asm volatile (
        \\.insn r 0x0c, 0b000, 0b0000000, %[out], %[in], %[size]
        :
        : [out] "r" (output_offset),
          [in] "r" (input_offset),
          [size] "r" (input_size),
    );
}
