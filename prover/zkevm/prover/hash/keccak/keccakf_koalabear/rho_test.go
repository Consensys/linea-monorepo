package keccakfkoalabear

import (
	"math/rand/v2"
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf"
	"github.com/stretchr/testify/assert"
)

// a module definition method specifically for testing the rho submodule
func rhoTestingModule(
	// parameters for the wizard
	maxNumKeccakf int,
) (
	wizard.DefineFunc, // the define function of testing wizard
	func(
		traces keccak.PermTraces,
		runRet **wizard.ProverRuntime,
	) wizard.MainProverStep,
	*Module,
) {

	// The module is only used a placeholder to let us the `assignInput`
	// function
	var (
		mod       = &Module{}
		size      = int(utils.NextPowerOfTwo(uint64(maxNumKeccakf) * 24))
		stateCurr = stateInBits{} // input to the rho module
	)

	/*
		Initializes the builder function
	*/
	builder := func(b *wizard.Builder) {
		comp := b.CompiledIOP

		// Initializes the input current state
		for x := 0; x < 5; x++ {
			for y := 0; y < 5; y++ {
				for z := 0; z < 64; z++ {
					stateCurr[x][y][z] = comp.InsertCommit(0, ifaces.ColIDf("RHO_STATE_CURR", x, y, z), size)
				}
			}
		}

		mod.RhoPi = newRho(comp, maxNumKeccakf, stateCurr)
	}

	prover := func(
		traces keccak.PermTraces,
		// pointers to access the runtime once they are available, when the
		// prover has been run. This allows the caller test to "open" the box
		// and checks that the assigned columns are consistent with the traces.
		runRet **wizard.ProverRuntime,
	) wizard.MainProverStep {
		return func(run *wizard.ProverRuntime) {
			*runRet = run

			// Number of permutation used for the current instance
			numKeccakf := len(traces.KeccakFInps)

			// If the number of keccakf constraints is larger than what the
			// module is sized for, then, we cannot prove everything.
			if numKeccakf > maxNumKeccakf {
				utils.Panic("Too many keccakf %v > %v", numKeccakf, maxNumKeccakf)
			}

			// Initializes the input columns
			stateCurrWit := [5][5][64][]field.Element{}
			for permId := 0; permId < numKeccakf; permId++ {
				state := traces.KeccakFInps[permId]

				for rnd := 0; rnd < keccak.NumRound; rnd++ {
					// Pre-permute using the theta transformation before running
					// the rho permutation.
					state.Theta()

					// Convert the state in sliced from in base 2
					for x := 0; x < 5; x++ {
						for y := 0; y < 5; y++ {
							a := BitsLE(state[x][y])
							for k := 0; k < 64; k++ {
								// If the column is not already assigned, then
								// allocate it with the proper length.
								if stateCurrWit[x][y][k] == nil {
									stateCurrWit[x][y][k] = make(
										[]field.Element,
										size,
									)
								}

								r := keccak.NumRound*permId + rnd
								stateCurrWit[x][y][k][r] = field.NewElement(uint64(a[k]))
							}
						}
					}

					// Then finalize the permutation normally
					state.Rho()
					b := state.Pi()
					state.Chi(&b)
					state.Iota(rnd)
				}
			}

			for x := 0; x < 5; x++ {
				for y := 0; y < 5; y++ {
					for k := 0; k < 64; k++ {
						run.AssignColumn(
							stateCurr[x][y][k].GetColID(),
							smartvectors.RightZeroPadded(
								stateCurrWit[x][y][k],
								size,
							),
						)
					}
				}
			}

			// Then assigns all the columns of the rho module
			mod.RhoPi.assignRoh(run, stateCurr)
		}
	}

	return builder, prover, mod
}

// Test the correctness of the theta wizard function
func TestRho(t *testing.T) {

	const numCases int = 30
	maxKeccaf := 10

	// #nosec G404 --we don't need a cryptographic RNG for testing purpose
	rnd := rand.New(rand.NewChaCha8([32]byte{}))

	// Every time the prover function is called, the traces will be updated.
	// Likewise, run will be set by the prover.
	var run *wizard.ProverRuntime

	// Parametrizes the wizard and the input generator.
	builder, prover, mod := rhoTestingModule(maxKeccaf)

	comp := wizard.Compile(builder, dummy.Compile)

	for i := 0; i < numCases; i++ {

		// Generate new traces
		traces := genKeccakfTrace(rnd, maxKeccaf)

		// Recall that this will set the values of `traces` and ``
		proof := wizard.Prove(comp, prover(traces, &run))
		assert.NoErrorf(t, wizard.Verify(comp, proof), "verifier failed")

		effNumKeccak := len(traces.KeccakFInps)

		for permId := 0; permId < effNumKeccak; permId++ {

			// Copy the corresponding input state and apply the rho
			// transformation.
			expectedStateRho := traces.KeccakFInps[permId]
			expectedStateRho.Theta()
			expectedStateRho.Rho()

			// Reconstruct the same state from the assignment of the prover
			reconstructed := keccak.State{}
			recomposed := [8]field.Element{}
			for x := 0; x < 5; x++ {
				for y := 0; y < 5; y++ {
					for z := 0; z < 8; z++ {
						// Recompose the slice into a complete base 2 representation
						// of aTheta[x][y] in base 2
						recomposed[z] = mod.RhoPi.stateNext[x][y][z].GetColAssignmentAt(
							run,
							permId*keccak.NumRound,
						)
					}
					reconstructed[x][y] = reconstructU64(recomposed)
				}
			}

			assert.Equal(t, expectedStateRho, reconstructed,
				"could not reconstruct the state. permutation %v", permId)

			// Exiting on the first failed case to not spam the test logs
			if t.Failed() {
				t.Fatalf("stopping here as we encountered errors")
			}
		}
	}
}

func BitsLE(x uint64) [64]uint8 {
	var bits [64]uint8
	for i := 0; i < 64; i++ {
		bits[i] = uint8((x >> i) & 1)
	}
	return bits
}

// Input provider for the tests. Return traces corresponding to random hashes
func genKeccakfTrace(rnd *rand.Rand, maxNumKeccakf int) keccak.PermTraces {

	res := keccak.PermTraces{}
	// The number of effective permutation is a random fraction of the
	// max number of keccakf.
	effNumKeccak := rnd.Int() % maxNumKeccakf
	for effNumKeccak == 0 {
		effNumKeccak = rnd.Int() % maxNumKeccakf
	}

	for len(res.Blocks) < effNumKeccak {
		// Each hash is for a random string taking at most 3 permutations
		streamLen := rnd.IntN(3*keccak.Rate-1) + 1
		stream := make([]byte, streamLen)
		utils.ReadPseudoRand(rnd, stream)

		for i := 0; i < effNumKeccak; i++ {
			keccak.Hash(stream, &res)
		}
	}

	return res

}

func reconstructU64(slices [8]field.Element) uint64 {
	a := []field.Element{}
	var res uint64
	for _, slice := range slices {
		decomposedF := keccakf.DecomposeFr(slice, 11, 64)
		a = append(a, decomposedF...)
	}

	for i, limb := range a {
		bit := (limb.Uint64()) & 1
		res |= bit << i
	}
	return res
}
