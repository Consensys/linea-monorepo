// https://github.com/eth-act/zkvm-standards/blob/282cd356c3a0498416bb0619f9c8a347ce9933fb/standards/c-interface-accelerators/zkvm_accelerators.h#L42
pub const zkvm_status = enum(c_int) {
    ZKVM_EOK = 0, // Success
    ZKVM_EFAIL = -1, // Failure
};

pub fn exit(code: u32) noreturn {
    // no OS to return to, signal halt via ecall
    asm volatile (
        \\mv a0, %[code]
        \\li a7, 93
        \\ecall
        :
        : [code] "r" (code),
    );
    unreachable;
}

pub fn panic() noreturn {
    exit(1);
}
