package datatransfer

import (
	"testing"

	permTrace "github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/stretchr/testify/assert"
)

// It generates Define and Assign function of ImportAndPadd module, for testing
func makeTestCaseImport(hashType int) (
	define wizard.DefineFunc,
	prover wizard.ProverStep,
) {
	round := 0
	size := 512
	gbmSize := 128
	gbm := generic.GenericByteModule{}
	iPadd := importAndPadd{}
	def := generic.PHONEY_RLP

	define = func(build *wizard.Builder) {
		comp := build.CompiledIOP
		lu := newLookupTables(comp)
		gbm = CommitGBM(comp, round, def, gbmSize)
		iPadd.newImportAndPadd(comp, round, size, gbm, lu, hashType)
	}
	prover = func(run *wizard.ProverRuntime) {
		traces := permTrace.PermTraces{}
		gt := generic.GenTrace{}
		AssignGBMfromTable(run, &gbm)
		gbm.AppendTraces(run, &gt, &traces)
		iPadd.assignImportAndPadd(run, gt, size, hashType)

	}
	return define, prover
}

func TestLImportAndPaddModule(t *testing.T) {
	// test keccak
	define, prover := makeTestCaseImport(Keccak)
	comp := wizard.Compile(define, dummy.Compile)
	proof := wizard.Prove(comp, prover)
	assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof for keccak")

	// test sha2
	define, prover = makeTestCaseImport(Sha2)
	comp = wizard.Compile(define, dummy.Compile)
	proof = wizard.Prove(comp, prover)
	assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof for sha2")
}
