const fiat_shamir = @import("../crypto/fiat_shamir.zig");
const poseidon2 = @import("../crypto/poseidon2.zig");
const field = @import("../field/koalabear.zig");
const ext = @import("../field/koalabear_ext.zig");

pub const Digest = poseidon2.Digest;
pub const PairBase = [2]field.Element;
pub const PairExt = [2]ext.Ext;
pub const ProofOfWork = fiat_shamir.ProofOfWork;

pub const Rail = enum {
    base,
    ext,
};

pub const MerklePath = struct {
    leaf_idx: u32,
    siblings: []const Digest = &.{},
};

pub const MerkleProof = struct {
    raw_leaf_base: []const PairBase = &.{},
    raw_leaf_ext: []const PairExt = &.{},
    path: MerklePath,
};

pub const QueryLayer = struct {
    rail: Rail,
    leaf_p_base: field.Element = field.Element.zero(),
    leaf_q_base: field.Element = field.Element.zero(),
    leaf_p_ext: ext.Ext = ext.Ext.zero(),
    leaf_q_ext: ext.Ext = ext.Ext.zero(),
    path: MerklePath,

    pub fn basePair(self: QueryLayer) PairBase {
        return .{ self.leaf_p_base, self.leaf_q_base };
    }

    pub fn extPair(self: QueryLayer) PairExt {
        return .{ self.leaf_p_ext, self.leaf_q_ext };
    }
};

pub const Query = struct {
    layers: []const QueryLayer = &.{},
};

pub const FriProof = struct {
    fri_roots: []const Digest = &.{},
    final_rail: Rail = .ext,
    final_poly_base: []const field.Element = &.{},
    final_poly_ext: []const ext.Ext = &.{},
    queries: []const Query = &.{},
    level_queries: []const []const QueryLayer = &.{},
    pow: []const ?ProofOfWork = &.{},
};

pub const Params = struct {
    n: u32,
    d: u32,
    num_queries: u32,
    num_rounds: u32,
    domain_gens: []const field.Element = &.{},
    domain_gens_inv: []const field.Element = &.{},
    grinding: u32,
};

pub const Proof = struct {
    deep_quotient_commitment: []const Digest = &.{},
    level_ds: []const u32 = &.{},
    fri: FriProof = .{},
    point_samplings: []const []const MerkleProof = &.{},
};
