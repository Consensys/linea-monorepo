const std = @import("std");
const builtin = @import("builtin");

const guest_options = @import("guest_options");
const executor = @import("zesu_executor");
const ssz_decode = @import("zesu_ssz_decode");
const ssz_output = @import("zesu_ssz_output");
const zesu_allocator = @import("zesu_allocator");

const GUEST_INPUT_OFFSET: usize = guest_options.input_offset;
const GUEST_HEAP_OFFSET: usize = 0x50000000;
const GUEST_HEAP_SIZE: usize = 256 * 1024 * 1024;

// This guest is a thin wrapper over zesu's vanilla stateless execution: it decodes an SSZ-encoded
// StatelessInput, executes the block, and serializes the SSZ validation result — the same pipeline
// as zesu's `runner.runStateless` / `zkevm-blockchain-test-runner`.
//
// The crypto accelerators (zkvm_*) are intentionally left UNRESOLVED in the freestanding object:
// zesu's extern_bridge only declares them, and the proving system (Zkc) resolves/intercepts them at
// link/run time. There is deliberately no in-guest software provider — see build.zig (addObject).

/// Result of running one SSZ-encoded StatelessInput:
///   `out`     — the 105-byte SSZ SszStatelessValidationResult
///   `success` — successful_validation: execution succeeded AND the computed post-state and
///               receipts roots match the values claimed in the payload.
pub const Result = struct {
    out: [105]u8,
    success: bool,
};

/// Vanilla zesu stateless block execution. Fed an explicit byte slice so it runs identically on the
/// native host (tests) and from the zkVM guest entry below.
pub fn runStateless(allocator: std.mem.Allocator, ssz_input: []const u8) !Result {
    zesu_allocator.set(allocator);

    const si = try ssz_decode.decode(allocator, ssz_input);
    const ep = &si.new_payload_request.execution_payload;

    const success = blk: {
        const proof = executor.executeStatelessInput(allocator, si, si.chain_config.fork_name) catch break :blk false;
        if (!std.mem.eql(u8, &proof.post_state_root, &ep.state_root)) break :blk false;
        if (!std.mem.eql(u8, &proof.receipts_root, &ep.receipts_root)) break :blk false;
        break :blk true;
    };

    const out = try ssz_output.serialize(allocator, si.new_payload_request, si.chain_config.chain_id, success);
    return .{ .out = out, .success = success };
}

/// zkVM guest entry. The host maps the input at GUEST_INPUT_OFFSET as
/// [u64 big-endian length][SSZ StatelessInput bytes]; we execute it and exit 0 on
/// successful_validation, 1 otherwise.
fn guestMain() callconv(.c) noreturn {
    const heap = @as([*]u8, @ptrFromInt(GUEST_HEAP_OFFSET))[0..GUEST_HEAP_SIZE];
    var fba = std.heap.FixedBufferAllocator.init(heap);
    const allocator = fba.allocator();

    const base = @as([*]const u8, @ptrFromInt(GUEST_INPUT_OFFSET));
    const len: usize = @intCast(std.mem.readInt(u64, base[0..8], .big));
    const ssz_input = base[8..][0..len];

    const result = runStateless(allocator, ssz_input) catch exit(1);
    exit(if (result.success) 0 else 1);
}

comptime {
    if (!builtin.is_test) {
        @export(&guestMain, .{ .name = "main" });
    }
}

fn exit(code: u64) noreturn {
    if (builtin.cpu.arch == .riscv64) {
        asm volatile (
            \\mv a0, %[code]
            \\li a7, 93
            \\ecall
            :
            : [code] "r" (code),
            : .{ .x10 = true, .x17 = true });
        unreachable;
    }

    std.debug.panic("guest exit({d})", .{code});
}
