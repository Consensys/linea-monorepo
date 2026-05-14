const std = @import("std");

pub const Modulus: u64 = 2130706433;
pub const MaxOrderRoot: usize = 24;
pub const RootOfUnityValue: u64 = 1791270792;

pub const Field = struct {
    value: u64,

    pub fn init(v: u64) Field {
        return .{ .value = v % Modulus };
    }

    pub fn add(self: Field, other: Field) Field {
        const sum = self.value + other.value;
        return .{ .value = if (sum >= Modulus) sum - Modulus else sum };
    }

    pub fn sub(self: Field, other: Field) Field {
        if (self.value >= other.value) {
            return .{ .value = self.value - other.value };
        }
        return .{ .value = self.value + Modulus - other.value };
    }

    pub fn mul(self: Field, other: Field) Field {
        return .{ .value = (self.value * other.value) % Modulus };
    }

    pub fn neg(self: Field) Field {
        return if (self.value == 0) self else .{ .value = Modulus - self.value };
    }

    pub fn square(self: Field) Field {
        return self.mul(self);
    }

    pub fn pow(self: Field, exponent: u64) Field {
        var result = f(1);
        var base = self;
        var e = exponent;
        while (e != 0) : (e >>= 1) {
            if ((e & 1) == 1) result = result.mul(base);
            base = base.square();
        }
        return result;
    }

    pub fn inv(self: Field) Field {
        return self.pow(Modulus - 2);
    }

    pub fn div(self: Field, other: Field) Field {
        return self.mul(other.inv());
    }

    pub fn isZero(self: Field) bool {
        return self.value == 0;
    }
};

pub fn f(v: u64) Field {
    return Field.init(v);
}

pub const Ext = struct {
    c0: Field,
    c1: Field,
    c2: Field,
    c3: Field,

    pub fn init(c0: Field, c1: Field, c2: Field, c3: Field) Ext {
        return .{ .c0 = c0, .c1 = c1, .c2 = c2, .c3 = c3 };
    }

    pub fn zero() Ext {
        return init(f(0), f(0), f(0), f(0));
    }

    pub fn one() Ext {
        return init(f(1), f(0), f(0), f(0));
    }

    pub fn fromBase(x: Field) Ext {
        return init(x, f(0), f(0), f(0));
    }

    pub fn add(self: Ext, other: Ext) Ext {
        return init(
            self.c0.add(other.c0),
            self.c1.add(other.c1),
            self.c2.add(other.c2),
            self.c3.add(other.c3),
        );
    }

    pub fn sub(self: Ext, other: Ext) Ext {
        return init(
            self.c0.sub(other.c0),
            self.c1.sub(other.c1),
            self.c2.sub(other.c2),
            self.c3.sub(other.c3),
        );
    }

    pub fn neg(self: Ext) Ext {
        return init(self.c0.neg(), self.c1.neg(), self.c2.neg(), self.c3.neg());
    }

    pub fn mulByField(self: Ext, scalar: Field) Ext {
        return init(
            self.c0.mul(scalar),
            self.c1.mul(scalar),
            self.c2.mul(scalar),
            self.c3.mul(scalar),
        );
    }

    pub fn mul(self: Ext, other: Ext) Ext {
        const three = f(3);

        // Tower representation: c0 + c1*u + c2*v + c3*u*v,
        // with u^2 = 3 and v^2 = u.
        const c0 = self.c0.mul(other.c0)
            .add(three.mul(self.c1.mul(other.c1)))
            .add(three.mul(self.c2.mul(other.c3)))
            .add(three.mul(self.c3.mul(other.c2)));

        const c1 = self.c0.mul(other.c1)
            .add(self.c1.mul(other.c0))
            .add(self.c2.mul(other.c2))
            .add(three.mul(self.c3.mul(other.c3)));

        const c2 = self.c0.mul(other.c2)
            .add(three.mul(self.c1.mul(other.c3)))
            .add(self.c2.mul(other.c0))
            .add(three.mul(self.c3.mul(other.c1)));

        const c3 = self.c0.mul(other.c3)
            .add(self.c1.mul(other.c2))
            .add(self.c2.mul(other.c1))
            .add(self.c3.mul(other.c0));

        return init(c0, c1, c2, c3);
    }

    pub fn square(self: Ext) Ext {
        return self.mul(self);
    }

    pub fn pow(self: Ext, exponent: u128) Ext {
        var result = Ext.one();
        var base = self;
        var e = exponent;
        while (e != 0) : (e >>= 1) {
            if ((e & 1) == 1) result = result.mul(base);
            base = base.square();
        }
        return result;
    }

    pub fn inv(self: Ext) Ext {
        const p = @as(u128, Modulus);
        return self.pow((p * p * p * p) - 2);
    }

    pub fn div(self: Ext, other: Ext) Ext {
        return self.mul(other.inv());
    }

    pub fn isZero(self: Ext) bool {
        return self.c0.isZero() and self.c1.isZero() and self.c2.isZero() and self.c3.isZero();
    }
};

pub const Gen = struct {
    ext: Ext,
    is_base: bool,

    pub fn base(x: Field) Gen {
        return .{ .ext = Ext.fromBase(x), .is_base = true };
    }

    pub fn fromExt(x: Ext) Gen {
        return .{ .ext = x, .is_base = false };
    }

    pub fn zero() Gen {
        return base(f(0));
    }

    pub fn one() Gen {
        return base(f(1));
    }

    pub fn isBase(self: Gen) bool {
        return self.is_base;
    }

    pub fn asBase(self: Gen) Field {
        if (!self.is_base) @panic("Gen.asBase called on an extension element");
        return self.ext.c0;
    }

    pub fn asExt(self: Gen) Ext {
        return self.ext;
    }

    pub fn add(self: Gen, other: Gen) Gen {
        return .{
            .ext = self.ext.add(other.ext),
            .is_base = self.is_base and other.is_base,
        };
    }

    pub fn sub(self: Gen, other: Gen) Gen {
        return .{
            .ext = self.ext.sub(other.ext),
            .is_base = self.is_base and other.is_base,
        };
    }

    pub fn mul(self: Gen, other: Gen) Gen {
        if (self.is_base and other.is_base) {
            return base(self.ext.c0.mul(other.ext.c0));
        }
        if (self.is_base) {
            return fromExt(other.ext.mulByField(self.ext.c0));
        }
        if (other.is_base) {
            return fromExt(self.ext.mulByField(other.ext.c0));
        }
        return fromExt(self.ext.mul(other.ext));
    }

    pub fn neg(self: Gen) Gen {
        return .{ .ext = self.ext.neg(), .is_base = self.is_base };
    }

    pub fn square(self: Gen) Gen {
        if (self.is_base) {
            return base(self.ext.c0.square());
        }
        return fromExt(self.ext.square());
    }

    pub fn inverse(self: Gen) Gen {
        if (self.is_base) {
            return base(self.ext.c0.inv());
        }
        return fromExt(self.ext.inv());
    }

    pub fn div(self: Gen, other: Gen) Gen {
        return self.mul(other.inverse());
    }

    pub fn isZero(self: Gen) bool {
        if (self.is_base) return self.ext.c0.isZero();
        return self.ext.isZero();
    }
};

pub fn expFieldElem(base: Gen, exp: usize) Gen {
    var result = Gen.one();
    var b = base;
    var e = exp;
    while (e != 0) : (e >>= 1) {
        if ((e & 1) == 1) result = result.mul(b);
        b = b.square();
    }
    return result;
}

pub fn expFieldInt(base: Field, exp: i64) Field {
    if (exp < 0) {
        return base.inv().pow(@as(u64, @intCast(-exp)));
    }
    return base.pow(@as(u64, @intCast(exp)));
}

pub fn computeAnnihilator(r: Gen, n: usize) Gen {
    return expFieldElem(r, n).sub(Gen.one());
}

pub fn rootOfUnityBy(n: usize) Field {
    if (n == 0 or !isPowerOfTwo(n)) {
        @panic("n must be positive and a power of two");
    }

    const log_n = log2PowerOfTwo(n);
    if (log_n > MaxOrderRoot) {
        @panic("requested root of unity is too large");
    }

    var res = f(RootOfUnityValue);
    var i = log_n;
    while (i < MaxOrderRoot) : (i += 1) {
        res = res.square();
    }
    return res;
}

pub fn evalCancellationAtPoint(comptime n: usize, r: Gen, cancelled: []const i64) Gen {
    if (cancelled.len == 0) return Gen.one();

    const omega = rootOfUnityBy(n);
    var result = Gen.one();
    for (cancelled) |pos| {
        const norm = if (pos < 0) @as(i64, @intCast(n)) + pos else pos;
        const omega_k = expFieldInt(omega, norm);
        result = result.mul(r.sub(Gen.base(omega_k)));
    }
    return result;
}

fn isPowerOfTwo(n: usize) bool {
    return n != 0 and (n & (n - 1)) == 0;
}

fn log2PowerOfTwo(n: usize) usize {
    var x = n;
    var log: usize = 0;
    while (x > 1) : (x >>= 1) {
        log += 1;
    }
    return log;
}

test "field arithmetic" {
    try std.testing.expectEqual(@as(u64, 3), f(1).add(f(2)).value);
    try std.testing.expectEqual(@as(u64, Modulus - 1), f(1).sub(f(2)).value);
    try std.testing.expectEqual(@as(u64, 6), f(2).mul(f(3)).value);
    try std.testing.expectEqual(@as(u64, Modulus - 1), f(1).neg().value);
    try std.testing.expectEqual(@as(u64, 1), f(7).div(f(7)).value);
}

test "extension arithmetic" {
    const a = Ext.init(f(1), f(2), f(3), f(4));
    const b = Ext.init(f(5), f(6), f(7), f(8));
    const product = a.mul(b);

    try std.testing.expectEqual(@as(u64, 197), product.c0.value);
    try std.testing.expectEqual(@as(u64, 133), product.c1.value);
    try std.testing.expectEqual(@as(u64, 142), product.c2.value);
    try std.testing.expectEqual(@as(u64, 60), product.c3.value);
    try std.testing.expect(a.mul(a.inv()).sub(Ext.one()).isZero());
}

test "root helpers" {
    const omega_8 = rootOfUnityBy(8);
    try std.testing.expect(omega_8.pow(8).sub(f(1)).isZero());
    try std.testing.expect(rootOfUnityBy(16).square().sub(omega_8).isZero());
}

test "generic field helpers" {
    const r = Gen.fromExt(Ext.init(f(2), f(3), f(4), f(5)));
    try std.testing.expect(expFieldElem(r, 0).sub(Gen.one()).isZero());
    try std.testing.expect(expFieldElem(r, 3).sub(r.mul(r).mul(r)).isZero());

    const cancelled = [_]i64{ 0, -1 };
    const c = evalCancellationAtPoint(8, Gen.base(f(9)), &cancelled);
    const expected = Gen.base(f(9).sub(rootOfUnityBy(8).pow(0)))
        .mul(Gen.base(f(9).sub(rootOfUnityBy(8).pow(7))));
    try std.testing.expect(c.sub(expected).isZero());
}
