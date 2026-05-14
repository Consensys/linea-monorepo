const field = @import("../field/koalabear.zig");

pub fn evaluateHorner(coefficients: []const field.Element, point: field.Element) field.Element {
    var acc = field.Element.zero();
    var i = coefficients.len;
    while (i != 0) {
        i -= 1;
        acc = acc.mul(point).add(coefficients[i]);
    }
    return acc;
}
