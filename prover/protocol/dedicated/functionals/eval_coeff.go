package functionals

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

const (
	EVAL_COEFF_POLY                 string = "EVAL_COEFF_COEFF_EVAL"
	EVAL_COEFF_LOCAL_CONSTRAINT_END string = "EVAL_COEFF_LOCAL_CONSTRAINT_END"
	EVAL_COEFF_FIXED_POINT_BEGIN    string = "EVAL_COEFF_FIXED_POINT_BEGIN"
	EVAL_COEFF_GLOBAL               string = "EVAL_COEFF_GLOBAL"
)

// Create a dedicated wizard to perform an evaluation in coefficient basis.
// Returns an accessor for the value of the polynomial. Takes an expression
// as input.
func CoeffEval(comp *wizard.CompiledIOP, name string, x coin.Info, pol ifaces.Column) ifaces.Accessor {

	length := pol.Size()
	maxRound := utils.Max(x.Round, pol.Round())

	hornerPoly := comp.InsertCommit(
		maxRound,
		ifaces.ColIDf("%v_%v", name, EVAL_COEFF_POLY),
		length,
	)

	// (x * h[i+1]) + expr[i] == h[i]
	// This will be cancelled at the border already
	globalExpr := ifaces.ColumnAsVariable(column.Shift(hornerPoly, 1)).
		Mul(x.AsVariable()).
		Add(ifaces.ColumnAsVariable(pol)).
		Sub(ifaces.ColumnAsVariable(hornerPoly))

	comp.InsertGlobal(
		maxRound,
		ifaces.QueryIDf("%v_%v", name, EVAL_COEFF_GLOBAL),
		globalExpr,
	)

	// p[-1] = h[-1]
	comp.InsertLocal(
		maxRound,
		ifaces.QueryIDf("%v_%v", name, EVAL_COEFF_LOCAL_CONSTRAINT_END),
		ifaces.ColumnAsVariable(column.Shift(hornerPoly, -1)).
			Sub(ifaces.ColumnAsVariable(column.Shift(pol, -1))),
	)

	// The result is given by
	localOpening := comp.InsertLocalOpening(
		maxRound,
		ifaces.QueryIDf("%v_%v", name, EVAL_COEFF_FIXED_POINT_BEGIN),
		hornerPoly,
	)

	comp.RegisterProverAction(maxRound, &coeffEvalAssignProverAction{
		name:       name,
		x:          x,
		pol:        pol,
		hornerPoly: hornerPoly,
		length:     length,
	})

	return accessors.NewLocalOpeningAccessor(localOpening, maxRound)
}

// coeffEvalAssignProverAction assigns the coefficient evaluation columns.
// It implements the [wizard.ProverAction] interface.
type coeffEvalAssignProverAction struct {
	name       string
	x          coin.Info
	pol        ifaces.Column
	hornerPoly ifaces.Column
	length     int
}

// Run executes the assignment of the coefficient evaluation.
func (a *coeffEvalAssignProverAction) Run(assi *wizard.ProverRuntime) {
	// Get the value of the coin and of pol
	x := assi.GetRandomCoinField(a.x.Name)
	p := a.pol.GetColAssignment(assi)

	// Now needs to evaluate the Horner poly
	h := make([]field.Element, a.length)
	h[a.length-1] = p.Get(a.length - 1)

	for i := a.length - 2; i >= 0; i-- {
		pi := p.Get(i)
		h[i].Mul(&h[i+1], &x).Add(&h[i], &pi)
	}

	assi.AssignColumn(ifaces.ColIDf("%v_%v", a.name, EVAL_COEFF_POLY), smartvectors.NewRegular(h))
	assi.AssignLocalPoint(ifaces.QueryIDf("%v_%v", a.name, EVAL_COEFF_FIXED_POINT_BEGIN), h[0])
}
