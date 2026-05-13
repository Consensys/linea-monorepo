# Packed vs Unpacked Bootstrappers

This document compares the two packing approaches
explains why packed is slower on mixed-size workloads, and identifies the
regime where packed is the better choice.

---

## 1. The Unpacked way

`InsertBootstrapperOpenings` groups committed columns by `numVars` and emits
one `MultilinearEval` query per distinct size. Each query is compiled into its
own Vortex commit on a matrix whose rows are the K polynomials of that size
stacked vertically (each poly viewed as a `2^nRow × 2^nCol` matrix with
`nRow + nCol = nv`).

For `mixedScaledDist` (24 cols, 7 distinct nv):

```
group     K   nRow nCol     data matrix (K·2^nRow × 2^nCol)
─────────────────────────────────────────────────────────────
nv=14     4    7    7         512  ×   128         (65,536 cells)
nv=16     4    8    8        1024  ×   256        (262,144 cells)
nv=17     8    9    8        4096  ×   256      (1,048,576 cells)
nv=18     4    9    9        2048  ×   512      (1,048,576 cells)
nv=19     2   10    9        2048  ×   512      (1,048,576 cells)
nv=20     1   10   10        1024  ×  1024      (1,048,576 cells)
nv=21     1   11   10        2048  ×  1024      (2,097,152 cells)
─────────────────────────────────────────────────────────────
                          Σ data cells = 6,619,136   (no padding)
                          Σ commits    = 7  (running concurrently
                                             via mlOrigCommitBatch)
```

Each group gets its own goroutine in `mlOrigCommitBatch`; all 7 commits
RS-encode and SIS-hash in parallel. Downstream rounds emit K UAlpha + K
RowEvals proof columns per group, grouped by size into ~5 CommitMLColumns
commits per recursion round.

---

## 2. The Packed way

`InsertBootstrapperOpeningsPacked` runs two layers of packing.

**Step A — buddy-allocate every poly into one `Q` of size `2^N`.**
$N = \lceil \log_2(\sum_i 2^{nv_i}) \rceil$. Each poly $P_k$ is placed at an
offset aligned to its size; the integer `b_k = offset_k / 2^{nv_k}` is the
**locator** — the path in a binary buddy tree to that slot. Original claims
transform via

$$
P_k(z_k) \;=\; Q\bigl(b_{k,0},\ldots,b_{k,N-nv_k-1},\; z_0,\ldots,z_{nv_k-1}\bigr).
$$

**Step B — commit `Q` as a single Vortex matrix.** With
`nRow = ⌈N/2⌉, nCol = N - nRow`, `Q` is reshaped into a `2^nRow × 2^nCol`
matrix and RS-encoded + SIS-hashed once.

For `mixedScaledDist`:

```
N = ⌈log₂(6,619,136)⌉ = 23
|Q|                   = 2²³ = 8,388,608 cells   →  1,769,472 zeros (21% padding)

data matrix           : 4,096 × 2,048
codeword (×2 blowup)  : 4,096 × 4,096

         row range     content                        rows  cells
        ────────────  ────────────────────────────   ─────  ──────────
            0 – 1023   P[nv=21]                       1024   2,097,152
         1024 – 1535   P[nv=20]                        512   1,048,576
         1536 – 2047   P[nv=19] ×2                     512   1,048,576
         2048 – 2559   P[nv=18] ×4                     512   1,048,576
         2560 – 3071   P[nv=17] ×8                     512   1,048,576
         3072 – 3199   P[nv=16] ×4                     128     262,144
         3200 – 3231   P[nv=14] ×4                      32      65,536
         3232 – 4095   PADDING                         864   1,769,472   ← waste
```

Only **one** Vortex commit happens at round 0. Downstream rounds use
the SharedRowEvals optimization (1 UAlpha + 1 RowEvals proof column per round,
regardless of K), so each recursion round emits only 2 commits instead of ~5.

Commit counts across the full pipeline (round 0 + 5 recursion rounds):

| | Unpacked | Packed |
|---|---|---|
| Round 0 (orig) | 7 | **1** |
| Each recursion round (×5) | ~5 | **2** |
| **Total** | **32** | **11** |

---

## 3. Benchmark on a mixed-size distribution

For a mixed-size column distribution, packing all polys into a single poly
(which builds a single matrix) has **padding overhead** — it pays more cells
in total for computation. On `mixedScaledDist`:

```
                           Unpacked        Packed
data cells (rd-0 commit)   6,619,136       8,388,608  (+27% padding)
codeword cells             13,238,272      16,777,216 (+27%)
```

Measured wall + CPU (3-iter avg, 192 cores):

| | Unpacked | Packed | Δ |
|---|---|---|---|
| SIS (`RSis.InnerHash`) cum | 10.82 s | 11.44 s | +5.7% |
| FFT (`Domain.FFT`) cum | 9.49 s | 9.69 s | +2.1% |
| **Total CPU** | **13.63 s** | **14.68 s** | **+8%** |
| Effective cores | 16.6 | 16.1 | — |
| **Wall (3 iters)** | **0.82 s** | **0.91 s** | **+11%** |

Packed loses by ≈ 11 % on this distribution.

---

## 4. Why packed is slower

Vortex commit cost has two components, and **both scale with cells** — packed's
padding inflates each one:

```
SIS  cost  ∝  codeword cells
FFT  cost  ∝  codeword cells × log₂(nColSize_codeword)
```

Plugging in:

```
                              Unpacked (weighted)        Packed
codeword cells                13.24 M                    16.78 M     (+27%)
log₂(nColSize_codeword)       ~10  (mixed: 8/9/10/11)     12         (+20%)

→  SIS  cost ratio  P/U  ≈  1.27
→  FFT  cost ratio  P/U  ≈  1.27 × 1.20 ≈ 1.52
```

Two compounding effects:
1. **Padding** — Q has +27% cells that aren't real data but still get
   RS-encoded and SIS-hashed.
2. **`log₂(nColSize)` grows** — Q's `nColSize = 2048` is bigger than any
   unpacked group's, so per-row FFT cost rises with the log factor.

### Why the prediction is an upper bound

The +27 % SIS and +52 % FFT ratios apply **per Vortex commit**, but the
prediction only describes **round 0** — the original commit on the data.
The full pipeline runs **six rounds** (round 0 + 5 recursion rounds), and the
two sides do very different amounts of work in the recursion rounds.

```
                      Unpacked share of CPU       Packed share of CPU
Round 0 (orig)        ~70%   (7 parallel          ~95%   (1 big commit
                              commits sum to              on Q including
                              6.62M cells)                21% padding)
Rounds 2–6 (recursion)  ~30%   (K UAlpha + K        ~5%    (1 UAlpha + 1
                              RowEvals per group,         RowEvals total,
                              5 CommitMLColumns           2 commits per
                              commits per round)          round via
                                                          SharedRowEvals)
```

So the cost equation isn't `total_packed = total_unpacked × 1.27`; it's

```
total_packed   = (round-0 packed)    + (recursion packed)
               ≈ (1.4 × round-0 unpacked) + (0.15 × recursion unpacked)

≈ 1.4 × 0.70 + 0.15 × 0.30        (normalised to unpacked total = 1.0)
≈ 0.98 + 0.045
≈ 1.025                            (predicts ≈ +2.5 % vs observed +8 %)
```

The round-0 penalty is mostly cancelled by the recursion-round savings —
**packed redistributes work**: bigger round-0, almost-free downstream rounds.

This is also why doubling the padding (e.g. moving from a 21 %-padded
distribution to one with 5 % padding) makes packed flip from "slight loss"
to "clear win" — the round-0 penalty drops below the recursion savings.

The round-0 penalty is unavoidable as long as the data has padding, but
since round 0 is no longer the only round that matters, the **effective**
penalty on total CPU is much smaller than the per-cell prediction.

---

## 5. When packed wins

**If the padding overhead is small, the packed way is cheaper end-to-end.**
For mixed distributions where padding is significant, packed remains the right
choice whenever the recursion verifier dominates (BN254 wrap, deep aggregation
trees) — the native-prover penalty is small and the recursion savings are
structural.


---

## 6. How to Test

```

# Native prover bench:
go test -bench BenchmarkBootstrapperOpeningsPacked -run '^$' \
  -benchtime=10x -timeout=15m ./protocol/compiler/multilinvortex/
```
