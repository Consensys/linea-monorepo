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
	NAME_BLS_MSM       = "BLS_MSM"
	NAME_UNALIGNED_MSM = "UNALIGNED_BLS_MSM"
)

type BlsMsmDataSource struct {
	ID           ifaces.Column
	CsMul        ifaces.Column
	CsMembership ifaces.Column
	Limb         ifaces.Column
	Index        ifaces.Column
	Counter      ifaces.Column
	IsData       ifaces.Column
	IsRes        ifaces.Column
}

func newMsmDataSource(comp *wizard.CompiledIOP, g group) *BlsMsmDataSource {
	return &BlsMsmDataSource{
		ID:           comp.Columns.GetHandle("bls.ID"),
		CsMul:        comp.Columns.GetHandle(ifaces.ColIDf("bls.CIRCUIT_SELECTOR_%s_MSM", g.String())),
		CsMembership: comp.Columns.GetHandle(ifaces.ColIDf("bls.CIRCUIT_SELECTOR_%s_MEMBERSHIP", g.String())),
		Limb:         comp.Columns.GetHandle("bls.LIMB"),
		Index:        comp.Columns.GetHandle("bls.INDEX"),
		Counter:      comp.Columns.GetHandle("bls.CT"),
		IsData:       comp.Columns.GetHandle("bls.IS_BLS_MUL_DATA"),
		IsRes:        comp.Columns.GetHandle("bls.IS_BLS_MUL_RESULT"),
	}
}

type BlsMsm struct {
	*BlsMsmDataSource
	*unalignedMsmData
	AlignedGnarkMsmData             *plonk.Alignment
	AlignedGnarkGroupMembershipData *plonk.Alignment
	size                            int
	*Limits
	group
}

func newMsm(comp *wizard.CompiledIOP, g group, limits *Limits, src *BlsMsmDataSource, plonkOptions []query.PlonkOption) *BlsMsm {
	size := limits.sizeMulIntegration(g)
	umsm := newUnalignedMsmData(comp, g, limits, src)

	toAlignMsm := &plonk.CircuitAlignmentInput{
		Name:               fmt.Sprintf("%s_%s_MSM", NAME_BLS_MSM, g.String()),
		Round:              ROUND_NR,
		DataToCircuitMask:  umsm.GnarkIsActiveMsm,
		DataToCircuit:      umsm.GnarkDataMsm,
		Circuit:            newMulCircuit(g, limits),
		NbCircuitInstances: limits.nbMulCircuitInstances(g),
		PlonkOptions:       plonkOptions,
	}
	toAlignMembership := &plonk.CircuitAlignmentInput{
		Name:               fmt.Sprintf("%s_%s_GROUP_MEMBERSHIP", NAME_BLS_MSM, g.String()),
		Round:              ROUND_NR,
		DataToCircuitMask:  umsm.GnarkIsActiveMembership,
		DataToCircuit:      umsm.GnarkDataMembership,
		Circuit:            newCheckCircuit(g, GROUP, limits),
		NbCircuitInstances: limits.nbGroupMembershipCircuitInstances(g),
		PlonkOptions:       plonkOptions,
	}

	return &BlsMsm{
		BlsMsmDataSource:                src,
		AlignedGnarkMsmData:             plonk.DefineAlignment(comp, toAlignMsm),
		AlignedGnarkGroupMembershipData: plonk.DefineAlignment(comp, toAlignMembership),
		size:                            size,
		Limits:                          limits,
		group:                           g,
	}
}

func (bm *BlsMsm) Assign(run *wizard.ProverRuntime) {
	bm.unalignedMsmData.Assign(run)
	bm.AlignedGnarkMsmData.Assign(run)
}

type unalignedMsmData struct {
	*BlsMsmDataSource
	// this part is used to define the accumulators and indicate if the
	IsActive            ifaces.Column
	Scalar              [nbFrLimbs]ifaces.Column
	Point               []ifaces.Column // length nbG1Limbs or nbG2Limbs
	CurrentAccumulator  []ifaces.Column // length nbG1Limbs or nbG2Limbs
	NextAccumulator     []ifaces.Column // length nbG1Limbs or nbG2Limbs
	ToMsmCircuit        ifaces.Column
	ToMembershipCircuit ifaces.Column

	// data which is projected from above columns going into the MSM circuit
	GnarkIsActiveMsm ifaces.Column
	GnarkDataMsm     ifaces.Column

	// data which is projected from above columns going into the membership circuit
	GnarkIsActiveMembership ifaces.Column
	GnarkDataMembership     ifaces.Column

	group group
}

func newUnalignedMsmData(comp *wizard.CompiledIOP, g group, limits *Limits, src *BlsMsmDataSource) *unalignedMsmData {
	size := limits.sizeMulUnalignedIntegration(g)
	createCol := createColFn(comp, fmt.Sprintf("UNALIGNED_%s_BLS_MSM", g.String()), size)
	res := &unalignedMsmData{
		BlsMsmDataSource:    src,
		IsActive:            createCol("IS_ACTIVE"),
		Point:               make([]ifaces.Column, nbLimbs(g)),
		CurrentAccumulator:  make([]ifaces.Column, nbLimbs(g)),
		NextAccumulator:     make([]ifaces.Column, nbLimbs(g)),
		ToMsmCircuit:        createCol("TO_MSM_CIRCUIT"),
		ToMembershipCircuit: createCol("TO_GROUP_MEMBERSHIP_CIRCUIT"),

		GnarkIsActiveMsm: createCol("GNARK_IS_ACTIVE_MSM"),
		GnarkDataMsm:     createCol("GNARK_DATA_MSM"),

		GnarkIsActiveMembership: createCol("GNARK_IS_ACTIVE_MEMBERSHIP"),
		GnarkDataMembership:     createCol("GNARK_DATA_GROUP_MEMBERSHIP"),
		group:                   g,
	}

	for i := range res.Scalar {
		res.Scalar[i] = createCol(fmt.Sprintf("SCALAR_%d", i))
	}
	for i := range res.Point {
		res.Point[i] = createCol(fmt.Sprintf("POINT_%d", i))
	}
	for i := range res.CurrentAccumulator {
		res.CurrentAccumulator[i] = createCol(fmt.Sprintf("CURRENT_ACCUMULATOR_%d", i))
	}
	for i := range res.NextAccumulator {
		res.NextAccumulator[i] = createCol(fmt.Sprintf("NEXT_ACCUMULATOR_%d", i))
	}

	return res
}

func (d *unalignedMsmData) Assign(run *wizard.ProverRuntime) {
	var (
		srcID      = d.ID.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcLimb    = d.Limb.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcIndex   = d.Index.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcCounter = d.Counter.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcCsMul   = d.CsMul.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcIsData = d.IsData.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcIsRes  = d.IsRes.GetColAssignment(run).IntoRegVecSaveAlloc()
	)

	var (
		dstIsActive = common.NewVectorBuilder(d.IsActive)

		dstIsFirstLine        = common.NewVectorBuilder(d.IsFirstLine)
		dstIsLastLine         = common.NewVectorBuilder(d.IsLastLine)
		dstScalar             = make([]*common.VectorBuilder, nbFrLimbs)
		dstPoint              = make([]*common.VectorBuilder, nbLimbs(d.group))
		dstCurrentAccumulator = make([]*common.VectorBuilder, nbLimbs(d.group))
		dstNextAccumulator    = make([]*common.VectorBuilder, nbLimbs(d.group))
	)
	for i := range dstScalar {
		dstScalar[i] = common.NewVectorBuilder(d.Scalar[i])
	}
	for i := range nbLimbs(d.group) {
		dstPoint[i] = common.NewVectorBuilder(d.Point[i])
		dstCurrentAccumulator[i] = common.NewVectorBuilder(d.CurrentAccumulator[i])
		dstNextAccumulator[i] = common.NewVectorBuilder(d.NextAccumulator[i])
	}

	nbL := nbLimbs(d.group)

	var ptr int

	for ptr < len(srcLimb) {
		// first detect if it is a new MSM instance
		if !(srcIsData[ptr].IsOne() && srcIndex[ptr].IsZero() && srcCounter[ptr].IsZero()) {
			ptr++
			continue
		}
		// we are now on the first line of a new MSM instance. Every MSM
		// instance is ((P1, n1), (P2, n2), ..., (Pk, nk), R), where Pi, ni are
		// the points and scalars and R is the result.

		// we get the current instance ID so that we know when to stop
		currentID := srcID[ptr]
		// MSM input index
		idx := 0
		// we initialize the current accumulator for computing the running sum
		currAccumulator := make([]field.Element, nbL)
		var nextAccumulator []field.Element

		// now we copy the actual
		for ptr < len(srcID) && currentID.Equal(&srcID[ptr]) {
			// we either have input or result limbs. Switch to see if we need to
			// copy point+scalar or only result limbs.
			switch {
			case srcIsData[ptr].IsOne():
				// copy the point limbs
				for i := range nbL {
					dstPoint[i].PushField(srcLimb[ptr+i])
				}
				// copy the scalar limbs
				for i := range nbFrLimbs {
					dstScalar[i].PushField(srcLimb[ptr+nbL+i])
				}
				dstIsLastLine.PushZero()
				// compute the next accumulator
				nextAccumulator = nativeScalarMulAndSum(d.group, currAccumulator, srcLimb[ptr:ptr+nbL], srcLimb[ptr+nbL:ptr+nbL+nbFrLimbs])
				for i := range nbL {
					// copy the next accumulator limbs
					dstNextAccumulator[i].PushField(nextAccumulator[i])
					// we also copy the current accumulator, which is the same as the next
					dstCurrentAccumulator[i].PushField(currAccumulator[i])
				}
				currAccumulator = nextAccumulator
				ptr += nbL + nbFrLimbs
				dstIsActive.PushOne()
			case srcIsRes[ptr].IsOne():
				// if it is the last line then we don't need to copy the result limbs - we have already computed it.
				// its consistency will be checked by gnark circuit and projection queries.
				dstIsLastLine.Pop()
				dstIsLastLine.PushOne()
				ptr += nbL
			}
			if idx == 0 {
				dstIsFirstLine.PushOne()
			}
			idx++
		}
	}

	var (
		dstDataMsm         = common.NewVectorBuilder(d.GnarkDataMsm)
		dstDataIsActiveMsm = common.NewVectorBuilder(d.GnarkIsActiveMsm)
	)

	dstIsActive.PadAndAssign(run, field.Zero())
	dstIsFirstLine.PadAndAssign(run, field.Zero())
	dstIsLastLine.PadAndAssign(run, field.Zero())
	for i := range nbFrLimbs {
		dstScalar[i].PadAndAssign(run, field.Zero())
	}
	for i := range nbLimbs(d.group) {
		dstPoint[i].PadAndAssign(run, field.Zero())
		dstCurrentAccumulator[i].PadAndAssign(run, field.Zero())
		dstNextAccumulator[i].PadAndAssign(run, field.Zero())
	}

	dstDataMsm.PadAndAssign(run, field.Zero())
	dstDataIsActiveMsm.PadAndAssign(run, field.Zero())
}
