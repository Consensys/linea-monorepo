package query

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/maths/common/poly"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

type DistributedProjectionInput struct {
	ColumnA, ColumnB         *symbolic.Expression
	FilterA, FilterB         *symbolic.Expression
	IsAInModule, IsBInModule bool
}

type DistributedProjection struct {
	Round int
	ID    ifaces.QueryID
	Inp   []*DistributedProjectionInput
}

type DistributedProjectionParams struct {
	HornerVal field.Element
	EvalRand  field.Element
}

func NewDistributedProjection(round int, id ifaces.QueryID, inp []*DistributedProjectionInput) DistributedProjection {
	for _, in := range inp {
		if err := in.ColumnA.Validate(); err != nil {
			utils.Panic("ColumnA for the distributed projection query %v is not a valid expression", id)
		}
		if err := in.ColumnB.Validate(); err != nil {
			utils.Panic("ColumnB for the distributed projection query %v is not a valid expression", id)
		}
		if err := in.FilterA.Validate(); err != nil {
			utils.Panic("FilterA for the distributed projection query %v is not a valid expression", id)
		}
		if err := in.FilterB.Validate(); err != nil {
			utils.Panic("FilterB for the distributed projection query %v is not a valid expression", id)
		}
		if !in.IsAInModule && !in.IsBInModule {
			utils.Panic("Invalid distributed projection query %v, both A and B are not in the module", id)
		}
	}
	return DistributedProjection{Round: round, ID: id, Inp: inp}
}

// Constructor for distributed projection query parameters
func NewDistributedProjectionParams(hornerVal, evalRand field.Element) DistributedProjectionParams {
	return DistributedProjectionParams{HornerVal: hornerVal, EvalRand: evalRand}
}

// Name returns the unique identifier of the GrandProduct query.
func (dp DistributedProjection) Name() ifaces.QueryID {
	return dp.ID
}

// Updates a Fiat-Shamir state
func (dpp DistributedProjectionParams) UpdateFS(fs *fiatshamir.State) {
	fs.Update(dpp.HornerVal, dpp.EvalRand)
}

func (dp DistributedProjection) Check(run ifaces.Runtime) error {
	var (
		actualParam = field.One()
		params       = run.GetParams(dp.ID).(DistributedProjectionParams)
	)
	for _, inp := range dp.Inp {
		var (
			colABoard    = inp.ColumnA.Board()
			colBBoard    = inp.ColumnB.Board()
			filterABorad = inp.FilterA.Board()
			filterBBoard = inp.FilterB.Board()
			colA         = column.EvalExprColumn(run, colABoard).IntoRegVecSaveAlloc()
			colB         = column.EvalExprColumn(run, colBBoard).IntoRegVecSaveAlloc()
			filterA      = column.EvalExprColumn(run, filterABorad).IntoRegVecSaveAlloc()
			filterB      = column.EvalExprColumn(run, filterBBoard).IntoRegVecSaveAlloc()
			elemParam = field.One()
		)
		if inp.IsAInModule && !inp.IsBInModule {
			hornerA := poly.CmptHorner(colA, filterA, params.EvalRand)
			elemParam = hornerA[0]
		} else if !inp.IsAInModule && inp.IsBInModule {
			hornerB := poly.CmptHorner(colB, filterB, params.EvalRand)
			elemParam = hornerB[0]
			elemParam.Inverse(&elemParam)
		} else if inp.IsAInModule && inp.IsBInModule {
			hornerA := poly.CmptHorner(colA, filterA, params.EvalRand)
			hornerB := poly.CmptHorner(colB, filterB, params.EvalRand)
			elemParam = hornerB[0]
			elemParam.Inverse(&elemParam)
			elemParam.Mul(&elemParam, &hornerA[0])
		} else {
			utils.Panic("Invalid distributed projection query %v", dp.ID)
		}
		actualParam.Mul(&actualParam, &elemParam)

	}

	if actualParam != params.HornerVal {
		return fmt.Errorf("the distributed projection query %v is not satisfied, actualParam = %v, param.HornerVal = %v", dp.ID, actualParam, params.HornerVal)
	}

	return nil
}

func (dp DistributedProjection) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {
	panic("UNSUPPORTED : can't check an Projection query directly into the circuit")
}
