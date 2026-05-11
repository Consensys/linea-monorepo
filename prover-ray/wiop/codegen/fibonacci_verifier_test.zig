const std = @import("std");
const koalabear = @import("koalabear_field.zig");

const ModuleSize: usize = 8;
const Field = koalabear.Field;
const f = koalabear.f;

fn shifted(row: usize, offset: i64) usize {
    const n: i64 = @intCast(ModuleSize);
    const row_i: i64 = @intCast(row);
    return @intCast(@mod(row_i + offset, n));
}

fn isVanishing0Cancelled(row: usize) bool {
    return row == 0 or row == 1;
}

fn evalVanishing0(col0: *const [ModuleSize]Field, row: usize) Field {
    return col0[shifted(row, 0)].sub(col0[shifted(row, -1)]).sub(col0[shifted(row, -2)]);
}

fn checkVerifierAction0(col0: *const [ModuleSize]Field) !void {
    var row: usize = 0;
    while (row < ModuleSize) : (row += 1) {
        if (isVanishing0Cancelled(row)) continue;
        if (!evalVanishing0(col0, row).isZero()) {
            return error.VanishingConstraintFailed;
        }
    }
}

pub fn verifyColumnAssignment(col0: *const [ModuleSize]Field) bool {
    checkVerifierAction0(col0) catch return false;
    return true;
}

const honest = [_]Field{ f(1), f(1), f(2), f(3), f(5), f(8), f(13), f(21) };
const invalid = [_]Field{ f(1), f(1), f(2), f(3), f(5), f(8), f(13), f(22) };

test "generated verifier accepts the Fibonacci column assignment" {
    try checkVerifierAction0(&honest);
    try std.testing.expect(verifyColumnAssignment(&honest));
}

test "generated verifier rejects a broken Fibonacci column assignment" {
    try std.testing.expectError(error.VanishingConstraintFailed, checkVerifierAction0(&invalid));
    try std.testing.expect(!verifyColumnAssignment(&invalid));
}
