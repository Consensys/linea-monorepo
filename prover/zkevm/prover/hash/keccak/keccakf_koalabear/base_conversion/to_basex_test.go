package baseconversion_test

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/pragmas"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	baseconversion "github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf_koalabear/base_conversion"
	"github.com/stretchr/testify/assert"
)

func TestKeccakFBlockPreparation(t *testing.T) {
	var (
		b    = &baseconversion.ToBaseX{}
		size = 8
	)

	define := func(build *wizard.Builder) {
		var (
			comp      = build.CompiledIOP
			createCol = common.CreateColFn(comp, "BASE_CONVERSION_TEST", size, pragmas.RightPadded)
		)

		inp := baseconversion.ToBaseXInputs{
			Lane:           createCol("LANE"),
			IsLaneActive:   createCol("IS_ACTIVE"),
			BaseX:          []int{3, 5, 7},
			NbBitsPerBaseX: 4,
			IsBaseX:        []ifaces.Column{},
		}
		for i := range inp.BaseX {
			inp.IsBaseX = append(inp.IsBaseX, createCol("IS_BASEX_%v", i))
		}

		b = baseconversion.NewToBaseX(comp, inp)

	}
	prover := func(run *wizard.ProverRuntime) {

		var (
			lane          = common.NewVectorBuilder(b.Inputs.Lane)
			isActive      = common.NewVectorBuilder(b.Inputs.IsLaneActive)
			effectiveSize = size - 3 // leave some padding at the end
			isBaseX       = make([]*common.VectorBuilder, len(b.Inputs.BaseX))
			isBaseXFr     = make([][]field.Element, len(b.Inputs.BaseX))
		)

		for i := range b.Inputs.BaseX {
			isBaseX[i] = common.NewVectorBuilder(b.Inputs.IsBaseX[i])
		}

		for row := 0; row < effectiveSize; row++ {
			f := field.NewElement(0x35ca) // bytes = [0xca, 0x35]
			lane.PushField(f)
			isActive.PushInt(1)
			if row%3 == 0 {
				isBaseX[0].PushInt(1)
				isBaseX[1].PushInt(0)
				isBaseX[2].PushInt(0)
			}
			if row%3 == 1 {
				isBaseX[0].PushInt(0)
				isBaseX[1].PushInt(1)
				isBaseX[2].PushInt(0)
			}
			if row%3 == 2 {
				isBaseX[0].PushInt(0)
				isBaseX[1].PushInt(0)
				isBaseX[2].PushInt(1)
			}

		}

		lane.PadAndAssign(run, field.Zero())
		isActive.PadAndAssign(run, field.Zero())
		for i := range b.Inputs.BaseX {
			isBaseX[i].PadAndAssign(run, field.Zero())
		}

		b.Run(run)

		expected0 := []uint64{30, 36, 10, 4} // see example in comment
		expected1 := []uint64{130, 150, 26, 6}
		expected3 := []uint64{350, 392, 50, 8}

		laneX := make([][]field.Element, len(b.LaneX))
		for i := range b.LaneX {
			laneX[i] = run.GetColumn(b.LaneX[i].GetColID()).IntoRegVecSaveAlloc()
		}
		for i := range b.Inputs.BaseX {
			isBaseXFr[i] = isBaseX[i].Slice()
		}

		for row := 0; row < size; row++ {
			for i := 0; i < len(b.LaneX); i++ {
				val := laneX[i][row]
				if row%3 == 0 && row < effectiveSize {
					assert.Equalf(t, fmt.Sprintf("%d", expected0[i]),
						fmt.Sprintf("%d", val.Uint64()), "invalid base conversion at laneX[%d] row %d is %d", i, 0, val.Uint64())
				}
				if row%3 == 1 && row < effectiveSize {
					assert.Equalf(t, fmt.Sprintf("%d", expected1[i]),
						fmt.Sprintf("%d", val.Uint64()), "invalid base conversion at laneX[%d] row %d is %d", i, 0, val.Uint64())
				}
				if row%3 == 2 && row < effectiveSize {
					assert.Equalf(t, fmt.Sprintf("%d", expected3[i]),
						fmt.Sprintf("%d", val.Uint64()), "invalid base conversion at laneX[%d] row %d is %d", i, 0, val.Uint64())
				}
				if row >= effectiveSize {
					assert.Equalf(t, "0",
						fmt.Sprintf("%d", val.Uint64()), "invalid base conversion at laneX[%d] row %d is %d", i, 0, val.Uint64())
				}
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
//
// and in base 5:
//	val[0] = 0*5^0 + 1*5^1 + 0*5^2 + 1*5^3 = 130
//	val[1] = 0*5^0 + 0*5^1 + 1*5^2 + 1*5^3 = 150
//	val[2] = 1*5^0 + 0*5^1 + 1*5^2 + 0*5^3 = 26
//	val[3] = 1*5^0 + 1*5^1 + 0*5^2 + 0*5^3 = 6
// So the function returns:
//	[]uint64{130, 130, 26, 6}
//
// and in base 7:
//	val[0] = 0*7^0 + 1*7^1 + 0*7^2 + 1*7^3 = 350
//	val[1] = 0*7^0 + 0*7^1 + 1*7^2 + 1*7^3 = 392
//	val[2] = 1*7^0 + 0*7^1 + 1*7^2 + 0*7^3 = 50
//	val[3] = 1*7^0 + 1*7^1 + 0*7^2 + 0*7^3 = 8
// So the function returns:
//	[]uint64{350, 392, 50, 8}
//
