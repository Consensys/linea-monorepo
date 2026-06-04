const std = @import("std");
const verifier_ray = @import("verifier_ray");

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

fn elem(value: u32) field.Element {
    return field.Element.init(value);
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

fn digestEql(left: fri.Digest, right: fri.Digest) bool {
    for (left, right) |left_limb, right_limb| {
        if (!left_limb.eql(right_limb)) return false;
    }
    return true;
}
