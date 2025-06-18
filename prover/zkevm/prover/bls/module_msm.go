package bls

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/plonk"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

const (
	NAME_BLS_MSM = "BLS_MSM"
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
	AlignedGnarkData *plonk.Alignment
	size             int
	*Limits
	group
}

func newMsm(comp *wizard.CompiledIOP, g group, limits *Limits, src *BlsMsmDataSource, plonkOptions []query.PlonkOption) *BlsMsm {
	size := limits.sizeMulIntegration(g)

	toAlign := &plonk.CircuitAlignmentInput{
		Name:               fmt.Sprintf("%s_%s", NAME_BLS_MSM, g.String()),
		Round:              ROUND_NR,
		DataToCircuitMask:  src.CsMul,
		DataToCircuit:      src.Limb,
		Circuit:            newMulCircuit(g, limits),
		NbCircuitInstances: limits.nbMulCircuitInstances(g),
		PlonkOptions:       plonkOptions,
	}

	return &BlsMsm{
		BlsMsmDataSource: src,
		AlignedGnarkData: plonk.DefineAlignment(comp, toAlign),
		size:             size,
		Limits:           limits,
		group:            g,
	}
}

func (bm *BlsMsm) Assign(run *wizard.ProverRuntime) {
	bm.AlignedGnarkData.Assign(run)
}
