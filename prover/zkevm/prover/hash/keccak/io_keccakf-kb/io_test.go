package iokeccakf_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/pragmas"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
	iokeccakf "github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/io_keccakf-kb"
	"github.com/stretchr/testify/assert"
)

func TestBaseConversionInput(t *testing.T) {
	var (
		b              = &iokeccakf.KeccakfBlockPreparation{}
		numRowsPerLane = 4
		numBlocks      = 12
		keccak         = generic.KeccakUsecase
		size           = utils.NextPowerOfTwo(keccak.NbOfLanesPerBlock() * numRowsPerLane * numBlocks)
	)

	define := func(build *wizard.Builder) {
		var (
			comp      = build.CompiledIOP
			createCol = common.CreateColFn(comp, "BASE_CONVERSION_TEST", size, pragmas.RightPadded)
		)

		inp := iokeccakf.KeccakfBlockPreparationInputs{
			Lane:                 createCol("LANE"),
			IsBeginningOfNewHash: createCol("IS_FIRST_LANE_NEW_HASH"),
			IsLaneActive:         createCol("IS_ACTIVE"),
			BaseA:                3,
			BaseB:                3,
			NbBitsPerBaseX:       4,
		}

		b = iokeccakf.NewKeccakfBlockPreparation(comp, inp)

	}
	prover := func(run *wizard.ProverRuntime) {

		var (
			lane            = common.NewVectorBuilder(b.Inputs.Lane)
			isFirst         = common.NewVectorBuilder(b.Inputs.IsBeginningOfNewHash)
			isActive        = common.NewVectorBuilder(b.Inputs.IsLaneActive)
			unmRowsPerBlock = generic.KeccakUsecase.NbOfLanesPerBlock() * numRowsPerLane
			effectiveSize   = unmRowsPerBlock * numBlocks
		)

		for row := 0; row < effectiveSize; row++ {
			// input lanes are uint64 big-endian
			// choose 8 random bytes
			f := field.NewElement(0x35CA) // bytes = [0x35, 0xCA]
			lane.PushField(f)
			isActive.PushInt(1)
			if row%unmRowsPerBlock == 0 && (row/unmRowsPerBlock == 2) {
				isFirst.PushInt(1)
			} else {
				isFirst.PushInt(0)
			}
		}

		lane.PadAndAssign(run, field.Zero())
		isFirst.PadAndAssign(run, field.Zero())
		isActive.PadAndAssign(run, field.Zero())

		b.Run(run)

		expected := []uint64{10, 12, 13, 0} // see example comment below
		laneX := make([][]field.Element, len(b.LaneX))
		for i := range b.LaneX {
			laneX[i] = run.GetColumn(b.LaneX[i].GetColID()).IntoRegVecSaveAlloc()
		}
		//for row := 0; row < effectiveSize; row++ {
		for i := 0; i < len(b.LaneX); i++ {
			val := laneX[i][0]
			expectedVal := field.NewElement(expected[i])
			assert.Equalf(t, expectedVal, val, "invalid base conversion at laneX[%d] row %d", i, 0)
		}
		// }

	}

	comp := wizard.Compile(define, dummy.Compile)
	proof := wizard.Prove(comp, prover)
	assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")

}

// Example:
//
//   Suppose:
//     MAXNBYTE        = 2
//     nbBitsPerBaseX  = 4
//     len(b.LaneX)    = 4
//     b.Inputs.BaseA  = 3
//     b.Inputs.BaseB  = 5
//
//   And for a given lane element:
//     lane[j] = field.NewElement(0x35CA) // bytes = [0x35, 0xCA]
//
//   Steps:
//     - Big-endian bytes: [0x35, 0xCA] = [00110101, 11001010]
//     - After reversal (little-endian): [0xCA, 0x35]
//     - Bitstream (LSB-first):
//         0,1,0,1,0,0,1,1,1,0,1,1,0,0,0,0
//
//   Extract 4 chunks of 4 bits each (in base 3):
//     Chunk 0: 0101₂ = 10₁₀
//     Chunk 1: 0011₂ = 12₁₀
//     Chunk 2: 1011₂ = 13₁₀
//     Chunk 3: 0000₂ = 0₁₀
//
//   Therefore:
//     a = []uint64{10, 12, 13, 0}
//
//   These values are pushed into laneX[0..3]:
//     laneX[0].PushInt(10)
