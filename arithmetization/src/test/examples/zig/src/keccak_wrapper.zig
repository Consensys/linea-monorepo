const custom_std = @import("custom_std.zig");

export fn main() noreturn {
    // buf_* variables represent all-zeros inputs
    const buf_0 = [_]u8{0} ** 0;
    const buf_32 = [_]u8{0} ** 32;
    const buf_64 = [_]u8{0} ** 64;
    const buf_135 = [_]u8{0} ** 135;
    const buf_136 = [_]u8{0} ** 136;
    const buf_137 = [_]u8{0} ** 137;

    // extract pointers to the inputs
    const data_0: [*c]const u8 = &buf_0;
    const data_32: [*c]const u8 = &buf_32;
    const data_64: [*c]const u8 = &buf_64;
    const data_135: [*c]const u8 = &buf_135;
    const data_136: [*c]const u8 = &buf_136;
    const data_137: [*c]const u8 = &buf_137;

    // pointer for writing output
    const output: [*c]zkvm_keccak256_hash = @ptrFromInt(0x08000000);

    _ = zkvm_keccak256(data_0, 0, output); // empty keccak
    _ = zkvm_keccak256(data_32, 32, output); // keccak of "00".repeat(32)
    _ = zkvm_keccak256(data_64, 64, output); // keccak of "00".repeat(64)
    _ = zkvm_keccak256(data_135, 135, output); // keccak of "00".repeat(135)
    _ = zkvm_keccak256(data_136, 136, output); // keccak of "00".repeat(136)
    _ = zkvm_keccak256(data_137, 137, output); // keccak of "00".repeat(137)


    // inputs:
    // 32: 0000000000000000000000000000000000000000000000000000000000000000
    // 64: 00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
    // 135: 000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
    // 136: 00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
    // 137: 0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
    //
    // hashes are discarded; for reference:
    // keccak( <empty> )          = c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470
    // keccak( "00".repeat(32) )  = 290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e563
    // keccak( "00".repeat(64) )  = ad3228b676f7d3cd4284a5443f17f1962b36e491b30a40b2405849e597ba5fb5
    // keccak( "00".repeat(135) ) = 29e3704feeca7fb9ba229f0fa04d9b36449cf3ad6e1d85d9cfff3a10df9abc3e
    // keccak( "00".repeat(136) ) = 3a5912a7c5faa06ee4fe906253e339467a9ce87d533c65be3c15cb231cdb25f9
    // keccak( "00".repeat(137) ) = bee7fbb405cb0d91a8775e338c4a5e4b5d6b2d051f687fa942043cffdc73bd28

    custom_std.exit(0);
}

// https://github.com/eth-act/zkvm-standards/blob/282cd356c3a0498416bb0619f9c8a347ce9933fb/standards/c-interface-accelerators/zkvm_accelerators.h#L42
pub const zkvm_status = enum(c_int) {
    ZKVM_EOK = 0, // Success
    ZKVM_EFAIL = -1, // Failure
};

// https://github.com/eth-act/zkvm-standards/blob/282cd356c3a0498416bb0619f9c8a347ce9933fb/standards/c-interface-accelerators/zkvm_accelerators.h#L72
pub const zkvm_keccak256_hash = extern struct {
    data: [32]u8 align(8),
};

// https://github.com/eth-act/zkvm-standards/blob/282cd356c3a0498416bb0619f9c8a347ce9933fb/standards/c-interface-accelerators/zkvm_accelerators.h#L166
export fn zkvm_keccak256(data: [*c]const u8, len: usize, output: [*c]zkvm_keccak256_hash) zkvm_status {
    if (data == null or output == null) {
        custom_std.panic();
    }

    // invoke custom opcode for keccak
    // opcode format: opcode(0x0c = custom-1) | funct3(0b000) | funct7(0b0000000) | rd(output_offset) | rs1(input_offset) | rs2(input_size)
    asm volatile (
        \\.insn r 0x0c, 0b000, 0b0000000, %[out], %[in], %[size]
        :
        : [out] "r" (@intFromPtr(output)),
          [in] "r" (@intFromPtr(data)),
          [size] "r" (len),
    );
    return .ZKVM_EOK;
}
