const base = @import("koalabear.zig");

pub const degree = 6;
pub const bytes = degree * base.bytes;

pub const Ext = struct {
    /// Limb order matches prover-ray and gnark-crypto:
    /// B0.A0, B0.A1, B1.A0, B1.A1, B2.A0, B2.A1
    /// for the tower F_{p^6} = F_{p^2}[v]/(v^3 - (u+1)) with F_{p^2} = F_p[u]/(u^2 - 3).
    limbs: [degree]base.Element,

    pub fn zero() Ext {
        return .{ .limbs = .{
            base.Element.zero(),
            base.Element.zero(),
            base.Element.zero(),
            base.Element.zero(),
            base.Element.zero(),
            base.Element.zero(),
        } };
    }

    pub fn one() Ext {
        return lift(base.Element.one());
    }

    pub fn lift(value: base.Element) Ext {
        return .{ .limbs = .{
            value,
            base.Element.zero(),
            base.Element.zero(),
            base.Element.zero(),
            base.Element.zero(),
            base.Element.zero(),
        } };
    }

    pub fn isZero(self: Ext) bool {
        for (self.limbs) |limb| {
            if (!limb.isZero()) return false;
        }
        return true;
    }

    pub fn isBase(self: Ext) bool {
        return self.limbs[1].isZero() and self.limbs[2].isZero() and self.limbs[3].isZero() and
            self.limbs[4].isZero() and self.limbs[5].isZero();
    }

    pub fn eql(self: Ext, rhs: Ext) bool {
        for (self.limbs, rhs.limbs) |lhs_limb, rhs_limb| {
            if (!lhs_limb.eql(rhs_limb)) return false;
        }
        return true;
    }

    pub fn add(self: Ext, rhs: Ext) Ext {
        var out = Ext.zero();
        for (&out.limbs, self.limbs, rhs.limbs) |*dst, lhs_limb, rhs_limb| {
            dst.* = lhs_limb.add(rhs_limb);
        }
        return out;
    }

    pub fn sub(self: Ext, rhs: Ext) Ext {
        var out = Ext.zero();
        for (&out.limbs, self.limbs, rhs.limbs) |*dst, lhs_limb, rhs_limb| {
            dst.* = lhs_limb.sub(rhs_limb);
        }
        return out;
    }

    pub fn neg(self: Ext) Ext {
        var out = Ext.zero();
        for (&out.limbs, self.limbs) |*dst, limb| {
            dst.* = limb.neg();
        }
        return out;
    }

    pub fn mulByBase(self: Ext, rhs: base.Element) Ext {
        var out = Ext.zero();
        for (&out.limbs, self.limbs) |*dst, limb| {
            dst.* = limb.mul(rhs);
        }
        return out;
    }

    pub fn divByBase(self: Ext, rhs: base.Element) Ext {
        return self.mulByBase(rhs.inverse());
    }

    pub fn mul(self: Ext, rhs: Ext) Ext {
        const a0 = E2{ .a0 = self.limbs[0], .a1 = self.limbs[1] };
        const a1 = E2{ .a0 = self.limbs[2], .a1 = self.limbs[3] };
        const a2 = E2{ .a0 = self.limbs[4], .a1 = self.limbs[5] };
        const b0 = E2{ .a0 = rhs.limbs[0], .a1 = rhs.limbs[1] };
        const b1 = E2{ .a0 = rhs.limbs[2], .a1 = rhs.limbs[3] };
        const b2 = E2{ .a0 = rhs.limbs[4], .a1 = rhs.limbs[5] };

        // F_{p^6} = F_{p^2}[v]/(v^3 - (u+1)), so v^3 = u+1
        // D0 = A0*B0 + (A1*B2 + A2*B1)*(u+1)
        // D1 = A0*B1 + A1*B0 + A2*B2*(u+1)
        // D2 = A0*B2 + A2*B0 + A1*B1
        const d0 = a0.mul(b0).add(a1.mul(b2).add(a2.mul(b1)).mulByNonResidue());
        const d1 = a0.mul(b1).add(a1.mul(b0)).add(a2.mul(b2).mulByNonResidue());
        const d2 = a0.mul(b2).add(a2.mul(b0)).add(a1.mul(b1));

        return .{ .limbs = .{ d0.a0, d0.a1, d1.a0, d1.a1, d2.a0, d2.a1 } };
    }

    pub fn square(self: Ext) Ext {
        return self.mul(self);
    }

    pub fn inverse(self: Ext) Ext {
        if (self.isZero()) unreachable;

        const b0 = E2{ .a0 = self.limbs[0], .a1 = self.limbs[1] };
        const b1 = E2{ .a0 = self.limbs[2], .a1 = self.limbs[3] };
        const b2 = E2{ .a0 = self.limbs[4], .a1 = self.limbs[5] };

        // Adjugate elements for the cubic extension inverse:
        //   A = b0^2 - (u+1)*b1*b2
        //   B = (u+1)*b2^2 - b0*b1
        //   C = b1^2 - b0*b2
        const cap_a = b0.mul(b0).sub(b1.mul(b2).mulByNonResidue());
        const cap_b = b2.mul(b2).mulByNonResidue().sub(b0.mul(b1));
        const cap_c = b1.mul(b1).sub(b0.mul(b2));

        // Norm: d = b0*A + (u+1)*(b2*B + b1*C)
        const d = b0.mul(cap_a).add(b2.mul(cap_b).add(b1.mul(cap_c)).mulByNonResidue());
        const d_inv = d.inverse();

        const e0 = cap_a.mul(d_inv);
        const e1 = cap_b.mul(d_inv);
        const e2 = cap_c.mul(d_inv);

        return .{ .limbs = .{ e0.a0, e0.a1, e1.a0, e1.a1, e2.a0, e2.a1 } };
    }

    pub fn div(self: Ext, rhs: Ext) Ext {
        return self.mul(rhs.inverse());
    }

    pub fn pow(self: Ext, exponent: u64) Ext {
        var result = Ext.one();
        var b = self;
        var exp = exponent;
        while (exp != 0) : (exp >>= 1) {
            if ((exp & 1) == 1) result = result.mul(b);
            b = b.square();
        }
        return result;
    }

    pub fn toBytes(self: Ext) [bytes]u8 {
        var out: [bytes]u8 = undefined;
        for (self.limbs, 0..) |limb, i| {
            const encoded = limb.toBytes();
            @memcpy(out[i * base.bytes .. (i + 1) * base.bytes], &encoded);
        }
        return out;
    }

    pub fn fromBytesCanonical(encoded: [bytes]u8) base.Error!Ext {
        return .{ .limbs = .{
            try base.Element.fromBytesCanonical(encoded[0..4].*),
            try base.Element.fromBytesCanonical(encoded[4..8].*),
            try base.Element.fromBytesCanonical(encoded[8..12].*),
            try base.Element.fromBytesCanonical(encoded[12..16].*),
            try base.Element.fromBytesCanonical(encoded[16..20].*),
            try base.Element.fromBytesCanonical(encoded[20..24].*),
        } };
    }
};

const E2 = struct {
    a0: base.Element,
    a1: base.Element,

    fn add(self: E2, rhs: E2) E2 {
        return .{ .a0 = self.a0.add(rhs.a0), .a1 = self.a1.add(rhs.a1) };
    }

    fn sub(self: E2, rhs: E2) E2 {
        return .{ .a0 = self.a0.sub(rhs.a0), .a1 = self.a1.sub(rhs.a1) };
    }

    fn mul(self: E2, rhs: E2) E2 {
        const c0 = self.a0.mul(rhs.a0).add(self.a1.mul(rhs.a1).mul(base.Element.init(3)));
        const c1 = self.a0.mul(rhs.a1).add(self.a1.mul(rhs.a0));
        return .{ .a0 = c0, .a1 = c1 };
    }

    // Multiply by the non-residue (u+1) of the cubic extension: (a0 + 3*a1) + (a0 + a1)*u
    fn mulByNonResidue(self: E2) E2 {
        const three = base.Element.init(3);
        return .{
            .a0 = self.a0.add(self.a1.mul(three)),
            .a1 = self.a0.add(self.a1),
        };
    }

    // Invert in F_p[u]/(u^2 - 3): norm = a0^2 - 3*a1^2
    fn inverse(self: E2) E2 {
        const three = base.Element.init(3);
        const norm = self.a0.mul(self.a0).sub(self.a1.mul(self.a1).mul(three));
        const norm_inv = norm.inverse();
        return .{
            .a0 = self.a0.mul(norm_inv),
            .a1 = self.a1.neg().mul(norm_inv),
        };
    }
};

pub fn zero() Ext {
    return Ext.zero();
}

pub fn one() Ext {
    return Ext.one();
}
