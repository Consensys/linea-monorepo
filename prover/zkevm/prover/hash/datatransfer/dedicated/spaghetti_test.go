package dedicated

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/assert"
)

func makeTestCaseIsSpaghetti() (
	define wizard.DefineFunc,
	prover wizard.ProverStep,
) {
	round := 0
	size := 64
	spaghettiSize := 128
	ncol := 6
	nMatrix := 3

	matrix := make([][]ifaces.Column, nMatrix)
	spaghettiOfMatrix := make([]ifaces.Column, nMatrix)
	filter := make([]ifaces.Column, ncol)

	define = func(build *wizard.Builder) {
		comp := build.CompiledIOP

		// declare the columns
		for k := range matrix {
			matrix[k] = make([]ifaces.Column, ncol)
			spaghettiOfMatrix[k] = comp.InsertCommit(round, ifaces.ColIDf("SpaghettiOfMatrix_%v", k), spaghettiSize)
		}

		for i := range filter {
			for k := range matrix {
				matrix[k][i] = comp.InsertCommit(round, ifaces.ColIDf("Matrix_%v_%v", k, i), size)
			}

			filter[i] = comp.InsertCommit(round, ifaces.ColIDf("Filters_%v", i), size)
		}

		// insert the query
		InsertIsSpaghetti(comp, round, ifaces.QueryIDf("IsSpaghetti"),
			matrix, filter, spaghettiOfMatrix, spaghettiSize)
	}

	// assign matrix and filter (spaghettiOfMatrix is assigned by the query itself).
	prover = func(run *wizard.ProverRuntime) {

		matrixWit := make([][][]field.Element, nMatrix)
		filtersWit := make([][]field.Element, ncol)
		witSize := 7

		for k := range matrix {
			matrixWit[k] = make([][]field.Element, ncol)
			for j := range filter {
				matrixWit[k][j] = make([]field.Element, witSize)
				for i := 0; i < witSize; i++ {
					nBig, _ := rand.Int(rand.Reader, big.NewInt(int64(1024)))
					a := nBig.Uint64()
					matrixWit[k][j][i] = field.NewElement(a)
				}
			}
		}

		for j := range filter {
			filtersWit[j] = make([]field.Element, witSize)
			for i := 0; i < witSize; i++ {

				// Such a filter has the right form for the query;
				// starting with zeroes ending with ones
				if i%7 == 0 {
					filtersWit[j][i] = field.One()
				}

			}
		}

		for j := range filter {
			run.AssignColumn(filter[j].GetColID(), smartvectors.RightZeroPadded(filtersWit[j], size))
			for k := range matrix {
				run.AssignColumn(matrix[k][j].GetColID(), smartvectors.RightZeroPadded(matrixWit[k][j], size))
			}
		}

		for j := range matrix {
			spaghetti := makeSpaghetti(filtersWit, matrixWit[j])
			run.AssignColumn(spaghettiOfMatrix[j].GetColID(), smartvectors.RightZeroPadded(spaghetti[0], spaghettiSize))
		}

	}
	return define, prover
}
func TestIsSpaghetti(t *testing.T) {
	define, prover := makeTestCaseIsSpaghetti()
	comp := wizard.Compile(define, dummy.Compile)

	proof := wizard.Prove(comp, prover)
	assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")
}

// It receives multiple matrices and a filter, it returns the spaghetti form of the matrices
func makeSpaghetti(filter [][]field.Element, matrix ...[][]field.Element) (spaghetti [][]field.Element) {
	spaghetti = make([][]field.Element, len(matrix))

	// populate spaghetties
	for i := range filter[0] {
		for j := range filter {
			if filter[j][i].Uint64() == 1 {
				for k := range matrix {
					spaghetti[k] = append(spaghetti[k], matrix[k][j][i])
				}
			}

		}
	}
	return spaghetti
}
