package keccakfkoalabear

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/assert"
)

const (
	limit = 214358881 // 11^8
)

func TestConvertAndClean(t *testing.T) {
	var (
		size      = 1024
		stateCurr [5][5][8]ifaces.Column
		fromBaseX *convertAndClean
	)

	define := func(b *wizard.Builder) {
		comp := b.CompiledIOP
		/// commit to the input state
		for x := 0; x < 5; x++ {
			for y := 0; y < 5; y++ {
				for z := 0; z < 8; z++ {
					//get a random integer smaller than 11^8
					stateCurr[x][y][z] = comp.InsertCommit(0, ifaces.ColIDf("BC_STATE_CURR_%v_%v_%v", x, y, z), size)

				}
			}
		}

		fromBaseX = newConvertAndClean(b.CompiledIOP, stateCurr)
	}
	prover := func(run *wizard.ProverRuntime) {
		// assign values to input state
		for x := 0; x < 5; x++ {
			for y := 0; y < 5; y++ {
				for z := 0; z < 8; z++ {
					var col = make([]field.Element, size)
					for i := 0; i < size; i++ {
						var a uint64
						if i%2 == 0 {
							a = 10*14641 + 19487171 // 11^7+ 10 * 11^4  ---> 0*4^7+1*4^4 conversion in base 4 clean.
						} else {
							a = 7*14641 + 2*1771561 // 2*11^6 +7*11^4  ---> 1*4^6+1*4^4 conversion in base 4 clean.
						}
						col[i] = field.NewElement(uint64(a))
					}
					run.AssignColumn(stateCurr[x][y][z].GetColID(), smartvectors.NewRegular(col))
				}
			}
		}

		// assign the base conversion module
		fromBaseX.Run(run)
		expected0 := uint64(256)  // 4^4
		expected1 := uint64(4352) // 1*4^6 +1*4^4
		for x := 0; x < 5; x++ {
			for y := 0; y < 5; y++ {
				for z := 0; z < 8; z++ {
					// verify the output values are correct
					actualState := fromBaseX.StateNext[x][y][z].GetColAssignment(run).IntoRegVecSaveAlloc()
					for i := 0; i < size; i++ {
						if i%2 == 0 {
							if actualState[i].Uint64() != expected0 {
								t.Fatalf("Base conversion failed at position (%v,%v,%v) at row %v: expected %v, got %v", x, y, z, i, expected0, actualState[i].Uint64())
							}
						} else {
							if actualState[i].Uint64() != expected1 {
								t.Fatalf("Base conversion failed at position (%v,%v,%v) at row %v: expected %v, got %v", x, y, z, i, expected1, actualState[i].Uint64())
							}
						}
					}
				}
			}
		}
	}

	compiled := wizard.Compile(define, dummy.Compile)
	proof := wizard.Prove(compiled, prover)
	assert.NoErrorf(t, wizard.Verify(compiled, proof), "verifier failed")

}
