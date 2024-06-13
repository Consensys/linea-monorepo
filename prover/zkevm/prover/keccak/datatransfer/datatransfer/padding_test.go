package datatransfer

import (
	"testing"

	permTrace "github.com/consensys/zkevm-monorepo/prover/crypto/keccak"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/keccak/generic"
	"github.com/stretchr/testify/assert"
)

// It generates Define and Assign function of Padding (sub) submodule, for testing
func GetPaddingForTest() (
	define wizard.DefineFunc,
	prover wizard.ProverStep,
) {
	round := 0
	gbmSize := 128
	size := 512
	iPadd := importAndPadd{}
	gbm := generic.GenericByteModule{}
	def := generic.PHONEY_RLP
	define = func(build *wizard.Builder) {
		comp := build.CompiledIOP
		iPadd.insertCommit(comp, round, size)
		gbm = CommitGBM(comp, round, def, gbmSize)
		iPadd.insertPadding(comp, round)

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
func TestPadding(t *testing.T) {
	define, prover := GetPaddingForTest()
	comp := wizard.Compile(define, dummy.Compile)

	proof := wizard.Prove(comp, prover)
	assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")
}
