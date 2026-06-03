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
    const program_offset = b.option([]const u8, "program-offset", "Program memory origin") orelse "0x00000000";
    const input_offset = b.option([]const u8, "input-offset", "Input memory origin") orelse "0x08800000";
    const sp = b.option([]const u8, "sp", "Initial stack pointer") orelse "0x08800000";
    const input_offset_value: usize = @intCast(parseAddress("input-offset", input_offset));
    const stack_origin = stackOrigin(b, sp);
    const guest_options = b.addOptions();
    guest_options.addOption(usize, "input_offset", input_offset_value);
    const guest_options_mod = guest_options.createModule();
    const linker_script = b.addWriteFiles().add("linker_script.ld", b.fmt(
        \\ENTRY(_start)
        \\
        \\MEMORY {{
        \\    PROGRAM (rx) : ORIGIN = {s}, LENGTH = 0x8000000
        \\    STACK   (rw) : ORIGIN = {s}, LENGTH = 0x800000
        \\    IN      (r)  : ORIGIN = {s}, LENGTH = 0x40000000
        \\}}
        \\
        \\SECTIONS {{
        \\    . = ORIGIN(PROGRAM);
        \\
        \\    .text : {{
        \\        *(.text .text.*)
        \\    }} > PROGRAM
        \\
        \\    .rodata : {{
        \\        *(.rodata .rodata.*)
        \\    }} > PROGRAM
        \\
        \\    .data : {{
        \\        *(.data .data.*)
        \\    }} > PROGRAM
        \\
        \\    .bss : {{
        \\        *(.bss .bss.*)
        \\    }} > PROGRAM
        \\
        \\    _stack_start = ORIGIN(STACK) + LENGTH(STACK);
        \\    _input_start = ORIGIN(IN);
        \\}}
        \\
    , .{ program_offset, stack_origin, input_offset }));

    const guest_name = "evm_execution_guest";
    const source = "l2-execution/src/evm_execution_guest.zig";

    const exe = b.addExecutable(.{
        .name = guest_name,
        .root_module = b.createModule(.{
            .root_source_file = b.path(source),
            .target = target,
            .optimize = optimize,
        }),
    });

    exe.root_module.addAssemblyFile(b.path("l2-execution/src/start.s"));
    exe.setLinkerScript(linker_script);

    const guest_imports = createZesuImports(b, target, optimize, .guest, null);
    addExecutionImports(exe.root_module, guest_imports);
    exe.root_module.addImport("guest_options", guest_options_mod);
    clearFreestandingNativeLinkage(b, exe.root_module);
    exe.link_gc_sections = true;
    b.installArtifact(exe);

    const test_step = b.step("test", "Run native Zig unit tests for the EVM execution guest");
    const write_fixtures_step = b.step("write-fixtures", "Regenerate native Zig guest fixture files");
    const native_target = b.resolveTargetQuery(.{});
    const native_crypto = resolveNativeCrypto(b, native_target);
    const zesu_imports = createZesuImports(b, native_target, optimize, .native, native_crypto);

    const guest_mod = b.createModule(.{
        .root_source_file = b.path(source),
        .target = native_target,
        .optimize = optimize,
    });
    addExecutionImports(guest_mod, zesu_imports);
    guest_mod.addImport("guest_options", guest_options_mod);

    const fixtures_mod = b.createModule(.{
        .root_source_file = b.path("l2-execution/src/evm_execution_fixtures.zig"),
        .target = native_target,
        .optimize = optimize,
    });
    addExecutionImports(fixtures_mod, zesu_imports);
    const fixture_data_mod = b.createModule(.{
        .root_source_file = b.path("l2-execution/fixtures/evm_execution_fixture_data.zig"),
        .target = native_target,
        .optimize = optimize,
    });
    fixtures_mod.addImport("evm_execution_fixture_data", fixture_data_mod);

    const tests = b.addTest(.{
        .root_module = b.createModule(.{
            .root_source_file = b.path("l2-execution/src/evm_execution_guest_test.zig"),
            .target = native_target,
            .optimize = optimize,
        }),
    });
    addExecutionImports(tests.root_module, zesu_imports);
    tests.root_module.addImport("evm_execution_guest", guest_mod);
    tests.root_module.addImport("evm_execution_fixtures", fixtures_mod);
    linkNativeZesuCrypto(tests, native_target, native_crypto);

    const run_tests = b.addRunArtifact(tests);
    test_step.dependOn(&run_tests.step);

    const fixture_writer = b.addExecutable(.{
        .name = "evm_execution_fixture_writer",
        .root_module = b.createModule(.{
            .root_source_file = b.path("l2-execution/src/evm_execution_fixture_writer.zig"),
            .target = native_target,
            .optimize = optimize,
        }),
    });
    addExecutionImports(fixture_writer.root_module, zesu_imports);
    fixture_writer.root_module.addImport("evm_execution_fixtures", fixtures_mod);
    const fixture_writer_options = b.addOptions();
    fixture_writer_options.addOption(
        []const u8,
        "output_path",
        b.pathFromRoot("l2-execution/fixtures/contract_creation_then_ecrecover.json"),
    );
    fixture_writer.root_module.addImport("fixture_writer_options", fixture_writer_options.createModule());
    linkNativeZesuCrypto(fixture_writer, native_target, native_crypto);

    const run_fixture_writer = b.addRunArtifact(fixture_writer);
    write_fixtures_step.dependOn(&run_fixture_writer.step);
}

const BuildFlavor = enum {
    guest,
    native,
};

const ZesuImports = struct {
    allocator: *std.Build.Module,
    executor: *std.Build.Module,
    input: *std.Build.Module,
    mpt: *std.Build.Module,
    rlp_decode: *std.Build.Module,
    precompile: *std.Build.Module,
};

fn stackOrigin(b: *std.Build, sp: []const u8) []const u8 {
    const sp_value = parseAddress("sp", sp);
    if (sp_value < 0x800000) @panic("-Dsp must be at least 0x800000");
    return b.fmt("0x{x:0>8}", .{sp_value - 0x800000});
}

fn parseAddress(comptime name: []const u8, value: []const u8) u64 {
    return std.fmt.parseInt(u64, value, 0) catch @panic("invalid -D" ++ name ++ " value");
}

fn createZesuImports(
    b: *std.Build,
    target: std.Build.ResolvedTarget,
    optimize: std.builtin.OptimizeMode,
    flavor: BuildFlavor,
    native_crypto: ?NativeCrypto,
) ZesuImports {
    const zesu = b.dependency("zesu", .{
        .target = target,
        .optimize = optimize,
    });

    const allocator_module = b.createModule(.{
        .root_source_file = b.path("l2-execution/src/evm_guest_allocator.zig"),
        .target = target,
        .optimize = optimize,
    });

    const primitives_module = b.createModule(.{
        .root_source_file = zesu.path("src/evm/primitives/main.zig"),
        .target = target,
        .optimize = optimize,
    });

    // Zesu routes precompile crypto through one injected accel_impl module:
    //
    //   precompile/default_impls.zig -> accelerators.zig -> accel_impl
    //
    // Keep that shape here so native tests can use local host implementations
    // while the RISC-V guest keeps stable zkvm_* hook points for future
    // interpreter/circuit interception.
    const accel_impl_module = b.createModule(.{
        .root_source_file = b.path(switch (flavor) {
            .guest => "l2-execution/src/zesu_guest_accelerators.zig",
            .native => "l2-execution/src/zesu_native_accelerators.zig",
        }),
        .target = target,
        .optimize = optimize,
    });
    if (native_crypto) |crypto| {
        accel_impl_module.addIncludePath(.{ .cwd_relative = crypto.include_path });
    }

    const accelerators_module = b.createModule(.{
        .root_source_file = zesu.path("src/crypto/accelerators.zig"),
        .target = target,
        .optimize = optimize,
    });
    accelerators_module.addImport("accel_impl", accel_impl_module);

    const precompile_types_module = b.createModule(.{
        .root_source_file = zesu.path("src/evm/precompile/types.zig"),
        .target = target,
        .optimize = optimize,
    });

    const bytecode_module = b.createModule(.{
        .root_source_file = zesu.path("src/evm/bytecode/main.zig"),
        .target = target,
        .optimize = optimize,
    });
    bytecode_module.addImport("primitives", primitives_module);
    bytecode_module.addImport("zesu_allocator", allocator_module);
    bytecode_module.addImport("accelerators", accelerators_module);

    const state_module = b.createModule(.{
        .root_source_file = zesu.path("src/evm/state/main.zig"),
        .target = target,
        .optimize = optimize,
    });
    state_module.addImport("primitives", primitives_module);
    state_module.addImport("bytecode", bytecode_module);
    state_module.addImport("zesu_allocator", allocator_module);

    const database_module = b.createModule(.{
        .root_source_file = zesu.path("src/evm/database/main.zig"),
        .target = target,
        .optimize = optimize,
    });
    database_module.addImport("primitives", primitives_module);
    database_module.addImport("state", state_module);
    database_module.addImport("bytecode", bytecode_module);

    const context_module = b.createModule(.{
        .root_source_file = zesu.path("src/evm/context/main.zig"),
        .target = target,
        .optimize = optimize,
    });
    context_module.addImport("primitives", primitives_module);
    context_module.addImport("bytecode", bytecode_module);
    context_module.addImport("state", state_module);
    context_module.addImport("database", database_module);
    context_module.addImport("zesu_allocator", allocator_module);

    const precompile_module = b.createModule(.{
        .root_source_file = zesu.path("src/evm/precompile/main.zig"),
        .target = target,
        .optimize = optimize,
    });
    precompile_module.addImport("primitives", primitives_module);
    precompile_module.addImport("zesu_allocator", allocator_module);
    precompile_module.addImport("precompile_types", precompile_types_module);
    precompile_module.addImport("accelerators", accelerators_module);

    const interpreter_module = b.createModule(.{
        .root_source_file = zesu.path("src/evm/interpreter/main.zig"),
        .target = target,
        .optimize = optimize,
    });
    interpreter_module.addImport("primitives", primitives_module);
    interpreter_module.addImport("bytecode", bytecode_module);
    interpreter_module.addImport("context", context_module);
    interpreter_module.addImport("database", database_module);
    interpreter_module.addImport("state", state_module);
    interpreter_module.addImport("precompile", precompile_module);
    interpreter_module.addImport("zesu_allocator", allocator_module);
    interpreter_module.addImport("accelerators", accelerators_module);

    const handler_module = b.createModule(.{
        .root_source_file = zesu.path("src/evm/handler/main.zig"),
        .target = target,
        .optimize = optimize,
    });
    handler_module.addImport("primitives", primitives_module);
    handler_module.addImport("bytecode", bytecode_module);
    handler_module.addImport("state", state_module);
    handler_module.addImport("database", database_module);
    handler_module.addImport("interpreter", interpreter_module);
    handler_module.addImport("context", context_module);
    handler_module.addImport("precompile", precompile_module);
    handler_module.addImport("zesu_allocator", allocator_module);

    const input_module = b.createModule(.{
        .root_source_file = zesu.path("src/stateless/input.zig"),
        .target = target,
        .optimize = optimize,
    });
    input_module.addImport("primitives", primitives_module);

    const output_module = b.createModule(.{
        .root_source_file = zesu.path("src/stateless/output.zig"),
        .target = target,
        .optimize = optimize,
    });
    output_module.addImport("primitives", primitives_module);

    const hardfork_module = b.createModule(.{
        .root_source_file = zesu.path("src/stateless/hardfork.zig"),
        .target = target,
        .optimize = optimize,
    });
    hardfork_module.addImport("primitives", primitives_module);

    const rlp_decode_module = b.createModule(.{
        .root_source_file = zesu.path("src/stateless/rlp_decode.zig"),
        .target = target,
        .optimize = optimize,
    });
    rlp_decode_module.addImport("primitives", primitives_module);
    rlp_decode_module.addImport("input", input_module);

    const mpt_module = b.createModule(.{
        .root_source_file = zesu.path("src/stateless/mpt/main.zig"),
        .target = target,
        .optimize = optimize,
    });
    mpt_module.addImport("primitives", primitives_module);
    mpt_module.addImport("input", input_module);
    mpt_module.addImport("accelerators", accelerators_module);
    rlp_decode_module.addImport("mpt", mpt_module);

    const executor_types_module = b.createModule(.{
        .root_source_file = zesu.path("src/stateless/executor/types.zig"),
        .target = target,
        .optimize = optimize,
    });

    const db_module = b.createModule(.{
        .root_source_file = zesu.path("src/stateless/db/main.zig"),
        .target = target,
        .optimize = optimize,
    });
    db_module.addImport("primitives", primitives_module);
    db_module.addImport("state", state_module);
    db_module.addImport("bytecode", bytecode_module);
    db_module.addImport("mpt", mpt_module);
    db_module.addImport("executor_types", executor_types_module);

    const executor_module = b.createModule(.{
        .root_source_file = zesu.path("src/stateless/executor/main.zig"),
        .target = target,
        .optimize = optimize,
    });
    executor_module.addImport("executor_types", executor_types_module);
    executor_module.addImport("zesu_allocator", allocator_module);
    executor_module.addImport("primitives", primitives_module);
    executor_module.addImport("input", input_module);
    executor_module.addImport("output", output_module);
    executor_module.addImport("mpt", mpt_module);
    executor_module.addImport("rlp_decode", rlp_decode_module);
    executor_module.addImport("hardfork", hardfork_module);
    executor_module.addImport("db", db_module);
    executor_module.addImport("context", context_module);
    executor_module.addImport("state", state_module);
    executor_module.addImport("bytecode", bytecode_module);
    executor_module.addImport("database", database_module);
    executor_module.addImport("handler", handler_module);
    executor_module.addImport("precompile", precompile_module);
    executor_module.addImport("accelerators", accelerators_module);

    return .{
        .allocator = allocator_module,
        .executor = executor_module,
        .input = input_module,
        .mpt = mpt_module,
        .rlp_decode = rlp_decode_module,
        .precompile = precompile_module,
    };
}

fn addExecutionImports(module: *std.Build.Module, imports: ZesuImports) void {
    module.addImport("evm_guest_allocator", imports.allocator);
    module.addImport("zesu_executor", imports.executor);
    module.addImport("zesu_input", imports.input);
    module.addImport("zesu_mpt", imports.mpt);
    module.addImport("zesu_rlp_decode", imports.rlp_decode);
    module.addImport("zesu_precompile", imports.precompile);
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
    enable_secp256k1: bool,
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
            .enable_secp256k1 = false,
        };
    }

    const has_secp256k1 = fileExists(b.fmt("{s}/secp256k1.h", .{include_path})) and
        anyFileExists(&.{
            b.fmt("{s}/libsecp256k1.a", .{lib_path}),
            b.fmt("{s}/libsecp256k1.dylib", .{lib_path}),
            b.fmt("{s}/libsecp256k1.so", .{lib_path}),
        });

    return .{
        .include_path = include_path,
        .lib_path = lib_path,
        .enable_secp256k1 = b.option(bool, "secp256k1", "Enable native secp256k1-backed recovery") orelse has_secp256k1,
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

fn linkNativeZesuCrypto(
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
    linkCompileSystemLibrary(step, "c");
    if (target.result.os.tag != .windows) {
        linkCompileSystemLibrary(step, "m");
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
