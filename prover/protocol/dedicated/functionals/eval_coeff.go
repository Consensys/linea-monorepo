package functionals

import (
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/accessors"
	"github.com/consensys/zkevm-monorepo/prover/protocol/coin"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils"
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

	comp.SubProvers.AppendToInner(maxRound, func(assi *wizard.ProverRuntime) {

		// Get the value of the coin and of pol
		x := assi.GetRandomCoinField(x.Name)
		p := pol.GetColAssignment(assi)

		// Now needs to evaluate the Horner poly
		h := make([]field.Element, length)
		h[length-1] = p.Get(length - 1)

		for i := length - 2; i >= 0; i-- {
			pi := p.Get(i)
			h[i].Mul(&h[i+1], &x).Add(&h[i], &pi)
		}

		assi.AssignColumn(ifaces.ColIDf("%v_%v", name, EVAL_COEFF_POLY), smartvectors.NewRegular(h))
		assi.AssignLocalPoint(ifaces.QueryIDf("%v_%v", name, EVAL_COEFF_FIXED_POINT_BEGIN), h[0])
	})

	return accessors.NewLocalOpeningAccessor(localOpening, maxRound)
}
