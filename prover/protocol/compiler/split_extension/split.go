package splitextension

import (
	"errors"
	"fmt"
	"reflect"

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
	baseNameSplit       = "splitted_col_"
	errUnconsistentEval = errors.New("unconsistent evaluation")
	fextQuery           = ifaces.QueryID("fextUnivariate")
	basefieldQuery      = ifaces.QueryID("baseUnivariate")
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
	QueryFext ifaces.QueryID

	// QueryBaseField Univariate query for base field columns
	QueryBaseField ifaces.QueryID

	// InputPolynomials polynomials to split
	// i-th column goes to 4*i, 4*i+1, etc
	InputPolynomials []ifaces.Column

	// OutputPolynomials splitted polynomials
	OutputPolynomials []ifaces.Column
}

func CompileSplitExtToBase(comp *wizard.CompiledIOP) {

	logrus.Trace("started naturalization compiler")
	defer logrus.Trace("finished naturalization compiler")

	// The compilation process is applied separately for each query
	for roundID := 0; roundID < comp.NumRounds(); roundID++ {

		for _, qName := range comp.QueriesParams.AllKeysAt(roundID) {

			if comp.QueriesParams.IsIgnored(qName) {
				continue
			}

			q_ := comp.QueriesParams.Data(qName)
			// TODO add a check to verify that the univariate query corresponds to split ext to base query (see queryID in testcase)
			if _, ok := q_.(query.UnivariateEval); !ok {
				utils.Panic("query %v has type %v expected only univariate", qName, reflect.TypeOf(q_))
			}

			q := q_.(query.UnivariateEval)

			var ctx SplitCtx
			ctx.QueryFext = fextQuery
			ctx.QueryBaseField = basefieldQuery
			ctx.InputPolynomials = make([]ifaces.Column, len(q.Pols))
			ctx.OutputPolynomials = make([]ifaces.Column, 4*len(q.Pols))

			for i, pol := range q.Pols {
				if pol.IsComposite() {
					panic(fmt.Sprintf("column %d should be natural", i)) // TODO use utils package
				}
				// if pol.IsBase() {
				// 	continue
				// }
				// ctx.InputPolynomials = append(ctx.InputPolynomials, pol)
				ctx.InputPolynomials[i] = pol
			}
			var proverCtx ProverCtx
			proverCtx.Ctx = ctx

			// Skip if it was already compiled, else insert.
			if comp.QueriesParams.MarkAsIgnored(qName) {
				continue
			}

			for i := 0; i < len(ctx.InputPolynomials); i++ {
				for j := 0; j < 4; j++ {
					curName := fmt.Sprintf("%s_%d", baseNameSplit, 4*i+j)
					// ctx.OutputPolynomials = append(ctx.OutputPolynomials, comp.InsertCommit(roundID, ifaces.ColID(curName), ctx.InputPolynomials[i].Size()))
					ctx.OutputPolynomials[4*i+j] = comp.InsertCommit(roundID, ifaces.ColID(curName), ctx.InputPolynomials[i].Size())
				}
			}
			comp.InsertUnivariate(0, basefieldQuery, ctx.OutputPolynomials)

			comp.RegisterProverAction(0, &proverCtx)

			var vctx VerifierCtx
			vctx.Ctx = ctx
			comp.RegisterVerifierAction(0, &vctx)

		}
	}
}

// Run implements ProverAction interface
func (pctx *ProverCtx) Run(runtime *wizard.ProverRuntime) {

	ctx := pctx.Ctx

	// filter cols defined over fext and split them, ignore other columns
	evalFextParams := runtime.GetUnivariateParams(ctx.QueryFext)
	y := make([]fext.Element, 4*len(ctx.InputPolynomials))
	for i, pol := range ctx.InputPolynomials {
		cc := pol.GetColAssignment(runtime)
		sv := splitVector(cc)

		runtime.AssignColumn(ctx.OutputPolynomials[4*i].GetColID(), sv[0])
		runtime.AssignColumn(ctx.OutputPolynomials[4*i+1].GetColID(), sv[1])
		runtime.AssignColumn(ctx.OutputPolynomials[4*i+2].GetColID(), sv[2])
		runtime.AssignColumn(ctx.OutputPolynomials[4*i+3].GetColID(), sv[3])

		for j := 0; j < 4; j++ {
			y[4*i+j] = smartvectors.EvaluateLagrangeFullFext(sv[j], evalFextParams.X)
		}
	}

	runtime.AssignUnivariate(basefieldQuery, evalFextParams.X, y...)
}

func (vctx *VerifierCtx) Run(run wizard.Runtime) error {

	ctx := vctx.Ctx

	// checks that P(x) = P_0(x) + w*P_1(x) + w**2*P_2(x) + w**3*P_3(x)
	// where P is the polynomial to split, and the P_i are the splitted
	// polynomials, corrersponding to the imaginary parts of P
	evalFextParams := run.GetUnivariateParams(ctx.QueryFext)

	evalBaseFieldParams := run.GetUnivariateParams(ctx.QueryBaseField)

	nbPolyToSplit := len(evalFextParams.Ys)

	for i := 0; i < nbPolyToSplit; i++ {
		var purportedEval [4]field.Element
		purportedEval[0].Set(&evalFextParams.Ys[i].B0.A0)
		purportedEval[1].Set(&evalFextParams.Ys[i].B0.A1)
		purportedEval[2].Set(&evalFextParams.Ys[i].B1.A0)
		purportedEval[3].Set(&evalFextParams.Ys[i].B1.A1)
		for j := 0; j < 4; j++ {

			// TODO check that evalBaseFieldParams.Ys[4*i+j] is real
			if !evalBaseFieldParams.Ys[4*i+j].B0.A0.Equal(&purportedEval[j]) {
				return errUnconsistentEval
			}
		}
	}

	return nil

}

func (vctx *VerifierCtx) RunGnark(api frontend.API, run wizard.GnarkRuntime) {
	panic("todo")
}

func printVector(v []field.Element) {
	for i := 0; i < len(v); i++ {
		fmt.Printf("%s\n", v[i].String())
	}
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
