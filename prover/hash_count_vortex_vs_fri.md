# Hash Count: Vortex × 4 + Recursion  vs  FRI + Recursion

Counting Poseidon2 (P2) compressions in the verifier, including the inherent
Fiat-Shamir transcript, for two proof-system architectures over the same input.

## Setup

| Symbol | Meaning | Value |
|--------|---------|-------|
| λ | security level | 128 bits |
| m | number of committed polynomials | **256** (= 2²⁷ cells / 2¹⁹ col-size) |
| n | polynomial size (evaluations) | 2¹⁹ |
| B | blowup factor (rate inverse) | varies |
| p, s | spot-check / query count | = 2λ / log₂ B  (Guruswami-Sudan) |
| d | Merkle tree depth | log₂(n × B) |
| C | Vortex commitment rounds (all layers) | **8** (7 committed + 1 precomputed) |
| P2 | one Poseidon2 compression (8 elements → 8 elements) | atomic unit |

**SIS params** (LogTwoDegree=6, LogTwoBound=16):
- SIS output = 2⁶ = **64 elements** → compressed to leaf in 64/8 = **8 P2 calls**
- Direct Poseidon2 leaf (no SIS): ⌈m/8⌉ P2 calls

---

## Pipeline A — Current: 4 × Vortex + SelfRecurse + Recursion

```
Initial wizard (2²⁷ cells)
  ↓  Vortex(B=2,  p=256) + SelfRecurse
  ↓  Vortex(B=8,  p=86)  + SelfRecurse
  ↓  Vortex(B=16, p=64)  + SelfRecurse
  ↓  Vortex(B=16, p=64)
  ↓  Recursion (gnark / BN254 PLONK)
```

### Vortex single-layer verifier cost

For a layer with blowup B, spot checks p, commitment rounds C, polynomial size n:

| Component | P2 formula | Notes |
|-----------|-----------|-------|
| SIS leaf hashing | `p × C × 8` | SIS always applied (threshold = 0 for V1–V3) |
| Merkle path verification | `p × C × log₂(nB)` | one path per spot-check per round |
| **FS — absorb UAlpha for Q coin** | **`nB / 8`** | fixed per codeword size; *independent of m and C* |
| FS — absorb Merkle roots for α coin | `C` | negligible |

The **UAlpha** polynomial (the linear combination of all committed polynomials evaluated at
all `nB` codeword positions) is sent by the prover before the column-selection challenge Q
is drawn.  The verifier must absorb every element of UAlpha into the Fiat-Shamir sponge to
derive Q, costing `nB/8` P2 calls regardless of security parameters.

### Concrete numbers per layer (C = 8 for all layers)

V1–V3 use SIS (threshold = 0), so leaf hash = 8 P2 regardless of m.
V4 uses non-SIS (threshold = 2²⁰).  After three rounds of self-recursion the committed
row count m₄ converges to ≈ 38 (dominated by the proof structure, not the original input
size), giving NextPow2(38) = 64 → leaf hash = 64/8 = **8 P2** — same as SIS by coincidence.

| Layer | B | p | numEnc | depth d | m (rows) | leaf P2 | Leaf total | Merkle total | FS (UAlpha) | **Total** |
|-------|---|---|--------|---------|-----------|---------|------------|--------------|-------------|-----------|
| V1 → SR1 | 2  | 256 | 2²⁰ | 20 | 256 | 8 (SIS) | 256×8×8 = **16,384** | 256×8×20 = **40,960** | **131,072** | **188,416** |
| V2 → SR2 | 8  | 86  | 2²⁰ | 20 | ~12 | 8 (SIS) | 86×8×8 = **5,504** | 86×8×20 = **13,760** | **131,072** | **150,336** |
| V3 → SR3 | 16 | 64  | 2¹⁹ | 19 | ~37 | 8 (SIS) | 64×8×8 = **4,096** | 64×8×19 = **9,728** | **65,536** | **79,360** |
| V4 → gnark | 16 | 64 | 2¹⁸ | 18 | ~38 | 8 (non-SIS) | 64×8×8 = **4,096** | 64×8×18 = **9,216** | **32,768** | **46,080** |

> V1 and V2 share the same UAlpha size (numEnc = 2²⁰) despite different B, because
> the smaller polynomial size in V2 (n = 2¹⁷) is exactly offset by the larger blowup (B = 8).

> m at each level is estimated from the Vortex proof cell sizes propagating through
> the pipeline.  V1 proof is dominated by UAlpha (1M elements); SR verification columns
> add ~330K cells; after Arcane(targetColSize=2¹⁷) this gives m₂ ≈ 1.56M/131K ≈ 12.
> V3 and V4 follow similarly.  The exact value is unimportant for V1–V3 (SIS, leaf=8 P2)
> and coincidentally unimportant for V4 (NextPow2(38)=64, also 8 P2).

### Pipeline A totals (C = 8)

| Source | P2 |
|--------|----|
| FS (UAlpha) — V1+V2 | 2 × 131,072 = **262,144** |
| FS (UAlpha) — V3 | **65,536** |
| FS (UAlpha) — V4 | **32,768** |
| Leaf + Merkle — V1 | 16,384 + 40,960 = **57,344** |
| Leaf + Merkle — V2 | 5,504 + 13,760 = **19,264** |
| Leaf + Merkle — V3 | 4,096 + 9,728 = **13,824** |
| Leaf + Merkle — V4 | 4,096 + 9,216 = **13,312** |
| **Grand total** | **464,192** |

FS (UAlpha) = 360,448 P2 = **78%** of total.
Spot checks = 103,744 P2 = **22%** of total.

---

## Pipeline B — Alternative: FRI + Recursion

```
Initial wizard (2²⁷ cells)
  ↓  FRI on trace polynomial of degree d = 2²⁷
  ↓  Recursion (gnark / BN254 PLONK)
```

FRI has **no UAlpha**.  Each folding challenge is derived from a single Merkle root
(8 elements ≡ 1 P2).  Total FS ≈ n P2 for n folding rounds.

### Model

All 2²⁷ trace cells are treated as evaluations of a single polynomial of degree
**d = 2²⁷** over the native KoalaBear field (≈ 2³¹).  FRI proves proximity to this
low-degree polynomial on an evaluation domain of size d/ρ.

| Symbol | Value |
|--------|-------|
| d | 2²⁷ (trace polynomial degree) |
| field | KoalaBear ≈ 2³¹ |
| k | extension degree (see below) |
| λ | 128 |

### FRI round count

n = log₂(d/ρ) = 27 + log₂(1/ρ)

| ρ | n |
|---|---|
| 1/2  | 28 |
| 1/4  | 29 |
| 1/8  | 30 |
| 1/16 | 31 |

### Algebraic soundness

With k=4 extension, |F_{p⁴}| ≈ 2¹²⁴.  Total algebraic error over n FRI rounds:

ε_alg = n · d / |F_{p^k}| = n · 2²⁷ / 2¹²⁴

| ρ | n | λ_alg = −log₂(ε_alg) |
|---|---|----------------------|
| 1/2  | 28 | 124 − 27 − log₂(28) ≈ **92.1 bits** |
| 1/4  | 29 | **92.0 bits** |
| 1/8  | 30 | **91.9 bits** |
| 1/16 | 31 | **91.8 bits** |

Grinding requirement to reach 128 bits: **b = 128 − 92 = 36 bits** → 2³⁶ ≈ 68 billion
hashes.  Prohibitively expensive.

**With k=5 extension**, |F_{p⁵}| ≈ 2¹⁵⁵:
λ_alg ≈ 155 − 27 − log₂(n) ≈ 155 − 27 − 5 = **123 bits** → grinding b = 5 bits.
**k=5 is the natural choice for d = 2²⁷ on KoalaBear.**

### Query count (DEEP-FRI conjecture)

t = ⌈128 / log₂(1/ρ)⌉

| ρ | bits/query | t |
|---|-----------|---|
| 1/2  | 1 | 128 |
| 1/4  | 2 | 64 |
| 1/8  | 3 | 43 |
| 1/16 | 4 | 32 |

### Hashes per query

Each query opens 2 positions at each of n FRI rounds.  Round i has a Merkle tree
of depth (n − i), so each opening costs (n − i) path hashes.

Hashes/query = Σᵢ₌₀ⁿ⁻¹ 2(n − i) = n(n + 1)

(Leaf hash ≈ 1 P2 per opening since each leaf = 1 field element; negligible.)

| ρ | n | n(n+1) |
|---|---|--------|
| 1/2  | 28 | 812 |
| 1/4  | 29 | 870 |
| 1/8  | 30 | 930 |
| 1/16 | 31 | 992 |

### FRI total hashes across all rates

| ρ | t | n(n+1) | **t × n(n+1)** | FS (n P2) | **Grand total** |
|---|---|--------|----------------|-----------|-----------------|
| 1/2  | 128 | 812 | 103,936 | 28 | **103,964** |
| 1/4  |  64 | 870 |  55,680 | 29 |  **55,709** |
| 1/8  |  43 | 930 |  39,990 | 30 |  **40,020** |
| 1/16 |  32 | 992 |  31,744 | 31 |  **31,775** |

Total hashes decrease monotonically as ρ decreases (blowup increases), because the
reduction in query count outweighs the slight growth in path depth.

---

## Side-by-side comparison

### Segmented input: d = 2²⁷ cells

Using the unified d = 2²⁷ polynomial model for FRI (see Pipeline B section):

| Scheme | ρ | Total P2 | vs Pipeline A |
|--------|---|----------|--------------|
| **Pipeline A** — Vortex × 4 (C=8) | 2/8/16/16 | **464,192** | — |
| FRI | 1/2  | **103,964** | 4.5× cheaper |
| FRI | 1/4  |  **55,709** | 8.3× cheaper |
| FRI | 1/8  |  **40,020** | 11.6× cheaper |
| FRI | 1/16 |  **31,775** | **14.6× cheaper** |

FRI at ρ=1/16 is **14.6× cheaper** than the current pipeline.
