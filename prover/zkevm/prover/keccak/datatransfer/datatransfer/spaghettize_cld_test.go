package datatransfer

import (
	"testing"

	permTrace "github.com/consensys/zkevm-monorepo/prover/crypto/keccak"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/keccak/generic"
	"github.com/stretchr/testify/assert"
)

// It generates Define and Assign function of spaghettizedCLD module, for testing
func makeTestCaseSpaghettiModule() (
	define wizard.DefineFunc,
	prover wizard.ProverStep,
) {
	round := 0
	gbm := generic.GenericByteModule{}
	iPadd := importAndPadd{}
	cld := cleanLimbDecomposition{}
	s := &spaghettizedCLD{}
	def := generic.PHONEY_RLP
	// spaghetti Size
	size := 2048
	spaghettiSize := 8 * size
	gbmSize := 512

	define = func(build *wizard.Builder) {
		comp := build.CompiledIOP
		gbm = CommitGBM(comp, round, def, gbmSize)
		iPadd.insertCommit(comp, round, size)
		cld.insertCommit(comp, round, size)
		s.newSpaghetti(comp, round, iPadd, cld, spaghettiSize)

	}
	prover = func(run *wizard.ProverRuntime) {
		traces := permTrace.PermTraces{}
		gt := generic.GenTrace{}
		AssignGBMfromTable(run, &gbm)
		gbm.AppendTraces(run, &traces, &gt)
		iPadd.assignImportAndPadd(run, traces, gt, size)
		cld.assignCLD(run, iPadd, size)
		s.assignSpaghetti(run, iPadd, cld, spaghettiSize)

	}
	return define, prover
}
func TestSpaghettiModule(t *testing.T) {
	define, prover := makeTestCaseSpaghettiModule()
	comp := wizard.Compile(define, dummy.Compile)

	proof := wizard.Prove(comp, prover)
	assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")
}
