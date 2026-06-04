const std = @import("std");
const builtin = @import("builtin");

const required_zig_version = "0.16.0-dev.3153+d6f43caad";

pub fn build(b: *std.Build) void {
    if (!std.mem.eql(u8, builtin.zig_version_string, required_zig_version)) {
        std.debug.print(
            "riscv-guests requires Zig {s}; found {s}\n",
            .{ required_zig_version, builtin.zig_version_string },
        );
        @panic("unsupported Zig version");
    }

    const target = b.standardTargetOptions(.{
        .default_target = .{
            .cpu_arch = .riscv64,
            .cpu_model = .{ .explicit = &std.Target.riscv.cpu.generic_rv64 },
            .cpu_features_add = std.Target.riscv.featureSet(&.{.m}),
            .cpu_features_sub = std.Target.riscv.featureSet(&.{ .a, .c, .d, .f, .zicsr, .zaamo, .zalrsc }),
            .os_tag = .freestanding,
            .abi = .none,
        },
    });

    const optimize = b.standardOptimizeOption(.{
        .preferred_optimize_mode = .ReleaseSmall,
    });
    const input_offset = b.option([]const u8, "input-offset", "Input memory origin") orelse "0x08800000";
    const input_offset_value: usize = @intCast(parseAddress("input-offset", input_offset));
    const guest_options = b.addOptions();
    guest_options.addOption(usize, "input_offset", input_offset_value);
    const guest_options_mod = guest_options.createModule();

    const guest_name = "evm_execution_guest";
    const source = "l2-execution/src/evm_execution_guest.zig";

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
    clearFreestandingNativeLinkage(b, obj.root_module);
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

    // The SSZ fixture comes from the execution-spec-tests zkevm dependency (lazy: only fetched when
    // this test is built). An empty-block vector → no transactions → no secp256k1/curve precompiles,
    // so the full native crypto backend is linked but only keccak is exercised.
    const fixture_rel = "blockchain_tests/for_amsterdam/amsterdam/eip7928_block_level_access_lists/block_access_lists/bal_empty_block_no_coinbase.json";
    if (b.lazyDependency("execution_spec_tests_zkevm", .{})) |fixtures_dep| {
        const fixtures_mod = b.createModule(.{
            .root_source_file = b.path("l2-execution/src/evm_execution_fixtures.zig"),
            .target = native_target,
            .optimize = optimize,
        });
        // Embed the chosen fixture straight from the dependency tree (no committed copy).
        fixtures_mod.addAnonymousImport("zkevm_stateless_block.json", .{
            .root_source_file = fixtures_dep.path(fixture_rel),
        });

        const tests = b.addTest(.{
            .root_module = b.createModule(.{
                .root_source_file = b.path("l2-execution/src/evm_execution_guest_test.zig"),
                .target = native_target,
                .optimize = optimize,
            }),
        });
        tests.root_module.addImport("evm_execution_guest", guest_mod);
        tests.root_module.addImport("evm_execution_fixtures", fixtures_mod);
        linkNativeZesuCrypto(tests, native_target, native_crypto);

        test_step.dependOn(&b.addRunArtifact(tests).step);
    }
}

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

fn parseAddress(comptime name: []const u8, value: []const u8) u64 {
    return std.fmt.parseInt(u64, value, 0) catch @panic("invalid -D" ++ name ++ " value");
}

fn clearFreestandingNativeLinkage(b: *std.Build, module: *std.Build.Module) void {
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

// Links the full native crypto backing zesu's default.zig accelerator: secp256k1 (ecrecover),
// OpenSSL (P-256), blst (BLS12-381 + KZG), mcl (BN254). No-op for freestanding targets, whose
// crypto is the extern zkvm_* bridge. zesu sets the @cImport include path on its accel module, so
// here we only need the library search path + the libraries.
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
