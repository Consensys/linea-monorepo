package fastpoly

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/fft"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/gnarkutil"
)

// Evaluate a polynomial in lagrange basis on a gnark circuit
func InterpolateGnark(api frontend.API, poly []frontend.Variable, x frontend.Variable) frontend.Variable {

	if !utils.IsPowerOfTwo(len(poly)) {
		utils.Panic("only support powers of two but poly has length %v", len(poly))
	}

	// When the poly is of length 1 it means it is a constant polynomial and its
	// evaluation is trivial.
	if len(poly) == 1 {
		return poly[0]
	}

	n := len(poly)
	domain := fft.NewDomain(n)
	one := field.One()

	// Test that x is not a root of unity. In the other case, we would
	// have to divide by zero. In practice this constraint is not necessary
	// (because the division constraint would be non-satisfiable anyway)
	// But doing an explicit check clarifies the need.
	xN := gnarkutil.Exp(api, x, n)
	api.AssertIsDifferent(xN, 1)

	// Compute the term-wise summand of the interpolation formula.
	// This will allow the gnark solver to process the expensive
	// inverses in parallel.
	terms := make([]frontend.Variable, n)

	// omegaMinI carries the domain's inverse root of unity generator raised to
	// the power I in the following loop. It is initialized with omega**0 = 1.
	omegaI := frontend.Variable(1)

	for i := 0; i < n; i++ {

		if i > 0 {
			omegaI = api.Mul(omegaI, domain.GeneratorInv)
		}

		// If the current term is the constant zero, we continue without generating
		// constraints. As a result, 'terms' may contain nil elements. Therefore,
		// we will need need to remove them later
		if c, isC := api.Compiler().ConstantValue(poly[i]); isC && c.IsInt64() && c.Int64() == 0 {
			continue
		}

		xOmegaN := api.Mul(x, omegaI)
		terms[i] = api.Sub(xOmegaN, 1)
		// No point doing a batch inverse in a circuit
		terms[i] = api.Inverse(terms[i])
		terms[i] = api.Mul(terms[i], poly[i])
	}

	nonNilTerms := make([]frontend.Variable, 0, len(terms))
	for i := range terms {
		if terms[i] == nil {
			continue
		}

		nonNilTerms = append(nonNilTerms, terms[i])
	}

	// Then sum all the terms
	var res frontend.Variable

	switch {
	case len(nonNilTerms) == 0:
		res = 0
	case len(nonNilTerms) == 1:
		res = nonNilTerms[0]
	case len(nonNilTerms) == 2:
		res = api.Add(nonNilTerms[0], nonNilTerms[1])
	default:
		res = api.Add(nonNilTerms[0], nonNilTerms[1], nonNilTerms[2:]...)
	}

	/*
		Then multiply the res by a factor \frac{g^{1 - n}X^n -g}{n}
	*/
	factor := xN
	factor = api.Sub(factor, one)
	factor = api.Mul(factor, domain.CardinalityInv)
	res = api.Mul(res, factor)

	return res

}
