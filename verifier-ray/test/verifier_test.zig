const std = @import("std");
const verifier_ray = @import("verifier_ray");

const protocol = verifier_ray.protocol;
const verifier = verifier_ray.verifier;
const vanishing = verifier_ray.query.vanishing;

test "verify rejects proof with wrong round count" {
    const spec = protocol.Spec{
        .round_coin_counts = &[_]usize{ 0, 1 },
        .round_coin_offsets = &[_]usize{ 0, 0 },
        .total_round_coins = 1,
    };
    const systems = verifier.Systems{
        .vanishing = vanishing.System{
            .modules = &.{},
            .round_coin_counts = &[_]usize{ 0, 1 },
            .round_coin_offsets = &[_]usize{ 0, 0 },
            .total_round_coins = 1,
        },
    };
    try std.testing.expectError(
        error.InvalidRoundCount,
        verifier.verify(spec, systems, .{
            .rounds = &.{},
            .witness_claims = &.{},
            .quotient_claims = &.{},
        }),
    );
}
