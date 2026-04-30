# MPTS: Multi-Point-to-Single-Point Reduction via Sumcheck

This document describes the mathematical technique used to convert
`LagrangeEval` queries (univariate polynomial evaluation claims) into
`MultilinearEval` queries (multilinear polynomial evaluation claims) in the
Vortex compilation pipeline.

---

## Problem Statement

We have $m$ univariate polynomials $P_0, \ldots, P_{m-1}$ of degree $< n = 2^k$,
each stored in **Lagrange form** over a multiplicative coset $\omega$-domain:

$$\text{table}_j[i] = P_j(\omega^i), \quad i = 0, \ldots, n-1$$

We hold evaluation claims $(z_j, y_j)$ with $P_j(z_j) = y_j$ for each $j$.
The goal is to reduce all these claims to a single `MultilinearEval` query
that the rest of the protocol can handle uniformly.

---

## Why Not the Lagrange-Direct Approach

A natural attempt is to write
$$P(z) = \sum_{i=0}^{n-1} P(\omega^i)\, L_i(z)$$
and interpret this as a sumcheck over the boolean hypercube with the MLE
$\widetilde{P}$ of the evaluation table and a mask $\lambda_z(b) = L_{r(b)}(z)$
(where $r(b)$ is the MSB-first integer index):
$$P(z) = \sum_{b \in \{0,1\}^k} \widetilde{P}(b)\, \lambda_z(b).$$

The problem: after the sumcheck with challenges $h' = (h'_0, \ldots, h'_{k-1})$,
the verifier must evaluate $\widetilde{\lambda}_z(h')$, which is the inner product
$$\widetilde{\lambda}_z(h') = \sum_{i=0}^{n-1} L_i(z)\cdot\mathrm{eq}(r^{-1}(i),\, h').$$
This requires $O(n)$ field operations — no product formula exists — making the
verifier cost linear in the table size. This is unacceptable for recursive settings.

---

## The Monomial Correspondence

### Setup

Let $\hat{P} = [c_0, c_1, \ldots, c_{n-1}]$ be the **monomial coefficient vector** of $P$:
$$P(x) = \sum_{i=0}^{n-1} c_i\, x^i.$$

Given the Lagrange table, $\hat{P}$ is obtained by an inverse NTT:
$$\hat{P} = \mathrm{iNTT}(\text{table}).$$

Define the MLE $\bar{P}$ of $\hat{P}$ using the **product-monomial basis**
and the MSB-first bit-index map $r(b) = b_0 2^{k-1} + b_1 2^{k-2} + \cdots + b_{k-1} 2^0$:
$$\bar{P}(X_0, \ldots, X_{k-1}) = \sum_{b \in \{0,1\}^k} c_{r(b)}\, \prod_{i=0}^{k-1} X_i^{b_i}.$$

### Key Identity

$$P(z) = \sum_{i=0}^{n-1} c_i\, z^i
= \sum_{b \in \{0,1\}^k} c_{r(b)}\, z^{r(b)}
= \bar{P}(z^{2^{k-1}},\, z^{2^{k-2}},\, \ldots,\, z^{2^0}).$$

So a claim $P(z) = y$ is equivalent to a multilinear evaluation claim
$\bar{P}\!\left(z^{2^{k-1}}, \ldots, z\right) = y$.

### Efficient Verifier Check

After $k$ rounds of sumcheck with challenges $h' = (h'_0, \ldots, h'_{k-1})$
(MSB-first, so $h'_0$ selects the most significant bit), the verifier evaluates
the MLE of the monomial mask $g(b) = z^{r(b)}$ at $h'$:

$$\widetilde{g}(h') = \prod_{i=0}^{k-1}\Bigl(1 + h'_i\bigl(z^{2^{k-1-i}} - 1\bigr)\Bigr).$$

**Derivation**: For variable $i$, the bilinear interpolation of $z^{r(b)}$ on $b_i$ gives:
$$
(1 - h'_i)\cdot\underbrace{(z^{2^{k-1-i}})^0}_{=\,1} + h'_i\cdot\underbrace{(z^{2^{k-1-i}})^1}_{=\,z^{2^{k-1-i}}}
= 1 + h'_i\bigl(z^{2^{k-1-i}} - 1\bigr).
$$
The product-monomial basis contributes factor $1$ when $b_i = 0$ (not $1 - z^{2^{k-1-i}}$
as the eq formula would give).

**Cost**: $k - 1$ squarings to obtain $z, z^2, z^4, \ldots, z^{2^{k-1}}$, then $k$
multiplications — total $O(k) = O(\log n)$, independent of $n$.

This function is implemented as `EvalMonomialMaskExt` in
`maths/koalabear/polynomials/product_mle.go`.

---

## Protocol: Batching $m$ Polynomials

**Setup**: $m$ polynomials, each of size $n = 2^k$, with claims $P_j(z_j) = y_j$.
Evaluation points $z_j$ may all differ.

### Step 1 — Prover: iNTT

For each $j$, compute:
$$\hat{P}_j = \mathrm{iNTT}(\text{table}_j) \quad O(n \log n).$$

### Step 2 — Batching

The verifier sends a recombination scalar $\rho$. Define:
$$\hat{Q}[i] = \sum_{j=0}^{m-1} \rho^j\, \hat{P}_j[i], \qquad
M[i] = \sum_{j=0}^{m-1} \rho^j\, z_j^i, \qquad
Y = \sum_{j=0}^{m-1} \rho^j\, y_j.$$

The combined claim is:
$$Y = \sum_{i=0}^{n-1} \hat{Q}[i]\, M[i]
= \sum_{b \in \{0,1\}^k} \hat{Q}[r(b)]\, M[r(b)].$$

The mask table $M$ is built by `BuildMonomialMaskExt` in `product_mle.go`:
$M[i] = \sum_j \rho^j z_j^i$, computed in $O(mn)$.

### Step 3 — Sumcheck

Run a single sumcheck over the boolean hypercube on the identity gate:
$$Y = \sum_{b \in \{0,1\}^k} \hat{Q}[r(b)]\cdot M[r(b)].$$

Here $M$ plays the role of the "eq table" in the prover (injected via
`NewProverStateWithMask`). After $k$ rounds the prover sends round polynomials
and the verifier sends challenges $h' = (h'_0, \ldots, h'_{k-1})$.

### Step 4 — Final Verifier Check

At the end of the sumcheck, the prover has asserted $u = \hat{Q}(h')$.
The verifier checks:
$$y_{\mathrm{final}} = u \cdot \widetilde{G}(h'), \quad \text{where} \quad
\widetilde{G}(h') = \sum_{j=0}^{m-1} \rho^j\, \widetilde{g}_j(h')$$
and each $\widetilde{g}_j(h') = \prod_{i=0}^{k-1}\bigl(1 + h'_i(z_j^{2^{k-1-i}}-1)\bigr)$
in $O(k)$. Total verifier work: $O(mk)$.

### Step 5 — Individual MultilinearEval Claims

The sumcheck produces one combined claim $u = \hat{Q}(h')$. To recover individual
claims, each polynomial $P_j$ contributes:
$$u_j = \bar{P}_j(h')$$
and the verifier checks $\sum_j \rho^j u_j = u$. Each $u_j$ becomes a `MultilinearEval`
claim on the coefficient column $\hat{P}_j$ at the shared point $h'$.

---

## Handling Shifts

A `LagrangeEval` with `ShiftingOffset = s` evaluates $P$ at $\omega^s z$. This
is absorbed entirely into the evaluation point: use $\omega^s z$ in place of $z$
when computing $\widetilde{g}_j$. The coefficient vector $\hat{P}_j$ is the same
(no rotation needed); the shift changes only the structured evaluation coordinate
$((\omega^s z)^{2^{k-1}}, \ldots, \omega^s z)$.

---

## Summary of Costs

| Party | Step | Cost |
|---|---|---|
| Prover | iNTT per polynomial | $O(mn \log n)$ |
| Prover | Build combined tables $\hat{Q}$, $M$ | $O(mn)$ |
| Prover | Sumcheck ($k$ rounds) | $O(n)$ |
| Verifier | Sumcheck check | $O(k)$ per round |
| Verifier | Final mask check | $O(mk)$ |

The iNTT is $O(n \log n)$, comparable to the RS-encoding the prover already
performs for Vortex rows — no asymptotic overhead is added.

---

## Code Pointers

| Component | Location |
|---|---|
| `EvalMonomialMaskExt` | `maths/koalabear/polynomials/product_mle.go` |
| `BuildMonomialMaskExt` | `maths/koalabear/polynomials/product_mle.go` |
| `NewProverStateWithMask` | `crypto/koalabear/sumcheck/prover.go` (to be added) |
| `IdentityGate` | `crypto/koalabear/sumcheck/gate.go` (to be added) |
| `LagrangeEval` source query | `wiop/query_lagrange_eval.go` |
| `MultilinearEval` target query | `wiop/query_multilinear_eval.go` |
| MPTS compiler | `wiop/compilers/mpts/` (to be added) |
