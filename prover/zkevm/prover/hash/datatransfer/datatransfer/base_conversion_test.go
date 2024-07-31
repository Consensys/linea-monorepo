package datatransfer

import (
	"testing"

	permTrace "github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/stretchr/testify/assert"
)

func makeTestCaseBaseConversionModule() (
	define wizard.DefineFunc,
	prover wizard.ProverStep,
) {
	round := 0
	gbm := generic.GenericByteModule{}
	iPadd := importAndPadd{}
	cld := cleanLimbDecomposition{nbCld: maxLanesFromLimb, nbCldSlices: numBytesInLane}
	s := spaghettizedCLD{}
	l := lane{}
	b := baseConversion{}
	def := generic.PHONEY_RLP
	cldSize := 2048
	gbmSize := 512
	spaghettiSize := 8 * cldSize
	laneSize := 4 * cldSize

	define = func(build *wizard.Builder) {
		comp := build.CompiledIOP
		gbm = CommitGBM(comp, round, def, gbmSize)
		lu := newLookupTables(comp)
		iPadd.insertCommit(comp, round, cldSize)
		cld.insertCommit(comp, round, cldSize)
		s.insertCommit(comp, round, cld, spaghettiSize)
		l.insertCommitForTest(comp, round, spaghettiSize, laneSize)
		b.newBaseConversionOfLanes(comp, round, laneSize, l, lu)

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
		b.assignBaseConversion(run, l, laneSize)

	}
	return define, prover
}
func TestBaseConversionModule(t *testing.T) {
	// test keccak
	define, prover := makeTestCaseBaseConversionModule()
	comp := wizard.Compile(define, dummy.Compile)

	proof := wizard.Prove(comp, prover)
	assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")
}

func (l *lane) insertCommitForTest(comp *wizard.CompiledIOP, round, maxNumRows, laneSize int) {
	l.lane = comp.InsertCommit(round, ifaces.ColIDf("Lane"), laneSize)
	l.coeff = comp.InsertCommit(round, ifaces.ColIDf("Coefficient"), maxNumRows)
	l.isLaneActive = comp.InsertCommit(round, ifaces.ColIDf("LaneIsActive"), laneSize)
	l.isFirstLaneOfNewHash = comp.InsertCommit(round, ifaces.ColIDf("IsFirstLane_Of_NewHash"), laneSize)
	l.isLaneCompleteShifted = comp.InsertCommit(round, ifaces.ColIDf("IsLaneCompleteShifted"), maxNumRows)
}
