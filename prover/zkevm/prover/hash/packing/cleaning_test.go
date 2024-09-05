package packing

import (
	"testing"

	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/stretchr/testify/assert"
)

// It generates Define and Assign function of Cleaning module
func makeTestCaseCleaningModule(uc generic.HashingUsecase) (
	define wizard.DefineFunc,
	prover wizard.ProverStep,
) {
	var (
		// max number of blocks that can be extracted from limbs
		// if the number of blocks passes the max, newPack() would panic.
		maxNumBlock = 67
		// if the blockSize is not consistent with PackingParam, newPack() would panic.
		blockSize = uc.BlockSizeBytes()
		// for testing; used to populate the importation columns
		// since we have at least one block per hash, the umber of hashes should be less than maxNumBlocks
		numHash = 33
		// max number of limbs
		size = utils.NextPowerOfTwo(maxNumBlock * blockSize)
	)

	imported := Importation{}
	cleaning := cleaningCtx{}

	define = func(build *wizard.Builder) {
		comp := build.CompiledIOP
		imported = createImportationColumns(comp, size)
		lookup := NewLookupTables(comp)
		cleaning = NewClean(comp, newCleaningInputs(imported, lookup, "TEST"))
	}
	prover = func(run *wizard.ProverRuntime) {
		var (
			imported = cleaning.Inputs.imported
		)
		// assign the importation columns
		assignImportationColumns(run, &imported, numHash, blockSize, size)

		// assign the cleaning module.
		cleaning.Assign(run)
	}
	return define, prover
}

func TestCleaningModule(t *testing.T) {
	for _, uc := range testCases {
		t.Run(uc.Name, func(t *testing.T) {
			define, prover := makeTestCaseCleaningModule(uc.UseCase)
			comp := wizard.Compile(define, dummy.Compile)
			proof := wizard.Prove(comp, prover)
			assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")
		})
	}
}
