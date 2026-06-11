const builtin = @import("builtin");
const std = @import("std");
const verifier_ray = @import("verifier_ray");
const bench_data = @import("vanishing_bench_data");
const bench_options = @import("vanishing_bench_options");

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
    vanishing.verify(selected_case.system, input) catch exitR5(1);
    exitR5(0);
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
