package p256verify

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/plonk"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

const (
	NAME_P256_VERIFY = "P256_VERIFY"
	moduleName       = "ecdata"
	ROUND_NR         = 0
)

func colNameFn(colName string) ifaces.ColID {
	return ifaces.ColID(fmt.Sprintf("%s.%s", moduleName, colName))
}

type P256VerifyDataSource struct {
	ID       ifaces.Column
	CS       ifaces.Column
	Limb     ifaces.Column
	Index    ifaces.Column
	IsData   ifaces.Column
	IsResult ifaces.Column
}

func newP256VerifyDataSource(comp *wizard.CompiledIOP) *P256VerifyDataSource {
	return &P256VerifyDataSource{
		ID:       comp.Columns.GetHandle(colNameFn("ID")),
		CS:       comp.Columns.GetHandle(colNameFn("CIRCUIT_SELECTOR_P256_VERIFY")),
		Limb:     comp.Columns.GetHandle(colNameFn("LIMB")),
		Index:    comp.Columns.GetHandle(colNameFn("INDEX")),
		IsData:   comp.Columns.GetHandle(colNameFn("IS_P256_VERIFY_DATA")),
		IsResult: comp.Columns.GetHandle(colNameFn("IS_P256_VERIFY_RESULT")),
	}
}

type P256Verify struct {
	*P256VerifyDataSource
	P256VerifyGnarkData *plonk.Alignment
	*Limits
}

func newP256Verify(_ *wizard.CompiledIOP, limits *Limits, src *P256VerifyDataSource) *P256Verify {
	res := &P256Verify{
		P256VerifyDataSource: src,
		Limits:               limits,
	}

	return res
}

func (pv *P256Verify) WithCircuit(comp *wizard.CompiledIOP, options ...query.PlonkOption) *P256Verify {
	nbRowsPerCircuit := nbRows * pv.Limits.NbInputInstances
	maxNbCircuits := (pv.P256VerifyDataSource.CS.Size() + nbRowsPerCircuit - 1) / nbRowsPerCircuit

	toAlign := &plonk.CircuitAlignmentInput{
		Name:               fmt.Sprintf("%s_ALIGNMENT", NAME_P256_VERIFY),
		Round:              ROUND_NR,
		DataToCircuitMask:  pv.P256VerifyDataSource.CS,
		DataToCircuit:      pv.P256VerifyDataSource.Limb,
		Circuit:            newP256VerifyCircuit(pv.Limits),
		NbCircuitInstances: maxNbCircuits,
		PlonkOptions:       options,
		InputFillerKey:     input_filler_key,
	}
	pv.P256VerifyGnarkData = plonk.DefineAlignment(comp, toAlign)
	return pv
}

func (pv *P256Verify) Assign(run *wizard.ProverRuntime) {
	if pv.P256VerifyGnarkData != nil {
		pv.P256VerifyGnarkData.Assign(run)
	}
}

func NewP256VerifyZkEvm(comp *wizard.CompiledIOP, limits *Limits) *P256Verify {
	return newP256Verify(comp, limits, newP256VerifyDataSource(comp)).
		WithCircuit(comp, query.PlonkRangeCheckOption(16, 6, true))
}
