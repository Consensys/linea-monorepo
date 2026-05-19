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
            .cpu_model = .{ .explicit = &std.Target.riscv.cpu.generic_rv64 },
            .cpu_features_add = std.Target.riscv.featureSet(&.{.m}), // Zicclsm extension does not affect the generated ELF so it is omitted
            .cpu_features_sub = std.Target.riscv.featureSet(&.{ .a, .c, .d, .f, .zicsr, .zaamo, .zalrsc }),
            .os_tag = .freestanding,
            .abi = .none, // LP64 (soft-float) is relevant only for float numbers, which we do not use, so it can be omitted
        },
    });

    // Optimize for binary size by default; can be overridden with -Doptimize=
    const optimize = b.standardOptimizeOption(.{
        .preferred_optimize_mode = .ReleaseSmall,
    });

    const name = b.option([]const u8, "name", "Name of the program (source: src/<name>.zig, binary: <name>)") orelse @panic("'-Dname=<name>' is required");
    const strip = b.option(bool, "strip", "Whether to strip the binary (default: false)") orelse false;

    const source = b.fmt("src/{s}.zig", .{name});
    const uses_zevm = std.mem.eql(u8, name, "rollup_zevm_guest");

    const exe = b.addExecutable(.{
        .name = name,
        .root_module = b.createModule(.{
            .root_source_file = b.path(source),
            .target = target,
            .optimize = optimize,
            .strip = strip, // Removes debug symbols and other metadata
        }),
    });

    // Point to assembly overwriting default SP
    exe.root_module.addAssemblyFile(b.path("src/start.s"));

    if (uses_zevm) {
        const exe_crypto = if (target.result.os.tag == .freestanding) null else resolveNativeCrypto(b, target);
        const zevm_imports = createZevmImports(b, target, optimize, exe_crypto);
        addZevmImports(exe.root_module, zevm_imports);
        if (exe_crypto) |crypto| {
            linkNativeZevmCrypto(b, exe, target, crypto);
        }
        if (target.result.os.tag == .freestanding) {
            clearFreestandingNativeLinkage(b, exe.root_module);
        }
    }

    // Remove unused code sections
    exe.link_gc_sections = true;

    b.installArtifact(exe);

    const test_step = b.step("test", "Run native Zig unit tests for the selected example");
    const write_fixtures_step = b.step("write-fixtures", "Regenerate native Zig guest fixture files");
    if (uses_zevm) {
        const native_target = b.resolveTargetQuery(.{});
        const native_crypto = resolveNativeCrypto(b, native_target);
        const zevm_imports = createZevmImports(b, native_target, optimize, native_crypto);

        const guest_mod = b.createModule(.{
            .root_source_file = b.path(source),
            .target = native_target,
            .optimize = optimize,
        });
        addZevmImports(guest_mod, zevm_imports);

        const fixtures_mod = b.createModule(.{
            .root_source_file = b.path("src/rollup_zevm_guest_fixtures.zig"),
            .target = native_target,
            .optimize = optimize,
        });
        addZevmImports(fixtures_mod, zevm_imports);
        const fixture_data_mod = b.createModule(.{
            .root_source_file = b.path("fixtures/rollup_zevm_guest_fixture_data.zig"),
            .target = native_target,
            .optimize = optimize,
        });
        fixtures_mod.addImport("rollup_zevm_guest_fixture_data", fixture_data_mod);

        const tests = b.addTest(.{
            .root_module = b.createModule(.{
                .root_source_file = b.path("src/rollup_zevm_guest_test.zig"),
                .target = native_target,
                .optimize = optimize,
            }),
        });
        addZevmImports(tests.root_module, zevm_imports);
        tests.root_module.addImport("rollup_zevm_guest", guest_mod);
        tests.root_module.addImport("rollup_zevm_guest_fixtures", fixtures_mod);
        linkNativeZevmCrypto(b, tests, native_target, native_crypto);

        const run_tests = b.addRunArtifact(tests);
        test_step.dependOn(&run_tests.step);

        const fixture_writer = b.addExecutable(.{
            .name = "rollup_zevm_guest_fixture_writer",
            .root_module = b.createModule(.{
                .root_source_file = b.path("src/rollup_zevm_guest_fixture_writer.zig"),
                .target = native_target,
                .optimize = optimize,
            }),
        });
        addZevmImports(fixture_writer.root_module, zevm_imports);
        fixture_writer.root_module.addImport("rollup_zevm_guest_fixtures", fixtures_mod);
        const fixture_writer_options = b.addOptions();
        fixture_writer_options.addOption(
            []const u8,
            "output_path",
            b.pathFromRoot("fixtures/contract_creation_then_ecrecover.json"),
        );
        fixture_writer.root_module.addImport("fixture_writer_options", fixture_writer_options.createModule());
        linkNativeZevmCrypto(b, fixture_writer, native_target, native_crypto);

        const run_fixture_writer = b.addRunArtifact(fixture_writer);
        write_fixtures_step.dependOn(&run_fixture_writer.step);
    } else {
        const tests = b.addTest(.{
            .root_module = b.createModule(.{
                .root_source_file = b.path(source),
                .target = b.resolveTargetQuery(.{}),
                .optimize = optimize,
            }),
        });

        const run_tests = b.addRunArtifact(tests);
        test_step.dependOn(&run_tests.step);
    }
}

const ZevmImports = struct {
    zevm_stateless: *std.Build.Module,
    zevm_stateless_rlp: *std.Build.Module,
    rollup_guest_allocator: *std.Build.Module,
    zevm_precompile: ?*std.Build.Module,
};

fn createZevmImports(
    b: *std.Build,
    target: std.Build.ResolvedTarget,
    optimize: std.builtin.OptimizeMode,
    native_crypto: ?NativeCrypto,
) ZevmImports {
    const zevm_stateless_dep = b.dependency("zevm_stateless", .{
        .target = target,
        .optimize = optimize,
    });
    const is_guest_target = target.result.os.tag == .freestanding;

    const guest_allocator_mod = b.createModule(.{
        .root_source_file = b.path("src/rollup_guest_allocator.zig"),
        .target = target,
        .optimize = optimize,
    });

    const guest_secp256k1_mod = b.createModule(.{
        .root_source_file = b.path("src/rollup_guest_secp256k1.zig"),
        .target = target,
        .optimize = optimize,
    });

    const executor_mod = zevm_stateless_dep.module("executor");
    const transition_mod = zevmImport(executor_mod, &.{"executor_transition"});
    const precompile_mod = zevmImport(transition_mod, &.{"precompile"});

    transition_mod.addImport("executor_allocator", guest_allocator_mod);
    if (is_guest_target) {
        const guest_precompile_impls_mod = b.createModule(.{
            .root_source_file = b.path("src/rollup_guest_precompile_implementations.zig"),
            .target = target,
            .optimize = optimize,
        });
        guest_precompile_impls_mod.addImport(
            "precompile_types",
            zevmImport(precompile_mod, &.{"precompile_types"}),
        );

        transition_mod.addImport("secp256k1_wrapper", guest_secp256k1_mod);
        precompile_mod.addImport("precompile_implementations", guest_precompile_impls_mod);
    } else {
        const crypto = native_crypto orelse @panic("missing native crypto configuration");

        const native_secp256k1_impl_mod = zevmImport(transition_mod, &.{"secp256k1_wrapper"});
        native_secp256k1_impl_mod.addIncludePath(.{
            .cwd_relative = crypto.include_path,
        });

        const native_secp256k1_mod = b.createModule(.{
            .root_source_file = b.path("src/rollup_native_secp256k1.zig"),
            .target = target,
            .optimize = optimize,
        });
        native_secp256k1_mod.addImport("zevm_native_secp256k1", native_secp256k1_impl_mod);
        transition_mod.addImport("secp256k1_wrapper", native_secp256k1_mod);

        const native_precompile_impls_mod = configureNativePrecompileImplementations(
            b,
            precompile_mod,
            guest_allocator_mod,
            crypto,
        );
        const native_precompile_wrapper_mod = b.createModule(.{
            .root_source_file = b.path("src/rollup_native_precompile_implementations.zig"),
            .target = target,
            .optimize = optimize,
        });
        native_precompile_wrapper_mod.addImport("zevm_native_precompile_implementations", native_precompile_impls_mod);
        precompile_mod.addImport("precompile_implementations", native_precompile_wrapper_mod);
    }

    addGuestAllocatorImport(transition_mod, guest_allocator_mod);
    addGuestAllocatorImport(zevmImport(transition_mod, &.{"bytecode"}), guest_allocator_mod);
    addGuestAllocatorImport(zevmImport(transition_mod, &.{"state"}), guest_allocator_mod);
    addGuestAllocatorImport(zevmImport(transition_mod, &.{"context"}), guest_allocator_mod);
    addGuestAllocatorImport(zevmImport(transition_mod, &.{"handler"}), guest_allocator_mod);
    addGuestAllocatorImport(zevmImport(transition_mod, &.{"precompile"}), guest_allocator_mod);
    addGuestAllocatorImport(zevmImport(transition_mod, &.{ "handler", "interpreter" }), guest_allocator_mod);

    const stateless_rlp_mod = b.createModule(.{
        .root_source_file = zevm_stateless_dep.path("src/stateless/io.zig"),
        .target = target,
        .optimize = optimize,
    });
    stateless_rlp_mod.addImport("input", zevm_stateless_dep.module("input"));
    stateless_rlp_mod.addImport("rlp_decode", zevm_stateless_dep.module("rlp_decode"));
    stateless_rlp_mod.addImport("mpt", zevm_stateless_dep.module("mpt"));

    return .{
        .zevm_stateless = zevm_stateless_dep.module("zevm_stateless"),
        .zevm_stateless_rlp = stateless_rlp_mod,
        .rollup_guest_allocator = guest_allocator_mod,
        .zevm_precompile = if (is_guest_target) null else precompile_mod,
    };
}

fn addZevmImports(module: *std.Build.Module, imports: ZevmImports) void {
    module.addImport("zevm_stateless", imports.zevm_stateless);
    module.addImport("zevm_stateless_rlp", imports.zevm_stateless_rlp);
    module.addImport("rollup_guest_allocator", imports.rollup_guest_allocator);
    if (imports.zevm_precompile) |precompile_mod| module.addImport("zevm_precompile", precompile_mod);
}

fn zevmImport(module: *std.Build.Module, names: []const []const u8) *std.Build.Module {
    var current = module;
    for (names) |name| {
        current = current.import_table.get(name) orelse @panic("missing ZEVM module import");
    }
    return current;
}

fn addGuestAllocatorImport(module: *std.Build.Module, guest_allocator_mod: *std.Build.Module) void {
    module.addImport("zevm_allocator", guest_allocator_mod);
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

fn configureNativePrecompileImplementations(
    b: *std.Build,
    precompile_mod: *std.Build.Module,
    guest_allocator_mod: *std.Build.Module,
    crypto: NativeCrypto,
) *std.Build.Module {
    const native_options = b.addOptions();
    native_options.addOption(bool, "enable_blst", crypto.enable_blst);
    native_options.addOption(bool, "enable_mcl", crypto.enable_mcl);
    native_options.addOption(bool, "enable_secp256k1", crypto.enable_secp256k1);
    native_options.addOption(bool, "enable_openssl", crypto.enable_openssl);

    const native_impls_mod = zevmImport(precompile_mod, &.{"precompile_implementations"});
    native_impls_mod.addImport("build_options", native_options.createModule());
    native_impls_mod.addImport("zevm_allocator", guest_allocator_mod);
    native_impls_mod.addIncludePath(.{ .cwd_relative = crypto.include_path });
    return native_impls_mod;
}

const NativeCrypto = struct {
    include_path: []const u8,
    lib_path: []const u8,
    enable_blst: bool,
    enable_mcl: bool,
    enable_secp256k1: bool,
    enable_openssl: bool,
};

fn resolveNativeCrypto(b: *std.Build, target: std.Build.ResolvedTarget) NativeCrypto {
    const default_prefix = if (b.graph.host.result.os.tag == .linux) "/usr/local" else "/opt/homebrew";
    const prefix = b.option([]const u8, "crypto-prefix", "Native crypto dependency prefix") orelse default_prefix;
    const include_path = b.fmt("{s}/include", .{prefix});
    const lib_path = b.fmt("{s}/lib", .{prefix});

    if (target.result.os.tag == .freestanding) {
        return .{
            .include_path = include_path,
            .lib_path = lib_path,
            .enable_blst = false,
            .enable_mcl = false,
            .enable_secp256k1 = false,
            .enable_openssl = false,
        };
    }

    const has_blst = fileExists(b.fmt("{s}/blst.h", .{include_path})) and
        anyFileExists(&.{
            b.fmt("{s}/libblst.a", .{lib_path}),
            b.fmt("{s}/libblst.dylib", .{lib_path}),
            b.fmt("{s}/libblst.so", .{lib_path}),
        });
    const has_mcl = fileExists(b.fmt("{s}/mcl/bn.h", .{include_path})) and
        anyFileExists(&.{
            b.fmt("{s}/libmcl.a", .{lib_path}),
            b.fmt("{s}/libmcl.dylib", .{lib_path}),
            b.fmt("{s}/libmcl.so", .{lib_path}),
        });
    const has_secp256k1 = fileExists(b.fmt("{s}/secp256k1.h", .{include_path})) and
        anyFileExists(&.{
            b.fmt("{s}/libsecp256k1.a", .{lib_path}),
            b.fmt("{s}/libsecp256k1.dylib", .{lib_path}),
            b.fmt("{s}/libsecp256k1.so", .{lib_path}),
        });
    const has_openssl = fileExists(b.fmt("{s}/openssl/ec.h", .{include_path})) and
        anyFileExists(&.{
            b.fmt("{s}/libssl.a", .{lib_path}),
            b.fmt("{s}/libssl.dylib", .{lib_path}),
            b.fmt("{s}/libssl.so", .{lib_path}),
        }) and
        anyFileExists(&.{
            b.fmt("{s}/libcrypto.a", .{lib_path}),
            b.fmt("{s}/libcrypto.dylib", .{lib_path}),
            b.fmt("{s}/libcrypto.so", .{lib_path}),
        });

    return .{
        .include_path = include_path,
        .lib_path = lib_path,
        .enable_blst = b.option(bool, "blst", "Enable native blst-backed precompiles") orelse has_blst,
        .enable_mcl = b.option(bool, "mcl", "Enable native mcl-backed precompiles") orelse has_mcl,
        .enable_secp256k1 = b.option(bool, "secp256k1", "Enable native secp256k1-backed precompiles") orelse has_secp256k1,
        .enable_openssl = b.option(bool, "openssl", "Enable native OpenSSL-backed precompiles") orelse has_openssl,
    };
}

fn fileExists(path: []const u8) bool {
    if (@hasDecl(std.fs, "openFileAbsolute")) {
        const file = std.fs.openFileAbsolute(path, .{}) catch return false;
        file.close();
    } else {
        var file = std.Io.Dir.openFileAbsolute(std.Options.debug_io, path, .{}) catch return false;
        file.close(std.Options.debug_io);
    }
    return true;
}

fn anyFileExists(paths: []const []const u8) bool {
    for (paths) |path| {
        if (fileExists(path)) return true;
    }
    return false;
}

fn linkNativeZevmCrypto(
    b: *std.Build,
    step: *std.Build.Step.Compile,
    target: std.Build.ResolvedTarget,
    crypto: NativeCrypto,
) void {
    if (target.result.os.tag == .freestanding) return;

    addCompileIncludePath(step, .{ .cwd_relative = crypto.include_path });
    addCompileLibraryPath(step, .{ .cwd_relative = crypto.lib_path });

    if (crypto.enable_secp256k1) {
        linkCompileSystemLibrary(step, "secp256k1");
    }
    if (crypto.enable_openssl) {
        linkCompileSystemLibrary(step, "ssl");
        linkCompileSystemLibrary(step, "crypto");
    }
    linkCompileSystemLibrary(step, "c");
    if (target.result.os.tag != .windows) {
        linkCompileSystemLibrary(step, "m");
    }
    if (crypto.enable_blst) {
        const libblst = b.fmt("{s}/libblst.a", .{crypto.lib_path});
        if (fileExists(libblst)) {
            addCompileObjectFile(step, .{ .cwd_relative = libblst });
        } else {
            linkCompileSystemLibrary(step, "blst");
        }
    }
    if (crypto.enable_mcl) {
        const libmcl = b.fmt("{s}/libmcl.a", .{crypto.lib_path});
        if (target.result.os.tag != .macos) {
            linkCompileSystemLibrary(step, "stdc++");
        }
        if (fileExists(libmcl) and target.result.os.tag == .macos) {
            addCompileObjectFile(step, .{ .cwd_relative = libmcl });
            linkCompileLibCpp(step);
        } else {
            linkCompileSystemLibrary(step, "mcl");
        }
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

fn addCompileObjectFile(step: *std.Build.Step.Compile, path: std.Build.LazyPath) void {
    if (@hasDecl(std.Build.Step.Compile, "addObjectFile")) {
        step.addObjectFile(path);
    } else {
        step.root_module.addObjectFile(path);
    }
}

fn linkCompileSystemLibrary(step: *std.Build.Step.Compile, name: []const u8) void {
    if (@hasDecl(std.Build.Step.Compile, "linkSystemLibrary")) {
        step.linkSystemLibrary(name);
    } else {
        step.root_module.linkSystemLibrary(name, .{});
    }
}

fn linkCompileLibCpp(step: *std.Build.Step.Compile) void {
    if (@hasDecl(std.Build.Step.Compile, "linkLibCpp")) {
        step.linkLibCpp();
    } else {
        step.root_module.linkSystemLibrary("c++", .{});
    }
}
