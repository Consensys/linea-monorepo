# Vortex Proof Size Optimization

This document describes the optimizations applied to reduce the proof size output by
`fullInitialCompilationSuite` — specifically the final KoalaBear Vortex step.

---

## Background

The final Vortex in `fullInitialCompilationSuite` (see `zkevm/full.go`) compresses the
accumulated wizard state into a proof. Its proof columns are the direct input to the
BN254/BLS12-377 PLONK verifier, so minimizing them reduces both proof size and constraint
count in the outer circuit.

The three main proof components are:

| Component | Description |
|---|---|
| `U_alpha` | Random linear combination of all committed polynomials |
| `SELECTED_COL` | Opened column values at K random positions |
| `MERKLEPROOF` | Sibling hashes for Merkle path verification |

---

## Optimization 1: Skip Duplicated Proof Columns

**Option:** `SkipSelfRecursionProofColumns()`

### What is duplicated?

The Vortex prover registers three opened-column proof objects:

- `SELECTED_COL` — all rounds combined (SIS + non-SIS), used by the verifier
- `SELECTED_COL_SIS` — SIS rounds only
- `SELECTED_COL_NON_SIS` — non-SIS rounds only

The relationship is:

```
SELECTED_COL = concat(SELECTED_COL_NON_SIS, SELECTED_COL_SIS)
```

The split exists because the **self-recursion** step needs SIS and non-SIS openings
separately — SIS columns are hashed via lattice-SIS and non-SIS columns via Poseidon2,
and these are verified independently. However, `SELECTED_COL` is also registered because
the overall verifier (Schwartz-Zippel check) needs the concatenated view.

### Why they are dead weight on the final Vortex

On the final Vortex (marked `PremarkAsSelfRecursed()`, no subsequent self-recursion),
`SELECTED_COL_SIS` and `SELECTED_COL_NON_SIS` are registered as proof columns but
**no verifier ever reads them**. The gnark verifier only consumes `SELECTED_COL`.

`SkipSelfRecursionProofColumns()` suppresses their registration entirely.

### Impact

This was the primary optimization that brought proof cells from ~540K to ~308K cells.

---

## Optimization 2: Reduce U_alpha Blowup Factor

**Option:** `WithUAlphaCoefficients()`

### Evaluation form vs. coefficient form

In standard Vortex, `U_alpha` is the random linear combination of all committed polynomials,
sent as a **Reed-Solomon codeword** of N evaluations over the extension field:

```
Eval mode:   N = T × RS_factor   extension-field elements
Coeff mode:  T                   extension-field elements
```

`WithUAlphaCoefficients()` switches to coefficient form: the prover sends T coefficients
instead of N evaluations. The gnark verifier reconstructs the N evaluations via a forward
FFT hint.

> **Important:** Each U_alpha cell is a **degree-4 KoalaBear extension field element**
> (`fext.Element`, 16 bytes), not a base field element (4 bytes). The savings from
> coefficient mode are 4× larger in bytes than the cell count suggests.

### Why only on the final Vortex

Coefficient mode requires a gnark FFT hint in the verifier circuit. The self-recursion
gnark circuit does not support this hint, so `WithUAlphaCoefficients()` can only be applied
to the final (pre-marked) Vortex.

### Interaction with column size

Enlarging the codeword (increasing RS factor or targeting a larger `T`) reduces the number
of rows per column — reducing `SELECTED_COL` size — at the cost of a larger `U_alpha` in
eval mode. With coefficient mode, the codeword size can be enlarged freely because U_alpha
stays at T (not N), so `SELECTED_COL` shrinks without penalty.

### Impact (Vortex-4, T=8192, RS=16)

| Mode | U_alpha cells | Bytes |
|---|---|---|
| Eval (before) | N = 131,072 ext | 131,072 × 16 = **2,097,152** |
| Coeff (after) | T = 8,192 ext | 8,192 × 16 = **131,072** |
| **Saving** | 122,880 ext | **1,966,080 bytes (~1.9 MB)** |

---

## Optimization 3: Skip Precomputed Merkle Proof

**Option:** `SkipPrecomputedMerkleProof()`

### Why the precomputed Merkle proof is redundant

For **committed** columns, the Merkle proof is necessary: the verifier cannot know
`selectedCol[j]` without it, so it must verify the column is genuinely part of the
committed codeword:

| Y_i source | Can prover fake selectedCol? | Merkle needed? |
|---|---|---|
| Committed (prover claim) | Yes — by adjusting Y_i to match | **Yes** |
| Precomputed (verifier computes) | No — Y_precomp is a fixed known value | **No** |

For **precomputed** columns, `ExplicitPolynomialEval` (in `verifier.go`) runs at the last
round and directly evaluates the known precomputed polynomials at the challenge point x,
pinning `Y_precomp` to a fixed verifier-computed value. The Schwartz-Zippel check then
enforces that `selectedCol_precomp` is consistent with `Y_precomp`. Since `Y_precomp` is
fixed independently of the prover, the Merkle path adds no additional security.

This redundancy was present from the beginning — `ExplicitPolynomialEval` has always
run unconditionally (it is not guarded by `IsSelfrecursed`).

### MerkleProofSize formula

```
MerkleProofSize = NextPowerOfTwo(K × numRounds × depth) × 8
```

where `depth = log2(codewordSize)`, K = NbColsToOpen, numRounds = committed rounds
(precomputed excluded when `SkipPrecomputedMerkleProof` is set).

### Why the saving is binary (power-of-two boundary)

The saving only materialises when removing the precomputed round crosses a `NextPowerOfTwo`
boundary downward. With the current parameters (K=64, 7 committed rounds, depth=17):

| Rounds | K × rounds × depth | NextPow2 | Cells |
|---|---|---|---|
| 8 (with precomp) | 64 × 8 × 17 = 8,704 | **16,384** | 16,384 × 8 = **131,072** |
| 7 (skip precomp) | 64 × 7 × 17 = 7,616 | **8,192** | 8,192 × 8 = **65,536** |
| **Saving** | | | **65,536 cells = 262,144 bytes** |

At depth=16 (T=4096), both cases map to NextPow2=8192, so there is no saving.
The optimization is effective here because `WithTargetColSize(1<<13)` gives T=8192 →
depth=17.

### Propagation into self-recursion

When `SkipPrecomputedMerkleProof` is set on intermediate Vortex rounds (not just the
final one), the self-recursion step (`linearhashandmerkle.go`) must also skip the
precomputed round in its own Merkle tree checks. The same argument applies: the
self-recursion's Merkle check for precomputed is equally redundant because
`ExplicitPolynealEval` runs for **all** Vortex rounds (it is not guarded by
`IsSelfrecursed`).

---

## Vortex-4 Final Proof Breakdown (setup8.log)

Compilation parameters from `fullInitialCompilationSuite` final Vortex (`full.go:141`):

```
vortex.Compile(16, false,
    ForceNumOpenedColumns(64),         // K = 64
    WithOptionalSISHashingThreshold(1<<20),
    PremarkAsSelfRecursed(),
    WithUAlphaCoefficients(),          // opt 2
    SkipSelfRecursionProofColumns(),   // opt 1
    SkipPrecomputedMerkleProof(),      // opt 3
)
```

Setup log (2026-03-13):

```
processed the precomputed columns  nbPrecomputedRows=37  isSISAppliedForCommitment=false
Compiled Vortex round  round=26  numComs=68   polynomialSize=8192  codewordSize=131072  columnHashingMode=Poseidon2
Compiled Vortex round  round=27  numComs=374  polynomialSize=8192  codewordSize=131072  columnHashingMode=Poseidon2
Compiled Vortex round  round=28  numComs=272  polynomialSize=8192  codewordSize=131072  columnHashingMode=Poseidon2
Compiled Vortex round  round=29  numComs=24   polynomialSize=8192  codewordSize=131072  columnHashingMode=Poseidon2
Compiled Vortex round  round=30  numComs=36   polynomialSize=8192  codewordSize=131072  columnHashingMode=Poseidon2
Compiled Vortex round  round=31  numComs=28   polynomialSize=8192  codewordSize=131072  columnHashingMode=Poseidon2
Compiled Vortex round  round=33  numComs=4    polynomialSize=8192  codewordSize=131072  columnHashingMode=Poseidon2
```

Parameters: T=8192, N=131072, RS=16, depth=17, K=64, 7 committed rounds, precomp=37 rows (non-SIS, skipped).
Total numComs: 37 + (68+374+272+24+36+28+4) = 37 + 806 = **843** → NextPow2 = **1024**.

| Component | Cells | Element type | Bytes |
|---|---|---|---|
| U_alpha (coeff mode) | 8,192 | ext (16 B each) | **131,072** |
| SELECTED_COL | 65,536 | base (4 B each) | **262,144** |
| MERKLEPROOF | 65,536 | base (4 B each) | **262,144** |
| **Total** | **139,264** | | **655,360 (~640 KB)** |

> **Note:** U_alpha cells are `fext.Element` (degree-4 KoalaBear extension, 16 bytes each).
> SELECTED_COL and MERKLEPROOF cells are base KoalaBear elements (4 bytes each).
> The wizard `NumCellsProof` metric mixes both — actual byte size requires per-component accounting.

---

## Benchmark Result

`BenchmarkProfileSelfRecursion` (realistic-segment, T3=4096, T4=8192, after all optimizations):

```
BenchmarkProfileSelfRecursion/realistic-segment/T3=4096/T4=8192-192
    1   1015610395 ns/op   7127040 #committed-cells   177864 #proof-cells
```

**177,864 proof cells** for the full pre-recursion chain.

---

## Optimization 4: GKR Poseidon2

TODO

---

## Optimization 5: WHIR

TODO
