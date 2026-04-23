# Compiler Verifier Actions

Each compiler in the protocol registers one or more **verifier actions** via
`comp.RegisterVerifierAction()`. A verifier action's `Run` function is called
during proof verification to check that the prover behaved honestly. This
document catalogues every verifier action, grouped by compiler, and explains the
mathematics behind each check.

---

## Table of Contents

1. [globalcs — Global Constraint System](#1-globalcs--global-constraint-system)
2. [vortex — Vortex Polynomial Commitment Scheme](#2-vortex--vortex-polynomial-commitment-scheme)
3. [logderivativesum — Log-Derivative Lookup Argument](#3-logderivativesum--log-derivative-lookup-argument)
4. [permutation — Permutation Grand Product](#4-permutation--permutation-grand-product)
5. [splitextension — Field Extension Splitting](#5-splitextension--field-extension-splitting)
6. [innerproduct — Inner Product Batching](#6-innerproduct--inner-product-batching)
7. [univariates — Naturalize Composite Columns](#7-univariates--naturalize-composite-columns)
8. [plonkinwizard — Activator and Mask Checking](#8-plonkinwizard--activator-and-mask-checking)
9. [stitchsplit — Stitched Constraint Verification](#9-stitchsplit--stitched-constraint-verification)
10. [mpts — Multipoint-to-Singlepoint Reduction](#10-mpts--multipoint-to-singlepoint-reduction)
11. [horner — Horner Projection Verification](#11-horner--horner-projection-verification)
12. [recursion — Recursion Consistency Check](#12-recursion--recursion-consistency-check)
13. [dummy — Brute-Force Constraint Verification](#13-dummy--brute-force-constraint-verification)

---

## 1. globalcs — Global Constraint System

**File:** `globalcs/compile.go`, `globalcs/evaluation.go`, `globalcs/quotient.go`

**Registered Action:** `EvaluationVerifier` (at `evaluationRound`)

**Core Function:** `recombineQuotientSharesEvaluation`

### What It Verifies

The global constraint compiler checks that all global constraint expressions
`C(X)` are satisfied over the evaluation domain. The fundamental identity being
verified at a random challenge point `r` is:

$$
C(r) = (r^n - 1) \cdot Q(r)
$$

where:
- $C(r)$ is the aggregated constraint polynomial evaluated at $r$
- $r^n - 1$ is the vanishing polynomial (annihilator) of the domain of size $n$
- $Q(r)$ is the quotient polynomial evaluated at $r$

If $C(X)$ vanishes on the domain (i.e., the constraints hold), then $Q(X) =
C(X) / (X^n - 1)$ exists as a polynomial. Checking the identity at a random
point $r$ provides soundness via the Schwartz-Zippel lemma.

### Quotient Shares

The quotient polynomial $Q(X)$ can be large, so it is decomposed into
**quotient shares**. For each constraint ratio $\rho_i$, the quotient is split
into $\rho_i$ shares, each evaluated on a different coset of the multiplicative
group:

$$
Q_j(X) \text{ is the share on coset } H \cdot g^j, \quad j \in \{0, \ldots, \rho_i - 1\}
$$

where $g$ is the multiplicative group generator and $H$ is the base domain.

### Recombination (`recombineQuotientSharesEvaluation`)

The function reconstructs $Q(r)$ from its share evaluations using Lagrange
interpolation on roots of unity. For each ratio $\rho_i$ with shares
$\{y_0, y_1, \ldots, y_{\rho_i - 1}\}$:

1. **Shift the evaluation point:** $\tilde{r} = r / g$ (to account for coset
   positioning).

2. **Validate evaluation points:** For each share $k$, verify the prover
   evaluated at $\tilde{r} \cdot \omega_N^{-k}$ where $\omega_N$ is the
   appropriate root of unity.

3. **Lagrange interpolation on the shares:**

$$
\text{inner} = \sum_{k=0}^{\rho_i - 1} \frac{y_k}{\tilde{r}^m \cdot \omega_{\rho_i}^{-k} - 1}
$$

   where $m$ is the base domain size and $\omega_{\rho_i}$ is a primitive
   $\rho_i$-th root of unity.

4. **Apply the outer correction factor:**

$$
Q_i(r) = \frac{1}{\rho_i} \cdot (\tilde{r}^n - 1) \cdot \text{inner}
$$

   where $n = m \cdot \rho_i$ is the full domain size.

### Verification Steps (in `Run`)

1. Evaluate the constraint expression $C(r)$ using witness column openings,
   random coins, and periodic column evaluations.
2. Reconstruct $Q(r)$ from shares via `recombineQuotientSharesEvaluation`.
3. Check: $C(r) \stackrel{?}{=} (r^n - 1) \cdot Q(r)$.

---

## 2. vortex — Vortex Polynomial Commitment Scheme

**File:** `vortex/compiler.go`, `vortex/verifier.go`

Two verifier actions are registered:

### 2a. `ExplicitPolynomialEval` (at `lastRound`)

**What It Verifies:**

A lightweight pre-check that directly evaluates **public polynomials** at the
query point using Lagrange interpolation and compares against claimed values.

For each public polynomial $P$ in the univariate query:

$$
P(x) = \text{EvaluateLagrange}(\text{coefficients}, x)
$$

The verifier computes $P(x)$ from the known coefficients and asserts:

$$
P(x) \stackrel{?}{=} y_{\text{claimed}}
$$

This catches evaluation inconsistencies before the expensive cryptographic check.

### 2b. `VortexVerifierAction` (at `lastRound + 2`)

**What It Verifies:**

The core cryptographic commitment verification. Dispatches to either Koalabear
or BLS12-377 variants. The verification performs:

1. **Merkle root collection:** Gathers commitment roots from all rounds
   (precomputed, non-SIS, and SIS rounds).
2. **Proof component extraction:** Extracts $\alpha$ (random linear combination
   coin), the linear combination evaluation, entry list (randomly selected
   column indices), and all claimed $Y$ evaluations.
3. **Merkle inclusion proof validation:** Verifies that opened columns are
   committed in the Merkle trees via path proofs.
4. **Reed-Solomon codeword integrity:** Checks that opened columns satisfy
   Reed-Solomon error correction properties.
5. **Linear combination correctness:** Verifies the random linear combination
   across opened columns was computed honestly.

---

## 3. logderivativesum — Log-Derivative Lookup Argument

**File:** `logderivativesum/logderivativesum.go`, `logderivativesum/lookup2logderivsum.go`

Two verifier actions are registered:

### 3a. `FinalEvaluationCheck` (at `lastRound`)

**What It Verifies:**

Checks that the sum of all $Z$-accumulator polynomial final values equals the
claimed log-derivative sum. For $k$ accumulator polynomials $Z_1, \ldots, Z_k$
(one per distinct table size):

$$
\sum_{i=1}^{k} Z_i[n_i - 1] \stackrel{?}{=} \text{claimed\_sum}
$$

where $Z_i[n_i - 1]$ is the opening of the $i$-th accumulator at its final
position. This bridges the polynomial identities (verified globally) with the
specific claimed sum value.

### 3b. `CheckLogDerivativeSumMustBeZero` (at `lastRound + 1`, non-segmented only)

**What It Verifies:**

In the non-segmented case, the global log-derivative sum must be exactly zero.
This is the fundamental property of the log-derivative lookup argument:

$$
\Sigma = \sum_{\text{all log-derivative terms}} = 0
$$

The `Run` function simply retrieves the computed sum and asserts
`Sum.IsZero()`. A non-zero sum means the lookup relation is violated.

> **Note:** This action is skipped when a segmenter is used, because
> segmentation invalidates the global-sum-equals-zero property.

---

## 4. permutation — Permutation Grand Product

**File:** `permutation/permutation.go`, `permutation/grand_product.go`, `permutation/verifier.go`

Two verifier actions are registered:

### 4a. `CheckGrandProductIsOne` (at `query.Round`)

**What It Verifies:**

The grand product argument for permutations must yield exactly 1. The verifier:

1. Retrieves the prover's grand product value $y$.
2. Multiplies in explicit public numerators: $y \leftarrow y \cdot \prod_j \text{Num}_j$.
3. Accumulates explicit public denominators: $d = \prod_j \text{Den}_j$.
4. Checks:

$$
\frac{y}{d} \stackrel{?}{=} 1
$$

A result of 1 confirms that the multiset equality (permutation relation) holds.

### 4b. `FinalProductCheck` (at the grand product round)

**What It Verifies:**

Validates consistency of all $Z$-accumulator columns with the committed grand
product value. For $Z$-columns $Z_1, \ldots, Z_k$:

$$
\left(\prod_{i=1}^{k} Z_i[\text{end}]\right) \cdot \left(\prod_j \text{ExplicitEval}_j\right) \stackrel{?}{=} \text{claimed\_grand\_product}
$$

This ensures the prover's accumulation process was honest.

---

## 5. splitextension — Field Extension Splitting

**File:** `splitextension/split.go`

**Registered Action:** `VerifierCtx` (at `roundID`)

### What It Verifies

Checks that extension field polynomial evaluations are correctly reconstructed
from their base field limbs. Each extension element is decomposed over the basis
$\{1, u, v, uv\}$:

$$
P(x) = P_0(x) + u \cdot P_1(x) + v \cdot P_2(x) + uv \cdot P_3(x)
$$

where $P_0, P_1, P_2, P_3$ are the split base-field polynomials. The verifier
reconstructs the extension value from the four limb evaluations and checks it
matches the alleged original evaluation.

---

## 6. innerproduct — Inner Product Batching

**File:** `innerproduct/context.go`, `innerproduct/verifier.go`

**Registered Action:** `VerifierForSize` (at `lastRound`)

### What It Verifies

Verifies that multiple inner-product queries are correctly batched using random
linear combinations. Given inner-product results $y_1, \ldots, y_m$ and a
batching coin $\beta$:

$$
\sum_{i=1}^{m} y_i \cdot \beta^{i-1} \stackrel{?}{=} \text{SummationOpening}
$$

The left side is computed via Horner's method. `SummationOpening` is the local
opening of the Summation column at its final position. This collapses $m$
inner-product checks into a single batched check.

---

## 7. univariates — Naturalize Composite Columns

**File:** `univariates/naturalize.go`

**Registered Action:** `NaturalizeVerifierAction` (at `roundID`)

### What It Verifies

Verifies that non-natural polynomial representations (columns with offsets,
repeats, or interleaving) are consistent with their naturalized sub-queries.
The verifier:

1. **Derives evaluation points** for sub-queries from the original query point,
   accounting for transformations:
   - **Offset:** $P(x \omega^k)$ becomes $P(x)$ at $x = \omega^k t$
   - **Repeat:** $P(x^2)$ becomes $P(x)$ at $x = t^2$
   - **Interleaving:** $I(X) = \frac{1}{2} P(X)(X^n - 1) - \frac{1}{2} P(-X)(X^n + 1)$

2. **Validates** that all derived $X$ values match what was submitted in the
   sub-queries.

3. **Recovers and checks** that $Y$ values are consistent with the
   transformations.

---

## 8. plonkinwizard — Activator and Mask Checking

**File:** `plonkinwizard/compile.go`, `plonkinwizard/verifier.go`

**Registered Action:** `CheckActivatorAndMask` (at `round`)

### What It Verifies

Verifies that binary activator columns (circuit instance selectors) correctly
match the circuit mask at their evaluation points. For each activator instance
$i$:

$$
\text{mask}[i \cdot \text{offset}] \stackrel{?}{=} \text{activator}[i]
$$

This ensures activators only activate where the underlying circuit mask permits,
maintaining the integrity of the PLONK gate selector mechanism.

---

## 9. stitchsplit — Stitched Constraint Verification

**File:** `stitchsplit/stitcher_constraints.go`

**Registered Action:** `QueryVerifierAction` (one per round with explicit queries)

### What It Verifies

After column stitching merges multiple smaller columns into larger ones for
efficiency, the verifier explicitly checks the rewritten constraints. For each
query in the collection:

- **Local constraints:** Verifies the constraint expression evaluates to zero at
  the queried positions.
- **Global constraints:** Verifies the constraint holds over the subsampled
  domain.

The `Run` function iterates over all queries and calls `Check()` on each.

---

## 10. mpts — Multipoint-to-Singlepoint Reduction

**File:** `mpts/multipoint_to_singlepoint.go`, `mpts/verifier.go`

**Registered Action:** `VerifierAction` (at `getNumRound + 1`)

### What It Verifies

Reduces multiple univariate evaluation queries at different points into a single
evaluation using a quotient polynomial and randomness. Given evaluation queries
$\{(x_i, \{y_{i,k}\})\}$ for polynomials $\{P_k\}$, the verifier checks:

$$
Q(r) = \sum_{i} \zeta_i \cdot \left[\sum_{k} \rho^k \cdot (P_k(r) - y_{k,i})\right]
$$

where:
- $\zeta_i = \lambda^i / (r - x_i)$ are barycentric-style weights
- $\lambda, \rho$ are random coins for batching across points and polynomials
- $r$ is the random evaluation point

The inner sum over $k$ is computed via Horner's method for efficiency. The
final result is compared against the quotient column's evaluation $Q(r)$.

---

## 11. horner — Horner Projection Verification

**File:** `horner/projection_to_horner.go`

**Registered Action:** `CheckHornerQuery` (at `round`)

### What It Verifies

Verifies that a Horner-form polynomial evaluation yields zero. The Horner query
evaluates:

$$
\text{result} = (\cdots((a_0 \cdot x + a_1) \cdot x + a_2) \cdot x + \cdots + a_n)
$$

The verifier checks two conditions:

1. **Final result is zero:** $\text{FinalResult} \stackrel{?}{=} 0$
2. **All initial values are zero:** $N_{0,j} \stackrel{?}{=} 0$ for each part $j$

This encodes that a specific polynomial constraint evaluates to zero at the
queried point.

---

## 12. recursion — Recursion Consistency Check

**File:** `recursion/recursion.go`, `recursion/actions.go`

**Registered Action:** `ConsistencyCheck` (at round 0)

### What It Verifies

Verifies consistency between the recursed inner proof's Vortex statements and
the Plonk-in-Wizard circuit's public inputs. The check ensures the recursion
circuit correctly extracted all cryptographic data from the inner proof:

1. **Evaluation point $X$ consistency:** The 4 base-field limbs of the
   extension element match:

$$
\text{circX}[0..3] \stackrel{?}{=} (X.B_0.A_0,\; X.B_0.A_1,\; X.B_1.A_0,\; X.B_1.A_1)
$$

2. **Evaluation results $Y$ consistency:** For each of the $n$ extension
   values, the 4 limbs match:

$$
\text{circYs}[4j..4j+3] \stackrel{?}{=} (Y_j.B_0.A_0,\; Y_j.B_0.A_1,\; Y_j.B_1.A_0,\; Y_j.B_1.A_1)
$$

3. **Merkle root consistency:** Each commitment root (8 limbs per root) matches
   between the inner proof and the circuit's public inputs.

---

## 13. dummy — Brute-Force Constraint Verification

**File:** `dummy/dummy.go`

**Registered Action:** `DummyVerifierAction` (at `numRounds - 1`)

### What It Verifies

A testing/debugging compiler that bypasses all PCS optimizations and **manually
verifies every single constraint**. The `Run` function iterates (in parallel)
over all queries and calls `Check()` on each:

- **Parametrized queries:** Univariate evaluations, inner products, local
  openings, log-derivative sums, grand products, etc.
- **Non-parametrized queries:** Local constraints, global constraints.

This compiler is not used in production. It serves as a correctness oracle
during development.

> **Note:** The Gnark circuit path does not implement the dummy verifier.
