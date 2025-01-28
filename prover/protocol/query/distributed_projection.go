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
	ColumnA, ColumnB            *symbolic.Expression
	FilterA, FilterB            *symbolic.Expression
	IsAInModule, IsBInModule bool
}

type DistributedProjection struct {
	Round int
	ID    ifaces.QueryID
	Inp   DistributedProjectionInput
}

type DistributedProjectionParams struct {
	HornerVal field.Element
	EvalRand  field.Element
}

func NewDistributedProjection(round int, id ifaces.QueryID, inp DistributedProjectionInput) DistributedProjection {
	if err := inp.ColumnA.Validate(); err != nil {
		utils.Panic("ColumnA for the distributed projection query %v is not a valid expression", id)
	}
	if err := inp.ColumnB.Validate(); err != nil {
		utils.Panic("ColumnB for the distributed projection query %v is not a valid expression", id)
	}
	if err := inp.FilterA.Validate(); err != nil {
		utils.Panic("FilterA for the distributed projection query %v is not a valid expression", id)
	}
	if err := inp.FilterB.Validate(); err != nil {
		utils.Panic("FilterB for the distributed projection query %v is not a valid expression", id)
	}
	if !inp.IsAInModule && !inp.IsBInModule {
		utils.Panic("Invalid distributed projection query %v, both A and B are not in the module", id)
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
		params       = run.GetParams(dp.ID).(DistributedProjectionParams)
		actualHorner = field.One()
		colABoard    = dp.Inp.ColumnA.Board()
		colBBoard    = dp.Inp.ColumnB.Board()
		filterABorad = dp.Inp.FilterA.Board()
		filterBBoard = dp.Inp.FilterB.Board()
		colA         = column.EvalExprColumn(run, colABoard).IntoRegVecSaveAlloc()
		colB         = column.EvalExprColumn(run, colBBoard).IntoRegVecSaveAlloc()
		filterA      = column.EvalExprColumn(run, filterABorad).IntoRegVecSaveAlloc()
		filterB      = column.EvalExprColumn(run, filterBBoard).IntoRegVecSaveAlloc()
	)
	if dp.Inp.IsAInModule && !dp.Inp.IsBInModule {
		hornerA := poly.CmptHorner(colA, filterA, params.EvalRand)
		actualHorner = hornerA[0]
	} else if !dp.Inp.IsAInModule && dp.Inp.IsBInModule {
		hornerB := poly.CmptHorner(colB, filterB, params.EvalRand)
		actualHorner = hornerB[0]
		actualHorner.Inverse(&actualHorner)
	} else if dp.Inp.IsAInModule && dp.Inp.IsBInModule {
		hornerA := poly.CmptHorner(colA, filterA, params.EvalRand)
		hornerB := poly.CmptHorner(colB, filterB, params.EvalRand)
		actualHorner = hornerB[0]
		actualHorner.Inverse(&actualHorner)
		actualHorner.Mul(&actualHorner, &hornerA[0])
	} else {
		utils.Panic("Invalid distributed projection query %v", dp.ID)
	}

	if actualHorner != params.HornerVal {
		return fmt.Errorf("the distributed projection query %v is not satisfied, actualHorner = %v, param.HornerVal = %v", dp.ID, actualHorner, params.HornerVal)
	}

	return nil
}

func (dp DistributedProjection) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {
	panic("UNSUPPORTED : can't check an Projection query directly into the circuit")
}
