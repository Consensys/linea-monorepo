pub const protocol = @import("protocol/root.zig");
pub const verifier = @import("verifier.zig");
pub const layout = @import("layout.zig");
pub const dq_layout = @import("dq_layout.zig");

pub const field = struct {
    pub const koalabear = @import("field/koalabear.zig");
    pub const koalabear_ext = @import("field/koalabear_ext.zig");
    pub const value = @import("field/value.zig");
};

pub const crypto = struct {
    pub const commitment = @import("crypto/commitment.zig");
    pub const fiat_shamir = @import("crypto/fiat_shamir.zig");
    pub const poseidon2 = @import("crypto/poseidon2.zig");
};

pub const polynomial = struct {
    pub const lagrange = @import("polynomial/lagrange.zig");
    pub const canonical = @import("polynomial/canonical.zig");
};

pub const query = struct {
    pub const vanishing = @import("query/vanishing.zig");
};

pub const fri = struct {
    pub const types = @import("fri/types.zig");
    pub const leaf_hash = @import("fri/leaf_hash.zig");
    pub const merkle = @import("fri/merkle.zig");
    pub const verify = @import("fri/verify.zig");
    pub const bridge = @import("fri/bridge.zig");
    pub const pcs = @import("fri/pcs.zig");

    pub const Digest = types.Digest;
    pub const PairBase = types.PairBase;
    pub const PairExt = types.PairExt;
    pub const ProofOfWork = types.ProofOfWork;
    pub const Rail = types.Rail;
    pub const MerklePath = types.MerklePath;
    pub const MerkleProof = types.MerkleProof;
    pub const QueryLayer = types.QueryLayer;
    pub const Query = types.Query;
    pub const FriProof = types.FriProof;
    pub const Params = types.Params;
};

pub const vortex = struct {
    pub const reed_solomon = @import("vortex/reed_solomon.zig");
    pub const ringsis = @import("vortex/ringsis.zig");
    pub const smt = @import("vortex/smt.zig");
    pub const verifier = @import("vortex/verifier.zig");
};

pub const generated = struct {
    pub const stub = @import("generated/stub.zig");
};

pub const Proof = proof.Proof;
pub const VerifyError = verifier.VerifyError;

pub fn verify(p: Proof) VerifyError!void {
    return verifier.verify(p);
}
