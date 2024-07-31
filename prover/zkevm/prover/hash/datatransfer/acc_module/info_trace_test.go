package acc_module

import (
	"testing"

	permTrace "github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/datatransfer/datatransfer"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/stretchr/testify/assert"
)

// It generates Define and Assign function of Bridging module, for testing
func makeTestCaseInfoModule(numModules int) (
	define wizard.DefineFunc,
	prover wizard.ProverStep,
) {
	round := 0

	gbmSize := make([]int, numModules)
	def := make([]generic.GenericByteModuleDefinition, numModules)
	gbms := make([]generic.GenericByteModule, numModules)
	maxNumKeccakF := 128

	info := InfoModule{}
	d := &DataModule{}
	dt := datatransfer.Module{}
	d.MaxNumRows = 1024
	dt.HashOutput.MaxNumRows = maxNumKeccakF

	def[0] = generic.RLP_ADD
	def[1] = generic.SHAKIRA
	def[2] = module1
	def[3] = module2
	gbmSize[0] = 32
	gbmSize[1] = 32
	gbmSize[2] = 32
	gbmSize[3] = 32

	define = func(build *wizard.Builder) {
		comp := build.CompiledIOP
		for i := range gbms {
			gbms[i] = CommitGBM(comp, round, def[i], gbmSize[i])
		}
		d.declareColumns(comp, round, gbms)
		dt.HashOutput.DeclareColumns(comp, round)
		info.NewInfoModule(comp, round, maxNumKeccakF, gbms, dt.HashOutput, *d)
	}
	prover = func(run *wizard.ProverRuntime) {
		traces := permTrace.PermTraces{}
		gt := make([]generic.GenTrace, numModules)
		for i := range gbms {
			AssignGBMfromTable(run, &gbms[i], gbmSize[i]-7, gbmSize[i]/5)
			gbms[i].AppendTraces(run, &gt[i], &traces)
		}
		d.AssignDataModule(run, gbms)
		dt.HashOutput.AssignHashOutPut(run, traces)

		info.AssignInfoModule(run, gbms)

	}
	return define, prover
}

func TestInfoModule(t *testing.T) {
	define, prover := makeTestCaseInfoModule(4)
	comp := wizard.Compile(define, dummy.Compile)

	proof := wizard.Prove(comp, prover)
	assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")
}
