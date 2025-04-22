package mpts

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// verifierAction implements [wizard.VerifierAction]. It is tasked with
//   - (1) checking that the X value of the univariate query is correct.
//   - (2) checking that q(r) = sum_{i,k \in claim} [\lambda^i \rho^k (Pk(r) - y_{ik})] / (r - xi).
type verifierAction struct {
	*MultipointToSinglepointCompilation
}

func (va verifierAction) Run(run wizard.Runtime) error {

	var (
		queryParams = run.GetUnivariateParams(va.NewQuery.QueryID)
		// polyOfRs stores the values of P_k(r) as returned in the query.
		// The last value of the slice is the value of Q(r) where q
		// is the quotient polynomial.
		polyOfRs = queryParams.Ys[:len(va.NewQuery.Pols)-1]
		qr       = queryParams.Ys[len(va.NewQuery.Pols)-1]
		r        = queryParams.X
		rCoin    = run.GetRandomCoinField(va.EvaluationPoint.Name)

		// zetasOfR stores the values zetas[i] = lambda^i / (r - xi).
		// These values are precomputed for efficiency.
		zetasOfR = make([]field.Element, len(va.Queries))

		lambda = run.GetRandomCoinField(va.LinCombCoeffLambda.Name)
		rho    = run.GetRandomCoinField(va.LinCombCoeffRho.Name)
	)

	if r != rCoin {
		return fmt.Errorf("(verifier of %v) : Evaluation point of %v is incorrect (%v, expected %v)",
			va.NewQuery.QueryID, va.NewQuery.QueryID, r.String(), rCoin.String())
	}

	var (
		lambdaPowI = field.One()
		rhoK       = field.One()
		// res stores the right-hand of the equality check. Namely,
		// sum_{i,k \in claim} [\lambda^i \rho^k (Pk(r) - y_{ik})] / (r - xi).
		res = field.Zero()
	)

	for i, q := range va.Queries {

		xi := run.GetUnivariateParams(q.Name()).X
		zetasOfR[i].Sub(&r, &xi)
		// NB: this is very sub-optimal. We should use a batch-inverse instead
		// but the native verifier time is not very important in this context.
		zetasOfR[i].Inverse(&zetasOfR[i])
		zetasOfR[i].Mul(&zetasOfR[i], &lambdaPowI)
		lambdaPowI.Mul(&lambdaPowI, &lambda)
	}

	// This loop computes the value of [res]
	for k := range va.Polys {
		for _, i := range va.EvalPointOfPolys[k] {
			// This sets tmp with the value of yik
			posOfYik := getPositionOfPolyInQueryYs(va.Queries[i], va.Polys[k])
			tmp := run.GetUnivariateParams(va.Queries[i].Name()).Ys[posOfYik]
			tmp.Sub(&polyOfRs[k], &tmp) // Pk(r) - y_{ik}
			tmp.Mul(&tmp, &zetasOfR[i])
			tmp.Mul(&tmp, &rhoK)
			res.Add(&res, &tmp)
		}

		rhoK.Mul(&rhoK, &rho)
	}

	if !res.Equal(&qr) {
		return fmt.Errorf("(verifier of %v) : Q(r) is incorrect (%v, expected %v)",
			va.NewQuery.QueryID, qr.String(), res.String())
	}

	return nil
}

func (va verifierAction) RunGnark(api frontend.API, run wizard.GnarkRuntime) {

	var (
		queryParams = run.GetUnivariateParams(va.NewQuery.QueryID)
		// polyOfRs stores the values of P_k(r) as returned in the query.
		// The last value of the slice is the value of Q(r) where q
		// is the quotient polynomial.
		polyOfRs = queryParams.Ys[:len(va.NewQuery.Pols)-1]
		qr       = queryParams.Ys[len(va.NewQuery.Pols)-1]
		r        = queryParams.X
		rCoin    = run.GetRandomCoinField(va.EvaluationPoint.Name)

		// zetasOfR stores the values zetas[i] = lambda^i / (r - xi).
		// These values are precomputed for efficiency.
		zetasOfR = make([]frontend.Variable, len(va.Queries))

		lambda = run.GetRandomCoinField(va.LinCombCoeffLambda.Name)
		rho    = run.GetRandomCoinField(va.LinCombCoeffRho.Name)
	)

	api.AssertIsEqual(r, rCoin)

	var (
		lambdaPowI = frontend.Variable(1)
		rhoK       = frontend.Variable(1)
		res        = frontend.Variable(0)
	)

	for i, q := range va.Queries {

		xi := run.GetUnivariateParams(q.Name()).X
		zetasOfR[i] = api.Sub(xi, r)
		// NB: this is very sub-optimal. We should use a batch-inverse instead
		// but the native verifier time is not very important in this context.
		zetasOfR[i] = api.Inverse(zetasOfR[i])
		lambdaPowI = api.Mul(lambdaPowI, lambda)
	}

	// res stores the right-hand of the equality check. Namely,
	// sum_{i,k \in claim} [\lambda^i \rho^k (Pk(r) - y_{ik})] / (r - xi).
	for k := range va.Polys {
		for _, i := range va.EvalPointOfPolys[k] {
			// This sets tmp with the value of yik
			posOfYik := getPositionOfPolyInQueryYs(va.Queries[i], va.Polys[k])
			tmp := run.GetUnivariateParams(va.Queries[i].Name()).Ys[posOfYik]
			tmp = api.Sub(polyOfRs[k], tmp) // Pk(r) - y_{ik}
			tmp = api.Mul(tmp, zetasOfR[i])
			tmp = api.Mul(tmp, rhoK)
			res = api.Add(res, tmp)
		}

		rhoK = api.Mul(rhoK, rho)
	}

	api.AssertIsEqual(res, qr)
}
