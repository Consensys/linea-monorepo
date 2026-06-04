const std = @import("std");
const verifier_ray = @import("verifier_ray");
const loom_vectors = @import("loom_test_vectors");

const field = verifier_ray.field.koalabear;
const ext = verifier_ray.field.koalabear_ext;
const fiat_shamir = verifier_ray.crypto.fiat_shamir;
const fri = verifier_ray.fri;

test "named transcript chains challenge digests" {
    var ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
    try ts.newChallenge("alpha");
    try ts.bindElements("alpha", &.{ elem(1), elem(2) });
    const alpha = try ts.computeChallenge("alpha");

    try ts.newChallenge("beta");
    try ts.bindElements("beta", &.{elem(1)});
    const beta = try ts.computeChallenge("beta");

    try std.testing.expect(!digestEql(alpha, beta));
    try std.testing.expect(digestEql(alpha, try ts.computeChallenge("alpha")));
}

test "named transcript rejects invalid proof of work" {
    var ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
    try ts.newChallenge("fri_fold_0");
    try ts.bindElements("fri_fold_0", &.{elem(3)});

    try std.testing.expectError(
        error.InvalidProofOfWork,
        ts.setProofOfWork("fri_fold_0", .{ .nb_bits = 32, .salt = elem(0) }),
    );
}

test "loom FRI leaf hashes match static vectors" {
    for (loom_vectors.loom_leaf_hash_cases) |case| {
        var base_pairs: [8]fri.PairBase = undefined;
        var ext_pairs: [8]fri.PairExt = undefined;
        fillBasePairs(&base_pairs, case.base_pairs);
        fillExtPairs(&ext_pairs, case.ext_pairs);

        try expectDigest(
            fri.leaf_hash.hashLeaf(base_pairs[0..case.base_pairs.len], ext_pairs[0..case.ext_pairs.len]),
            case.expected,
        );
    }
}

test "loom FRI node hashes match static vectors" {
    for (loom_vectors.loom_node_hash_cases) |case| {
        try expectDigest(fri.leaf_hash.hashNode(digest(case.left), digest(case.right)), case.expected);
    }
}

test "loom FRI merkle paths match static vectors" {
    for (loom_vectors.loom_merkle_cases) |case| {
        var siblings: [8]fri.Digest = undefined;
        fillDigests(&siblings, case.siblings);

        const path = fri.MerklePath{
            .leaf_idx = case.leaf_idx,
            .siblings = siblings[0..case.siblings.len],
        };
        try std.testing.expect(fri.merkle.merkleVerify(digest(case.root), path, digest(case.leaf)));
    }
}

test "loom named transcript challenges match static vectors" {
    for (loom_vectors.loom_named_transcript_cases) |case| {
        var ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
        var first_bindings: [16]field.Element = undefined;
        var second_bindings: [16]field.Element = undefined;

        try ts.newChallenge(case.first_name);
        fillElems(&first_bindings, case.first_bindings);
        try ts.bindElements(case.first_name, first_bindings[0..case.first_bindings.len]);
        try expectDigest(try ts.computeChallenge(case.first_name), case.first_expected);

        try ts.newChallenge(case.second_name);
        fillElems(&second_bindings, case.second_bindings);
        try ts.bindElements(case.second_name, second_bindings[0..case.second_bindings.len]);
        try expectDigest(try ts.computeChallenge(case.second_name), case.second_expected);
        try expectExt(try ts.computeChallengeExt(case.second_name), uintsToExt(case.second_ext_expected));
    }
}

test "loom proof-of-work transcript challenges match static vectors" {
    for (loom_vectors.loom_pow_transcript_cases) |case| {
        var ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
        var bindings: [16]field.Element = undefined;

        try ts.newChallenge(case.name);
        fillElems(&bindings, case.bindings);
        try ts.bindElements(case.name, bindings[0..case.bindings.len]);
        try ts.setProofOfWork(case.name, .{
            .nb_bits = case.nb_bits,
            .salt = elem(case.salt),
        });
        try expectDigest(try ts.computeChallenge(case.name), case.expected);
    }
}

test "fri leaf hash and merkle proof verify paired leaves" {
    const leaf0_base = [_]fri.PairBase{.{ elem(1), elem(2) }};
    const leaf1_ext = [_]fri.PairExt{.{ ext.Ext.fromUints(3, 4, 5, 6, 7, 8), ext.Ext.fromUints(8, 7, 6, 5, 4, 3) }};

    const leaf0 = fri.leaf_hash.hashLeaf(leaf0_base[0..], &.{});
    const leaf1 = fri.leaf_hash.hashLeaf(&.{}, leaf1_ext[0..]);
    const root = fri.leaf_hash.hashNode(leaf0, leaf1);

    const path0 = fri.MerklePath{ .leaf_idx = 0, .siblings = &.{leaf1} };
    const path1 = fri.MerklePath{ .leaf_idx = 1, .siblings = &.{leaf0} };

    try std.testing.expect(fri.merkle.merkleVerify(root, path0, leaf0));
    try std.testing.expect(fri.merkle.merkleVerify(root, path1, leaf1));
    try std.testing.expect(!fri.merkle.merkleVerify(root, path1, leaf0));

    const proof0 = fri.MerkleProof{
        .raw_leaf_base = leaf0_base[0..],
        .path = path0,
    };
    try std.testing.expect(fri.merkle.proofVerify(root, proof0));
}

test "point samplings verify every query tree proof" {
    const leaf_base = [_]fri.PairBase{.{ elem(11), elem(12) }};
    const root = fri.leaf_hash.hashLeaf(leaf_base[0..], &.{});
    const roots = [_]fri.Digest{root};
    const proof = fri.MerkleProof{
        .raw_leaf_base = leaf_base[0..],
        .path = .{ .leaf_idx = 0, .siblings = &.{} },
    };
    const row = [_]fri.MerkleProof{proof};
    const samplings = [_][]const fri.MerkleProof{row[0..]};

    try fri.bridge.checkPointSamplings(roots[0..], samplings[0..]);
}

test "fri verifier accepts empty-query shape and rejects missing grinding" {
    const root = zeroDigest();
    const roots = [_]fri.Digest{root};
    const level_ds = [_]u32{4};
    const gens = [_]field.Element{field.Element.one()};
    const pow = [_]?fri.ProofOfWork{null};
    const final_poly = [_]ext.Ext{ext.Ext.one()};
    const proof = fri.FriProof{
        .final_poly_ext = final_poly[0..],
        .pow = pow[0..],
    };
    const params = fri.Params{
        .n = 16,
        .d = 4,
        .num_queries = 0,
        .num_rounds = 1,
        .domain_gens = gens[0..],
        .domain_gens_inv = gens[0..],
        .grinding = 0,
    };

    var ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
    try fri.verify.friVerify(params, roots[0..], level_ds[0..], proof, &ts);

    var grinding_ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
    var grinding_params = params;
    grinding_params.grinding = 1;
    try std.testing.expectError(
        error.MissingProofOfWork,
        fri.verify.friVerify(grinding_params, roots[0..], level_ds[0..], proof, &grinding_ts),
    );
}

test "fri verifier reports unsupported before accepting unchecked queries" {
    const root = zeroDigest();
    const roots = [_]fri.Digest{root};
    const level_ds = [_]u32{4};
    const gens = [_]field.Element{field.Element.one()};
    const pow = [_]?fri.ProofOfWork{null};
    const final_poly = [_]ext.Ext{ext.Ext.one()};
    const layers = [_]fri.QueryLayer{.{
        .rail = .ext,
        .leaf_p_ext = ext.Ext.one(),
        .leaf_q_ext = ext.Ext.one(),
        .path = .{ .leaf_idx = 0, .siblings = &.{} },
    }};
    const queries = [_]fri.Query{.{ .layers = layers[0..] }};
    const proof = fri.FriProof{
        .final_poly_ext = final_poly[0..],
        .queries = queries[0..],
        .pow = pow[0..],
    };
    const params = fri.Params{
        .n = 16,
        .d = 4,
        .num_queries = 1,
        .num_rounds = 1,
        .domain_gens = gens[0..],
        .domain_gens_inv = gens[0..],
        .grinding = 0,
    };

    var ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
    try std.testing.expectError(
        error.Unsupported,
        fri.verify.friVerify(params, roots[0..], level_ds[0..], proof, &ts),
    );
}

test "fri verifier tags ext final poly before query challenge" {
    const root = zeroDigest();
    const roots = [_]fri.Digest{root};
    const level_ds = [_]u32{4};
    const gens = [_]field.Element{field.Element.one()};
    const pow = [_]?fri.ProofOfWork{null};
    const final_poly = [_]ext.Ext{
        ext.Ext.fromUints(2, 3, 5, 7, 11, 13),
        ext.Ext.fromUints(17, 19, 23, 29, 31, 37),
    };
    const layers = [_]fri.QueryLayer{.{
        .rail = .ext,
        .leaf_p_ext = ext.Ext.one(),
        .leaf_q_ext = ext.Ext.one(),
        .path = .{ .leaf_idx = 0, .siblings = &.{} },
    }};
    const queries = [_]fri.Query{.{ .layers = layers[0..] }};
    const proof = fri.FriProof{
        .final_poly_ext = final_poly[0..],
        .queries = queries[0..],
        .pow = pow[0..],
    };
    const params = fri.Params{
        .n = 16,
        .d = 4,
        .num_queries = 1,
        .num_rounds = 1,
        .domain_gens = gens[0..],
        .domain_gens_inv = gens[0..],
        .grinding = 0,
    };

    var ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
    try std.testing.expectError(error.Unsupported, fri.verify.friVerify(params, roots[0..], level_ds[0..], proof, &ts));
    const actual = try ts.computeChallenge("fri_query_0");

    var expected_ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
    try bindSingleRoundFold(&expected_ts, root);
    try expected_ts.newChallenge("fri_query_0");
    try expected_ts.bindElements("fri_query_0", &.{
        field.Element.init(0x45585450),
        field.Element.init(final_poly.len),
    });
    for (final_poly) |value| {
        var limbs = extLimbs(value);
        try expected_ts.bindElements("fri_query_0", limbs[0..]);
    }

    try std.testing.expect(digestEql(actual, try expected_ts.computeChallenge("fri_query_0")));
}

test "fri verifier tags base final poly before query challenge" {
    const root = zeroDigest();
    const roots = [_]fri.Digest{root};
    const level_ds = [_]u32{4};
    const gens = [_]field.Element{field.Element.one()};
    const pow = [_]?fri.ProofOfWork{null};
    const final_poly = [_]field.Element{ elem(41), elem(43), elem(47) };
    const layers = [_]fri.QueryLayer{.{
        .rail = .base,
        .leaf_p_base = elem(1),
        .leaf_q_base = elem(2),
        .path = .{ .leaf_idx = 0, .siblings = &.{} },
    }};
    const queries = [_]fri.Query{.{ .layers = layers[0..] }};
    const proof = fri.FriProof{
        .final_rail = .base,
        .final_poly_base = final_poly[0..],
        .queries = queries[0..],
        .pow = pow[0..],
    };
    const params = fri.Params{
        .n = 16,
        .d = 4,
        .num_queries = 1,
        .num_rounds = 1,
        .domain_gens = gens[0..],
        .domain_gens_inv = gens[0..],
        .grinding = 0,
    };

    var ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
    try std.testing.expectError(error.Unsupported, fri.verify.friVerify(params, roots[0..], level_ds[0..], proof, &ts));
    const actual = try ts.computeChallenge("fri_query_0");

    var expected_ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
    try bindSingleRoundFold(&expected_ts, root);
    try expected_ts.newChallenge("fri_query_0");
    try expected_ts.bindElements("fri_query_0", &.{
        field.Element.init(0x42415345),
        field.Element.init(final_poly.len),
    });
    try expected_ts.bindElements("fri_query_0", final_poly[0..]);

    try std.testing.expect(digestEql(actual, try expected_ts.computeChallenge("fri_query_0")));
}

test "fri verifier rejects duplicate level introduction rounds" {
    const roots = [_]fri.Digest{ zeroDigest(), zeroDigest(), zeroDigest() };
    const level_ds = [_]u32{ 8, 4, 4 };
    const fri_roots = [_]fri.Digest{ zeroDigest(), zeroDigest() };
    const gens = [_]field.Element{ field.Element.one(), field.Element.one(), field.Element.one() };
    const pow = [_]?fri.ProofOfWork{ null, null, null };
    const final_poly = [_]ext.Ext{ext.Ext.one()};
    const level_queries = [_][]const fri.QueryLayer{ &.{}, &.{} };
    const proof = fri.FriProof{
        .fri_roots = fri_roots[0..],
        .final_poly_ext = final_poly[0..],
        .level_queries = level_queries[0..],
        .pow = pow[0..],
    };
    const params = fri.Params{
        .n = 32,
        .d = 8,
        .num_queries = 0,
        .num_rounds = 3,
        .domain_gens = gens[0..],
        .domain_gens_inv = gens[0..],
        .grinding = 0,
    };

    var ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
    try std.testing.expectError(
        error.BadDimensions,
        fri.verify.friVerify(params, roots[0..], level_ds[0..], proof, &ts),
    );
}

fn elem(value: u32) field.Element {
    return field.Element.init(value);
}

fn uintsToExt(limbs: [6]u32) ext.Ext {
    return ext.Ext.fromUints(limbs[0], limbs[1], limbs[2], limbs[3], limbs[4], limbs[5]);
}

fn digest(values: [8]u32) fri.Digest {
    var out: fri.Digest = undefined;
    for (&out, values) |*dst, value| {
        dst.* = elem(value);
    }
    return out;
}

fn fillElems(out: []field.Element, values: []const u32) void {
    for (values, 0..) |value, i| {
        out[i] = elem(value);
    }
}

fn fillBasePairs(out: []fri.PairBase, values: []const [2]u32) void {
    for (values, 0..) |value, i| {
        out[i] = .{ elem(value[0]), elem(value[1]) };
    }
}

fn fillExtPairs(out: []fri.PairExt, values: []const [2][6]u32) void {
    for (values, 0..) |value, i| {
        out[i] = .{ uintsToExt(value[0]), uintsToExt(value[1]) };
    }
}

fn fillDigests(out: []fri.Digest, values: []const [8]u32) void {
    for (values, 0..) |value, i| {
        out[i] = digest(value);
    }
}

fn zeroDigest() fri.Digest {
    return .{
        field.Element.zero(),
        field.Element.zero(),
        field.Element.zero(),
        field.Element.zero(),
        field.Element.zero(),
        field.Element.zero(),
        field.Element.zero(),
        field.Element.zero(),
    };
}

fn bindSingleRoundFold(ts: *fiat_shamir.Transcript, root: fri.Digest) !void {
    try ts.newChallenge("fri_fold_0");
    try ts.bindDigest("fri_fold_0", root);
    _ = try ts.computeChallengeExt("fri_fold_0");
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

fn expectDigest(actual: fri.Digest, expected: [8]u32) !void {
    for (actual, expected) |actual_limb, expected_limb| {
        try std.testing.expectEqual(expected_limb, actual_limb.value);
    }
}

fn expectExt(actual: ext.Ext, expected: ext.Ext) !void {
    try std.testing.expect(actual.eql(expected));
}

fn digestEql(left: fri.Digest, right: fri.Digest) bool {
    for (left, right) |left_limb, right_limb| {
        if (!left_limb.eql(right_limb)) return false;
    }
    return true;
}
