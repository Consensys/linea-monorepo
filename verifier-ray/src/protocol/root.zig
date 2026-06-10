const types = @import("types.zig");
const sampler_mod = @import("sampler.zig");
const fiat_shamir = @import("../crypto/fiat_shamir.zig");

pub const Error = error{InvalidRoundCount};

pub const Visibility = types.Visibility;
pub const Vector = types.Vector;
pub const Scalar = types.Scalar;
pub const Coin = types.Coin;
pub const Commitment = types.Commitment;
pub const ColumnMessage = types.ColumnMessage;
pub const RoundMessage = types.RoundMessage;

/// Compile-time parametric Fiat-Shamir coin sampler. Exposed so that callers
/// (tests, codegen) can drive the transcript directly when needed.
pub const Sampler = sampler_mod.Sampler;

/// Compile-time coin-routing specification shared across all sub-verifiers.
/// Extracted from the compiled IOP system by the Go codegen and emitted as a
/// standalone constant in the generated file alongside `verifier_mod.Systems`.
pub const Spec = struct {
    /// Number of coins squeezed after each round. Index 0 is always 0;
    /// the first coins are derived after the first round message is absorbed.
    round_coin_counts: []const usize,
    /// Starting position of each round's coins in the flat `all_coins` array.
    round_coin_offsets: []const usize,
    /// Total number of coins across all rounds; length of `all_coins`.
    total_round_coins: usize,
};

/// All protocol-level data derived from a proof by the higher-level verifier.
/// Produced by `replay`; consumed by sub-verifiers.
pub const Context = struct {
    /// All Fiat-Shamir coins derived across every round, laid out flat.
    /// Indexed by the compiled system's `round_coin_offsets`.
    all_coins: []const Coin,
    /// The verifier-visible round messages bound into the shared transcript.
    /// Sub-verifiers read cell openings directly from `rounds[i].cells`.
    rounds: []const RoundMessage,
};

/// Replays the prover–verifier transcript to derive all Fiat-Shamir coins.
///
/// For each round i, absorbs `rounds[i]` (oracle commitments, public columns,
/// cell scalars) into the Poseidon2 Merkle-Damgård hasher, then squeezes
/// exactly `advance_counts[i]` coins and stores them at
/// `all_coins[coin_offsets[i+1] .. coin_offsets[i+1] + advance_counts[i]]`.
///
/// Parameters expected by callers:
///   `advance_counts`  — `round_coin_counts[1..]`: squeeze counts starting from
///                       round 1 (round 0 always produces 0 coins).
///   `coin_offsets`    — `round_coin_offsets[1..]`: start positions for each
///                       round's coins; same length as `advance_counts`.
///   `total_coins`     — `round_coin_offsets[0].total_round_coins`; length of the
///                       returned array.
pub fn replay(
    comptime advance_counts: []const usize,
    comptime coin_offsets: []const usize,
    comptime total_coins: usize,
    rounds: []const RoundMessage,
) Error![total_coins]Coin {
    if (rounds.len != advance_counts.len) return error.InvalidRoundCount;

    comptime if (coin_offsets.len != advance_counts.len)
        @compileError("coin_offsets must have the same length as advance_counts");
    comptime for (0..advance_counts.len) |i| {
        if (coin_offsets[i] + advance_counts[i] > total_coins)
            @compileError("coin_offsets/advance_counts inconsistent with total_coins");
    };

    var transcript = fiat_shamir.Transcript.init();
    var all_coins: [total_coins]Coin = undefined;

    inline for (0..advance_counts.len) |advance_index| {
        const offset = coin_offsets[advance_index];
        const count = advance_counts[advance_index];
        Sampler(advance_counts).advanceRoundWithMessage(advance_index, &transcript, rounds[advance_index], all_coins[offset..][0..count]);
    }

    return all_coins;
}
