package distributedprojection

import (
	"math/big"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/gnarkutil"
)

type distributedProjectionVerifierAction struct {
	Name                   ifaces.QueryID
	HornerA0, HornerB0     []query.LocalOpening
	isA, isB               []bool
	skipped                bool
	EvalCoins              []coin.Info
	cumNumOnesPrevSegments []big.Int
}

// Run implements the [wizard.VerifierAction]
func (va *distributedProjectionVerifierAction) Run(run wizard.Runtime) error {
	var (
		actualParam = field.Zero()
	)
	for index := range va.HornerA0 {
		var (
			elemParam = field.Zero()
		)
		if va.isA[index] && va.isB[index] {
			elemParam = run.GetLocalPointEvalParams(va.HornerB0[index].ID).Y
			elemParam.Neg(&elemParam)
			temp := run.GetLocalPointEvalParams(va.HornerA0[index].ID).Y
			elemParam.Add(&elemParam, &temp)
			// Multiply the horner value with the evalCoin^(CumulativeNumOnesPrevSegments)
			multiplier := run.GetRandomCoinField(va.EvalCoins[index].Name)
			multiplier.Exp(multiplier, &va.cumNumOnesPrevSegments[index])
			elemParam.Mul(&elemParam, &multiplier)
		} else if va.isA[index] && !va.isB[index] {
			elemParam = run.GetLocalPointEvalParams(va.HornerA0[index].ID).Y
			// Multiply the horner value with the evalCoin^(CumulativeNumOnesPrevSegments)
			multiplier := run.GetRandomCoinField(va.EvalCoins[index].Name)
			multiplier.Exp(multiplier, &va.cumNumOnesPrevSegments[index])
			elemParam.Mul(&elemParam, &multiplier)
		} else if !va.isA[index] && va.isB[index] {
			elemParam = run.GetLocalPointEvalParams(va.HornerB0[index].ID).Y
			elemParam.Neg(&elemParam)
			// Multiply the horner value with the evalCoin^(CumulativeNumOnesPrevSegments)
			multiplier := run.GetRandomCoinField(va.EvalCoins[index].Name)
			multiplier.Exp(multiplier, &va.cumNumOnesPrevSegments[index])
			elemParam.Mul(&elemParam, &multiplier)
		} else {
			utils.Panic("Unsupported verifier action registered for %v", va.Name)
		}
		actualParam.Add(&actualParam, &elemParam)
	}
	queryParam := run.GetDistributedProjectionParams(va.Name).HornerVal
	if actualParam != queryParam {
		utils.Panic("The distributed projection query %v did not pass, query param %v and actual param %v", va.Name, queryParam, actualParam)
	}
	return nil
}

// RunGnark implements the [wizard.VerifierAction] interface.
func (va *distributedProjectionVerifierAction) RunGnark(api frontend.API, run wizard.GnarkRuntime) {

	var (
		actualParam = frontend.Variable(0)
	)
	for index := range va.HornerA0 {
		var (
			elemParam, multiplier, evalCoin frontend.Variable
		)
		if va.isA[index] && va.isB[index] {
			var (
				a, b frontend.Variable
			)
			a = run.GetLocalPointEvalParams(va.HornerA0[index].ID).Y
			b = run.GetLocalPointEvalParams(va.HornerB0[index].ID).Y
			elemParam = api.Sub(a, b)
			// Multiply the horner value with the evalCoin^(CumulativeNumOnesPrevSegments)
			evalCoin = run.GetRandomCoinField(va.EvalCoins[index].Name)
			multiplier = gnarkutil.Exp(api, evalCoin, int(va.cumNumOnesPrevSegments[index].Int64()))
			elemParam = api.Mul(elemParam, multiplier)
		} else if va.isA[index] && !va.isB[index] {
			a := run.GetLocalPointEvalParams(va.HornerA0[index].ID).Y
			elemParam = api.Add(elemParam, a)
			// Multiply the horner value with the evalCoin^(CumulativeNumOnesPrevSegments)
			evalCoin = run.GetRandomCoinField(va.EvalCoins[index].Name)
			multiplier = gnarkutil.Exp(api, evalCoin, int(va.cumNumOnesPrevSegments[index].Int64()))
			elemParam = api.Mul(elemParam, multiplier)
		} else if !va.isA[index] && va.isB[index] {
			b := run.GetLocalPointEvalParams(va.HornerB0[index].ID).Y
			elemParam = api.Sub(elemParam, b)
			// Multiply the horner value with the evalCoin^(CumulativeNumOnesPrevSegments)
			evalCoin = run.GetRandomCoinField(va.EvalCoins[index].Name)
			multiplier = gnarkutil.Exp(api, evalCoin, int(va.cumNumOnesPrevSegments[index].Int64()))
			elemParam = api.Mul(elemParam, multiplier)
		} else {
			utils.Panic("Unsupported verifier action registered for %v", va.Name)
		}
		actualParam = api.Add(actualParam, elemParam)
	}
	queryParam := run.GetDistributedProjectionParams(va.Name).Sum

	api.AssertIsEqual(actualParam, queryParam)
}

// Skip implements the [wizard.VerifierAction]
func (va *distributedProjectionVerifierAction) Skip() {
	va.skipped = true
}

// IsSkipped implements the [wizard.VerifierAction]
func (va *distributedProjectionVerifierAction) IsSkipped() bool {
	return va.skipped
}
