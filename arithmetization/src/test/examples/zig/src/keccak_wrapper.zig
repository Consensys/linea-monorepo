const custom_std = @import("custom_std.zig");

export fn main() noreturn {

    const buf_0 = [_]u8{0} ** 0;
    const buf_32 = [_]u8{0} ** 32;
    const buf_64 = [_]u8{0} ** 64;

    const data_0: [*c]const u8 = &buf_0;
    const data_32: [*c]const u8 = &buf_32;
    const data_64: [*c]const u8 = &buf_64;

    const output: [*c]zkvm_keccak256_hash = @ptrFromInt(0x08000000);

    _ = zkvm_keccak256(data_0, 0, output); // empty keccak
    _ = zkvm_keccak256(data_32, 32, output); // keccak of "00".repeat(32)
    _ = zkvm_keccak256(data_64, 64, output); // keccak of "00".repeat(64)

    // hashes are discarded; for reference:
    // keccak( <empty> )         = c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470
    // keccak( "00".repeat(32) ) = 290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e563
    // keccak( "00".repeat(64) ) = ad3228b676f7d3cd4284a5443f17f1962b36e491b30a40b2405849e597ba5fb5

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

// 64: 0000000000000000000000000000000000000000000000000000000000000000
// 32: 00000000000000000000000000000000
