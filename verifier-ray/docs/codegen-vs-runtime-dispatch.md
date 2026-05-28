# Codegen vs Runtime Dispatch: PR #3230 vs PR #3202

This document compares two approaches to the global-constraint verifier in
`verifier-ray`, using a concrete example constraint to make the differences
tangible.

---

## The Example Constraint

```
A · (B − 1) · (C + D²)
```

Four oracle columns, one vanishing constraint. The verifier must evaluate this
expression at a random point `r` and check the quotient identity
`P_agg(r) = (r^n − 1) · Q(r)`.

---

## PR #3202 — Runtime Dispatch (`verifier-ray/global-compiler`)

### What it does

The Go side compiles the protocol and exports the expression tree as data via
`VerifierExport`. The Zig binary stores that tree and interprets it for every
proof at runtime.

### Expression representation

The expression graph is a flat `[]const ExprNode` array carried inside the
binary. Interior nodes reference their children by integer index.

```zig
pub const ExprOp = struct {
    operator: Operator,
    operands: []const usize,  // indices into the expressions array
};

pub const ExprNode = union(enum) {
    column_view: ColumnView,    // leaf: column index + shift
    constant:    field.Element, // leaf: a field constant
    op:          ExprOp,        // interior node
};
```

For `A·(B−1)·(C+D²)` the array in the binary at runtime:

```
expressions[0]  column_view D
expressions[1]  op{ square, operands=[0] }       → D²
expressions[2]  column_view C
expressions[3]  op{ add,    operands=[2,1] }     → C + D²
expressions[4]  column_view B
expressions[5]  constant(1)
expressions[6]  op{ sub,    operands=[4,5] }     → B − 1
expressions[7]  op{ mul,    operands=[6,3] }     → (B−1)·(C+D²)
expressions[8]  column_view A
expressions[9]  op{ mul,    operands=[8,7] }     → A·(B−1)·(C+D²)
```

### Interpreter (runs per proof, exact code from PR #3202)

```zig
fn evalExprAtPoint(self: CompiledModule, expr_index: usize, input: CheckInput) Error!ext.Ext {
    return switch (self.spec.expressions[expr_index]) {  // dispatch per node
        .column_view => |view| blk: {
            const i = self.findWitnessView(view) orelse return Error.InvalidExpression;
            break :blk input.witness_claims[self.witness_claim_offset + i];
        },
        .constant => |c| ext.Ext.lift(c),
        .op       => |o| try self.evalOpAtPoint(o, input),
    };
}

fn evalOpAtPoint(self: CompiledModule, op: ExprOp, input: CheckInput) Error!ext.Ext {
    return switch (op.operator) {
        .add    => (try self.evalExprAtPoint(op.operands[0], input))
                       .add(try self.evalExprAtPoint(op.operands[1], input)),
        .sub    => (try self.evalExprAtPoint(op.operands[0], input))
                       .sub(try self.evalExprAtPoint(op.operands[1], input)),
        .mul    => (try self.evalExprAtPoint(op.operands[0], input))
                       .mul(try self.evalExprAtPoint(op.operands[1], input)),
        .square => (try self.evalExprAtPoint(op.operands[0], input)).square(),
        // ... div, double, negate, inverse
    };
}
```

For `A·(B−1)·(C+D²)`: **10 recursive calls + 10 switch dispatches** per proof,
plus the 4 arithmetic operations the constraint actually requires.

---

## PR #3230 (This Work) — Compile-time Codegen (`verifier-ray/global-constraint`)

### What it does

The Go codegen (`emitExprCode` in `codegen/generate.go`) walks the expression
tree once at build time and writes flat Zig arithmetic into
`src/generated/stub.zig`. The tree is consumed by Go and is gone before the
Zig binary is compiled.

### Code generation (runs once at build time, in Go)

```go
// emitExprCode recursively traverses wiop.Expression and returns a Zig string.
// Runs during `make generate-stub`, not at verification time.
func emitExprCode(expr wiop.Expression, gv *global.Verifier) (string, error) {
    switch e := expr.(type) {
    case *wiop.ColumnView:
        pos := gv.WitnessClaimEvalCellPos(e)
        return fmt.Sprintf("proof.eval_cells[%d]", pos), nil
    case *wiop.Constant:
        return "ext.Ext.one()", nil  // simplified for value=1
    case *wiop.ArithmeticOperation:
        left,  _ := emitExprCode(e.Operands[0], gv)
        right, _ := emitExprCode(e.Operands[1], gv)
        return fmt.Sprintf("(%s).%s(%s)", left, ops[e.Operator], right), nil
    }
}
```

### Generated Zig (runs per proof, no dispatch)

```zig
// === Round 2 ===
const _n0: usize = 4;

const _v0_0_0 = (proof.eval_cells[0]).mul(
    ((proof.eval_cells[1]).sub(ext.Ext.one())).mul(
        (proof.eval_cells[2]).add((proof.eval_cells[3]).square())
    )
);

const _c0_0_0 = ext.Ext.one();
const _pc0_0_0 = _v0_0_0.mul(_c0_0_0);

const _ve0_0 = [_]ext.Ext{ _pc0_0_0 };
try gc.verify(_n0, coins_r1[0], coins_r2[0], &_ve0_0, proof.eval_cells[4..8]);
```

For `A·(B−1)·(C+D²)`: **4 arithmetic calls**, nothing else. No branches, no
recursion, no data structure reads.

---

## Steps 2–5: Cancellation, Aggregation, Quotient, Identity Check

Steps 1–5 are defined in [global-constraint.md](./global-constraint.md).
Step 1 (expression evaluation) is covered above. This section covers the rest.

### Step 2 — Cancellation: `P_j(r) = V_j(r) · ∏_{k ∈ cancelled_j}(r − ω_n^k)`

The cancelled-positions set is fixed by the circuit and never changes per proof.

**PR #3202** stores cancelled positions as a runtime slice and calls
`evalCancellation` unconditionally — one function call and one branch per
vanishing, even when no rows are cancelled.

**PR #3230** resolves two cases at compile time:

- **Case 1 — no cancelled rows** (constraints referencing only unshifted
  columns, e.g. `A·(A−1) = 0`): emits `ext.Ext.one()` — no function call, no
  branch.
- **Case 2 — log-derivative sums** (`cancelled = [0]`, row 0 only): `ω_n^0 = 1`
  regardless of module size `n`, so the factor is inlined as
  `r.sub(ext.Ext.one())`.

All other cases (e.g. local/transition constraints that also cancel the last row)
require `gc.evalCancellation` at runtime, because `ω_n^{n−1}` depends on `n`
which is not known until the proof is presented.

---

### Steps 3–5 — Aggregation, Quotient, Identity Check

The field arithmetic is identical in both PRs. For large vanishing counts the
loops run unconditionally regardless; the per-iteration cost difference is the
Step 1 expression-eval overhead, not the loop machinery.

---

## Side-by-side: What Runs Per Proof

| | PR #3202 (dispatch) | PR #3230 (codegen, this work) |
|---|---|---|
| **Step 1 — expression eval** | | |
| Expression tree in binary | yes — `ExprNode[]` array | no |
| Array reads per eval | 1 per node (10 here) | 0 |
| Switch dispatches per eval | 1 per node (10 here) | 0 |
| Recursive calls per eval | 1 per node (10 here) | 0 |
| Arithmetic ops | 4 | 4 |
| **Step 2 — cancellation** | | |
| `cancelled = []` (no shift) | `evalCancellationAtPoint` called, returns `Ext.one()` | `ext.Ext.one()` literal, no call |
| `cancelled = [0]` (log-derivative, row 0) | `evalCancellationAtPoint` called | `r.sub(ext.Ext.one())` inlined |
| other cancelled positions | `evalCancellationAtPoint` called | `gc.evalCancellation` called |
| **Steps 3–5 — aggregate / quotient / identity** | | |
| Implementation | inline in `checkQuotientIdentity` | delegates to `gc.verify` |
| `n` source | always runtime (from spec data) | literal (static) or proof len (dynamic) |
| Loop counts | runtime (from spec data) | compile-time-known at call site (may unroll for small counts) |
| **Overall** | | |
| Protocol IR in binary | yes | no |
| Binary size | larger (IR + interpreter) | smaller (arithmetic only) |
| Generated source size | small fixed runtime | grows with protocol |

---

## Tradeoffs

### Flexibility vs. fixed output

PR #3202's Zig runtime is generic — change the Go constraint definition and the
Zig verifier adapts without regenerating any file. PR #3230 requires running
`make generate-stub` and committing the result whenever the protocol changes.
The `verify-stub` Makefile target catches drift in CI.

### Development friction

PR #3202 has no codegen step; the interpreter works for any protocol structure
immediately. PR #3230 adds a build-time dependency (Go codegen → Zig source →
Zig compiler) and a committed generated file that must stay in sync.

### Cost inside a zkVM

A zkVM (e.g. SP1, RISC Zero) proves a RISC-V execution trace. Every instruction
— including branches, memory reads, and function calls — becomes a constraint.

For `A·(B−1)·(C+D²)`, approximate RISC-V instruction cost:

| Step | PR #3202 | PR #3230 (this work)|
|---|---|---|
| Array bounds check + load | ~3–5 per node | 0 |
| Switch dispatch | ~3–5 per node | 0 |
| Stack frame (call/return) | ~4–8 per node | 0 |
| Ext field arithmetic | ~50–200 per op | ~50–200 per op |

The arithmetic cost is identical. The dispatch overhead adds roughly 10–18
instructions per expression node on top. For a real protocol with thousands of
expression nodes this multiplies linearly.

### Memory consistency proofs

Reading `expressions[expr_index]` is a random-index array access. In
Merkle-tree-based zkVMs each such access requires a Merkle path proof to
demonstrate the read value matches committed memory. For 10 nodes that is 10
Merkle paths per constraint evaluation. PR #3230 has zero such accesses: every
index is a literal in the instruction stream.

### Precompile circuits

The README targets a zkVM precompile — a hand-optimized arithmetic circuit, not
RISC-V interpretation. In that model a switch dispatch compiles to multiplexer
gates and a memory read becomes a lookup argument, both costing real gates. In
a precompile the dispatch overhead is on equal footing with arithmetic, not
cheaper.

### Verification key and binary size

In PLONK-style circuits the verification key encodes circuit structure. More
gates = larger key. PR #3202's interpreter gates appear in the key regardless
of which protocol is verified. PR #3230 produces a different, smaller circuit
per protocol with no shared interpreter overhead.

### Testing and correctness

PR #3202 is described by its author as an "interim/golden-vector verifier" —
useful for validating that prover-ray outputs align with the Zig algebra. In
that role the interpreter overhead is irrelevant and the flexibility of not
needing a codegen step is valuable.

---

## Which approach is better?

The answer depends on which use case you are optimising for.

### For the production zkVM precompile

**PR #3230 (codegen) is strictly better**, for all five steps.

Step 1 accounts for most of the difference (see expression-eval section above).
For Steps 2–5 both PRs implement the same algorithm with identical field
arithmetic (PR #3202 inline in `checkQuotientIdentity`; PR #3230 via `gc.verify`).
PR #3230's advantages are:

- **Step 2**: two cases are fully compile-time (no function call, no branch):
  unshifted-only constraints (`cancelled = []`) and log-derivative sums
  (`cancelled = [0]`). All other cancellations still call `gc.evalCancellation`.
- **Steps 3–4**: fixed-size arrays at the call site may allow loop unrolling for
  small counts (ratio 1–4 for quotient shares is typical). For large vanishing
  counts the loop runs regardless; the per-iteration saving comes from Step 1.
- **Step 5**: static modules emit a literal `n`, so `r.pow(n)` has a
  compile-time-known exponent.

In a zkVM precompile there is no branch predictor — every conditional costs the
same as arithmetic, and a fixed-DAG circuit cannot represent runtime branches.
PR #3230's fixed loop counts and elided branches are therefore mandatory, not
an optional optimisation.

### For development and testing

**PR #3202 (runtime dispatch) is better** — not for performance, but for
iteration speed. There is no codegen step, no generated file to commit, and the
Zig binary adapts to any protocol change immediately. PR #3202's author
describes it as an interim verifier for validating alignment between prover-ray
and the Zig field algebra. In that role the dispatch overhead is irrelevant.

### Recommendation

The two PRs are not alternatives for the same role — they are different tools:

- **PR #3202**: use as a reference / golden-vector verifier during development
  to validate that the prover produces outputs the Zig algebra agrees with.
- **PR #3230**: the production path. Once the real circuit is wired into
  `buildSystem()`, the generated stub is the verifier that goes into the zkVM.

They can coexist: PR #3202's test vectors can serve as an independent oracle to
cross-check PR #3230's generated Zig. A proof that passes both gives much higher
confidence than either alone.

---

## Summary

| Characteristic | PR #3202 (dispatch) | PR #3230 (codegen, this work) |
|---|---|---|
| Protocol flexibility | any protocol, no regen needed | fixed at codegen time, regen on change |
| Development iteration | fast — no build step | slower — regen + commit cycle |
| Expression tree at runtime | present | absent |
| zkVM instruction overhead | scales with expression nodes | zero |
| Random memory reads in zkVM | one per expression node | zero |
| Precompile gate overhead | multiplexer + lookup gates | none |
| Binary size | larger | smaller |
| Generated source size | fixed | grows with protocol |
| Stated purpose | interim testing / alignment | production verifier |
