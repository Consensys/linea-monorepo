const field = @import("../field/koalabear.zig");

pub const Digest = [8]field.Element;

pub fn compress(input: []const field.Element) Digest {
    var out = [_]field.Element{field.Element.zero()} ** 8;
    for (input, 0..) |value, i| {
        out[i % out.len] = out[i % out.len].add(value);
    }
    return out;
}
