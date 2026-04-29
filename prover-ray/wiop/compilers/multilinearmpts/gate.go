// Package multilinearmpts implements the Many-Point-To-Single-Point (MPTS)
// compiler step of the Vortex pipeline. It converts [wiop.LagrangeEval]
// queries into [wiop.MultilinearEval] queries via the monomial-correspondence
// sumcheck described in docs/mpts-sumcheck.md.
package multilinearmpts

import "github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"

// EvalGate is the degree-2 sumcheck gate for the MPTS protocol. It evaluates
//
//	Gate(P̂₀, M₀, P̂₁, M₁, …) = Σᵢ Lambdas[i] · P̂ᵢ · Mᵢ
//
// Input tables are interleaved: inputs[2i] holds P̂ᵢ (the iNTT'd coefficient
// table of polynomial i) and inputs[2i+1] holds Mᵢ (the geometric mask table,
// Mᵢ[h] = rᵢ^{index(h)} for evaluation point rᵢ).
//
// Lambdas holds the batching powers [λ⁰, λ¹, …, λ^{m-1}] for the Fiat-Shamir
// recombination challenge λ.
//
// The gate is used with [sumcheck.NewProverStateWithMask](mask = nil) so that
// all 2m tables are folded uniformly each sumcheck round. After logN rounds
// with challenges h', FinalClaims yields individual claims P̂ᵢ(h') and Mᵢ(h').
// The verifier checks Mᵢ(h') = EvalMonomialMaskExt(rᵢ, h') in O(log n) per
// query, and P̂ᵢ(h') feeds directly into a MultilinearEval query.
type EvalGate struct {
	// Lambdas holds the batching powers [λ⁰, λ¹, …, λ^{m-1}].
	Lambdas []field.Ext
}

// NewEvalGate constructs an EvalGate for m polynomial–mask pairs batched with
// Fiat-Shamir challenge lambda. Lambdas[i] is set to lambda^i.
func NewEvalGate(lambda field.Ext, m int) *EvalGate {
	lambdas := make([]field.Ext, m)
	lambdas[0].SetOne()
	for i := 1; i < m; i++ {
		lambdas[i].Mul(&lambdas[i-1], &lambda)
	}
	return &EvalGate{Lambdas: lambdas}
}

// Degree returns 2.
func (g *EvalGate) Degree() int { return 2 }

// EvalBatch sets res[j] = Σᵢ Lambdas[i] · inputs[2i][j] · inputs[2i+1][j].
// Panics if len(inputs) != 2·len(g.Lambdas).
func (g *EvalGate) EvalBatch(res []field.Ext, inputs ...[]field.Ext) {
	for j := range res {
		res[j].SetZero()
		for i, λ := range g.Lambdas {
			var prod, term field.Ext
			prod.Mul(&inputs[2*i][j], &inputs[2*i+1][j])
			term.Mul(&prod, &λ)
			res[j].Add(&res[j], &term)
		}
	}
}
