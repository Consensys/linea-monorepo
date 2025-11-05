package iokeccakf_test

import (
	"fmt"
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

func TestKeccakFBlockPreparation(t *testing.T) {
	var (
		b              = &iokeccakf.KeccakfInputPreparation{}
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

		inp := iokeccakf.KeccakfInputPreparationData{
			Lane:                 createCol("LANE"),
			IsBeginningOfNewHash: createCol("IS_FIRST_LANE_NEW_HASH"),
			IsLaneActive:         createCol("IS_ACTIVE"),
			BaseA:                3,
			BaseB:                3,
			NbBitsPerBaseX:       4,
		}

		b = iokeccakf.NewKeccakfInputPreparation(comp, inp)

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
			f := field.NewElement(0x35ca) // bytes = [0xca, 0x35]
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

		expected := []uint64{30, 36, 10, 4} // see example in comment
		laneX := make([][]field.Element, len(b.LaneX))
		for i := range b.LaneX {
			laneX[i] = run.GetColumn(b.LaneX[i].GetColID()).IntoRegVecSaveAlloc()
		}
		for row := 0; row < effectiveSize; row++ {
			for i := 0; i < len(b.LaneX); i++ {
				val := laneX[i][0]
				assert.Equalf(t, fmt.Sprintf("%d", expected[i]),
					fmt.Sprintf("%d", val.Uint64()), "invalid base conversion at laneX[%d] row %d is %d", i, 0, val.Uint64())
			}
		}

	}

	comp := wizard.Compile(define, dummy.Compile)
	proof := wizard.Prove(comp, prover)
	assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")

}

//  Example:
//  input: 0x35ca
//	inputBytes := []byte{0b11001010, 0b00110101}
//	bitsPerChunk := 4
//	nbSlices := 4
//	base := uint64(3)
//
//	vals := extractLittleEndianBaseX(data, bitsPerChunk, nbSlices, base)
//
// Combined bitstream (LSB-first):
//
//	[0 1 0 1 0 0 1 1 1 0 1 0 1 1 0 0]
//
// Grouped into 4-bit chunks (little-endian within each group):
//
//	chunk 0 → [0 1 0 1]
//	chunk 1 → [0 0 1 1]
//	chunk 2 → [1 0 1 0]
//	chunk 3 → [1 1 0 0]
//
// Interpretation in base 3 (each bit contributes bit_j * 3^j):
//
//	val[0] = 0*3^0 + 1*3^1 + 0*3^2 + 1*3^3 = 30
//	val[1] = 0*3^0 + 0*3^1 + 1*3^2 + 1*3^3 = 36
//	val[2] = 1*3^0 + 0*3^1 + 1*3^2 + 0*3^3 = 10
//	val[3] = 1*3^0 + 1*3^1 + 0*3^2 + 0*3^3 = 4
//
// So the function returns:
//
//	[]uint64{30, 36, 10, 4}
