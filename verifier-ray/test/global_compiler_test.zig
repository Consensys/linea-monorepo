const std = @import("std");
const verifier_ray = @import("verifier_ray");
const vectors = @import("test_vectors");

const field = verifier_ray.field.koalabear;
const ext = verifier_ray.field.koalabear_ext;
const global = verifier_ray.compiler.global;
const runtime = verifier_ray.runtime;

test "global compiler verifier matches prover-ray vanishing scenarios" {
    const allocator = std.testing.allocator;

    for (vectors.global_compiler_cases) |case| {
        var system = try buildSystem(allocator, case.modules);
        defer freeSystem(allocator, &system);

        var compiled = try global.Compile(allocator, system);
        defer compiled.deinit();

        var honest = GlobalProofBacking{};
        const honest_input = try honest.fill(case.honest, compiled.modules.len);
        try compiled.Check(allocator, honest_input);

        var invalid = GlobalProofBacking{};
        const invalid_input = try invalid.fill(case.invalid, compiled.modules.len);
        try std.testing.expectError(global.Error.QuotientIdentityMismatch, compiled.Check(allocator, invalid_input));
    }
}

fn buildSystem(allocator: std.mem.Allocator, modules: []const vectors.GlobalModule) !global.System {
    const out = try allocator.alloc(global.Module, modules.len);
    errdefer {
        for (out) |module| freeModule(allocator, module);
        allocator.free(out);
    }

    for (modules, out) |src, *dst| {
        const expressions = try allocator.alloc(global.ExprNode, src.expressions.len);
        errdefer allocator.free(expressions);
        for (src.expressions, expressions) |expr, *target| {
            target.* = convertExpression(expr);
        }

        const vanishings = try allocator.alloc(global.Vanishing, src.vanishings.len);
        errdefer allocator.free(vanishings);
        for (src.vanishings, vanishings) |vanishing, *target| {
            target.* = .{
                .expression = vanishing.expression,
                .cancelled_positions = vanishing.cancelled_positions,
            };
        }

        dst.* = .{
            .size = src.size,
            .expressions = expressions,
            .vanishings = vanishings,
        };
    }

    return .{ .modules = out };
}

fn freeSystem(allocator: std.mem.Allocator, system: *global.System) void {
    for (system.modules) |module| freeModule(allocator, module);
    allocator.free(system.modules);
    system.* = undefined;
}

fn freeModule(allocator: std.mem.Allocator, module: global.Module) void {
    allocator.free(module.expressions);
    allocator.free(module.vanishings);
}

fn convertExpression(expr: vectors.GlobalExprNode) global.ExprNode {
    return switch (expr) {
        .column_view => |view| .{ .column_view = .{
            .column = view.column,
            .shift = view.shift,
        } },
        .constant => |value| .{ .constant = elem(value) },
        .op => |op| .{ .op = .{
            .operator = convertOperator(op.operator),
            .operands = op.operands,
        } },
    };
}

fn convertOperator(op: vectors.GlobalOperator) global.Operator {
    return switch (op) {
        .add => .add,
        .mul => .mul,
        .sub => .sub,
        .div => .div,
        .double => .double,
        .square => .square,
        .negate => .negate,
        .inverse => .inverse,
    };
}

const global_dimensions = globalDimensions(vectors.global_compiler_cases);
const max_global_columns = @max(1, global_dimensions.columns);
const max_global_values = @max(1, global_dimensions.values);
const max_global_cells = @max(1, global_dimensions.cells);
const max_global_witness_claims = @max(1, global_dimensions.witness_claims);
const max_global_quotient_claims = @max(1, global_dimensions.quotient_claims);

const GlobalDimensions = struct {
    columns: usize = 0,
    values: usize = 0,
    cells: usize = 0,
    witness_claims: usize = 0,
    quotient_claims: usize = 0,
};

fn globalDimensions(comptime cases: anytype) GlobalDimensions {
    var dimensions = GlobalDimensions{};
    for (cases) |case| {
        updateProofDimensions(&dimensions, case.honest);
        updateProofDimensions(&dimensions, case.invalid);
    }
    return dimensions;
}

fn updateProofDimensions(comptime dimensions: *GlobalDimensions, comptime proof: vectors.GlobalProofView) void {
    updateRoundDimensions(dimensions, proof.initial_round);
    updateRoundDimensions(dimensions, proof.quotient_round);
    dimensions.witness_claims = @max(dimensions.witness_claims, proof.witness_claims.len);
    dimensions.quotient_claims = @max(dimensions.quotient_claims, proof.quotient_claims.len);
}

fn updateRoundDimensions(comptime dimensions: *GlobalDimensions, comptime round: vectors.GlobalRoundView) void {
    dimensions.columns = @max(dimensions.columns, round.columns.len);
    dimensions.cells = @max(dimensions.cells, round.cells.len);
    for (round.columns) |column| {
        dimensions.values = @max(dimensions.values, column.base_values.len);
        dimensions.values = @max(dimensions.values, column.ext_values.len);
    }
}

const GlobalProofBacking = struct {
    initial: RoundBacking = .{},
    quotient: RoundBacking = .{},
    witness_claims: [max_global_witness_claims]ext.Ext = undefined,
    quotient_claims: [max_global_quotient_claims]ext.Ext = undefined,

    fn fill(self: *GlobalProofBacking, proof: vectors.GlobalProofView, coin_count: usize) !global.CheckInput {
        const initial = try self.initial.fill(proof.initial_round, coin_count);
        const quotient = try self.quotient.fill(proof.quotient_round, coin_count);

        for (proof.witness_claims, 0..) |claim, i| {
            self.witness_claims[i] = uintsToExt(claim);
        }
        for (proof.quotient_claims, 0..) |claim, i| {
            self.quotient_claims[i] = uintsToExt(claim);
        }

        return .{
            .initial_round = initial,
            .quotient_round = quotient,
            .witness_claims = self.witness_claims[0..proof.witness_claims.len],
            .quotient_claims = self.quotient_claims[0..proof.quotient_claims.len],
        };
    }
};

const RoundBacking = struct {
    columns: [max_global_columns]runtime.ColumnAssignment = undefined,
    cells: [max_global_cells]runtime.Scalar = undefined,
    base_values: [max_global_columns][max_global_values]field.Element = undefined,
    ext_values: [max_global_columns][max_global_values]ext.Ext = undefined,

    fn fill(self: *RoundBacking, round: vectors.GlobalRoundView, coin_count: usize) !runtime.RoundMessage {
        try std.testing.expect(round.columns.len <= max_global_columns);
        try std.testing.expect(round.cells.len <= max_global_cells);

        for (round.columns, 0..) |column_case, i| {
            try std.testing.expect(column_case.is_assigned);
            const assignment: runtime.Vector = if (column_case.is_ext) assignment: {
                fillExts(&self.ext_values[i], column_case.ext_values);
                break :assignment .{ .ext = self.ext_values[i][0..column_case.ext_values.len] };
            } else assignment: {
                fillElems(&self.base_values[i], column_case.base_values);
                break :assignment .{ .base = self.base_values[i][0..column_case.base_values.len] };
            };
            self.columns[i] = .{
                .visibility = try visibility(column_case.visibility),
                .assignment = assignment,
            };
        }

        for (round.cells, 0..) |cell_case, i| {
            try std.testing.expect(cell_case.is_assigned);
            self.cells[i] = if (cell_case.is_ext)
                .{ .ext = uintsToExt(cell_case.ext_value) }
            else
                .{ .base = elem(cell_case.base_value) };
        }

        return .{
            .columns = self.columns[0..round.columns.len],
            .cells = self.cells[0..round.cells.len],
            .next_round_coin_count = coin_count,
        };
    }
};

fn elem(value: u32) field.Element {
    return field.Element.init(value);
}

fn uintsToExt(limbs: [6]u32) ext.Ext {
    return ext.Ext.fromUints(limbs[0], limbs[1], limbs[2], limbs[3], limbs[4], limbs[5]);
}

fn fillElems(out: []field.Element, values: []const u32) void {
    for (values, 0..) |value, i| {
        out[i] = elem(value);
    }
}

fn fillExts(out: []ext.Ext, values: []const [6]u32) void {
    for (values, 0..) |value, i| {
        out[i] = uintsToExt(value);
    }
}

fn visibility(value: u8) !runtime.Visibility {
    return switch (value) {
        1 => .oracle,
        2 => .public,
        else => error.InvalidVisibility,
    };
}
