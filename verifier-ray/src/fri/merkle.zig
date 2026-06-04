const leaf_hash = @import("leaf_hash.zig");
const types = @import("types.zig");

pub const Digest = types.Digest;
pub const MerklePath = types.MerklePath;
pub const MerkleProof = types.MerkleProof;

pub fn merkleVerify(root: Digest, path: MerklePath, leaf: Digest) bool {
    var cur = leaf;
    var idx = path.leaf_idx;

    for (path.siblings) |sibling| {
        cur = if ((idx & 1) == 0)
            leaf_hash.hashNode(cur, sibling)
        else
            leaf_hash.hashNode(sibling, cur);
        idx >>= 1;
    }

    return digestEql(cur, root);
}

pub fn proofVerify(root: Digest, proof: MerkleProof) bool {
    const leaf = leaf_hash.hashLeaf(proof.raw_leaf_base, proof.raw_leaf_ext);
    return merkleVerify(root, proof.path, leaf);
}

pub fn digestEql(left: Digest, right: Digest) bool {
    for (left, right) |left_limb, right_limb| {
        if (!left_limb.eql(right_limb)) return false;
    }
    return true;
}
