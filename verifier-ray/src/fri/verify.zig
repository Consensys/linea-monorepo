const std = @import("std");

const fiat_shamir = @import("../crypto/fiat_shamir.zig");
const field = @import("../field/koalabear.zig");
const ext = @import("../field/koalabear_ext.zig");
const leaf_hash = @import("leaf_hash.zig");
const merkle = @import("merkle.zig");
const types = @import("types.zig");

const tag_final_poly_base = field.Element.init(0x42415345); // "BASE"
const tag_final_poly_ext = field.Element.init(0x45585450); // "EXTP"
const max_fri_rounds = field.max_order_root;
const max_fri_levels = field.max_order_root;

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
    const challenges = try bindCommitPhase(params, level_roots, level_ds, proof, ts);
    try verifyQueries(params, level_roots, level_ds, proof, ts, challenges);
}

const CommitChallenges = struct {
    alphas: [max_fri_rounds]ext.Ext,
    gammas: [max_fri_levels]ext.Ext,

    fn empty() CommitChallenges {
        return .{
            .alphas = [_]ext.Ext{ext.Ext.zero()} ** max_fri_rounds,
            .gammas = [_]ext.Ext{ext.Ext.zero()} ** max_fri_levels,
        };
    }
};

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
    if (round_count > max_fri_rounds) return FriError.BadDimensions;
    if (level_roots.len > max_fri_levels) return FriError.BadDimensions;
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

    for (level_ds[1..], 0..) |level_d, offset| {
        const round = try introducedRound(level_ds[0], level_d, params.num_rounds);
        if (round == 0) return FriError.BadDimensions;
        for (level_ds[1..][0..offset]) |previous_level_d| {
            const previous_round = try introducedRound(level_ds[0], previous_level_d, params.num_rounds);
            if (previous_round == round) return FriError.BadDimensions;
        }
    }
}

fn bindCommitPhase(
    params: types.Params,
    level_roots: []const types.Digest,
    level_ds: []const u32,
    proof: types.FriProof,
    ts: *fiat_shamir.Transcript,
) FriError!CommitChallenges {
    var challenges = CommitChallenges.empty();

    var round: u32 = 0;
    while (round < params.num_rounds) : (round += 1) {
        for (level_ds[1..], 1..) |level_d, level_index| {
            const intro = try introducedRound(level_ds[0], level_d, params.num_rounds);
            if (intro != 0 and intro == round) {
                var gamma_name_buf: [48]u8 = undefined;
                const gamma_name = std.fmt.bufPrint(&gamma_name_buf, "fri_level_{d}_gamma", .{level_index}) catch {
                    return FriError.BadDimensions;
                };
                ts.newChallenge(gamma_name) catch return FriError.BadDimensions;
                ts.bindDigest(gamma_name, level_roots[level_index]) catch return FriError.BadDimensions;
                challenges.gammas[level_index] = ts.computeChallengeExt(gamma_name) catch |err| switch (err) {
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

        challenges.alphas[round_index] = ts.computeChallengeExt(fold_name) catch |err| switch (err) {
            error.InvalidProofOfWork => return FriError.InvalidProofOfWork,
            else => return FriError.BadDimensions,
        };
    }

    if (params.num_queries != 0) {
        var query_index_value: u32 = 0;
        while (query_index_value < params.num_queries) : (query_index_value += 1) {
            var query_name_buf: [32]u8 = undefined;
            const query_name = std.fmt.bufPrint(&query_name_buf, "fri_query_{d}", .{query_index_value}) catch {
                return FriError.BadDimensions;
            };
            ts.newChallenge(query_name) catch return FriError.BadDimensions;
        }
        try bindFinalPoly(ts, proof);
    }

    return challenges;
}

fn bindFinalPoly(ts: *fiat_shamir.Transcript, proof: types.FriProof) FriError!void {
    switch (proof.final_rail) {
        .base => {
            const header = [_]field.Element{
                tag_final_poly_base,
                field.Element.init(proof.final_poly_base.len),
            };
            ts.bindElements("fri_query_0", header[0..]) catch return FriError.BadDimensions;
            ts.bindElements("fri_query_0", proof.final_poly_base) catch return FriError.BadDimensions;
        },
        .ext => {
            const header = [_]field.Element{
                tag_final_poly_ext,
                field.Element.init(proof.final_poly_ext.len),
            };
            ts.bindElements("fri_query_0", header[0..]) catch return FriError.BadDimensions;
            for (proof.final_poly_ext) |value| {
                var limbs = extLimbs(value);
                ts.bindElements("fri_query_0", limbs[0..]) catch return FriError.BadDimensions;
            }
        },
    }
}

fn verifyQueries(
    params: types.Params,
    level_roots: []const types.Digest,
    level_ds: []const u32,
    proof: types.FriProof,
    ts: *fiat_shamir.Transcript,
    challenges: CommitChallenges,
) FriError!void {
    var query_index_value: u32 = 0;
    while (query_index_value < params.num_queries) : (query_index_value += 1) {
        var query_name_buf: [32]u8 = undefined;
        const query_name = std.fmt.bufPrint(&query_name_buf, "fri_query_{d}", .{query_index_value}) catch {
            return FriError.BadDimensions;
        };
        const challenge = ts.computeChallenge(query_name) catch return FriError.BadDimensions;
        const position = queryIndex(challenge, params.n / 2);

        if (query_index_value + 1 < params.num_queries) {
            var next_name_buf: [32]u8 = undefined;
            const next_name = std.fmt.bufPrint(&next_name_buf, "fri_query_{d}", .{query_index_value + 1}) catch {
                return FriError.BadDimensions;
            };
            ts.bindDigest(next_name, challenge) catch return FriError.BadDimensions;
        }

        const proof_query_index: usize = @intCast(query_index_value);
        switch (proof.final_rail) {
            .base => try checkQueryBase(
                position,
                proof_query_index,
                params,
                level_roots,
                level_ds,
                proof,
                challenges,
            ),
            .ext => try checkQueryExt(
                position,
                proof_query_index,
                params,
                level_roots,
                level_ds,
                proof,
                challenges,
            ),
        }
    }
}

fn checkQueryBase(
    position: u32,
    proof_query_index: usize,
    params: types.Params,
    level_roots: []const types.Digest,
    level_ds: []const u32,
    proof: types.FriProof,
    challenges: CommitChallenges,
) FriError!void {
    try verifyLevelQueryMerkleBase(position, proof_query_index, params, level_roots, level_ds, proof);

    var round: u32 = 0;
    while (round < params.num_rounds) : (round += 1) {
        const round_index: usize = @intCast(round);
        const domain_size = params.n >> @intCast(round);
        if (domain_size < 2) return FriError.BadDimensions;
        const base = position % (domain_size / 2);
        const layer = proof.queries[proof_query_index].layers[round_index];
        if (layer.rail != .base) return FriError.BadDimensions;
        try checkQueryLeafIndex(layer.path, base);

        const leaf = hashBasePair(layer.basePair());
        if (!merkle.merkleVerify(roundRoot(level_roots, proof, round), layer.path, leaf)) {
            return FriError.InvalidMerkleProof;
        }

        const expected = foldBasePair(
            layer.leaf_p_base,
            layer.leaf_q_base,
            challenges.alphas[round_index].B0.a0,
            xInv(params, round, base),
        );

        if (round + 1 < params.num_rounds) {
            const next = proof.queries[proof_query_index].layers[round_index + 1];
            const is_leaf_p = base < (domain_size / 4);
            var expected_next = expected;

            if (levelIntroducedAt(level_ds, params, round + 1)) |level_index| {
                const level_layer = proof.level_queries[level_index - 1][proof_query_index];
                if (level_layer.rail != .base) return FriError.BadDimensions;
                const level_value = if (is_leaf_p) level_layer.leaf_p_base else level_layer.leaf_q_base;
                const term = level_value.mul(challenges.gammas[level_index].B0.a0);
                expected_next = expected_next.add(term);
            }

            const actual_next = if (is_leaf_p) next.leaf_p_base else next.leaf_q_base;
            if (!expected_next.eql(actual_next)) return FriError.FoldMismatch;
        } else {
            const final_value = proof.final_poly_base[position % @as(u32, @intCast(proof.final_poly_base.len))];
            if (!expected.eql(final_value)) return FriError.FoldMismatch;
        }
    }
}

fn checkQueryExt(
    position: u32,
    proof_query_index: usize,
    params: types.Params,
    level_roots: []const types.Digest,
    level_ds: []const u32,
    proof: types.FriProof,
    challenges: CommitChallenges,
) FriError!void {
    try verifyLevelQueryMerkleExt(position, proof_query_index, params, level_roots, level_ds, proof);

    var round: u32 = 0;
    while (round < params.num_rounds) : (round += 1) {
        const round_index: usize = @intCast(round);
        const domain_size = params.n >> @intCast(round);
        if (domain_size < 2) return FriError.BadDimensions;
        const base = position % (domain_size / 2);
        const layer = proof.queries[proof_query_index].layers[round_index];
        if (layer.rail != .ext) return FriError.BadDimensions;
        try checkQueryLeafIndex(layer.path, base);

        const leaf = hashExtPair(layer.extPair());
        if (!merkle.merkleVerify(roundRoot(level_roots, proof, round), layer.path, leaf)) {
            return FriError.InvalidMerkleProof;
        }

        const expected = foldExtPair(
            layer.leaf_p_ext,
            layer.leaf_q_ext,
            challenges.alphas[round_index],
            xInv(params, round, base),
        );

        if (round + 1 < params.num_rounds) {
            const next = proof.queries[proof_query_index].layers[round_index + 1];
            const is_leaf_p = base < (domain_size / 4);
            var expected_next = expected;

            if (levelIntroducedAt(level_ds, params, round + 1)) |level_index| {
                const level_layer = proof.level_queries[level_index - 1][proof_query_index];
                if (level_layer.rail != .ext) return FriError.BadDimensions;
                const level_value = if (is_leaf_p) level_layer.leaf_p_ext else level_layer.leaf_q_ext;
                const term = level_value.mul(challenges.gammas[level_index]);
                expected_next = expected_next.add(term);
            }

            const actual_next = if (is_leaf_p) next.leaf_p_ext else next.leaf_q_ext;
            if (!expected_next.eql(actual_next)) return FriError.FoldMismatch;
        } else {
            const final_value = proof.final_poly_ext[position % @as(u32, @intCast(proof.final_poly_ext.len))];
            if (!expected.eql(final_value)) return FriError.FoldMismatch;
        }
    }
}

fn verifyLevelQueryMerkleBase(
    position: u32,
    proof_query_index: usize,
    params: types.Params,
    level_roots: []const types.Digest,
    level_ds: []const u32,
    proof: types.FriProof,
) FriError!void {
    for (proof.level_queries, 0..) |queries, level_offset| {
        const level_index = level_offset + 1;
        const intro_round = try introducedRound(level_ds[0], level_ds[level_index], params.num_rounds);
        const domain_size = params.n >> @intCast(intro_round);
        if (domain_size < 2) return FriError.BadDimensions;
        const base = position % (domain_size / 2);
        const layer = queries[proof_query_index];
        if (layer.rail != .base) return FriError.BadDimensions;
        try checkQueryLeafIndex(layer.path, base);

        if (!merkle.merkleVerify(level_roots[level_index], layer.path, hashBasePair(layer.basePair()))) {
            return FriError.InvalidMerkleProof;
        }
    }
}

fn verifyLevelQueryMerkleExt(
    position: u32,
    proof_query_index: usize,
    params: types.Params,
    level_roots: []const types.Digest,
    level_ds: []const u32,
    proof: types.FriProof,
) FriError!void {
    for (proof.level_queries, 0..) |queries, level_offset| {
        const level_index = level_offset + 1;
        const intro_round = try introducedRound(level_ds[0], level_ds[level_index], params.num_rounds);
        const domain_size = params.n >> @intCast(intro_round);
        if (domain_size < 2) return FriError.BadDimensions;
        const base = position % (domain_size / 2);
        const layer = queries[proof_query_index];
        if (layer.rail != .ext) return FriError.BadDimensions;
        try checkQueryLeafIndex(layer.path, base);

        if (!merkle.merkleVerify(level_roots[level_index], layer.path, hashExtPair(layer.extPair()))) {
            return FriError.InvalidMerkleProof;
        }
    }
}

fn levelIntroducedAt(level_ds: []const u32, params: types.Params, round: u32) ?usize {
    for (level_ds[1..], 1..) |level_d, level_index| {
        const intro_round = introducedRound(level_ds[0], level_d, params.num_rounds) catch return null;
        if (intro_round == round) return level_index;
    }
    return null;
}

fn roundRoot(level_roots: []const types.Digest, proof: types.FriProof, round: u32) types.Digest {
    if (round == 0) return level_roots[0];
    return proof.fri_roots[@intCast(round - 1)];
}

fn checkQueryLeafIndex(path: types.MerklePath, expected: u32) FriError!void {
    // The opened leaf position is fixed by the Fiat-Shamir query, not by the prover-supplied Merkle path.
    if (path.leaf_idx != expected) return FriError.BadDimensions;
}

fn hashBasePair(pair: types.PairBase) types.Digest {
    const pairs = [_]types.PairBase{pair};
    return leaf_hash.hashLeaf(pairs[0..], &.{});
}

fn hashExtPair(pair: types.PairExt) types.Digest {
    const pairs = [_]types.PairExt{pair};
    return leaf_hash.hashLeaf(&.{}, pairs[0..]);
}

fn foldBasePair(p: field.Element, q: field.Element, alpha: field.Element, x_inv: field.Element) field.Element {
    const sum = p.add(q).halve();
    const diff = p.sub(q).halve().mul(x_inv).mul(alpha);
    return sum.add(diff);
}

fn foldExtPair(p: ext.Ext, q: ext.Ext, alpha: ext.Ext, x_inv: field.Element) ext.Ext {
    const sum = p.add(q).halve();
    const diff = p.sub(q).halve().mulByBase(x_inv).mul(alpha);
    return sum.add(diff);
}

fn xInv(params: types.Params, round: u32, base: u32) field.Element {
    return params.domain_gens_inv[@intCast(round)].pow(base);
}

fn queryIndex(challenge: types.Digest, modulus: u32) u32 {
    if (modulus == 0) return 0;
    const wide = (@as(u64, challenge[0].value) << 31) ^ @as(u64, challenge[1].value);
    return @intCast(wide % modulus);
}

fn introducedRound(root_d: u32, level_d: u32, num_rounds: u32) FriError!u32 {
    if (root_d == 0 or level_d == 0 or level_d > root_d) return FriError.BadDimensions;
    if (!field.isPowerOfTwo(@intCast(level_d))) return FriError.BadDimensions;
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
