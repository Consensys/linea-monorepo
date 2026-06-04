const std = @import("std");

const ext = @import("../field/koalabear_ext.zig");
const dq_layout_mod = @import("../dq_layout.zig");
const layout_mod = @import("../layout.zig");
const merkle = @import("merkle.zig");
const types = @import("types.zig");

pub const PointSamplingError = error{
    BadDimensions,
    InvalidMerkleProof,
};

pub const BridgeError = error{
    MissingValueAtZeta,
    BridgeMismatch,
    BadShifts,
    Unsupported,
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

pub fn checkFRIBridge(
    layout: layout_mod.Layout,
    dq_layout: dq_layout_mod.DQLayout,
    proof: types.Proof,
    values_at_zeta: *const std.StringHashMap(ext.Ext),
    zeta: ext.Ext,
    alpha: ext.Ext,
) BridgeError!void {
    _ = layout;
    _ = dq_layout;
    _ = proof;
    _ = values_at_zeta;
    _ = zeta;
    _ = alpha;
    return BridgeError.Unsupported;
}
