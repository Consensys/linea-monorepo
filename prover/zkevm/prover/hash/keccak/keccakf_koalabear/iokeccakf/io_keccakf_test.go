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
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf_koalabear/iokeccakf"
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
		f                 field.Element
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

		// assign values to input columns here
		for row := 0; row < laneEffectiveSize; row++ {
			if (row/4)%2 == 0 {
				f = field.NewElement(0x35ca) // bytes = [0xca, 0x35]
			} else {
				f = field.NewElement(0x56ad) // bytes = [0xad, 0x56]
			}
			lane.PushField(f)
			isActive.PushInt(1)
		}
		isBeginning.PadAndAssign(run, field.Zero())
		lane.PadAndAssign(run, field.Zero())
		isActive.PadAndAssign(run, field.Zero())

		// assign io module
		io.Run(run)

		// verify output blocks here, see the example below
		var expected uint64
		for i := range io.Blocks() {
			for j := 0; j < iokeccakf.NumSlices; j++ {
				actualBlock := io.Blocks()[i][j].GetColAssignment(run).IntoRegVecSaveAlloc()
				for row := 0; row < len(actualBlock); row++ {

					switch {
					case (row == 0 || row == 48) && i%2 == 0:
						expected = []uint64{20548, 1297}[j%2]
					case (row == 0 || row == 48) && i%2 == 1:
						expected = []uint64{17489, 4372}[j%2]
					case (row == 23 || row == 71) && i%2 == 0:
						expected = []uint64{19649675, 1786334}[j%2]
					case (row == 23 || row == 71) && i%2 == 1:
						expected = []uint64{21260074, 175814}[j%2]
					default:
						expected = 0
					}

					assert.Equalf(t, fmt.Sprintf("%d", expected),
						fmt.Sprintf("%d", actualBlock[row].Uint64()),
						"invalid block value at block %d slice %d row %d ,value %d", i, j, row, actualBlock[row].Uint64())

				}
			}

		}

	}

	compiled := wizard.Compile(define, dummy.Compile)
	proof := wizard.Prove(compiled, prover)
	assert.NoErrorf(t, wizard.Verify(compiled, proof), "ioKeccakF-verifier failed")

}

//  Example:
//  input: 0x35ca
//	inputBytes := []byte{0b11001010, 0b00110101}
//	bitsPerChunk := 8
//	nbSlices := 2
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
//	chunk 0 → [0 1 0 1 0 0 1 1]
//	chunk 2 → [1 0 1 0  1 1 0 0]
//
//
// Interpretation in base 3 (each bit contributes bit_j * 3^j):
//
//	laneX[i][j] = 0*4^0 + 1*4^1 + 0*4^2 + 1*4^3 + 0*4^4 + 0*4^5 + 1*4^6 + 1*4^7 = 20548
//	laneX[i][j] = 1*4^0 + 0*4^1 + 1*4^2 + 0*4^3 + 1*4^4 + 1*4^5 + 0*4^6 + 0*4^7 = 1297
//
// So the the laneX is:
//
//	[]uint64{20548 , 1297 }
//
//
// and in base 11:
//  val0 = 0*11^0 + 1*11^1 + 0*11^2 + 1*11^3 + 0*11^4 + 0*11^5 + 1*11^6 + 1*11^7 = 21260074
//	val1 = 1*11^0 + 0*11^1 + 1*11^2 + 0*11^3 + 1*11^4 + 1*11^5 + 0*11^6 + 0*11^7 = 175814
//
// So the lanex is:
//	[]uint64{21260074, 175814}
//
// for 0x56ad → 0x56, 0xad → [0b01010110, 0b10101101]
// we have  inputBytes := []byte{0b10101101, 0b01010110},
// in base 4:
// laneX = []uint64{17489, 4372}
//
// in base 11
//	laneX = []uint64{19649675, 1786334}

//
// we have 2 hashes, one start at row 0 and the other at row 136.
// the first block is in base 4 and the second block is in base 11.
//
// block Columns:
// the laneX columns are flattened such that the first numRowsPerBlock rows correspond to the first blocks and so on.
// blocks are in positions 0, 23, 48, 71.
