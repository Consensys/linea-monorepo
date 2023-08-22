package functionals

import (
	"fmt"
	"math/big"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/column"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/query"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/variables"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/gnarkutil"
	"github.com/consensys/gnark/frontend"
)

const (
	EVAL_BIVARIATE_POLY                 string = "HORNER"
	EVAL_BIVARIATE_LOCAL_CONSTRAINT_END string = "LOCAL_CONSTRAINT_END"
	EVAL_BIVARIATE_FIXED_POINT_BEGIN    string = "FIXED_POINT_BEGIN"
	EVAL_BIVARIATE_GLOBAL               string = "GLOBAL"
)

/*
Create a dedicated wizard to perform an evaluation in coefficient form
of a bivariate polynomial in x and y of degree respectively k and l.

Example : p commits to [1, 2, 3, 4, 5, 6, 7, 8]
and we want to evaluate p for (X=2, nPow=4) (Y=3, nPow=2)

This will return an accessor to value worth

	1 + 2X + 3X^2 + 4X^3 + 5Y + 6XY + 7XY^2 + 8XY^3 = 376
*/
func EvalCoeffBivariate(
	comp *wizard.CompiledIOP, name string,
	pCom ifaces.Column,
	x, y *ifaces.Accessor,
	nPowX, nPowY int, // corresponds to the degrees + 1 in each variable respectively
) *ifaces.Accessor {

	// sanity-check the dimensions of the vector
	if nPowX*nPowY != pCom.Size() {
		utils.Panic("mismatch in the dimensions (%v * %v) != %v", nPowX, nPowY, pCom.Size())
	}

	length := pCom.Size()
	maxRound := utils.Max(
		x.Round,
		y.Round,
		pCom.Round(),
	)

	hCom := comp.InsertCommit(
		maxRound,
		ifaces.ColIDf("%v_%v", name, EVAL_BIVARIATE_POLY),
		length,
	)

	/*
		Prepare a "bank" with all the variable we will need to encode the following constraints

		Global : H[i] = P[i] + H[i+1] (x + Zl,n[i-1] * (yx^{−k+1} - x))
		Local Constraint : H[−1]=P[−1]
		Local Opening : H[0] (will contain the result)
	*/

	h := ifaces.ColumnAsVariable(hCom)
	p := ifaces.ColumnAsVariable(pCom)
	hNext := ifaces.ColumnAsVariable(column.Shift(hCom, 1))
	zPrev := variables.NewPeriodicSample(nPowX, nPowX-1)
	x_v := x.AsVariable()

	yx_pow_1mk_acc := MakeYXPow1MNAccessor(name, x, y, nPowX)
	yx_pow_1mk := yx_pow_1mk_acc.AsVariable()
	h_prev := ifaces.ColumnAsVariable(column.Shift(hCom, -1))
	p_prev := ifaces.ColumnAsVariable(column.Shift(pCom, -1))

	/*
		start with the global constraints
		factor := x + Zl,n[i-1] * (yx^{−k+1} - x)
	*/
	factor := yx_pow_1mk.Sub(x_v).Mul(zPrev).Add(x_v)

	globalExpr := hNext.Mul(factor).Add(p).Sub(h)
	comp.InsertGlobal(
		maxRound,
		ifaces.QueryIDf("%v_%v", name, EVAL_BIVARIATE_GLOBAL),
		globalExpr,
	)

	// Local constraint : p[-1] = h[-1]
	comp.InsertLocal(
		maxRound,
		ifaces.QueryIDf("%v_%v", name, EVAL_BIVARIATE_LOCAL_CONSTRAINT_END),
		p_prev.Sub(h_prev),
	)

	// Local opening containing the result
	comp.InsertLocalOpening(
		maxRound,
		ifaces.QueryIDf("%v_%v", name, EVAL_BIVARIATE_FIXED_POINT_BEGIN),
		hCom,
	)

	comp.SubProvers.AppendToInner(maxRound, func(assi *wizard.ProverRuntime) {

		// Get the value of the coin and of pol
		x, yx_pow_1mk := x.GetVal(assi), yx_pow_1mk_acc.GetVal(assi)
		p := pCom.GetColAssignment(assi)

		// Now needs to evaluate the Horner poly
		h := make([]field.Element, length)
		h[length-1] = p.Get(length - 1)

		for i := length - 2; i >= 0; i-- {
			pi := p.Get(i)

			// Transition to a new "power of y"
			if (i+1)%nPowX == 0 {
				h[i].Mul(&h[i+1], &yx_pow_1mk).Add(&h[i], &pi)
				continue
			}

			h[i].Mul(&h[i+1], &x).Add(&h[i], &pi)
		}

		assi.AssignColumn(ifaces.ColIDf("%v_%v", name, EVAL_BIVARIATE_POLY), smartvectors.NewRegular(h))
		assi.AssignLocalPoint(ifaces.QueryIDf("%v_%v", name, EVAL_BIVARIATE_FIXED_POINT_BEGIN), h[0])
	})

	return ifaces.NewAccessor(
		fmt.Sprintf("EVAL_BIVARIATE_RES_%v", name),
		func(run ifaces.Runtime) field.Element {
			// This will panic if the accessor is called too soon
			// i.e. in a round too early
			params := run.GetParams(
				ifaces.QueryIDf("%v_%v", name, EVAL_BIVARIATE_FIXED_POINT_BEGIN),
			).(query.LocalOpeningParams)
			return params.Y
		},
		func(api frontend.API, run ifaces.GnarkRuntime) frontend.Variable {
			params := run.GetParams(
				ifaces.QueryIDf("%v_%v", name, EVAL_BIVARIATE_FIXED_POINT_BEGIN),
			).(query.GnarkLocalOpeningParams)
			return params.Y
		},
		maxRound,
	)

}

// returns an accessor for the intermediate variable yx^{1-n}
func MakeYXPow1MNAccessor(name string, x, y *ifaces.Accessor, n int) *ifaces.Accessor {
	return ifaces.NewAccessor(
		fmt.Sprintf("YX_POW_1_MIN_n_%v", name),
		func(run ifaces.Runtime) field.Element {
			x_ := x.GetVal(run)
			y_ := y.GetVal(run)
			var res field.Element
			res.Exp(x_, big.NewInt(int64(1-n)))
			res.Mul(&res, &y_)
			return res
		},
		func(api frontend.API, run ifaces.GnarkRuntime) frontend.Variable {
			x_ := x.GetFrontendVariable(api, run)
			y_ := y.GetFrontendVariable(api, run)
			res := gnarkutil.Exp(api, x_, 1-n)
			res = api.Mul(res, y_)
			return res
		},
		utils.Max(x.Round, y.Round),
	)
}
