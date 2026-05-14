const field = @import("../field/koalabear.zig");
const ext = @import("../field/koalabear_ext.zig");
const poseidon2 = @import("poseidon2.zig");

pub const Transcript = struct {
    state: poseidon2.Digest,

    pub fn init() Transcript {
        return .{ .state = [_]field.Element{field.Element.zero()} ** 8 };
    }

    pub fn updateElement(self: *Transcript, value: field.Element) void {
        self.state[0] = self.state[0].add(value);
    }

    pub fn updateElements(self: *Transcript, values: []const field.Element) void {
        const digest = poseidon2.compress(values);
        for (&self.state, digest) |*state_limb, digest_limb| {
            state_limb.* = state_limb.add(digest_limb);
        }
    }

    pub fn challengeExt(self: *Transcript) ext.Ext {
        const challenge = ext.Ext.lift(self.state[0]);
        self.updateElement(field.Element.zero());
        return challenge;
    }
};
