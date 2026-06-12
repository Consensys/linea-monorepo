const builtin = @import("builtin");
const std = @import("std");
const verifier_ray = @import("verifier_ray");
const bench_data = @import("vanishing_bench_data");
const bench_options = @import("vanishing_bench_options");

const protocol = verifier_ray.protocol;
const vanishing = verifier_ray.query.vanishing;

const is_r5_zkvm = builtin.target.cpu.arch == .riscv64 and builtin.target.os.tag == .freestanding;
const selected_case = if (bench_options.invalid)
    bench_data.getInvalid(bench_options.case_index)
else
    bench_data.get(bench_options.case_index);

pub fn main() noreturn {
    if (comptime !is_r5_zkvm) {
        @compileError("vanishing benchmark guest is currently only wired for the R5 zkVM target");
    }
    r5_main();
}

pub export fn r5_main() noreturn {
    if (comptime !is_r5_zkvm) {
        @compileError("vanishing benchmark guest is currently only wired for the R5 zkVM target");
    }

    bench_data.keepRuntime(bench_options.case_index, bench_options.invalid);
    var input = selected_case.input;
    std.mem.doNotOptimizeAway(&input);

    markR5(1);
    var replay_stats = protocol.replayWithStats(selected_case.spec, input.ctx.rounds) catch exitR5(1);
    var all_coins = replay_stats.coins;
    std.mem.doNotOptimizeAway(&all_coins);
    std.mem.doNotOptimizeAway(&replay_stats.compression_count);

    markR5Value(2, replay_stats.compression_count);
    var replayed_input = vanishing.CheckInput{
        .ctx = .{
            .all_coins = &all_coins,
            .rounds = input.ctx.rounds,
        },
        .witness_claims = input.witness_claims,
        .quotient_claims = input.quotient_claims,
        .module_sizes = input.module_sizes,
    };
    std.mem.doNotOptimizeAway(&replayed_input);
    vanishing.verify(selected_case.system, replayed_input) catch exitR5(1);

    markR5(3);
    exitR5(0);
}

fn markR5(comptime phase: u64) void {
    markR5Value(phase, 0);
}

fn markR5Value(comptime phase: u64, value: usize) void {
    asm volatile (
        \\mv a0, %[phase]
        \\mv a1, %[value]
        \\li a7, 4242
        \\ecall
        :
        : [phase] "r" (phase),
          [value] "r" (value),
        : .{ .a0 = true, .a1 = true, .a7 = true, .memory = true });
}

fn exitR5(code: u8) noreturn {
    asm volatile (
        \\mv a0, %[code]
        \\li a7, 93
        \\ecall
        :
        : [code] "r" (code),
    );
    unreachable;
}
