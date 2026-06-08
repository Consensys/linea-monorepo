const fiat_shamir = @import("../crypto/fiat_shamir.zig");
const types = @import("types.zig");

pub const Coin = types.Coin;
pub const RoundMessage = types.RoundMessage;

pub const Error = error{UnexpectedRound};

/// Returns a coin-sampling type whose round boundaries and per-round squeeze
/// counts are fixed at compile time.
///
/// `advance_counts[i]` is the number of Fiat-Shamir extension-field coins
/// derived after absorbing round i's transcript messages and cells. The total
/// number of valid round transitions equals `advance_counts.len`; calling
/// `advanceRoundWithMessage` with an out-of-range index is a compile error.
pub fn Sampler(comptime advance_counts: []const usize) type {
    return struct {
        transcript: fiat_shamir.Transcript,
        current_round: usize,

        const Self = @This();

        pub fn init() Self {
            return .{
                .transcript = fiat_shamir.Transcript.init(),
                .current_round = 0,
            };
        }

        /// Absorbs one round's verifier-visible messages, advances the round
        /// counter, and squeezes exactly `advance_counts[round_index]` coins.
        ///
        /// `round_index` is comptime-known; an out-of-bounds index is a
        /// compile error. Calling with a round index that does not equal the
        /// current round returns `UnexpectedRound`.
        pub fn advanceRoundWithMessage(
            self: *Self,
            comptime round_index: usize,
            message: RoundMessage,
        ) Error![advance_counts[round_index]]Coin {
            if (self.current_round != round_index) return Error.UnexpectedRound;

            for (message.columns) |entry| {
                switch (entry) {
                    .oracle_commitment => |c| self.transcript.updateElements(&c),
                    .public_column => |col| self.transcript.absorbVector(col),
                }
            }
            for (message.cells) |cell| {
                self.transcript.absorbScalar(cell);
            }

            self.current_round += 1;
            var coins: [advance_counts[round_index]]Coin = undefined;
            for (&coins) |*coin| coin.* = self.transcript.randomExt();
            return coins;
        }
    };
}
