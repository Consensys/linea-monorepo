const std = @import("std");
const verifier_ray = @import("verifier_ray");

const field = verifier_ray.field.koalabear;
const ext = verifier_ray.field.koalabear_ext;
const fiat_shamir = verifier_ray.crypto.fiat_shamir;
const protocol = verifier_ray.protocol;

test "transcript absorbs elements deterministically" {
    var transcript = fiat_shamir.Transcript.init();
    transcript.updateElements(&.{
        field.Element.init(3),
        field.Element.init(4),
    });

    const challenge = transcript.randomExt();
    try std.testing.expect(!challenge.isZero());
}

test "sampler rejects skipped and replayed rounds" {
    // Three advances; coin counts don't matter for the ordering check.
    var sampler = protocol.Sampler(&.{ 1, 1, 1 }).init();

    try std.testing.expectError(
        error.UnexpectedRound,
        sampler.advanceRoundWithMessage(1, .{}),
    );

    _ = try sampler.advanceRoundWithMessage(0, .{});

    try std.testing.expectError(
        error.UnexpectedRound,
        sampler.advanceRoundWithMessage(0, .{}),
    );
    try std.testing.expectError(
        error.UnexpectedRound,
        sampler.advanceRoundWithMessage(2, .{}),
    );
}

test "sampler absorbs commitments, public columns, and squeezes matching coin count" {
    var sampler = protocol.Sampler(&.{2}).init();

    const entries = [_]protocol.ColumnMessage{
        .{ .oracle_commitment = .{
            field.Element.init(1),
            field.Element.init(2),
            field.Element.init(3),
            field.Element.init(4),
            field.Element.init(5),
            field.Element.init(6),
            field.Element.init(7),
            field.Element.init(8),
        } },
        .{ .public_column = .{ .base = &.{field.Element.init(9)} } },
    };
    const cells = [_]protocol.Scalar{
        .{ .base = field.Element.init(10) },
    };
    const got = try sampler.advanceRoundWithMessage(0, .{
        .columns = &entries,
        .cells = &cells,
    });

    // The return type is [2]Coin; .len is comptime-known.
    try std.testing.expectEqual(@as(usize, 2), got.len);
    try std.testing.expectEqual(@as(usize, 1), sampler.current_round);
    // Consecutive coins must be distinct: randomDigest() absorbs a zero
    // separator between squeezes, so identical back-to-back outputs indicate
    // a broken separator mechanism.
    try std.testing.expect(!got[0].eql(got[1]));
}
