const std = @import("std");

const field = @import("../field/koalabear.zig");
const ext = @import("../field/koalabear_ext.zig");
const field_value = @import("../field/value.zig");
const poseidon2 = @import("poseidon2.zig");

pub const Error = poseidon2.Error || error{
    TooManyChallenges,
    TooManyBindings,
    UnknownChallenge,
    DuplicateChallenge,
    ChallengeNameTooLong,
    ChallengeOutOfOrder,
    ChallengeAlreadyComputed,
    InvalidProofOfWork,
};

pub const ProofOfWork = struct {
    nb_bits: u32,
    salt: field.Element,
};

const max_challenges = 96;
const max_challenge_name_len = 64;
const max_bindings_per_challenge = 1024;

const tag_fsid = field.Element.init(0x46534944); // "FSID"
const tag_fspw = field.Element.init(0x46535057); // "FSPW"
const string_chunk_size = 3;

const ChallengeSlot = struct {
    name: [max_challenge_name_len]u8,
    name_len: usize,
    bindings: [max_bindings_per_challenge]field.Element,
    binding_len: usize,
    pow: ?ProofOfWork,
    computed: bool,
    digest: poseidon2.Digest,

    fn empty() ChallengeSlot {
        return .{
            .name = undefined,
            .name_len = 0,
            .bindings = undefined,
            .binding_len = 0,
            .pow = null,
            .computed = false,
            .digest = poseidon2.zeroDigest(),
        };
    }
};

pub const Transcript = struct {
    hasher: poseidon2.MDHasher,
    backend_id: []const u8,
    challenges: [max_challenges]ChallengeSlot,
    challenge_len: usize,

    pub fn init() Transcript {
        return initWithBackend("poseidon2");
    }

    pub fn initWithBackend(backend_id: []const u8) Transcript {
        return .{
            .hasher = poseidon2.MDHasher.init(),
            .backend_id = backend_id,
            .challenges = emptyChallenges(),
            .challenge_len = 0,
        };
    }

    pub fn updateElement(self: *Transcript, value: field.Element) void {
        self.hasher.writeElement(value);
    }

    pub fn updateElements(self: *Transcript, values: field_value.ElementSlice) void {
        self.hasher.writeElements(values);
    }

    pub fn updateExt(self: *Transcript, values: field_value.ExtSlice) void {
        for (values) |ext_value| {
            self.hasher.writeElements(&.{ ext_value.B0.a0, ext_value.B0.a1, ext_value.B1.a0, ext_value.B1.a1, ext_value.B2.a0, ext_value.B2.a1 });
        }
    }

    pub fn absorbVector(self: *Transcript, vector: field_value.Vector) void {
        switch (vector) {
            .base => |values| self.updateElements(values),
            .ext => |values| self.updateExt(values),
        }
    }

    pub fn absorbScalar(self: *Transcript, scalar: field_value.Scalar) void {
        switch (scalar) {
            .base => |scalar_value| self.updateElement(scalar_value),
            .ext => |scalar_value| self.updateExt(&.{scalar_value}),
        }
    }

    pub fn randomDigest(self: *Transcript) poseidon2.Digest {
        const challenge = self.hasher.sumDigest();
        self.updateElement(field.Element.zero());
        return challenge;
    }

    pub fn randomExt(self: *Transcript) ext.Ext {
        const challenge = self.randomDigest();
        return .{
            .B0 = .{ .a0 = challenge[0], .a1 = challenge[1] },
            .B1 = .{ .a0 = challenge[2], .a1 = challenge[3] },
            .B2 = .{ .a0 = challenge[4], .a1 = challenge[5] },
        };
    }

    pub fn state(self: *const Transcript) poseidon2.Digest {
        return self.hasher.getState();
    }

    pub fn setState(self: *Transcript, digest: poseidon2.Digest) void {
        self.hasher.setState(digest);
    }

    /// Registers a named Fiat-Shamir challenge.
    ///
    /// The named-challenge API is intentionally separate from the legacy
    /// absorb/random API above. Existing prover-ray compatibility code should
    /// keep using update*/random*. FRI code should use only newChallenge,
    /// bindElements, setProofOfWork, and computeChallenge* for a transcript.
    pub fn newChallenge(self: *Transcript, name: []const u8) Error!void {
        if (self.challenge_len == self.challenges.len) return Error.TooManyChallenges;
        if (name.len > max_challenge_name_len) return Error.ChallengeNameTooLong;
        if (self.findChallenge(name) != null) return Error.DuplicateChallenge;

        self.challenges[self.challenge_len] = ChallengeSlot.empty();
        @memcpy(self.challenges[self.challenge_len].name[0..name.len], name);
        self.challenges[self.challenge_len].name_len = name.len;
        self.challenge_len += 1;
    }

    pub fn bindElements(self: *Transcript, name: []const u8, vals: []const field.Element) Error!void {
        const index = self.findChallenge(name) orelse return Error.UnknownChallenge;
        var slot = &self.challenges[index];
        if (slot.computed) return Error.ChallengeAlreadyComputed;
        if (slot.binding_len + vals.len > slot.bindings.len) return Error.TooManyBindings;

        @memcpy(slot.bindings[slot.binding_len .. slot.binding_len + vals.len], vals);
        slot.binding_len += vals.len;
    }

    pub fn bindDigest(self: *Transcript, name: []const u8, digest: poseidon2.Digest) Error!void {
        try self.bindElements(name, digest[0..]);
    }

    pub fn setProofOfWork(self: *Transcript, name: []const u8, pow: ProofOfWork) Error!void {
        const index = self.findChallenge(name) orelse return Error.UnknownChallenge;
        var slot = &self.challenges[index];
        if (slot.computed) return Error.ChallengeAlreadyComputed;
        if (pow.nb_bits > 31) return Error.InvalidProofOfWork;
        slot.pow = pow;
    }

    pub fn computeChallenge(self: *Transcript, name: []const u8) Error!poseidon2.Digest {
        const index = self.findChallenge(name) orelse return Error.UnknownChallenge;
        if (index != 0 and !self.challenges[index - 1].computed) {
            return Error.ChallengeOutOfOrder;
        }

        var slot = &self.challenges[index];
        if (slot.computed) return slot.digest;

        var h = poseidon2.SpongeHasher.init();
        writeStringElements(&h, tag_fsid, slot.name[0..slot.name_len]);
        if (index != 0) {
            h.writeElements(self.challenges[index - 1].digest[0..]);
        }
        h.writeElements(slot.bindings[0..slot.binding_len]);

        if (slot.pow) |pow| {
            h.writeElement(tag_fspw);
            h.writeElement(field.Element.init(pow.nb_bits));
            h.writeElement(pow.salt);
        }

        const digest = h.sumDigest();
        if (slot.pow) |pow| {
            if (!hasLowZeroBits(digest, pow.nb_bits)) return Error.InvalidProofOfWork;
        }

        slot.digest = digest;
        slot.computed = true;
        return digest;
    }

    pub fn computeChallengeExt(self: *Transcript, name: []const u8) Error!ext.Ext {
        const digest = try self.computeChallenge(name);
        return .{
            .B0 = .{ .a0 = digest[0], .a1 = digest[1] },
            .B1 = .{ .a0 = digest[2], .a1 = digest[3] },
            .B2 = .{ .a0 = digest[4], .a1 = digest[5] },
        };
    }

    fn findChallenge(self: *const Transcript, name: []const u8) ?usize {
        for (self.challenges[0..self.challenge_len], 0..) |slot, index| {
            if (std.mem.eql(u8, slot.name[0..slot.name_len], name)) return index;
        }
        return null;
    }
};

fn emptyChallenges() [max_challenges]ChallengeSlot {
    var challenges: [max_challenges]ChallengeSlot = undefined;
    for (&challenges) |*challenge| {
        challenge.* = ChallengeSlot.empty();
    }
    return challenges;
}

fn writeStringElements(hasher: *poseidon2.SpongeHasher, domain_tag: field.Element, bytes: []const u8) void {
    hasher.writeElement(domain_tag);
    hasher.writeElement(field.Element.init(@intCast(bytes.len)));
    var i: usize = 0;
    while (i < bytes.len) : (i += string_chunk_size) {
        var limb: u64 = 0;
        var j: usize = 0;
        while (j < string_chunk_size and i + j < bytes.len) : (j += 1) {
            const shift: u6 = @intCast(8 * j);
            limb |= @as(u64, bytes[i + j]) << shift;
        }
        hasher.writeElement(field.Element.init(limb));
    }
}

fn hasLowZeroBits(digest: poseidon2.Digest, nb_bits: u32) bool {
    var remaining = nb_bits;
    for (digest) |limb| {
        if (remaining == 0) return true;
        const take = @min(remaining, 31);
        const shift: u5 = @intCast(take);
        const mask = (@as(u32, 1) << shift) - 1;
        if ((limb.value & mask) != 0) return false;
        remaining -= take;
    }
    return remaining == 0;
}
