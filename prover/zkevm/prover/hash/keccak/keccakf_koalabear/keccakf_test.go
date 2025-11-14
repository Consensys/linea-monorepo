package keccakfkoalabear

import (
	"encoding/binary"
	"math/rand/v2"
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	kcommon "github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf_koalabear/common"
	"github.com/stretchr/testify/assert"
)

func TestKeccakf(t *testing.T) {

	// #nosec G404 --we don't need a cryptographic RNG for testing purpose
	rng := rand.New(rand.NewChaCha8([32]byte{}))
	numCases := 30
	maxNumKeccakf := 2
	// The -1 is here to prevent the generation of a padding block
	maxInputBytes := maxNumKeccakf*keccak.Rate - 1

	definer, prover := keccakfTestingModule(maxNumKeccakf)
	comp := wizard.Compile(definer, dummy.Compile)

	for i := 0; i < numCases; i++ {
		// Generate a random piece of data
		dataSize := rng.IntN(maxInputBytes + 1)
		data := make([]byte, dataSize)
		utils.ReadPseudoRand(rng, data)

		// Generate permutation traces for the data
		traces := keccak.PermTraces{}
		keccak.Hash(data, &traces)

		proof := wizard.Prove(comp, prover(t, traces))
		assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")
	}
}

func keccakfTestingModule(
	maxNumKeccakf int,
) (
	define wizard.DefineFunc,
	prover func(t *testing.T, traces keccak.PermTraces) wizard.MainProverStep,
) {

	var (
		mod    = &Module{}
		size   = numRows(maxNumKeccakf)
		blocks = make([][kcommon.NumSlices]ifaces.Column, kcommon.NumLanesInBlock)
	)

	// The testing wizard uniquely calls the keccakf module
	define = func(b *wizard.Builder) {

		comp := b.CompiledIOP
		for m := 0; m < kcommon.NumLanesInBlock; m++ {
			for z := 0; z < kcommon.NumSlices; z++ {
				blocks[m][z] = comp.InsertCommit(0, ifaces.ColIDf("BLOCK_%v_%v", m, z), size)
			}
		}

		mod = NewModule(b.CompiledIOP, KeccakfInputs{
			Blocks:       blocks,
			IsBlock:      comp.InsertCommit(0, "IS_BLOCK", size),
			IsFirstBlock: comp.InsertCommit(0, "IS_FIRST_BLOCK", size),
			IsBlockBaseB: comp.InsertCommit(0, "IS_BLOCK_BASEB", size),
			IsActive:     comp.InsertCommit(0, "IS_ACTIVE", size),
			KeccakfSize:  size,
		})
	}

	// And the prover (instanciated for traces) is called
	prover = func(
		t *testing.T,
		traces keccak.PermTraces,
	) wizard.MainProverStep {
		return func(run *wizard.ProverRuntime) {
			// assign the input columns

			mod.assignBlocks(run, traces)
			mod.assignBlockFlags(run, traces)

			// Assigns the module
			mod.Assign(run, traces)

			// Asserts that the last value in aIota is the correct one. `pos` is
			// the last active row of the module (given the traces we got). We
			// use it to reconstruct what the module "believes" to be the final
			// keccak state. Then, we compare this value with one generated in
			// the traces.
			numPerm := len(traces.KeccakFInps)
			pos := numPerm*keccak.NumRound - 1
			expectedState := traces.KeccakFOuts[numPerm-1]
			extractedState := keccak.State{}
			for x := 0; x < 5; x++ {
				for y := 0; y < 5; y++ {
					var a [8]uint8
					for z := 0; z < kcommon.NumSlices; z++ {
						v := mod.BackToThetaOrOutput.StateNext[x][y][z].GetColAssignmentAt(run, pos)
						a[z] = uint8(v.Uint64())
					}
					extractedState[x][y] = binary.LittleEndian.Uint64(a[:])
				}
			}

			assert.Equal(t, expectedState, extractedState)
		}
	}

	return define, prover
}

/*func BenchmarkDataTransferModule(b *testing.B) {
	b.Skip()
	maxNumKeccakF := []int{
		1 << 13,
		// 1 << 16,
		// 1 << 18,
		// 1 << 20,
	}
	once := &sync.Once{}

	for _, numKeccakF := range maxNumKeccakF {

		b.Run(fmt.Sprintf("%v-numKeccakF", numKeccakF), func(b *testing.B) {

			define := func(build *wizard.Builder) {
				comp := build.CompiledIOP
				mod := &Module{}
				*mod = NewModule(comp, 0, numKeccakF)
			}

			var (
				compiled = wizard.Compile(
					define,
					specialqueries.RangeProof,
					specialqueries.CompileFixedPermutations,
					permutation.CompileViaGrandProduct,
					logderivativesum.CompileLookups,
					innerproduct.Compile(),
				)
				numCells = 0
				numCols  = 0
			)

			for _, colID := range compiled.Columns.AllKeys() {
				numCells += compiled.Columns.GetSize(colID)
				numCols += 1
			}

			b.ReportMetric(float64(numCells), "#cells")
			b.ReportMetric(float64(numCols), "#columns")

			once.Do(func() {

				for _, colID := range compiled.Columns.AllKeys() {
					fmt.Printf("%v, %v\n", colID, compiled.Columns.GetSize(colID))
				}

			})

		})

	}
}*/

// it extracts two bytes chunks from a given u64 value.
func TwoBytesFromU64(block uint64) (twoBytes []field.Element) {
	twoBytes = make([]field.Element, 4)
	for j := 0; j < 4; j++ {
		twoBytes[j] = field.NewElement((block >> (4 * j)) & 0xff)
	}
	return twoBytes
}

// Assigns blocks using the keccak traces.
func (mod *Module) assignBlocks(
	run *wizard.ProverRuntime,
	traces keccak.PermTraces,
) {

	var (
		colSize      = mod.Inputs.KeccakfSize
		numKeccakF   = len(traces.KeccakFInps)
		unpaddedSize = numKeccakF * keccak.NumRound
	)

	run.AssignColumn(
		mod.Inputs.IsActive.GetColID(),
		smartvectors.RightZeroPadded(
			vector.Repeat(field.One(), unpaddedSize),
			colSize,
		),
	)

	// Assign the block in BaseB.
	blocksVal := [kcommon.NumLanesInBlock][kcommon.NumSlices][]field.Element{}

	for m := range blocksVal {
		for z := 0; z < kcommon.NumSlices; z++ {
			blocksVal[m][z] = make([]field.Element, unpaddedSize)
		}
	}

	parallel.Execute(numKeccakF, func(start, stop int) {

		for nperm := start; nperm < stop; nperm++ {

			currBlock := traces.Blocks[nperm]

			for r := 0; r < keccak.NumRound; r++ {
				// Current row that we are assigning
				currRow := nperm*keccak.NumRound + r

				// Retro-actively assign the block in BaseB if we are not on
				// the first row. The condition over nperm is to ensure that we
				// do not underflow although in practice isNewHash[0] will
				// always be true because this is the first perm of the first
				// hash by definition.
				if r == 0 && nperm > 0 && !traces.IsNewHash[nperm] {
					block := cleanBaseBlock(currBlock, &kcommon.BaseChiFr)
					for m := 0; m < kcommon.NumLanesInBlock; m++ {
						for z := 0; z < kcommon.NumSlices; z++ {
							blocksVal[m][z][currRow-1] = block[m][z]
						}
					}
				}
				//assign the firstBlock in BaseA
				if r == 0 && traces.IsNewHash[nperm] {
					block := cleanBaseBlock(currBlock, &kcommon.BaseThetaFr)
					for m := 0; m < kcommon.NumLanesInBlock; m++ {
						for z := 0; z < kcommon.NumSlices; z++ {
							blocksVal[m][z][currRow] = block[m][z]
						}
					}
				}
			}
		}
	})

	for m := 0; m < kcommon.NumLanesInBlock; m++ {
		for z := 0; z < kcommon.NumSlices; z++ {
			run.AssignColumn(
				mod.Inputs.Blocks[m][z].GetColID(),
				smartvectors.RightZeroPadded(blocksVal[m][z], colSize),
			)
		}
	}

}

// It assigns the columns specific to the submodule.
func (mod *Module) assignBlockFlags(
	run *wizard.ProverRuntime,
	permTrace keccak.PermTraces,
) {
	var (
		isFirstBlock = common.NewVectorBuilder(mod.Inputs.IsFirstBlock)
		isBlockBaseB = common.NewVectorBuilder(mod.Inputs.IsBlockBaseB)
		isBlock      = common.NewVectorBuilder(mod.Inputs.IsBlock)
	)

	zeroes := make([]field.Element, keccak.NumRound-1)
	for i := range permTrace.IsNewHash {
		if permTrace.IsNewHash[i] {
			isFirstBlock.PushOne()
			isFirstBlock.PushSliceF(zeroes)
			// append 24 zeroes
			isBlockBaseB.PushSliceF(zeroes)
			isBlockBaseB.PushInt(0)
			// populate IsBlock
			isBlock.PushOne()
			isBlock.PushSliceF(zeroes)

		} else {
			isFirstBlock.PushZero()
			isFirstBlock.PushSliceF(zeroes)

			isBlockBaseB.OverWriteInt(1)
			// append 24 zeroes
			isBlockBaseB.PushSliceF(zeroes)
			isBlockBaseB.PushZero()

			//populate IsBlock
			isBlock.OverWriteInt(1)
			isBlock.PushSliceF(zeroes)
			isBlock.PushZero()
		}
	}

	isBlock.PadAndAssign(run)
	isFirstBlock.PadAndAssign(run)
	isBlockBaseB.PadAndAssign(run)

}
