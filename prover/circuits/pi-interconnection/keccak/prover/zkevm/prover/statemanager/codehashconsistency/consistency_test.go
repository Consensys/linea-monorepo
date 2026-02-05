package codehashconsistency

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils/csvtraces"
)

// TestCaseGeneration generates testdata csv and does not test anything per se.
// To re-generate the testcases, you need to unskip it.
func TestCaseGeneration(t *testing.T) {

	t.Skip()

	var (
		rng                = rand.New(utils.NewRandSource(67569)) // nolint
		shared             = [][3]field.Element{}
		numSharedRow       = 1
		numStateSummaryRow = 128
		numCodeHashRow     = 128
	)

	randRow := func() [3]field.Element {
		row := [3]field.Element{}
		row[0] = field.PseudoRand(rng)
		row[1] = field.PseudoRandTruncated(rng, 16)
		row[2] = field.PseudoRandTruncated(rng, 16)
		return row
	}

	for i := 0; i < numSharedRow; i++ {
		shared = append(shared, randRow())
	}

	var (
		ssIsActive      = make([]field.Element, 0)
		ssIsStorage     = make([]field.Element, 0)
		ssInitMimc      = make([]field.Element, 0)
		ssInitKeccakHi  = make([]field.Element, 0)
		ssInitKeccakLo  = make([]field.Element, 0)
		ssFinalMimc     = make([]field.Element, 0)
		ssFinalKeccakHi = make([]field.Element, 0)
		ssFinalKeccakLo = make([]field.Element, 0)
	)

	for i := 0; i < numStateSummaryRow; i++ {

		c := rng.IntN(4)

		if i == 0 {
			c = 0
		}

		switch {

		default:

			ssIsActive = append(ssIsActive, field.One())
			ssIsStorage = append(ssIsStorage, field.One())
			ssInitMimc = append(ssInitMimc, ssInitMimc[i-1])
			ssInitKeccakHi = append(ssInitKeccakHi, ssInitKeccakHi[i-1])
			ssInitKeccakLo = append(ssInitKeccakLo, ssInitKeccakLo[i-1])
			ssFinalMimc = append(ssFinalMimc, ssFinalMimc[i-1])
			ssFinalKeccakHi = append(ssFinalKeccakHi, ssFinalKeccakHi[i-1])
			ssFinalKeccakLo = append(ssFinalKeccakLo, ssFinalKeccakLo[i-1])

		case c == 0 || c == 1:

			var newInit, newFinal [3]field.Element

			if c == 0 {
				newInit = randRow()
				newFinal = randRow()
			}

			if c == 1 {
				fmt.Printf("row %v of state summary is shared\n", i)
				newInit = randChoose(rng, shared)
				newFinal = randChoose(rng, shared)
			}

			ssIsActive = append(ssIsActive, field.One())
			ssIsStorage = append(ssIsStorage, field.Zero())
			ssInitMimc = append(ssInitMimc, newInit[0])
			ssInitKeccakHi = append(ssInitKeccakHi, newInit[1])
			ssInitKeccakLo = append(ssInitKeccakLo, newInit[2])
			ssFinalMimc = append(ssFinalMimc, newFinal[0])
			ssFinalKeccakHi = append(ssFinalKeccakHi, newFinal[1])
			ssFinalKeccakLo = append(ssFinalKeccakLo, newFinal[2])

		}
	}

	csvtraces.WriteExplicit(
		files.MustOverwrite("./testdata/state-summary.csv"),
		[]string{"IS_ACTIVE", "IS_STORAGE", "INITIAL_MIMC", "INITIAL_KECCAK_HI", "INITIAL_KECCAK_LO", "FINAL_MIMC", "FINAL_KECCAK_HI", "FINAL_KECCAK_LO"},
		[][]field.Element{ssIsActive, ssIsStorage, ssInitMimc, ssInitKeccakHi, ssInitKeccakLo, ssFinalMimc, ssFinalKeccakHi, ssFinalKeccakLo},
		false,
	)

	/*
		For the codehash module, the rows are generated in reverse order and
		then reverted.
	*/

	var (
		romIsActive  = make([]field.Element, 0)
		romIsHashEnd = make([]field.Element, 0)
		romNewState  = make([]field.Element, 0)
		romKeccakHi  = make([]field.Element, 0)
		romKeccakLo  = make([]field.Element, 0)
	)

	for i := 0; i < numCodeHashRow; i++ {

		c0 := rng.Int64N(4)

		switch {

		case c0 == 0 || c0 == 1 || i == 0:

			row := randRow()

			if c0 == 0 {
				fmt.Printf("row %v of rom is shared\n", i)
				row = randChoose(rng, shared)
			}

			romIsActive = append(romIsActive, field.One())
			romIsHashEnd = append(romIsHashEnd, field.One())
			romNewState = append(romNewState, row[0])
			romKeccakHi = append(romKeccakHi, row[1])
			romKeccakLo = append(romKeccakLo, row[2])

		default:

			romIsActive = append(romIsActive, field.One())
			romIsHashEnd = append(romIsHashEnd, field.Zero())
			romNewState = append(romNewState, field.PseudoRand(rng))
			romKeccakHi = append(romKeccakHi, romKeccakHi[i-1])
			romKeccakLo = append(romKeccakLo, romKeccakLo[i-1])

		}
	}

	slices.Reverse(romIsActive)
	slices.Reverse(romIsHashEnd)
	slices.Reverse(romNewState)
	slices.Reverse(romKeccakHi)
	slices.Reverse(romKeccakLo)

	csvtraces.WriteExplicit(
		files.MustOverwrite("./testdata/mimc-codehash.csv"),
		[]string{"IS_ACTIVE", "IS_HASH_END", "NEW_STATE", "KECCAK_HI", "KECCAK_LO"},
		[][]field.Element{romIsActive, romIsHashEnd, romNewState, romKeccakHi, romKeccakLo},
		false,
	)
}

func randChoose[T any](rand *rand.Rand, slice []T) T {
	return slice[rand.IntN(len(slice))]
}
