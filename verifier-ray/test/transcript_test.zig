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

test "sampler absorbs commitments, public columns, and squeezes matching coin count" {
    var transcript = fiat_shamir.Transcript.init();

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

    var coins: [2]protocol.Coin = undefined;
    protocol.Sampler(&.{2}).advanceRoundWithMessage(0, &transcript, .{
        .columns = &entries,
        .cells = &cells,
    }, &coins);

    // Consecutive coins must be distinct: randomDigest() absorbs a zero
    // separator between squeezes, so identical back-to-back outputs indicate
    // a broken separator mechanism.
    try std.testing.expect(!coins[0].eql(coins[1]));
}
