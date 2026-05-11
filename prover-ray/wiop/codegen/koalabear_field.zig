const std = @import("std");

pub const Modulus: u64 = 2130706433;

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

    pub fn pow(self: Field, exponent: u64) Field {
        var result = f(1);
        var base = self;
        var e = exponent;
        while (e != 0) : (e >>= 1) {
            if ((e & 1) == 1) result = result.mul(base);
            base = base.mul(base);
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

test "field arithmetic" {
    try std.testing.expectEqual(@as(u64, 3), f(1).add(f(2)).value);
    try std.testing.expectEqual(@as(u64, Modulus - 1), f(1).sub(f(2)).value);
    try std.testing.expectEqual(@as(u64, 6), f(2).mul(f(3)).value);
    try std.testing.expectEqual(@as(u64, Modulus - 1), f(1).neg().value);
    try std.testing.expectEqual(@as(u64, 1), f(7).div(f(7)).value);
}
