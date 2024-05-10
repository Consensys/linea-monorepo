package datatransfer

import (
	"testing"

	permTrace "github.com/consensys/zkevm-monorepo/prover/crypto/keccak"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/keccak/generic"
	"github.com/stretchr/testify/assert"
)

// It generates Define and Assign function of LCD module, for testing
func makeTestCaseCLD() (
	define wizard.DefineFunc,
	prover wizard.ProverStep,
) {
	round := 0
	size := 2048
	gbmSize := 512
	gbm := generic.GenericByteModule{}
	iPadd := importAndPadd{}
	cld := cleanLimbDecomposition{}
	def := generic.PHONEY_RLP

	define = func(build *wizard.Builder) {
		comp := build.CompiledIOP
		lu := newLookupTables(comp)
		gbm = CommitGBM(comp, round, def, gbmSize)
		iPadd.insertCommit(comp, round, size)
		cld.newCLD(comp, round, lu, iPadd, size)
	}
	prover = func(run *wizard.ProverRuntime) {
		traces := permTrace.PermTraces{}
		gt := generic.GenTrace{}
		AssignGBMfromTable(run, &gbm)
		gbm.AppendTraces(run, &traces, &gt)
		iPadd.assignImportAndPadd(run, traces, gt, size)
		cld.assignCLD(run, iPadd, size)
	}
	return define, prover
}

func TestCLDModule(t *testing.T) {
	define, prover := makeTestCaseCLD()
	comp := wizard.Compile(define, dummy.Compile)

	proof := wizard.Prove(comp, prover)
	assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")
}
