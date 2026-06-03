const field = @import("../field/koalabear.zig");
const ext = @import("../field/koalabear_ext.zig");
const runtime = @import("../runtime.zig");

pub const Error = runtime.Error || error{
    MissingDynamicModuleSize,
    InvalidModuleSize,
    InvalidClaimCount,
    QuotientIdentityMismatch,
};

pub const ModuleSize = union(enum) {
    static: usize,
    dynamic: usize,
};

pub const Operator = enum {
    add,
    mul,
    sub,
    div,
    double,
    square,
    negate,
    inverse,
};

pub const ExprOp = struct {
    operator: Operator,
    operands: []const usize,
};

pub const ScalarRef = struct {
    round: usize,
    index: usize,
};

pub const ExprNode = union(enum) {
    column_claim: usize,
    cell_value: ScalarRef,
    coin_value: usize,
    constant: field.Element,
    op: ExprOp,
};

pub const Vanishing = struct {
    expression: usize,
    cancelled_positions: []const i32 = &.{},
};

pub const Bucket = struct {
    ratio: usize,
    vanishings: []const Vanishing,
    quotient_claim_offset: usize,
};

pub const Module = struct {
    size: ModuleSize,
    expressions: []const ExprNode,
    buckets: []const Bucket,
    witness_claim_offset: usize,
};

pub const System = struct {
    modules: []const Module,
    dynamic_module_count: usize = 0,
    total_witness_claims: usize = 0,
    total_quotient_claims: usize = 0,
};

pub const CheckInput = struct {
    initial_round: runtime.RoundMessage,
    quotient_round: runtime.RoundMessage,
    witness_claims: []const ext.Ext,
    quotient_claims: []const ext.Ext,
    module_sizes: []const usize = &.{},
};

pub fn verify(comptime system: System, input: CheckInput) Error!void {
    if (input.witness_claims.len != system.total_witness_claims) return error.InvalidClaimCount;
    if (input.quotient_claims.len != system.total_quotient_claims) return error.InvalidClaimCount;
    if (input.module_sizes.len < system.dynamic_module_count) return error.MissingDynamicModuleSize;

    var rt = runtime.Runtime.initWithRoundCount(3);

    var merge_coins: [system.modules.len]runtime.Coin = undefined;
    const merge_message = runtime.RoundMessage{
        .columns = input.initial_round.columns,
        .cells = input.initial_round.cells,
        .next_round_coin_count = system.modules.len,
    };
    const merges = try rt.advanceRoundWithMessage(0, merge_message, &merge_coins);

    var eval_coins: [system.modules.len]runtime.Coin = undefined;
    const eval_message = runtime.RoundMessage{
        .columns = input.quotient_round.columns,
        .cells = input.quotient_round.cells,
        .next_round_coin_count = system.modules.len,
    };
    const evals = try rt.advanceRoundWithMessage(1, eval_message, &eval_coins);

    inline for (system.modules, 0..) |module, module_index| {
        switch (module.size) {
            .static => |n| try verifyModule(module, n, 0, input, merges[module_index], evals[module_index]),
            .dynamic => |size_index| {
                if (size_index >= input.module_sizes.len) return error.MissingDynamicModuleSize;
                try verifyModule(module, 0, input.module_sizes[size_index], input, merges[module_index], evals[module_index]);
            },
        }
    }
}

fn verifyModule(
    comptime module: Module,
    comptime static_n: usize,
    dynamic_n: usize,
    input: CheckInput,
    merge_coin: ext.Ext,
    eval_coin: ext.Ext,
) Error!void {
    // Static module sizes are embedded in the generated System, so Zig can
    // specialize this function at comptime. Dynamic modules use static_n == 0
    // as a sentinel and read n from CheckInput.module_sizes at runtime.
    //
    // The inline loops below are intentional: they traverse generated metadata
    // whose indices must stay comptime-known to avoid runtime expression-DAG
    // dispatch. Data loops, such as quotient-share recombination, remain plain
    // for loops.
    comptime {
        if (static_n != 0 and !validModuleSize(static_n)) {
            @compileError("static vanishing module size must be a non-zero power of two");
        }
    }
    if (static_n == 0 and !validModuleSize(dynamic_n)) return error.InvalidModuleSize;

    // Let r be the evaluation coin and H the module domain of size n.
    // The prover computes the domain annihilator Z_H(r) = r^n - 1.
    const r_pow_n = powModuleSize(eval_coin, static_n, dynamic_n);
    const annihilator = r_pow_n.sub(ext.Ext.one());

    inline for (module.buckets) |bucket| {
        try verifyBucket(module, bucket, static_n, dynamic_n, input, merge_coin, eval_coin, r_pow_n, annihilator);
    }
}

fn powModuleSize(r: ext.Ext, comptime static_n: usize, dynamic_n: usize) ext.Ext {
    // When static_n is non-zero, the exponent n is part of the comptime System
    // and powComptime emits a fixed exponentiation chain. Otherwise n is known
    // only from the verifier input and we use the runtime exponentiation path.
    if (static_n != 0) {
        return r.powComptime(static_n);
    }
    return r.pow(@as(u256, dynamic_n));
}

fn verifyBucket(
    comptime module: Module,
    comptime bucket: Bucket,
    comptime static_n: usize,
    dynamic_n: usize,
    input: CheckInput,
    merge_coin: ext.Ext,
    eval_coin: ext.Ext,
    r_pow_n: ext.Ext,
    annihilator: ext.Ext,
) Error!void {
    var quotient = ext.Ext.zero();
    var r_pow_kn = ext.Ext.one();
    for (0..bucket.ratio) |i| {
        // Recombine quotient-share claims:
        // Q(r) = sum_k r^(k*n) * Q_k(r) = sum_k (r^n)^k * Q_k(r).
        quotient = quotient.add(r_pow_kn.mul(input.quotient_claims[bucket.quotient_claim_offset + i]));
        r_pow_kn = r_pow_kn.mul(r_pow_n);
    }

    var aggregate = ext.Ext.zero();
    var coin_power = ext.Ext.one();
    inline for (bucket.vanishings) |v| {
        // Aggregate the vanished numerators with the merge coin alpha:
        // P_agg(r) = sum_i alpha^i * P_i(r) * C_i(r).
        const value = evalExpr(module, v.expression, input);
        const cancellation = try cancellationAtPoint(v.cancelled_positions, static_n, dynamic_n, eval_coin);
        aggregate = aggregate.add(coin_power.mul(value.mul(cancellation)));
        coin_power = coin_power.mul(merge_coin);
    }

    // PLONK quotient identity checked by prover-ray/global.Verifier.Check:
    // P_agg(r) = Z_H(r) * Q(r) = (r^n - 1) * Q(r).
    if (!aggregate.eql(annihilator.mul(quotient))) return error.QuotientIdentityMismatch;
}

fn evalExpr(comptime module: Module, comptime expr_index: usize, input: CheckInput) ext.Ext {
    const node = module.expressions[expr_index];
    return switch (node) {
        .column_claim => |claim_index| input.witness_claims[module.witness_claim_offset + claim_index],
        .constant => |value| ext.Ext.lift(value),
        .op => |op| evalOp(module, op, input),
    };
}

fn evalOp(comptime module: Module, comptime op: ExprOp, input: CheckInput) ext.Ext {
    const a = evalExpr(module, op.operands[0], input);
    return switch (op.operator) {
        .add => a.add(evalExpr(module, op.operands[1], input)),
        .mul => a.mul(evalExpr(module, op.operands[1], input)),
        .sub => a.sub(evalExpr(module, op.operands[1], input)),
        .div => a.div(evalExpr(module, op.operands[1], input)),
        .double => a.add(a),
        .square => a.square(),
        .negate => a.neg(),
        .inverse => a.inverse(),
    };
}

fn cancellationAtPoint(
    comptime positions: []const i32,
    comptime static_n: usize,
    dynamic_n: usize,
    r: ext.Ext,
) Error!ext.Ext {
    if (positions.len == 0) return ext.Ext.one();

    const omega = if (static_n == 0) field.rootOfUnityBy(dynamic_n) catch return error.InvalidModuleSize else field.Element.one();
    var result = ext.Ext.one();

    inline for (positions) |position| {
        // Cancellation polynomial for openings already enforced elsewhere:
        // C(r) = product_{k in cancelled} (r - omega_n^norm(k)).
        const root = if (static_n != 0) comptime staticRootPower(static_n, normalizePosition(position, static_n, 0)) else omega.pow(@as(u64, @intCast(normalizePosition(position, 0, dynamic_n))));
        result = result.mul(r.sub(ext.Ext.lift(root)));
    }
    return result;
}

fn staticRootPower(comptime n: usize, comptime k: usize) field.Element {
    const omega = field.rootOfUnityBy(n) catch unreachable;
    return omega.pow(@as(u64, k));
}

fn normalizePosition(comptime position: i32, comptime static_n: usize, dynamic_n: usize) usize {
    const n = if (static_n != 0) static_n else dynamic_n;
    if (position < 0) return n - @as(usize, @intCast(-position));
    return @as(usize, @intCast(position));
}

fn validModuleSize(n: usize) bool {
    return field.isPowerOfTwo(n);
}
