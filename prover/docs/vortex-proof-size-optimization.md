# Vortex Proof Size Optimization

This document describes the optimizations applied to reduce the final vortex proof size output by
`fullInitialCompilationSuite`.

---

## Background


The three main proof components in the final Vortex in `fullInitialCompilationSuite` (see `zkevm/full.go`) are:

| Component | Description |
|---|---|
| `U_alpha` | Random linear combination of all committed polynomials |
| `SELECTED_COL` | Opened column values at K random positions |
| `MERKLEPROOF` | Sibling hashes for Merkle path verification |


---

## Optimization 1: GKR Poseidon2

Without GKR: Every Poseidon2 hash call is proven via explicit polynomial columns — one column per intermediate round state, S-box, and round constant. For a batch of N hash calls, this explodes into O(N × permutation_rounds) committed columns.

With GKR: The entire batch of Poseidon2 evaluations is proven via a sumcheck argument. The GKR verifier only requires a short transcript (the sumcheck messages), which becomes one small polynomial column — GKR_Poseidon2_TRANSCRIPT — regardless of how many hashes are batched.


### Why fewer committed cells → smaller proof

committed_cells = numRows × T

numRows = total number of committed polynomials across all rounds
T       = polynomial degree (= codewordSize / blowup)

The Vortex proof components:

- U_alpha and MERKLEPROOF are scale with T 

- SELECTED_COL is scale with numRows

So a smaller committed_cells could result two smaller factors (committed rows and polynomial degree T)

---

## Optimization 2: Reduce U_alpha size by the Blowup Factor

**Option:** `WithUAlphaCoefficients()`

### Evaluation form vs. coefficient form

In standard Vortex, `U_alpha` is the random linear combination of all committed polynomials,
sent as a **Reed-Solomon codeword** of N evaluations over the extension field:

```
Eval mode:   N = T × blowup_factor   extension-field elements
Coeff mode:  T                   extension-field elements
```

`WithUAlphaCoefficients()` switches to coefficient form: the prover sends T coefficients
instead of N evaluations. 


### Interaction with column size

With a fixed size of commited cells, enlarging the codeword (increasing RS factor or targeting a larger `T`) reduces the number
of rows per column — reducing `SELECTED_COL` size — at the cost of a larger `U_alpha` in
eval mode. With coefficient mode, the codeword size can be enlarged 'freely' because U_alpha
stays at T (not T x blowup_factor), so `SELECTED_COL` shrinks without penalty.

### Impact (Vortex-4, T=8192, blowup_factor=16)

| Mode | U_alpha cells | Bytes |
|---|---|---|
| Eval (before) | N = 131,072 ext | 131,072 × 16 = **2,097,152** |
| Coeff (after) | T = 8,192 ext | 8,192 × 16 = **131,072** |
| **Saving** | 122,880 ext | **1,966,080 bytes (~1.9 MB)** |

---

## Optimization 3: Skip Duplicated Proof Columns

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
**no verifier ever reads them**. 

`SkipSelfRecursionProofColumns()` suppresses their registration entirely.



---

## Optimization 4: Skip Precomputed Merkle Proof

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


---
## Benchmarks compare

baseline: 
BenchmarkProfileSelfRecursion/realistic-segment/T3=4096/T4=8192-192                    1        2463224418 ns/op          16891904 #committed-cells            919240 #proof-cells 2616099528 B/op 27536786 allocs/op

Opt1 GKR:

BenchmarkProfileSelfRecursion/realistic-segment/T3=4096/T4=8192
BenchmarkProfileSelfRecursion/realistic-segment/T3=4096/T4=8192-192                    1        1012616333 ns/op           7127040 #committed-cells            800456 #proof-cells 1604045944 B/op 11868091 allocs/op

Opt 2 WithUAlphaCoefficients

BenchmarkProfileSelfRecursion/realistic-segment/T3=4096/T4=8192
BenchmarkProfileSelfRecursion/realistic-segment/T3=4096/T4=8192-192                    8         140331964 ns/op           7127040 #committed-cells            308936 #proof-cells 200118093 B/op   1483023 allocs/op

Opt 3 SkipSelfRecursionProofColumns
BenchmarkProfileSelfRecursion/realistic-segment/T3=4096/T4=8192
BenchmarkProfileSelfRecursion/realistic-segment/T3=4096/T4=8192-192                    7         146418357 ns/op           7127040 #committed-cells            243400 #proof-cells 228708181 B/op   1694843 allocs/op

Opt4 SkipPrecomputedMerkleProof


BenchmarkProfileSelfRecursion/realistic-segment/T3=4096/T4=8192-192                    5         204445118 ns/op           7127040 #committed-cells            177864 #proof-cells 320229758 B/op   2372886 allocs/op


---

## Vortex Final Proof Breakdown (setup8.log)

Compilation parameters from `fullInitialCompilationSuite` final Vortex (`full.go:141`):

```
vortex.Compile(16, false,
    ForceNumOpenedColumns(64),         // K = 64
    WithOptionalSISHashingThreshold(1<<20),
    PremarkAsSelfRecursed(),
    WithUAlphaCoefficients(),          // opt 2
    SkipSelfRecursionProofColumns(),   // opt 3
    SkipPrecomputedMerkleProof(),      // opt 4
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


---

## Step-by-Step Benchmark Results

`BenchmarkOptimizationSteps` (realistic-segment, T3=4096, T4=8192).

All cell counts use the base-field unit (4 bytes): extension-field elements (`U_alpha`) are
weighted ×4 to convert to base-field cells.

### Total proof cells per step

| Step | Active optimizations | Total cells | Δ vs prev | Bytes |
|---|---|---|---|---|
| baseline | — | **800,456** | — | ~3.1 MB |
| opt1 | SkipSelfRecursionProofColumns | **734,920** | −65,536 | ~2.9 MB |
| opt1+2 | + WithUAlphaCoefficients | **243,400** | −491,520 | ~951 KB |
| opt1+2+3 | + SkipPrecompFinalVortex | **177,864** | −65,536 | ~695 KB |
| opt1+2+3+4 | + SkipPrecompAllVortexes | **177,864** | 0 | ~695 KB |

### Per-component breakdown

| Component | baseline | opt1 | opt1+2 | opt1+2+3 | opt1+2+3+4 |
|---|---|---|---|---|---|
| `U_alpha` | 524,288 | 524,288 | 32,768 | 32,768 | 32,768 |
| `SELECTED_COL` | 65,536 | 65,536 | 65,536 | 65,536 | 65,536 |
| `SELECTED_COL_NON_SIS` | 65,536 | — | — | — | — |
| `MERKLEPROOF` | 131,072 | 131,072 | 131,072 | 65,536 | 65,536 |
| `MERKLEROOT` + `OTHER` | ~14,024 | ~14,024 | ~14,024 | ~14,024 | ~14,024 |
| **Total** | **800,456** | **734,920** | **243,400** | **177,864** | **177,864** |

### Notes

- **opt1** removes `SELECTED_COL_NON_SIS` (65,536 cells): the split columns are dead on the
  final Vortex since no subsequent self-recursion reads them.
- **opt2** reduces `U_alpha` from N=131,072 ext evaluations to T=8,192 ext coefficients.
  Each ext element counts as 4 base cells, so the saving is (131,072 − 8,192) × 4 = **491,520
  cells**.
- **opt3** halves `MERKLEPROOF` from 131,072 → 65,536 by removing the precomputed round,
  which crosses a NextPowerOfTwo boundary (8,704 → 7,616 → NPow2: 16,384 → 8,192).
- **opt4** (SkipPrecomputedMerkleProof on all intermediate Vortexes) gives **zero benefit**
  here because the intermediate Vortexes use SIS-hashed precomputed columns
  (`IsSISAppliedToPrecomputed=true`) and their parameters do not cross a NextPow2 boundary.

---

## Full-Pipeline Benchmark Result

`BenchmarkProfileSelfRecursion` (realistic-segment, T3=4096, T4=8192, all optimizations active):

```
BenchmarkProfileSelfRecursion/realistic-segment/T3=4096/T4=8192-192
    1   1015610395 ns/op   7127040 #committed-cells   177864 #proof-cells
```

**177,864 proof cells = ~695 KB** for the full pre-recursion chain. The figure is slightly
larger than the 139,264 cells from the setup log breakdown because it includes Merkle roots,
GKR transcript columns, and self-recursion auxiliary columns (counted in OTHER above).

---

## Optimization 5: WHIR

TODO
