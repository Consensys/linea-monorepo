package packing

import (
	"testing"

	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/vector"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/stretchr/testify/assert"
)

// It generates Define and Assign function of Packing module, for testing
func makeTestCaseBlockModule(uc generic.HashingUsecase) (
	define wizard.DefineFunc,
	prover wizard.ProverStep,
) {
	var (
		// max number of blocks that can be extracted from limbs
		// if the number of blocks passes the max, newPack() would panic.
		maxNumBlock = 103
		// if the blockSize is not consistent with PackingParam, newPack() would panic.
		nbOfLanesPerBlock = uc.BlockSizeBytes()
		size              = utils.NextPowerOfTwo(maxNumBlock * nbOfLanesPerBlock)
		effectiveSize     = maxNumBlock / 2 * nbOfLanesPerBlock
	)

	block := block{}
	var isActive ifaces.Column

	define = func(build *wizard.Builder) {
		comp := build.CompiledIOP

		// commit to isActive
		isActive = comp.InsertCommit(0, "IsActive", size)

		inp := blockInput{
			lanes: laneRepacking{
				IsLaneActive: isActive,
				Size:         size,
			},
			param: uc,
		}

		// constraints
		block = newBlock(comp, inp)

	}
	prover = func(run *wizard.ProverRuntime) {

		// assign isActive
		col := vector.Repeat(field.One(), effectiveSize)
		run.AssignColumn(isActive.GetColID(), smartvectors.RightZeroPadded(col, size))

		block.Assign(run)
	}
	return define, prover
}

func TestBlockModule(t *testing.T) {
	for _, uc := range testCases {
		t.Run(uc.Name, func(t *testing.T) {
			define, prover := makeTestCaseBlockModule(uc.UseCase)
			comp := wizard.Compile(define, dummy.Compile)
			proof := wizard.Prove(comp, prover)
			assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")
		},
		)
	}
}
