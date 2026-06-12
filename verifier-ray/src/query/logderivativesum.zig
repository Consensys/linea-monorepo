const protocol = @import("../protocol/root.zig");
const ext = @import("../field/koalabear_ext.zig");

pub const Error = error{
    FinalSumMismatch,
    LookupResultNonZero,
};

// ScalarRef locates a cell opening as (round, index) into the verifier-visible
// transcript: ctx.rounds[round].cells[index]. Mirrors vanishing.ScalarRef.
pub const ScalarRef = struct {
    round: usize,
    index: usize,
};

// Query is one reduced LogDerivativeSum query. The logderivativesum compiler
// turns each query into Z running-sum columns whose recurrence and L_0 initial
// condition are ordinary vanishing constraints — already discharged by the
// vanishing sub-verifier. All that remains is the boundary identity:
//
//     Σ_i Z_i[n-1] == Result        (and, for lookups, Result == 0)
//
// z_finals are the openings of each Z column's last row; result is the query's
// claimed aggregated value. Every operand is a cell opening, so no expression
// evaluation is required.
pub const Query = struct {
    z_finals: []const ScalarRef,
    result: ScalarRef,
    result_is_zero: bool = false,
};

pub const System = struct {
    queries: []const Query = &.{},
};

pub fn verify(comptime system: System, ctx: protocol.Context) Error!void {
    inline for (system.queries) |query| {
        // Σ_i Z_i[n-1].
        var sum = ext.Ext.zero();
        inline for (query.z_finals) |ref| {
            sum = sum.add(cellExt(ctx, ref));
        }

        // The final-sum identity links the Z endpoints to the claimed result.
        const result = cellExt(ctx, query.result);
        if (!sum.eql(result)) return error.FinalSumMismatch;

        // Lookup queries reduce to a LogDerivativeSum whose result must be 0.
        if (query.result_is_zero and !result.isZero()) return error.LookupResultNonZero;
    }
}

fn cellExt(ctx: protocol.Context, comptime ref: ScalarRef) ext.Ext {
    return scalarToExt(ctx.rounds[ref.round].cells[ref.index]);
}

fn scalarToExt(value: protocol.Scalar) ext.Ext {
    return switch (value) {
        .base => |base| ext.Ext.lift(base),
        .ext => |extended| extended,
    };
}
