const field = @import("../field/koalabear.zig");
const ext = @import("../field/koalabear_ext.zig");

pub const Error = field.Error || error{InvalidCardinality};

pub fn evaluateBaseAtBase(values: []const field.Element, point: field.Element) Error!field.Element {
    if (values.len == 0) return field.Element.zero();
    try ensureCardinality(values.len);

    const omega = try field.rootOfUnityBy(values.len);
    const inv_n = field.Element.init(values.len).inverse();
    const vanishing = point.pow(@intCast(values.len)).sub(field.Element.one());

    var omega_i = field.Element.one();
    var sum = field.Element.zero();
    for (values) |value| {
        const weighted = omega_i.mul(inv_n).mul(value);
        const inv_denom = point.sub(omega_i).inverse();
        sum = sum.add(weighted.mul(inv_denom));
        omega_i = omega_i.mul(omega);
    }

    return vanishing.mul(sum);
}

pub fn evaluateBaseAtExt(values: []const field.Element, point: ext.Ext) Error!ext.Ext {
    if (values.len == 0) return ext.Ext.zero();
    try ensureCardinality(values.len);

    const omega = try field.rootOfUnityBy(values.len);
    const inv_n = field.Element.init(values.len).inverse();
    const vanishing = point.pow(@intCast(values.len)).sub(ext.Ext.one());

    var omega_i = field.Element.one();
    var sum = ext.Ext.zero();
    for (values) |value| {
        const weighted = omega_i.mul(inv_n).mul(value);
        const denom = point.sub(ext.Ext.lift(omega_i));
        sum = sum.add(denom.inverse().mulByBase(weighted));
        omega_i = omega_i.mul(omega);
    }

    return vanishing.mul(sum);
}

pub fn evaluateExtAtBase(values: []const ext.Ext, point: field.Element) Error!ext.Ext {
    if (values.len == 0) return ext.Ext.zero();
    try ensureCardinality(values.len);

    const omega = try field.rootOfUnityBy(values.len);
    const inv_n = field.Element.init(values.len).inverse();
    const vanishing = point.pow(@intCast(values.len)).sub(field.Element.one());

    var omega_i = field.Element.one();
    var sum = ext.Ext.zero();
    for (values) |value| {
        const weighted = value.mulByBase(omega_i.mul(inv_n));
        const inv_denom = point.sub(omega_i).inverse();
        sum = sum.add(weighted.mulByBase(inv_denom));
        omega_i = omega_i.mul(omega);
    }

    return sum.mulByBase(vanishing);
}

pub fn evaluateExtAtExt(values: []const ext.Ext, point: ext.Ext) Error!ext.Ext {
    if (values.len == 0) return ext.Ext.zero();
    try ensureCardinality(values.len);

    const omega = try field.rootOfUnityBy(values.len);
    const inv_n = field.Element.init(values.len).inverse();
    const vanishing = point.pow(@intCast(values.len)).sub(ext.Ext.one());

    var omega_i = field.Element.one();
    var sum = ext.Ext.zero();
    for (values) |value| {
        const weighted = value.mulByBase(omega_i.mul(inv_n));
        const denom = point.sub(ext.Ext.lift(omega_i));
        sum = sum.add(weighted.mul(denom.inverse()));
        omega_i = omega_i.mul(omega);
    }

    return vanishing.mul(sum);
}

pub fn evaluateBaseBatchAtBase(
    out: []field.Element,
    values: []const field.Element,
    points: []const field.Element,
) Error!void {
    if (out.len != points.len) return error.InvalidCardinality;
    for (points, out) |point, *dst| {
        dst.* = try evaluateBaseAtBase(values, point);
    }
}

pub fn evaluateBaseBatchAtExt(
    out: []ext.Ext,
    values: []const field.Element,
    points: []const ext.Ext,
) Error!void {
    if (out.len != points.len) return error.InvalidCardinality;
    for (points, out) |point, *dst| {
        dst.* = try evaluateBaseAtExt(values, point);
    }
}

pub fn evaluateExtBatchAtBase(
    out: []ext.Ext,
    values: []const ext.Ext,
    points: []const field.Element,
) Error!void {
    if (out.len != points.len) return error.InvalidCardinality;
    for (points, out) |point, *dst| {
        dst.* = try evaluateExtAtBase(values, point);
    }
}

pub fn evaluateExtBatchAtExt(
    out: []ext.Ext,
    values: []const ext.Ext,
    points: []const ext.Ext,
) Error!void {
    if (out.len != points.len) return error.InvalidCardinality;
    for (points, out) |point, *dst| {
        dst.* = try evaluateExtAtExt(values, point);
    }
}

pub fn evaluate(values: []const field.Element, point: field.Element) Error!field.Element {
    return evaluateBaseAtBase(values, point);
}

fn ensureCardinality(cardinality: usize) Error!void {
    if (!field.isPowerOfTwo(cardinality)) {
        return error.InvalidCardinality;
    }
    if (field.log2PowerOfTwo(cardinality) > field.max_order_root) {
        return error.InvalidCardinality;
    }
}
