package sumcheck

import (
	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
)

// Gate represents a low-degree multivariate polynomial to be sumchecked.
// Inputs and outputs are extension-field slices (tables are lifted to ext at
// prover construction so that extension-field challenges fold them correctly).
type Gate interface {
	// Degree returns the total degree of the polynomial.
	Degree() int

	// EvalBatch sets res[j] = Gate(inputs[0][j], inputs[1][j], …) for all j.
	// res must be pre-allocated with the same length as each inputs[k] slice.
	EvalBatch(res []field.Ext, inputs ...[]field.Ext)
}

// IdentityGate is a degree-1 gate that returns its single input unchanged:
// Gate(x) = x. Used when the weighting is entirely carried by the mask table.
type IdentityGate struct{}

// Degree returns 1.
func (IdentityGate) Degree() int { return 1 }

// EvalBatch implements Gate. Copies inputs[0] into res.
func (IdentityGate) EvalBatch(res []field.Ext, inputs ...[]field.Ext) {
	copy(res, inputs[0])
}

// ProductSumGate evaluates
//
//	Gate(x) = Σ_i Lambdas[i] · inputs[2i](x) · inputs[2i+1](x)
//
// where inputs = [A₀, B₀, A₁, B₁, …].  It has degree 2.
// Lambdas must be pre-computed by the caller (e.g. powers [1, λ, λ², …]).
type ProductSumGate struct {
	Lambdas []field.Element
}

// Degree returns 2.
func (g *ProductSumGate) Degree() int { return 2 }

// EvalBatch implements Gate. len(inputs) must equal 2·len(g.Lambdas).
func (g *ProductSumGate) EvalBatch(res []field.Ext, inputs ...[]field.Ext) {
	for j := range res {
		res[j].SetZero()
		for i, λ := range g.Lambdas {
			var prod, tmp field.Ext
			prod.Mul(&inputs[2*i][j], &inputs[2*i+1][j])
			tmp.MulByElement(&prod, &λ)
			res[j].Add(&res[j], &tmp)
		}
	}
}
