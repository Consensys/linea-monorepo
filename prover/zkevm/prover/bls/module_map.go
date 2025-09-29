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
	var mapString string
	if g == G1 {
		mapString = "MAP_FP_TO_G1"
	} else {
		mapString = "MAP_FP2_TO_G2"
	}
	return &blsMapDataSource{
		ID:      comp.Columns.GetHandle(colNameFn("ID")),
		CsMap:   comp.Columns.GetHandle(colNameFn("CIRCUIT_SELECTOR_BLS_" + mapString)),
		Index:   comp.Columns.GetHandle(colNameFn("INDEX")),
		Counter: comp.Columns.GetHandle(colNameFn("CT")),
		Limb:    comp.Columns.GetHandle(colNameFn("LIMB")),
		IsData:  comp.Columns.GetHandle(colNameFn("DATA_BLS_" + mapString + "_FLAG")),
		IsRes:   comp.Columns.GetHandle(colNameFn("RSLT_BLS_" + mapString + "_FLAG")),
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
	// gnark circuits takes the inputs as is. To get the bound on the number of circuits we need
	// to divide the maximum number of inputs instances with the number of instances per circuit
	maxNbInstances := bm.blsMapDataSource.CsMap.Size() / nbRowsPerMap(bm.group)
	maxNbCircuits := maxNbInstances/bm.Limits.nbMapInputInstances(bm.group) + 1
	toAlign := &plonk.CircuitAlignmentInput{
		Name:               fmt.Sprintf("%s_%s_ALIGNMENT", NAME_BLS_MAP, bm.group.String()),
		Round:              ROUND_NR,
		DataToCircuitMask:  bm.blsMapDataSource.CsMap,
		DataToCircuit:      bm.blsMapDataSource.Limb,
		Circuit:            NewMapCircuit(bm.group, bm.Limits),
		NbCircuitInstances: maxNbCircuits,
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
