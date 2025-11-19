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
	moduleName       = "ECDATA"
	ROUND_NR         = 0
)

func colNameFn(colName string) ifaces.ColID {
	return ifaces.ColID(fmt.Sprintf("%s.%s", moduleName, colName))
}

type p256VerifyDataSource struct {
	ID       ifaces.Column
	CS       ifaces.Column
	Limb     ifaces.Column
	Index    ifaces.Column
	IsData   ifaces.Column
	IsResult ifaces.Column
}

func newP256VerifyDataSource(comp *wizard.CompiledIOP) *p256VerifyDataSource {
	return &p256VerifyDataSource{
		ID:       comp.Columns.GetHandle(colNameFn("ID")),
		CS:       comp.Columns.GetHandle(colNameFn("CIRCUIT_SELECTOR_P256_VERIFY")),
		Limb:     comp.Columns.GetHandle(colNameFn("LIMB")),
		Index:    comp.Columns.GetHandle(colNameFn("INDEX")),
		IsData:   comp.Columns.GetHandle(colNameFn("DATA_P256_VERIFY_FLAG")),
		IsResult: comp.Columns.GetHandle(colNameFn("RSLT_P256_VERIFY_FLAG")),
	}
}

type P256Verify struct {
	*p256VerifyDataSource
	p256VerifyGnarkData *plonk.Alignment
	*Limits
}

func newP256Verify(_ *wizard.CompiledIOP, limits *Limits, src *p256VerifyDataSource) *P256Verify {
	res := &P256Verify{
		p256VerifyDataSource: src,
		Limits:               limits,
	}

	return res
}

func (pv *P256Verify) WithCircuit(comp *wizard.CompiledIOP, options ...query.PlonkOption) *P256Verify {
	nbRowsPerCircuit := nbRows * pv.Limits.NbInputInstances
	maxNbCircuits := (pv.p256VerifyDataSource.CS.Size() + nbRowsPerCircuit - 1) / nbRowsPerCircuit

	toAlign := &plonk.CircuitAlignmentInput{
		Name:               fmt.Sprintf("%s_ALIGNMENT", NAME_P256_VERIFY),
		Round:              ROUND_NR,
		DataToCircuitMask:  pv.p256VerifyDataSource.CS,
		DataToCircuit:      pv.p256VerifyDataSource.Limb,
		Circuit:            newP256VerifyCircuit(pv.Limits),
		NbCircuitInstances: maxNbCircuits,
		PlonkOptions:       options,
		InputFillerKey:     input_filler_key,
	}
	pv.p256VerifyGnarkData = plonk.DefineAlignment(comp, toAlign)
	return pv
}

func (pv *P256Verify) Assign(run *wizard.ProverRuntime) {
	if pv.p256VerifyGnarkData != nil {
		pv.p256VerifyGnarkData.Assign(run)
	}
}

func NewP256VerifyZkEvm(comp *wizard.CompiledIOP, limits *Limits) *P256Verify {
	return newP256Verify(comp, limits, newP256VerifyDataSource(comp))
}
