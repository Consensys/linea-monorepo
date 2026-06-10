const std = @import("std");
const common = @import("build_common");

pub fn build(b: *std.Build) void {
    common.requireZigVersion();

    // All guests target the same freestanding rv64im ZkC profile (shared helper).
    const target = common.standardGuestTarget(b);

    const optimize = b.standardOptimizeOption(.{
        .preferred_optimize_mode = .ReleaseSmall,
    });
    const input_offset = b.option([]const u8, "input-offset", "Input memory origin") orelse "0x08800000";
    const input_offset_value: usize = @intCast(common.parseAddress("input-offset", input_offset));
    const guest_options = b.addOptions();
    guest_options.addOption(usize, "input_offset", input_offset_value);
    const guest_options_mod = guest_options.createModule();

    const guest_name = "evm_execution_guest";
    const source = "src/evm_execution_guest.zig";

    // ── Guest relocatable object ──────────────────────────────────────────────
    // Built the same way zesu builds zesu.rv64im.o: a relocatable rv64im object with all execution
    // logic compiled in, but the zkvm-standards crypto accelerators (zkvm_*) left as UNRESOLVED
    // extern references. zesu's freestanding accel backend (extern_bridge.zig) only DECLARES zkvm_*;
    // the proving system (Zkc) is the consumer that resolves/intercepts them. So there is no in-guest
    // software provider, no final link, and no entry/linker script here — the prover toolchain links
    // this object, supplies zkvm_* (and _start), and runs it.
    // NOTE: this object is NOT a runnable ELF on its own; it cannot execute until the prover links it.
    const zesu_guest = b.dependency("zesu", .{ .target = target, .optimize = optimize });
    const obj = b.addObject(.{
        .name = guest_name,
        .root_module = b.createModule(.{
            .root_source_file = b.path(source),
            .target = target,
            .optimize = optimize,
        }),
    });
    obj.root_module.code_model = .medium;
    addExecutionImports(obj.root_module, zesuImports(zesu_guest));
    obj.root_module.addImport("guest_options", guest_options_mod);
    // Delegate the in-guest precompile implementations to zesu-zkvm's stdlibs_accel (consumed by
    // zkvm_provide.zig's `.native` arms). It is std-only, so it builds for freestanding rv64im.
    const zesu_zkvm = b.dependency("zesu_zkvm", .{});
    const accel_src = zesu_zkvm.path("linea/src/runtime/stdlibs_accel.zig");
    obj.root_module.addImport("zesu_zkvm_accel", b.createModule(.{
        .root_source_file = accel_src,
        .target = target,
        .optimize = optimize,
    }));
    common.clearFreestandingNativeLinkage(b, obj.root_module);
    const install_obj = b.addInstallFile(obj.getEmittedBin(), b.fmt("lib/{s}.o", .{guest_name}));
    b.getInstallStep().dependOn(&install_obj.step);

    // ── Native test ───────────────────────────────────────────────────────────
    // Runs the thin wrapper (vanilla zesu stateless execution) on the host against a real
    // execution-spec-tests zkevm SSZ fixture, asserting the serialized validation result matches —
    // the same end-to-end check as zesu's zkevm-blockchain-test-runner. Links zesu's full native
    // crypto backend (default.zig); linea adds the library search path so it links on macOS. The
    // committed fixture is an empty block (only keccak), but the full backend is linked so the suite
    // can grow to tx-bearing fixtures (ecrecover/curves) without further build changes.
    const native_target = b.resolveTargetQuery(.{});
    const native_crypto = resolveNativeCrypto(b, native_target);
    const zesu_native = b.dependency("zesu", .{ .target = native_target, .optimize = optimize });
    const native_imports = zesuImports(zesu_native);

    const guest_mod = b.createModule(.{
        .root_source_file = b.path(source),
        .target = native_target,
        .optimize = optimize,
    });
    addExecutionImports(guest_mod, native_imports);
    guest_mod.addImport("guest_options", guest_options_mod);

    const test_step = b.step("test", "Run native Zig unit tests for the EVM execution guest");
    const spec_step = b.step("spec-tests", "Run the guest against all EF zkevm stateless fixtures (host)");

    // Integration smoke test for the delegated precompiles: verifies zesu-zkvm's stdlibs_accel
    // imports and that its ecrecover round-trips (the in-guest precompiles delegate to it). std +
    // the dependency only — no fixtures, no native crypto libs.
    const accel_tests = b.addTest(.{
        .root_module = b.createModule(.{
            .root_source_file = b.path("src/stdlibs_accel_test.zig"),
            .target = native_target,
            .optimize = optimize,
        }),
    });
    accel_tests.root_module.addImport("zesu_zkvm_accel", b.createModule(.{
        .root_source_file = accel_src,
        .target = native_target,
        .optimize = optimize,
    }));
    test_step.dependOn(&b.addRunArtifact(accel_tests).step);

    // The SSZ fixture comes from the execution-spec-tests zkevm dependency (lazy: only fetched when
    // this test is built). An empty-block vector → no transactions → no secp256k1/curve precompiles,
    // so the full native crypto backend is linked but only keccak is exercised.
    const fixture_rel = "blockchain_tests/for_amsterdam/amsterdam/eip7928_block_level_access_lists/block_access_lists/bal_empty_block_no_coinbase.json";
    if (b.lazyDependency("execution_spec_tests_zkevm", .{})) |fixtures_dep| {
        const fixtures_mod = b.createModule(.{
            .root_source_file = b.path("src/evm_execution_fixtures.zig"),
            .target = native_target,
            .optimize = optimize,
        });
        // Embed the chosen fixture straight from the dependency tree (no committed copy).
        fixtures_mod.addAnonymousImport("zkevm_stateless_block.json", .{
            .root_source_file = fixtures_dep.path(fixture_rel),
        });

        const tests = b.addTest(.{
            .root_module = b.createModule(.{
                .root_source_file = b.path("src/evm_execution_guest_test.zig"),
                .target = native_target,
                .optimize = optimize,
            }),
        });
        tests.root_module.addImport("evm_execution_guest", guest_mod);
        tests.root_module.addImport("evm_execution_fixtures", fixtures_mod);
        linkNativeZesuCrypto(tests, native_target, native_crypto);

        test_step.dependOn(&b.addRunArtifact(tests).step);

        // ── Spec-test runner ────────────────────────────────────────────────────
        // Standalone host executable that walks the WHOLE zkevm fixture tree and runs every block
        // through this guest (mirrors zesu's zkevm-blockchain-test-runner). Fixtures come from the
        // same lazy dependency — no curl, no embedding; `zig build spec-tests` passes the
        // blockchain_tests/ directory as --fixtures. Pass-through extra args after `--`, e.g.
        // `zig build spec-tests -- --fork Amsterdam -x`.
        const spec_runner_exe = b.addExecutable(.{
            .name = "evm-execution-spec-runner",
            .root_module = b.createModule(.{
                .root_source_file = b.path("src/evm_spec_runner.zig"),
                .target = native_target,
                .optimize = optimize,
            }),
        });
        spec_runner_exe.root_module.addImport("evm_execution_guest", guest_mod);
        linkNativeZesuCrypto(spec_runner_exe, native_target, native_crypto);

        const run_spec = b.addRunArtifact(spec_runner_exe);
        run_spec.addArg("--fixtures");
        run_spec.addDirectoryArg(fixtures_dep.path("blockchain_tests"));
        if (b.args) |extra| run_spec.addArgs(extra);
        spec_step.dependOn(&run_spec.step);
    }
}

// ── zesu / EVM-execution wiring (l2-execution-specific) ───────────────────────────────────────────
// These are NOT in build_common: only guests that run the EVM via zesu need them. The rollup and
// rollup-aggregation guests (KZG/compression + recursive proof verification) do not.

const ZesuImports = struct {
    allocator: *std.Build.Module,
    executor: *std.Build.Module,
    ssz_decode: *std.Build.Module,
    ssz_output: *std.Build.Module,
};

/// Pull zesu's exposed modules by name. Backends are selected inside zesu by target:
/// freestanding → extern zkvm_* bridge; native → default.zig (full crypto).
fn zesuImports(zesu: *std.Build.Dependency) ZesuImports {
    return .{
        .allocator = zesu.module("zesu_allocator"),
        .executor = zesu.module("executor"),
        .ssz_decode = zesu.module("ssz_decode"),
        .ssz_output = zesu.module("ssz_output"),
    };
}

fn addExecutionImports(module: *std.Build.Module, imports: ZesuImports) void {
    module.addImport("zesu_allocator", imports.allocator);
    module.addImport("zesu_executor", imports.executor);
    module.addImport("zesu_ssz_decode", imports.ssz_decode);
    module.addImport("zesu_ssz_output", imports.ssz_output);
}

const NativeCrypto = struct {
    include_path: []const u8,
    lib_path: []const u8,
    blst_path: []const u8,
    mcl_path: []const u8,
    is_linux: bool,
};

fn resolveNativeCrypto(b: *std.Build, target: std.Build.ResolvedTarget) NativeCrypto {
    _ = target;
    const default_prefix = if (b.graph.host.result.os.tag == .linux) "/usr/local" else "/opt/homebrew";
    const prefix = b.option([]const u8, "crypto-prefix", "Native crypto dependency prefix") orelse default_prefix;
    const lib_path = b.fmt("{s}/lib", .{prefix});
    return .{
        .include_path = b.fmt("{s}/include", .{prefix}),
        .lib_path = lib_path,
        .blst_path = b.fmt("{s}/libblst.a", .{lib_path}),
        .mcl_path = b.fmt("{s}/libmcl.a", .{lib_path}),
        .is_linux = b.graph.host.result.os.tag == .linux,
    };
}

/// Links the full native crypto backing zesu's default.zig accelerator: secp256k1 (ecrecover),
/// OpenSSL (P-256), blst (BLS12-381 + KZG), mcl (BN254). No-op for freestanding targets, whose
/// crypto is the extern zkvm_* bridge. zesu sets the @cImport include path on its accel module, so
/// here we only need the library search path + the libraries.
fn linkNativeZesuCrypto(
    step: *std.Build.Step.Compile,
    target: std.Build.ResolvedTarget,
    crypto: NativeCrypto,
) void {
    if (target.result.os.tag == .freestanding) return;

    addCompileIncludePath(step, .{ .cwd_relative = crypto.include_path });
    addCompileLibraryPath(step, .{ .cwd_relative = crypto.lib_path });

    linkCompileSystemLibrary(step, "c");
    if (target.result.os.tag != .windows) {
        linkCompileSystemLibrary(step, "m");
    }
    linkCompileSystemLibrary(step, "secp256k1");
    linkCompileSystemLibrary(step, "ssl");
    linkCompileSystemLibrary(step, "crypto");
    step.root_module.addObjectFile(.{ .cwd_relative = crypto.blst_path });
    if (crypto.is_linux) {
        linkCompileSystemLibrary(step, "mcl");
    } else {
        step.root_module.addObjectFile(.{ .cwd_relative = crypto.mcl_path });
        step.root_module.link_libcpp = true;
    }
}

fn addCompileIncludePath(step: *std.Build.Step.Compile, path: std.Build.LazyPath) void {
    if (@hasDecl(std.Build.Step.Compile, "addIncludePath")) {
        step.addIncludePath(path);
    } else {
        step.root_module.addIncludePath(path);
    }
}

fn addCompileLibraryPath(step: *std.Build.Step.Compile, path: std.Build.LazyPath) void {
    if (@hasDecl(std.Build.Step.Compile, "addLibraryPath")) {
        step.addLibraryPath(path);
    } else {
        step.root_module.addLibraryPath(path);
    }
}

fn linkCompileSystemLibrary(step: *std.Build.Step.Compile, name: []const u8) void {
    if (@hasDecl(std.Build.Step.Compile, "linkSystemLibrary")) {
        step.linkSystemLibrary(name);
    } else {
        step.root_module.linkSystemLibrary(name, .{});
    }
}
