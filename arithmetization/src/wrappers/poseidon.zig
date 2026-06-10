const custom_std = @import("custom_std.zig");

// RATE = 15 field elements
// general KOALABEAR elements hold 4 field elements
pub const zkvm_poseidon_hash = extern struct {
    data: [60]u8 align(8),
};

// Poseidon (1) in-line assembly call
// Poseidon IS NOT part of the standardized zkvm interface
pub fn zkvmPoseidon(data: [*c]const u8, len: usize, output: [*c]zkvm_poseidon_hash) custom_std.zkvm_status {
    if (data == null or output == null or len == 0) {
        custom_std.panic();
    }

    // invoke custom opcode for poseidon
    // opcode format: opcode(0x0b = custom-0) | funct3(0b111) | funct7(0b1111111) | rd(output_offset) | rs1(input_offset) | rs2(input_size)
    asm volatile (
        \\.insn r 0x0b, 0b111, 0b1111111, %[out], %[in], %[size]
        :
        : [out] "r" (@intFromPtr(output)),
          [in] "r" (@intFromPtr(data)),
          [size] "r" (len),
    );
    return .ZKVM_EOK;
}
