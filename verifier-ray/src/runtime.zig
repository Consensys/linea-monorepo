const fiat_shamir = @import("crypto/fiat_shamir.zig");
const field = @import("field/koalabear.zig");
const ext = @import("field/koalabear_ext.zig");

pub const Error = error{
    NoRounds,
    UnexpectedRound,
    LastRound,
    MissingCellAssignment,
    OutputTooSmall,
};

pub const Visibility = enum(u8) {
    oracle = 0,
    public = 1,
};

pub const Vector = union(enum) {
    base: []const field.Element,
    ext: []const ext.Ext,

    fn absorb(self: Vector, transcript: *fiat_shamir.Transcript) void {
        switch (self) {
            .base => |values| transcript.updateElements(values),
            .ext => |values| transcript.updateExt(values),
        }
    }
};

pub const Scalar = union(enum) {
    base: field.Element,
    ext: ext.Ext,

    fn absorb(self: Scalar, transcript: *fiat_shamir.Transcript) void {
        switch (self) {
            .base => |value| transcript.updateElement(value),
            .ext => |value| transcript.updateExt(&.{value}),
        }
    }
};

pub const ColumnAssignment = struct {
    visibility: Visibility,
    assignment: Vector,
};

/// Verifier-visible data sent before deriving the next round's coins.
pub const RoundMessage = struct {
    columns: []const ColumnAssignment = &.{},
    cells: []const ?Scalar = &.{},
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
        out_coins: []ext.Ext,
    ) Error![]const ext.Ext {
        if (self.total_rounds == 0) return Error.NoRounds;
        if (self.current_round != expected_round or expected_round >= self.total_rounds) {
            return Error.UnexpectedRound;
        }
        if (self.current_round + 1 >= self.total_rounds) return Error.LastRound;
        if (message.next_round_coin_count > out_coins.len) return Error.OutputTooSmall;

        for (message.cells) |cell| {
            if (cell == null) return Error.MissingCellAssignment;
        }

        for (message.columns) |column| {
            column.assignment.absorb(&self.transcript);
        }
        for (message.cells) |cell| {
            cell.?.absorb(&self.transcript);
        }

        self.current_round += 1;
        const coins = out_coins[0..message.next_round_coin_count];
        for (coins) |*coin| {
            coin.* = self.transcript.randomExt();
        }
        return coins;
    }
};
