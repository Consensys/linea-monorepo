package datatransfer

import (
	"testing"

	permTrace "github.com/consensys/zkevm-monorepo/prover/crypto/keccak"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/stretchr/testify/assert"
)

// It generates Define and Assign function of SpaghetizedCLD module, for testing
func makeTestCaseLaneModule() (
	define wizard.DefineFunc,
	prover wizard.ProverStep,
) {
	round := 0
	gbm := generic.GenericByteModule{}
	iPadd := importAndPadd{}
	cld := cleanLimbDecomposition{nbCld: maxLanesFromLimb, nbCldSlices: numBytesInLane}
	s := spaghettizedCLD{}
	l := lane{}
	def := generic.PHONEY_RLP

	gbmSize := 512
	cldSize := 2048
	spaghettiSize := 8 * cldSize
	laneSize := 4 * cldSize

	define = func(build *wizard.Builder) {
		comp := build.CompiledIOP
		gbm = CommitGBM(comp, round, def, gbmSize)
		iPadd.insertCommit(comp, round, cldSize)
		cld.insertCommit(comp, round, cldSize)
		s.insertCommit(comp, round, cld, spaghettiSize)
		l.newLane(comp, round, spaghettiSize, laneSize, s)

	}
	prover = func(run *wizard.ProverRuntime) {
		permTrace := permTrace.PermTraces{}
		gt := generic.GenTrace{}
		AssignGBMfromTable(run, &gbm)
		gbm.AppendTraces(run, &gt, &permTrace)
		iPadd.assignImportAndPadd(run, gt, cldSize, 0)
		cld.assignCLD(run, iPadd, cldSize)
		s.assignSpaghetti(run, iPadd, cld, spaghettiSize)
		l.assignLane(run, iPadd, s, permTrace, spaghettiSize, laneSize)

	}
	return define, prover
}
func TestLaneModule(t *testing.T) {
	// test keccak
	define, prover := makeTestCaseLaneModule()
	comp := wizard.Compile(define, dummy.Compile)

	proof := wizard.Prove(comp, prover)
	assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")
}
