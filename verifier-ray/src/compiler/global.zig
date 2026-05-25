const std = @import("std");

const runtime = @import("../runtime.zig");
const field = @import("../field/koalabear.zig");
const ext = @import("../field/koalabear_ext.zig");
const lagrange = @import("../pcs/lagrange.zig");

pub const Error = std.mem.Allocator.Error || runtime.Error || lagrange.Error || error{
    InvalidExpression,
    InvalidOperatorArity,
    NonPolynomialExpression,
    EmptyModule,
    InvalidCoinCount,
    InvalidClaimCount,
    InvalidColumnCount,
    WitnessClaimMismatch,
    QuotientClaimMismatch,
    QuotientIdentityMismatch,
};

pub const Operator = enum(u8) {
    add,
    mul,
    sub,
    div,
    double,
    square,
    negate,
    inverse,
};

pub const ColumnView = struct {
    column: usize,
    shift: i32 = 0,
};

pub const ExprOp = struct {
    operator: Operator,
    operands: []const usize,
};

pub const ExprNode = union(enum) {
    column_view: ColumnView,
    constant: field.Element,
    op: ExprOp,
};

pub const Vanishing = struct {
    expression: usize,
    cancelled_positions: []const i32 = &.{},
};

pub const Module = struct {
    size: usize,
    expressions: []const ExprNode,
    vanishings: []const Vanishing,
};

pub const System = struct {
    modules: []const Module,
};

pub const CheckInput = struct {
    /// Verifier-visible messages from the round before the global quotient
    /// compiler's quotient round. This usually contains the trace columns.
    initial_round: runtime.RoundMessage,
    /// Verifier-visible messages from the quotient round. This contains the
    /// quotient-share columns committed after the merge coin is known.
    quotient_round: runtime.RoundMessage,
    /// Claimed evaluations of all unique witness column views at evalCoin.
    witness_claims: []const ext.Ext,
    /// Claimed evaluations of quotient shares at evalCoin, flattened in the
    /// same order as the quotient columns are declared.
    quotient_claims: []const ext.Ext,
};

pub const Compiled = struct {
    allocator: std.mem.Allocator,
    modules: []CompiledModule,
    total_witness_claims: usize,
    total_quotient_claims: usize,
    total_quotient_columns: usize,

    pub fn deinit(self: *Compiled) void {
        for (self.modules) |*module| {
            self.allocator.free(module.witness_views);
            for (module.buckets) |*bucket| {
                self.allocator.free(bucket.vanishing_indices);
            }
            self.allocator.free(module.buckets);
        }
        self.allocator.free(self.modules);
        self.* = undefined;
    }

    pub fn Check(self: *const Compiled, allocator: std.mem.Allocator, input: CheckInput) Error!void {
        if (input.initial_round.next_round_coin_count != self.modules.len) return Error.InvalidCoinCount;
        if (input.quotient_round.next_round_coin_count != self.modules.len) return Error.InvalidCoinCount;
        if (input.witness_claims.len != self.total_witness_claims) return Error.InvalidClaimCount;
        if (input.quotient_claims.len != self.total_quotient_claims) return Error.InvalidClaimCount;
        if (input.quotient_round.columns.len != self.total_quotient_columns) return Error.InvalidColumnCount;

        var rt = runtime.Runtime.initWithRoundCount(3);

        const merge_coin_buf = try allocator.alloc(runtime.Coin, self.modules.len);
        defer allocator.free(merge_coin_buf);
        const merge_coins = try rt.advanceRoundWithMessage(0, input.initial_round, merge_coin_buf);
        if (merge_coins.len != self.modules.len) return Error.InvalidCoinCount;

        const eval_coin_buf = try allocator.alloc(runtime.Coin, self.modules.len);
        defer allocator.free(eval_coin_buf);
        const eval_coins = try rt.advanceRoundWithMessage(1, input.quotient_round, eval_coin_buf);
        if (eval_coins.len != self.modules.len) return Error.InvalidCoinCount;

        for (self.modules, 0..) |module, module_index| {
            const merge_coin = merge_coins[module_index];
            const eval_coin = eval_coins[module_index];
            try module.checkLagrangeClaims(input, eval_coin);
            try module.checkQuotientIdentity(input, merge_coin, eval_coin);
        }
    }
};

pub const CompiledModule = struct {
    spec: Module,
    witness_claim_offset: usize,
    witness_views: []ColumnView,
    buckets: []CompiledBucket,

    fn checkLagrangeClaims(self: CompiledModule, input: CheckInput, eval_coin: ext.Ext) Error!void {
        for (self.witness_views, 0..) |view, i| {
            if (view.column >= input.initial_round.columns.len) return Error.InvalidColumnCount;
            const column = input.initial_round.columns[view.column];
            const point = try shiftedPoint(self.spec.size, eval_coin, view.shift);
            const got = try evaluateVectorAtExt(column.assignment, point);
            const want = input.witness_claims[self.witness_claim_offset + i];
            if (!got.eql(want)) return Error.WitnessClaimMismatch;
        }

        for (self.buckets) |bucket| {
            for (0..bucket.ratio) |k| {
                const column_index = bucket.quotient_column_offset + k;
                if (column_index >= input.quotient_round.columns.len) return Error.InvalidColumnCount;
                const got = try evaluateVectorAtExt(input.quotient_round.columns[column_index].assignment, eval_coin);
                const want = input.quotient_claims[bucket.quotient_claim_offset + k];
                if (!got.eql(want)) return Error.QuotientClaimMismatch;
            }
        }
    }

    fn checkQuotientIdentity(
        self: CompiledModule,
        input: CheckInput,
        merge_coin: ext.Ext,
        eval_coin: ext.Ext,
    ) Error!void {
        const r_pow_n = eval_coin.pow(@as(u256, @intCast(self.spec.size)));
        const annihilator = r_pow_n.sub(ext.Ext.one());

        for (self.buckets) |bucket| {
            var quotient_at_r = ext.Ext.zero();
            var r_pow_kn = ext.Ext.one();
            for (0..bucket.ratio) |k| {
                const qk = input.quotient_claims[bucket.quotient_claim_offset + k];
                quotient_at_r = quotient_at_r.add(r_pow_kn.mul(qk));
                r_pow_kn = r_pow_kn.mul(r_pow_n);
            }

            var aggregate_at_r = ext.Ext.zero();
            var coin_pow = ext.Ext.one();
            for (bucket.vanishing_indices) |vanishing_index| {
                const vanishing = self.spec.vanishings[vanishing_index];
                const p_at_r = try self.evalExprAtPoint(vanishing.expression, input);
                const c_at_r = try evalCancellationAtPoint(vanishing.cancelled_positions, self.spec.size, eval_coin);
                aggregate_at_r = aggregate_at_r.add(coin_pow.mul(p_at_r.mul(c_at_r)));
                coin_pow = coin_pow.mul(merge_coin);
            }

            const rhs = annihilator.mul(quotient_at_r);
            if (!aggregate_at_r.eql(rhs)) return Error.QuotientIdentityMismatch;
        }
    }

    fn evalExprAtPoint(self: CompiledModule, expr_index: usize, input: CheckInput) Error!ext.Ext {
        if (expr_index >= self.spec.expressions.len) return Error.InvalidExpression;
        return switch (self.spec.expressions[expr_index]) {
            .column_view => |view| blk: {
                const claim_index = self.findWitnessView(view) orelse return Error.InvalidExpression;
                break :blk input.witness_claims[self.witness_claim_offset + claim_index];
            },
            .constant => |constant| ext.Ext.lift(constant),
            .op => |op| blk: {
                break :blk try self.evalOpAtPoint(op, input);
            },
        };
    }

    fn evalOpAtPoint(self: CompiledModule, op: ExprOp, input: CheckInput) Error!ext.Ext {
        return switch (op.operator) {
            .add => {
                if (op.operands.len != 2) return Error.InvalidOperatorArity;
                return (try self.evalExprAtPoint(op.operands[0], input)).add(try self.evalExprAtPoint(op.operands[1], input));
            },
            .mul => {
                if (op.operands.len != 2) return Error.InvalidOperatorArity;
                return (try self.evalExprAtPoint(op.operands[0], input)).mul(try self.evalExprAtPoint(op.operands[1], input));
            },
            .sub => {
                if (op.operands.len != 2) return Error.InvalidOperatorArity;
                return (try self.evalExprAtPoint(op.operands[0], input)).sub(try self.evalExprAtPoint(op.operands[1], input));
            },
            .div => {
                if (op.operands.len != 2) return Error.InvalidOperatorArity;
                return (try self.evalExprAtPoint(op.operands[0], input)).div(try self.evalExprAtPoint(op.operands[1], input));
            },
            .double => {
                if (op.operands.len != 1) return Error.InvalidOperatorArity;
                const value = try self.evalExprAtPoint(op.operands[0], input);
                return value.add(value);
            },
            .square => {
                if (op.operands.len != 1) return Error.InvalidOperatorArity;
                const value = try self.evalExprAtPoint(op.operands[0], input);
                return value.square();
            },
            .negate => {
                if (op.operands.len != 1) return Error.InvalidOperatorArity;
                return (try self.evalExprAtPoint(op.operands[0], input)).neg();
            },
            .inverse => {
                if (op.operands.len != 1) return Error.InvalidOperatorArity;
                return (try self.evalExprAtPoint(op.operands[0], input)).inverse();
            },
        };
    }

    fn findWitnessView(self: CompiledModule, needle: ColumnView) ?usize {
        for (self.witness_views, 0..) |view, i| {
            if (view.column == needle.column and view.shift == needle.shift) return i;
        }
        return null;
    }
};

pub const CompiledBucket = struct {
    ratio: usize,
    vanishing_indices: []usize,
    quotient_claim_offset: usize,
    quotient_column_offset: usize,
};

pub fn Compile(allocator: std.mem.Allocator, system: System) Error!Compiled {
    var modules: std.ArrayList(CompiledModule) = .empty;
    errdefer {
        for (modules.items) |*module| {
            allocator.free(module.witness_views);
            for (module.buckets) |*bucket| allocator.free(bucket.vanishing_indices);
            allocator.free(module.buckets);
        }
        modules.deinit(allocator);
    }

    var total_witness_claims: usize = 0;
    var total_quotient_claims: usize = 0;
    var total_quotient_columns: usize = 0;

    for (system.modules) |module| {
        if (module.vanishings.len == 0) continue;
        if (module.size == 0) return Error.EmptyModule;

        var ratios = try allocator.alloc(usize, module.vanishings.len);
        defer allocator.free(ratios);

        var ratio_order: std.ArrayList(usize) = .empty;
        defer ratio_order.deinit(allocator);

        for (module.vanishings, 0..) |vanishing, i| {
            const factor = try degreeFactor(module, vanishing.expression);
            const ratio = computeRatio(factor, vanishing.cancelled_positions.len);
            ratios[i] = ratio;
            if (!containsRatio(ratio_order.items, ratio)) {
                try ratio_order.append(allocator, ratio);
            }
        }

        var witness_views_list: std.ArrayList(ColumnView) = .empty;
        defer witness_views_list.deinit(allocator);
        for (ratio_order.items) |ratio| {
            for (module.vanishings, 0..) |vanishing, vanishing_index| {
                if (ratios[vanishing_index] != ratio) continue;
                try collectColumnViews(allocator, module, vanishing.expression, &witness_views_list);
            }
        }
        const witness_views = try witness_views_list.toOwnedSlice(allocator);
        errdefer allocator.free(witness_views);

        const buckets = try allocator.alloc(CompiledBucket, ratio_order.items.len);
        var buckets_initialized: usize = 0;
        errdefer {
            for (buckets[0..buckets_initialized]) |*bucket| allocator.free(bucket.vanishing_indices);
            allocator.free(buckets);
        }

        for (ratio_order.items, 0..) |ratio, bucket_index| {
            var count: usize = 0;
            for (ratios) |r| {
                if (r == ratio) count += 1;
            }
            const indices = try allocator.alloc(usize, count);
            var j: usize = 0;
            for (ratios, 0..) |r, vanishing_index| {
                if (r == ratio) {
                    indices[j] = vanishing_index;
                    j += 1;
                }
            }
            buckets[bucket_index] = .{
                .ratio = ratio,
                .vanishing_indices = indices,
                .quotient_claim_offset = total_quotient_claims,
                .quotient_column_offset = total_quotient_columns,
            };
            buckets_initialized += 1;
            total_quotient_claims += ratio;
            total_quotient_columns += ratio;
        }

        try modules.append(allocator, .{
            .spec = module,
            .witness_claim_offset = total_witness_claims,
            .witness_views = witness_views,
            .buckets = buckets,
        });
        total_witness_claims += witness_views.len;
    }

    return .{
        .allocator = allocator,
        .modules = try modules.toOwnedSlice(allocator),
        .total_witness_claims = total_witness_claims,
        .total_quotient_claims = total_quotient_claims,
        .total_quotient_columns = total_quotient_columns,
    };
}

fn collectColumnViews(
    allocator: std.mem.Allocator,
    module: Module,
    expr_index: usize,
    views: *std.ArrayList(ColumnView),
) Error!void {
    if (expr_index >= module.expressions.len) return Error.InvalidExpression;
    switch (module.expressions[expr_index]) {
        .column_view => |view| {
            for (views.items) |existing| {
                if (existing.column == view.column and existing.shift == view.shift) return;
            }
            try views.append(allocator, view);
        },
        .constant => {},
        .op => |op| {
            for (op.operands) |operand| try collectColumnViews(allocator, module, operand, views);
        },
    }
}

fn degreeFactor(module: Module, expr_index: usize) Error!usize {
    if (expr_index >= module.expressions.len) return Error.InvalidExpression;
    return switch (module.expressions[expr_index]) {
        .column_view => 1,
        .constant => 0,
        .op => |op| try degreeFactorOp(module, op),
    };
}

fn degreeFactorOp(module: Module, op: ExprOp) Error!usize {
    return switch (op.operator) {
        .add, .sub => {
            if (op.operands.len != 2) return Error.InvalidOperatorArity;
            return @max(try degreeFactor(module, op.operands[0]), try degreeFactor(module, op.operands[1]));
        },
        .mul => {
            if (op.operands.len != 2) return Error.InvalidOperatorArity;
            return (try degreeFactor(module, op.operands[0])) + (try degreeFactor(module, op.operands[1]));
        },
        .double, .negate => {
            if (op.operands.len != 1) return Error.InvalidOperatorArity;
            return try degreeFactor(module, op.operands[0]);
        },
        .square => {
            if (op.operands.len != 1) return Error.InvalidOperatorArity;
            return 2 * (try degreeFactor(module, op.operands[0]));
        },
        .div, .inverse => Error.NonPolynomialExpression,
    };
}

fn computeRatio(factor: usize, cancelled_count: usize) usize {
    const numerator_tail = @as(isize, @intCast(cancelled_count)) - @as(isize, @intCast(factor)) + 1;
    if (numerator_tail > 0) {
        return nextPowerOfTwo(@max(@as(usize, 1), factor));
    }
    return nextPowerOfTwo(@max(@as(usize, 1), factor -| 1));
}

fn nextPowerOfTwo(value: usize) usize {
    var result: usize = 1;
    while (result < value) : (result <<= 1) {}
    return result;
}

fn containsRatio(values: []const usize, needle: usize) bool {
    for (values) |value| {
        if (value == needle) return true;
    }
    return false;
}

fn evaluateVectorAtExt(vector: runtime.Vector, point: ext.Ext) Error!ext.Ext {
    return switch (vector) {
        .base => |values| try lagrange.evaluateBaseAtExt(values, point),
        .ext => |values| try lagrange.evaluateExtAtExt(values, point),
    };
}

fn shiftedPoint(cardinality: usize, point: ext.Ext, shift: i32) Error!ext.Ext {
    if (shift == 0) return point;
    const omega_shift = try rootPower(cardinality, shift);
    return point.mul(ext.Ext.lift(omega_shift));
}

fn rootPower(cardinality: usize, shift: i32) Error!field.Element {
    const omega = try field.rootOfUnityBy(cardinality);
    const n = @as(i64, @intCast(cardinality));
    const normalized = @mod(@as(i64, shift), n);
    return omega.pow(@as(u64, @intCast(normalized)));
}

fn evalCancellationAtPoint(cancelled: []const i32, cardinality: usize, point: ext.Ext) Error!ext.Ext {
    if (cancelled.len == 0) return ext.Ext.one();

    const omega = try field.rootOfUnityBy(cardinality);
    const n = @as(i64, @intCast(cardinality));
    var result = ext.Ext.one();
    for (cancelled) |position| {
        const normalized = @mod(@as(i64, position), n);
        const omega_k = omega.pow(@as(u64, @intCast(normalized)));
        result = result.mul(point.sub(ext.Ext.lift(omega_k)));
    }
    return result;
}
