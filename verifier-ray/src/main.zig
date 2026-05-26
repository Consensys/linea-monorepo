const builtin = @import("builtin");
const verifier_ray = @import("verifier_ray");

const field = verifier_ray.field.koalabear;
const polynomial = verifier_ray.pcs.polynomial;
const poseidon2 = verifier_ray.crypto.poseidon2;
const Transcript = verifier_ray.crypto.fiat_shamir.Transcript;

const is_r5_zkvm = builtin.target.cpu.arch == .riscv64 and builtin.target.os.tag == .freestanding;
const is_native_linux_x86_64 = builtin.target.cpu.arch == .x86_64 and builtin.target.os.tag == .linux;

const Commitment = verifier_ray.proof.Commitment;
const Digest = poseidon2.Digest;

const commitment_count = 1;
const commitment_limb_count = 8;
const public_input_count = 2;
const proof_byte_count = 3;
const coefficient_count = 4;
const digest_limb_count = 8;

const native_input_path: [:0]const u8 = "zig-out/input.bin";

extern const _input_start: u8;

const SmokeInput = struct {
    commitments: [commitment_count]Commitment,
    public_inputs: [public_input_count]field.Element,
    proof_bytes: [proof_byte_count]u8,
    coefficients: [coefficient_count]field.Element,
    point: field.Element,
    expected_challenge: Digest,
};

const raw_input_len = @sizeOf(SmokeInput);

pub fn main() noreturn {
    if (comptime is_r5_zkvm) {
        // this entry point should only be called from native build (`make build` or `make build-release`)
        unreachable;
    }
    if (comptime !is_native_linux_x86_64) {
        @compileError("native verifier syscall path currently supports x86_64-linux only");
    }

    const fd = nativeSyscall3(sys_open, @intFromPtr(native_input_path.ptr), o_rdonly, 0);
    if (isNativeSyscallError(fd)) exitNative(1);

    const mapped_addr = nativeSyscall6(sys_mmap, 0, raw_input_len, prot_read, map_private, fd, 0);
    if (isNativeSyscallError(mapped_addr)) exitNative(1);

    const input: *const SmokeInput = @ptrCast(@alignCast(@as([*]const u8, @ptrFromInt(mapped_addr))));
    exitNative(runVerifierSmoke(input));
}

pub export fn r5_main() noreturn {
    if (comptime !is_r5_zkvm) {
        // this entry point should only be called from R5 zkVM build (`make build-r5` or `make build-r5-release`)
        unreachable;
    }
    const input: *const SmokeInput = @ptrCast(@alignCast(&_input_start));
    const res = runVerifierSmoke(input);
    exitR5(res);
}

fn runVerifierSmoke(input: *const SmokeInput) u8 {
    if (!exerciseTemporaryTraceWork(input)) {
        return 1;
    }

    verifier_ray.verify(.{
        .commitments = input.commitments[0..],
        .public_inputs = input.public_inputs[0..],
        .proof_bytes = input.proof_bytes[0..],
    }) catch |err| switch (err) {
        verifier_ray.VerifyError.Unsupported => {},
        else => return 1,
    };

    return 0;
}

fn exerciseTemporaryTraceWork(input: *const SmokeInput) bool {
    // Temporary zkVM trace exercise. Remove this once main verifies a realistic proof.
    const evaluation = polynomial.evaluateBaseCanonical(input.coefficients[0..], input.point);

    var transcript = Transcript.init();
    transcript.updateElements(input.coefficients[0..]);
    transcript.updateElement(evaluation);
    const challenge = transcript.randomField();

    return digestEql(challenge, input.expected_challenge);
}

fn digestEql(lhs: Digest, rhs: Digest) bool {
    for (lhs, rhs) |lhs_limb, rhs_limb| {
        if (!lhs_limb.eql(rhs_limb)) return false;
    }
    return true;
}

const sys_open: u64 = 2;
const sys_mmap: u64 = 9;
const sys_exit: u64 = 60;
const o_rdonly: u64 = 0;
const prot_read: u64 = 1;
const map_private: u64 = 2;

fn isNativeSyscallError(result: u64) bool {
    const signed: i64 = @bitCast(result);
    return signed < 0 and signed > -4096;
}

fn nativeSyscall1(number: u64, arg1: u64) u64 {
    return asm volatile ("syscall"
        : [ret] "={rax}" (-> u64),
        : [number] "{rax}" (number),
          [arg1] "{rdi}" (arg1),
        : .{ .rcx = true, .r11 = true, .memory = true });
}

fn nativeSyscall3(number: u64, arg1: u64, arg2: u64, arg3: u64) u64 {
    return asm volatile ("syscall"
        : [ret] "={rax}" (-> u64),
        : [number] "{rax}" (number),
          [arg1] "{rdi}" (arg1),
          [arg2] "{rsi}" (arg2),
          [arg3] "{rdx}" (arg3),
        : .{ .rcx = true, .r11 = true, .memory = true });
}

fn nativeSyscall6(number: u64, arg1: u64, arg2: u64, arg3: u64, arg4: u64, arg5: u64, arg6: u64) u64 {
    return asm volatile ("syscall"
        : [ret] "={rax}" (-> u64),
        : [number] "{rax}" (number),
          [arg1] "{rdi}" (arg1),
          [arg2] "{rsi}" (arg2),
          [arg3] "{rdx}" (arg3),
          [arg4] "{r10}" (arg4),
          [arg5] "{r8}" (arg5),
          [arg6] "{r9}" (arg6),
        : .{ .rcx = true, .r11 = true, .memory = true });
}

fn exitNative(code: u8) noreturn {
    if (comptime !is_native_linux_x86_64) {
        @compileError("native verifier syscall exit currently supports x86_64-linux only");
    }

    _ = nativeSyscall1(sys_exit, @intCast(code));
    unreachable;
}

fn exitR5(code: u8) noreturn {
    if (comptime !is_r5_zkvm) {
        @compileError("R5 exit currently supports only R5 zkVM target");
    }
    switch (code) {
        0 => exitR5Success(),
        else => exitR5Failure(),
    }
}

fn exitR5Success() noreturn {
    asm volatile (
        \\li a0, 0
        \\li a7, 93
        \\ecall
    );
    unreachable;
}

fn exitR5Failure() noreturn {
    asm volatile (
        \\li a0, 1
        \\li a7, 93
        \\ecall
    );
    unreachable;
}
