const base = @import("koalabear.zig");

pub const degree = 4;

pub const Ext = struct {
    limbs: [degree]base.Element,

    pub fn zero() Ext {
        return .{ .limbs = .{
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
        } };
    }

    pub fn isZero(self: Ext) bool {
        for (self.limbs) |limb| {
            if (!limb.isZero()) return false;
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

    pub fn mulByBase(self: Ext, rhs: base.Element) Ext {
        var out = Ext.zero();
        for (&out.limbs, self.limbs) |*dst, limb| {
            dst.* = limb.mul(rhs);
        }
        return out;
    }
};
