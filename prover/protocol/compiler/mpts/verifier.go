package mpts

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/fastpolyext"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// VerifierAction implements [wizard.VerifierAction]. It is tasked with
//   - (1) checking that the X value of the univariate query is correct.
//   - (2) checking that q(r) = sum_{i,k \in claim} [\lambda^i \rho^k (Pk(r) - y_{ik})] / (r - xi).
type VerifierAction struct {
	*MultipointToSinglepointCompilation
}

func (va VerifierAction) Run(run wizard.Runtime) error {

	var (
		queryParams = run.GetUnivariateParams(va.NewQuery.QueryID)
		// _polyOfRs stores the values of P_k(r) as returned in the query.
		// The last value of the slice is the value of Q(r) where q
		// is the quotient polynomial.
		//
		// However, the other values cannot be directly used by the verifier
		// and should instead use the ysMap.
		qr       = queryParams.ExtYs[len(va.NewQuery.Pols)-1]
		polysAtR = va.cptEvaluationMapExt(run)
		r        = queryParams.ExtX
		rCoin    = run.GetRandomCoinFieldExt(va.EvaluationPoint.Name)

		// zetasOfR stores the values zetas[i] = lambda^i / (r - xi).
		// These values are precomputed for efficiency.
		zetasOfR = make([]fext.Element, len(va.Queries))

		lambda = run.GetRandomCoinFieldExt(va.LinCombCoeffLambda.Name)
		rho    = run.GetRandomCoinFieldExt(va.LinCombCoeffRho.Name)
	)

	if r != rCoin {
		return fmt.Errorf("(verifier of %v) : Evaluation point of %v is incorrect (%v, expected %v)",
			va.NewQuery.QueryID, va.NewQuery.QueryID, r.String(), rCoin.String())
	}

	var (
		lambdaPowI = fext.One()
		rhoK       = fext.One()
		// res stores the right-hand of the equality check. Namely,
		// sum_{i,k \in claim} [\lambda^i \rho^k (Pk(r) - y_{ik})] / (r - xi).
		res = fext.Zero()
	)

	for i, q := range va.Queries {

		xi := run.GetUnivariateParams(q.Name()).ExtX
		zetasOfR[i].Sub(&r, &xi)
		// NB: this is very sub-optimal. We should use a batch-inverse instead
		// but the native verifier time is not very important in this context.
		zetasOfR[i].Inverse(&zetasOfR[i])
		zetasOfR[i].Mul(&zetasOfR[i], &lambdaPowI)
		lambdaPowI.Mul(&lambdaPowI, &lambda)
	}

	// This loop computes the value of [res]
	for k, p := range va.Polys {

		pr := polysAtR[p.GetColID()]
		for _, i := range va.EvalPointOfPolys[k] {
			// This sets tmp with the value of yik
			posOfYik := getPositionOfPolyInQueryYs(va.Queries[i], va.Polys[k])
			tmp := run.GetUnivariateParams(va.Queries[i].Name()).ExtYs[posOfYik]
			tmp.Sub(&pr, &tmp) // Pk(r) - y_{ik}
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

func (va VerifierAction) RunGnark(api frontend.API, run wizard.GnarkRuntime) {
	var (
		queryParams = run.GetUnivariateParams(va.NewQuery.QueryID)
		// polyOfRs stores the values of P_k(r) as returned in the query.
		// The last value of the slice is the value of Q(r) where q
		// is the quotient polynomial.
		qr       = queryParams.ExtYs[len(va.NewQuery.Pols)-1]
		polysAtR = va.cptEvaluationMapGnarkExt(api, run)
		r        = queryParams.ExtX
		rCoin    = run.GetRandomCoinFieldExt(va.EvaluationPoint.Name)

		// zetasOfR stores the values zetas[i] = lambda^i / (r - xi).
		// These values are precomputed for efficiency.
		zetasOfR = make([]koalagnark.Ext, len(va.Queries))

		lambda = run.GetRandomCoinFieldExt(va.LinCombCoeffLambda.Name)
		rho    = run.GetRandomCoinFieldExt(va.LinCombCoeffRho.Name)
	)

	koalaAPI := koalagnark.NewAPI(api)

	koalaAPI.AssertIsEqualExt(r, rCoin)

	lambdaPowI := koalaAPI.OneExt()
	rhoK := koalaAPI.OneExt()

	// res stores the right-hand of the equality check. Namely,
	// sum_{i,k \in claim} [\lambda^i \rho^k (Pk(r) - y_{ik})] / (r - xi).
	res := koalaAPI.ZeroExt()

	for i, q := range va.Queries {

		xi := run.GetUnivariateParams(q.Name()).ExtX
		// zetasOfR[i].Sub(api, r, xi)
		zetasOfR[i] = koalaAPI.SubExt(r, xi)

		// NB: this is very sub-optimal. We should use a batch-inverse instead
		// but the native verifier time is not very important in this context.
		// zetasOfR[i].Inverse(api, zetasOfR[i])
		// zetasOfR[i].Mul(api, zetasOfR[i], lambdaPowI)
		// lambdaPowI.Mul(api, lambdaPowI, lambda)
		zetasOfR[i] = koalaAPI.InverseExt(zetasOfR[i])
		zetasOfR[i] = koalaAPI.MulExt(zetasOfR[i], lambdaPowI)
		lambdaPowI = koalaAPI.MulExt(lambdaPowI, lambda)
	}

	// This loop computes the value of [res]
	for k, p := range va.Polys {
		pr := polysAtR[p.GetColID()]
		for _, i := range va.EvalPointOfPolys[k] {
			// This sets tmp with the value of yik
			posOfYik := getPositionOfPolyInQueryYs(va.Queries[i], va.Polys[k])
			tmp := run.GetUnivariateParams(va.Queries[i].Name()).ExtYs[posOfYik]
			tmp = koalaAPI.SubExt(pr, tmp)
			tmp = koalaAPI.MulExt(tmp, zetasOfR[i])
			tmp = koalaAPI.MulExt(tmp, rhoK)
			res = koalaAPI.AddExt(res, tmp)

		}

		// rhoK.Mul(api, rhoK, rho)
		rhoK = koalaAPI.MulExt(rhoK, rho)
	}

	koalaAPI.AssertIsEqualExt(res, qr)
}

// cptEvaluationMap returns an evaluation map [Column] -> [Y] for all the
// polynomials handled by [ctx]. This includes the columns of the new query
// but also the explictly evaluated columns.
func (ctx *MultipointToSinglepointCompilation) cptEvaluationMapExt(run wizard.Runtime) map[ifaces.ColID]fext.Element {

	var (
		evaluationMap = make(map[ifaces.ColID]fext.Element)
		univParams    = run.GetParams(ctx.NewQuery.QueryID).(query.UnivariateEvalParams)
		x             = univParams.ExtX
	)

	for i := range ctx.NewQuery.Pols {
		colID := ctx.NewQuery.Pols[i].GetColID()
		evaluationMap[colID] = univParams.ExtYs[i]
	}
	for i, c := range ctx.ExplicitlyEvaluated {
		colID := ctx.ExplicitlyEvaluated[i].GetColID()
		poly := c.GetColAssignment(run)

		evaluationMap[colID] = smartvectors.EvaluateFextPolyLagrange(poly, x)
	}

	return evaluationMap
}

// cptEvaluationMapGnark is the same as [cptEvaluationMap] but for a gnark circuit.
func (ctx *MultipointToSinglepointCompilation) cptEvaluationMapGnarkExt(api frontend.API, run wizard.GnarkRuntime) map[ifaces.ColID]koalagnark.Ext {

	var (
		evaluationMap = make(map[ifaces.ColID]koalagnark.Ext)
		univParams    = run.GetUnivariateParams(ctx.NewQuery.QueryID)
		x             = univParams.ExtX
		polys         = make([][]koalagnark.Ext, 0)
	)

	for i := range ctx.NewQuery.Pols {
		colID := ctx.NewQuery.Pols[i].GetColID()
		evaluationMap[colID] = univParams.ExtYs[i]
	}

	for _, c := range ctx.ExplicitlyEvaluated {
		poly := c.GetColAssignmentGnarkExt(run)
		polys = append(polys, poly)
	}

	ys := fastpolyext.BatchEvaluateLagrangeGnark(api, polys, x)

	for i := range ctx.ExplicitlyEvaluated {
		colID := ctx.ExplicitlyEvaluated[i].GetColID()
		evaluationMap[colID] = ys[i]
	}

	return evaluationMap
}

func getPositionOfPolyInQueryYs(q query.UnivariateEval, poly ifaces.Column) int {
	// TODO @gbotrel this appears on the traces quite a lot -- lot of string comparisons
	toFind := poly.GetColID()
	for i, p := range q.Pols {
		if p.GetColID() == toFind {
			return i
		}
	}

	utils.Panic("not found, poly=%v in query=%v", toFind, q.Name())
	return 0
}
