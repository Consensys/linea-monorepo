package functionals

import (
	"fmt"
	"math/big"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/fft"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/column"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/query"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/symbolic"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/gnarkutil"
	"github.com/consensys/gnark/frontend"
)

const (
	INTERPOLATION_POLY          string = "INTERPOLATION_COEFF_EVAL"
	INTERPOLATION_LOCAL_BEGIN_0 string = "INTERPOLATION_LOCAL_BEGIN_0"
	INTERPOLATION_LOCAL_BEGIN_1 string = "INTERPOLATION_LOCAL_BEGIN_1"
	INTERPOLATION_OPEN_END      string = "INTERPOLATION_OPEN_END"
	INTERPOLATION_GLOBAL        string = "INTERPOLATION_GLOBAL"
)

// See the explainer here : https://hackmd.io/S78bJUa0Tk-T256iduE22g#Evaluate-in-Lagrange-form
// The variable names are the same as the one in the hackmd
func Interpolation(comp *wizard.CompiledIOP, name string, a *ifaces.Accessor, p ifaces.Column) *ifaces.Accessor {

	length := p.Size()
	maxRound := utils.Max(a.Round, p.Round())

	i := comp.InsertCommit(
		maxRound,
		ifaces.ColIDf("%v_%v", name, INTERPOLATION_POLY),
		length,
	)

	/*
		Common variables
	*/
	aV := a.AsVariable()
	pPrev := ifaces.ColumnAsVariable(column.Shift(p, -1))
	pV := ifaces.ColumnAsVariable(p)
	pNext := ifaces.ColumnAsVariable(column.Shift(p, 1))
	iPrevPrev := ifaces.ColumnAsVariable(column.Shift(i, -2))
	iPrev := ifaces.ColumnAsVariable(column.Shift(i, -1))
	iV := ifaces.ColumnAsVariable(i)
	iNext := ifaces.ColumnAsVariable(column.Shift(i, 1))
	one := symbolic.NewConstant(1)
	omega := symbolic.NewConstant(fft.GetOmega(p.Size()))
	omegaMin1 := omega.Sub(one)

	/*
		For the global constraint
			(ω−1)Δ[i]Δ[i−1] + ωΔ[i−1]P[i] − Δ[i]P[i−1] == 0
	*/
	delta := iV.Sub(iPrev)
	deltaPrev := iPrev.Sub(iPrevPrev)
	t1 := omegaMin1.Mul(delta).Mul(deltaPrev)
	t2 := omega.Mul(deltaPrev).Mul(ifaces.ColumnAsVariable(p))
	t3 := pPrev.Mul(delta)

	comp.InsertGlobal(
		maxRound,
		ifaces.QueryIDf("%v_%v", name, INTERPOLATION_GLOBAL),
		t1.Add(t2).Sub(t3),
	)

	/*
		For the local constraint at the beginning
			(a − 1)I[0] = P[0]
	*/

	comp.InsertLocal(
		maxRound,
		ifaces.QueryIDf("%v_%v", name, INTERPOLATION_LOCAL_BEGIN_0),
		aV.Sub(one).Mul(iV).Sub(pV),
	)

	/*
		For the the second local constraint
			(a − ω)(I[1] − I[0]) = ωP[1]
	*/

	comp.InsertLocal(
		maxRound,
		ifaces.QueryIDf("%v_%v", name, INTERPOLATION_LOCAL_BEGIN_1),
		aV.Sub(omega).Mul(iNext.Sub(iV)).Sub(pNext.Mul(omega)),
	)

	/*
		Local opening at the end
	*/

	comp.InsertLocalOpening(
		maxRound,
		ifaces.QueryIDf("%v_%v", name, INTERPOLATION_OPEN_END),
		column.Shift(i, -1),
	)

	comp.SubProvers.AppendToInner(maxRound, func(assi *wizard.ProverRuntime) {

		n := p.Size()
		a := a.GetVal(assi)
		one := field.One()
		p := p.GetColAssignment(assi)

		omegaInv := fft.GetOmega(n)
		omegaInv.Inverse(&omegaInv)

		// Compute the accumulator
		// witi will first contain the values of
		// omega^i / a - omega^i
		witi := make([]field.Element, n)
		witi[0] = a

		aRootOfUnityFlag := false

		for i := 1; i < n; i++ {
			witi[i].Mul(&witi[i-1], &omegaInv)
			witi[i-1].Sub(&witi[i-1], &one)
			if witi[i-1].IsZero() {
				aRootOfUnityFlag = true
			}
		}

		witi[n-1].Sub(&witi[n-1], &one)

		if witi[n-1].IsZero() || aRootOfUnityFlag {
			utils.Panic("detected that a is a root of unity")
		}

		witi = field.BatchInvert(witi)

		// Now we use it to compute the accumulation polyno
		for i := range witi {
			pi := p.Get(i)
			witi[i].Mul(&pi, &witi[i])
			if i > 0 {
				witi[i].Add(&witi[i], &witi[i-1])
			}
		}

		// Now we have the full witness of i
		assi.AssignColumn(ifaces.ColIDf("%v_%v", name, INTERPOLATION_POLY), smartvectors.NewRegular(witi))
		assi.AssignLocalPoint(ifaces.QueryIDf("%v_%v", name, INTERPOLATION_OPEN_END), witi[n-1])
	})

	// Returns an accessor to the result of the interpolation
	return ifaces.NewAccessor(
		fmt.Sprintf("INTERPOLATION_RES_%v", name),
		func(run ifaces.Runtime) field.Element {
			// Get the last point from the inner local opening and the random coin a
			end := run.GetParams(ifaces.QueryIDf("%v_%v", name, INTERPOLATION_OPEN_END)).(query.LocalOpeningParams).Y
			a := a.GetVal(run)
			// We return `(a^n - 1) end / n`
			one := field.One()
			nInv := field.NewElement(uint64(p.Size()))
			nInv.Inverse(&nInv)
			a.Exp(a, big.NewInt(int64(p.Size())))
			a.Sub(&a, &one)
			end.Mul(&end, &nInv)
			end.Mul(&end, &a)
			return end
		},
		func(api frontend.API, c ifaces.GnarkRuntime) frontend.Variable {
			// Get the last point from the inner local opening and the random coin a
			end := c.GetParams(ifaces.QueryIDf("%v_%v", name, INTERPOLATION_OPEN_END)).(query.GnarkLocalOpeningParams).Y
			a := a.GetFrontendVariable(api, c)
			// We return `(a^n - 1) end / n`
			one := field.One()
			nInv := field.NewElement(uint64(p.Size()))
			nInv.Inverse(&nInv)
			a = gnarkutil.Exp(api, a, p.Size())
			a = api.Sub(a, one)
			end = api.Mul(end, nInv)
			end = api.Mul(end, a)
			return end
		},
		maxRound,
	)

}
