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
	Query                   query.DistributedProjection
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
	for index, inp := range va.Query.Inp {
		if va.IsA[index] && va.IsB[index] {
			va.CumNumOnesPrevSegmentsA[index] = inp.CumulativeNumOnesPrevSegmentsA
			va.NumOnesCurrSegmentA[index] = inp.CurrNumOnesA
			va.CumNumOnesPrevSegmentsB[index] = inp.CumulativeNumOnesPrevSegmentsB
			va.NumOnesCurrSegmentB[index] = inp.CurrNumOnesB
		} else if va.IsA[index] && !va.IsB[index] {
			va.CumNumOnesPrevSegmentsA[index] = inp.CumulativeNumOnesPrevSegmentsA
			va.NumOnesCurrSegmentA[index] = inp.CurrNumOnesA
		} else if !va.IsA[index] && va.IsB[index] {
			va.CumNumOnesPrevSegmentsB[index] = inp.CumulativeNumOnesPrevSegmentsB
			va.NumOnesCurrSegmentB[index] = inp.CurrNumOnesB
		}
	}
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
				multA, multB     field.Element
				hornerA, hornerB field.Element
			)
			hornerA = run.GetLocalPointEvalParams(va.HornerA0[index].ID).Y
			multA = run.GetRandomCoinField(va.EvalCoins[index].Name)
			multA.Exp(multA, &va.CumNumOnesPrevSegmentsA[index])
			hornerA.Mul(&hornerA, &multA)
			elemParam.Add(&elemParam, &hornerA)

			hornerB = run.GetLocalPointEvalParams(va.HornerB0[index].ID).Y
			multB = run.GetRandomCoinField(va.EvalCoins[index].Name)
			multB.Exp(multB, &va.CumNumOnesPrevSegmentsB[index])
			hornerB.Mul(&elemParam, &multB)
			elemParam.Sub(&elemParam, &hornerB)
		} else if va.IsA[index] && !va.IsB[index] {
			var (
				multA   field.Element
				hornerA field.Element
			)
			hornerA = run.GetLocalPointEvalParams(va.HornerA0[index].ID).Y
			multA = run.GetRandomCoinField(va.EvalCoins[index].Name)
			multA.Exp(multA, &va.CumNumOnesPrevSegmentsA[index])
			hornerA.Mul(&hornerA, &multA)
			elemParam.Add(&elemParam, &hornerA)
		} else if !va.IsA[index] && va.IsB[index] {
			var (
				multB   field.Element
				hornerB field.Element
			)
			hornerB = run.GetLocalPointEvalParams(va.HornerB0[index].ID).Y
			multB = run.GetRandomCoinField(va.EvalCoins[index].Name)
			multB.Exp(multB, &va.CumNumOnesPrevSegmentsB[index])
			hornerB.Mul(&hornerB, &multB)
			elemParam.Sub(&elemParam, &hornerB)
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
				return fmt.Errorf("number of one for filterA does not match, actual = %v, assigned = %v", numOnesA, va.NumOnesCurrSegmentA[index])
			}
			fB := column.EvalExprColumn(run, va.FilterB[index].Board()).IntoRegVecSaveAlloc()
			for i := 0; i < len(fB); i++ {
				if fB[i] == field.One() {
					numOnesB.Add(&numOnesB, &one)
				}
			}
			if numOnesB != va.NumOnesCurrSegmentB[index] {
				return fmt.Errorf("number of one for filterB does not match, actual = %v, assigned = %v", numOnesB, va.NumOnesCurrSegmentB[index])
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
				return fmt.Errorf("number of one for filterA does not match, actual = %v, assigned = %v", numOnesA, va.NumOnesCurrSegmentA[index])
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
				return fmt.Errorf("number of one for filterB does not match, actual = %v, assigned = %v", numOnesB, va.NumOnesCurrSegmentB[index])
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
				sumA, sumB field.Element
			)
			sumA = field.NewElement(va.CumNumOnesPrevSegmentsA[index].Uint64())
			sumA.Add(&sumA, &va.NumOnesCurrSegmentA[index])
			sumB = field.NewElement(va.CumNumOnesPrevSegmentsB[index].Uint64())
			sumB.Add(&sumB, &va.NumOnesCurrSegmentB[index])
			oldState = mimc.BlockCompression(oldState, sumA)
			oldState = mimc.BlockCompression(oldState, sumB)
		} else if va.IsA[index] && !va.IsB[index] {
			sumA := field.NewElement(va.CumNumOnesPrevSegmentsA[index].Uint64())
			sumA.Add(&sumA, &va.NumOnesCurrSegmentA[index])
			oldState = mimc.BlockCompression(oldState, sumA)
		} else if !va.IsA[index] && va.IsB[index] {
			sumB := field.NewElement(va.CumNumOnesPrevSegmentsB[index].Uint64())
			sumB.Add(&sumB, &va.NumOnesCurrSegmentB[index])
			oldState = mimc.BlockCompression(oldState, sumB)
		} else {
			panic("Invalid distributed projection query encountered during current hash verification")
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
}

// Skip implements the [wizard.VerifierAction]
func (va *distributedProjectionVerifierAction) Skip() {
	va.skipped = true
}

// IsSkipped implements the [wizard.VerifierAction]
func (va *distributedProjectionVerifierAction) IsSkipped() bool {
	return va.skipped
}
