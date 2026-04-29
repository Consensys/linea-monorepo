// Package sumcheck implements the sumcheck protocol over the KoalaBear field
// and its degree-4 extension.
//
// # Overview
//
// The sumcheck protocol reduces a claim of the form
//
//	H = Σ_{x ∈ {0,1}ⁿ} eq(q, x) · Gate(A₁(x), A₂(x), …)
//
// to a single evaluation query at a random point in the extension field.
// Each of the n rounds produces a univariate "round polynomial" and consumes
// one verifier challenge. At the end the prover supplies the evaluation of
// each input table and the eq polynomial at the accumulated challenge point.
//
// # Multi-sumcheck
//
// Multiple claims at different evaluation points q₁, q₂, … are combined via
// a caller-supplied recombination challenge μ: the effective eq table is
//
//	Eq = Σ_j μʲ · FoldedEqTable(q_j)
//
// The caller is responsible for deriving μ from a Fiat-Shamir transcript.
//
// # API
//
// The prover is round-by-round and FS-agnostic:
//
//  1. Pre-allocate all scratch memory once: [NewProverConfig]
//  2. Start a proof: [NewProverStateWithEqMask]
//  3. Per round: [ProverState.ComputeRoundPoly], then [ProverState.FoldAndAdvance]
//  4. After n rounds: [ProverState.FinalClaims]
//
// The verifier is also FS-agnostic; the caller provides a callback to hash
// each round polynomial and sample the next challenge: [Verify]
//
// # Gruen optimisation
//
// Round polynomials are stored in Gruen compressed format: evaluations at
// {0, 2, 3, …, d} with P(1) omitted. The verifier reconstructs
// P(1) = claim − P(0), saving one gate evaluation and one extension-field
// element per round in the proof transcript.
//
// # Coordinate convention
//
// Multilinear polynomials follow the same MSB-first convention as the
// maths/koalabear/polynomials package: coords[0] selects the most-significant
// bit of the boolean-hypercube index.
package sumcheck
