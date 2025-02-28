package query

import (
	"math/big"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

type DistributedProjectionInput struct {
	ColumnA, ColumnB                                               *symbolic.Expression
	FilterA, FilterB                                               *symbolic.Expression
	SizeA, SizeB                                                   int
	EvalCoin                                                       coin.Name
	IsAInModule, IsBInModule                                       bool
	CumulativeNumOnesPrevSegmentsA, CumulativeNumOnesPrevSegmentsB big.Int
	CurrNumOnesA, CurrNumOnesB                                     field.Element
}

// func (dpInp *DistributedProjectionInput) completeAssign(run *ifaces.Runtime) {
// 	dpInp.CumulativeNumOnesPrevSegments.Run(run)
// }

type DistributedProjection struct {
	Round int
	ID    ifaces.QueryID
	Inp   []*DistributedProjectionInput
}

type DistributedProjectionParams struct {
	ScaledHorner                         field.Element
	HashCumSumOnePrev, HashCumSumOneCurr field.Element
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
func NewDistributedProjectionParams(scaledHorner, hashCumSumOnePrev, hashCumSumOneCurr field.Element) DistributedProjectionParams {
	return DistributedProjectionParams{
		ScaledHorner:      scaledHorner,
		HashCumSumOnePrev: hashCumSumOnePrev,
		HashCumSumOneCurr: hashCumSumOneCurr}
}

// Name returns the unique identifier of the GrandProduct query.
func (dp DistributedProjection) Name() ifaces.QueryID {
	return dp.ID
}

// Updates a Fiat-Shamir state
func (dpp DistributedProjectionParams) UpdateFS(fs *fiatshamir.State) {
	fs.Update(dpp.ScaledHorner)
}

// Unimplemented
func (dp DistributedProjection) Check(run ifaces.Runtime) error {
	return nil
}

func (dp DistributedProjection) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {
	panic("UNSUPPORTED : can't check a Projection query directly into the circuit")
}
