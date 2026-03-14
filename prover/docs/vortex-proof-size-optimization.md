# Vortex Proof Size Optimization

This document describes the four optimizations applied to reduce the final Vortex proof size
produced by `fullInitialCompilationSuite` (`zkevm/full.go`).

---

## Background

### Proof components

The three main components of the final Vortex proof are:

| Component | Description | Scales with |
|---|---|---|
| `U_alpha` | Random linear combination of all committed polynomials | (codeword = blowup x polynomial degree) |
| `SELECTED_COL` | Revealed column values at K randomly chosen positions | numRows (# committed polynomials) |
| `MERKLEPROOF` | Sibling hashes for Merkle path verification | log₂(codeword) × K × numRounds |

### Committed cells and their effect on proof size

The Vortex commitment matrix has dimensions:

```
committed_cells = numRows × T

  numRows  total number of committed polynomials across all rounds
  T        polynomial degree  (= codewordSize / blowup)
```

Reducing `numRows` shrinks `SELECTED_COL`; reducing `T` shrinks `U_alpha` and `MERKLEPROOF`.
A single large reduction in `committed_cells` can fundamentally reduce both factors.

---

## Optimization 1: GKR Poseidon2

**Compiler step:** `poseidon2.CompileGKRPoseidon2`

### Mechanism

Without GKR, every Poseidon2 hash call is proven by `poseidon2.checkPoseidon2BlockCompressionExpression` — 206 committed cells. Batching N hash calls therefore produces O(N × 206) committed cells.

With GKR, the entire batch is proven via a sumcheck argument. The GKR verifier needs only a
short transcript column (`GKR_Poseidon2_TRANSCRIPT`) regardless of batch size, replacing the
dense intermediate-state columns entirely.

### Effect on proof size

GKR primarily reduces committed cells (`numRows` × fewer committed polynomials). This directly shrinks `SELECTED_COL` and the Vortex matrix, and fewer columns fed into each self-recursion round compounds the savings across all recursion levels.

| | Committed cells | Proof cells |
|---|---|---|
| Before GKR | 16,891,904 | 919,240 |
| After GKR | 7,127,040 | 800,456 |
| **Reduction** | **−57.8%** | **−118,784 (−12.9%)** |

---

## Optimization 2: Reduce U_alpha by the Blowup Factor

**Option:** `WithUAlphaCoefficients()`

### Evaluation form vs. coefficient form

In standard Vortex, `U_alpha` — the random linear combination of all committed polynomials —
is sent as a Reed-Solomon codeword of N evaluations over the extension field:

```
Eval mode:   N = T × blowup_factor   extension-field elements
Coeff mode:  T                       extension-field elements
```

`WithUAlphaCoefficients()` switches to coefficient form: the prover sends T polynomial
coefficients instead of N codeword evaluations. 

### Interaction with column size

With a fixed committed-cell budget, increasing `T` (larger polynomial degree) reduces
`numRows`, which shrinks `SELECTED_COL`. In evaluation mode this comes at the cost of a
larger `U_alpha` (N = T × blowup grows). In coefficient mode `U_alpha` stays at T regardless
of blowup, so `T` can be increased 'freely' without any U_alpha penalty — `SELECTED_COL`
shrinks while `U_alpha` stays relatively small.

### Impact (Vortex-4, T=8192, blowup=16)

| Mode | U_alpha size | Bytes |
|---|---|---|
| Eval (before) | N = 131,072 ext elements | 131,072 × 16 = **2,097,152** |
| Coeff (after) | T = 8,192 ext elements | 8,192 × 16 = **131,072** |
| **Saving** | 122,880 ext elements | **~1.9 MB** |


Except for the proof size benefit, verifier circuit's Horner evaluation is ~16× cheaper in gnark constraints than Lagrange interpolation using the coeff mode. 

---

## Optimization 3: Skip Duplicated Proof Columns

**Option:** `SkipSelfRecursionProofColumns()`

### What is duplicated

The Vortex prover registers three opened-column proof objects:

- `SELECTED_COL` — all rounds combined (SIS + non-SIS), consumed by the Schwartz-Zippel verifier
- `SELECTED_COL_SIS` — SIS rounds only
- `SELECTED_COL_NON_SIS` — non-SIS rounds only

The split is needed by self-recursion: SIS openings are hashed via lattice-SIS and non-SIS
openings via Poseidon2, and these are verified independently. The concatenated
`SELECTED_COL` is also registered for the overall verifier:

```
SELECTED_COL = concat(SELECTED_COL_NON_SIS, SELECTED_COL_SIS)
```

### Why they are dead weight on the final Vortex

On the final Vortex (marked `PremarkAsSelfRecursed()`) there is no subsequent self-recursion
step, so `SELECTED_COL_SIS` and `SELECTED_COL_NON_SIS` are registered as proof columns but
**no verifier ever reads them**. `SkipSelfRecursionProofColumns()` suppresses their
registration entirely, saving one full copy of the split-column data.

### Impact

| Proof column | Before | After |
|---|---|---|
| `SELECTED_COL_NON_SIS` | cols=64, cells=32,768 | **removed** |
| `SELECTED_COL` | cols=64, cells=32,768 | cols=64, cells=32,768 |
| **Total opened-column cells** | **65,536** | **32,768 (−50%)** |

---

## Optimization 4: Skip Precomputed Merkle Proof

**Option:** `SkipPrecomputedMerkleProof()`

### Why the precomputed Merkle proof is redundant

For **committed** columns the Merkle proof is essential: without it the prover could supply
a fabricated `selectedCol[j]` that satisfies the linear-combination check but does not
correspond to the committed codeword.

For **precomputed** columns the situation is different:

| Column type | Can prover fake `selectedCol`? | Merkle proof needed? |
|---|---|---|
| Committed | Yes — by back-solving Y_i | **Yes** |
| Precomputed | No — Y_precomp is verifier-computed | **No** |

`ExplicitPolynomialEval` (in `verifier.go`) runs unconditionally at the last round and
evaluates each precomputed polynomial directly at the challenge point x, pinning `Y_precomp`
to a fixed verifier-computed value. The Schwartz-Zippel check then enforces consistency
with `selectedCol_precomp`. Since `Y_precomp` is independent of the prover, the Merkle path
adds no additional security and can be dropped.

### MerkleProofSize formula

```
MerkleProofSize = NextPowerOfTwo(K × numRounds × depth) × 8
```

where `depth = log₂(codewordSize)`, K = number of opened columns, and `numRounds` counts
only committed rounds (precomputed round excluded when `SkipPrecomputedMerkleProof` is set).

### Why the saving is binary

The saving only materialises when removing the precomputed round crosses a `NextPowerOfTwo`
boundary downward. With the current parameters (K=64, 7 committed rounds, depth=17):

| Rounds | K × rounds × depth | NextPow2 | Cells |
|---|---|---|---|
| 8 (with precomp) | 64 × 8 × 17 = 8,704 | 16,384 | 16,384 × 8 = **131,072** |
| 7 (skip precomp) | 64 × 7 × 17 = 7,616 | 8,192 | 8,192 × 8 = **65,536** |
| **Saving** | | | **65,536 cells (262,144 bytes)** |

At depth=17 (`WithTargetColSize(1<<13)` results codeword=1<<13 x 16, depth=log2(codeword)=17), 7,616 is below 8,192 and 8,704 above, so the optimization halved the merkle proof cells.

---

## Benchmark Results

`BenchmarkProfileSelfRecursion` — realistic-segment, T3=4096, T4=8192, `-benchtime=1x`.
Cell counts use the base-field unit (4 bytes); extension-field elements in `U_alpha` are
weighted ×4 when computing totals.

### Cumulative impact per optimization


| Step | Optimizations active | Committed cells | Δ committed | Proof cells | Δ proof cells | Proof size |  Compile time | Compile mem |
|---|---|---:|---:|---:|---:|---:|---:|---:|
| 0 — baseline | none | 16,891,904 | — | 919,240 | — | ~3.5 MB | 2.45 s | 2.44 GB |
| 1 | + GKR Poseidon2 | 7,127,040 | −57.8% | 800,456 | −118,784 | ~3.1 MB | 0.95 s | 1.49 GB |
| 2 | + WithUAlphaCoefficients (all rounds) | 3,686,400 | −48.3% | 243,400 | −557,056 | ~951 KB | 0.87 s | 1.30 GB |
| 3 | + SkipSelfRecursionProofColumns | 3,686,400 | — | 210,632 | −32,768 | ~823 KB | 0.85 s | 1.30 GB |
| 4 | + SkipPrecomputedMerkleProof | 3,686,400 | — | **145,096** | −65,536 | **~567 KB** | 0.88 s | 1.30 GB |
| | **Total** | **−78.2%** | | **−774,144 (−84.2%)** | | | | |

---

## Vortex-4 Final Proof Breakdown in the ZK-EVM full prover (setup.log)

Compilation parameters for the final Vortex in `fullInitialCompilationSuite` (`full.go:141`):

```go
vortex.Compile(16, false,
    ForceNumOpenedColumns(64),              // K = 64
    WithOptionalSISHashingThreshold(1<<20),
    PremarkAsSelfRecursed(),
    WithUAlphaCoefficients(),              // opt 2
    SkipSelfRecursionProofColumns(),       // opt 3
    SkipPrecomputedMerkleProof(),          // opt 4
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

Parameters: T=8192, N=131,072, blowup=16, depth=17, K=64, 7 committed rounds, precomp=37 rows
(non-SIS, Merkle proof skipped). Total polynomials: 37 + (68+374+272+24+36+28+4) = **843**
→ NextPow2 = **1024**.

| Component | Cells | Element type | Bytes |
|---|---:|---|---:|
| U_alpha (coeff mode) | 8,192 | ext (16 B each) | **131,072** |
| SELECTED_COL | 65,536 | base (4 B each) | **262,144** |
| MERKLEPROOF | 65,536 | base (4 B each) | **262,144** |
| **Total** | | | **655,360 (~640 KB)** |



There's a small discrepancy between the benchmark and the production data.
The benchmark `SELECTED_COL` is half the production values because the synthetic circuit (Fibo/Lookup/Permutation modules) produces 3.6M committed cells (~450 committed polynomials<512) → NextPow2=**512**, while the full ZK-EVM has 7M committed cells (843 committed polynomials>512) → NextPow2=**1024**, doubling both components. 

