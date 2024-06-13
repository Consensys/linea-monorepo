package datatransfer

import (
	"testing"

	permTrace "github.com/consensys/zkevm-monorepo/prover/crypto/keccak"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/keccak/generic"
	"github.com/stretchr/testify/assert"
)

// It generates Define and Assign function of ImportAndPadd module, for testing
func makeTestCaseImport() (
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
		gbm = CommitGBM(comp, round, def, gbmSize)
		iPadd.newImportAndPadd(comp, round, size, gbm)
	}
	prover = func(run *wizard.ProverRuntime) {
		traces := permTrace.PermTraces{}
		gt := generic.GenTrace{}
		AssignGBMfromTable(run, &gbm)
		gbm.AppendTraces(run, &traces, &gt)
		iPadd.assignImportAndPadd(run, traces, gt, size)

	}
	return define, prover
}

func TestLImportAndPaddModule(t *testing.T) {
	define, prover := makeTestCaseImport()
	comp := wizard.Compile(define, dummy.Compile)

	proof := wizard.Prove(comp, prover)
	assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")
}
