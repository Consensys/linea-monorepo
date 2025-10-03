package globalcs

import (
	"fmt"
	"math/big"
	"reflect"
	"sync"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark/frontend"
	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors_mixed"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/sirupsen/logrus"
)

// EvaluationCtx collects the compilation artefacts related to the evaluation
// part of the Plonk quotient technique.
type EvaluationCtx[T zk.Element] struct {
	QuotientCtx[T]
	QuotientEvals []query.UnivariateEval[T]
	WitnessEval   query.UnivariateEval[T]
	EvalCoin      coin.Info[T]
}

// EvaluationProver wraps [evaluationCtx] to implement the [wizard.ProverAction]
// interface.
type EvaluationProver[T zk.Element] EvaluationCtx[T]

// Run computes the evaluation of the univariate queries and implements the
// [wizard.ProverAction] interface.
func (pa EvaluationProver[T]) Run(run *wizard.ProverRuntime[T]) {

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

	ys := smartvectors_mixed.BatchEvaluateLagrange(witnesses, r)
	run.AssignUnivariateExt(pa.WitnessEval.QueryID, r, ys...)

	/*
		For the quotient evaluate it on `x = r / g`, where g is the coset
		shift. The generation of the domain is memoized.
	*/

	var (
		maxRatio          = utils.Max(pa.Ratios...)
		mulGenInv         = fft.NewDomain(uint64(maxRatio*pa.DomainSize), fft.WithCache()).FrMultiplicativeGenInv
		rootInv, _        = fft.Generator(uint64(maxRatio * pa.DomainSize))
		quotientEvalPoint fext.Element
		wg                = &sync.WaitGroup{}
	)
	rootInv.Inverse(&rootInv)
	quotientEvalPoint.MulByElement(&r, &mulGenInv)

	for i := range pa.QuotientEvals {
		wg.Add(1)
		go func(i int, evalPoint fext.Element) {
			var (
				q  = pa.QuotientEvals[i]
				cs = make([]sv.SmartVector, len(q.Pols))
			)

			parallel.Execute(len(q.Pols), func(start, stop int) {
				for i := start; i < stop; i++ {
					cs[i] = q.Pols[i].GetColAssignment(run)
				}
			})
			ys := smartvectors_mixed.BatchEvaluateLagrange(cs, evalPoint)

			run.AssignUnivariateExt(q.Name(), evalPoint, ys...)
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

// EvaluationVerifier wraps [evaluationCtx] to implement the [wizard.VerifierAction]
// interface.
type EvaluationVerifier[T zk.Element] struct {
	EvaluationCtx[T]
	Skipped bool
}

// declareUnivariateQueries declares the univariate queries over all the quotient
// shares, making sure that the shares needing to be evaluated over the same
// point are in the same query. This is a req from the upcoming naturalization
// compiler.
func declareUnivariateQueries[T zk.Element](
	comp *wizard.CompiledIOP[T],
	qCtx QuotientCtx[T],
) EvaluationCtx[T] {

	var (
		round       = qCtx.QuotientShares[0][0].Round()
		ratios      = qCtx.Ratios
		maxRatio    = utils.Max(ratios...)
		queriesPols = make([][]ifaces.Column[T], maxRatio)
		res         = EvaluationCtx[T]{
			QuotientCtx: qCtx,
			EvalCoin: comp.InsertCoin(
				round+1,
				coin.Name(deriveName(comp, EVALUATION_RANDOMESS)),
				coin.FieldExt,
			),
			WitnessEval: comp.InsertUnivariate(
				round+1,
				ifaces.QueryID(deriveName(comp, UNIVARIATE_EVAL_ALL_HANDLES)),
				qCtx.AllInvolvedColumns,
			),
			QuotientEvals: make([]query.UnivariateEval[T], maxRatio),
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
func (pa EvaluationCtx[T]) Run(run *wizard.ProverRuntime[T]) {

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

	ys := smartvectors_mixed.BatchEvaluateLagrange(witnesses, r)
	run.AssignUnivariateExt(pa.WitnessEval.QueryID, r, ys...)

	/*
		For the quotient evaluate it on `x = r / g`, where g is the coset
		shift. The generation of the domain is memoized.
	*/

	var (
		maxRatio          = utils.Max(pa.Ratios...)
		mulGenInv         = fft.NewDomain(uint64(maxRatio*pa.DomainSize), fft.WithCache()).FrMultiplicativeGenInv
		rootInv, _        = fft.Generator(uint64(maxRatio * pa.DomainSize))
		quotientEvalPoint fext.Element
		wg                = &sync.WaitGroup{}
	)
	rootInv.Inverse(&rootInv)
	quotientEvalPoint.MulByElement(&r, &mulGenInv)

	for i := range pa.QuotientEvals {
		wg.Add(1)
		go func(i int, evalPoint fext.Element) {
			var (
				q  = pa.QuotientEvals[i]
				cs = make([]sv.SmartVector, len(q.Pols))
			)

			parallel.Execute(len(q.Pols), func(start, stop int) {
				for i := start; i < stop; i++ {
					cs[i] = q.Pols[i].GetColAssignment(run)
				}
			})
			ys := smartvectors_mixed.BatchEvaluateLagrange(cs, evalPoint)

			run.AssignUnivariateExt(q.Name(), evalPoint, ys...)
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
func (ctx *EvaluationVerifier[T]) Run(run wizard.Runtime[T]) error {

	var (
		// Will be assigned to "X", the random point at which we check the constraint.
		r = run.GetRandomCoinFieldExt(ctx.EvalCoin.Name)
		// Map all the evaluations and checks the evaluations points
		mapYs = make(map[ifaces.ColID]fext.GenericFieldElem)
		// Get the parameters
		params           = run.GetUnivariateParams(ctx.WitnessEval.QueryID)
		univQuery        = run.GetUnivariateEval(ctx.WitnessEval.QueryID)
		quotientYs, errQ = ctx.recombineQuotientSharesEvaluation(run, r)
	)

	if errQ != nil {
		return fmt.Errorf("invalid evaluation point for the quotients: %v", errQ.Error())
	}

	// Check the evaluation point is consistent with r
	if params.ExtX != r {
		return fmt.Errorf("(verifier of global queries) : Evaluation point of %v is incorrect (%v, expected %v)",
			ctx.WitnessEval.QueryID, params.X.String(), r.String())
	}

	// Collect the evaluation points
	for j, handle := range univQuery.Pols {
		var genericElem fext.GenericFieldElem
		if params.IsBase {
			genericElem = fext.NewESHashFromBase(params.Ys[j])
		} else {
			genericElem = fext.NewESHashFromExt(params.ExtYs[j])
		}
		mapYs[handle.GetColID()] = genericElem
	}

	// Annulator = r^n - 1, common for all ratios
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
			case ifaces.Column[T]:
				entry := mapYs[metadata.GetColID()]
				if entry.IsBase {
					elem, _ := entry.GetBase()
					evalInputs[k] = sv.NewConstant(elem, 1)
				} else {
					elem := entry.GetExt()
					evalInputs[k] = sv.NewConstantExt(elem, 1)
				}
			case coin.Info[T]:
				evalInputs[k] = sv.NewConstantExt(run.GetRandomCoinFieldExt(metadata.Name), 1)
			case variables.X[T]:
				evalInputs[k] = sv.NewConstantExt(r, 1)
			case variables.PeriodicSample[T]:
				evalInputs[k] = sv.NewConstantExt(metadata.EvalAtOutOfDomainExt(ctx.DomainSize, r), 1)
			case ifaces.Accessor[T]:
				evalInputs[k] = sv.NewConstantExt(metadata.GetValExt(run), 1)
			default:
				utils.Panic("Not a variable type %v in global query (ratio %v)", reflect.TypeOf(metadataInterface), ratio)
			}
		}

		left := board.Evaluate(evalInputs).GetExt(0)

		// right : r^{n}-1 Q(r)
		qr := quotientYs[i]
		var right fext.Element
		right.Mul(&annulator, &qr)

		if left != right {
			return fmt.Errorf("global constraint - ratio %v - mismatch at random point - %v != %v", ratio, left.String(), right.String())
		}
	}

	return nil
}

// Verifier step, evaluate the constraint and checks that
func (ctx *EvaluationVerifier[T]) RunGnark(api frontend.API, c wizard.GnarkRuntime[T]) {

	// Will be assigned to "X", the random point at which we check the constraint.

	e4Api, err := gnarkfext.NewExt4[T](api)
	if err != nil {
		panic(err)
	}
	r := e4Api.NewFromBase(2) // r := c.GetRandomCoinFieldExt(ctx.EvalCoin.Name) TODO @thomas fixme
	annulator := e4Api.Exp(r, big.NewInt(int64(ctx.DomainSize)))

	quotientYs := ctx.recombineQuotientSharesEvaluationGnark(api, c, *r)
	params := c.GetUnivariateParams(ctx.WitnessEval.QueryID)
	univQuery := c.GetUnivariateEval(ctx.WitnessEval.QueryID)

	// annulator = api.Sub(annulator, T(1))

	// Get the parameters
	api.AssertIsEqual(r, params.X) // check the evaluation is consistent with the other stuffs

	// Map all the evaluations and checks the evaluations points
	mapYs := make(map[ifaces.ColID]T)

	// Collect the evaluation points
	for j, handle := range univQuery.Pols {
		mapYs[handle.GetColID()] = params.Ys[j]
	}

	apiGen, err := zk.NewApi[T](api)
	if err != nil {
		panic(err)
	}

	for i, ratio := range ctx.Ratios {

		board := ctx.AggregateExpressionsBoard[i]
		metadatas := board.ListVariableMetadata()

		evalInputs := make([]T, len(metadatas))

		for k, metadataInterface := range metadatas {
			switch metadata := metadataInterface.(type) {
			case ifaces.Column[T]:
				evalInputs[k] = mapYs[metadata.GetColID()]
			case coin.Info[T]:
				if metadata.IsBase() {
					evalInputs[k] = c.GetRandomCoinField(metadata.Name)
				} else {
					// TODO @thomas fixme
					// evalInputs[k] = c.GetRandomCoinFieldExt(metadata.Name)
				}
			case variables.X[T]:
				evalInputs[k] = r.B0.A0 // TODO @thomas should be extension fixme
			case variables.PeriodicSample[T]:
				evalInputs[k] = metadata.GnarkEvalAtOutOfDomain(api, ctx.DomainSize, r.B0.A0) // TODO @thomas should be extension fixme
			case ifaces.Accessor[T]:
				evalInputs[k] = metadata.GetFrontendVariable(apiGen, c)
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
func (ctx EvaluationVerifier[T]) recombineQuotientSharesEvaluation(run wizard.Runtime[T], r fext.Element) ([]fext.Element, error) {

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
		mulGenInv = fft.NewDomain(uint64(maxRatio*ctx.DomainSize), fft.WithCache()).FrMultiplicativeGenInv
		// omegaN is a root of unity generating the domain of size `domainSize
		// * maxRatio`
		omegaN, _ = fft.Generator(uint64(ctx.DomainSize * maxRatio))
	)

	shiftedR.MulByElement(&r, &mulGenInv)

	for i, q := range ctx.QuotientEvals {
		params := run.GetUnivariateParams(q.Name())
		qYs[i] = params.ExtYs

		// Check that the provided value for x is the right one
		providedX := params.ExtX
		var expectedXinit field.Element
		var expectedX fext.Element

		expectedXinit.Inverse(&omegaN)
		expectedXinit.Exp(expectedXinit, big.NewInt(int64(i)))
		expectedX.MulByElement(&shiftedR, &expectedXinit)

		if !providedX.Equal(&expectedX) {
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
			m             = ctx.DomainSize
			n             = ctx.DomainSize * ratio
			omegaRatio, _ = fft.Generator(uint64(ratio))
			rPowM         fext.Element
			// outerFactor stores m/n*(r^n - 1)
			outerFactor   = shiftedR
			one           = fext.One()
			omegaRatioInv field.Element
			res           fext.Element
			ratioInvField = field.NewElement(uint64(ratio))
		)

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
func (ctx EvaluationVerifier[T]) recombineQuotientSharesEvaluationGnark(api frontend.API, run wizard.GnarkRuntime[T], r gnarkfext.E4Gen[T]) []gnarkfext.E4Gen[T] {

	var (
		// res stores the list of the recombined quotient evaluations for each
		// combination.
		recombinedYs = make([]gnarkfext.E4Gen[T], len(ctx.Ratios))
		// ys stores the values of the quotient shares ordered by ratio
		qYs      = make([][]gnarkfext.E4Gen[T], utils.Max(ctx.Ratios...))
		maxRatio = utils.Max(ctx.Ratios...)
		// shiftedR = r / g where g is the generator of the multiplicative group
		shiftedR gnarkfext.E4Gen[T]
		// mulGen is the generator of the multiplicative group
		mulGenInv = fft.NewDomain(uint64(maxRatio*ctx.DomainSize), fft.WithCache()).FrMultiplicativeGenInv
		// omegaN is a root of unity generating the domain of size `domainSize
		// * maxRatio`
		omegaN, _ = fft.Generator(uint64(ctx.DomainSize * maxRatio))
	)

	e4Api, err := gnarkfext.NewExt4[T](api)
	if err != nil {
		panic(err)
	}

	apiGen, err := zk.NewApi[T](api)
	if err != nil {
		panic(err)
	}

	// shiftedR.MulByFp(api, r, mulGenInv)
	shiftedR = *e4Api.MulByFp(&r, zk.ValueOf[T](mulGenInv))

	for i, q := range ctx.QuotientEvals {
		params := run.GetUnivariateParams(q.Name())
		qYs[i] = params.ExtYs

		// Check that the provided value for x is the right one
		providedX := params.X
		_expectedX := apiGen.Inverse(zk.ValueOf[T](omegaN))
		_expectedX = field.Exp[T](apiGen, *_expectedX, big.NewInt(int64(i)))
		expectedX := e4Api.MulByFp(&shiftedR, _expectedX)
		e4Api.AssertIsEqual(expectedX, gnarkfext.Lift(&providedX)) // TODO @thomas fixme
	}

	for i, ratio := range ctx.Ratios {
		var (
			jumpBy = maxRatio / ratio
			ys     = make([]gnarkfext.E4Gen[T], ratio)
		)

		for j := range ctx.QuotientShares[i] {
			ys[j] = qYs[j*jumpBy][0]
			qYs[j*jumpBy] = qYs[j*jumpBy][1:]
		}

		var (
			m             = ctx.DomainSize
			n             = ctx.DomainSize * ratio
			omegaRatio, _ = fft.Generator(uint64(ratio))
			// outerFactor stores m/n*(r^n - 1)
			omegaRatioInv field.Element
			res           gnarkfext.E4Gen[T]
			ratioInvField = field.NewElement(uint64(ratio))
		)
		res = *e4Api.NewFromBase(0)

		rPowM := e4Api.Exp(&shiftedR, big.NewInt(int64(m)))
		ratioInvField.Inverse(&ratioInvField)
		omegaRatioInv.Inverse(&omegaRatio)

		for k := range ys {

			// tmp stores ys[k] / ((r^m / omegaRatio^k) - 1)
			var omegaInvPowK field.Element
			omegaInvPowK.Exp(omegaRatioInv, big.NewInt(int64(k)))
			tmp := e4Api.MulByFp(rPowM, zk.ValueOf[T](omegaInvPowK))
			tmp = e4Api.Sub(tmp, e4Api.NewFromBase(1))
			tmp = e4Api.Div(&ys[k], tmp)
			res = *e4Api.Mul(&res, tmp)
		}

		outerFactor := e4Api.Exp(&shiftedR, big.NewInt(int64(n)))
		outerFactor = e4Api.Sub(outerFactor, e4Api.NewFromBase(1))
		outerFactor = e4Api.Mul(outerFactor, e4Api.NewFromBase(ratioInvField))
		res = *e4Api.Mul(&res, outerFactor)
		recombinedYs[i] = res
	}

	return recombinedYs
}

func (ctx *EvaluationVerifier[T]) Skip() {
	ctx.Skipped = true
}

func (ctx *EvaluationVerifier[T]) IsSkipped() bool {
	return ctx.Skipped
}
