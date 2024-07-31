package datatransfer

import (
	"fmt"
	"sync"
	"testing"

	permTrace "github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/innerproduct"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/lookup"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/permutation"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/specialqueries"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/stretchr/testify/assert"
)

func TestDataTransferModule(t *testing.T) {
	mod := &Module{}
	round := 0
	def := generic.PHONEY_RLP
	gbmSize := 512
	maxNumKeccakF := 128
	define := func(build *wizard.Builder) {
		comp := build.CompiledIOP
		mod.Provider = CommitGBM(comp, round, def, gbmSize)
		mod.NewDataTransfer(comp, round, maxNumKeccakF, 0)
	}

	prover := func(run *wizard.ProverRuntime) {
		permTrace := permTrace.PermTraces{}
		gt := generic.GenTrace{}
		AssignGBMfromTable(run, &mod.Provider)
		mod.Provider.AppendTraces(run, &gt, &permTrace)
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
				mod := &Module{}
				mod.Provider = CommitGBM(comp, 0, generic.PHONEY_RLP, 2)
				mod.NewDataTransfer(comp, 0, numKeccakF, 0)
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
