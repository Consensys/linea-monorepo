const std = @import("std");
const verifier_ray = @import("verifier_ray");

const field = verifier_ray.field.koalabear;
const protocol = verifier_ray.protocol;
const logderivativesum = verifier_ray.query.logderivativesum;

// A one-query LogDerivativeSum system with a single Z column: its final opening
// is read from (round 0, index 0) and the Result cell from (round 0, index 1).
// The golden fixtures already cover honest end-to-end proofs across every
// LogDerivativeSumCompiler/Lookup scenario; these hand-built cases pin the
// boundary checks (and their error paths) directly.
const z_finals = [_]logderivativesum.ScalarRef{.{ .round = 0, .index = 0 }};

fn oneQuerySystem(comptime result_is_zero: bool) logderivativesum.System {
    const queries = &[_]logderivativesum.Query{.{
        .z_finals = &z_finals,
        .result = .{ .round = 0, .index = 1 },
        .result_is_zero = result_is_zero,
    }};
    return .{ .queries = queries };
}

test "logderiv accepts a matching final sum" {
    const cells = [_]protocol.Scalar{
        .{ .base = field.Element.init(5) }, // z-final
        .{ .base = field.Element.init(5) }, // result
    };
    const rounds = [_]protocol.RoundMessage{.{ .cells = &cells }};
    const ctx = protocol.Context{ .all_coins = &.{}, .rounds = &rounds };
    try logderivativesum.verify(oneQuerySystem(false), ctx);
}

test "logderiv rejects a final sum that disagrees with Result" {
    const cells = [_]protocol.Scalar{
        .{ .base = field.Element.init(5) }, // z-final
        .{ .base = field.Element.init(7) }, // result
    };
    const rounds = [_]protocol.RoundMessage{.{ .cells = &cells }};
    const ctx = protocol.Context{ .all_coins = &.{}, .rounds = &rounds };
    try std.testing.expectError(
        error.FinalSumMismatch,
        logderivativesum.verify(oneQuerySystem(false), ctx),
    );
}

test "lookup rejects a non-zero aggregated result" {
    // Final-sum holds (3 == 3) so the FinalSumMismatch guard passes; the
    // result-is-zero guard must then reject the non-zero Result.
    const cells = [_]protocol.Scalar{
        .{ .base = field.Element.init(3) }, // z-final
        .{ .base = field.Element.init(3) }, // result
    };
    const rounds = [_]protocol.RoundMessage{.{ .cells = &cells }};
    const ctx = protocol.Context{ .all_coins = &.{}, .rounds = &rounds };
    try std.testing.expectError(
        error.LookupResultNonZero,
        logderivativesum.verify(oneQuerySystem(true), ctx),
    );
}

test "lookup accepts a zero aggregated result" {
    const cells = [_]protocol.Scalar{
        .{ .base = field.Element.init(0) }, // z-final
        .{ .base = field.Element.init(0) }, // result
    };
    const rounds = [_]protocol.RoundMessage{.{ .cells = &cells }};
    const ctx = protocol.Context{ .all_coins = &.{}, .rounds = &rounds };
    try logderivativesum.verify(oneQuerySystem(true), ctx);
}
