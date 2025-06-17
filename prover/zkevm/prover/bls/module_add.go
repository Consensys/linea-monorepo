package bls

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/plonk"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

const (
	NAME_BLS_ADD = "BLS_ADD"
)

type blsAddDataSource struct {
	CsAdd             ifaces.Column
	CsCurveMembership ifaces.Column
	Limb              ifaces.Column
	Index             ifaces.Column
	Counter           ifaces.Column
	IsData            ifaces.Column
	IsRes             ifaces.Column
}

func newAddDataSource(comp *wizard.CompiledIOP, g group) *blsAddDataSource {
	return &blsAddDataSource{
		CsAdd:             comp.Columns.GetHandle(ifaces.ColIDf("bls.CIRCUIT_SELECTOR_%s_ADD", g.String())),
		CsCurveMembership: comp.Columns.GetHandle(ifaces.ColIDf("bls.CURVE_MEMBERSHIP_%s_ADD", g.StringCurve())),
		Limb:              comp.Columns.GetHandle("bls.LIMB"),
		Index:             comp.Columns.GetHandle("bls.INDEX"),
		Counter:           comp.Columns.GetHandle("bls.CT"),
		IsData:            comp.Columns.GetHandle(ifaces.ColIDf("bls.DATA_%s_ADD", g.String())),
		IsRes:             comp.Columns.GetHandle(ifaces.ColIDf("bls.RSLT_%s_ADD", g.String())),
	}
}

type BlsAdd struct {
	*blsAddDataSource
	*unalignedCurveMembershipData
	alignedAddGnarkData             *plonk.Alignment
	alignedCurveMembershipGnarkData *plonk.Alignment

	size int
	*Limits
	group
}

func newAdd(comp *wizard.CompiledIOP, g group, limits *Limits, src *blsAddDataSource, plonkOptions []query.PlonkOption) *BlsAdd {
	size := limits.sizeAddIntegration(g)
	ucmd := newUnalignedCurveMembershipData(comp, g, size, &unalignedCurveMembershipDataSource{
		Limb:              src.Limb,
		CsCurveMembership: src.CsCurveMembership,
		Counter:           src.Counter,
	})

	toAlignAdd := &plonk.CircuitAlignmentInput{
		Name:               fmt.Sprintf("%s_%s_ALIGNMENT", NAME_BLS_ADD, g.String()),
		Round:              ROUND_NR,
		DataToCircuitMask:  src.CsAdd,
		DataToCircuit:      src.Limb,
		Circuit:            newAddCircuit(g, limits),
		NbCircuitInstances: limits.nbAddCircuitInstances(g),
		PlonkOptions:       plonkOptions,
	}
	toAlignCurveMembership := &plonk.CircuitAlignmentInput{
		Name:               fmt.Sprintf("%s_%s_CURVE_MEMBERSHIP_ALIGNMENT", NAME_BLS_ADD, g.StringCurve()),
		Round:              ROUND_NR,
		DataToCircuitMask:  ucmd.IsActive,
		DataToCircuit:      ucmd.GnarkData,
		Circuit:            newCheckCircuit(g, CURVE, limits),
		NbCircuitInstances: limits.nbCurveMembershipCircuitInstances(g),
		PlonkOptions:       plonkOptions,
		InputFillerKey:     membershipInputFillerKey(g, CURVE),
	}

	res := &BlsAdd{
		blsAddDataSource:                src,
		unalignedCurveMembershipData:    ucmd,
		alignedAddGnarkData:             plonk.DefineAlignment(comp, toAlignAdd),
		alignedCurveMembershipGnarkData: plonk.DefineAlignment(comp, toAlignCurveMembership),
		size:                            size,
		Limits:                          limits,
		group:                           g,
	}

	return res
}

func (ba *BlsAdd) Assign(run *wizard.ProverRuntime) {
	ba.unalignedCurveMembershipData.Assign(run)
	ba.alignedAddGnarkData.Assign(run)
	ba.alignedCurveMembershipGnarkData.Assign(run)
}
