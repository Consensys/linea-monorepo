package bls

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/plonk"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

const (
	NAME_BLS_PAIR       = "BLS_PAIR"
	NAME_UNALIGNED_PAIR = "UNALIGNED_BLS_PAIR"
)

type BlsPairDataSource struct {
	ID             ifaces.Column
	CsPair         ifaces.Column
	CsG1Membership ifaces.Column
	CsG2Membership ifaces.Column
	IsData         ifaces.Column
	IsRes          ifaces.Column
	Index          ifaces.Column
	Counter        ifaces.Column
	Limb           ifaces.Column
	SuccessBit     ifaces.Column
}

func newPairDataSource(comp *wizard.CompiledIOP) *BlsPairDataSource {
	return &BlsPairDataSource{
		ID:             comp.Columns.GetHandle("bls.ID"),
		CsPair:         comp.Columns.GetHandle("bls.CIRCUIT_SELECTOR_BLS_PAIRING_CHECK"),
		CsG1Membership: comp.Columns.GetHandle("bls.CIRCUIT_SELECTOR_G1_MEMBERSHIP"),
		CsG2Membership: comp.Columns.GetHandle("bls.CIRCUIT_SELECTOR_G2_MEMBERSHIP"),
		IsData:         comp.Columns.GetHandle("bls.DATA_BLS_PAIRING_CHECK"),
		IsRes:          comp.Columns.GetHandle("bls.RSLT_BLS_PAIRING_CHECK"),
		Index:          comp.Columns.GetHandle("bls.INDEX"),
		Counter:        comp.Columns.GetHandle("bls.CT"),
		Limb:           comp.Columns.GetHandle("bls.LIMB"),
		SuccessBit:     comp.Columns.GetHandle("bls.SUCCESS_BIT"),
	}
}

type BlsPair struct {
	*BlsPairDataSource
	*unalignedPairData
	alignedMillerLoopData        *plonk.Alignment
	alignedFinalExpData          *plonk.Alignment
	alignedG1MembershipGnarkData *plonk.Alignment
	alignedG2MembershipGnarkData *plonk.Alignment
	*Limits
}

func newPair(comp *wizard.CompiledIOP, limits *Limits, src *BlsPairDataSource) *BlsPair {
	ucmd := newUnalignedPairData(comp, limits, src)

	res := &BlsPair{
		BlsPairDataSource: src,
		unalignedPairData: ucmd,
		Limits:            limits,
	}

	return res
}

func (bp *BlsPair) WithPairingCircuit(comp *wizard.CompiledIOP, options ...query.PlonkOption) *BlsPair {
	toAlignMl := &plonk.CircuitAlignmentInput{
		Name:               fmt.Sprintf("%s_ML", NAME_BLS_PAIR),
		Round:              ROUND_NR,
		DataToCircuitMask:  bp.unalignedPairData.GnarkIsActiveMillerLoop,
		DataToCircuit:      bp.unalignedPairData.GnarkDataMillerLoop,
		Circuit:            newMultiMillerLoopMulCircuit(bp.Limits),
		NbCircuitInstances: bp.Limits.NbMillerLoopCircuitInstances,
		// InputFillerKey:     "", // TODO
		PlonkOptions: options,
	}
	toAlignFe := &plonk.CircuitAlignmentInput{
		Name:               fmt.Sprintf("%s_FE", NAME_BLS_PAIR),
		Round:              ROUND_NR,
		DataToCircuitMask:  bp.unalignedPairData.GnarkIsActiveFinalExp,
		DataToCircuit:      bp.unalignedPairData.GnarkDataFinalExp,
		Circuit:            newMultiMillerLoopFinalExpCircuit(bp.Limits),
		NbCircuitInstances: bp.Limits.NbFinalExpCircuitInstances,
		// InputFillerKey:     "", // TODO
		PlonkOptions: options,
	}
	bp.alignedMillerLoopData = plonk.DefineAlignment(comp, toAlignMl)
	bp.alignedFinalExpData = plonk.DefineAlignment(comp, toAlignFe)
	return bp
}

func (bp *BlsPair) WithG1MembershipCircuit(comp *wizard.CompiledIOP, options ...query.PlonkOption) *BlsPair {
	toAlignG1Ms := &plonk.CircuitAlignmentInput{
		Name:               fmt.Sprintf("%s_G1_MEMBERSHIP", NAME_BLS_PAIR),
		Round:              ROUND_NR,
		DataToCircuitMask:  bp.unalignedPairData.GnarkIsActiveG1Membership,
		DataToCircuit:      bp.BlsPairDataSource.Limb,
		Circuit:            newCheckCircuit(G1, GROUP, bp.Limits),
		NbCircuitInstances: bp.Limits.nbGroupMembershipCircuitInstances(G1),
		InputFillerKey:     membershipInputFillerKey(G1, GROUP),
		PlonkOptions:       options,
	}
	bp.alignedG1MembershipGnarkData = plonk.DefineAlignment(comp, toAlignG1Ms)
	return bp
}

func (bp *BlsPair) WithG2MembershipCircuit(comp *wizard.CompiledIOP, options ...query.PlonkOption) *BlsPair {
	toAlignG2Ms := &plonk.CircuitAlignmentInput{
		Name:               fmt.Sprintf("%s_G2_MEMBERSHIP", NAME_BLS_PAIR),
		Round:              ROUND_NR,
		DataToCircuitMask:  bp.unalignedPairData.GnarkIsActiveG2Membership,
		DataToCircuit:      bp.BlsPairDataSource.Limb,
		Circuit:            newCheckCircuit(G2, GROUP, bp.Limits),
		NbCircuitInstances: bp.Limits.nbGroupMembershipCircuitInstances(G2),
		InputFillerKey:     membershipInputFillerKey(G2, GROUP),
		PlonkOptions:       options,
	}
	bp.alignedG2MembershipGnarkData = plonk.DefineAlignment(comp, toAlignG2Ms)
	return bp
}

func (bp *BlsPair) Assign(run *wizard.ProverRuntime) {
	bp.unalignedPairData.Assign(run)
	if bp.alignedMillerLoopData != nil {
		bp.alignedMillerLoopData.Assign(run)
	}
	if bp.alignedFinalExpData != nil {
		bp.alignedFinalExpData.Assign(run)
	}
	if bp.alignedG1MembershipGnarkData != nil {
		bp.alignedG1MembershipGnarkData.Assign(run)
	}
	if bp.alignedG2MembershipGnarkData != nil {
		bp.alignedG2MembershipGnarkData.Assign(run)
	}
}

type unalignedPairData struct {
	*BlsPairDataSource
	IsActive           ifaces.Column
	IsFirstLine        ifaces.Column
	IsLastLine         ifaces.Column
	PointG1            [nbG1Limbs]ifaces.Column
	PointG2            [nbG2Limbs]ifaces.Column
	PrevAccumulator    [nbGtLimbs]ifaces.Column
	CurrentAccumulator [nbGtLimbs]ifaces.Column
	ExpectedResult     [2]ifaces.Column

	GnarkIsActiveG1Membership ifaces.Column
	GnarkIsActiveG2Membership ifaces.Column

	// data which is project from above to columns going into the gnark pairing check circuit
	GnarkIsActiveMillerLoop ifaces.Column
	GnarkDataMillerLoop     ifaces.Column

	GnarkIsActiveFinalExp ifaces.Column
	GnarkDataFinalExp     ifaces.Column
}

func newUnalignedPairData(comp *wizard.CompiledIOP, limits *Limits, src *BlsPairDataSource) *unalignedPairData {
	createColFnMembership := createColFn(comp, NAME_BLS_PAIR, src.SuccessBit.Size())
	createColFnUa := createColFn(comp, NAME_UNALIGNED_PAIR, limits.sizePairUnalignedIntegration())
	createColFnMl := createColFn(comp, NAME_UNALIGNED_PAIR, limits.sizePairMillerLoopIntegration())
	createColFnFe := createColFn(comp, NAME_UNALIGNED_PAIR, limits.sizePairFinalExpIntegration())
	ucmd := &unalignedPairData{
		BlsPairDataSource:         src,
		IsActive:                  createColFnUa("IS_ACTIVE"),
		IsFirstLine:               createColFnUa("IS_FIRST_LINE"),
		IsLastLine:                createColFnUa("IS_LAST_LINE"),
		GnarkIsActiveMillerLoop:   createColFnMl("GNARK_IS_ACTIVE_ML"),
		GnarkDataMillerLoop:       createColFnMl("GNARK_DATA_ML"),
		GnarkIsActiveFinalExp:     createColFnFe("GNARK_IS_ACTIVE_FE"),
		GnarkDataFinalExp:         createColFnFe("GNARK_DATA_FE"),
		GnarkIsActiveG1Membership: createColFnMembership("GNARK_IS_ACTIVE_G1_MEMBERSHIP"),
		GnarkIsActiveG2Membership: createColFnMembership("GNARK_IS_ACTIVE_G2_MEMBERSHIP"),
	}

	for i := range ucmd.PointG1 {
		ucmd.PointG1[i] = createColFnUa(fmt.Sprintf("POINT_G1_%d", i))
	}
	for i := range ucmd.PointG2 {
		ucmd.PointG2[i] = createColFnUa(fmt.Sprintf("POINT_G2_%d", i))
	}
	for i := range ucmd.PrevAccumulator {
		ucmd.PrevAccumulator[i] = createColFnUa(fmt.Sprintf("PREV_ACCUMULATOR_%d", i))
	}
	for i := range ucmd.CurrentAccumulator {
		ucmd.CurrentAccumulator[i] = createColFnUa(fmt.Sprintf("CURRENT_ACCUMULATOR_%d", i))
	}
	for i := range ucmd.ExpectedResult {
		ucmd.ExpectedResult[i] = createColFnUa(fmt.Sprintf("EXPECTED_RESULT_%d", i))
	}

	// TODO: projection from source to unaligned
	// TODO: projection from unaligned to gnark data
	// TODO: that SUCCESS AND MEMBERSHIP is correctly set
	// TODO: first line is correct
	// TODO: last line is correct

	return ucmd
}

func (d *unalignedPairData) Assign(run *wizard.ProverRuntime) {
	d.assignMembershipMask(run)
	d.assignUnaligned(run)
	d.assignGnarkData(run)
}

func (d *unalignedPairData) assignMembershipMask(run *wizard.ProverRuntime) {
	// assigns the masks (CS_G1_MEMBERSHIP AND !SUCCESS_BIT) and
	// (CS_G2_MEMBERSHIP AND !SUCCESS_BIT) columns which are used for filtering
	// the inputs going to group non-membership circuit.
	var (
		srcCsG1Membership = d.CsG1Membership.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcCsG2Membership = d.CsG2Membership.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcSuccessBit     = d.SuccessBit.GetColAssignment(run).IntoRegVecSaveAlloc()
	)

	var (
		dstG1Membership = common.NewVectorBuilder(d.GnarkIsActiveG1Membership)
		dstG2Membership = common.NewVectorBuilder(d.GnarkIsActiveG2Membership)
	)

	for i := range srcCsG1Membership {
		dstG1Membership.PushBoolean(srcCsG1Membership[i].IsOne() && srcSuccessBit[i].IsZero())
	}
	for i := range srcCsG2Membership {
		dstG2Membership.PushBoolean(srcCsG2Membership[i].IsOne() && srcSuccessBit[i].IsZero())
	}

	dstG1Membership.PadAndAssign(run, field.Zero())
	dstG2Membership.PadAndAssign(run, field.Zero())
}

func (d *unalignedPairData) assignUnaligned(run *wizard.ProverRuntime) {
	var (
		srcID         = d.ID.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcSuccessBit = d.SuccessBit.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcLimb       = d.Limb.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcIndex      = d.Index.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcCounter    = d.Counter.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcIsData     = d.IsData.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcIsRes      = d.IsRes.GetColAssignment(run).IntoRegVecSaveAlloc()
		// srcCsPair         = d.CsPair.GetColAssignment(run).IntoRegVecSaveAlloc()
		// srcCsG1Membership = d.CsG1Membership.GetColAssignment(run).IntoRegVecSaveAlloc()
		// srcCsG2Membership = d.CsG2Membership.GetColAssignment(run).IntoRegVecSaveAlloc()
	)

	var (
		dstIsActive    = common.NewVectorBuilder(d.IsActive)
		dstIsFirstLine = common.NewVectorBuilder(d.IsFirstLine)
		dstIsLastLine  = common.NewVectorBuilder(d.IsLastLine)
	)

	var dstPointG1 [nbG1Limbs]*common.VectorBuilder
	var dstPointG2 [nbG2Limbs]*common.VectorBuilder
	var dstPrevAccumulator [nbGtLimbs]*common.VectorBuilder
	var dstCurrentAccumulator [nbGtLimbs]*common.VectorBuilder
	var dstExpectedResult [2]*common.VectorBuilder

	for i := range d.PointG1 {
		dstPointG1[i] = common.NewVectorBuilder(d.PointG1[i])
	}
	for i := range d.PointG2 {
		dstPointG2[i] = common.NewVectorBuilder(d.PointG2[i])
	}
	for i := range d.PrevAccumulator {
		dstPrevAccumulator[i] = common.NewVectorBuilder(d.PrevAccumulator[i])
	}
	for i := range d.CurrentAccumulator {
		dstCurrentAccumulator[i] = common.NewVectorBuilder(d.CurrentAccumulator[i])
	}
	for i := range d.ExpectedResult {
		dstExpectedResult[i] = common.NewVectorBuilder(d.ExpectedResult[i])
	}

	var ptr int

	for ptr < len(srcLimb) {
		// we detect if we're in a new pairing check instance
		if !(srcIsData[ptr].IsOne() && srcIndex[ptr].IsZero() && srcCounter[ptr].IsZero()) {
			ptr++
			continue
		}
		// the inputs which go to the pairing circuit are:
		//  - valid nontrivial input (cs_pairing_check = 1)
		//  - valid half-trivial ((0,Q) or (P,0)) input (cs_pairing_check = 0 and cs_gN_membership = 1 and success_bit = 1)
		//  - valid trivial input (0,0) (cs_pairing_check=0, cs_gn_membership=0, success_bit = 1)
		// in contrast, the inputs which go to the group membership circuit are:
		//  - invalid nontrivial input (cs_gN_membership = 1 and success_bit = 0)
		//
		// now, we don't do anything with the inputs going to the group
		// membership circuits, they are passed directly to the gnark circuit. From the above it implies that we send all
		// inputs with success_bit = 1 to here.

		// get the current ID to know when to stop
		currentID := srcID[ptr]
		// pair input index
		idx := 0
		// we initialize the current accumulator to zero
		prevAccumulator := nativeGtZero()
		var currentAccumulator []field.Element
		for ptr < len(srcID) && srcID[ptr].Equal(&currentID) {
			// if we have success_bit = 0, then we advance the pointer to the
			// next input. But we need to move the pointer according to if we
			// are in data or result part.
			if srcSuccessBit[ptr].IsZero() {
				switch {
				case srcIsData[ptr].IsOne():
					ptr += nbG1Limbs + nbG2Limbs
				case srcIsRes[ptr].IsOne():
					ptr += 2
				}
				continue
			}
			// now, the success_bit is 1. If we are in the data part, then we copy the point limbs,
			// otherwise we copy the expected result limbs.
			if srcIsData[ptr].IsOne() {
				for i := range nbG1Limbs {
					dstPointG1[i].PushField(srcLimb[ptr+i])
				}
				for i := range nbG2Limbs {
					dstPointG2[i].PushField(srcLimb[ptr+nbG1Limbs+i])
				}
				dstIsLastLine.PushZero()
				// compute the next accumulator
				currentAccumulator = nativeMillerLoopAndSum(prevAccumulator, srcLimb[ptr:ptr+nbG1Limbs], srcLimb[ptr+nbG1Limbs:ptr+nbG1Limbs+nbG2Limbs])
				// copy the accumulator limbs
				for i := range nbGtLimbs {
					dstPrevAccumulator[i].PushField(prevAccumulator[i])
					dstCurrentAccumulator[i].PushField(currentAccumulator[i])
				}
				for i := range 2 {
					dstExpectedResult[i].PushZero()
				}
				dstIsActive.PushOne()
				if idx == 0 {
					dstIsFirstLine.PushOne()
				} else {
					dstIsFirstLine.PushZero()
				}
				idx++
				ptr += nbG1Limbs + nbG2Limbs
			} else if srcIsRes[ptr].IsOne() {
				// we are in the result part. However, we don't add new line but reuse the last last line.
				// thus we need to pop the data before pushing the expected result limbs.
				dstIsLastLine.Pop()
				dstIsLastLine.PushOne()
				for i := range 2 {
					dstExpectedResult[i].Pop()
					dstExpectedResult[i].PushField(srcLimb[ptr+i])
				}
				ptr += 2
			}
		}
	}
	dstIsActive.PadAndAssign(run, field.Zero())
	dstIsFirstLine.PadAndAssign(run, field.Zero())
	dstIsLastLine.PadAndAssign(run, field.Zero())
	for i := range d.PointG1 {
		dstPointG1[i].PadAndAssign(run, field.Zero())
	}
	for i := range d.PointG2 {
		dstPointG2[i].PadAndAssign(run, field.Zero())
	}
	for i := range d.PrevAccumulator {
		dstPrevAccumulator[i].PadAndAssign(run, field.Zero())
	}
	for i := range d.CurrentAccumulator {
		dstCurrentAccumulator[i].PadAndAssign(run, field.Zero())
	}
	for i := range d.ExpectedResult {
		dstExpectedResult[i].PadAndAssign(run, field.Zero())
	}
}

func (d *unalignedPairData) assignGnarkData(run *wizard.ProverRuntime) {
	var (
		srcIsActive = d.IsActive.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcPointG1  = make([][]field.Element, nbG1Limbs)
		srcPointG2  = make([][]field.Element, nbG2Limbs)
		srcPrev     = make([][]field.Element, nbGtLimbs)
		srcCurrent  = make([][]field.Element, nbGtLimbs)
		srcExpected = make([][]field.Element, 2)
	)

	for i := range nbG1Limbs {
		srcPointG1[i] = d.PointG1[i].GetColAssignment(run).IntoRegVecSaveAlloc()
	}
	for i := range nbG2Limbs {
		srcPointG2[i] = d.PointG2[i].GetColAssignment(run).IntoRegVecSaveAlloc()
	}
	for i := range nbGtLimbs {
		srcPrev[i] = d.PrevAccumulator[i].GetColAssignment(run).IntoRegVecSaveAlloc()
		srcCurrent[i] = d.CurrentAccumulator[i].GetColAssignment(run).IntoRegVecSaveAlloc()
	}
	for i := range 2 {
		srcExpected[i] = d.ExpectedResult[i].GetColAssignment(run).IntoRegVecSaveAlloc()
	}

	var (
		dstGnarkIsActiveMillerLoop = common.NewVectorBuilder(d.GnarkIsActiveMillerLoop)
		dstGnarkDataMillerLoop     = common.NewVectorBuilder(d.GnarkDataMillerLoop)
		dstGnarkIsActiveFinalExp   = common.NewVectorBuilder(d.GnarkIsActiveFinalExp)
		dstGnarkDataFinalExp       = common.NewVectorBuilder(d.GnarkDataFinalExp)
	)

	dstGnarkIsActiveMillerLoop.PadAndAssign(run, field.Zero())
	dstGnarkDataMillerLoop.PadAndAssign(run, field.Zero())
	dstGnarkIsActiveFinalExp.PadAndAssign(run, field.Zero())
	dstGnarkDataFinalExp.PadAndAssign(run, field.Zero())
}
