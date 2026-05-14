const base = @import("koalabear.zig");
const ext = @import("koalabear_ext.zig");

pub const ElementSlice = []const base.Element;
pub const ExtSlice = []const ext.Ext;

pub fn allZero(values: ElementSlice) bool {
    for (values) |value| {
        if (!value.isZero()) return false;
    }
    return true;
}
