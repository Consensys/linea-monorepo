const std = @import("std");

const fiat_shamir = @import("../crypto/fiat_shamir.zig");
const field = @import("../field/koalabear.zig");
const ext = @import("../field/koalabear_ext.zig");
const dq_layout_mod = @import("../dq_layout.zig");
const layout_mod = @import("../layout.zig");
const merkle = @import("merkle.zig");
const types = @import("types.zig");

const code_rate = 4; // loom constants.RATE
pub const final_evaluation_challenge = "__zeta";
pub const deep_alpha_challenge = "alpha_DEEP";

pub const PointSamplingError = error{
    BadDimensions,
    InvalidMerkleProof,
};

pub const BridgeError = error{
    BadDimensions,
    MissingColumnSlot,
    MissingValueAtZeta,
    BridgeMismatch,
    BadShifts,
};

pub const DeepChallengeError = fiat_shamir.Error || error{
    BadDimensions,
    MissingValueAtZeta,
};

pub fn checkPointSamplings(
    roots: []const types.Digest,
    samplings: []const []const types.MerkleProof,
) PointSamplingError!void {
    for (samplings) |query_samplings| {
        if (query_samplings.len != roots.len) return PointSamplingError.BadDimensions;
        for (roots, query_samplings) |root, sampling| {
            if (!merkle.proofVerify(root, sampling)) return PointSamplingError.InvalidMerkleProof;
        }
    }
}

/// Binds every DEEP-quotient evaluation claim to `alpha_DEEP` in loom's
/// canonical order: per level, all shifted column keys, then AIR chunks.
pub fn bindDeepEvaluationClaims(
    dq_layout: dq_layout_mod.DQLayout,
    values_at_zeta: *const std.StringHashMap(ext.Ext),
    ts: *fiat_shamir.Transcript,
) DeepChallengeError!void {
    try checkDeepChallengeDimensions(dq_layout);

    for (dq_layout.sizes, 0..) |_, level_index| {
        for (dq_layout.column_keys[level_index]) |keys_at_point| {
            for (keys_at_point) |key| {
                try bindValueAtZeta(values_at_zeta, ts, key);
            }
        }
        for (dq_layout.air_chunks[level_index]) |chunk_name| {
            try bindValueAtZeta(values_at_zeta, ts, chunk_name);
        }
    }
}

pub fn deriveDeepAlpha(
    dq_layout: dq_layout_mod.DQLayout,
    values_at_zeta: *const std.StringHashMap(ext.Ext),
    ts: *fiat_shamir.Transcript,
) DeepChallengeError!ext.Ext {
    try bindDeepEvaluationClaims(dq_layout, values_at_zeta, ts);
    return try ts.computeChallengeExt(deep_alpha_challenge);
}

/// Verifies the DEEP quotient bridge after `friVerify` has bound the FRI query
/// positions and `checkPointSamplings` has authenticated `proof.point_samplings`.
/// `dq_layout.eval_points` must contain the runtime points zeta * omega_N^shift.
pub fn checkFRIBridge(
    layout: layout_mod.Layout,
    dq_layout: dq_layout_mod.DQLayout,
    proof: types.Proof,
    values_at_zeta: *const std.StringHashMap(ext.Ext),
    zeta: ext.Ext,
    alpha: ext.Ext,
) BridgeError!void {
    try checkBridgeDimensions(layout, dq_layout, proof);

    for (proof.fri.queries, 0..) |query, query_index| {
        if (query.layers.len == 0) return BridgeError.BadDimensions;
        const position = query.layers[0].path.leaf_idx;

        for (dq_layout.sizes, 0..) |degree_bound, level_index| {
            const domain_size = checkedDomainSize(degree_bound) catch return BridgeError.BadShifts;
            const half_domain = domain_size / 2;
            if (half_domain == 0) return BridgeError.BadShifts;
            const query_base = position % half_domain;
            const generator = field.rootOfUnityBy(domain_size) catch return BridgeError.BadShifts;
            const x_base = generator.pow(query_base);
            const x = ext.Ext.lift(x_base);
            const neg_x = ext.Ext.lift(x_base.neg());

            var dq_p = ext.Ext.zero();
            var dq_q = ext.Ext.zero();
            var alpha_acc = ext.Ext.one();

            for (dq_layout.eval_points[level_index], 0..) |eval_point, eval_index| {
                const names = dq_layout.column_names[level_index][eval_index];
                const keys = dq_layout.column_keys[level_index][eval_index];
                var v_at_z = ext.Ext.zero();
                var c_at_x = ext.Ext.zero();
                var c_at_neg_x = ext.Ext.zero();

                for (names, keys) |name, key| {
                    const eval = values_at_zeta.get(key) orelse return BridgeError.MissingValueAtZeta;
                    const slot = layout.col_slot.get(name) orelse return BridgeError.MissingColumnSlot;
                    const leaf = try samplePair(layout, proof.point_samplings, query_index, position, slot);

                    v_at_z = v_at_z.add(eval.mul(alpha_acc));
                    c_at_x = c_at_x.add(leaf.p.mul(alpha_acc));
                    c_at_neg_x = c_at_neg_x.add(leaf.q.mul(alpha_acc));
                    alpha_acc = alpha_acc.mul(alpha);
                }

                try accumulateBridgeTerm(eval_point, x, neg_x, v_at_z, c_at_x, c_at_neg_x, &dq_p, &dq_q);
            }

            if (dq_layout.air_chunks[level_index].len != 0) {
                var v_air = ext.Ext.zero();
                var c_at_x = ext.Ext.zero();
                var c_at_neg_x = ext.Ext.zero();

                for (dq_layout.air_chunks[level_index]) |chunk_name| {
                    const eval = values_at_zeta.get(chunk_name) orelse return BridgeError.MissingValueAtZeta;
                    const slot = layout.air_chunk_slot.get(chunk_name) orelse return BridgeError.MissingColumnSlot;
                    const leaf = try samplePair(layout, proof.point_samplings, query_index, position, slot);

                    v_air = v_air.add(eval.mul(alpha_acc));
                    c_at_x = c_at_x.add(leaf.p.mul(alpha_acc));
                    c_at_neg_x = c_at_neg_x.add(leaf.q.mul(alpha_acc));
                    alpha_acc = alpha_acc.mul(alpha);
                }

                try accumulateBridgeTerm(zeta, x, neg_x, v_air, c_at_x, c_at_neg_x, &dq_p, &dq_q);
            }

            const actual = try friLevelPair(proof, query_index, level_index);
            if (!dq_p.eql(actual.p) or !dq_q.eql(actual.q)) return BridgeError.BridgeMismatch;
        }
    }
}

fn checkDeepChallengeDimensions(dq_layout: dq_layout_mod.DQLayout) DeepChallengeError!void {
    const levels = dq_layout.sizes.len;
    if (levels == 0) return DeepChallengeError.BadDimensions;
    if (dq_layout.column_keys.len != levels) return DeepChallengeError.BadDimensions;
    if (dq_layout.air_chunks.len != levels) return DeepChallengeError.BadDimensions;
}

fn bindValueAtZeta(
    values_at_zeta: *const std.StringHashMap(ext.Ext),
    ts: *fiat_shamir.Transcript,
    key: []const u8,
) DeepChallengeError!void {
    const value = values_at_zeta.get(key) orelse return DeepChallengeError.MissingValueAtZeta;
    const limbs = [_]field.Element{
        value.B0.a0,
        value.B0.a1,
        value.B1.a0,
        value.B1.a1,
        value.B2.a0,
        value.B2.a1,
    };
    try ts.bindElements(deep_alpha_challenge, limbs[0..]);
}

fn checkBridgeDimensions(
    layout: layout_mod.Layout,
    dq_layout: dq_layout_mod.DQLayout,
    proof: types.Proof,
) BridgeError!void {
    const levels = dq_layout.sizes.len;
    if (levels == 0) return BridgeError.BadDimensions;
    if (proof.level_ds.len != levels) return BridgeError.BadDimensions;
    if (proof.deep_quotient_commitment.len != levels) return BridgeError.BadDimensions;
    if (proof.fri.level_queries.len + 1 != levels) return BridgeError.BadDimensions;
    if (proof.point_samplings.len != proof.fri.queries.len) return BridgeError.BadDimensions;
    if (layout.tree_size.len != @as(usize, @intCast(layout.num_trees))) return BridgeError.BadDimensions;
    if (dq_layout.eval_points.len != levels) return BridgeError.BadDimensions;
    if (dq_layout.column_names.len != levels) return BridgeError.BadDimensions;
    if (dq_layout.column_keys.len != levels) return BridgeError.BadDimensions;
    if (dq_layout.air_chunks.len != levels) return BridgeError.BadDimensions;

    for (dq_layout.sizes, proof.level_ds, 0..) |size, level_d, level_index| {
        if (size != level_d) return BridgeError.BadDimensions;
        if (dq_layout.column_names[level_index].len != dq_layout.eval_points[level_index].len) {
            return BridgeError.BadDimensions;
        }
        if (dq_layout.column_keys[level_index].len != dq_layout.eval_points[level_index].len) {
            return BridgeError.BadDimensions;
        }
        for (dq_layout.column_names[level_index], dq_layout.column_keys[level_index]) |names, keys| {
            if (names.len != keys.len) return BridgeError.BadDimensions;
        }
    }
}

fn checkedDomainSize(degree_bound: u32) BridgeError!u32 {
    if (degree_bound == 0) return BridgeError.BadShifts;
    if (degree_bound > std.math.maxInt(u32) / code_rate) return BridgeError.BadShifts;
    const domain_size = degree_bound * code_rate;
    if (!field.isPowerOfTwo(@intCast(domain_size))) return BridgeError.BadShifts;
    return domain_size;
}

const ExtPair = struct {
    p: ext.Ext,
    q: ext.Ext,
};

fn samplePair(
    layout: layout_mod.Layout,
    point_samplings: []const []const types.MerkleProof,
    query_index: usize,
    position: u32,
    slot: layout_mod.Slot,
) BridgeError!ExtPair {
    if (query_index >= point_samplings.len) return BridgeError.BadDimensions;
    const query_samplings = point_samplings[query_index];
    if (layout.num_trees != 0 and query_samplings.len != @as(usize, @intCast(layout.num_trees))) {
        return BridgeError.BadDimensions;
    }
    const tree_index: usize = @intCast(slot.tree_idx);
    if (tree_index >= query_samplings.len) return BridgeError.BadDimensions;
    if (tree_index >= layout.tree_size.len) return BridgeError.BadDimensions;
    const sampling = query_samplings[tree_index];
    const expected_leaf_index = try samplingLeafIndex(position, layout.tree_size[tree_index]);
    if (sampling.path.leaf_idx != expected_leaf_index) return BridgeError.BadDimensions;

    const raw_index: usize = @intCast(slot.poly_idx);

    return switch (slot.rail) {
        .base => blk: {
            if (raw_index >= sampling.raw_leaf_base.len) return BridgeError.BadDimensions;
            const pair = sampling.raw_leaf_base[raw_index];
            break :blk .{ .p = ext.Ext.lift(pair[0]), .q = ext.Ext.lift(pair[1]) };
        },
        .ext => blk: {
            if (raw_index >= sampling.raw_leaf_ext.len) return BridgeError.BadDimensions;
            const pair = sampling.raw_leaf_ext[raw_index];
            break :blk .{ .p = pair[0], .q = pair[1] };
        },
    };
}

fn samplingLeafIndex(position: u32, tree_size: u32) BridgeError!u32 {
    const domain_size = try checkedDomainSize(tree_size);
    return position % (domain_size / 2);
}

fn accumulateBridgeTerm(
    eval_point: ext.Ext,
    x: ext.Ext,
    neg_x: ext.Ext,
    v_at_z: ext.Ext,
    c_at_x: ext.Ext,
    c_at_neg_x: ext.Ext,
    dq_p: *ext.Ext,
    dq_q: *ext.Ext,
) BridgeError!void {
    const denom_p = eval_point.sub(x);
    const denom_q = eval_point.sub(neg_x);
    if (denom_p.isZero() or denom_q.isZero()) return BridgeError.BadShifts;
    dq_p.* = dq_p.*.add(v_at_z.sub(c_at_x).mul(denom_p.inverse()));
    dq_q.* = dq_q.*.add(v_at_z.sub(c_at_neg_x).mul(denom_q.inverse()));
}

fn friLevelPair(proof: types.Proof, query_index: usize, level_index: usize) BridgeError!ExtPair {
    if (level_index == 0) {
        if (query_index >= proof.fri.queries.len) return BridgeError.BadDimensions;
        const query = proof.fri.queries[query_index];
        if (query.layers.len == 0) return BridgeError.BadDimensions;
        const layer = query.layers[0];
        if (layer.rail != .ext) return BridgeError.BadDimensions;
        return .{ .p = layer.leaf_p_ext, .q = layer.leaf_q_ext };
    }

    const extra_level_index = level_index - 1;
    if (extra_level_index >= proof.fri.level_queries.len) return BridgeError.BadDimensions;
    const level_queries = proof.fri.level_queries[extra_level_index];
    if (query_index >= level_queries.len) return BridgeError.BadDimensions;
    const layer = level_queries[query_index];
    if (layer.rail != .ext) return BridgeError.BadDimensions;
    return .{ .p = layer.leaf_p_ext, .q = layer.leaf_q_ext };
}
