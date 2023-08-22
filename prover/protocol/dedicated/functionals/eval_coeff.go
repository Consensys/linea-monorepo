package functionals

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/column"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/query"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/gnark/frontend"
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
func CoeffEval(comp *wizard.CompiledIOP, name string, x coin.Info, pol ifaces.Column) *ifaces.Accessor {

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
	comp.InsertLocalOpening(
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

	return ifaces.NewAccessor(
		fmt.Sprintf("EVAL_UNIVARIATE_RESULT_%v", name),
		func(run ifaces.Runtime) field.Element {
			// This will panic if the accessor is called too soon
			// i.e. in a round too early
			params := run.GetParams(ifaces.QueryIDf("%v_%v", name, EVAL_COEFF_FIXED_POINT_BEGIN)).(query.LocalOpeningParams)
			return params.Y
		},
		func(api frontend.API, c ifaces.GnarkRuntime) frontend.Variable {
			params := c.GetParams(ifaces.QueryIDf("%v_%v", name, EVAL_COEFF_FIXED_POINT_BEGIN)).(query.GnarkLocalOpeningParams)
			return params.Y
		},
		maxRound,
	)

}
