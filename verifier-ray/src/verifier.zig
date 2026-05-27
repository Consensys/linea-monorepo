const proof_mod = @import("proof.zig");
const runtime_mod = @import("runtime.zig");
const generated = @import("generated/stub.zig");

pub const VerifyError = error{
    EmptyProof,
    InvalidProof,
};

pub fn verify(p: proof_mod.Proof) VerifyError!void {
    if (p.proof_bytes.len == 0 and p.commitments.len == 0 and p.public_inputs.len == 0 and
        p.columns.len == 0 and p.cells.len == 0 and p.eval_cells.len == 0)
    {
        return VerifyError.EmptyProof;
    }

    if (p.columns.len < generated.min_columns or
        p.cells.len < generated.min_cells or
        p.eval_cells.len < generated.min_eval_cells)
    {
        return VerifyError.InvalidProof;
    }

    for (generated.cell_is_extension, 0..) |is_extension, idx| {
        switch (p.cells[idx]) {
            .base => if (is_extension) return VerifyError.InvalidProof,
            .ext => if (!is_extension) return VerifyError.InvalidProof,
        }
    }

    var rt = runtime_mod.Runtime.initWithRoundCount(generated.round_count);
    generated.verifyGenerated(&rt, p) catch return VerifyError.InvalidProof;
}
