package packing

import (
	"testing"

	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/common"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/stretchr/testify/assert"
)

// It generates Define and Assign function of Packing module, for testing
func makeTestCaseCLDModule(uc generic.HashingUsecase) (
	define wizard.DefineFunc,
	prover wizard.ProverStep,
) {
	var (
		// max number of blocks that can be extracted from limbs
		// if the number of blocks passes the max, newPack() would panic.
		maxNumBlock = 108
		// if the blockSize is not consistent with PackingParam, newPack() would panic.
		blockSize = uc.BlockSizeBytes()
		// for testing; used to populate the importation columns
		// since we have at least one block per hash, the umber of hashes should be less than maxNumBlocks
		numHash = 73
		// max number of limbs
		size = utils.NextPowerOfTwo(maxNumBlock * blockSize)
	)

	imported := Importation{}
	ctx := cleaningCtx{}
	decomposed := decomposition{}

	define = func(build *wizard.Builder) {
		comp := build.CompiledIOP
		imported = createImportationColumns(comp, size)

		createCol := common.CreateColFn(comp, CLEANING, imported.Limb.Size())
		ctx = cleaningCtx{
			CleanLimb: createCol("CleanLimb"),
			Inputs: &cleaningInputs{
				imported: imported,
				lookup:   NewLookupTables(comp)},
		}

		inp := decompositionInputs{
			param:       uc,
			cleaningCtx: ctx,
		}

		decomposed = newDecomposition(comp, inp)
	}
	prover = func(run *wizard.ProverRuntime) {
		// assign the importation columns
		assignImportationColumns(run, &imported, numHash, blockSize, size)

		// assign all the Packing module.
		ctx.assignCleanLimbs(run)
		decomposed.Assign(run)
	}
	return define, prover
}

func TestCLDModule(t *testing.T) {
	for _, uc := range testCases {
		t.Run(uc.Name, func(t *testing.T) {
			define, prover := makeTestCaseCLDModule(uc.UseCase)
			comp := wizard.Compile(define, dummy.Compile)
			proof := wizard.Prove(comp, prover)
			assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")
		})
	}
}
