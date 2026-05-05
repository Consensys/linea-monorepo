# Multilinear Vortex: Protocol, Implementation, and Soundness

## Overview

The protocol has three interleaved layers, repeating until all claims reach $n = 1$:

1. **Entry (`InsertBootstrapperOpenings`).** Convert each committed column $P_k$ into a
   `MultilinearEval` claim $\widetilde{P}_k(c) = y_k$ at a Fiat-Shamir point $c$.

2. **Multilinear Vortex.** For each active `MultilinearEval` claim, reshape $P_k$ as a
   matrix, RS-encode and Merkle-commit to bind the prover, draw a batching coin $\alpha$,
   and produce two half-size claims — UCols and RowClaims — that feed into step 3.

3. **Sumcheck (`Batch`).** Batch all active `MultilinearEval` claims (possibly at mixed
   sizes) into one sumcheck via an eq-selector.  Output: one shared challenge point $c'$
   and $K$ residual oracle queries, all at $c'$.

Steps 2–3 repeat until every claim has $n = 1$.  For columns up to $2^{24}$, five rounds
suffice.

---

## 1. Multilinear Polynomials

### 1.1 Definition

A `MultilinearEval` query declares:
- Columns $P_0, \ldots, P_{K-1}$, each of size $2^{n_k}$.
- Evaluation points $c_k \in \mathbb{F}_\text{ext}^{n_k}$, set at runtime via Fiat-Shamir.
- Claimed values $y_k \in \mathbb{F}_\text{ext}$.

The claim is $\widetilde{P}_k(c_k) = y_k$, where the **multilinear extension** $\widetilde{P}_k$
is the unique multilinear polynomial that agrees with $P_k$ on the Boolean hypercube:

$$\widetilde{P}_k(x_0, \ldots, x_{n_k-1})
  = \sum_{b \in \{0,1\}^{n_k}} P_k[b]
    \cdot \prod_{i=0}^{n_k-1} \bigl(b_i x_i + (1-b_i)(1-x_i)\bigr)$$

$b$ is indexed in binary with $b_0$ as the most-significant bit (gnark-crypto convention).

### 1.2 Why Hypercube Domain (No FFT)

A committed column $P_k$ is an array of field elements; interpreting it as a hypercube
evaluation table costs nothing.  Multilinear evaluation is a tensor product of linear
factors, costing $O(2^n)$ — the same as reading the array once.

Univariate evaluation, by contrast, requires an encode/decode pipeline:

$$\text{array} \xrightarrow{\mathrm{IFFT}} \text{coefficients} \xrightarrow{\times\,\text{coset}} \text{shifted coefficients} \xrightarrow{\mathrm{FFT}} \text{codeword}$$

Skipping this is the root cause of the multilinear prover speedup.

---

## 2. Entry Point: Register Multi-size Claims

`InsertBootstrapperOpenings` scans all round-0 committed columns and registers the initial
`MultilinearEval` claims that feed the vortex pipeline.

**Group by size.**  Columns with the same $n = \log_2(\text{size})$ form one group.
Each group gets one shared evaluation point: $n$ independent `FieldExt` coins
$c = (c_0, \ldots, c_{n-1})$ drawn at round 1.

**Emit one claim per group.**  One `MultilinearEval` query is inserted at round 1 covering
all $K$ columns in the group, evaluated at the shared point $c$.  The prover computes
$y_k = \widetilde{P}_k(c)$ by multilinear folding.

**Why group by $n$?**  The vortex step requires all polynomials in one batch to share the
same number of variables.  Grouping by $n$ satisfies this at round 1.  Columns of different
sizes merge naturally in later rounds when their residuals land at the same $n$.

**Why no Arcane?**  The univariate Vortex requires uniform column sizes (enforced by
Arcane = Stitcher + Splitter).  The multilinear approach sidesteps this: each $n$-group runs
its own recursion at its native size, so no normalization preprocessing is needed.

---

## 3. Sumcheck and Batching

After one vortex round there are many residual `MultilinearEval` claims at mixed sizes.
`Batch` batches all of them into one sumcheck.

### 3.1 Key Identity

Every multilinear evaluation claim is a sum over the Boolean hypercube:

$$\widetilde{P}_k(c_k) = \sum_{x \in \{0,1\}^n} \mathrm{eq}(c_k, x) \cdot P_k(x),
  \quad
  \mathrm{eq}(c_k, x) = \prod_i \bigl({c_k}_i x_i + (1-{c_k}_i)(1-x_i)\bigr)$$

So the claim $\widetilde{P}_k(c_k) = y_k$ is equivalent to the sumcheck target
$\sum_x \mathrm{eq}(c_k, x) \cdot P_k(x) = y_k$.

### 3.2 Batching $K$ Claims into One Sumcheck

Draw Fiat-Shamir coin $\lambda$.  Embed each $P_k$ of $n_k$ variables into
$n_\text{max}$-variable space by value-repetition: $E(P_k)[j] = P_k[j \bmod 2^{n_k}]$.
Form the combined polynomial:

$$F(x) = \sum_{k=0}^{K-1} \lambda^k \cdot \mathrm{eq}(c_k^{\text{ext}}, x) \cdot E(P_k)(x)$$

where $c_k^{\text{ext}}$ zero-pads $c_k$ to $n_\text{max}$ coordinates.  The combined
sumcheck target is $\sum_{x \in \{0,1\}^{n_\text{max}}} F(x) = \sum_k \lambda^k \cdot y_k$.

The sumcheck runs $n_\text{max}$ rounds.  At the end the verifier holds a shared challenge
$c = (c_0, \ldots, c_{n_\text{max}-1})$ and checks:

$$\sum_k \lambda^k \cdot \mathrm{eq}(c_k^{\text{ext}}, c) \cdot P_k(c_{0:n_k})
  \stackrel{?}{=} \text{last round polynomial evaluated at } c_{n_\text{max}-1}$$

The verifier computes $\lambda^k$ and $\mathrm{eq}$ itself; it needs only each $P_k(c_{0:n_k})$.
The output is $K$ residual `MultilinearEval` claims — one per $P_k$ — all at the first $n_k$
coordinates of the shared challenge $c$.

### 3.3 Why Batch (Proof Size)

$F$ is never committed; it exists only as a virtual polynomial inside the sumcheck.  The
savings come from sharing the $n_\text{max}$ round polynomials across all $K$ claims:

| | Round polynomial cost | Final evals |
|---|---|---|
| 1 combined sumcheck | $n_\text{max} \times 3$ fext elements | $K$ fext elements |
| $K$ separate sumchecks | $\sum_k n_k \times 3$ fext elements | $K$ fext elements |

With $K = 2626$, $n_\text{max} = 24$: combined sends **72** round-poly elements; separate
sends up to **189,000** — a factor-$K$ saving.  The shared output challenge also means all
$K$ residuals land at the same point, so the next vortex round can commit to them all in one
Merkle tree.

### 3.4 Why Batch After, Not Before, the Vortex Step

Right after `InsertBootstrapperOpenings` the active claims have up to $n = 24$ variables
($2^{24}$ rows).  Batching first would launch a 24-round sumcheck with $O(2^{23})$ operations
per round, and the final oracle query would demand a Vortex opening of a $2^{24}$-row matrix.

`CompileRound` first halves the variable count: each $n$-variable claim becomes two
$\lfloor n/2 \rfloor$-variable claims (UCols and RowEvals), committed before any sumcheck
challenge is drawn.  `Batch` then runs a $\lfloor n/2 \rfloor$-round sumcheck over
$2^{n/2}$-row committed oracles — a factor-$2^{n/2}$ cost reduction.

### 3.5 Prover Work

Naively, processing $F$ costs $O(K \cdot 2^{n_\text{max}})$.  The eq-selector avoids this:
for sumcheck rounds $r \geq n_k$, $P_k$'s contribution collapses to a scalar (the high
coordinates of $c_k^\text{ext}$ are zero).  Total prover cost: $O\!\left(\sum_k 2^{n_k}\right)$
— the same as $K$ separate sumchecks.

### 3.6 Why No MPTS

In univariate schemes, batching $K$ claims at $K$ distinct points requires **MPTS**
(Multi-Point To Sumcheck): form quotient $q_k(x) = (P_k(x)-y_k)/(x-c_k)$, commit to $q_k$,
then open a random linear combination.  This introduces $K$ additional commitments.

Multilinear evaluation avoids this entirely.  The eq-selector is degree-1 per variable and
factors over the hypercube, so a claim at any point $c_k$ is already a sum over $\{0,1\}^n$:

$$\widetilde{P}_k(c_k) = \sum_{x \in \{0,1\}^n} \mathrm{eq}(c_k, x) \cdot P_k(x)$$

Claims at different points combine directly into one sumcheck via $F$ without any quotient
polynomial or extra commitment.  The $\lambda^k$ weights ensure linear independence
(Schwartz-Zippel), and $F$ is never committed.

---

## 4. Multilinear Vortex

Given $K$ committed columns $P_0, \ldots, P_{K-1}$ of size $2^n$, a single vortex round
proves $\widetilde{P}_k(c) = y_k$ for each $k$ by reducing each $n$-variable claim to two
$\lfloor n/2 \rfloor$-variable claims, while binding the prover to the committed columns via
RS-encoding and Merkle spot-checks.

### 4.1 Matrix Split

Split $n = n_\text{row} + n_\text{col}$ with $n_\text{row} = \lceil n/2 \rceil$.  Reshape
$P_k$ row-major into a $2^{n_\text{row}} \times 2^{n_\text{col}}$ matrix:

$$M_k[b][j] = P_k\!\left[b \cdot 2^{n_\text{col}} + j\right],
  \quad b \in \{0,\ldots,2^{n_\text{row}}-1\},\; j \in \{0,\ldots,2^{n_\text{col}}-1\}$$

Split the evaluation point $c = (c_\text{row} \| c_\text{col})$ with
$c_\text{row} \in \mathbb{F}_\text{ext}^{n_\text{row}}$ and
$c_\text{col} \in \mathbb{F}_\text{ext}^{n_\text{col}}$.  The multilinear extension factors
as:

$$\widetilde{P}_k(c) = \sum_{b \in \{0,1\}^{n_\text{row}}} \mathrm{eq}(b, c_\text{row})
  \cdot \widetilde{M_k[b,\cdot]}(c_\text{col})$$

### 4.2 Prover: Commit, Then Combine

**Step 1 — Commit to the original matrix.**  RS-encode each row of $M_k$ (rate 2) and
insert all rows into a Merkle tree.  The root enters the Fiat-Shamir transcript, binding
the prover to $P_k$ before the batching randomness $\alpha$ is chosen.

**Step 2 — Draw $\alpha$ and compute derived columns.**  Draw one Fiat-Shamir coin
$\alpha \in \mathbb{F}_\text{ext}$, shared across all $K$ polynomials in this batch.  For
each $k$, compute:

**UAlpha** (length $2^{n_\text{col}}$) — $\alpha$-weighted combination of matrix rows:
$$U_\alpha^k[j] = \sum_{b=0}^{2^{n_\text{row}}-1} \alpha^b \cdot M_k[b][j]$$

**RowEvals** (length $2^{n_\text{row}}$) — each row evaluated at $c_\text{col}$:
$$\mathrm{RowEvals}_k[b] = \widetilde{M_k[b,\cdot]}(c_\text{col})$$

**Linking value** $v_k = \sum_{b} \alpha^b \cdot \mathrm{RowEvals}_k[b]$

**Step 3 — Commit to derived columns.**  Group UAlpha and RowEvals columns by size and
insert them into Merkle trees (one tree per size group).

### 4.3 Two Reduced Claims per Polynomial

| Claim | Column | Evaluation point | Variables |
|---|---|---|---|
| **UCols** | $U_\alpha^k$ | $c_\text{col}$ | $n_\text{col} = \lfloor n/2 \rfloor$ |
| **RowClaims** | $\mathrm{RowEvals}_k$ | $c_\text{row}$ | $n_\text{row} = \lceil n/2 \rceil$ |

**Why these two claims suffice.**  RowClaims directly proves the original:

$$\widetilde{P}_k(c)
  = \sum_b \mathrm{eq}(b, c_\text{row}) \cdot \widetilde{M_k[b,\cdot]}(c_\text{col})
  = \sum_b \mathrm{eq}(b, c_\text{row}) \cdot \mathrm{RowEvals}_k[b]
  = \widetilde{\mathrm{RowEvals}_k}(c_\text{row})
  = y_k$$

However, RowEvals is a fresh committed column; the verifier must also check that it
contains the correct row evaluations of the committed $P_k$.  Checks 2–4 (§4.4) handle
this: by Schwartz-Zippel over $\alpha$, the combination
$U_\alpha^k = \sum_b \alpha^b \cdot M_k[b, \cdot]$ holds with high probability, which
forces $\mathrm{RowEvals}_k[b] = \widetilde{M_k[b,\cdot]}(c_\text{col})$ for all $b$.

### 4.4 Checks Overview

A single vortex round introduces five checks.  Checks 1–3 are verified in the current round;
checks 4–5 become `MultilinearEval` queries for the next round.

| # | Check | Mechanism | Equation |
|---|---|---|---|
| 1 | **Merkle spot-check** | Merkle path reconstruction | sibling paths for $t$ opened positions reconstruct to root |
| 2 | **Linear combination check** | Field arithmetic at $t$ positions | $\mathrm{enc}(U_\alpha)[j] \stackrel{?}{=} \sum_b \alpha^b \cdot \mathrm{enc}(M[b,\cdot])[j]$ |
| 3 | **Evaluation check** | Field arithmetic (dot product) | $v \stackrel{?}{=} \sum_b \alpha^b \cdot \mathrm{RowEvals}[b]$ |
| 4 | **UCols claim** | Multilinear opening (next round) | $\widetilde{U_\alpha}(c_\text{col}) \stackrel{?}{=} v$ |
| 5 | **RowClaims** | Multilinear opening (next round) | $y \stackrel{?}{=} \widetilde{\mathrm{RowEvals}}(c_\text{row})$ |

Checks 1–2 bind $U_\alpha$ to the committed $P_k$ before $\alpha$ is drawn.  Check 3 ties
the linking value $v$ to the committed RowEvals column.  Checks 4–5 are deferred to the
next recursion level.

The half-size claims proved in the next round are:

$$y = \underbrace{\widetilde{\mathrm{RowEvals}}(c_\text{row})}_{\text{Check 5}} \qquad
v = \underbrace{\widetilde{U_\alpha}(c_\text{col})}_{\text{Check 4}} = \underbrace{\sum_b \alpha^b \cdot \mathrm{RowEvals}[b]}_{\text{Check 3}}$$

$y$ and $v$ are **distinct**: $y$ is eq-weighted (the original multilinear evaluation),
while $v$ is $\alpha$-power-weighted (an auxiliary linking value).

### 4.5 Termination

When $n = 1$ the column has size 2 and is evaluated directly:
$(1-c) \cdot P[0] + c \cdot P[1] \stackrel{?}{=} y$.  No further recursion.

---

## 5. Comparison with Univariate Vortex

| Property | Multilinear Vortex | Univariate Vortex |
|---|---|---|
| Column normalization | Not needed — native sizes per group | Arcane (Stitcher + Splitter) required |
| Commitment | Poseidon2 + Merkle, one tree per group | SIS hash + Merkle, one tree per round |
| Recursion depth | $\lceil \log_2 n \rceil$ rounds | Fixed 4 rounds (SelfRecurse) |
| Multi-point batching | Sumcheck via eq-selector, no extra commitments | MPTS required (quotient polynomials) |
| Prover time (78M cells, 5 rounds) | **~4.5 s** (rate=2, t=256) | **~10.6 s** (~2.3× slower) |
| Gnark verifier | Not yet implemented | Fully implemented |

---

## 6. Compile Chain

```
wizard.Compile(
  InsertBootstrapperOpenings,       ← group round-0 cols by n; emit MultilinearEval claims
  multilinvortex.CompileRound,      ← matrix split → UAlpha, RowEvals; RS+Merkle commit; register all checks
  multilineareval.Batch,            ← batch residuals into one sumcheck; shared challenge
  ... × 4 more rounds ...
  dummy.Compile,                    ← no-op terminator; all queries resolved
)
```

Each `CompileRound + Batch` iteration halves the number of variables of every active claim.

---

## 7. Worked Example: Two Rounds for $n = 4$

Three committed columns: $P_0, P_1$ (size $2^4 = 16$) and $Q$ (size $2^2 = 4$).

### Setup — InsertBootstrapperOpenings

| Group | Columns | FS evaluation point |
|---|---|---|
| $n=4$ | $P_0, P_1$ | $c = (c_0,c_1,c_2,c_3) \in \mathbb{F}_\text{ext}^4$ |
| $n=2$ | $Q$ | $d = (d_0,d_1) \in \mathbb{F}_\text{ext}^2$ |

Claims entering the pipeline:
$\widetilde{P}_0(c) = y_0$, $\widetilde{P}_1(c) = y_1$, $\widetilde{Q}(d) = z$.

---

### Round 1 — CompileRound

**$n=4$ group** ($K=2$, reshape into $4 \times 4$ matrices):

Commit to RS-encoded $P_0, P_1$; roots enter FS transcript.  Draw $\alpha$.

$$U_\alpha^k[j] = \sum_{b=0}^{3} \alpha^b \cdot P_k[4b+j], \quad k=0,1 \quad \text{(size 4)}$$
$$\mathrm{RowEvals}_k[b] = \widetilde{P_k[b,\cdot]}(c_2,c_3), \quad k=0,1 \quad \text{(size 4)}$$

Evaluation check: $\sum_{b=0}^{3} \alpha^b \cdot \mathrm{RowEvals}_k[b] \stackrel{?}{=} v_k$.

New claims: UCols$_k$ at $(c_2,c_3)$ claiming $v_k$; RowClaims$_k$ at $(c_0,c_1)$ claiming $y_k$.

**$n=2$ group** ($K=1$, reshape into $2 \times 2$ matrix):

Commit to RS-encoded $Q$; root enters FS transcript.  Draw $\beta$.

$$U_\beta[j] = Q[j] + \beta \cdot Q[2+j], \quad j=0,1 \quad \text{(size 2)}$$
$$\mathrm{RowEvals}_Q[b] = \widetilde{Q[b,\cdot]}(d_1), \quad b=0,1 \quad \text{(size 2)}$$

Evaluation check: $\mathrm{RowEvals}_Q[0] + \beta \cdot \mathrm{RowEvals}_Q[1] \stackrel{?}{=} v_Q$.

New claims: UCols$_Q$ at $(d_1)$ claiming $v_Q$; RowClaims$_Q$ at $(d_0)$ claiming $z$.

**Merkle commits** — derived columns grouped by size:

| Tree | Columns | Size |
|---|---|---|
| size-4 | $U_\alpha^0,\, U_\alpha^1,\, \mathrm{RowEvals}_0,\, \mathrm{RowEvals}_1$ | 4 fext elements each |
| size-2 | $U_\beta,\, \mathrm{RowEvals}_Q$ | 2 fext elements each |

Linear combination checks at $t$ opened positions:
$$\mathrm{enc}(U_\alpha^k)[j] \stackrel{?}{=} \sum_{b=0}^{3} \alpha^b \cdot \mathrm{enc}(P_k)[b][j], \qquad
\mathrm{enc}(U_\beta)[j] \stackrel{?}{=} \mathrm{enc}(Q)[0][j] + \beta \cdot \mathrm{enc}(Q)[1][j]$$

### Round 1 — Batch

Six residual claims ($n_\text{max} = 2$):

| Column | $n_k$ | Evaluation point |
|---|---|---|
| $U_\alpha^0$, $U_\alpha^1$, $\mathrm{RowEvals}_0$, $\mathrm{RowEvals}_1$ | 2 | $(c_2,c_3)$ or $(c_0,c_1)$ |
| $U_\beta$, $\mathrm{RowEvals}_Q$ | 1 | $d_1$ or $d_0$ |

One 2-round sumcheck on $F(x_0,x_1) = \sum_k \lambda^k \cdot \mathrm{eq}(c_k^\text{ext},x) \cdot E(P_k)(x)$.

Output: shared challenge $c' = (c'_0, c'_1)$.  All six residuals now at $c'$ (using first $n_k$ coordinates).

---

### Round 2 — CompileRound

**$n=2$ group** (four columns $U_\alpha^0, U_\alpha^1, \mathrm{RowEvals}_0, \mathrm{RowEvals}_1$, each $K=1$):

Draw $\alpha'$.  For each column $P_k$:

$$\mathrm{UAlpha2}_k[j] = P_k[j] + \alpha' \cdot P_k[2+j], \quad j=0,1$$
$$\mathrm{RowEvals2}_k[b] = \widetilde{P_k[b,\cdot]}(c'_1), \quad b=0,1$$

Evaluation check: $\mathrm{RowEvals2}_k[0] + \alpha' \cdot \mathrm{RowEvals2}_k[1] \stackrel{?}{=} v'_k$.

New claims: UCols2$_k$ at $(c'_1)$ claiming $v'_k$; RowClaims2$_k$ at $(c'_0)$ claiming $y'_k$.  Both $n=1$.

**$n=1$ group** ($U_\beta$, $\mathrm{RowEvals}_Q$ — terminal): evaluated directly, no further splitting.

**Merkle commit** — 8 new size-2 columns into one tree.

### Round 2 — Batch

Output: shared challenge $c''$.  All 8 residuals from the $n=2$ group now at $n=1$.

---

### Terminal

For each size-2 column $C_k$ at challenge $c''$:
$$(1 - c'') \cdot C_k[0] + c'' \cdot C_k[1] \stackrel{?}{=} \text{claimed value}$$

---

### Complete Verifier Check Inventory

All checks are registered at the final wizard round:

| Check | Registered by | What it verifies |
|---|---|---|
| Merkle spot-check: orig $P_0, P_1$ | CompileRound 1 | RS-encoded rows of $P_0, P_1$ hash to committed root |
| Linear combination check: $P_0, P_1$ | CompileRound 1 | $\mathrm{enc}(U_\alpha^k)[j] = \sum_b \alpha^b \cdot \mathrm{enc}(P_k)[b][j]$ |
| Merkle spot-check: orig $Q$ | CompileRound 1 | RS-encoded rows of $Q$ hash to committed root |
| Linear combination check: $Q$ | CompileRound 1 | $\mathrm{enc}(U_\beta)[j] = \mathrm{enc}(Q)[0][j] + \beta \cdot \mathrm{enc}(Q)[1][j]$ |
| Merkle spot-check: size-4 derived cols | CompileRound 1 | $U_\alpha^{0,1}$, RowEvals$_{0,1}$ hash to roots |
| Merkle spot-check: size-2 derived cols | CompileRound 1 | $U_\beta$, RowEvals$_Q$ hash to roots |
| Evaluation check, $k=0,1$ | CompileRound 1 | $\sum_b \alpha^b \cdot \mathrm{RowEvals}_k[b] = v_k$ |
| Evaluation check, $Q$ | CompileRound 1 | $\mathrm{RowEvals}_Q[0] + \beta \cdot \mathrm{RowEvals}_Q[1] = v_Q$ |
| Sumcheck consistency | Batch 1 | $\sum_k \lambda^k \cdot \mathrm{eq}(c_k^{\text{ext}}, c') \cdot P_k(c'_{0:n_k}) = \text{last round poly at } c'_1$ |
| Merkle spot-check: 8 size-2 cols | CompileRound 2 | UAlpha2 / RowEvals2 hash to roots |
| Evaluation check, round 2 | CompileRound 2 | $\mathrm{RowEvals2}_k[0] + \alpha' \cdot \mathrm{RowEvals2}_k[1] = v'_k$ |
| Sumcheck consistency | Batch 2 | $\sum_k \lambda^k \cdot \mathrm{eq}(c_k^{\text{ext}}, c'') \cdot P_k(c''_{0:n_k}) = \text{last round poly at } c''_1$ |
| Terminal evaluations | CompileRound 2 | $(1-c'') C_k[0] + c'' C_k[1] = \text{claimed value}$ |

**Soundness chain.**  $P_0, P_1, Q$ are committed before $\alpha$/$\beta$ are drawn, so the
linear combination checks bind each $U_\alpha^k$ to the committed rows.  By Schwartz-Zippel
over $\alpha$, checks 2–4 together force $\mathrm{RowEvals}_k[b] = \widetilde{M_k[b,\cdot]}(c_\text{col})$
for all $b$ with high probability.  RowClaims (check 5) then directly proves the original
multilinear claim.  The evaluation checks propagate this argument through each recursion
level until the terminal $n=1$ check closes the chain.
