pub const modulus: u32 = 2_130_706_433;
pub const multiplicative_gen: u32 = 3;
pub const root_of_unity: u32 = 1_791_270_792;

pub const Element = struct {
    value: u32,

    pub fn init(raw: u64) Element {
        return .{ .value = @as(u32, @intCast(raw % modulus)) };
    }

    pub fn zero() Element {
        return .{ .value = 0 };
    }

    pub fn one() Element {
        return .{ .value = 1 };
    }

    pub fn eql(self: Element, rhs: Element) bool {
        return self.value == rhs.value;
    }

    pub fn isZero(self: Element) bool {
        return self.value == 0;
    }

    pub fn add(self: Element, rhs: Element) Element {
        return init(@as(u64, self.value) + @as(u64, rhs.value));
    }

    pub fn sub(self: Element, rhs: Element) Element {
        if (self.value >= rhs.value) {
            return .{ .value = self.value - rhs.value };
        }
        return .{ .value = modulus - (rhs.value - self.value) };
    }

    pub fn neg(self: Element) Element {
        if (self.isZero()) return self;
        return .{ .value = modulus - self.value };
    }

    pub fn mul(self: Element, rhs: Element) Element {
        return init(@as(u64, self.value) * @as(u64, rhs.value));
    }

    pub fn square(self: Element) Element {
        return self.mul(self);
    }

    pub fn pow(self: Element, exponent: u64) Element {
        var result = Element.one();
        var base = self;
        var exp = exponent;
        while (exp != 0) : (exp >>= 1) {
            if ((exp & 1) == 1) {
                result = result.mul(base);
            }
            base = base.square();
        }
        return result;
    }

    pub fn inverse(self: Element) Element {
        return self.pow(modulus - 2);
    }
};
