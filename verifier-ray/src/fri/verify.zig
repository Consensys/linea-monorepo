const std = @import("std");

const fiat_shamir = @import("../crypto/fiat_shamir.zig");
const field = @import("../field/koalabear.zig");
const ext = @import("../field/koalabear_ext.zig");
const types = @import("types.zig");

pub const FriError = error{
    BadDimensions,
    BadFinalPoly,
    MissingProofOfWork,
    InvalidProofOfWork,
    InvalidMerkleProof,
    FoldMismatch,
    Unsupported,
};

pub fn friVerify(
    params: types.Params,
    level_roots: []const types.Digest,
    level_ds: []const u32,
    proof: types.FriProof,
    ts: *fiat_shamir.Transcript,
) FriError!void {
    try checkDimensions(params, level_roots, level_ds, proof);
    try bindCommitPhase(params, level_roots, level_ds, proof, ts);

    // The settled loom proof-byte schema and golden fold vectors are still
    // prerequisites for the per-query fold verifier. Avoid returning success
    // for a non-empty query proof until those checks are implemented.
    if (proof.queries.len != 0) return FriError.Unsupported;
}

fn checkDimensions(
    params: types.Params,
    level_roots: []const types.Digest,
    level_ds: []const u32,
    proof: types.FriProof,
) FriError!void {
    if (level_roots.len == 0 or level_roots.len != level_ds.len) return FriError.BadDimensions;
    if (level_ds[0] != params.d) return FriError.BadDimensions;
    if (params.num_rounds == 0) return FriError.BadDimensions;
    const round_count: usize = @intCast(params.num_rounds);
    const query_count: usize = @intCast(params.num_queries);
    if (proof.fri_roots.len + 1 != round_count) return FriError.BadDimensions;
    if (proof.pow.len != round_count) return FriError.BadDimensions;
    if (proof.queries.len != query_count) return FriError.BadDimensions;
    if (proof.level_queries.len + 1 != level_roots.len) return FriError.BadDimensions;
    if (params.domain_gens.len < round_count) return FriError.BadDimensions;
    if (params.domain_gens_inv.len < round_count) return FriError.BadDimensions;

    switch (proof.final_rail) {
        .base => if (proof.final_poly_base.len == 0) return FriError.BadFinalPoly,
        .ext => if (proof.final_poly_ext.len == 0) return FriError.BadFinalPoly,
    }

    for (proof.queries) |query| {
        if (query.layers.len != round_count) return FriError.BadDimensions;
    }

    for (proof.level_queries) |level_query| {
        if (level_query.len != query_count) return FriError.BadDimensions;
    }

    for (level_ds[1..]) |level_d| {
        _ = try introducedRound(level_ds[0], level_d, params.num_rounds);
    }
}

fn bindCommitPhase(
    params: types.Params,
    level_roots: []const types.Digest,
    level_ds: []const u32,
    proof: types.FriProof,
    ts: *fiat_shamir.Transcript,
) FriError!void {
    var round: u32 = 0;
    while (round < params.num_rounds) : (round += 1) {
        for (level_ds[1..], 1..) |level_d, level_index| {
            const intro = try introducedRound(level_ds[0], level_d, params.num_rounds);
            if (intro == round) {
                var gamma_name_buf: [48]u8 = undefined;
                const gamma_name = std.fmt.bufPrint(&gamma_name_buf, "fri_level_{d}_gamma", .{level_index}) catch {
                    return FriError.BadDimensions;
                };
                ts.newChallenge(gamma_name) catch return FriError.BadDimensions;
                ts.bindDigest(gamma_name, level_roots[level_index]) catch return FriError.BadDimensions;
                _ = ts.computeChallengeExt(gamma_name) catch |err| switch (err) {
                    error.InvalidProofOfWork => return FriError.InvalidProofOfWork,
                    else => return FriError.BadDimensions,
                };
            }
        }

        var fold_name_buf: [32]u8 = undefined;
        const fold_name = std.fmt.bufPrint(&fold_name_buf, "fri_fold_{d}", .{round}) catch {
            return FriError.BadDimensions;
        };
        const round_index: usize = @intCast(round);
        const round_root = if (round == 0) level_roots[0] else proof.fri_roots[round_index - 1];

        ts.newChallenge(fold_name) catch return FriError.BadDimensions;
        ts.bindDigest(fold_name, round_root) catch return FriError.BadDimensions;

        if (params.grinding != 0) {
            const pow = proof.pow[round_index] orelse return FriError.MissingProofOfWork;
            if (pow.nb_bits != params.grinding) return FriError.InvalidProofOfWork;
            ts.setProofOfWork(fold_name, pow) catch return FriError.InvalidProofOfWork;
        } else if (proof.pow[round_index]) |pow| {
            if (pow.nb_bits != 0) return FriError.InvalidProofOfWork;
            ts.setProofOfWork(fold_name, pow) catch return FriError.InvalidProofOfWork;
        }

        _ = ts.computeChallengeExt(fold_name) catch |err| switch (err) {
            error.InvalidProofOfWork => return FriError.InvalidProofOfWork,
            else => return FriError.BadDimensions,
        };
    }

    if (params.num_queries != 0) {
        ts.newChallenge("fri_query_0") catch return FriError.BadDimensions;
        switch (proof.final_rail) {
            .base => ts.bindElements("fri_query_0", proof.final_poly_base) catch return FriError.BadDimensions,
            .ext => for (proof.final_poly_ext) |value| {
                var limbs = extLimbs(value);
                ts.bindElements("fri_query_0", limbs[0..]) catch return FriError.BadDimensions;
            },
        }
    }
}

fn introducedRound(root_d: u32, level_d: u32, num_rounds: u32) FriError!u32 {
    if (root_d == 0 or level_d == 0 or level_d > root_d) return FriError.BadDimensions;
    if (root_d % level_d != 0) return FriError.BadDimensions;

    const ratio = root_d / level_d;
    if (!field.isPowerOfTwo(@intCast(ratio))) return FriError.BadDimensions;
    const round: u32 = @intCast(field.log2PowerOfTwo(@intCast(ratio)));
    if (round >= num_rounds) return FriError.BadDimensions;
    return round;
}

fn extLimbs(value: ext.Ext) [6]field.Element {
    return .{
        value.B0.a0,
        value.B0.a1,
        value.B1.a0,
        value.B1.a1,
        value.B2.a0,
        value.B2.a1,
    };
}
