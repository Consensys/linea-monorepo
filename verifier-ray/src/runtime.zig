const fiat_shamir = @import("crypto/fiat_shamir.zig");
const ext = @import("field/koalabear_ext.zig");
const value = @import("field/value.zig");

pub const Error = error{
    NoRounds,
    UnexpectedRound,
    LastRound,
    OutputTooSmall,
};

/// For the verifier, there are only two meaningful visibilty, the internal visisbilty
/// of prover-ray is not relevant. The `oracle` visibility is for columns that are only visible to the prover, and the `public` visibility is for columns that are visible to both the prover and verifier.  
pub const Visibility = enum(u8) {
    oracle = 0,
    public = 1,
};

pub const Vector = value.Vector;
pub const Scalar = value.Scalar;
pub const Coin = ext.Ext;

pub const ColumnAssignment = struct {
    visibility: Visibility,
    assignment: Vector,
};

/// Verifier-visible data sent before deriving the next round's coins.
/// Columns and cells are expected to be assigned before the round advances.
pub const RoundMessage = struct {
    columns: []const ColumnAssignment = &.{},
    cells: []const Scalar = &.{},
    next_round_coin_count: usize = 0,
};

pub const Runtime = struct {
    transcript: fiat_shamir.Transcript,
    current_round: usize,
    total_rounds: usize,

    pub fn init() Runtime {
        return initWithRoundCount(0);
    }

    pub fn initWithRoundCount(round_count: usize) Runtime {
        return .{
            .transcript = fiat_shamir.Transcript.init(),
            .current_round = 0,
            .total_rounds = round_count,
        };
    }

    /// Absorb one round's verifier-visible messages, advance the round counter,
    /// and derive the next round's extension-field coins.
    pub fn advanceRoundWithMessage(
        self: *Runtime,
        expected_round: usize,
        message: RoundMessage,
        out_coins: []Coin,
    ) Error![]const Coin {
        if (self.total_rounds == 0) return Error.NoRounds;
        if (self.current_round != expected_round or expected_round >= self.total_rounds) {
            return Error.UnexpectedRound;
        }
        if (self.current_round + 1 >= self.total_rounds) return Error.LastRound;
        if (message.next_round_coin_count > out_coins.len) return Error.OutputTooSmall;

        for (message.columns) |column| {
            self.transcript.absorbVector(column.assignment);
        }
        for (message.cells) |cell| {
            self.transcript.absorbScalar(cell);
        }

        self.current_round += 1;
        const coins = out_coins[0..message.next_round_coin_count];
        for (coins) |*coin| {
            coin.* = self.transcript.randomExt();
        }
        return coins;
    }
};
