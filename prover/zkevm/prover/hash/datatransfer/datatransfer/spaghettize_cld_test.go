package datatransfer

import (
	"testing"

	permTrace "github.com/consensys/zkevm-monorepo/prover/crypto/keccak"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/generic"
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
	cld := cleanLimbDecomposition{nbCld: maxLanesFromLimb, nbCldSlices: numBytesInLane}
	s := &spaghettizedCLD{}
	def := generic.PHONEY_RLP
	// spaghetti Size
	size := 64
	spaghettiSize := 4 * size
	gbmSize := 32

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
		gbm.AppendTraces(run, &gt, &traces)
		iPadd.assignImportAndPadd(run, gt, size, 0)
		cld.assignCLD(run, iPadd, size)
		s.assignSpaghetti(run, iPadd, cld, spaghettiSize)
	}
	return define, prover
}
func TestSpaghettiModule(t *testing.T) {
	// test keccak
	define, prover := makeTestCaseSpaghettiModule()
	comp := wizard.Compile(define, dummy.Compile)
	proof := wizard.Prove(comp, prover)
	assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")
}
