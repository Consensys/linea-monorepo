package distributedprojection

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

type distributedProjectionVerifierAction struct {
	Name                    ifaces.QueryID
	HornerA0, HornerB0      []query.LocalOpening
	IsA, IsB                []bool
	skipped                 bool
	EvalCoins               []coin.Info
	CumNumOnesPrevSegmentsA []big.Int
	CumNumOnesPrevSegmentsB []big.Int
	NumOnesCurrSegmentA     []field.Element
	NumOnesCurrSegmentB     []field.Element
	FilterA, FilterB        []*sym.Expression
}

// Run implements the [wizard.VerifierAction]
func (va *distributedProjectionVerifierAction) Run(run wizard.Runtime) error {
	errorCheckHorner := va.scaledHornerCheck(run)
	if errorCheckHorner != nil {
		return errorCheckHorner
	}
	errorCheckCurrNumOnes := va.currSumOneCheck(run)
	if errorCheckCurrNumOnes != nil {
		return errorCheckCurrNumOnes
	}

	errorCheckHash := va.hashCheck(run)
	if errorCheckHash != nil {
		return errorCheckHash
	}

	return nil
}

// method to check consistancy of the scaled horner
func (va *distributedProjectionVerifierAction) scaledHornerCheck(run wizard.Runtime) error {
	var (
		actualParam = field.Zero()
	)
	for index := range va.HornerA0 {
		var (
			elemParam = field.Zero()
		)
		if va.IsA[index] && va.IsB[index] {
			var (
				multA = field.One()
				multB = field.One()
			)
			elemParam = run.GetLocalPointEvalParams(va.HornerB0[index].ID).Y
			multB = run.GetRandomCoinField(va.EvalCoins[index].Name)
			multB.Exp(multB, &va.CumNumOnesPrevSegmentsB[index])
			elemParam.Mul(&elemParam, &multB)
			elemParam.Neg(&elemParam)
			temp := run.GetLocalPointEvalParams(va.HornerA0[index].ID).Y
			multA = run.GetRandomCoinField(va.EvalCoins[index].Name)
			multA.Exp(multA, &va.CumNumOnesPrevSegmentsA[index])
			temp.Mul(&temp, &multA)
			elemParam.Add(&elemParam, &temp)
		} else if va.IsA[index] && !va.IsB[index] {
			var (
				multA = field.One()
			)
			elemParam = run.GetLocalPointEvalParams(va.HornerA0[index].ID).Y
			multA = run.GetRandomCoinField(va.EvalCoins[index].Name)
			multA.Exp(multA, &va.CumNumOnesPrevSegmentsA[index])
			elemParam.Mul(&elemParam, &multA)
		} else if !va.IsA[index] && va.IsB[index] {
			var (
				multB = field.One()
			)
			elemParam = run.GetLocalPointEvalParams(va.HornerB0[index].ID).Y
			multB = run.GetRandomCoinField(va.EvalCoins[index].Name)
			multB.Exp(multB, &va.CumNumOnesPrevSegmentsB[index])
			elemParam.Mul(&elemParam, &multB)
		} else {
			utils.Panic("Unsupported verifier action registered for %v", va.Name)
		}
		actualParam.Add(&actualParam, &elemParam)
	}
	queryParam := run.GetDistributedProjectionParams(va.Name).ScaledHorner
	if actualParam != queryParam {
		utils.Panic("The distributed projection query %v did not pass, query param %v and actual param %v", va.Name, queryParam, actualParam)
	}
	return nil
}

func (va *distributedProjectionVerifierAction) currSumOneCheck(run wizard.Runtime) error {
	for index := range va.HornerA0 {
		if va.IsA[index] && va.IsB[index] {
			var (
				numOnesA = field.Zero()
				numOnesB = field.Zero()
				one      = field.One()
			)
			fA := column.EvalExprColumn(run, va.FilterA[index].Board()).IntoRegVecSaveAlloc()
			for i := 0; i < len(fA); i++ {
				if fA[i] == field.One() {
					numOnesA.Add(&numOnesA, &one)
				}
			}
			if numOnesA != va.NumOnesCurrSegmentA[index] {
				return fmt.Errorf("Number of one for filterA does not match, actual = %v, assigned = %v", numOnesA, va.NumOnesCurrSegmentA[index])
			}
			fB := column.EvalExprColumn(run, va.FilterB[index].Board()).IntoRegVecSaveAlloc()
			for i := 0; i < len(fB); i++ {
				if fB[i] == field.One() {
					numOnesB.Add(&numOnesB, &one)
				}
			}
			if numOnesB != va.NumOnesCurrSegmentB[index] {
				return fmt.Errorf("Number of one for filterB does not match, actual = %v, assigned = %v", numOnesB, va.NumOnesCurrSegmentB[index])
			}
		}
		if va.IsA[index] && !va.IsB[index] {
			var (
				numOnesA = field.Zero()
				one      = field.One()
			)
			fA := column.EvalExprColumn(run, va.FilterA[index].Board()).IntoRegVecSaveAlloc()
			for i := 0; i < len(fA); i++ {
				if fA[i] == field.One() {
					numOnesA.Add(&numOnesA, &one)
				}
			}
			if numOnesA != va.NumOnesCurrSegmentA[index] {
				return fmt.Errorf("Number of one for filterA does not match, actual = %v, assigned = %v", numOnesA, va.NumOnesCurrSegmentA[index])
			}
		}
		if !va.IsA[index] && va.IsB[index] {
			var (
				numOnesB = field.Zero()
				one      = field.One()
			)
			fB := column.EvalExprColumn(run, va.FilterB[index].Board()).IntoRegVecSaveAlloc()
			for i := 0; i < len(fB); i++ {
				if fB[i] == field.One() {
					numOnesB.Add(&numOnesB, &one)
				}
			}
			if numOnesB != va.NumOnesCurrSegmentB[index] {
				return fmt.Errorf("Number of one for filterB does not match, actual = %v, assigned = %v", numOnesB, va.NumOnesCurrSegmentB[index])
			}
		}
	}
	return nil
}

func (va *distributedProjectionVerifierAction) hashCheck(run wizard.Runtime) error {
	var (
		oldState = field.Zero()
	)
	for index := range va.HornerA0 {
		if va.IsA[index] && va.IsB[index] {
			var (
				sumA = field.Zero()
				sumB = field.Zero()
			)
			sumA = field.NewElement(va.CumNumOnesPrevSegmentsA[index].Uint64())
			sumA.Add(&sumA, &va.NumOnesCurrSegmentA[index])
			sumB = field.NewElement(va.CumNumOnesPrevSegmentsB[index].Uint64())
			sumB.Add(&sumB, &va.NumOnesCurrSegmentB[index])
			oldState = mimc.BlockCompression(oldState, sumA)
			oldState = mimc.BlockCompression(oldState, sumB)
		}
		if va.IsA[index] && !va.IsB[index] {
			var (
				sumA = field.Zero()
			)
			sumA = field.NewElement(va.CumNumOnesPrevSegmentsA[index].Uint64())
			sumA.Add(&sumA, &va.NumOnesCurrSegmentA[index])
			oldState = mimc.BlockCompression(oldState, sumA)
		}
		if !va.IsA[index] && va.IsB[index] {
			var (
				sumB = field.Zero()
			)
			sumB = field.NewElement(va.CumNumOnesPrevSegmentsB[index].Uint64())
			sumB.Add(&sumB, &va.NumOnesCurrSegmentB[index])
			oldState = mimc.BlockCompression(oldState, sumB)
		}
	}
	if oldState != run.GetDistributedProjectionParams(va.Name).HashCumSumOneCurr {
		return fmt.Errorf("HashCumSumOneCurr does not match, actual = %v, assigned = %v", oldState, run.GetDistributedProjectionParams(va.Name).HashCumSumOneCurr)
	}
	return nil
}

// RunGnark implements the [wizard.VerifierAction] interface.
func (va *distributedProjectionVerifierAction) RunGnark(api frontend.API, run wizard.GnarkRuntime) {

	panic("unimplemented")
	// var (
	// 	actualParam = frontend.Variable(0)
	// )
	// for index := range va.HornerA0 {
	// 	var (
	// 		elemParam, multiplier, evalCoin frontend.Variable
	// 	)
	// 	if va.IsA[index] && va.IsB[index] {
	// 		var (
	// 			a, b frontend.Variable
	// 		)
	// 		a = run.GetLocalPointEvalParams(va.HornerA0[index].ID).Y
	// 		b = run.GetLocalPointEvalParams(va.HornerB0[index].ID).Y
	// 		elemParam = api.Sub(a, b)
	// 		// Multiply the horner value with the evalCoin^(CumulativeNumOnesPrevSegments)
	// 		evalCoin = run.GetRandomCoinField(va.EvalCoins[index].Name)
	// 		multiplier = gnarkutil.Exp(api, evalCoin, int(va.CumNumOnesPrevSegments[index].Int64()))
	// 		elemParam = api.Mul(elemParam, multiplier)
	// 	} else if va.IsA[index] && !va.IsB[index] {
	// 		a := run.GetLocalPointEvalParams(va.HornerA0[index].ID).Y
	// 		elemParam = api.Add(elemParam, a)
	// 		// Multiply the horner value with the evalCoin^(CumulativeNumOnesPrevSegments)
	// 		evalCoin = run.GetRandomCoinField(va.EvalCoins[index].Name)
	// 		multiplier = gnarkutil.Exp(api, evalCoin, int(va.CumNumOnesPrevSegments[index].Int64()))
	// 		elemParam = api.Mul(elemParam, multiplier)
	// 	} else if !va.IsA[index] && va.IsB[index] {
	// 		b := run.GetLocalPointEvalParams(va.HornerB0[index].ID).Y
	// 		elemParam = api.Sub(elemParam, b)
	// 		// Multiply the horner value with the evalCoin^(CumulativeNumOnesPrevSegments)
	// 		evalCoin = run.GetRandomCoinField(va.EvalCoins[index].Name)
	// 		multiplier = gnarkutil.Exp(api, evalCoin, int(va.CumNumOnesPrevSegments[index].Int64()))
	// 		elemParam = api.Mul(elemParam, multiplier)
	// 	} else {
	// 		utils.Panic("Unsupported verifier action registered for %v", va.Name)
	// 	}
	// 	actualParam = api.Add(actualParam, elemParam)
	// }
	// queryParam := run.GetDistributedProjectionParams(va.Name).Sum

	// api.AssertIsEqual(actualParam, queryParam)
}

// Skip implements the [wizard.VerifierAction]
func (va *distributedProjectionVerifierAction) Skip() {
	va.skipped = true
}

// IsSkipped implements the [wizard.VerifierAction]
func (va *distributedProjectionVerifierAction) IsSkipped() bool {
	return va.skipped
}
