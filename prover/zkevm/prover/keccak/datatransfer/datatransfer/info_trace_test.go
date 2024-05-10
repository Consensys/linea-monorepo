package datatransfer

import (
	"testing"

	permTrace "github.com/consensys/zkevm-monorepo/prover/crypto/keccak"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/keccak/generic"
	"github.com/stretchr/testify/assert"
)

// It generates Define and Assign function of info module, for testing
func makeTestCaseHashOutput() (
	define wizard.DefineFunc,
	prover wizard.ProverStep,
) {
	round := 0
	gbmSize := 512
	maxNumKeccakF := 128
	gbm := generic.GenericByteModule{}
	def := generic.PHONEY_RLP
	h := HashOutput{}

	define = func(build *wizard.Builder) {
		comp := build.CompiledIOP
		gbm = CommitGBM(comp, round, def, gbmSize)
		lu := newLookupTables(comp)
		h.newHashOutput(comp, round, maxNumKeccakF, lu)
	}
	prover = func(run *wizard.ProverRuntime) {
		traces := permTrace.PermTraces{}
		gt := generic.GenTrace{}
		AssignGBMfromTable(run, &gbm)
		gbm.AppendTraces(run, &traces, &gt)
		h.AssignHashOutPut(run, traces)

	}
	return define, prover
}

func TestInfoTraceModule(t *testing.T) {
	define, prover := makeTestCaseHashOutput()
	comp := wizard.Compile(define, dummy.Compile)

	proof := wizard.Prove(comp, prover)
	assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")
}
