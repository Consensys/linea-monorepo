package codehashconsistency

import (
	"fmt"
	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/csvtraces"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/mimccodehash"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/statesummary"
	"math/big"
	"math/rand/v2"
	"slices"
	"testing"
	"unsafe"
)

func TestConsistency(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		runTestcase(t, "testdata/mimc-codehash.csv", "testdata/state-summary.csv")
	})
	t.Run("empty-rom", func(t *testing.T) {
		runTestcase(t, "testdata/mimc-codehash-empty.csv", "testdata/state-summary.csv")
	})
	t.Run("empty-state-summary", func(t *testing.T) {
		runTestcase(t, "testdata/mimc-codehash.csv", "testdata/state-summary-empty.csv")
	})
}

func runTestcase(t *testing.T, mimcCodeHashCsvPath, stateSummaryCsvPath string) {

	var (
		stateSummary     *statesummary.Module
		mimcCodeHash     *mimccodehash.Module
		consistency      Module
		sizeStateSummary = 128
		sizeMimcCodeHash = 256
	)

	define := func(b *wizard.Builder) {

		stateSummary = &statesummary.Module{
			IsActive:  b.InsertCommit(0, "SS_IS_ACTIVE", sizeStateSummary),
			IsStorage: b.InsertCommit(0, "SS_IS_STORAGE", sizeStateSummary),
			Account: statesummary.AccountPeek{
				Initial: statesummary.Account{
					KeccakCodeHash: common.NewHiLoColumns(b.CompiledIOP, sizeStateSummary, "SS_INITIAL_KECCAK"),
					Exists:         b.InsertCommit(0, "SS_INITIAL_EXISTS", sizeStateSummary),
				},
				Final: statesummary.Account{
					KeccakCodeHash: common.NewHiLoColumns(b.CompiledIOP, sizeStateSummary, "SS_FINAL_KECCAK"),
					Exists:         b.InsertCommit(0, "SS_FINAL_EXISTS", sizeStateSummary),
				},
			},
		}

		mimcCodeHash = &mimccodehash.Module{
			IsActive:         b.InsertCommit(0, "MCH_IS_ACTIVE", sizeMimcCodeHash),
			IsHashEnd:        b.InsertCommit(0, "MCH_IS_HASH_END", sizeMimcCodeHash),
			IsForConsistency: b.InsertCommit(0, "MCH_IS_FOR_CONSISTENCY", sizeMimcCodeHash),
		}

		for i := range common.NbLimbU256 {
			stateSummary.Account.Initial.MiMCCodeHash[i] = b.InsertCommit(0, ifaces.ColIDf("SS_INITIAL_MIMC_%v", i), sizeStateSummary)
			stateSummary.Account.Final.MiMCCodeHash[i] = b.InsertCommit(0, ifaces.ColIDf("SS_FINAL_MIMC_%v", i), sizeStateSummary)

			mimcCodeHash.NewState[i] = b.InsertCommit(0, ifaces.ColIDf("MCH_NEW_STATE_%v", i), sizeMimcCodeHash)
			mimcCodeHash.CodeHash[i] = b.InsertCommit(0, ifaces.ColIDf("MCH_KECCAK_%v", i), sizeMimcCodeHash)

		}

		consistency = NewModule(b.CompiledIOP, "CONSISTENCY", stateSummary, mimcCodeHash)
	}

	prover := func(run *wizard.ProverRuntime) {

		mchCt := csvtraces.MustOpenCsvFile(mimcCodeHashCsvPath)
		ssCt := csvtraces.MustOpenCsvFile(stateSummaryCsvPath)

		run.AssignColumn(mimcCodeHash.IsActive.GetColID(), smartvectors.RightZeroPadded(mchCt.Get("IS_ACTIVE"), sizeMimcCodeHash))
		run.AssignColumn(mimcCodeHash.IsHashEnd.GetColID(), smartvectors.RightZeroPadded(mchCt.Get("IS_HASH_END"), sizeMimcCodeHash))
		run.AssignColumn(mimcCodeHash.IsForConsistency.GetColID(), smartvectors.RightZeroPadded(mchCt.Get("IS_FOR_CONSISTENCY"), sizeMimcCodeHash))

		for i := range common.NbLimbU256 {
			run.AssignColumn(mimcCodeHash.NewState[i].GetColID(), smartvectors.RightZeroPadded(mchCt.Get(fmt.Sprintf("NEW_STATE_%d", i)), sizeMimcCodeHash))
			run.AssignColumn(mimcCodeHash.CodeHash[i].GetColID(), smartvectors.RightZeroPadded(mchCt.Get(fmt.Sprintf("KECCAK_%d", i)), sizeMimcCodeHash))

			run.AssignColumn(stateSummary.Account.Initial.MiMCCodeHash[i].GetColID(), smartvectors.RightZeroPadded(ssCt.Get(fmt.Sprintf("INITIAL_MIMC_%d", i)), sizeStateSummary))
			run.AssignColumn(stateSummary.Account.Final.MiMCCodeHash[i].GetColID(), smartvectors.RightZeroPadded(ssCt.Get(fmt.Sprintf("FINAL_MIMC_%d", i)), sizeStateSummary))
		}

		for i := range common.NbLimbU128 {
			run.AssignColumn(stateSummary.Account.Initial.KeccakCodeHash.Hi[i].GetColID(), smartvectors.RightZeroPadded(ssCt.Get(fmt.Sprintf("INITIAL_KECCAK_HI_%v", i)), sizeStateSummary))
			run.AssignColumn(stateSummary.Account.Final.KeccakCodeHash.Hi[i].GetColID(), smartvectors.RightZeroPadded(ssCt.Get(fmt.Sprintf("FINAL_KECCAK_HI_%v", i)), sizeStateSummary))
			run.AssignColumn(stateSummary.Account.Initial.KeccakCodeHash.Lo[i].GetColID(), smartvectors.RightZeroPadded(ssCt.Get(fmt.Sprintf("INITIAL_KECCAK_LO_%v", i)), sizeStateSummary))
			run.AssignColumn(stateSummary.Account.Final.KeccakCodeHash.Lo[i].GetColID(), smartvectors.RightZeroPadded(ssCt.Get(fmt.Sprintf("FINAL_KECCAK_LO_%v", i)), sizeStateSummary))
		}

		run.AssignColumn(stateSummary.IsActive.GetColID(), smartvectors.RightZeroPadded(ssCt.Get("IS_ACTIVE"), sizeStateSummary))
		run.AssignColumn(stateSummary.IsStorage.GetColID(), smartvectors.RightZeroPadded(ssCt.Get("IS_STORAGE"), sizeStateSummary))
		run.AssignColumn(stateSummary.Account.Initial.Exists.GetColID(), smartvectors.RightZeroPadded(ssCt.Get("INITIAL_EXISTS"), sizeStateSummary))
		run.AssignColumn(stateSummary.Account.Final.Exists.GetColID(), smartvectors.RightZeroPadded(ssCt.Get("FINAL_EXISTS"), sizeStateSummary))

		consistency.Assign(run)
	}

	comp := wizard.Compile(define, dummy.CompileAtProverLvl())
	_ = wizard.Prove(comp, prover)

}

var isForConsistencyMocked = []uint64{1, 1, 0, 0, 1, 1, 1, 0, 1, 0, 1, 1, 1, 1, 0, 1, 1, 0, 1, 1, 1, 1, 1, 0, 1, 0, 0,
	0, 0, 0, 1, 0, 1, 1, 0, 0, 1, 0, 0, 1, 0, 1, 1, 1, 1, 0, 1, 1, 1, 1, 0, 1, 0, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 1, 0,
	1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0, 0, 1, 1, 0, 1, 1, 1, 0, 0, 1, 0, 1, 0, 0, 1, 0, 1, 1, 1, 1,
	1, 0, 0, 0, 0, 0, 1, 1, 1, 0, 1, 0, 0, 0, 0, 0, 1, 1, 1, 0, 1, 1, 1}

// TestCaseGeneration generates testdata csv and does not test anything per se.
// To re-generate the testcases, you need to unskip it.
func TestCaseGeneration(t *testing.T) {

	t.Skip()

	var isForConsistencyMockedElement []field.Element
	for _, num := range isForConsistencyMocked {
		isForConsistencyMockedElement = append(isForConsistencyMockedElement, field.NewElement(num))
	}

	var (
		rng                = rand.New(utils.NewRandSource(67569)) // nolint
		shared             = [][3][]field.Element{}
		numSharedRow       = 1
		numStateSummaryRow = 128
		numCodeHashRow     = 128
	)

	randRow := func() [3][]field.Element {
		row := [3][]field.Element{}
		row0 := GenerateRandomLimbs(common.NbLimbU256, rng)
		row1 := GenerateRandomLimbs(common.NbLimbU128, rng)
		row2 := GenerateRandomLimbs(common.NbLimbU128, rng)

		row[0] = make([]field.Element, common.NbLimbU256)
		copy(row[0], row0)
		row[1] = make([]field.Element, common.NbLimbU128)
		copy(row[1], row1)
		row[2] = make([]field.Element, common.NbLimbU128)
		copy(row[2], row2)
		return row
	}

	for i := 0; i < numSharedRow; i++ {
		shared = append(shared, randRow())
	}

	var (
		ssIsActive      = make([]field.Element, 0)
		ssIsStorage     = make([]field.Element, 0)
		ssInitMimc      = make([][]field.Element, 0)
		ssInitKeccakHi  = make([][]field.Element, 0)
		ssInitKeccakLo  = make([][]field.Element, 0)
		ssFinalMimc     = make([][]field.Element, 0)
		ssFinalKeccakHi = make([][]field.Element, 0)
		ssFinalKeccakLo = make([][]field.Element, 0)
		ssInitialExists = make([]field.Element, 0)
		ssFinalExists   = make([]field.Element, 0)
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

			var newInit, newFinal [3][]field.Element

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

		ssInitialExists = append(ssInitialExists, field.One())
		ssFinalExists = append(ssInitialExists, field.One())
	}

	colNamesStateSummary := []string{"IS_ACTIVE", "IS_STORAGE"}
	for i := range common.NbLimbU256 {
		colNamesStateSummary = append(colNamesStateSummary, fmt.Sprintf("INITIAL_MIMC_%d", i))
	}

	for i := range common.NbLimbU256 {
		colNamesStateSummary = append(colNamesStateSummary, fmt.Sprintf("FINAL_MIMC_%d", i))
	}

	for i := range common.NbLimbU128 {
		colNamesStateSummary = append(colNamesStateSummary, fmt.Sprintf("INITIAL_KECCAK_HI_%d", i))
	}

	for i := range common.NbLimbU128 {
		colNamesStateSummary = append(colNamesStateSummary, fmt.Sprintf("INITIAL_KECCAK_LO_%d", i))
	}

	for i := range common.NbLimbU128 {
		colNamesStateSummary = append(colNamesStateSummary, fmt.Sprintf("FINAL_KECCAK_HI_%d", i))
	}

	for i := range common.NbLimbU128 {
		colNamesStateSummary = append(colNamesStateSummary, fmt.Sprintf("FINAL_KECCAK_LO_%d", i))
	}

	colNamesStateSummary = append(colNamesStateSummary, "INITIAL_EXISTS", "FINAL_EXISTS")

	var colValuesStateSummary = [][]field.Element{ssIsActive, ssIsStorage}
	colValuesStateSummary = append(colValuesStateSummary, transposeLimbs(ssInitMimc)...)
	colValuesStateSummary = append(colValuesStateSummary, transposeLimbs(ssFinalMimc)...)

	colValuesStateSummary = append(colValuesStateSummary, transposeLimbs(ssInitKeccakHi)...)
	colValuesStateSummary = append(colValuesStateSummary, transposeLimbs(ssInitKeccakLo)...)
	colValuesStateSummary = append(colValuesStateSummary, transposeLimbs(ssFinalKeccakHi)...)
	colValuesStateSummary = append(colValuesStateSummary, transposeLimbs(ssFinalKeccakLo)...)

	colValuesStateSummary = append(colValuesStateSummary, ssInitialExists)
	colValuesStateSummary = append(colValuesStateSummary, ssFinalExists)

	csvtraces.WriteExplicit(
		files.MustOverwrite("./testdata/state-summary.csv"),
		colNamesStateSummary,
		colValuesStateSummary,
		false,
	)

	/*
		For the codehash module, the rows are generated in reverse order and
		then reverted.
	*/

	var (
		romIsActive  = make([]field.Element, 0)
		romIsHashEnd = make([]field.Element, 0)
		romNewState  = make([][]field.Element, 0)
		romKeccak    = make([][]field.Element, 0)
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
			romKeccak = append(romKeccak, append(row[1], row[2]...))
		default:

			romIsActive = append(romIsActive, field.One())
			romIsHashEnd = append(romIsHashEnd, field.Zero())
			romNewState = append(romNewState, GenerateRandomLimbs(common.NbLimbU256, rng))
			romKeccak = append(romKeccak, romKeccak[i-1])

		}
	}

	slices.Reverse(romIsActive)
	slices.Reverse(romIsHashEnd)
	slices.Reverse(romNewState)
	slices.Reverse(romKeccak)

	colNamesMimcCodehash := []string{"IS_FOR_CONSISTENCY", "IS_ACTIVE", "IS_HASH_END"}
	for i := range common.NbLimbU256 {
		colNamesMimcCodehash = append(colNamesMimcCodehash, fmt.Sprintf("NEW_STATE_%d", i))
	}

	for i := range common.NbLimbU256 {
		colNamesMimcCodehash = append(colNamesMimcCodehash, fmt.Sprintf("KECCAK_%d", i))
	}

	var colValuesMimcCodehash = [][]field.Element{isForConsistencyMockedElement, romIsActive, romIsHashEnd}
	colValuesMimcCodehash = append(colValuesMimcCodehash, transposeLimbs(romNewState)...)
	colValuesMimcCodehash = append(colValuesMimcCodehash, transposeLimbs(romKeccak)...)

	csvtraces.WriteExplicit(
		files.MustOverwrite("./testdata/mimc-codehash.csv"),
		colNamesMimcCodehash,
		colValuesMimcCodehash,
		false,
	)
}

func randChoose[T any](rand *rand.Rand, slice []T) T {
	return slice[rand.IntN(len(slice))]
}

// GenerateRandomLimbs generates a slice of random limbs of the specified limbs number.
func GenerateRandomLimbs(limbs int, rng *rand.Rand) []field.Element {
	var (
		bigInt    = &big.Int{}
		res       = make([]field.Element, limbs)
		bareU64   = [4]uint64{rng.Uint64(), rng.Uint64(), rng.Uint64(), rng.Uint64()}
		bareBytes = *(*[32]byte)(unsafe.Pointer(&bareU64))
	)

	bigInt.SetBytes(bareBytes[:limbs*common.LimbBytes]).Mod(bigInt, field.Modulus())

	limbBytes := common.SplitBytes(bareBytes[:])
	for i := range limbs {
		res[i].SetBytes(limbBytes[i])
	}

	return res
}

func transposeLimbs(inputMatrix [][]field.Element) [][]field.Element {
	if len(inputMatrix) == 0 || len(inputMatrix[0]) == 0 {
		return [][]field.Element{}
	}

	rows := len(inputMatrix)
	cols := len(inputMatrix[0])

	outputMatrix := make([][]field.Element, cols)
	for i := range outputMatrix {
		outputMatrix[i] = make([]field.Element, rows)
	}

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			outputMatrix[j][i] = inputMatrix[i][j]
		}
	}
	return outputMatrix
}
