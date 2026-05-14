const field = @import("../field/koalabear.zig");

pub const Error = error{Unsupported};

pub fn evaluate(_: []const field.Element, _: field.Element) Error!field.Element {
    return Error.Unsupported;
}
