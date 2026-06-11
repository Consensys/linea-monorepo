const std = @import("std");

const fiat_shamir = @import("../crypto/fiat_shamir.zig");
const ext = @import("../field/koalabear_ext.zig");
const dq_layout_mod = @import("../dq_layout.zig");
const layout_mod = @import("../layout.zig");
const bridge = @import("bridge.zig");
const fri_verify = @import("verify.zig");
const types = @import("types.zig");

pub const Error = bridge.DeepChallengeError ||
    fri_verify.FriError ||
    bridge.PointSamplingError ||
    bridge.BridgeError;

/// Runs the PCS verification checks in loom verifier order.
///
/// The caller must have already registered and computed `__zeta`, and
/// registered `alpha_DEEP`, on `ts`. This function derives `alpha_DEEP`, then
/// lets FRI extend the same named transcript with its fold and query
/// challenges.
pub fn verify(
    params: types.Params,
    layout: layout_mod.Layout,
    dq_layout: dq_layout_mod.DQLayout,
    roots: []const types.Digest,
    proof: types.Proof,
    values_at_zeta: *const std.StringHashMap(ext.Ext),
    zeta: ext.Ext,
    ts: *fiat_shamir.Transcript,
) Error!void {
    const alpha = try bridge.deriveDeepAlpha(dq_layout, values_at_zeta, ts);
    try fri_verify.friVerify(params, proof.deep_quotient_commitment, proof.level_ds, proof.fri, ts);
    try bridge.checkPointSamplings(roots, proof.point_samplings);
    try bridge.checkFRIBridge(layout, dq_layout, proof, values_at_zeta, zeta, alpha);
}
