package acc_module

import (
	"testing"

	permTrace "github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/datatransfer/datatransfer"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/stretchr/testify/assert"
)

// It generates Define and Assign function of Bridging module, for testing
func makeTestCaseBridging(numModules int) (
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

	def[0] = generic.RLP_ADD
	def[1] = generic.SHAKIRA

	gbmSize[0] = 8
	gbmSize[1] = 128

	define = func(build *wizard.Builder) {
		comp := build.CompiledIOP
		for i := range gbms {
			gbms[i] = CommitGBM(comp, round, def[i], gbmSize[i])
		}
		d.NewDataModule(comp, round, maxNumKeccakF, gbms[:])
		dt.Provider = d.Provider
		dt.NewDataTransfer(comp, round, maxNumKeccakF, 0)
		info.NewInfoModule(comp, round, maxNumKeccakF, gbms, dt.HashOutput, *d)
	}
	prover = func(run *wizard.ProverRuntime) {
		traces1 := permTrace.PermTraces{}
		traces2 := permTrace.PermTraces{}
		gt := make([]generic.GenTrace, numModules)
		for i := range gbms {
			AssignGBMfromTable(run, &gbms[i], gbmSize[i]-3, gbmSize[i]/5)
			gbms[i].AppendTraces(run, &gt[i], &traces1)
		}

		d.AssignDataModule(run, gbms)
		sGT := generic.GenTrace{}
		d.Provider.AppendTraces(run, &sGT, &traces2)
		if len(traces1.HashOutPut) != len(traces2.HashOutPut) {
			utils.Panic("trace are not the same")
		}
		dt.AssignModule(run, traces2, sGT)
		// dt.HashOutput.AssignHashOutPut(run, traces)
		info.AssignInfoModule(run, gbms)

	}
	return define, prover
}

func TestBridging(t *testing.T) {
	define, prover := makeTestCaseBridging(2)
	comp := wizard.Compile(define, dummy.Compile)
	proof := wizard.Prove(comp, prover)
	assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")
}
