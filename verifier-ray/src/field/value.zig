const base = @import("koalabear.zig");
const ext = @import("koalabear_ext.zig");

pub const ElementSlice = []const base.Element;
pub const ExtSlice = []const ext.Ext;

pub const Vector = union(enum) {
    base: ElementSlice,
    ext: ExtSlice,

    pub fn len(self: Vector) usize {
        return switch (self) {
            .base => |v| v.len,
            .ext => |v| v.len,
        };
    }
};

pub const Scalar = union(enum) {
    base: base.Element,
    ext: ext.Ext,
};

pub fn ElementArray(comptime len: usize) type {
    return [len]base.Element;
}

pub fn ExtArray(comptime len: usize) type {
    return [len]ext.Ext;
}

pub fn allZero(values: ElementSlice) bool {
    for (values) |value| {
        if (!value.isZero()) return false;
    }
    return true;
}

pub fn allZeroExt(values: ExtSlice) bool {
    for (values) |value| {
        if (!value.isZero()) return false;
    }
    return true;
}

pub fn batchInvertBase(out: []base.Element, values: []const base.Element) (base.Error || error{LengthMismatch})!void {
    if (out.len != values.len) return error.LengthMismatch;
    for (values, out) |value, *dst| {
        dst.* = if (value.isZero()) base.Element.zero() else try value.inverse();
    }
}

pub fn batchInvertExt(out: []ext.Ext, values: []const ext.Ext) (base.Error || error{LengthMismatch})!void {
    if (out.len != values.len) return error.LengthMismatch;
    for (values, out) |value, *dst| {
        dst.* = if (value.isZero()) ext.Ext.zero() else try value.inverse();
    }
}
