package keccak

import (
	"fmt"
	"sync"
	"testing"

	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/innerproduct"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/lookup"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/permutation"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/specialqueries"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/datatransfer/acc_module"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/datatransfer/datatransfer"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/stretchr/testify/assert"
)

func MakeTestCaseKeccakModule(numProviders int) (
	define wizard.DefineFunc,
	prover wizard.ProverStep,
) {
	mod := Module{}
	round := 0
	nbKeccakF := 128

	gbmSize := make([]int, numProviders)
	def := make([]generic.GenericByteModuleDefinition, numProviders)
	gbms := make([]generic.GenericByteModule, numProviders)

	def[0] = generic.RLP_ADD
	def[1] = generic.SHAKIRA

	gbmSize[0] = 512
	gbmSize[1] = 128

	define = func(builder *wizard.Builder) {
		comp := builder.CompiledIOP
		for i := range gbms {
			gbms[i] = acc_module.CommitGBM(comp, round, def[i], gbmSize[i])
		}
		mod.Define(comp, gbms, nbKeccakF)
	}

	prover = func(run *wizard.ProverRuntime) {
		for i := range gbms {
			acc_module.AssignGBMfromTable(run, &gbms[i], gbmSize[i]-4, gbmSize[i]/7)
		}
		mod.AssignKeccak(run)
	}

	return define, prover
}

func TestKeccakModule(t *testing.T) {
	definer, prover := MakeTestCaseKeccakModule(2)
	comp := wizard.Compile(definer, dummy.Compile)

	proof := wizard.Prove(comp, prover)
	assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")

}

func BenchmarkKeccakModule(b *testing.B) {

	nbKeccakF := []int{
		1 << 13,
		// 1 << 16,
		// 1 << 18,
		// 1 << 20,
	}
	once := &sync.Once{}

	for _, numKeccakF := range nbKeccakF {

		b.Run(fmt.Sprintf("%v-numKeccakF", numKeccakF), func(b *testing.B) {

			define := func(build *wizard.Builder) {
				comp := build.CompiledIOP
				mod := Module{}
				gbm0 := datatransfer.CommitGBM(comp, 0, generic.SHAKIRA, 2)
				gbm1 := datatransfer.CommitGBM(comp, 0, generic.RLP_ADD, 2)
				gbms := []generic.GenericByteModule{gbm0, gbm1}
				mod.Define(comp, gbms, numKeccakF)
			}

			var (
				compiled = wizard.Compile(
					define,
					specialqueries.RangeProof,
					specialqueries.CompileFixedPermutations,
					permutation.CompileGrandProduct,
					lookup.CompileLogDerivative,
					innerproduct.Compile,
				)
				numCells = 0
				numCols  = 0
			)

			for _, colID := range compiled.Columns.AllKeys() {
				numCells += compiled.Columns.GetSize(colID)
				numCols += 1
			}

			b.ReportMetric(float64(numCells), "#cells")
			b.ReportMetric(float64(numCols), "#columns")

			once.Do(func() {

				for _, colID := range compiled.Columns.AllKeys() {
					fmt.Printf("%v, %v\n", colID, compiled.Columns.GetSize(colID))
				}

			})

		})

	}
}
