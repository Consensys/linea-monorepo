package bls

import (
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/plonk"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

const (
	NAME_BLS_POINTEVAL = "BLS_POINTEVAL"
)

type blsPointEvalDataSource struct {
	ID                 ifaces.Column
	CsPointEval        ifaces.Column
	CsPointEvalInvalid ifaces.Column
	Limb               ifaces.Column
	Index              ifaces.Column
	Counter            ifaces.Column
	IsData             ifaces.Column
	IsRes              ifaces.Column
}

func newPointEvalDataSource(comp *wizard.CompiledIOP) *blsPointEvalDataSource {
	return &blsPointEvalDataSource{
		ID:                 comp.Columns.GetHandle("bls.ID"),
		CsPointEval:        comp.Columns.GetHandle("bls.CIRCUIT_SELECTOR_POINT_EVALUATION"),
		CsPointEvalInvalid: comp.Columns.GetHandle("bls.CIRCUIT_SELECTOR_POINT_EVALUATION_FAILURE"),
		Limb:               comp.Columns.GetHandle("bls.LIMB"),
		Index:              comp.Columns.GetHandle("bls.INDEX"),
		Counter:            comp.Columns.GetHandle("bls.CT"),
		IsData:             comp.Columns.GetHandle("bls.DATA_POINT_EVALUATION"),
		IsRes:              comp.Columns.GetHandle("bls.RSLT_POINT_EVALUATION"),
	}
}

type BlsPointEval struct {
	*blsPointEvalDataSource
	alignedGnarkData        *plonk.Alignment
	alignedFailureGnarkData *plonk.Alignment
	*Limits
}

func newPointEval(_ *wizard.CompiledIOP, limits *Limits, src *blsPointEvalDataSource) *BlsPointEval {
	res := &BlsPointEval{
		blsPointEvalDataSource: src,
		Limits:                 limits,
	}
	return res
}

func (bp *BlsPointEval) WithPointEvalCircuit(comp *wizard.CompiledIOP, options ...query.PlonkOption) *BlsPointEval {
	toAlign := &plonk.CircuitAlignmentInput{
		Name:               NAME_BLS_POINTEVAL,
		Round:              ROUND_NR,
		DataToCircuitMask:  bp.blsPointEvalDataSource.CsPointEval,
		DataToCircuit:      bp.blsPointEvalDataSource.Limb,
		Circuit:            newMultiPointEvalCircuit(bp.Limits),
		NbCircuitInstances: bp.NbPointEvalCircuitInstances,
		InputFillerKey:     pointEvalInputFillerKey,
		PlonkOptions:       options,
	}
	bp.alignedGnarkData = plonk.DefineAlignment(comp, toAlign)
	return bp
}

func (bp *BlsPointEval) WithPointEvalFailureCircuit(comp *wizard.CompiledIOP, options ...query.PlonkOption) *BlsPointEval {
	toAlign := &plonk.CircuitAlignmentInput{
		Name:               NAME_BLS_POINTEVAL + "_FAILURE",
		Round:              ROUND_NR,
		DataToCircuitMask:  bp.blsPointEvalDataSource.CsPointEvalInvalid,
		DataToCircuit:      bp.blsPointEvalDataSource.Limb,
		Circuit:            newMultiPointEvalFailureCircuit(bp.Limits),
		NbCircuitInstances: bp.NbPointEvalFailureCircuitInstances,
		InputFillerKey:     pointEvalFailureInputFillerKey,
		PlonkOptions:       options,
	}
	bp.alignedFailureGnarkData = plonk.DefineAlignment(comp, toAlign)
	return bp
}

func (bp *BlsPointEval) Assign(run *wizard.ProverRuntime) {
	if bp.alignedGnarkData != nil {
		bp.alignedGnarkData.Assign(run)
	}
	if bp.alignedFailureGnarkData != nil {
		bp.alignedFailureGnarkData.Assign(run)
	}
}
