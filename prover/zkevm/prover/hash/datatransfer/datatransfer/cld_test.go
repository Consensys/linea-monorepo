package datatransfer

import (
	"testing"

	permTrace "github.com/consensys/zkevm-monorepo/prover/crypto/keccak"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/stretchr/testify/assert"
)

// It generates Define and Assign function of LCD module, for testing
func makeTestCaseCLD(hashType int) (
	define wizard.DefineFunc,
	prover wizard.ProverStep,
) {
	round := 0
	size := 1024
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
		cld.newCLD(comp, round, lu, iPadd, size, hashType)
	}
	prover = func(run *wizard.ProverRuntime) {
		traces := permTrace.PermTraces{}
		gt := generic.GenTrace{}
		AssignGBMfromTable(run, &gbm)
		gbm.AppendTraces(run, &gt, &traces)
		iPadd.assignImportAndPadd(run, gt, size, hashType)
		cld.assignCLD(run, iPadd, size)
	}
	return define, prover
}

func TestCLDModule(t *testing.T) {
	// test cld for keccak
	define, prover := makeTestCaseCLD(0)
	comp := wizard.Compile(define, dummy.Compile)
	proof := wizard.Prove(comp, prover)
	assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")

	// test cld for sha2
	define, prover = makeTestCaseCLD(1)
	comp = wizard.Compile(define, dummy.Compile)
	proof = wizard.Prove(comp, prover)
	assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")
}
