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
		ID:                 comp.Columns.GetHandle(colNameFn("ID")),
		CsPointEval:        comp.Columns.GetHandle(colNameFn("CIRCUIT_SELECTOR_POINT_EVALUATION")),
		CsPointEvalInvalid: comp.Columns.GetHandle(colNameFn("CIRCUIT_SELECTOR_POINT_EVALUATION_FAILURE")),
		Limb:               comp.Columns.GetHandle(colNameFn("LIMB")),
		Index:              comp.Columns.GetHandle(colNameFn("INDEX")),
		Counter:            comp.Columns.GetHandle(colNameFn("CT")),
		IsData:             comp.Columns.GetHandle(colNameFn("DATA_POINT_EVALUATION_FLAG")),
		IsRes:              comp.Columns.GetHandle(colNameFn("RSLT_POINT_EVALUATION_FLAG")),
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
	// the gnark circuit takes exactly the same rows as provided by the arithmetization. So
	// to get the bound on the number of circuits we just need to divide by the size of the
	// addition circuit input instances
	maxNbInstances := bp.blsPointEvalDataSource.CsPointEval.Size() / nbRowsPerPointEval
	maxNbCircuits := maxNbInstances/bp.Limits.NbPointEvalInputInstances + 1
	toAlign := &plonk.CircuitAlignmentInput{
		Name:               NAME_BLS_POINTEVAL,
		Round:              ROUND_NR,
		DataToCircuitMask:  bp.blsPointEvalDataSource.CsPointEval,
		DataToCircuit:      bp.blsPointEvalDataSource.Limb,
		Circuit:            newMultiPointEvalCircuit(bp.Limits),
		NbCircuitInstances: maxNbCircuits,
		InputFillerKey:     pointEvalInputFillerKey,
		PlonkOptions:       options,
	}
	bp.alignedGnarkData = plonk.DefineAlignment(comp, toAlign)
	return bp
}

func (bp *BlsPointEval) WithPointEvalFailureCircuit(comp *wizard.CompiledIOP, options ...query.PlonkOption) *BlsPointEval {
	maxNbInstances := bp.blsPointEvalDataSource.CsPointEvalInvalid.Size() / nbRowsPerPointEval
	maxNbCircuits := maxNbInstances/bp.Limits.NbPointEvalFailureInputInstances + 1
	toAlign := &plonk.CircuitAlignmentInput{
		Name:               NAME_BLS_POINTEVAL + "_FAILURE",
		Round:              ROUND_NR,
		DataToCircuitMask:  bp.blsPointEvalDataSource.CsPointEvalInvalid,
		DataToCircuit:      bp.blsPointEvalDataSource.Limb,
		Circuit:            newMultiPointEvalFailureCircuit(bp.Limits),
		NbCircuitInstances: maxNbCircuits,
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
