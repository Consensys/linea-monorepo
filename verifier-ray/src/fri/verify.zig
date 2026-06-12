const field = @import("../field/koalabear.zig");
const ext = @import("../field/koalabear_ext.zig");
const leaf_hash = @import("leaf_hash.zig");
const merkle = @import("merkle.zig");
const types = @import("types.zig");

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

pub const FriChallenges = struct {
    /// Single batching challenge `fri_gamma` when extra FRI levels exist.
    /// Level l consumes gamma^l, matching prover-ray's multi-level FRI.
    gamma: ?ext.Ext = null,
    /// Fold challenges `fri_fold_j`, one per folding round. Base-rail proofs
    /// consume the first limb; extension-rail proofs consume the full value.
    fold_alphas: []const ext.Ext,
    /// Query positions already reduced modulo params.n / 2.
    query_positions: []const u32,
};

pub fn friVerify(
    params: types.Params,
    level_roots: []const types.Digest,
    level_ds: []const u32,
    proof: types.FriProof,
    challenges: FriChallenges,
) FriError!void {
    try checkDimensions(params, level_roots, level_ds, proof);
    const resolved = try resolveChallenges(params, level_roots, proof, challenges);
    try verifyQueries(params, level_roots, level_ds, proof, challenges.query_positions, resolved);
}

const ResolvedChallenges = struct {
    alphas: [max_fri_rounds]ext.Ext,
    gammas_base: [max_fri_levels]field.Element,
    gammas_ext: [max_fri_levels]ext.Ext,

    fn empty() ResolvedChallenges {
        return .{
            .alphas = [_]ext.Ext{ext.Ext.zero()} ** max_fri_rounds,
            .gammas_base = [_]field.Element{field.Element.zero()} ** max_fri_levels,
            .gammas_ext = [_]ext.Ext{ext.Ext.zero()} ** max_fri_levels,
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

fn resolveChallenges(
    params: types.Params,
    level_roots: []const types.Digest,
    proof: types.FriProof,
    challenges: FriChallenges,
) FriError!ResolvedChallenges {
    const round_count: usize = @intCast(params.num_rounds);
    const query_count: usize = @intCast(params.num_queries);
    if (challenges.fold_alphas.len != round_count) return FriError.BadDimensions;
    if (challenges.query_positions.len != query_count) return FriError.BadDimensions;
    try validateProofOfWork(params, proof, challenges.fold_alphas);

    for (challenges.query_positions) |position| {
        if (position >= params.n / 2) return FriError.BadDimensions;
    }

    var resolved = ResolvedChallenges.empty();
    @memcpy(resolved.alphas[0..round_count], challenges.fold_alphas);

    if (level_roots.len > 1) {
        const gamma = challenges.gamma orelse return FriError.BadDimensions;
        const gamma_base = gamma.B0.a0;
        resolved.gammas_base[1] = gamma_base;
        resolved.gammas_ext[1] = gamma;
        var level_index: usize = 2;
        while (level_index < level_roots.len) : (level_index += 1) {
            resolved.gammas_base[level_index] = resolved.gammas_base[level_index - 1].mul(gamma_base);
            resolved.gammas_ext[level_index] = resolved.gammas_ext[level_index - 1].mul(gamma);
        }
    } else if (challenges.gamma != null) {
        return FriError.BadDimensions;
    }

    return resolved;
}

fn validateProofOfWork(
    params: types.Params,
    proof: types.FriProof,
    fold_alphas: []const ext.Ext,
) FriError!void {
    // The pure verifier can enforce the declared PoW shape and low-bit
    // constraint on injected fold coins. The named transcript binding from
    // salt to coin is enforced by the challenge-production layer.
    if (params.grinding > 31) return FriError.InvalidProofOfWork;
    for (proof.pow, fold_alphas) |maybe_pow, alpha| {
        if (params.grinding != 0) {
            const pow = maybe_pow orelse return FriError.MissingProofOfWork;
            if (pow.nb_bits != params.grinding) return FriError.InvalidProofOfWork;
            if (!hasLowZeroBits(alpha, params.grinding)) return FriError.InvalidProofOfWork;
        } else if (maybe_pow) |pow| {
            if (pow.nb_bits != 0) return FriError.InvalidProofOfWork;
        }
    }
}

fn hasLowZeroBits(value: ext.Ext, nb_bits: u32) bool {
    var remaining = nb_bits;
    const limbs = [_]field.Element{
        value.B0.a0,
        value.B0.a1,
        value.B1.a0,
        value.B1.a1,
        value.B2.a0,
        value.B2.a1,
    };
    for (limbs) |limb| {
        if (remaining == 0) return true;
        const bits = @min(remaining, 31);
        const shift: u6 = @intCast(bits);
        const mask = (@as(u64, 1) << shift) - 1;
        if ((@as(u64, limb.value) & mask) != 0) return false;
        remaining -= bits;
    }
    return remaining == 0;
}

fn verifyQueries(
    params: types.Params,
    level_roots: []const types.Digest,
    level_ds: []const u32,
    proof: types.FriProof,
    query_positions: []const u32,
    challenges: ResolvedChallenges,
) FriError!void {
    var query_index_value: u32 = 0;
    while (query_index_value < params.num_queries) : (query_index_value += 1) {
        const proof_query_index: usize = @intCast(query_index_value);
        const position = query_positions[proof_query_index];
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
    challenges: ResolvedChallenges,
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
                const term = level_value.mul(challenges.gammas_base[level_index]);
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
    challenges: ResolvedChallenges,
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
                const term = level_value.mul(challenges.gammas_ext[level_index]);
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
