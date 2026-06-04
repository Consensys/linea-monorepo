const poseidon2 = @import("../crypto/poseidon2.zig");
const field = @import("../field/koalabear.zig");
const ext = @import("../field/koalabear_ext.zig");
const types = @import("types.zig");

pub const Digest = types.Digest;
pub const PairBase = types.PairBase;
pub const PairExt = types.PairExt;

const tag_leaf = field.Element.init(0x4c454146); // "LEAF"
const tag_node = field.Element.init(0x4e4f4445); // "NODE"

pub fn hashLeaf(base: []const PairBase, ext_: []const PairExt) Digest {
    var h = poseidon2.SpongeHasher.init();
    h.writeElement(tag_leaf);
    h.writeElement(field.Element.init(@intCast(base.len)));
    h.writeElement(field.Element.init(@intCast(ext_.len)));

    for (base) |pair| {
        h.writeElement(pair[0]);
        h.writeElement(pair[1]);
    }

    for (ext_) |pair| {
        writeExt(&h, pair[0]);
        writeExt(&h, pair[1]);
    }

    return h.sumDigest();
}

pub fn hashNode(left: Digest, right: Digest) Digest {
    return poseidon2.nodeCompress(tag_node, left, right);
}

fn writeExt(hasher: *poseidon2.SpongeHasher, value: ext.Ext) void {
    hasher.writeElements(&.{
        value.B0.a0,
        value.B0.a1,
        value.B1.a0,
        value.B1.a1,
        value.B2.a0,
        value.B2.a1,
    });
}
