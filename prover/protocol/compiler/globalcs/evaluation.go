package globalcs

import (
	"fmt"
	"math/big"
	"reflect"
	"sync"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark/frontend"
	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/gnarkutil"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/sirupsen/logrus"
)

// evaluationCtx collects the compilation artefacts related to the evaluation
// part of the Plonk quotient technique.
type evaluationCtx struct {
	quotientCtx
	QuotientEvals []query.UnivariateEval
	WitnessEval   query.UnivariateEval
	EvalCoin      coin.Info
}

// evaluationProver wraps [evaluationCtx] to implement the [wizard.ProverAction]
// interface.
type evaluationProver evaluationCtx

// evaluationVerifier wraps [evaluationCtx] to implement the [wizard.VerifierAction]
// interface.
type evaluationVerifier struct {
	evaluationCtx
	skipped bool
}

// declareUnivariateQueries declares the univariate queries over all the quotient
// shares, making sure that the shares needing to be evaluated over the same
// point are in the same query. This is a req from the upcoming naturalization
// compiler.
func declareUnivariateQueries(
	comp *wizard.CompiledIOP,
	qCtx quotientCtx,
) evaluationCtx {

	var (
		round       = qCtx.QuotientShares[0][0].Round()
		ratios      = qCtx.Ratios
		maxRatio    = utils.Max(ratios...)
		queriesPols = make([][]ifaces.Column, maxRatio)
		res         = evaluationCtx{
			quotientCtx: qCtx,
			EvalCoin: comp.InsertCoin(
				round+1,
				coin.Name(deriveName(comp, EVALUATION_RANDOMESS)),
				coin.Field,
			),
			WitnessEval: comp.InsertUnivariate(
				round+1,
				ifaces.QueryID(deriveName(comp, UNIVARIATE_EVAL_ALL_HANDLES)),
				qCtx.AllInvolvedColumns,
			),
			QuotientEvals: make([]query.UnivariateEval, maxRatio),
		}
	)

	for i, ratio := range ratios {
		var (
			jumpBy = maxRatio / ratio
		)
		for j := range qCtx.QuotientShares[i] {
			queriesPols[j*jumpBy] = append(queriesPols[j*jumpBy], qCtx.QuotientShares[i][j])
		}
	}

	for i := range queriesPols {
		res.QuotientEvals[i] = comp.InsertUnivariate(
			round+1,
			ifaces.QueryID(deriveName(comp, UNIVARIATE_EVAL_QUOTIENT_SHARES, i, maxRatio)),
			queriesPols[i],
		)
	}

	return res
}

// Run computes the evaluation of the univariate queries and implements the
// [wizard.ProverAction] interface.
func (pa evaluationProver) Run(run *wizard.ProverRuntime) {

	var (
		stoptimer = profiling.LogTimer("Evaluate the queries for the global constraints")
		r         = run.GetRandomCoinFieldExt(pa.EvalCoin.Name)
		witnesses = make([]sv.SmartVector, len(pa.AllInvolvedColumns))
	)

	// Compute the evaluations
	parallel.Execute(len(pa.AllInvolvedColumns), func(start, stop int) {
		for i := start; i < stop; i++ {
			handle := pa.AllInvolvedColumns[i]
			witness := handle.GetColAssignment(run)
			witnesses[i] = witness

			if witness.Len() == 0 {
				logrus.Errorf("found a witness of size zero: %v", handle.GetColID())
			}
		}
	})

	ys := sv.BatchEvaluateLagrangeOnFext(witnesses, r)
	run.AssignUnivariate(pa.WitnessEval.QueryID, r, ys...)

	/*
		For the quotient evaluate it on `x = r / g`, where g is the coset
		shift. The generation of the domain is memoized.
	*/

	var (
		maxRatio          = utils.Max(pa.Ratios...)
		mulGenInv         = fft.NewDomain(uint64(maxRatio * pa.DomainSize)).FrMultiplicativeGenInv
		rootInv           field.Element
		quotientEvalPoint fext.Element
		wg                = &sync.WaitGroup{}
	)

	rootInv, err := fft.Generator(uint64(maxRatio * pa.DomainSize))
	if err != nil {
		panic(err)
	}
	rootInv.Inverse(&rootInv)
	quotientEvalPoint.MulByElement(&r, &mulGenInv)

	for i := range pa.QuotientEvals {
		wg.Add(1)
		go func(i int, evalPoint fext.Element) {
			var (
				q  = pa.QuotientEvals[i]
				ys = make([]fext.Element, len(q.Pols))
			)

			parallel.Execute(len(q.Pols), func(start, stop int) {
				for i := start; i < stop; i++ {
					c := q.Pols[i].GetColAssignment(run)
					ys[i] = sv.EvaluateLagrangeOnFext(c, evalPoint)
				}
			})

			run.AssignUnivariate(q.Name(), evalPoint, ys...)
			wg.Done()
		}(i, quotientEvalPoint)
		quotientEvalPoint.MulByElement(&quotientEvalPoint, &rootInv)
	}

	wg.Wait()

	/*
		as we shifted the evaluation point. No need to do do coset evaluation
		here
	*/
	stoptimer()
}

// Run evaluate the constraint and checks that
func (ctx *evaluationVerifier) Run(run wizard.Runtime) error {

	var (
		// Will be assigned to "X", the random point at which we check the constraint.
		r = run.GetRandomCoinFieldExt(ctx.EvalCoin.Name)
		// Map all the evaluations and checks the evaluations points
		mapYs = make(map[ifaces.ColID]fext.Element)
		// Get the parameters
		params           = run.GetUnivariateParams(ctx.WitnessEval.QueryID)
		univQuery        = run.GetUnivariateEval(ctx.WitnessEval.QueryID)
		quotientYs, errQ = ctx.recombineQuotientSharesEvaluation(run, r)
	)

	if errQ != nil {
		return fmt.Errorf("invalid evaluation point for the quotients: %v", errQ.Error())
	}

	// Check the evaluation point is consistent with r
	if params.X != r {
		return fmt.Errorf("(verifier of global queries) : Evaluation point of %v is incorrect (%v, expected %v)",
			ctx.WitnessEval.QueryID, params.X.String(), r.String())
	}

	// Collect the evaluation points
	for j, handle := range univQuery.Pols {
		mapYs[handle.GetColID()] = params.Ys[j]
	}

	// Annulator = X^n - 1, common for all ratios
	one := fext.One()
	annulator := r
	annulator.Exp(annulator, big.NewInt(int64(ctx.DomainSize)))
	annulator.Sub(&annulator, &one)

	for i, ratio := range ctx.Ratios {

		board := ctx.AggregateExpressionsBoard[i]
		metadatas := board.ListVariableMetadata()

		evalInputs := make([]sv.SmartVector, len(metadatas))

		for k, metadataInterface := range metadatas {
			switch metadata := metadataInterface.(type) {
			case ifaces.Column:
				evalInputs[k] = sv.NewConstant(mapYs[metadata.GetColID()], 1)
			case coin.Info:
				evalInputs[k] = sv.NewConstant(run.GetRandomCoinFext(metadata.Name), 1)
			case variables.X:
				evalInputs[k] = sv.NewConstant(r, 1)
			case variables.PeriodicSample:
				evalInputs[k] = sv.NewConstant(metadata.EvalAtOutOfDomain(ctx.DomainSize, r), 1)
			case ifaces.Accessor:
				evalInputs[k] = sv.NewConstant(metadata.GetVal(run), 1)
			default:
				utils.Panic("Not a variable type %v in global query (ratio %v)", reflect.TypeOf(metadataInterface), ratio)
			}
		}

		left := board.Evaluate(evalInputs).Get(0)

		// right : r^{n}-1 Q(r)
		qr := quotientYs[i]
		var right field.Element
		right.Mul(&annulator, &qr)

		if left != right {
			return fmt.Errorf("global constraint - ratio %v - mismatch at random point - %v != %v", ratio, left.String(), right.String())
		}
	}

	return nil
}

// Verifier step, evaluate the constraint and checks that
func (ctx *evaluationVerifier) RunGnark(api frontend.API, c wizard.GnarkRuntime) {

	// Will be assigned to "X", the random point at which we check the constraint.
	r := c.GetRandomCoinFext(ctx.EvalCoin.Name)
	annulator := gnarkutil.Exp(api, r, ctx.DomainSize)
	quotientYs := ctx.recombineQuotientSharesEvaluationGnark(api, c, r)
	params := c.GetUnivariateParams(ctx.WitnessEval.QueryID)
	univQuery := c.GetUnivariateEval(ctx.WitnessEval.QueryID)

	annulator = api.Sub(annulator, frontend.Variable(1))

	// Get the parameters
	api.AssertIsEqual(r, params.X) // check the evaluation is consistent with the other stuffs

	// Map all the evaluations and checks the evaluations points
	mapYs := make(map[ifaces.ColID]frontend.Variable)

	// Collect the evaluation points
	for j, handle := range univQuery.Pols {
		mapYs[handle.GetColID()] = params.Ys[j]
	}

	for i, ratio := range ctx.Ratios {

		board := ctx.AggregateExpressionsBoard[i]
		metadatas := board.ListVariableMetadata()

		evalInputs := make([]frontend.Variable, len(metadatas))

		for k, metadataInterface := range metadatas {
			switch metadata := metadataInterface.(type) {
			case ifaces.Column:
				evalInputs[k] = mapYs[metadata.GetColID()]
			case coin.Info:
				evalInputs[k] = c.GetRandomCoinFext(metadata.Name)
			case variables.X:
				evalInputs[k] = r
			case variables.PeriodicSample:
				evalInputs[k] = metadata.GnarkEvalAtOutOfDomain(api, ctx.DomainSize, r)
			case ifaces.Accessor:
				evalInputs[k] = metadata.GetFrontendVariable(api, c)
			default:
				utils.Panic("Not a variable type %v in global query (ratio %v)", reflect.TypeOf(metadataInterface), ratio)
			}
		}

		left := board.GnarkEval(api, evalInputs)

		// right : r^{n}-1 Q(r)
		qr := quotientYs[i]
		right := api.Mul(annulator, qr)

		api.AssertIsEqual(left, right)
		logrus.Debugf("verifying global constraint : DONE")

	}
}

// recombineQuotientSharesEvaluation returns the evaluations of the quotients
// on point r
func (ctx evaluationVerifier) recombineQuotientSharesEvaluation(run wizard.Runtime, r fext.Element) ([]fext.Element, error) {

	var (
		// res stores the list of the recombined quotient evaluations for each
		// combination.
		recombinedYs = make([]fext.Element, len(ctx.Ratios))
		// ys stores the values of the quotient shares ordered by ratio
		qYs      = make([][]fext.Element, utils.Max(ctx.Ratios...))
		maxRatio = utils.Max(ctx.Ratios...)
		// shiftedR = r / g where g is the generator of the multiplicative group
		shiftedR fext.Element
		// mulGen is the generator of the multiplicative group
		mulGenInv = fft.NewDomain(uint64(maxRatio * ctx.DomainSize)).FrMultiplicativeGenInv
		// omegaN is a root of unity generating the domain of size `domainSize
		// * maxRatio`
		omegaN field.Element
	)

	omegaN, err := fft.Generator(uint64(ctx.DomainSize * maxRatio))
	if err != nil {
		panic(err)
	}

	shiftedR.MulByElement(&r, &mulGenInv)

	for i, q := range ctx.QuotientEvals {
		params := run.GetUnivariateParams(q.Name())
		qYs[i] = params.Ys

		// Check that the provided value for x is the right one
		providedX := params.X
		var expectedX field.Element
		expectedX.Inverse(&omegaN)
		expectedX.Exp(expectedX, big.NewInt(int64(i)))

		var expectedXExt fext.Element
		expectedXExt.MulByElement(&shiftedR, &expectedX)
		if providedX != expectedXExt {
			return nil, fmt.Errorf("bad X value")
		}
	}

	for i, ratio := range ctx.Ratios {
		var (
			jumpBy = maxRatio / ratio
			ys     = make([]fext.Element, ratio)
		)

		for j := range ctx.QuotientShares[i] {
			ys[j] = qYs[j*jumpBy][0]
			qYs[j*jumpBy] = qYs[j*jumpBy][1:]
		}

		var (
			m          = ctx.DomainSize
			n          = ctx.DomainSize * ratio
			omegaRatio field.Element
			rPowM      fext.Element
			// outerFactor stores m/n*(r^n - 1)
			outerFactor   = shiftedR
			one           = fext.One()
			omegaRatioInv field.Element
			res           fext.Element
			ratioInvField = field.NewElement(uint64(ratio))
		)
		omegaRatio, err := fft.Generator(uint64(ratio))
		if err != nil {
			panic(err)
		}

		rPowM.Exp(shiftedR, big.NewInt(int64(m)))
		ratioInvField.Inverse(&ratioInvField)
		omegaRatioInv.Inverse(&omegaRatio)

		for k := range ys {

			// tmp stores ys[k] / ((r^m / omegaRatio^k) - 1)
			var tmpinit field.Element
			var tmp fext.Element
			tmpinit.Exp(omegaRatioInv, big.NewInt(int64(k)))
			tmp.MulByElement(&rPowM, &tmpinit)
			tmp.Sub(&tmp, &one)
			tmp.Div(&ys[k], &tmp)

			res.Add(&res, &tmp)
		}

		outerFactor.Exp(shiftedR, big.NewInt(int64(n)))
		outerFactor.Sub(&outerFactor, &one)
		outerFactor.MulByElement(&outerFactor, &ratioInvField)
		res.Mul(&res, &outerFactor)
		recombinedYs[i] = res
	}

	return recombinedYs, nil
}

// recombineQuotientSharesEvaluation returns the evaluations of the quotients
// on point r
func (ctx evaluationVerifier) recombineQuotientSharesEvaluationGnark(api frontend.API, run wizard.GnarkRuntime, r frontend.Variable) []frontend.Variable {

	var (
		// res stores the list of the recombined quotient evaluations for each
		// combination.
		recombinedYs = make([]frontend.Variable, len(ctx.Ratios))
		// ys stores the values of the quotient shares ordered by ratio
		qYs      = make([][]frontend.Variable, utils.Max(ctx.Ratios...))
		maxRatio = utils.Max(ctx.Ratios...)
		// shiftedR = r / g where g is the generator of the multiplicative group
		shiftedR frontend.Variable
		// mulGen is the generator of the multiplicative group
		mulGenInv = fft.NewDomain(uint64(maxRatio * ctx.DomainSize)).FrMultiplicativeGenInv
		// omegaN is a root of unity generating the domain of size `domainSize
		// * maxRatio`
		omegaN field.Element
	)
	omegaN, err := fft.Generator(uint64(ctx.DomainSize * maxRatio))
	if err != nil {
		panic(err)
	}
	shiftedR = api.Mul(r, mulGenInv)

	for i, q := range ctx.QuotientEvals {
		params := run.GetUnivariateParams(q.Name())
		qYs[i] = params.Ys

		// Check that the provided value for x is the right one
		providedX := params.X
		var expectedX frontend.Variable
		expectedX = api.Inverse(omegaN)
		expectedX = gnarkutil.Exp(api, expectedX, i)
		expectedX = api.Mul(expectedX, shiftedR)
		api.AssertIsEqual(providedX, expectedX)
	}

	for i, ratio := range ctx.Ratios {
		var (
			jumpBy = maxRatio / ratio
			ys     = make([]frontend.Variable, ratio)
		)

		for j := range ctx.QuotientShares[i] {
			ys[j] = qYs[j*jumpBy][0]
			qYs[j*jumpBy] = qYs[j*jumpBy][1:]
		}

		var (
			m          = ctx.DomainSize
			n          = ctx.DomainSize * ratio
			omegaRatio field.Element
			// outerFactor stores m/n*(r^n - 1)
			one           = field.One()
			omegaRatioInv field.Element
			res           = frontend.Variable(0)
			ratioInvField = field.NewElement(uint64(ratio))
		)
		omegaRatio, err := fft.Generator(uint64(ratio))
		if err != nil {
			panic(err)
		}
		rPowM := gnarkutil.Exp(api, shiftedR, m)
		ratioInvField.Inverse(&ratioInvField)
		omegaRatioInv.Inverse(&omegaRatio)

		for k := range ys {

			// tmp stores ys[k] / ((r^m / omegaRatio^k) - 1)
			var omegaInvPowK field.Element
			omegaInvPowK.Exp(omegaRatioInv, big.NewInt(int64(k)))
			tmp := api.Mul(omegaInvPowK, rPowM)
			tmp = api.Sub(tmp, one)
			tmp = api.Div(ys[k], tmp)

			res = api.Add(res, tmp)
		}

		outerFactor := gnarkutil.Exp(api, shiftedR, n)
		outerFactor = api.Sub(outerFactor, one)
		outerFactor = api.Mul(outerFactor, ratioInvField)
		res = api.Mul(res, outerFactor)
		recombinedYs[i] = res
	}

	return recombinedYs
}

func (ctx *evaluationVerifier) Skip() {
	ctx.skipped = true
}

func (ctx *evaluationVerifier) IsSkipped() bool {
	return ctx.skipped
}
