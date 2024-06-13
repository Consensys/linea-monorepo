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
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/keccak/datatransfer/datatransfer"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/keccak/generic"
	"github.com/stretchr/testify/assert"
)

func MakeTestCaseKeccakModule() (
	define wizard.DefineFunc,
	prover wizard.ProverStep,
) {
	mod := Module{}
	round := 0
	gbm := generic.GenericByteModule{}
	gbmDef := generic.PHONEY_RLP
	gbmSize := 128
	set := Settings{}
	set.MaxNumKeccakf = 32
	mod.Settings = &set
	define = func(builder *wizard.Builder) {
		comp := builder.CompiledIOP
		gbm = datatransfer.CommitGBM(comp, round, gbmDef, gbmSize)
		mod.dataTransfer.GBM = gbm
		mod.Define(comp)
	}

	prover = func(run *wizard.ProverRuntime) {
		datatransfer.AssignGBMfromTable(run, &gbm)
		mod.AssignKeccak(run)
	}

	return define, prover
}

func TestKeccakModule(t *testing.T) {

	t.Skipf("removed for now: fails on CI")

	definer, prover := MakeTestCaseKeccakModule()
	comp := wizard.Compile(definer, dummy.Compile)

	proof := wizard.Prove(comp, prover)
	assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")

}

func BenchmarkKeccakModule(b *testing.B) {

	maxNumKeccakF := []int{
		1 << 13,
		// 1 << 16,
		// 1 << 18,
		// 1 << 20,
	}
	once := &sync.Once{}

	for _, numKeccakF := range maxNumKeccakF {

		b.Run(fmt.Sprintf("%v-numKeccakF", numKeccakF), func(b *testing.B) {

			define := func(build *wizard.Builder) {
				comp := build.CompiledIOP
				mod := Module{}
				set := Settings{}
				set.MaxNumKeccakf = numKeccakF
				mod.Settings = &set
				gbm := datatransfer.CommitGBM(comp, 0, generic.PHONEY_RLP, 2)
				mod.dataTransfer.GBM = gbm
				mod.Define(comp)
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
