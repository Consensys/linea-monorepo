package packing

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
)

// blockInput stores the inputs for [newBlock] function.
type blockInput struct {
	lanes laneRepacking
	param generic.HashingUsecase
}

type block struct {
	Inputs *blockInput
	// it is 1 iff the block is complete
	IsBlockComplete ifaces.Column
	// accumulator for the number of lanes
	accNumLane ifaces.Column
	// size of the submodule
	size int
	// the ProverAction for Iszero()
	pa wizard.ProverAction
}

// newBlock imposes the constraints  that the total sum of nBytes, given via imported.NBytes,
// indeed divides the blockSize.
func newBlock(comp *wizard.CompiledIOP, inp blockInput) block {

	var (
		name            = inp.lanes.Inputs.pckInp.Name
		size            = inp.lanes.Size
		createCol       = common.CreateColFn(comp, BLOCK+"_"+name, size)
		isLaneActive    = inp.lanes.IsLaneActive
		nbLanesPerBlock = inp.param.NbOfLanesPerBlock()
	)

	b := block{
		size:   inp.lanes.Size,
		Inputs: &inp,

		accNumLane: createCol("AccNumLane"),
	}

	b.IsBlockComplete, b.pa = dedicated.IsZero(comp, sym.Sub(b.accNumLane, nbLanesPerBlock))

	// constraints over accNumLanes (accumulate backward)
	// accNumLane[last] =isLaneActive[last]
	comp.InsertLocal(0, ifaces.QueryIDf(name+"_AccNumLane_Last"),
		sym.Sub(column.Shift(b.accNumLane, -1),
			column.Shift(isLaneActive, -1)),
	)

	// accNumLanes[i] = accNumLane[i+1]*(1-isBlockComplete[i+1]) + isLaneActive[i]
	res := sym.Sub(1, column.Shift(b.IsBlockComplete, 1)) // 1-isBlockComplete[i+1]

	expr :=
		sym.Sub(
			sym.Add(
				sym.Mul(
					column.Shift(b.accNumLane, 1), res),
				isLaneActive),
			b.accNumLane,
		)

	comp.InsertGlobal(0, ifaces.QueryIDf(name+"_AccNumLane_Glob"), expr)

	// isBlockComplete[0] = 1
	// NB: this guarantees that the total sum of  nybtes ,given via imported.Nbytes,
	// indeed divides the blockSize.
	// This fact can be used to guarantee that enough zeroes where padded during padding.
	comp.InsertLocal(
		0, ifaces.QueryIDf(name+"_IsBlockComplete"),
		sym.Mul(
			isLaneActive,
			sym.Sub(1, b.IsBlockComplete),
		),
	)

	// if isFirstLaneOfNewHash = 1 then isBlockComplete = 1.
	comp.InsertGlobal(0, ifaces.QueryIDf(name+"_EACH_HASH_HAS_COMPLETE_BLOCKS"),
		sym.Mul(
			inp.lanes.IsFirstLaneOfNewHash,
			sym.Sub(1, b.IsBlockComplete),
		),
	)

	return b
}

func (b *block) Assign(run *wizard.ProverRuntime) {

	var (
		size              = b.size
		accNumLane        = make([]field.Element, size)
		isActive          = b.Inputs.lanes.IsLaneActive.GetColAssignment(run).IntoRegVecSaveAlloc()
		nbOfLanesPerBlock = b.Inputs.param.NbOfLanesPerBlock()
	)
	accNumLane[size-1] = isActive[size-1]
	// accNumLanes[i] = accNumLane[i+1]*(1-isBlockComplete[i+1]) + isLaneActive[i]
	for row := size - 2; row >= 0; row-- {
		if field.ToInt(&accNumLane[row+1]) == nbOfLanesPerBlock {
			accNumLane[row] = field.One()
		} else {
			accNumLane[row].Add(&isActive[row], &accNumLane[row+1])
		}
	}
	run.AssignColumn(b.accNumLane.GetColID(),
		smartvectors.RightZeroPadded(accNumLane, size))

	b.pa.Run(run)
}

// it creates a blockInputs object.
func getBlockInputs(lane laneRepacking, param generic.HashingUsecase) blockInput {
	return blockInput{
		lanes: lane,
		param: param,
	}
}
