package codehashconsistency

import (
	"slices"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

// Assign assigns the columns internally defined in `mod` into `run`.
func (mod Module) Assign(run *wizard.ProverRuntime) {

	var (
		ssInput  = mod.StateSummaryInput
		mchInput = mod.MimcCodeHashInput
	)

	externalSs := struct {
		InitAccountExists  smartvectors.SmartVector
		FinalAccountExists smartvectors.SmartVector
		IsStorage          smartvectors.SmartVector
		InitMiMC           [common.NbLimbU256]smartvectors.SmartVector
		InitKeccakLo       [common.NbLimbU128]smartvectors.SmartVector
		InitKeccakHi       [common.NbLimbU128]smartvectors.SmartVector
		FinalMiMC          [common.NbLimbU256]smartvectors.SmartVector
		FinalKeccakLo      [common.NbLimbU128]smartvectors.SmartVector
		FinalKeccakHi      [common.NbLimbU128]smartvectors.SmartVector
	}{
		InitAccountExists:  ssInput.Account.Initial.Exists.GetColAssignment(run),
		FinalAccountExists: ssInput.Account.Final.Exists.GetColAssignment(run),
		IsStorage:          ssInput.IsStorage.GetColAssignment(run),
	}

	for i := range common.NbLimbU256 {
		externalSs.InitMiMC[i] = ssInput.Account.Initial.MiMCCodeHash[i].GetColAssignment(run)
		externalSs.FinalMiMC[i] = ssInput.Account.Final.MiMCCodeHash[i].GetColAssignment(run)
	}

	for i := range common.NbLimbU128 {
		externalSs.InitKeccakHi[i] = ssInput.Account.Initial.KeccakCodeHash.Hi[i].GetColAssignment(run)
		externalSs.InitKeccakLo[i] = ssInput.Account.Initial.KeccakCodeHash.Lo[i].GetColAssignment(run)
		externalSs.FinalKeccakHi[i] = ssInput.Account.Final.KeccakCodeHash.Hi[i].GetColAssignment(run)
		externalSs.FinalKeccakLo[i] = ssInput.Account.Final.KeccakCodeHash.Lo[i].GetColAssignment(run)
	}

	externalRom := struct {
		IsActive         smartvectors.SmartVector
		IsForConsistency smartvectors.SmartVector
		NewState         [common.NbLimbU256]smartvectors.SmartVector
		CodeHash         [common.NbLimbU256]smartvectors.SmartVector
	}{
		IsActive:         mchInput.IsActive.GetColAssignment(run),
		IsForConsistency: mchInput.IsForConsistency.GetColAssignment(run),
	}

	for i := range common.NbLimbU256 {
		externalRom.NewState[i] = mchInput.NewState[i].GetColAssignment(run)
		externalRom.CodeHash[i] = mchInput.CodeHash[i].GetColAssignment(run)
	}

	var (
		ssData  = make([][2][common.NbLimbU256]field.Element, 0, 2*externalSs.InitMiMC[0].Len())
		romData = make([][2][common.NbLimbU256]field.Element, 0, externalRom.IsForConsistency.Len())
	)

	for i := 0; i < externalSs.InitMiMC[0].Len(); i++ {

		if isStorage := externalSs.IsStorage.Get(i); isStorage.IsOne() {
			continue
		}

		var (
			initAccountExists  = externalSs.InitAccountExists.Get(i)
			finalAccountExists = externalSs.FinalAccountExists.Get(i)
		)

		if initAccountExists.IsOne() {
			var initMimc [common.NbLimbU256]field.Element
			var initKeccak [common.NbLimbU256]field.Element
			for j := range common.NbLimbU256 {
				initMimc[j] = externalSs.InitMiMC[j].Get(i)
			}

			for j := range common.NbLimbU128 {
				initKeccak[j] = externalSs.InitKeccakHi[j].Get(i)
				initKeccak[common.NbLimbU128+j] = externalSs.InitKeccakLo[j].Get(i)
			}

			ssData = append(ssData,
				[2][common.NbLimbU256]field.Element{
					initMimc,
					initKeccak,
				},
			)
		}

		if finalAccountExists.IsOne() {
			var finalMimc [common.NbLimbU256]field.Element
			var finalKeccak [common.NbLimbU256]field.Element
			for j := range common.NbLimbU256 {
				finalMimc[j] = externalSs.FinalMiMC[j].Get(i)
			}

			for j := range common.NbLimbU128 {
				finalKeccak[j] = externalSs.FinalKeccakHi[j].Get(i)
				finalKeccak[common.NbLimbU128+j] = externalSs.FinalKeccakLo[j].Get(i)
			}

			ssData = append(ssData,
				[2][common.NbLimbU256]field.Element{
					finalMimc,
					finalKeccak,
				},
			)
		}
	}

	for i := 0; i < externalRom.NewState[0].Len(); i++ {

		if isActive := externalRom.IsActive.Get(i); isActive.IsZero() {
			break
		}

		isForConsistency := externalRom.IsForConsistency.Get(i)
		isForConsistencyInt := isForConsistency.Uint64()
		isHashEnd := 1 - isForConsistencyInt

		if isHashEnd == 1 {
			continue
		}

		var newState [common.NbLimbU256]field.Element
		var codeHash [common.NbLimbU256]field.Element
		for j := range common.NbLimbU256 {
			newState[j] = externalRom.NewState[j].Get(i)
			codeHash[j] = externalRom.CodeHash[j].Get(i)
		}

		romData = append(romData,
			[2][common.NbLimbU256]field.Element{
				newState,
				codeHash,
			},
		)
	}

	cmp := func(a, b [2][common.NbLimbU256]field.Element) int {
		if res := CmpLimbs(a[1][:], b[1][:]); res != 0 {
			return res
		}
		return CmpLimbs(a[0][:], b[0][:])
	}

	slices.SortFunc(ssData, cmp)
	slices.SortFunc(romData, cmp)
	ssData = slices.Compact(ssData)
	romData = slices.Compact(romData)
	ssData = slices.Clip(ssData)
	romData = slices.Clip(romData)

	assignment := struct {
		IsActive        *common.VectorBuilder
		StateSumKeccak  common.HiLoAssignmentBuilder
		StateSumMiMC    [common.NbLimbU256]*common.VectorBuilder
		RomKeccak       common.HiLoAssignmentBuilder
		RomMiMC         [common.NbLimbU256]*common.VectorBuilder
		RomOngoing      *common.VectorBuilder
		StateSumOngoing *common.VectorBuilder
	}{
		IsActive:        common.NewVectorBuilder(mod.IsActive),
		StateSumKeccak:  common.NewHiLoAssignmentBuilder(mod.StateSumKeccak),
		RomKeccak:       common.NewHiLoAssignmentBuilder(mod.RomKeccak),
		RomOngoing:      common.NewVectorBuilder(mod.RomOngoing),
		StateSumOngoing: common.NewVectorBuilder(mod.StateSumOngoing),
	}

	for i := range common.NbLimbU256 {
		assignment.RomMiMC[i] = common.NewVectorBuilder(mod.RomMiMC[i])
		assignment.StateSumMiMC[i] = common.NewVectorBuilder(mod.StateSumMiMC[i])
	}

	var (
		cRom     int = 0
		cSS      int = 0
		nbRowMax     = len(romData) + len(ssData)
	)

	for i := 0; i < nbRowMax; i++ {

		if cSS >= len(ssData) && cRom >= len(romData) {
			break
		}

		// importantly, we have to account for the fact that romData and/or ssData
		// can perfectly be empty slices. This can happen when a block is full of
		// eoa-transactions only. Therefore, we need to check that cRom and cSS
		// are within bounds. Otherwise, it will panic.
		var (
			romRow = [2][common.NbLimbU256]field.Element{}
			ssRow  = [2][common.NbLimbU256]field.Element{}
		)

		if cRom < len(romData) {
			romRow = romData[cRom]
		}

		if cSS < len(ssData) {
			ssRow = ssData[cSS]
		}

		romCmpSs := cmp(romRow, ssRow)

		assignment.IsActive.PushOne()

		for j := range common.NbLimbU256 {
			assignment.RomMiMC[j].PushField(romRow[0][j])
			assignment.StateSumMiMC[j].PushField(ssRow[0][j])
		}

		for j := range common.NbLimbU128 {
			assignment.RomKeccak.Hi[j].PushField(romRow[1][j])
			assignment.RomKeccak.Lo[j].PushField(romRow[1][common.NbLimbU128+j])

			assignment.StateSumKeccak.Hi[j].PushField(ssRow[1][j])
			assignment.StateSumKeccak.Lo[j].PushField(ssRow[1][common.NbLimbU128+j])
		}

		assignment.RomOngoing.PushBoolean(cRom < len(romData))
		assignment.StateSumOngoing.PushBoolean(cSS < len(ssData))

		newCRom, newCSS := cRom, cSS

		if cRom < len(romData) && romCmpSs <= 0 {
			newCRom = cRom + 1
		}

		if cSS < len(ssData) && romCmpSs >= 0 {
			newCSS = cSS + 1
		}

		if cRom < len(romData) && newCSS >= len(ssData) {
			newCRom = cRom + 1
		}

		if cSS < len(ssData) && newCRom >= len(romData) {
			newCSS = cSS + 1
		}

		cRom, cSS = newCRom, newCSS
	}

	for i := range common.NbLimbU256 {
		assignment.RomMiMC[i].PadAndAssign(run, field.Zero())
		assignment.StateSumMiMC[i].PadAndAssign(run, field.Zero())
	}

	for i := range common.NbLimbU128 {
		assignment.RomKeccak.Hi[i].PadAndAssign(run, field.Zero())
		assignment.RomKeccak.Lo[i].PadAndAssign(run, field.Zero())

		assignment.StateSumKeccak.Hi[i].PadAndAssign(run, field.Zero())
		assignment.StateSumKeccak.Lo[i].PadAndAssign(run, field.Zero())
	}

	assignment.IsActive.PadAndAssign(run, field.Zero())
	assignment.RomOngoing.PadAndAssign(run, field.Zero())
	assignment.StateSumOngoing.PadAndAssign(run, field.Zero())

	for i := range common.NbLimbU128 {
		mod.CptStateSumKeccakLimbsHi[i].Run(run)
		mod.CptStateSumKeccakLimbsLo[i].Run(run)
		mod.CptRomKeccakLimbsHi[i].Run(run)
		mod.CptRomKeccakLimbsLo[i].Run(run)
	}

	mod.CmpStateSumLimbs.Run(run)
	mod.CmpRomLimbs.Run(run)
	mod.CmpRomVsStateSumLimbs.Run(run)
}

// trimLeadingZeros removes any leading zero limbs from a number's representation.
// For example, [0, 0, 5, 10] becomes [5, 10], and [0, 0, 0] becomes [].
// This ensures that numbers like 0 are consistently represented (e.g., as []field.Element{}).
func trimLeadingZeros(limbs []field.Element) []field.Element {
	firstNonZeroIdx := -1
	for i := len(limbs) - 1; i >= 0; i-- {
		if !limbs[i].IsZero() {
			firstNonZeroIdx = i
			break
		}
	}

	if firstNonZeroIdx == -1 {
		return []field.Element{}
	}

	return limbs[:firstNonZeroIdx+1]
}

// CmpLimbs compares two numbers represented as slices of uint64 limbs.
// The most significant limb is expected to be at the highest index.
//
// It returns:
//
//	 0 if a == b
//	 1 if a > b
//	-1 if a < b
func CmpLimbs(a, b []field.Element) int {
	trimmedA := trimLeadingZeros(a)
	trimmedB := trimLeadingZeros(b)

	lenA := len(trimmedA)
	lenB := len(trimmedB)

	if lenA > lenB {
		return 1
	}
	if lenA < lenB {
		return -1
	}

	for i := 0; i < lenA; i++ {
		if trimmedA[i].Cmp(&trimmedB[i]) == 1 {
			return 1
		}
		if trimmedA[i].Cmp(&trimmedB[i]) == -1 {
			return -1
		}
	}

	return 0
}
