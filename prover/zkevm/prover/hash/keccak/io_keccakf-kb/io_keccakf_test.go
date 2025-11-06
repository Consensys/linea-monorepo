package iokeccakf_test

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
	iokeccakf "github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/io_keccakf-kb"
	"github.com/stretchr/testify/assert"
)

func TestIoKeccakF(t *testing.T) {
	var (
		io                = &iokeccakf.IOKeccakF{}
		numBlocks         = 4
		numRowsPerBlock   = generic.KeccakUsecase.NbOfLanesPerBlock() * iokeccakf.NbOfRowsPerLane
		laneEffectiveSize = numRowsPerBlock * numBlocks
		laneSize          = utils.NextPowerOfTwo(laneEffectiveSize)
		keccakfSize       = utils.NextPowerOfTwo(numBlocks * keccak.NumRound)
	)

	define := func(build *wizard.Builder) {

		//declare input columns here
		lane := build.CompiledIOP.InsertCommit(0, "LANE", laneSize)
		isBeginningOfNewHash := build.CompiledIOP.InsertCommit(0, "IS_BEGINNING_OF_NEW_HASH", laneSize)
		isLaneActive := build.CompiledIOP.InsertCommit(0, "IS_LANE_ACTIVE", laneSize)

		io = iokeccakf.NewIOKeccakF(build.CompiledIOP, iokeccakf.IOKeccakFInputs{
			Lane:                 lane,
			IsBeginningOfNewHash: isBeginningOfNewHash,
			IsLaneActive:         isLaneActive,
			KeccakfSize:          keccakfSize,
		})
	}

	prover := func(run *wizard.ProverRuntime) {
		var (
			lane        = common.NewVectorBuilder(io.Inputs.Lane)
			isBeginning = common.NewVectorBuilder(io.Inputs.IsBeginningOfNewHash)
			zeros       = vector.Zero(numRowsPerBlock - 1)
			isActive    = common.NewVectorBuilder(io.Inputs.IsLaneActive)
		)
		isBeginning.PushInt(1)
		isBeginning.PushSliceF(zeros)
		isBeginning.PushSliceF(zeros)
		isBeginning.PushInt(0)
		isBeginning.PushInt(1)
		isBeginning.PushSliceF(zeros)
		isBeginning.PushSliceF(zeros)
		isBeginning.PushInt(0)
		isBeginningFr := isBeginning.Slice()
		fmt.Printf("len(isBeginningFr) = %d\n", len(isBeginningFr))
		// assign values to input columns here
		for row := 0; row < laneEffectiveSize; row++ {
			f := field.NewElement(0x35ca) // bytes = [0xca, 0x35]
			lane.PushField(f)
			isActive.PushInt(1)
			if isBeginningFr[row].IsOne() {
				fmt.Printf("isBeginningOfNewHash at row %d is one\n", row)
			}
		}
		isBeginning.PadAndAssign(run, field.Zero())
		lane.PadAndAssign(run, field.Zero())
		isActive.PadAndAssign(run, field.Zero())

		// assign io module
		io.Run(run)
	}

	compiled := wizard.Compile(define, dummy.Compile)
	proof := wizard.Prove(compiled, prover)
	assert.NoErrorf(t, wizard.Verify(compiled, proof), "ioKeccakF-verifier failed")

}
