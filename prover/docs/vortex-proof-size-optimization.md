# Vortex Proof Size Optimization

This document describes the two optimizations applied to reduce the final Vortex proof size
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

## Optimization 1: Reduce U_alpha by the Blowup Factor


### Evaluation form vs. coefficient form

In standard Vortex, `U_alpha` — the random linear combination of all committed polynomials —
is sent as a Reed-Solomon codeword of N evaluations over the extension field:

```
Eval mode:   N = T × blowup_factor   extension-field elements
Coeff mode:  T                       extension-field elements
```

This work switches to coefficient form: the prover sends T polynomial
coefficients instead of N codeword evaluations.

### Interaction with column size

With a fixed committed-cell budget, increasing `T` (larger polynomial degree) reduces
`numRows`, which shrinks `SELECTED_COL`. In evaluation mode this comes at the cost of a
larger `U_alpha` (N = T × blowup grows). In coefficient mode `U_alpha` stays at T regardless
of blowup, so `T` can be increased 'freely' without any U_alpha penalty — `SELECTED_COL`
shrinks while `U_alpha` stays relatively small.

### Impact (Vortex-4, T=16384, blowup=16)

| Mode | U_alpha size | Bytes |
|---|---|---|
| Eval (before) | N = 262,144 ext elements | 262,144 × 16 = **4,194,304** |
| Coeff (after) | T = 16,384 ext elements | 16,384 × 16 = **262,144** |
| **Saving** | 245,760 ext elements | **~3.8 MB** |

Beyond the proof size benefit (U_alpha is 16× smaller), coefficient mode is also cheaper to
compute: the prover builds U_alpha directly from T-length coefficient vectors instead of
N-length encoded rows. The verifier circuit's Horner evaluation is ~16× cheaper in gnark
constraints than Lagrange interpolation in evaluation mode.

---

### Vortex Verification

This section describes how the Vortex verifier checks correctness in both evaluation mode and
coefficient mode, to explain the mode switch and its cost implications.

The verifier checks four conditions:

#### 1. ReedSolomon Check

Verify/compute that $U_{\alpha,\text{Eval}}$ is a codeword of $U_{\alpha,\text{Coeff}}$.

- **[evaluation mode only]** Compute iFFT to obtain $U_{\alpha,\text{Coeff}}$; verify $U_{\alpha,\text{Coeff}}[j] = 0$ for $j \geq T$.

- **[coefficient mode only]** Compute FFT to obtain $U_{\alpha,\text{Eval}}$.

- **[gnark circuit and self-recursion only]** Check:
$$\text{LagrangeEval}(U_{\alpha,\text{Eval}}, \text{challenge}) = \text{CanonicalEval}(U_{\alpha,\text{Coeff}}, \text{challenge})$$
This check applies only in the gnark circuit. The native verifier runs the FFT itself as a
deterministic computation on trusted input — there is nothing to prove. But the gnark circuit
computes the FFT using a hint, so it must prove the hint is correct via Schwartz-Zippel.

> **Schwartz-Zippel appears in both modes symmetrically**
>
> | Mode | Direction | What the gnark hint verifies |
> |---|---|---|
> | Eval mode  | N evals → iFFT → T coeffs | hint = coefficients |
> | Coeff mode | T coeffs → FFT → N evals | hint = evaluations |
>
> Both use the same Schwartz-Zippel structure. It is the same check, just for opposite FFT directions.



#### 2. Statement Check

Compute $U_\alpha(x)$. Note that $U_{\alpha,\text{Coeff}}$ and $U_{\alpha,\text{Eval}}$ represent
the same polynomial:

$$U_\alpha(x) = \text{LagrangeEval}(U_{\alpha,\text{Eval}}, x) = \text{CanonicalEval}(U_{\alpha,\text{Coeff}}, x)$$

> $\text{CanonicalEval}(U_{\alpha,\text{Coeff}}, X)$ is a degree-(T-1) polynomial, which is **cheaper**
> to evaluate than the degree-(N - 1)  polynomial $\text{LagrangeEval}(U_{\alpha,\text{Eval}}, X)$.

Verify the evaluation:

$$U_\alpha(x) = \text{CanonicalEval}(y, \alpha)$$

where $\text{CanonicalEval}(y, \alpha) = y_0 + \alpha \cdot y_1 + \alpha^2 \cdot y_2 + \ldots$


#### 3. Linear Combination Check

Look up $U_{\alpha,\text{Eval}}$ at $Q$. Verify the linear combination: for $q_i \in Q$, $i \leq K$ opened columns:

$$\text{CanonicalEval}(\mathbf{col}_i, \alpha) = U_{\alpha,\text{Eval}}[q_i]$$



#### 4. Merkle Proof Verification

Ensure the commitment is consistent with the provided column:

$$H(\mathbf{col}_i) = h_{q_i}$$

Check that column $\mathbf{col}_i$ is consistent with $(\text{merkleproof}[i], \text{root}[i])$.
The leaves are:

- $\text{Poseidon2}(\text{sis}(\mathbf{col}_i))$ using SIS hash, or
- $\text{Poseidon2}(\mathbf{col}_i)$ using non-SIS hash

**Gnark circuit:** only checks the non-SIS case.



---

#### How Self-Recursion Verifies the Above Checks

**RowLinearCombinationPhase**
1. ReedSolomon Check: `reedsolomon.CheckReedSolomon`
2. Statement Check: `ctx.consistencyBetweenYsAndUalpha()`

**ColumnOpeningPhase**
3. Linear Combination Check: `ColSelection` performs the lookup and `CollapsingPhase` verifies the evaluation equations: the K individual equations are collapsed into one by sampling a collapse coin

**LinearHashAndMerkle**
4. Merkle proof verification: `LinearHashAndMerkle`

---

## Optimization 2: Automatic Proof Column Deduplication

### What was duplicated

The Vortex prover previously registered three opened-column proof objects:

- `SELECTED_COL` — all rounds combined (SIS + non-SIS), consumed by the Schwartz-Zippel verifier
- `SELECTED_COL_SIS` — SIS rounds only
- `SELECTED_COL_NON_SIS` — non-SIS rounds only

The split was needed by self-recursion: SIS openings are hashed via lattice-SIS and non-SIS
openings via Poseidon2, and these are verified independently. The concatenated
`SELECTED_COL` was also registered for the overall verifier:

```
SELECTED_COL = concat(SELECTED_COL_NON_SIS, SELECTED_COL_SIS)
```

### Current mechanism

In practice every Vortex context is either pure-SIS or pure-non-SIS (never mixed). The
compiler now detects this:

```go
ctx.IsPureSIS    = numRowsSIS != 0 && numRowsNonSIS == 0
ctx.IsPureNonSIS = numRowsSIS == 0 && numRowsNonSIS != 0
```

In the pure case the split column is reused directly as `SELECTED_COL` — there is no separate
combined column and no duplication, regardless of whether self-recursion follows.

### Impact

| Proof column | Before (explicit option) | After (automatic) |
|---|---|---|
| `SELECTED_COL_NON_SIS` | cols=64, cells=65,536 | cols=64, cells=65,536 |
| `SELECTED_COL` | cols=64, cells=65,536 | **removed** (reused as `SELECTED_COL_NON_SIS`) |
| **Total opened-column cells** | **131,072** | **65,536 (−50%)** |

---

## Benchmark Results

`BenchmarkProfileSelfRecursion/realistic-segment` — T4=16384, `-benchtime=1x`.
Cell counts use the base-field unit (4 bytes); extension-field elements in `U_alpha` are
weighted ×4 when computing totals.

Baseline uses T3=32768 (evaluation mode, no optimizations).
After applies both optimizations with T3=4096 (coefficient mode enables the smaller T3).

### Cumulative impact per optimization

| Step | Optimizations active | T3 | Committed cells | Δ committed | Proof cells | Δ proof cells | Proof size |
|---|---|---:|---:|---:|---:|---:|---:|
| 0 — baseline | none (eval mode) | 32,768 | 27,901,952 | — | 1,443,528 | — | ~5.5 MB |
| 1 | + U_alpha coefficient mode | 4,096 | 17,235,968 | −38.2% | 329,416 | −1,114,112 | ~1.3 MB |
| 2 | + Proof column deduplication | 4,096 | 17,235,968 | — | 263,880 | −65,536 | ~1.0 MB |
| | **Total** | | **−38.2%** | | **−1,179,648 (−81.7%)** | | |

### Proof breakdown after both optimizations (T3=4096, T4=16384)

Vortex params: numComs=7, depth=18, numOpening=64, MerkleProofSize=16384, numRows=1015 (all non-SIS).

| Component | Cols | Cells |
|---|---:|---:|
| LINEAR_COMBINATION (U_alpha) | 1 | 65,536 |
| MERKLEPROOF | 8 | 131,072 |
| SELECTED_COL_NON_SIS | 64 | 65,536 |
| MERKLEROOT | 200 | 200 |
| OTHER (self-recursion coins + duals) | 6 | 1,536 |
| **TOTAL** | | **263,880** |

---

## Vortex-4 Final Proof Breakdown in the ZK-EVM full prover (setup.log)

This section compares the above benchmark result with the final Vortex in `fullInitialCompilationSuite` (`full.go:141`):



Setup log (2026-03-20):

```
processed the precomputed columns nbPrecomputedRows=27 nbShadowRows=0 where isSISAppliedForCommitment=false
Compiled Vortex round  round=26  numComs=20   numUnconstrained=0   polynomialSize=16384  codewordSize=262144  columnHashingMode=Poseidon2  merkleHashingField=Koalabear
Compiled Vortex round  round=27  numComs=668  numUnconstrained=64  polynomialSize=16384  codewordSize=262144  columnHashingMode=Poseidon2  merkleHashingField=Koalabear
Compiled Vortex round  round=28  numComs=232  numUnconstrained=0   polynomialSize=16384  codewordSize=262144  columnHashingMode=Poseidon2  merkleHashingField=Koalabear
Compiled Vortex round  round=29  numComs=16   numUnconstrained=0   polynomialSize=16384  codewordSize=262144  columnHashingMode=Poseidon2  merkleHashingField=Koalabear
Compiled Vortex round  round=30  numComs=20   numUnconstrained=0   polynomialSize=16384  codewordSize=262144  columnHashingMode=Poseidon2  merkleHashingField=Koalabear
Compiled Vortex round  round=31  numComs=28   numUnconstrained=0   polynomialSize=16384  codewordSize=262144  columnHashingMode=Poseidon2  merkleHashingField=Koalabear
Compiled Vortex round  round=33  numComs=4    numUnconstrained=0   polynomialSize=16384  codewordSize=262144  columnHashingMode=Poseidon2  merkleHashingField=Koalabear
[wizard.analytic] msg=pre-recursion.post-vortex-4  NumColumnsCommitted=0  NumColumnsProof=1805  NumColumnsPrecomputed=0
```

Parameters: T=16384, N=262,144, blowup=16, depth=18, K=64, 7 committed rounds, precomp=27 rows
(non-SIS, `isSISAppliedForCommitment=false`). Precomputed round counted as 1 in the Merkle proof.
Total polynomials: 988 committed + 27 precomp = **1,015** → NextPow2 = **1,024**.

MerkleProofSize: depth × (numCommitted + numPrecomp) × K = 18 × (7+1) × 64 = 9,216
→ NextPow2 = 16,384 nodes × 8 elem/node (Poseidon2/KoalaBear) = **131,072** base cells.

| Component | Cells | Element type | Bytes |
|---|---:|---|---:|
| U_alpha (coeff mode) | 16,384 | ext (16 B each) | **262,144** |
| SELECTED_COL | 65,536 | base (4 B each) | **262,144** |
| MERKLEPROOF | 131,072 | base (4 B each) | **524,288** |
| **Total** | | | **1,048,576 (~1 MB)** |

This is consistent with the benchmark result
