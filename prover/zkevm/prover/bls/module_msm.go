package bls

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/plonk"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

const (
	NAME_BLS_MSM       = "BLS_MSM"
	NAME_UNALIGNED_MSM = "UNALIGNED_BLS_MSM"
)

type BlsMsmDataSource struct {
	CsMul  ifaces.Column
	Limb   ifaces.Column
	Index  ifaces.Column
	IsData ifaces.Column
	IsRes  ifaces.Column
}

func newMsmDataSource(comp *wizard.CompiledIOP, g group) *BlsMsmDataSource {
	var selectorCs ifaces.Column
	switch g {
	case G1:
		selectorCs = comp.Columns.GetHandle("bls.CIRCUIT_SELECTOR_BLS_G1_MSM")
	case G2:
		selectorCs = comp.Columns.GetHandle("bls.CIRCUIT_SELECTOR_BLS_G2_MSM")
	default:
		panic("unknown group for bls msm data source")
	}
	return &BlsMsmDataSource{
		CsMul:  selectorCs,
		Limb:   comp.Columns.GetHandle("bls.LIMB"),
		Index:  comp.Columns.GetHandle("bls.INDEX"),
		IsData: comp.Columns.GetHandle("bls.IS_BLS_MUL_DATA"),
		IsRes:  comp.Columns.GetHandle("bls.IS_BLS_MUL_RESULT"),
	}
}

type BlsMsm struct {
	*BlsMsmDataSource
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

}
