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
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf_koalabear/common"
	"github.com/stretchr/testify/assert"
)

func TestChi(t *testing.T) {

	const numCases int = 1
	maxKeccaf := 10

	// #nosec G404 --we don't need a cryptographic RNG for testing purpose
	rnd := rand.New(rand.NewChaCha8([32]byte{}))

	// Every time the prover function is called, the traces will be updated.
	// Likewise, run will be set by the prover.
	var run *wizard.ProverRuntime

	// Parametrizes the wizard and the input generator.
	builder, prover, mod := chiTestingModule(maxKeccaf)

	comp := wizard.Compile(builder, dummy.Compile)

	for i := 0; i < numCases; i++ {

		// Generate new traces
		traces := genKeccakfTrace(rnd, maxKeccaf)

		// Recall that this will set the values of `traces` and `run`
		proof := wizard.Prove(comp, prover(traces, &run))
		assert.NoErrorf(t, wizard.Verify(comp, proof), "verifier failed")

		effNumKeccak := len(traces.KeccakFInps)
		for permId := 0; permId < effNumKeccak; permId++ {
			// Get the input state for this permutation
			state := traces.KeccakFInps[permId]
			for round := 0; round < keccak.NumRound; round++ {

				state.ApplyKeccakfRound(round)
				expectedAIota := state
				// In that case, aIOTA should be in fact the next input because
				// the iota step will be responsible for xoring in the inputs.
				if round == 23 && permId+1 < len(traces.KeccakFInps) {
					if !traces.IsNewHash[permId+1] {
						expectedAIota = traces.KeccakFInps[permId+1]
					}
				}

				// Reconstruct the same state from the assignment of the prover
				reconstructed := keccak.State{}
				stateBits := [64]field.Element{}
				for x := 0; x < 5; x++ {
					for y := 0; y < 5; y++ {
						for z := 0; z < numSlices; z++ {
							k := mod.ChiIota.StateNext[x][y][z].GetColAssignmentAt(run, permId*keccak.NumRound+round)
							res := common.CleanBaseChi(common.DecomposeU64(k.Uint64(), BaseChi, numSlices))
							for j := range res {
								stateBits[z*numSlices+j] = res[j]
							}
						}
						reconstructed[x][y] = reconstructU64(stateBits)
					}
				}
				assert.Equal(t, expectedAIota, reconstructed,
					"could not reconstruct the state. permutation %v", permId)
				// Exiting on the first failed case to not spam the test logs

				if t.Failed() {
					t.Fatalf("stopping here as we encountered errors")
				}
			}
		}
	}
}

// a module definition method specifically for testing the rho submodule
func chiTestingModule(
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
		mod          = &Module{}
		size         = int(utils.NextPowerOfTwo(uint64(maxNumKeccakf) * 24))
		stateCurr    = stateInBits{} // input to the rho module
		blocks       = [numLanesInBlock][numSlices]ifaces.Column{}
		isBlockOther ifaces.Column
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
					stateCurr[x][y][z] = comp.InsertCommit(0, ifaces.ColIDf("CHI_STATE_CURR_%v_%v_%v", x, y, z), size)
					if z < numSlices && (5*y)+x < numLanesInBlock {
						blocks[(5*y)+x][z] = comp.InsertCommit(0, ifaces.ColIDf("MESSAGE_BLOCK_%v_%v", (5*y)+x, z), size)
					}
				}
			}
		}

		isBlockOther = comp.InsertCommit(0, "IS_BLOCK_OTHER", size)

		mod.ChiIota = newChi(comp, chiInputs{
			stateCurr:    stateCurr,
			blocks:       blocks,
			isBlockOther: isBlockOther,
			numKeccakf:   maxNumKeccakf,
		})
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
			blockWit := [numLanesInBlock][numSlices][]field.Element{}
			isBlockOtherWit := []field.Element{}
			blockNext := [numLanesInBlock][numSlices]field.Element{}
			for permId := 0; permId < numKeccakf; permId++ {
				state := traces.KeccakFInps[permId]
				block := cleanBaseBlock(traces.Blocks[permId])

				if permId+1 < numKeccakf {
					blockNext = cleanBaseBlock(traces.Blocks[permId+1])
				}

				for rnd := 0; rnd < keccak.NumRound; rnd++ {
					// Pre-permute using the theta transformation before running
					// the rho permutation.
					state.Theta()
					state.Rho()
					inputState := state.Pi()

					// create the block columns
					for i := 0; i < numLanesInBlock; i++ {
						for j := 0; j < numSlices; j++ {
							switch {
							case rnd == 0 && traces.IsNewHash[permId] == true:
								blockWit[i][j] = append(blockWit[i][j], block[i][j])
							case rnd == 23 && permId+1 < numKeccakf && traces.IsNewHash[permId+1] == false:
								blockWit[i][j] = append(blockWit[i][j], blockNext[i][j])
							default:
								blockWit[i][j] = append(blockWit[i][j], field.Zero())
							}
						}
					}

					if rnd == 23 && permId+1 < numKeccakf && traces.IsNewHash[permId+1] == false {
						isBlockOtherWit = append(isBlockOtherWit, field.One())
					} else {
						isBlockOtherWit = append(isBlockOtherWit, field.Zero())
					}

					// Convert the state in sliced from in base 2
					for x := 0; x < 5; x++ {
						for y := 0; y < 5; y++ {
							a := BitsLE(inputState[x][y])

							for k := 0; k < 64; k++ {
								// If the column is not already assigned, then
								// allocate it with the proper length.
								if stateCurrWit[x][y][k] == nil {
									stateCurrWit[x][y][k] = make([]field.Element, size)
								}

								r := keccak.NumRound*permId + rnd
								stateCurrWit[x][y][k][r] = field.NewElement(uint64(a[k]))
							}
						}
					}

					// Then finalize the permutation normally
					state.Chi(&inputState)
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

						if k < numSlices && (5*y)+x < numLanesInBlock {
							run.AssignColumn(
								blocks[(5*y)+x][k].GetColID(),
								smartvectors.RightZeroPadded(
									blockWit[(5*y)+x][k],
									size,
								),
							)
						}
					}
				}
			}

			run.AssignColumn(isBlockOther.GetColID(), smartvectors.RightZeroPadded(isBlockOtherWit, size))

			// Then assigns all the columns of the rho module
			mod.ChiIota.assignChi(run, stateCurr)
		}
	}

	return builder, prover, mod
}

func cleanBaseBlock(block keccak.Block) (res [numLanesInBlock][numSlices]field.Element) {
	eleven := field.NewElement(BaseChi)
	// extract the byte of each lane, in little endian
	for i := 0; i < numLanesInBlock; i++ {
		lanebytes := [numSlices]uint8{}
		for j := 0; j < numSlices; j++ {
			lanebytes[j] = uint8((block[i] >> (numSlices * j)) & 0xff)
		}
		// convert each byte to clean base BaseChi
		for j := 0; j < numSlices; j++ {
			res[i][j] = keccakf.U64ToBaseX(uint64(lanebytes[j]), &eleven)
		}
	}
	return res
}
