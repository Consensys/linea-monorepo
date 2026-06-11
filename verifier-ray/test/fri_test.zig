const std = @import("std");
const verifier_ray = @import("verifier_ray");
const loom_vectors = @import("loom_test_vectors");
const pcs_vectors = @import("pcs_test_vectors");

const field = verifier_ray.field.koalabear;
const ext = verifier_ray.field.koalabear_ext;
const fiat_shamir = verifier_ray.crypto.fiat_shamir;
const fri = verifier_ray.fri;
const layout_types = verifier_ray.layout;
const dq_layout_types = verifier_ray.dq_layout;

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

test "loom DEEP alpha binding order matches static vector" {
    const bridge_case = loom_vectors.loom_bridge_cases[0];
    const dq_layout = dq_layout_types.DQLayout{
        .sizes = bridge_case.sizes,
        .column_keys = bridge_case.column_keys,
        .air_chunks = bridge_case.air_chunks,
    };

    for (loom_vectors.loom_deep_alpha_cases) |case| {
        var values_at_zeta = std.StringHashMap(ext.Ext).init(std.testing.allocator);
        defer values_at_zeta.deinit();
        for (bridge_case.values_at_zeta) |value| {
            try values_at_zeta.put(value.name, uintsToExt(value.value));
        }

        var ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
        try ts.newChallenge(fri.bridge.final_evaluation_challenge);
        try ts.newChallenge(fri.bridge.deep_alpha_challenge);

        var zeta_bindings: [8]field.Element = undefined;
        try std.testing.expect(case.zeta_bindings.len <= zeta_bindings.len);
        fillElems(zeta_bindings[0..case.zeta_bindings.len], case.zeta_bindings);
        try ts.bindElements(fri.bridge.final_evaluation_challenge, zeta_bindings[0..case.zeta_bindings.len]);
        try expectDigest(try ts.computeChallenge(fri.bridge.final_evaluation_challenge), case.zeta_digest);

        const alpha = try fri.bridge.deriveDeepAlpha(dq_layout, &values_at_zeta, &ts);
        try expectExt(alpha, uintsToExt(case.alpha_ext));
        try expectDigest(try ts.computeChallenge(fri.bridge.deep_alpha_challenge), case.alpha_digest);

        _ = values_at_zeta.remove("base_col");
        var missing_ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
        try missing_ts.newChallenge(fri.bridge.deep_alpha_challenge);
        try std.testing.expectError(
            error.MissingValueAtZeta,
            fri.bridge.deriveDeepAlpha(dq_layout, &values_at_zeta, &missing_ts),
        );
    }
}

test "DEEP alpha rejects malformed DQ layout dimensions" {
    var values_at_zeta = std.StringHashMap(ext.Ext).init(std.testing.allocator);
    defer values_at_zeta.deinit();

    const sizes = [_]u32{4};
    const empty_key_levels = [_][]const []const []const u8{&.{}};
    const empty_air_levels = [_][]const []const u8{&.{}};

    try expectDeepAlphaBadDimensions(.{
        .sizes = &.{},
        .column_keys = &.{},
        .air_chunks = &.{},
    }, &values_at_zeta);

    try expectDeepAlphaBadDimensions(.{
        .sizes = sizes[0..],
        .column_keys = &.{},
        .air_chunks = empty_air_levels[0..],
    }, &values_at_zeta);

    try expectDeepAlphaBadDimensions(.{
        .sizes = sizes[0..],
        .column_keys = empty_key_levels[0..],
        .air_chunks = &.{},
    }, &values_at_zeta);
}

test "loom FRI base proof vector verifies" {
    for (loom_vectors.loom_fri_base_proof_cases) |case| {
        const round_count: usize = @intCast(case.num_rounds);
        const query_count: usize = @intCast(case.num_queries);
        try std.testing.expect(round_count <= 4);
        try std.testing.expect(query_count <= 4);
        try std.testing.expect(case.level_roots.len <= 4);
        try std.testing.expect(case.fri_roots.len <= 4);

        var level_roots: [4]fri.Digest = undefined;
        var fri_roots: [4]fri.Digest = undefined;
        var final_poly: [8]field.Element = undefined;
        var domain_gens: [4]field.Element = undefined;
        var domain_gens_inv: [4]field.Element = undefined;
        var pow: [4]?fri.ProofOfWork = undefined;
        var query_siblings: [4][4][4]fri.Digest = undefined;
        var query_layers: [4][4]fri.QueryLayer = undefined;
        var queries: [4]fri.Query = undefined;
        var level_siblings: [4][4][4]fri.Digest = undefined;
        var level_layers: [4][4]fri.QueryLayer = undefined;
        var level_query_rows: [4][]const fri.QueryLayer = undefined;

        fillDigests(level_roots[0..case.level_roots.len], case.level_roots);
        fillDigests(fri_roots[0..case.fri_roots.len], case.fri_roots);
        fillElems(final_poly[0..case.final_poly_base.len], case.final_poly_base);
        for (pow[0..round_count]) |*slot| {
            slot.* = null;
        }
        for (0..round_count) |round| {
            const shift: u5 = @intCast(round);
            const domain_size = case.n >> shift;
            domain_gens[round] = try field.rootOfUnityBy(@intCast(domain_size));
            domain_gens_inv[round] = domain_gens[round].inverse();
        }

        for (case.queries, 0..) |case_query, query_index| {
            try std.testing.expect(case_query.layers.len <= 4);
            for (case_query.layers, 0..) |case_layer, layer_index| {
                try std.testing.expect(case_layer.path.siblings.len <= 4);
                fillDigests(
                    query_siblings[query_index][layer_index][0..case_layer.path.siblings.len],
                    case_layer.path.siblings,
                );
                query_layers[query_index][layer_index] = .{
                    .rail = .base,
                    .leaf_p_base = elem(case_layer.leaf_p_base),
                    .leaf_q_base = elem(case_layer.leaf_q_base),
                    .path = .{
                        .leaf_idx = case_layer.path.leaf_idx,
                        .siblings = query_siblings[query_index][layer_index][0..case_layer.path.siblings.len],
                    },
                };
            }
            queries[query_index] = .{ .layers = query_layers[query_index][0..case_query.layers.len] };
        }

        for (case.level_queries, 0..) |case_level_queries, level_index| {
            try std.testing.expect(case_level_queries.len <= 4);
            for (case_level_queries, 0..) |case_layer, query_index| {
                try std.testing.expect(case_layer.path.siblings.len <= 4);
                fillDigests(
                    level_siblings[level_index][query_index][0..case_layer.path.siblings.len],
                    case_layer.path.siblings,
                );
                level_layers[level_index][query_index] = .{
                    .rail = .base,
                    .leaf_p_base = elem(case_layer.leaf_p_base),
                    .leaf_q_base = elem(case_layer.leaf_q_base),
                    .path = .{
                        .leaf_idx = case_layer.path.leaf_idx,
                        .siblings = level_siblings[level_index][query_index][0..case_layer.path.siblings.len],
                    },
                };
            }
            level_query_rows[level_index] = level_layers[level_index][0..case_level_queries.len];
        }

        const proof = fri.FriProof{
            .fri_roots = fri_roots[0..case.fri_roots.len],
            .final_rail = .base,
            .final_poly_base = final_poly[0..case.final_poly_base.len],
            .queries = queries[0..case.queries.len],
            .level_queries = level_query_rows[0..case.level_queries.len],
            .pow = pow[0..round_count],
        };
        const params = fri.Params{
            .n = case.n,
            .d = case.d,
            .num_queries = case.num_queries,
            .num_rounds = case.num_rounds,
            .domain_gens = domain_gens[0..round_count],
            .domain_gens_inv = domain_gens_inv[0..round_count],
            .grinding = 0,
        };

        var ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
        try fri.verify.friVerify(params, level_roots[0..case.level_roots.len], case.level_ds, proof, &ts);

        for (case.query_positions, 0..) |expected_position, query_index| {
            var name_buf: [32]u8 = undefined;
            const name = try std.fmt.bufPrint(&name_buf, "fri_query_{d}", .{query_index});
            const challenge = try ts.computeChallenge(name);
            try std.testing.expectEqual(expected_position, queryIndexForTest(challenge, case.n / 2));
        }

        if (case.level_queries.len > 0 and case.level_queries[0].len > 0) {
            level_layers[0][0].path.leaf_idx ^= 1;
            try expectFriVerifyError(error.BadDimensions, params, level_roots[0..case.level_roots.len], case.level_ds, proof);
            level_layers[0][0].path.leaf_idx ^= 1;
        }

        if (case.queries.len > 0 and case.queries[0].layers.len > 0) {
            query_layers[0][0].path.leaf_idx ^= 1;
            try expectFriVerifyError(error.BadDimensions, params, level_roots[0..case.level_roots.len], case.level_ds, proof);
            query_layers[0][0].path.leaf_idx ^= 1;
        }
    }
}

test "loom FRI ext proof vector verifies" {
    for (loom_vectors.loom_fri_ext_proof_cases) |case| {
        const round_count: usize = @intCast(case.num_rounds);
        const query_count: usize = @intCast(case.num_queries);
        try std.testing.expect(round_count <= 4);
        try std.testing.expect(query_count <= 4);
        try std.testing.expect(case.level_roots.len <= 4);
        try std.testing.expect(case.fri_roots.len <= 4);

        var level_roots: [4]fri.Digest = undefined;
        var fri_roots: [4]fri.Digest = undefined;
        var final_poly: [8]ext.Ext = undefined;
        var domain_gens: [4]field.Element = undefined;
        var domain_gens_inv: [4]field.Element = undefined;
        var pow: [4]?fri.ProofOfWork = undefined;
        var query_siblings: [4][4][4]fri.Digest = undefined;
        var query_layers: [4][4]fri.QueryLayer = undefined;
        var queries: [4]fri.Query = undefined;
        var level_siblings: [4][4][4]fri.Digest = undefined;
        var level_layers: [4][4]fri.QueryLayer = undefined;
        var level_query_rows: [4][]const fri.QueryLayer = undefined;

        fillDigests(level_roots[0..case.level_roots.len], case.level_roots);
        fillDigests(fri_roots[0..case.fri_roots.len], case.fri_roots);
        fillExtValues(final_poly[0..case.final_poly_ext.len], case.final_poly_ext);
        for (pow[0..round_count]) |*slot| {
            slot.* = null;
        }
        for (0..round_count) |round| {
            const shift: u5 = @intCast(round);
            const domain_size = case.n >> shift;
            domain_gens[round] = try field.rootOfUnityBy(@intCast(domain_size));
            domain_gens_inv[round] = domain_gens[round].inverse();
        }

        for (case.queries, 0..) |case_query, query_index| {
            try std.testing.expect(case_query.layers.len <= 4);
            for (case_query.layers, 0..) |case_layer, layer_index| {
                try std.testing.expect(case_layer.path.siblings.len <= 4);
                fillDigests(
                    query_siblings[query_index][layer_index][0..case_layer.path.siblings.len],
                    case_layer.path.siblings,
                );
                query_layers[query_index][layer_index] = .{
                    .rail = .ext,
                    .leaf_p_ext = uintsToExt(case_layer.leaf_p_ext),
                    .leaf_q_ext = uintsToExt(case_layer.leaf_q_ext),
                    .path = .{
                        .leaf_idx = case_layer.path.leaf_idx,
                        .siblings = query_siblings[query_index][layer_index][0..case_layer.path.siblings.len],
                    },
                };
            }
            queries[query_index] = .{ .layers = query_layers[query_index][0..case_query.layers.len] };
        }

        for (case.level_queries, 0..) |case_level_queries, level_index| {
            try std.testing.expect(case_level_queries.len <= 4);
            for (case_level_queries, 0..) |case_layer, query_index| {
                try std.testing.expect(case_layer.path.siblings.len <= 4);
                fillDigests(
                    level_siblings[level_index][query_index][0..case_layer.path.siblings.len],
                    case_layer.path.siblings,
                );
                level_layers[level_index][query_index] = .{
                    .rail = .ext,
                    .leaf_p_ext = uintsToExt(case_layer.leaf_p_ext),
                    .leaf_q_ext = uintsToExt(case_layer.leaf_q_ext),
                    .path = .{
                        .leaf_idx = case_layer.path.leaf_idx,
                        .siblings = level_siblings[level_index][query_index][0..case_layer.path.siblings.len],
                    },
                };
            }
            level_query_rows[level_index] = level_layers[level_index][0..case_level_queries.len];
        }

        const proof = fri.FriProof{
            .fri_roots = fri_roots[0..case.fri_roots.len],
            .final_rail = .ext,
            .final_poly_ext = final_poly[0..case.final_poly_ext.len],
            .queries = queries[0..case.queries.len],
            .level_queries = level_query_rows[0..case.level_queries.len],
            .pow = pow[0..round_count],
        };
        const params = fri.Params{
            .n = case.n,
            .d = case.d,
            .num_queries = case.num_queries,
            .num_rounds = case.num_rounds,
            .domain_gens = domain_gens[0..round_count],
            .domain_gens_inv = domain_gens_inv[0..round_count],
            .grinding = 0,
        };

        var ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
        try fri.verify.friVerify(params, level_roots[0..case.level_roots.len], case.level_ds, proof, &ts);

        for (case.query_positions, 0..) |expected_position, query_index| {
            var name_buf: [32]u8 = undefined;
            const name = try std.fmt.bufPrint(&name_buf, "fri_query_{d}", .{query_index});
            const challenge = try ts.computeChallenge(name);
            try std.testing.expectEqual(expected_position, queryIndexForTest(challenge, case.n / 2));
        }

        if (case.level_queries.len > 0 and case.level_queries[0].len > 0) {
            level_layers[0][0].path.leaf_idx ^= 1;
            try expectFriVerifyError(error.BadDimensions, params, level_roots[0..case.level_roots.len], case.level_ds, proof);
            level_layers[0][0].path.leaf_idx ^= 1;
        }

        if (case.queries.len > 0 and case.queries[0].layers.len > 0) {
            query_layers[0][0].path.leaf_idx ^= 1;
            try expectFriVerifyError(error.BadDimensions, params, level_roots[0..case.level_roots.len], case.level_ds, proof);
            query_layers[0][0].path.leaf_idx ^= 1;
        }
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

test "loom DEEP bridge vector verifies and detects tampering" {
    for (loom_vectors.loom_bridge_cases) |case| {
        try std.testing.expect(case.sizes.len <= 4);
        try std.testing.expect(case.queries.len != 0);
        try std.testing.expect(case.queries.len <= 4);

        var eval_points: [4][4]ext.Ext = undefined;
        var eval_point_rows: [4][]const ext.Ext = undefined;
        for (case.eval_points, 0..) |case_points, level_index| {
            try std.testing.expect(case_points.len <= 4);
            fillExtValues(eval_points[level_index][0..case_points.len], case_points);
            eval_point_rows[level_index] = eval_points[level_index][0..case_points.len];
        }

        const dq_layout = dq_layout_types.DQLayout{
            .sizes = case.sizes,
            .eval_points = eval_point_rows[0..case.eval_points.len],
            .column_names = case.column_names,
            .column_keys = case.column_keys,
            .air_chunks = case.air_chunks,
        };

        var values_at_zeta = std.StringHashMap(ext.Ext).init(std.testing.allocator);
        defer values_at_zeta.deinit();
        for (case.values_at_zeta) |value| {
            try values_at_zeta.put(value.name, uintsToExt(value.value));
        }

        var col_slots = std.StringHashMap(layout_types.Slot).init(std.testing.allocator);
        defer col_slots.deinit();
        for (case.col_slots) |slot| {
            try col_slots.put(slot.name, .{
                .tree_idx = slot.tree_idx,
                .poly_idx = slot.poly_idx,
                .rail = bridgeRail(slot.rail),
            });
        }

        var air_chunk_slots = std.StringHashMap(layout_types.Slot).init(std.testing.allocator);
        defer air_chunk_slots.deinit();
        for (case.air_chunk_slots) |slot| {
            try air_chunk_slots.put(slot.name, .{
                .tree_idx = slot.tree_idx,
                .poly_idx = slot.poly_idx,
                .rail = bridgeRail(slot.rail),
            });
        }

        const tree_count: u32 = @intCast(case.queries[0].point_samplings.len);
        try std.testing.expect(case.tree_sizes.len == @as(usize, @intCast(tree_count)));
        const layout = layout_types.Layout{
            .num_trees = tree_count,
            .air_begin = 0,
            .air_end = 0,
            .setup_begin = 0,
            .setup_end = 0,
            .tree_size = case.tree_sizes,
            .col_slot = col_slots,
            .air_chunk_slot = air_chunk_slots,
        };

        var raw_base: [4][8][4]fri.PairBase = undefined;
        var raw_ext: [4][8][4]fri.PairExt = undefined;
        var point_sampling_storage: [4][8]fri.MerkleProof = undefined;
        var point_sampling_rows: [4][]const fri.MerkleProof = undefined;
        var query_layers: [4][1]fri.QueryLayer = undefined;
        var queries: [4]fri.Query = undefined;
        var level_layers: [4][4]fri.QueryLayer = undefined;
        var level_query_rows: [4][]const fri.QueryLayer = undefined;

        const extra_levels = case.sizes.len - 1;
        for (case.queries, 0..) |case_query, query_index| {
            try std.testing.expect(case_query.point_samplings.len <= 8);
            try std.testing.expect(case_query.point_samplings.len == @as(usize, @intCast(tree_count)));
            try std.testing.expect(case_query.level_layers.len == extra_levels);

            for (case_query.point_samplings, 0..) |sampling, tree_index| {
                try std.testing.expect(sampling.base_pairs.len <= 4);
                try std.testing.expect(sampling.ext_pairs.len <= 4);
                fillBasePairs(raw_base[query_index][tree_index][0..sampling.base_pairs.len], sampling.base_pairs);
                fillExtPairs(raw_ext[query_index][tree_index][0..sampling.ext_pairs.len], sampling.ext_pairs);
                point_sampling_storage[query_index][tree_index] = .{
                    .raw_leaf_base = raw_base[query_index][tree_index][0..sampling.base_pairs.len],
                    .raw_leaf_ext = raw_ext[query_index][tree_index][0..sampling.ext_pairs.len],
                    .path = .{ .leaf_idx = sampling.leaf_idx, .siblings = &.{} },
                };
            }
            point_sampling_rows[query_index] = point_sampling_storage[query_index][0..case_query.point_samplings.len];

            query_layers[query_index][0] = bridgeLayer(case_query.fri_layer);
            queries[query_index] = .{ .layers = query_layers[query_index][0..1] };
        }

        for (0..extra_levels) |level_index| {
            for (case.queries, 0..) |case_query, query_index| {
                level_layers[level_index][query_index] = bridgeLayer(case_query.level_layers[level_index]);
            }
            level_query_rows[level_index] = level_layers[level_index][0..case.queries.len];
        }

        var level_roots: [4]fri.Digest = undefined;
        for (level_roots[0..case.sizes.len]) |*root| root.* = zeroDigest();
        const proof = fri.types.Proof{
            .deep_quotient_commitment = level_roots[0..case.sizes.len],
            .level_ds = case.sizes,
            .fri = .{
                .queries = queries[0..case.queries.len],
                .level_queries = level_query_rows[0..extra_levels],
            },
            .point_samplings = point_sampling_rows[0..case.queries.len],
        };

        const zeta = uintsToExt(case.zeta);
        const alpha = uintsToExt(case.alpha);
        try fri.bridge.checkFRIBridge(layout, dq_layout, proof, &values_at_zeta, zeta, alpha);

        const original_leaf_idx = point_sampling_storage[0][0].path.leaf_idx;
        point_sampling_storage[0][0].path.leaf_idx = original_leaf_idx + 1;
        try expectBridgeError(error.BadDimensions, layout, dq_layout, proof, &values_at_zeta, zeta, alpha);
        point_sampling_storage[0][0].path.leaf_idx = original_leaf_idx;

        const original_raw = raw_base[0][0][0][0];
        raw_base[0][0][0][0] = original_raw.add(elem(1));
        try expectBridgeError(error.BridgeMismatch, layout, dq_layout, proof, &values_at_zeta, zeta, alpha);
        raw_base[0][0][0][0] = original_raw;

        const original_value = values_at_zeta.get("base_col").?;
        try values_at_zeta.put("base_col", original_value.add(ext.Ext.one()));
        try expectBridgeError(error.BridgeMismatch, layout, dq_layout, proof, &values_at_zeta, zeta, alpha);
        try values_at_zeta.put("base_col", original_value);

        _ = values_at_zeta.remove("base_col");
        try expectBridgeError(error.MissingValueAtZeta, layout, dq_layout, proof, &values_at_zeta, zeta, alpha);
        try values_at_zeta.put("base_col", original_value);
    }
}

test "PCS integration verifies vectored end-to-end proof" {
    for (pcs_vectors.pcs_integration_cases) |case| {
        try std.testing.expect(case.num_rounds <= 4);
        try std.testing.expect(case.deep_quotient_commitment.len <= 4);
        try std.testing.expect(case.point_sampling_roots.len <= 4);
        try std.testing.expect(case.final_poly_ext.len <= 8);
        try std.testing.expect(case.point_sample_ext_pairs.len <= 4);
        try std.testing.expect(case.point_sample_siblings.len <= 4);
        try std.testing.expect(case.fri_layer_siblings.len <= 4);
        try std.testing.expect(case.air_chunk_slots.len <= 4);

        var values_at_zeta = std.StringHashMap(ext.Ext).init(std.testing.allocator);
        defer values_at_zeta.deinit();
        for (case.values_at_zeta) |value| {
            try values_at_zeta.put(value.name, uintsToExt(value.value));
        }

        var col_slots = std.StringHashMap(layout_types.Slot).init(std.testing.allocator);
        defer col_slots.deinit();
        var air_chunk_slots = std.StringHashMap(layout_types.Slot).init(std.testing.allocator);
        defer air_chunk_slots.deinit();
        for (case.air_chunk_slots) |slot| {
            try air_chunk_slots.put(slot.name, .{
                .tree_idx = slot.tree_idx,
                .poly_idx = slot.poly_idx,
                .rail = .ext,
            });
        }

        const layout = layout_types.Layout{
            .num_trees = @intCast(case.tree_sizes.len),
            .setup_begin = 0,
            .setup_end = 0,
            .air_begin = 0,
            .air_end = 0,
            .tree_size = case.tree_sizes,
            .col_slot = col_slots,
            .air_chunk_slot = air_chunk_slots,
        };

        var air_chunk_names: [4][]const u8 = undefined;
        for (case.air_chunk_slots, 0..) |slot, index| {
            air_chunk_names[index] = slot.name;
        }
        const eval_point_rows = [_][]const ext.Ext{&.{}};
        const column_name_rows = [_][]const []const []const u8{&.{}};
        const column_key_rows = [_][]const []const []const u8{&.{}};
        const air_chunk_rows = [_][]const []const u8{air_chunk_names[0..case.air_chunk_slots.len]};
        const dq_layout = dq_layout_types.DQLayout{
            .sizes = case.level_ds,
            .eval_points = eval_point_rows[0..],
            .column_names = column_name_rows[0..],
            .column_keys = column_key_rows[0..],
            .air_chunks = air_chunk_rows[0..],
        };

        var point_roots: [4]fri.Digest = undefined;
        var deep_quotient_roots: [4]fri.Digest = undefined;
        fillDigests(point_roots[0..case.point_sampling_roots.len], case.point_sampling_roots);
        fillDigests(deep_quotient_roots[0..case.deep_quotient_commitment.len], case.deep_quotient_commitment);

        var point_sample_siblings: [4]fri.Digest = undefined;
        var fri_layer_siblings: [4]fri.Digest = undefined;
        fillDigests(point_sample_siblings[0..case.point_sample_siblings.len], case.point_sample_siblings);
        fillDigests(fri_layer_siblings[0..case.fri_layer_siblings.len], case.fri_layer_siblings);

        var point_sample_ext_pairs: [4]fri.PairExt = undefined;
        fillExtPairs(point_sample_ext_pairs[0..case.point_sample_ext_pairs.len], case.point_sample_ext_pairs);
        var final_poly_ext: [8]ext.Ext = undefined;
        fillExtValues(final_poly_ext[0..case.final_poly_ext.len], case.final_poly_ext);

        const point_sampling = fri.MerkleProof{
            .raw_leaf_ext = point_sample_ext_pairs[0..case.point_sample_ext_pairs.len],
            .path = .{
                .leaf_idx = case.point_sample_leaf_idx,
                .siblings = point_sample_siblings[0..case.point_sample_siblings.len],
            },
        };
        const point_sampling_row = [_]fri.MerkleProof{point_sampling};
        const point_samplings = [_][]const fri.MerkleProof{point_sampling_row[0..]};

        const query_layer = fri.QueryLayer{
            .rail = .ext,
            .leaf_p_ext = uintsToExt(case.fri_layer_leaf_p_ext),
            .leaf_q_ext = uintsToExt(case.fri_layer_leaf_q_ext),
            .path = .{
                .leaf_idx = case.fri_layer_leaf_idx,
                .siblings = fri_layer_siblings[0..case.fri_layer_siblings.len],
            },
        };
        const query_layers = [_]fri.QueryLayer{query_layer};
        const queries = [_]fri.Query{.{ .layers = query_layers[0..] }};
        var pow: [4]?fri.ProofOfWork = undefined;
        for (pow[0..case.num_rounds]) |*slot| {
            slot.* = null;
        }
        const fri_proof = fri.FriProof{
            .final_rail = .ext,
            .final_poly_ext = final_poly_ext[0..case.final_poly_ext.len],
            .queries = queries[0..],
            .pow = pow[0..case.num_rounds],
        };
        const proof = fri.types.Proof{
            .deep_quotient_commitment = deep_quotient_roots[0..case.deep_quotient_commitment.len],
            .level_ds = case.level_ds,
            .fri = fri_proof,
            .point_samplings = point_samplings[0..],
        };

        var domain_gens: [4]field.Element = undefined;
        var domain_gens_inv: [4]field.Element = undefined;
        for (0..case.num_rounds) |round| {
            const shift: u5 = @intCast(round);
            const domain_size = case.n >> shift;
            domain_gens[round] = try field.rootOfUnityBy(@intCast(domain_size));
            domain_gens_inv[round] = domain_gens[round].inverse();
        }
        const params = fri.Params{
            .n = case.n,
            .d = case.d,
            .num_queries = case.num_queries,
            .num_rounds = case.num_rounds,
            .domain_gens = domain_gens[0..case.num_rounds],
            .domain_gens_inv = domain_gens_inv[0..case.num_rounds],
            .grinding = 0,
        };

        var ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
        const zeta = try preparePcsTranscript(case, &ts);
        try fri.pcs.verify(
            params,
            layout,
            dq_layout,
            point_roots[0..case.point_sampling_roots.len],
            proof,
            &values_at_zeta,
            zeta,
            &ts,
        );
        try expectDigest(try ts.computeChallenge(fri.bridge.deep_alpha_challenge), case.alpha_deep_digest);
        try expectExt(try ts.computeChallengeExt(fri.bridge.deep_alpha_challenge), uintsToExt(case.alpha_deep));
        try expectDigest(try ts.computeChallenge("fri_fold_0"), case.fri_alpha_digest);
        const query_digest = try ts.computeChallenge("fri_query_0");
        try expectDigest(query_digest, case.query_digest);
        try std.testing.expectEqual(case.query_position, queryIndexForTest(query_digest, case.n / 2));

        const final_poly_index: usize = @intCast(case.query_position);
        const original_final_poly = final_poly_ext[final_poly_index];
        final_poly_ext[final_poly_index] = original_final_poly.add(ext.Ext.one());
        try expectPcsVerifyError(
            error.BadDimensions,
            case,
            params,
            layout,
            dq_layout,
            point_roots[0..case.point_sampling_roots.len],
            proof,
            &values_at_zeta,
            zeta,
        );
        final_poly_ext[final_poly_index] = original_final_poly;

        const original_sample = point_sample_ext_pairs[0][0];
        point_sample_ext_pairs[0][0] = original_sample.add(ext.Ext.one());
        try expectPcsVerifyError(
            error.InvalidMerkleProof,
            case,
            params,
            layout,
            dq_layout,
            point_roots[0..case.point_sampling_roots.len],
            proof,
            &values_at_zeta,
            zeta,
        );
        point_sample_ext_pairs[0][0] = original_sample;

        try expectPcsVerifyError(
            error.BridgeMismatch,
            case,
            params,
            layout,
            dq_layout,
            point_roots[0..case.point_sampling_roots.len],
            proof,
            &values_at_zeta,
            zeta.add(ext.Ext.one()),
        );

        _ = values_at_zeta.remove("air1");
        try expectPcsVerifyError(
            error.MissingValueAtZeta,
            case,
            params,
            layout,
            dq_layout,
            point_roots[0..case.point_sampling_roots.len],
            proof,
            &values_at_zeta,
            zeta,
        );
    }
}

test "loom PCS proof vector verifies composed pipeline" {
    for (pcs_vectors.loom_pcs_cases) |case| {
        try std.testing.expect(case.num_rounds <= 4);
        try std.testing.expect(case.num_queries <= 4);
        try std.testing.expect(case.roots.len <= 8);
        try std.testing.expect(case.deep_quotient_commitment.len <= 4);
        try std.testing.expect(case.fri_roots.len <= 4);
        try std.testing.expect(case.final_poly_ext.len <= 8);
        try std.testing.expect(case.queries.len <= 4);
        try std.testing.expect(case.level_queries.len <= 4);
        try std.testing.expect(case.point_samplings.len <= 4);
        try std.testing.expect(case.dq_eval_points.len <= 4);

        var roots: [8]fri.Digest = undefined;
        var deep_quotient_roots: [4]fri.Digest = undefined;
        var fri_roots: [4]fri.Digest = undefined;
        var final_poly_ext: [8]ext.Ext = undefined;
        fillDigests(roots[0..case.roots.len], case.roots);
        fillDigests(deep_quotient_roots[0..case.deep_quotient_commitment.len], case.deep_quotient_commitment);
        fillDigests(fri_roots[0..case.fri_roots.len], case.fri_roots);
        fillExtValues(final_poly_ext[0..case.final_poly_ext.len], case.final_poly_ext);

        var values_at_zeta = std.StringHashMap(ext.Ext).init(std.testing.allocator);
        defer values_at_zeta.deinit();
        for (case.values_at_zeta) |value| {
            try values_at_zeta.put(value.name, uintsToExt(value.value));
        }

        var col_slots = std.StringHashMap(layout_types.Slot).init(std.testing.allocator);
        defer col_slots.deinit();
        for (case.col_slots) |slot| {
            try col_slots.put(slot.name, .{
                .tree_idx = slot.tree_idx,
                .poly_idx = slot.poly_idx,
                .rail = pcsRail(slot.rail),
            });
        }

        var air_chunk_slots = std.StringHashMap(layout_types.Slot).init(std.testing.allocator);
        defer air_chunk_slots.deinit();
        for (case.air_chunk_slots) |slot| {
            try air_chunk_slots.put(slot.name, .{
                .tree_idx = slot.tree_idx,
                .poly_idx = slot.poly_idx,
                .rail = pcsRail(slot.rail),
            });
        }

        const layout = layout_types.Layout{
            .num_trees = @intCast(case.tree_sizes.len),
            .setup_begin = 0,
            .setup_end = 0,
            .air_begin = 0,
            .air_end = 0,
            .tree_size = case.tree_sizes,
            .col_slot = col_slots,
            .air_chunk_slot = air_chunk_slots,
        };

        var eval_points: [4][4]ext.Ext = undefined;
        var eval_point_rows: [4][]const ext.Ext = undefined;
        for (case.dq_eval_points, 0..) |case_points, level_index| {
            try std.testing.expect(case_points.len <= 4);
            fillExtValues(eval_points[level_index][0..case_points.len], case_points);
            eval_point_rows[level_index] = eval_points[level_index][0..case_points.len];
        }
        const dq_layout = dq_layout_types.DQLayout{
            .sizes = case.level_ds,
            .eval_points = eval_point_rows[0..case.dq_eval_points.len],
            .column_names = case.dq_column_names,
            .column_keys = case.dq_column_keys,
            .air_chunks = case.dq_air_chunks,
        };

        var point_raw_base: [4][4][4]fri.PairBase = undefined;
        var point_raw_ext: [4][4][4]fri.PairExt = undefined;
        var point_siblings: [4][4][4]fri.Digest = undefined;
        var point_sampling_storage: [4][4]fri.MerkleProof = undefined;
        var point_sampling_rows: [4][]const fri.MerkleProof = undefined;
        for (case.point_samplings, 0..) |case_row, query_index| {
            try std.testing.expect(case_row.len <= 4);
            for (case_row, 0..) |sampling, tree_index| {
                try std.testing.expect(sampling.base_pairs.len <= 4);
                try std.testing.expect(sampling.ext_pairs.len <= 4);
                try std.testing.expect(sampling.path.siblings.len <= 4);
                fillBasePairs(point_raw_base[query_index][tree_index][0..sampling.base_pairs.len], sampling.base_pairs);
                fillExtPairs(point_raw_ext[query_index][tree_index][0..sampling.ext_pairs.len], sampling.ext_pairs);
                fillDigests(point_siblings[query_index][tree_index][0..sampling.path.siblings.len], sampling.path.siblings);
                point_sampling_storage[query_index][tree_index] = .{
                    .raw_leaf_base = point_raw_base[query_index][tree_index][0..sampling.base_pairs.len],
                    .raw_leaf_ext = point_raw_ext[query_index][tree_index][0..sampling.ext_pairs.len],
                    .path = .{
                        .leaf_idx = sampling.path.leaf_idx,
                        .siblings = point_siblings[query_index][tree_index][0..sampling.path.siblings.len],
                    },
                };
            }
            point_sampling_rows[query_index] = point_sampling_storage[query_index][0..case_row.len];
        }

        var query_siblings: [4][4][4]fri.Digest = undefined;
        var query_layers: [4][4]fri.QueryLayer = undefined;
        var queries: [4]fri.Query = undefined;
        for (case.queries, 0..) |case_query, query_index| {
            try std.testing.expect(case_query.layers.len <= 4);
            for (case_query.layers, 0..) |case_layer, layer_index| {
                try std.testing.expect(case_layer.path.siblings.len <= 4);
                fillDigests(
                    query_siblings[query_index][layer_index][0..case_layer.path.siblings.len],
                    case_layer.path.siblings,
                );
                query_layers[query_index][layer_index] = pcsLayer(
                    case_layer,
                    query_siblings[query_index][layer_index][0..case_layer.path.siblings.len],
                );
            }
            queries[query_index] = .{ .layers = query_layers[query_index][0..case_query.layers.len] };
        }

        var level_siblings: [4][4][4]fri.Digest = undefined;
        var level_layers: [4][4]fri.QueryLayer = undefined;
        var level_query_rows: [4][]const fri.QueryLayer = undefined;
        for (case.level_queries, 0..) |case_level_queries, level_index| {
            try std.testing.expect(case_level_queries.len <= 4);
            for (case_level_queries, 0..) |case_layer, query_index| {
                try std.testing.expect(case_layer.path.siblings.len <= 4);
                fillDigests(
                    level_siblings[level_index][query_index][0..case_layer.path.siblings.len],
                    case_layer.path.siblings,
                );
                level_layers[level_index][query_index] = pcsLayer(
                    case_layer,
                    level_siblings[level_index][query_index][0..case_layer.path.siblings.len],
                );
            }
            level_query_rows[level_index] = level_layers[level_index][0..case_level_queries.len];
        }

        var pow: [4]?fri.ProofOfWork = undefined;
        for (pow[0..case.num_rounds]) |*slot| {
            slot.* = null;
        }
        const fri_proof = fri.FriProof{
            .fri_roots = fri_roots[0..case.fri_roots.len],
            .final_rail = .ext,
            .final_poly_ext = final_poly_ext[0..case.final_poly_ext.len],
            .queries = queries[0..case.queries.len],
            .level_queries = level_query_rows[0..case.level_queries.len],
            .pow = pow[0..case.num_rounds],
        };
        const proof = fri.types.Proof{
            .deep_quotient_commitment = deep_quotient_roots[0..case.deep_quotient_commitment.len],
            .level_ds = case.level_ds,
            .fri = fri_proof,
            .point_samplings = point_sampling_rows[0..case.point_samplings.len],
        };

        var domain_gens: [4]field.Element = undefined;
        var domain_gens_inv: [4]field.Element = undefined;
        for (0..case.num_rounds) |round| {
            const shift: u5 = @intCast(round);
            const domain_size = case.n >> shift;
            domain_gens[round] = try field.rootOfUnityBy(@intCast(domain_size));
            domain_gens_inv[round] = domain_gens[round].inverse();
        }
        const params = fri.Params{
            .n = case.n,
            .d = case.d,
            .num_queries = case.num_queries,
            .num_rounds = case.num_rounds,
            .domain_gens = domain_gens[0..case.num_rounds],
            .domain_gens_inv = domain_gens_inv[0..case.num_rounds],
            .grinding = 0,
        };

        var ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
        const zeta = try prepareLoomPcsTranscript(case, roots[0..case.roots.len], &ts);
        try fri.pcs.verify(
            params,
            layout,
            dq_layout,
            roots[0..case.roots.len],
            proof,
            &values_at_zeta,
            zeta,
            &ts,
        );
        try expectDigest(try ts.computeChallenge(fri.bridge.deep_alpha_challenge), case.alpha_deep_digest);
        try expectExt(try ts.computeChallengeExt(fri.bridge.deep_alpha_challenge), uintsToExt(case.alpha_deep));
        for (case.fri_fold_digests, 0..) |expected_digest, round| {
            var name_buf: [32]u8 = undefined;
            const name = try std.fmt.bufPrint(&name_buf, "fri_fold_{d}", .{round});
            try expectDigest(try ts.computeChallenge(name), expected_digest);
        }
        for (case.query_digests, case.query_positions, 0..) |expected_digest, expected_position, query_index| {
            var name_buf: [32]u8 = undefined;
            const name = try std.fmt.bufPrint(&name_buf, "fri_query_{d}", .{query_index});
            const query_digest = try ts.computeChallenge(name);
            try expectDigest(query_digest, expected_digest);
            try std.testing.expectEqual(expected_position, queryIndexForTest(query_digest, case.n / 2));
        }

        const original_sample = point_raw_base[0][0][0][0];
        point_raw_base[0][0][0][0] = original_sample.add(elem(1));
        try expectLoomPcsVerifyError(
            error.InvalidMerkleProof,
            case,
            params,
            layout,
            dq_layout,
            roots[0..case.roots.len],
            proof,
            &values_at_zeta,
            zeta,
        );
        point_raw_base[0][0][0][0] = original_sample;

        try expectLoomPcsVerifyError(
            error.BridgeMismatch,
            case,
            params,
            layout,
            dq_layout,
            roots[0..case.roots.len],
            proof,
            &values_at_zeta,
            zeta.add(ext.Ext.one()),
        );
    }
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

test "fri verifier accepts one-round base query" {
    const domain_gen = try field.rootOfUnityBy(4);
    const domain_gens = [_]field.Element{domain_gen};
    const domain_gens_inv = [_]field.Element{domain_gen.inverse()};
    const level_ds = [_]u32{2};
    const pow = [_]?fri.ProofOfWork{null};
    const leaf_pairs = [_]fri.PairBase{
        .{ elem(3), elem(9) },
        .{ elem(12), elem(4) },
    };
    const leaf0 = hashBasePairForTest(leaf_pairs[0]);
    const leaf1 = hashBasePairForTest(leaf_pairs[1]);
    const root = fri.leaf_hash.hashNode(leaf0, leaf1);
    const roots = [_]fri.Digest{root};

    var expected_ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
    const alpha_digest = try bindSingleRoundFoldDigest(&expected_ts, root);
    const alpha = alpha_digest[0];
    const final_poly = [_]field.Element{
        foldBasePairForTest(leaf_pairs[0][0], leaf_pairs[0][1], alpha, field.Element.one()),
        foldBasePairForTest(leaf_pairs[1][0], leaf_pairs[1][1], alpha, domain_gen.inverse()),
    };

    const expected_query = try bindBaseFinalPolyAndQuery(&expected_ts, final_poly[0..]);
    const query_pos = queryIndexForTest(expected_query, 2);
    const query_pos_usize: usize = @intCast(query_pos);
    const sibling = [_]fri.Digest{if (query_pos == 0) leaf1 else leaf0};
    const layers = [_]fri.QueryLayer{.{
        .rail = .base,
        .leaf_p_base = leaf_pairs[query_pos_usize][0],
        .leaf_q_base = leaf_pairs[query_pos_usize][1],
        .path = .{ .leaf_idx = query_pos, .siblings = sibling[0..] },
    }};
    const queries = [_]fri.Query{.{ .layers = layers[0..] }};
    const proof = fri.FriProof{
        .final_rail = .base,
        .final_poly_base = final_poly[0..],
        .queries = queries[0..],
        .pow = pow[0..],
    };
    const params = fri.Params{
        .n = 4,
        .d = 2,
        .num_queries = 1,
        .num_rounds = 1,
        .domain_gens = domain_gens[0..],
        .domain_gens_inv = domain_gens_inv[0..],
        .grinding = 0,
    };

    var ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
    try fri.verify.friVerify(params, roots[0..], level_ds[0..], proof, &ts);
    try std.testing.expect(digestEql(expected_query, try ts.computeChallenge("fri_query_0")));
}

test "fri verifier accepts one-round ext query" {
    const domain_gen = try field.rootOfUnityBy(4);
    const domain_gens = [_]field.Element{domain_gen};
    const domain_gens_inv = [_]field.Element{domain_gen.inverse()};
    const level_ds = [_]u32{2};
    const pow = [_]?fri.ProofOfWork{null};
    const leaf_pairs = [_]fri.PairExt{
        .{ ext.Ext.fromUints(2, 3, 5, 7, 11, 13), ext.Ext.fromUints(17, 19, 23, 29, 31, 37) },
        .{ ext.Ext.fromUints(41, 43, 47, 53, 59, 61), ext.Ext.fromUints(67, 71, 73, 79, 83, 89) },
    };
    const leaf0 = hashExtPairForTest(leaf_pairs[0]);
    const leaf1 = hashExtPairForTest(leaf_pairs[1]);
    const root = fri.leaf_hash.hashNode(leaf0, leaf1);
    const roots = [_]fri.Digest{root};

    var expected_ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
    const alpha_digest = try bindSingleRoundFoldDigest(&expected_ts, root);
    const alpha = uintsToExt(.{
        alpha_digest[0].value,
        alpha_digest[1].value,
        alpha_digest[2].value,
        alpha_digest[3].value,
        alpha_digest[4].value,
        alpha_digest[5].value,
    });
    const final_poly = [_]ext.Ext{
        foldExtPairForTest(leaf_pairs[0][0], leaf_pairs[0][1], alpha, field.Element.one()),
        foldExtPairForTest(leaf_pairs[1][0], leaf_pairs[1][1], alpha, domain_gen.inverse()),
    };

    const expected_query = try bindExtFinalPolyAndQuery(&expected_ts, final_poly[0..]);
    const query_pos = queryIndexForTest(expected_query, 2);
    const query_pos_usize: usize = @intCast(query_pos);
    const sibling = [_]fri.Digest{if (query_pos == 0) leaf1 else leaf0};
    const layers = [_]fri.QueryLayer{.{
        .rail = .ext,
        .leaf_p_ext = leaf_pairs[query_pos_usize][0],
        .leaf_q_ext = leaf_pairs[query_pos_usize][1],
        .path = .{ .leaf_idx = query_pos, .siblings = sibling[0..] },
    }};
    const queries = [_]fri.Query{.{ .layers = layers[0..] }};
    const proof = fri.FriProof{
        .final_poly_ext = final_poly[0..],
        .queries = queries[0..],
        .pow = pow[0..],
    };
    const params = fri.Params{
        .n = 4,
        .d = 2,
        .num_queries = 1,
        .num_rounds = 1,
        .domain_gens = domain_gens[0..],
        .domain_gens_inv = domain_gens_inv[0..],
        .grinding = 0,
    };

    var ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
    try fri.verify.friVerify(params, roots[0..], level_ds[0..], proof, &ts);
    try std.testing.expect(digestEql(expected_query, try ts.computeChallenge("fri_query_0")));
}

test "fri verifier mixes extra level during query fold" {
    const domain_gen_0 = try field.rootOfUnityBy(8);
    const domain_gen_1 = try field.rootOfUnityBy(4);
    const domain_gens = [_]field.Element{ domain_gen_0, domain_gen_1 };
    const domain_gens_inv = [_]field.Element{ domain_gen_0.inverse(), domain_gen_1.inverse() };
    const level_ds = [_]u32{ 4, 2 };
    const pow = [_]?fri.ProofOfWork{ null, null };

    const layer0_pairs = [_]fri.PairBase{
        .{ elem(2), elem(11) },
        .{ elem(5), elem(13) },
        .{ elem(7), elem(17) },
        .{ elem(19), elem(23) },
    };
    const layer0_leaves = [_]fri.Digest{
        hashBasePairForTest(layer0_pairs[0]),
        hashBasePairForTest(layer0_pairs[1]),
        hashBasePairForTest(layer0_pairs[2]),
        hashBasePairForTest(layer0_pairs[3]),
    };
    const root0 = merkleRoot4(layer0_leaves);

    var expected_ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
    const alpha0_digest = try bindSingleRoundFoldDigest(&expected_ts, root0);
    const alpha0 = alpha0_digest[0];
    var folded0: [4]field.Element = undefined;
    for (&folded0, layer0_pairs, 0..) |*dst, pair, i| {
        dst.* = foldBasePairForTest(pair[0], pair[1], alpha0, domain_gens_inv[0].pow(@intCast(i)));
    }

    const extra_evals = [_]field.Element{ elem(29), elem(31), elem(37), elem(41) };
    const extra_pairs = [_]fri.PairBase{
        .{ extra_evals[0], extra_evals[2] },
        .{ extra_evals[1], extra_evals[3] },
    };
    const extra_leaves = [_]fri.Digest{
        hashBasePairForTest(extra_pairs[0]),
        hashBasePairForTest(extra_pairs[1]),
    };
    const extra_root = fri.leaf_hash.hashNode(extra_leaves[0], extra_leaves[1]);

    try expected_ts.newChallenge("fri_level_1_gamma");
    try expected_ts.bindDigest("fri_level_1_gamma", extra_root);
    const gamma_digest = try expected_ts.computeChallenge("fri_level_1_gamma");
    const gamma = gamma_digest[0];

    var mixed: [4]field.Element = undefined;
    for (&mixed, folded0, extra_evals) |*dst, folded, extra_value| {
        dst.* = folded.add(extra_value.mul(gamma));
    }
    const layer1_pairs = [_]fri.PairBase{
        .{ mixed[0], mixed[2] },
        .{ mixed[1], mixed[3] },
    };
    const layer1_leaves = [_]fri.Digest{
        hashBasePairForTest(layer1_pairs[0]),
        hashBasePairForTest(layer1_pairs[1]),
    };
    const root1 = fri.leaf_hash.hashNode(layer1_leaves[0], layer1_leaves[1]);

    try expected_ts.newChallenge("fri_fold_1");
    try expected_ts.bindDigest("fri_fold_1", root1);
    const alpha1_digest = try expected_ts.computeChallenge("fri_fold_1");
    const alpha1 = alpha1_digest[0];
    const final_poly = [_]field.Element{
        foldBasePairForTest(layer1_pairs[0][0], layer1_pairs[0][1], alpha1, field.Element.one()),
        foldBasePairForTest(layer1_pairs[1][0], layer1_pairs[1][1], alpha1, domain_gens_inv[1]),
    };

    const expected_query = try bindBaseFinalPolyAndQuery(&expected_ts, final_poly[0..]);
    const query_pos = queryIndexForTest(expected_query, 4);
    const layer0_base = query_pos % 4;
    const layer1_base = query_pos % 2;
    const layer0_path_siblings = merklePath4(layer0_leaves, layer0_base);
    const layer1_path_siblings = merklePath2(layer1_leaves, layer1_base);
    const extra_path_siblings = merklePath2(extra_leaves, layer1_base);

    const query_layers = [_]fri.QueryLayer{
        .{
            .rail = .base,
            .leaf_p_base = layer0_pairs[@intCast(layer0_base)][0],
            .leaf_q_base = layer0_pairs[@intCast(layer0_base)][1],
            .path = .{ .leaf_idx = layer0_base, .siblings = layer0_path_siblings[0..] },
        },
        .{
            .rail = .base,
            .leaf_p_base = layer1_pairs[@intCast(layer1_base)][0],
            .leaf_q_base = layer1_pairs[@intCast(layer1_base)][1],
            .path = .{ .leaf_idx = layer1_base, .siblings = layer1_path_siblings[0..] },
        },
    };
    const queries = [_]fri.Query{.{ .layers = query_layers[0..] }};
    const extra_query_layers = [_]fri.QueryLayer{.{
        .rail = .base,
        .leaf_p_base = extra_pairs[@intCast(layer1_base)][0],
        .leaf_q_base = extra_pairs[@intCast(layer1_base)][1],
        .path = .{ .leaf_idx = layer1_base, .siblings = extra_path_siblings[0..] },
    }};
    const level_queries = [_][]const fri.QueryLayer{extra_query_layers[0..]};
    const roots = [_]fri.Digest{ root0, extra_root };
    const fri_roots = [_]fri.Digest{root1};
    const proof = fri.FriProof{
        .fri_roots = fri_roots[0..],
        .final_rail = .base,
        .final_poly_base = final_poly[0..],
        .queries = queries[0..],
        .level_queries = level_queries[0..],
        .pow = pow[0..],
    };
    const params = fri.Params{
        .n = 8,
        .d = 4,
        .num_queries = 1,
        .num_rounds = 2,
        .domain_gens = domain_gens[0..],
        .domain_gens_inv = domain_gens_inv[0..],
        .grinding = 0,
    };

    var ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
    try fri.verify.friVerify(params, roots[0..], level_ds[0..], proof, &ts);
    try std.testing.expect(digestEql(expected_query, try ts.computeChallenge("fri_query_0")));
}

test "fri verifier rejects final fold mismatch" {
    const domain_gen = try field.rootOfUnityBy(4);
    const domain_gens = [_]field.Element{domain_gen};
    const domain_gens_inv = [_]field.Element{domain_gen.inverse()};
    const level_ds = [_]u32{2};
    const pow = [_]?fri.ProofOfWork{null};
    const leaf_pairs = [_]fri.PairBase{
        .{ elem(6), elem(7) },
        .{ elem(8), elem(10) },
    };
    const leaf0 = hashBasePairForTest(leaf_pairs[0]);
    const leaf1 = hashBasePairForTest(leaf_pairs[1]);
    const root = fri.leaf_hash.hashNode(leaf0, leaf1);
    const roots = [_]fri.Digest{root};

    var expected_ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
    _ = try bindSingleRoundFoldDigest(&expected_ts, root);
    const final_poly = [_]field.Element{ elem(123), elem(456) };

    const expected_query = try bindBaseFinalPolyAndQuery(&expected_ts, final_poly[0..]);
    const query_pos = queryIndexForTest(expected_query, 2);
    const query_pos_usize: usize = @intCast(query_pos);
    const sibling = [_]fri.Digest{if (query_pos == 0) leaf1 else leaf0};
    const layers = [_]fri.QueryLayer{.{
        .rail = .base,
        .leaf_p_base = leaf_pairs[query_pos_usize][0],
        .leaf_q_base = leaf_pairs[query_pos_usize][1],
        .path = .{ .leaf_idx = query_pos, .siblings = sibling[0..] },
    }};
    const queries = [_]fri.Query{.{ .layers = layers[0..] }};
    const proof = fri.FriProof{
        .final_rail = .base,
        .final_poly_base = final_poly[0..],
        .queries = queries[0..],
        .pow = pow[0..],
    };
    const params = fri.Params{
        .n = 4,
        .d = 2,
        .num_queries = 1,
        .num_rounds = 1,
        .domain_gens = domain_gens[0..],
        .domain_gens_inv = domain_gens_inv[0..],
        .grinding = 0,
    };

    var ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
    try std.testing.expectError(
        error.FoldMismatch,
        fri.verify.friVerify(params, roots[0..], level_ds[0..], proof, &ts),
    );
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

test "fri verifier rejects non-power-of-two extra level degree" {
    const roots = [_]fri.Digest{ zeroDigest(), zeroDigest() };
    const level_ds = [_]u32{ 12, 3 };
    const fri_roots = [_]fri.Digest{ zeroDigest(), zeroDigest() };
    const gens = [_]field.Element{ field.Element.one(), field.Element.one(), field.Element.one() };
    const pow = [_]?fri.ProofOfWork{ null, null, null };
    const final_poly = [_]ext.Ext{ext.Ext.one()};
    const level_queries = [_][]const fri.QueryLayer{&.{}};
    const proof = fri.FriProof{
        .fri_roots = fri_roots[0..],
        .final_poly_ext = final_poly[0..],
        .level_queries = level_queries[0..],
        .pow = pow[0..],
    };
    const params = fri.Params{
        .n = 32,
        .d = 12,
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

fn fillExtValues(out: []ext.Ext, values: []const [6]u32) void {
    for (values, 0..) |value, i| {
        out[i] = uintsToExt(value);
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

fn bindSingleRoundFoldDigest(ts: *fiat_shamir.Transcript, root: fri.Digest) !fri.Digest {
    try ts.newChallenge("fri_fold_0");
    try ts.bindDigest("fri_fold_0", root);
    return ts.computeChallenge("fri_fold_0");
}

fn bindBaseFinalPolyAndQuery(ts: *fiat_shamir.Transcript, final_poly: []const field.Element) !fri.Digest {
    try ts.newChallenge("fri_query_0");
    try ts.bindElements("fri_query_0", &.{
        field.Element.init(0x42415345),
        field.Element.init(final_poly.len),
    });
    try ts.bindElements("fri_query_0", final_poly);
    return ts.computeChallenge("fri_query_0");
}

fn bindExtFinalPolyAndQuery(ts: *fiat_shamir.Transcript, final_poly: []const ext.Ext) !fri.Digest {
    try ts.newChallenge("fri_query_0");
    try ts.bindElements("fri_query_0", &.{
        field.Element.init(0x45585450),
        field.Element.init(final_poly.len),
    });
    for (final_poly) |value| {
        var limbs = extLimbs(value);
        try ts.bindElements("fri_query_0", limbs[0..]);
    }
    return ts.computeChallenge("fri_query_0");
}

fn hashBasePairForTest(pair: fri.PairBase) fri.Digest {
    const pairs = [_]fri.PairBase{pair};
    return fri.leaf_hash.hashLeaf(pairs[0..], &.{});
}

fn hashExtPairForTest(pair: fri.PairExt) fri.Digest {
    const pairs = [_]fri.PairExt{pair};
    return fri.leaf_hash.hashLeaf(&.{}, pairs[0..]);
}

fn merkleRoot4(leaves: [4]fri.Digest) fri.Digest {
    const left = fri.leaf_hash.hashNode(leaves[0], leaves[1]);
    const right = fri.leaf_hash.hashNode(leaves[2], leaves[3]);
    return fri.leaf_hash.hashNode(left, right);
}

fn merklePath2(leaves: [2]fri.Digest, index: u32) [1]fri.Digest {
    return .{if (index == 0) leaves[1] else leaves[0]};
}

fn merklePath4(leaves: [4]fri.Digest, index: u32) [2]fri.Digest {
    const left = fri.leaf_hash.hashNode(leaves[0], leaves[1]);
    const right = fri.leaf_hash.hashNode(leaves[2], leaves[3]);
    return .{
        if ((index & 1) == 0) leaves[@as(usize, @intCast(index + 1))] else leaves[@as(usize, @intCast(index - 1))],
        if (index < 2) right else left,
    };
}

fn foldBasePairForTest(p: field.Element, q: field.Element, alpha: field.Element, x_inv: field.Element) field.Element {
    return p.add(q).halve().add(p.sub(q).halve().mul(x_inv).mul(alpha));
}

fn foldExtPairForTest(p: ext.Ext, q: ext.Ext, alpha: ext.Ext, x_inv: field.Element) ext.Ext {
    return p.add(q).halve().add(p.sub(q).halve().mulByBase(x_inv).mul(alpha));
}

fn queryIndexForTest(challenge: fri.Digest, modulus: u32) u32 {
    if (modulus == 0) return 0;
    const wide = (@as(u64, challenge[0].value) << 31) ^ @as(u64, challenge[1].value);
    return @intCast(wide % modulus);
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

fn bridgeRail(rail: loom_vectors.LoomBridgeRail) fri.Rail {
    return switch (rail) {
        .base => .base,
        .ext => .ext,
    };
}

fn bridgeLayer(layer: loom_vectors.LoomBridgeLayer) fri.QueryLayer {
    return .{
        .rail = .ext,
        .leaf_p_ext = uintsToExt(layer.leaf_p_ext),
        .leaf_q_ext = uintsToExt(layer.leaf_q_ext),
        .path = .{ .leaf_idx = layer.leaf_idx, .siblings = &.{} },
    };
}

fn preparePcsTranscript(case: pcs_vectors.PcsIntegrationCase, ts: *fiat_shamir.Transcript) !ext.Ext {
    var zeta_bindings: [8]field.Element = undefined;
    try std.testing.expect(case.zeta_bindings.len <= zeta_bindings.len);

    try ts.newChallenge(fri.bridge.final_evaluation_challenge);
    try ts.newChallenge(fri.bridge.deep_alpha_challenge);
    fillElems(zeta_bindings[0..case.zeta_bindings.len], case.zeta_bindings);
    try ts.bindElements(fri.bridge.final_evaluation_challenge, zeta_bindings[0..case.zeta_bindings.len]);

    try expectDigest(try ts.computeChallenge(fri.bridge.final_evaluation_challenge), case.zeta_digest);
    const zeta = try ts.computeChallengeExt(fri.bridge.final_evaluation_challenge);
    try expectExt(zeta, uintsToExt(case.zeta));
    return zeta;
}

fn expectDeepAlphaBadDimensions(
    dq_layout: dq_layout_types.DQLayout,
    values_at_zeta: *const std.StringHashMap(ext.Ext),
) !void {
    var ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
    try ts.newChallenge(fri.bridge.deep_alpha_challenge);
    try std.testing.expectError(
        error.BadDimensions,
        fri.bridge.deriveDeepAlpha(dq_layout, values_at_zeta, &ts),
    );
}

fn expectPcsVerifyError(
    expected_error: anyerror,
    case: pcs_vectors.PcsIntegrationCase,
    params: fri.Params,
    layout: layout_types.Layout,
    dq_layout: dq_layout_types.DQLayout,
    roots: []const fri.Digest,
    proof: fri.types.Proof,
    values_at_zeta: *const std.StringHashMap(ext.Ext),
    zeta: ext.Ext,
) !void {
    var ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
    _ = try preparePcsTranscript(case, &ts);
    try std.testing.expectError(
        expected_error,
        fri.pcs.verify(params, layout, dq_layout, roots, proof, values_at_zeta, zeta, &ts),
    );
}

fn expectBridgeError(
    expected_error: anyerror,
    layout: layout_types.Layout,
    dq_layout: dq_layout_types.DQLayout,
    proof: fri.types.Proof,
    values_at_zeta: *const std.StringHashMap(ext.Ext),
    zeta: ext.Ext,
    alpha: ext.Ext,
) !void {
    try std.testing.expectError(
        expected_error,
        fri.bridge.checkFRIBridge(layout, dq_layout, proof, values_at_zeta, zeta, alpha),
    );
}

fn expectFriVerifyError(
    expected_error: anyerror,
    params: fri.Params,
    level_roots: []const fri.Digest,
    level_ds: []const u32,
    proof: fri.FriProof,
) !void {
    var ts = fiat_shamir.Transcript.initWithBackend("poseidon2");
    try std.testing.expectError(expected_error, fri.verify.friVerify(params, level_roots, level_ds, proof, &ts));
}

fn digestEql(left: fri.Digest, right: fri.Digest) bool {
    for (left, right) |left_limb, right_limb| {
        if (!left_limb.eql(right_limb)) return false;
    }
    return true;
}
