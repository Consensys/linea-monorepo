const field = @import("../field/koalabear.zig");
const ext = @import("../field/koalabear_ext.zig");
const runtime = @import("../runtime.zig");

const Coin = runtime.Coin;

pub const Error = error{GlobalConstraintFailed};

/// Compute the cancellation polynomial C(r) = Π_{k ∈ cancelled} (r − ω_n^k).
///
/// Cancelled positions are normalised: negative values count from the end of
/// the domain (−1 is the last row, i.e. row n−1). Returns 1 when cancelled
/// is empty (no rows are exempt).
pub fn evalCancellation(
    n: usize,
    cancelled: []const i64,
    r: Coin,
) (field.Error || Error)!ext.Ext {
    if (cancelled.len == 0) return ext.Ext.one();
    const omega = try field.rootOfUnityBy(n);
    const n_i64: i64 = @intCast(n);
    var result = ext.Ext.one();
    for (cancelled) |raw_pos| {
        const shifted: i64 = if (raw_pos < 0) n_i64 + raw_pos else raw_pos;
        if (shifted < 0 or shifted >= n_i64) return Error.GlobalConstraintFailed;
        const pos: usize = @intCast(shifted);
        const root = ext.Ext.lift(omega.pow(@intCast(pos)));
        result = result.mul(r.sub(root));
    }
    return result;
}

/// Verify the PLONK global-quotient identity for one module.
///
/// Checks: P_agg(r) = (r^n − 1) · Q(r)
///
/// where:
///   P_agg(r) = Σ_i merge_coin^i · vanishing_evals[i]
///   Q(r)     = Σ_k (r^n)^k · quotient_evals[k]
///
/// Arguments:
///   n:               module size (power of 2, ≥ 1)
///   merge_coin:      extension-field challenge for batching vanishing constraints
///   r:               extension-field evaluation point
///   vanishing_evals: pre-computed P_i(r)·C_i(r) for each vanishing, in order
///   quotient_evals:  Q_k(r) for k = 0..ratio−1 (quotient-share evaluations)
///
/// The caller is responsible for supplying correct vanishing_evals, which are
/// the product of the expression value and the cancellation polynomial at r.
/// Both inputs are extension-field elements; base-field values should be lifted
/// via Ext.lift before passing.
pub fn verify(
    n: usize,
    merge_coin: Coin,
    r: Coin,
    vanishing_evals: []const ext.Ext,
    quotient_evals: []const ext.Ext,
) Error!void {
    // A prover-supplied n=0 would make ann=0 and trivially satisfy any P_agg.
    if (n == 0) return Error.GlobalConstraintFailed;
    const r_pow_n = r.pow(@intCast(n));
    const ann = r_pow_n.sub(ext.Ext.one()); // r^n - 1

    // Step 3: aggregate with merge coin — P_agg(r) = Σ_j merge_coin^j · P_j(r)
    var pagg = ext.Ext.zero();
    var coin_pow = ext.Ext.one();
    for (vanishing_evals) |pc| {
        pagg = pagg.add(coin_pow.mul(pc));
        coin_pow = coin_pow.mul(merge_coin);
    }

    // Step 4: reconstruct Q(r) from shares — Q(r) = Σ_k (r^n)^k · Q_k(r)
    var qr = ext.Ext.zero();
    var r_pow_kn = ext.Ext.one();
    for (quotient_evals) |qk| {
        qr = qr.add(r_pow_kn.mul(qk));
        r_pow_kn = r_pow_kn.mul(r_pow_n);
    }

    // Step 5: identity check — assert P_agg(r) == (r^n - 1) · Q(r)
    const rhs = ann.mul(qr);
    if (!pagg.eql(rhs)) return Error.GlobalConstraintFailed;
}
