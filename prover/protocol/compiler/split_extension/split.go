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

type ProverCtx struct {
	Ctx SplitCtx
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

				ctx.ToSplitPolynomials = append(ctx.ToSplitPolynomials, pol)
				for j := 0; j < 4; j++ {
					curName := ifaces.ColIDf("%s_%s_%d", pol.String(), fextSplitTag, 4*i+j)
					ctx.SplittedPolynomials = append(
						ctx.SplittedPolynomials,
						comp.InsertCommit(roundID, ifaces.ColID(curName), pol.Size()))
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

			comp.RegisterProverAction(roundID, &ProverCtx{
				Ctx: ctx,
			})

			comp.RegisterVerifierAction(roundID, &VerifierCtx{
				Ctx: ctx,
			})
		}
	}
}

// Run implements ProverAction interface
func (pctx *ProverCtx) Run(runtime *wizard.ProverRuntime) {

	var (
		ctx            = pctx.Ctx
		evalFextParams = runtime.GetUnivariateParams(ctx.QueryFext.Name())
		x              = evalFextParams.ExtX
		y              = make([]fext.Element, 0, len(ctx.QueryBaseField.Pols))
	)

	// This loop evaluates and assigns the polynomials that have been split and
	// append their evaluation "y" the assignment to the evaluation on the new
	// query. The implementation relies on the fact that these polynomials are
	// positionned at the beginning of the list of evaluated polynomials in the
	// new query.
	for i, pol := range ctx.ToSplitPolynomials {
		cc := pol.GetColAssignment(runtime)
		sv := splitVector(cc)

		runtime.AssignColumn(ctx.SplittedPolynomials[4*i].GetColID(), sv[0])
		runtime.AssignColumn(ctx.SplittedPolynomials[4*i+1].GetColID(), sv[1])
		runtime.AssignColumn(ctx.SplittedPolynomials[4*i+2].GetColID(), sv[2])
		runtime.AssignColumn(ctx.SplittedPolynomials[4*i+3].GetColID(), sv[3])

		for j := 0; j < 4; j++ {
			newY := smartvectors.EvaluateLagrangeFullFext(sv[j], x)
			y = append(y, newY)
		}
	}

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

		// TODO check that evalBaseFieldParams.Ys[4*i+j] is real
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
	var res [4]smartvectors.SmartVector
	size := sv.Len()
	var buf [4][]field.Element
	for i := 0; i < 4; i++ {
		buf[i] = make([]field.Element, size)
	}
	parallel.Execute(size, func(start, end int) {
		for i := start; i < end; i++ {
			elmt := sv.GetExt(i)
			buf[0][i].Set(&elmt.B0.A0)
			buf[1][i].Set(&elmt.B0.A1)
			buf[2][i].Set(&elmt.B1.A0)
			buf[3][i].Set(&elmt.B1.A1)
		}
	})

	res[0] = smartvectors.NewRegular(buf[0])
	res[1] = smartvectors.NewRegular(buf[1])
	res[2] = smartvectors.NewRegular(buf[2])
	res[3] = smartvectors.NewRegular(buf[3])
	return res
}

// build prover & verifier actions

// prover action -> create struct containing all the context
// get the challenge X (run.GetUnivariateParams(<queryName>))
// evaluate the splitted cols to get y' & assign splitted columns (run.AssignColumn())
