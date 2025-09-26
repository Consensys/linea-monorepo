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
	ID                ifaces.Column
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
		ID:                comp.Columns.GetHandle("bls.ID"),
		CsAdd:             comp.Columns.GetHandle(ifaces.ColIDf("bls.CIRCUIT_SELECTOR_BLS_%s_ADD", g.String())),
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
	alignedAddGnarkData             *plonk.Alignment
	alignedCurveMembershipGnarkData *plonk.Alignment

	*Limits
	group
}

func newAdd(_ *wizard.CompiledIOP, g group, limits *Limits, src *blsAddDataSource) *BlsAdd {
	res := &BlsAdd{
		blsAddDataSource: src,
		Limits:           limits,
		group:            g,
	}

	return res
}

func (ba *BlsAdd) WithAddCircuit(comp *wizard.CompiledIOP, options ...query.PlonkOption) *BlsAdd {
	// the gnark circuit takes exactly the same rows as provided by the arithmetization. So
	// to get the bound on the number of circuits we just need to divide by the size of the
	// addition circuit input instances
	maxNbInstances := ba.blsAddDataSource.CsAdd.Size() / nbRowsPerAdd(ba.group)
	maxNbCircuits := maxNbInstances/ba.Limits.nbAddInputInstances(ba.group) + 1

	toAlignAdd := &plonk.CircuitAlignmentInput{
		Name:               fmt.Sprintf("%s_%s_ALIGNMENT", NAME_BLS_ADD, ba.group.String()),
		Round:              ROUND_NR,
		DataToCircuitMask:  ba.blsAddDataSource.CsAdd,
		DataToCircuit:      ba.blsAddDataSource.Limb,
		Circuit:            newAddCircuit(ba.group, ba.Limits),
		NbCircuitInstances: maxNbCircuits,
		PlonkOptions:       options,
	}
	ba.alignedAddGnarkData = plonk.DefineAlignment(comp, toAlignAdd)
	return ba
}

func (ba *BlsAdd) WithCurveMembershipCircuit(comp *wizard.CompiledIOP, options ...query.PlonkOption) *BlsAdd {
	maxNbCircuits := ba.blsAddDataSource.CsCurveMembership.Size() / ba.Limits.nbCurveMembershipInputInstances(ba.group)

	toAlignCurveMembership := &plonk.CircuitAlignmentInput{
		Name:               fmt.Sprintf("%s_%s_CURVE_MEMBERSHIP_ALIGNMENT", NAME_BLS_ADD, ba.group.StringCurve()),
		Round:              ROUND_NR,
		DataToCircuitMask:  ba.blsAddDataSource.CsCurveMembership,
		DataToCircuit:      ba.blsAddDataSource.Limb,
		Circuit:            newCheckCircuit(ba.group, CURVE, ba.Limits),
		NbCircuitInstances: maxNbCircuits,
		PlonkOptions:       options,
		InputFillerKey:     membershipInputFillerKey(ba.group, CURVE),
	}
	ba.alignedCurveMembershipGnarkData = plonk.DefineAlignment(comp, toAlignCurveMembership)
	return ba
}

func (ba *BlsAdd) Assign(run *wizard.ProverRuntime) {
	if ba.alignedAddGnarkData != nil {
		ba.alignedAddGnarkData.Assign(run)
	}
	if ba.alignedCurveMembershipGnarkData != nil {
		ba.alignedCurveMembershipGnarkData.Assign(run)
	}
}
