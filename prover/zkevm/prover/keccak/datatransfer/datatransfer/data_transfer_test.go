package datatransfer

import (
	"fmt"
	"sync"
	"testing"

	permTrace "github.com/consensys/zkevm-monorepo/prover/crypto/keccak"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/innerproduct"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/lookup"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/permutation"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/specialqueries"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/keccak/generic"
	"github.com/stretchr/testify/assert"
)

func TestDataTransferModule(t *testing.T) {
	mod := &DataTransferModule{}
	round := 0
	def := generic.PHONEY_RLP
	gbmSize := 512
	maxNumKeccakF := 128
	define := func(build *wizard.Builder) {
		comp := build.CompiledIOP
		mod.GBM = CommitGBM(comp, round, def, gbmSize)
		mod.NewDataTransfer(comp, round, maxNumKeccakF)
	}

	prover := func(run *wizard.ProverRuntime) {
		permTrace := permTrace.PermTraces{}
		gt := generic.GenTrace{}
		AssignGBMfromTable(run, &mod.GBM)
		mod.GBM.AppendTraces(run, &permTrace, &gt)
		mod.AssignModule(run, permTrace, gt)
	}
	compiled := wizard.Compile(define, dummy.Compile)
	proof := wizard.Prove(compiled, prover)

	assert.NoErrorf(t, wizard.Verify(compiled, proof), "invalid proof")
}

func BenchmarkDataTransferModule(b *testing.B) {

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
				mod := &DataTransferModule{}
				mod.GBM = CommitGBM(comp, 0, generic.PHONEY_RLP, 2)
				mod.NewDataTransfer(comp, 0, numKeccakF)
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
