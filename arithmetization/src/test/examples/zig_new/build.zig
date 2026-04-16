const std = @import("std");

pub fn build(b: *std.Build) void {
    // Default to a strict RV64IM target (base integer + multiply/divide only):
    //   - baseline_rv64: RV64I base ISA
    //   - adds M extension: integer multiply/divide
    //   - strips A (atomics), C (compressed), D (double float), F (single float),
    //     Zicsr (CSR instructions), Zaamo/Zalrsc (atomic sub-extensions)
    const target = b.standardTargetOptions(.{
        .default_target = .{
            .cpu_arch = .riscv64,
            .cpu_model = .{ .explicit = &std.Target.riscv.cpu.baseline_rv64 },
            .cpu_features_add = std.Target.riscv.featureSet(&.{.m}),
            .cpu_features_sub = std.Target.riscv.featureSet(&.{ .a, .c, .d, .f, .zicsr, .zaamo, .zalrsc }),
            .os_tag = .linux,
            .abi = .musl,
        },
    });

    const optimize = b.standardOptimizeOption(.{});

    const exe = b.addExecutable(.{
        .name = "arithmetic_test",
        .root_module = b.createModule(.{
            .root_source_file = b.path("src/arithmetic_test.zig"),
            .target = target,
            .optimize = optimize,
        }),
    });

    b.installArtifact(exe);
}
