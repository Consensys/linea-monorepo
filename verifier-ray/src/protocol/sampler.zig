const fiat_shamir = @import("../crypto/fiat_shamir.zig");
const types = @import("types.zig");

pub const Coin = types.Coin;
pub const RoundMessage = types.RoundMessage;

/// Returns a comptime-only coin-sampling type parametrized by per-round squeeze
/// counts. The runtime Fiat-Shamir transcript is owned by the caller and passed
/// by pointer, keeping comptime protocol structure and runtime state separated.
///
/// Because there is no runtime state, call ordering is guaranteed entirely by
/// the `inline for` in `protocol.replay` — no runtime check is needed.
pub fn Sampler(comptime advance_counts: []const usize) type {
    return struct {
        /// Absorbs one round's verifier-visible messages into `transcript`,
        /// then writes exactly `advance_counts[round_index]` coins into `out`.
        ///
        /// `round_index` is comptime-known; an out-of-bounds index is a
        /// compile error. The caller's `inline for` guarantees rounds are
        /// advanced in order — no runtime ordering check is performed.
        pub fn advanceRoundWithMessage(
            comptime round_index: usize,
            transcript: *fiat_shamir.Transcript,
            message: RoundMessage,
            out: *[advance_counts[round_index]]Coin,
        ) void {
            for (message.columns) |entry| {
                switch (entry) {
                    .oracle_commitment => |c| transcript.updateElements(&c),
                    .public_column => |col| transcript.absorbVector(col),
                }
            }
            for (message.cells) |cell| transcript.absorbScalar(cell);
            for (out) |*coin| coin.* = transcript.randomExt();
        }
    };
}
