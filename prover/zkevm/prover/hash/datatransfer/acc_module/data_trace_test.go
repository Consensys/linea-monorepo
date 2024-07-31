package acc_module

import (
	"testing"

	permTrace "github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/stretchr/testify/assert"
)

// It generates Define and Assign function of Data module, for testing
func makeTestCaseDataModule(numModules int) (
	define wizard.DefineFunc,
	prover wizard.ProverStep,
) {
	round := 0

	gbmSize := make([]int, numModules)
	def := make([]generic.GenericByteModuleDefinition, numModules)
	gbms := make([]generic.GenericByteModule, numModules)
	maxNumKeccakF := 8

	d := DataModule{}

	def[0] = generic.RLP_ADD
	def[1] = generic.SHAKIRA
	def[2] = module1
	def[3] = module2
	gbmSize[0] = 8
	gbmSize[1] = 8
	gbmSize[2] = 32
	gbmSize[3] = 8

	define = func(build *wizard.Builder) {
		comp := build.CompiledIOP
		for i := range gbms {
			gbms[i] = CommitGBM(comp, round, def[i], gbmSize[i])
		}
		d.NewDataModule(comp, round, maxNumKeccakF, gbms[:])
	}
	prover = func(run *wizard.ProverRuntime) {
		traces := permTrace.PermTraces{}
		gt := make([]generic.GenTrace, numModules)
		for i := range gbms {
			AssignGBMfromTable(run, &gbms[i], gbmSize[i]-5, gbmSize[i]/5)
			gbms[i].AppendTraces(run, &gt[i], &traces)
		}
		d.AssignDataModule(run, gbms)
	}
	return define, prover
}

func TestDataModule(t *testing.T) {
	define, prover := makeTestCaseDataModule(4)
	comp := wizard.Compile(define, dummy.Compile)

	proof := wizard.Prove(comp, prover)
	assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")
}

var module1 = generic.GenericByteModuleDefinition{
	Data: generic.DataDef{
		HashNum: "module1_HashNum",
		Limb:    "module1_Limb",
		NBytes:  "module1_NBytes",
		TO_HASH: "module1_TO_Hash",
		Index:   "module1_Index",
	},
	Info: generic.InfoDef{
		HashNum:  "module1_HashNum_Info",
		HashLo:   "module1_HashLo",
		HashHi:   "module1_HashHi",
		IsHashLo: "module1_IsHashLo",
		IsHashHi: "module1_IsHashHi",
	},
}

var module2 = generic.GenericByteModuleDefinition{
	Data: generic.DataDef{
		HashNum: "module2_HashNum",
		Limb:    "module2_Limb",
		NBytes:  "module2_NBytes",
		TO_HASH: "module2_TO_Hash",
		Index:   "module2_Index",
	},
	Info: generic.InfoDef{
		HashNum:  "module2_HashNum_Info",
		HashLo:   "module2_HashLo",
		HashHi:   "module2_HashHi",
		IsHashLo: "module2_IsHashLo",
		IsHashHi: "module2_IsHashHi",
	},
}
