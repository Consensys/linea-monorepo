package keccak

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/pragmas"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
	keccakfkoalabear "github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf_koalabear"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf_koalabear/iokeccakf"
	"github.com/stretchr/testify/assert"
)

// makes Define and Prove function for testing [NewKeccakOverBlocks]
func MakeTestCaseCustomizedKeccak(t *testing.T, providers [][]byte) (
	define wizard.DefineFunc,
	prover wizard.MainProverStep,
) {
	var (
		mod = &KeccakOverBlocks{
			Outputs: &iokeccakf.KeccakFOutputs{},
		}
		maxNumKeccakF = 2
		nbRowsPerLane = generic.KeccakUsecase.LaneSizeBytes() / common.LimbBytes
		laneSize      = utils.NextPowerOfTwo(maxNumKeccakF * generic.KeccakUsecase.NbOfLanesPerBlock() * nbRowsPerLane)
		keccakfSize   = keccakfkoalabear.NumRows(maxNumKeccakF)
	)

	define = func(builder *wizard.Builder) {
		comp := builder.CompiledIOP
		createCol := common.CreateColFn(comp, "Test_Customized_Keccak", laneSize, pragmas.RightPadded)

		inp := KeccakOverBlockInputs{
			LaneInfo: iokeccakf.LaneInfo{
				Lane:                 createCol("Lanes"),
				IsBeginningOfNewHash: createCol("IsFirstLaneOfNewHash"),
				IsLaneActive:         createCol("IsLaneActive"),
			},

			KeccakfSize: keccakfSize,
			Provider:    providers,
		}
		mod = NewKeccakOverBlocks(comp, inp)
	}

	prover = func(run *wizard.ProverRuntime) {

		AssignLaneInfo(run, &mod.Inputs.LaneInfo, mod.Inputs.Provider)

		mod.Run(run)

		// check the hash result
		permTrace := keccak.GenerateTrace(mod.Inputs.Provider)
		fmt.Printf("expected hashes: %+v\n", permTrace.HashOutPut)
		// extract hash result from the module
		actualHashes := mod.Outputs.ExtractHashResult(run)
		fmt.Printf("actual hashes: %+v\n", actualHashes)

		if len(actualHashes) != len(permTrace.HashOutPut) {
			t.Fatalf("expected %d hashes, got %d", len(permTrace.HashOutPut), len(actualHashes))
		}

		for i, expectedHash := range permTrace.HashOutPut {
			// hashHi := hash[:16] ,  hashLo := hash[16:]
			assert.Equal(t, expectedHash, actualHashes[i], "hash %d mismatch", i)
		}
	}

	return define, prover
}

func TestCustomizedKeccak(t *testing.T) {
	var providers [][]byte
	// generate 20 random slices of bytes
	for i := 0; i < 1; i++ {
		length := (i + 1) * generic.KeccakUsecase.BlockSizeBytes()
		// generate random bytes
		slice := make([]byte, length)
		rand.Read(slice)
		providers = append(providers, slice)
	}

	definer, prover := MakeTestCaseCustomizedKeccak(t, providers)
	comp := wizard.Compile(definer, dummy.Compile)

	proof := wizard.Prove(comp, prover)
	assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")
}
