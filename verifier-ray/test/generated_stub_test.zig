const std = @import("std");
const verifier_ray = @import("verifier_ray");

test "generated stub returns NoRounds for empty runtime" {
    var rt = verifier_ray.runtime.Runtime.init();
    try std.testing.expectError(
        verifier_ray.runtime.Error.NoRounds,
        verifier_ray.generated.stub.verifyGenerated(&rt, verifier_ray.proof.Proof.empty()),
    );
}

test "public verify rejects truncated generated payloads" {
    const proof = verifier_ray.proof.Proof{
        .commitments = &.{},
        .public_inputs = &.{},
        .proof_bytes = &.{0},
        .columns = &.{},
        .cells = &.{},
        .eval_cells = &.{},
    };

    try std.testing.expectError(verifier_ray.VerifyError.InvalidProof, verifier_ray.verify(proof));
}

test "public verify accepts proof populated only via generated payload fields" {
    const runtime = verifier_ray.runtime;
    const ext = verifier_ray.field.koalabear_ext;
    const base = verifier_ray.field.koalabear;

    const column_values = [_]base.Element{
        base.Element.init(1),
        base.Element.init(2),
        base.Element.init(3),
        base.Element.init(4),
    };
    const columns = [_]runtime.ColumnAssignment{
        .{
            .visibility = .oracle,
            .assignment = .{ .base = column_values[0..] },
        },
    };

    var rt = runtime.Runtime.initWithRoundCount(verifier_ray.generated.stub.round_count);
    var round1_coins: [1]runtime.Coin = undefined;
    _ = try rt.advanceRoundWithMessage(0, .{
        .columns = &.{},
        .cells = &.{},
        .next_round_coin_count = round1_coins.len,
    }, &round1_coins);

    var round2_coins: [1]runtime.Coin = undefined;
    const coins_r2 = try rt.advanceRoundWithMessage(1, .{
        .columns = columns[0..],
        .cells = &.{},
        .next_round_coin_count = round2_coins.len,
    }, &round2_coins);

    const quotient = ext.Ext.one();
    const evals = [_]runtime.Coin{
        coins_r2[0].pow(4).sub(ext.Ext.one()),
        ext.Ext.one(),
        quotient,
    };

    const proof = verifier_ray.proof.Proof{
        .commitments = &.{},
        .public_inputs = &.{},
        .proof_bytes = &.{},
        .columns = columns[0..],
        .cells = &.{},
        .eval_cells = evals[0..],
    };

    try verifier_ray.verify(proof);
}
