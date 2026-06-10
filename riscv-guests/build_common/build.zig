//! Shared build helpers for the riscv-guests packages — deliberately only the **guest-agnostic**
//! bits: the pinned toolchain, the common freestanding rv64im target (every guest targets the same
//! ZkC profile), freestanding-linkage hygiene, and small utilities.
//!
//! Anything tied to a specific workload stays in the guest that uses it. In particular the zesu /
//! EVM-execution wiring (executor + SSZ module imports, native crypto linking) lives in
//! l2-execution's build.zig — the rollup and rollup-aggregation guests do NOT run the EVM (they do
//! KZG/compression and recursive proof verification), so that wiring is not common.
//!
//! Each guest pulls this in via a `build_common` path dependency in its build.zig.zon plus
//! `const common = @import("build_common");`. Never built on its own — `build()` is a no-op.

const std = @import("std");
const builtin = @import("builtin");

/// The one Zig toolchain pinned for all guests (kept in sync with riscv-guests/.zigversion).
pub const required_zig_version = "0.16.0-dev.3153+d6f43caad";

/// Abort unless the active Zig matches the pinned toolchain.
pub fn requireZigVersion() void {
    if (!std.mem.eql(u8, builtin.zig_version_string, required_zig_version)) {
        std.debug.print(
            "riscv-guests requires Zig {s}; found {s}\n",
            .{ required_zig_version, builtin.zig_version_string },
        );
        @panic("unsupported Zig version");
    }
}

/// The freestanding rv64im target every guest builds for (the Linea ZkC interpreter profile:
/// base RV64I + M, soft-float, no A/C/D/F/Zicsr). Overridable on the CLI like any standard target.
pub fn standardGuestTarget(b: *std.Build) std.Build.ResolvedTarget {
    return b.standardTargetOptions(.{
        .default_target = .{
            .cpu_arch = .riscv64,
            .cpu_model = .{ .explicit = &std.Target.riscv.cpu.generic_rv64 },
            .cpu_features_add = std.Target.riscv.featureSet(&.{.m}),
            .cpu_features_sub = std.Target.riscv.featureSet(&.{ .a, .c, .d, .f, .zicsr, .zaamo, .zalrsc }),
            .os_tag = .freestanding,
            .abi = .none,
        },
    });
}

pub fn parseAddress(comptime name: []const u8, value: []const u8) u64 {
    return std.fmt.parseInt(u64, value, 0) catch @panic("invalid -D" ++ name ++ " value");
}

/// Clear native libc/libcpp linkage across a module graph — freestanding rv64im guests must not
/// link a host libc, but imported modules may default to it.
pub fn clearFreestandingNativeLinkage(b: *std.Build, module: *std.Build.Module) void {
    var seen = std.AutoHashMap(*std.Build.Module, void).init(b.allocator);
    clearFreestandingNativeLinkageInner(&seen, module);
}

fn clearFreestandingNativeLinkageInner(
    seen: *std.AutoHashMap(*std.Build.Module, void),
    module: *std.Build.Module,
) void {
    if (seen.get(module) != null) return;
    seen.put(module, {}) catch @panic("OOM");

    if (@hasField(std.Build.Module, "link_libc")) {
        module.link_libc = false;
    }
    if (@hasField(std.Build.Module, "link_libcpp")) {
        module.link_libcpp = false;
    }

    var imports = module.import_table.iterator();
    while (imports.next()) |entry| {
        clearFreestandingNativeLinkageInner(seen, entry.value_ptr.*);
    }
}

/// build_common is consumed via `@import`, never built directly.
pub fn build(b: *std.Build) void {
    _ = b;
}
