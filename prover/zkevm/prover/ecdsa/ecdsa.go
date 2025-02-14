package ecdsa

import (
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/plonkinwizard/plonkinternal"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
)

type EcdsaZkEvm struct {
	ant *antichamber
}

func NewEcdsaZkEvm(
	comp *wizard.CompiledIOP,
	settings *Settings,
) *EcdsaZkEvm {
	return &EcdsaZkEvm{
		ant: newAntichamber(
			comp,
			&antichamberInput{
				settings:     settings,
				ecSource:     getEcdataArithmetization(comp),
				txSource:     getTxnDataArithmetization(comp),
				rlpTxn:       getRlpTxnArithmetization(comp),
				plonkOptions: []any{plonkinternal.WithRangecheck(16, 6, true)},
			},
		),
	}
}

func (e *EcdsaZkEvm) Assign(run *wizard.ProverRuntime, txSig TxSignatureGetter, nbTx int) {
	e.ant.assign(run, txSig, nbTx)
}

func (e *EcdsaZkEvm) GetProviders() []generic.GenericByteModule {
	return e.ant.Providers
}

func getEcdataArithmetization(comp *wizard.CompiledIOP) *ecDataSource {
	return &ecDataSource{
		CsEcrecover: comp.Columns.GetHandle("ecdata.CIRCUIT_SELECTOR_ECRECOVER"),
		ID:          comp.Columns.GetHandle("ecdata.ID"),
		Limb:        comp.Columns.GetHandle("ecdata.LIMB"),
		SuccessBit:  comp.Columns.GetHandle("ecdata.SUCCESS_BIT"),
		Index:       comp.Columns.GetHandle("ecdata.INDEX"),
		IsData:      comp.Columns.GetHandle("ecdata.IS_ECRECOVER_DATA"),
		IsRes:       comp.Columns.GetHandle("ecdata.IS_ECRECOVER_RESULT"),
	}
}

func getTxnDataArithmetization(comp *wizard.CompiledIOP) *txnData {
	td := &txnData{
		fromHi: comp.Columns.GetHandle("txndata.FROM_HI"),
		fromLo: comp.Columns.GetHandle("txndata.FROM_LO"),
		ct:     comp.Columns.GetHandle("txndata.CT"),
	}
	return td
}

func getRlpTxnArithmetization(comp *wizard.CompiledIOP) generic.GenDataModule {
	return generic.GenDataModule{
		HashNum: comp.Columns.GetHandle("rlptxn.ABS_TX_NUM"),
		Index:   comp.Columns.GetHandle("rlptxn.INDEX_LX"),
		Limb:    comp.Columns.GetHandle("rlptxn.LIMB"),
		NBytes:  comp.Columns.GetHandle("rlptxn.nBYTES"),
		ToHash:  comp.Columns.GetHandle("rlptxn.TO_HASH_BY_PROVER"),
	}
}
