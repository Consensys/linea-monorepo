const field = @import("field/koalabear.zig");
const runtime = @import("runtime.zig");

pub const Commitment = [8]field.Element;

pub const Proof = struct {
    commitments: []const Commitment,
    public_inputs: []const field.Element,
    proof_bytes: []const u8,
    /// Oracle and public columns committed by the prover in each round,
    /// concatenated in round order. Indexed by the flat offset emitted by codegen.
    columns: []const runtime.ColumnAssignment,
    /// Scalar cell openings, concatenated in round order.
    cells: []const runtime.Scalar,
    /// Extension-field evaluation claims (witness and quotient evaluations at r),
    /// concatenated across all global-quotient verifier actions.
    ///
    /// Layout invariant: within each global-verifier bucket, witness claims occupy
    /// contiguous positions assigned by `WitnessClaimEvalCellPos`, followed by
    /// quotient claims in contiguous positions. Codegen reads positions directly
    /// from `ObjectID.Position()` on the corresponding cells, so the runtime layout
    /// must match the compile-time layout produced by `global.Compile`.
    eval_cells: []const runtime.Coin,

    pub fn empty() Proof {
        return .{
            .commitments = &.{},
            .public_inputs = &.{},
            .proof_bytes = &.{},
            .columns = &.{},
            .cells = &.{},
            .eval_cells = &.{},
        };
    }
};
