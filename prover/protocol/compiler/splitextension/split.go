package splitextension

import (
	"errors"
	"reflect"

	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	"github.com/sirupsen/logrus"
)

var (
	fextSplitTag        = "FEXT2BASE"
	errInconsistentEval = errors.New("unconsistent evaluation claim")
	errInconsistentX    = errors.New("unconsistent evaluation point")
)

// AssignUnivProverAction implements the [wizard.ProverAction] interface and is
// responsible for assigning the new univariate query.
type AssignUnivProverAction struct {
	Ctx SplitCtx
}

// AssignSplitColumnProverAction implements the [wizard.ProverAction] interface
// and is responsible for assigning the splitted columns.
type AssignSplitColumnProverAction struct {
	Round int
	Ctx   SplitCtx
}

type VerifierCtx struct {
	Ctx SplitCtx
}

// SplitCtx context for splitting columns
type SplitCtx struct {

	// QueryFext Univariate query for field ext columns
	QueryFext query.UnivariateEval

	// QueryBaseField Univariate query for base field columns. The first
	// polynomials showing up in the query are the splitted ones and they are
	// followed by those that were already on the base field.
	QueryBaseField query.UnivariateEval

	// ToSplitPolynomials polynomials to split
	// i-th column goes to 4*i, 4*i+1, etc
	ToSplitPolynomials []ifaces.Column

	// AlreadyOnBasePolynomials is the list of columns that were already on
	// the base field in the order in which they show up in QueryFext.
	AlreadyOnBasePolynomials []ifaces.Column

	// SplittedPolynomials splitted polynomials
	SplittedPolynomials []ifaces.Column
}

func CompileSplitExtToBase(comp *wizard.CompiledIOP) {

	logrus.Trace("started naturalization compiler")
	defer logrus.Trace("finished naturalization compiler")

	// Expectedly, there should be only one query to compile as this compiler
	// is meant to be run after the MPTS compiler. This variable will detect
	// if more than one queries have been compiled.
	var numFoundQueries int

	// The compilation process is applied separately for each query
	for roundID := 0; roundID < comp.NumRounds(); roundID++ {
		for _, qName := range comp.QueriesParams.AllKeysAt(roundID) {

			if comp.QueriesParams.IsIgnored(qName) {
				continue
			}

			q, ok := comp.QueriesParams.Data(qName).(query.UnivariateEval)
			if !ok {
				utils.Panic("query %v has type %v expected only univariate", qName, reflect.TypeOf(q))
			}

			numFoundQueries++
			if numFoundQueries > 1 {
				utils.Panic("expected only one query to compile")
			}

			comp.QueriesParams.MarkAsIgnored(qName)

			basefieldQName := ifaces.QueryIDf("%s_%s", qName, fextSplitTag)

			ctx := SplitCtx{
				QueryFext:           q,
				ToSplitPolynomials:  make([]ifaces.Column, 0, len(q.Pols)),
				SplittedPolynomials: make([]ifaces.Column, 0, 4*len(q.Pols)),
			}

			for i, pol := range q.Pols {
				if pol.IsComposite() {
					utils.Panic("column %d should be natural", i)
				}

				if pol.IsBase() {
					ctx.AlreadyOnBasePolynomials = append(ctx.AlreadyOnBasePolynomials, pol)
					continue
				}

				comp.Columns.MarkAsIgnored(pol.GetColID())
				ctx.ToSplitPolynomials = append(ctx.ToSplitPolynomials, pol)
				polRound := pol.Round()

				for j := 0; j < 4; j++ {
					splittedColName := ifaces.ColIDf("%s_%s_%d", pol.String(), fextSplitTag, 4*i+j)
					ctx.SplittedPolynomials = append(
						ctx.SplittedPolynomials,
						comp.InsertCommit(polRound, ifaces.ColID(splittedColName), pol.Size()))
				}
			}

			// toEval is constructed using append over an empty slice to ensure
			// that it will do a deep-copy. Otherwise, it could have side-effects
			// over [ctx.ToSplitPolynomials] potentially causing
			// complex-to-diagnose issues in the future if the code came to evolve.
			toEval := append([]ifaces.Column{}, ctx.SplittedPolynomials...)
			toEval = append(toEval, ctx.AlreadyOnBasePolynomials...)

			ctx.QueryBaseField = comp.InsertUnivariate(
				roundID,
				basefieldQName,
				toEval,
			)

			for r := 0; r <= roundID; r++ {
				comp.RegisterProverAction(r, &AssignSplitColumnProverAction{
					Ctx:   ctx,
					Round: r,
				})
			}

			comp.RegisterProverAction(roundID, &AssignUnivProverAction{
				Ctx: ctx,
			})

			comp.RegisterVerifierAction(roundID, &VerifierCtx{
				Ctx: ctx,
			})
		}
	}
}

// Run implements the [wizard.ProverAction] interface
func (a *AssignSplitColumnProverAction) Run(runtime *wizard.ProverRuntime) {

	ctx := a.Ctx

	parallel.Execute(len(ctx.ToSplitPolynomials), func(start, end int) {
		for i := start; i < end; i++ {
			pol := ctx.ToSplitPolynomials[i]
			if pol.Round() != a.Round {
				continue
			}

			cc := pol.GetColAssignment(runtime)
			sv := splitVector(cc)

			runtime.AssignColumn(ctx.SplittedPolynomials[4*i].GetColID(), sv[0])
			runtime.AssignColumn(ctx.SplittedPolynomials[4*i+1].GetColID(), sv[1])
			runtime.AssignColumn(ctx.SplittedPolynomials[4*i+2].GetColID(), sv[2])
			runtime.AssignColumn(ctx.SplittedPolynomials[4*i+3].GetColID(), sv[3])
		}
	})
}

// Run implements ProverAction interface
func (pctx *AssignUnivProverAction) Run(runtime *wizard.ProverRuntime) {

	var (
		ctx            = pctx.Ctx
		evalFextParams = runtime.GetUnivariateParams(ctx.QueryFext.Name())
		x              = evalFextParams.ExtX
		svToEval       = make([]smartvectors.SmartVector, 0, len(ctx.QueryBaseField.Pols))
	)

	// This loop evaluates and assigns the polynomials that have been split and
	// append their evaluation "y" the assignment to the evaluation on the new
	// query. The implementation relies on the fact that these polynomials are
	// positionned at the beginning of the list of evaluated polynomials in the
	// new query.
	for _, pol := range ctx.SplittedPolynomials {
		sv := pol.GetColAssignment(runtime)
		svToEval = append(svToEval, sv)
	}

	y := smartvectors.BatchEvaluateBasePolyLagrange(svToEval, x)

	// This loops collect the evaluation claims of the already-on-base polynomials
	// from the new query to append them to the claims on the new query. This
	// relies on the fact that they appear in the new query *after* the splitted
	// polynomials and in the same order as in the original query.
	for i, pol := range ctx.QueryFext.Pols {

		if !pol.IsBase() {
			continue
		}

		oldY := evalFextParams.ExtYs[i]
		y = append(y, oldY)
	}

	runtime.AssignUnivariateExt(ctx.QueryBaseField.QueryID, evalFextParams.ExtX, y...)
}

var (
	_one = field.One()
	// fieldExtensionBasis is the list of the field extension basis elements.
	// 1, u, v, uv. The limbs are decomposed according to this basis.
	fieldExtensionBasis = [4]fext.Element{
		{B0: extensions.E2{A0: _one}},
		{B0: extensions.E2{A1: _one}},
		{B1: extensions.E2{A0: _one}},
		{B1: extensions.E2{A1: _one}},
	}
)

func (vctx *VerifierCtx) Run(run wizard.Runtime) error {
	ctx := vctx.Ctx

	// checks that P(x) = P_0(x) + w*P_1(x) + w**2*P_2(x) + w**3*P_3(x)
	// where P is the polynomial to split, and the P_i are the splitted
	// polynomials, corrersponding to the imaginary parts of P
	var (
		evalFextParams      = run.GetUnivariateParams(ctx.QueryFext.QueryID)
		evalBaseFieldParams = run.GetUnivariateParams(ctx.QueryBaseField.QueryID)
		nbPolyToSplit       = len(evalFextParams.ExtYs)
	)

	if evalBaseFieldParams.ExtX != evalFextParams.ExtX {
		return errInconsistentX
	}

	for i := 0; i < nbPolyToSplit; i++ {

		var tmp, reconstructedEval fext.Element

		for j := 0; j < 4; j++ {
			tmp.Mul(&fieldExtensionBasis[j], &evalBaseFieldParams.ExtYs[4*i+j])
			reconstructedEval.Add(&reconstructedEval, &tmp)
		}

		if !evalFextParams.ExtYs[i].Equal(&reconstructedEval) {
			return errInconsistentEval
		}
	}

	return nil
}

func (vctx *VerifierCtx) RunGnark(api frontend.API, run wizard.GnarkRuntime) {
	panic("RunGnark is unimplemented for split extension")
}

func splitVector(sv smartvectors.SmartVector) [4]smartvectors.SmartVector {

	size := sv.Len()

	if smartvectors.IsBase(sv) {
		return [4]smartvectors.SmartVector{
			sv,
			smartvectors.NewConstant(field.Zero(), size),
			smartvectors.NewConstant(field.Zero(), size),
			smartvectors.NewConstant(field.Zero(), size),
		}
	}

	// if it's a constant (ext) we don't need to allocate.
	if _, ok := sv.(*smartvectors.ConstantExt); ok {
		elmt := sv.GetExt(0)
		r0 := smartvectors.NewConstant(elmt.B0.A0, size)
		r1 := smartvectors.NewConstant(elmt.B0.A1, size)
		r2 := smartvectors.NewConstant(elmt.B1.A0, size)
		r3 := smartvectors.NewConstant(elmt.B1.A1, size)
		return [4]smartvectors.SmartVector{r0, r1, r2, r3}
	}

	r0 := make([]field.Element, size)
	r1 := make([]field.Element, size)
	r2 := make([]field.Element, size)
	r3 := make([]field.Element, size)

	for i := 0; i < size; i++ {
		elmt := sv.GetExt(i)
		r0[i] = elmt.B0.A0
		r1[i] = elmt.B0.A1
		r2[i] = elmt.B1.A0
		r3[i] = elmt.B1.A1
	}
	return [4]smartvectors.SmartVector{
		smartvectors.NewRegular(r0),
		smartvectors.NewRegular(r1),
		smartvectors.NewRegular(r2),
		smartvectors.NewRegular(r3),
	}
}

// build prover & verifier actions

// prover action -> create struct containing all the context
// get the challenge X (run.GetUnivariateParams(<queryName>))
// evaluate the splitted cols to get y' & assign splitted columns (run.AssignColumn())
