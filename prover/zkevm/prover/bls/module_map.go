package bls

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/plonk"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

const (
	NAME_BLS_MAP = "BLS_MAP"
)

type blsMapDataSource struct {
	ID      ifaces.Column
	CsMap   ifaces.Column
	Limb    ifaces.Column
	Index   ifaces.Column
	Counter ifaces.Column
	IsData  ifaces.Column
	IsRes   ifaces.Column
}

func newMapDataSource(comp *wizard.CompiledIOP, g group) *blsMapDataSource {
	return &blsMapDataSource{
		ID:      comp.Columns.GetHandle("bls.ID"),
		CsMap:   comp.Columns.GetHandle(ifaces.ColIDf("bls.CIRCUIT_SELECTOR_BLS_MAP_%s_TO_%s", g.StringMap(), g.String())),
		Limb:    comp.Columns.GetHandle("bls.LIMB"),
		Index:   comp.Columns.GetHandle("bls.INDEX"),
		Counter: comp.Columns.GetHandle("bls.CT"),
		IsData:  comp.Columns.GetHandle(ifaces.ColIDf("bls.DATA_BLS_MAP_%s_TO_%s", g.StringMap(), g.String())),
		IsRes:   comp.Columns.GetHandle(ifaces.ColIDf("bls.RSLT_BLS_MAP_%s_TO_%s", g.StringMap(), g.String())),
	}
}

type BlsMap struct {
	*blsMapDataSource
	alignedGnarkData *plonk.Alignment
	*Limits
	group
}

func newMap(_ *wizard.CompiledIOP, g group, limits *Limits, src *blsMapDataSource) *BlsMap {
	res := &BlsMap{
		blsMapDataSource: src,
		Limits:           limits,
		group:            g,
	}
	return res
}

func (bm *BlsMap) WithMapCircuit(comp *wizard.CompiledIOP, options ...query.PlonkOption) *BlsMap {
	toAlign := &plonk.CircuitAlignmentInput{
		Name:               fmt.Sprintf("%s_%s_ALIGNMENT", NAME_BLS_MAP, bm.group.String()),
		Round:              ROUND_NR,
		DataToCircuitMask:  bm.blsMapDataSource.CsMap,
		DataToCircuit:      bm.blsMapDataSource.Limb,
		Circuit:            NewMapCircuit(bm.group, bm.Limits),
		NbCircuitInstances: bm.Limits.nbMapCircuitInstances(bm.group),
		InputFillerKey:     mapToGroupInputFillerKey(bm.group),
		PlonkOptions:       options,
	}
	bm.alignedGnarkData = plonk.DefineAlignment(comp, toAlign)
	return bm
}

func (bm *BlsMap) Assign(run *wizard.ProverRuntime) {
	if bm.alignedGnarkData != nil {
		bm.alignedGnarkData.Assign(run)
	}
}
