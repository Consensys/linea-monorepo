package ecdsa

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
)

type EcdsaZkEvm struct {
	Ant *antichamber
}

func NewEcdsaZkEvm(
	comp *wizard.CompiledIOP,
	settings *Settings,
) *EcdsaZkEvm {
	return &EcdsaZkEvm{
		Ant: newAntichamber(
			comp,
			&antichamberInput{
				Settings:     settings,
				EcSource:     getEcdataArithmetization(comp),
				TxSource:     getTxnDataArithmetization(comp),
				RlpTxn:       getRlpTxnArithmetization(comp),
				PlonkOptions: []query.PlonkOption{query.PlonkRangeCheckOption(16, 6, true)},
			},
		),
	}
}

func (e *EcdsaZkEvm) Assign(run *wizard.ProverRuntime, txSig TxSignatureGetter, nbTx int) {
	e.Ant.assign(run, txSig, nbTx)
}

func (e *EcdsaZkEvm) GetProviders() []generic.GenericByteModule {
	return e.Ant.Providers
}

func getEcdataArithmetization(comp *wizard.CompiledIOP) *ecDataSource {
	src := &ecDataSource{
		CsEcrecover: comp.Columns.GetHandle("ecdata.CIRCUIT_SELECTOR_ECRECOVER"),
		ID:          comp.Columns.GetHandle("ecdata.ID"),
		SuccessBit:  comp.Columns.GetHandle("ecdata.SUCCESS_BIT"),
		Index:       comp.Columns.GetHandle("ecdata.INDEX"),
		IsData:      comp.Columns.GetHandle("ecdata.IS_ECRECOVER_DATA"),
		IsRes:       comp.Columns.GetHandle("ecdata.IS_ECRECOVER_RESULT"),
	}

	for i := 0; i < common.NbLimbU128; i++ {
		src.Limb[i] = comp.Columns.GetHandle(ifaces.ColIDf("ecdata.LIMB_%d", i))
	}

	return src
}

func getTxnDataArithmetization(comp *wizard.CompiledIOP) *txnData {
	td := &txnData{
		Ct:       comp.Columns.GetHandle("txndata.CT"),
		User:     comp.Columns.GetHandle("txndata.USER"),
		Selector: comp.Columns.GetHandle("txndata.HUB"),
	}

	for i := 0; i < common.NbLimbU256; i++ {
		utils.Panic("")
		td.From[i] = comp.Columns.GetHandle(ifaces.ColIDf("txndata.FROM_%d", i))
	}

	return td
}

func getRlpTxnArithmetization(comp *wizard.CompiledIOP) generic.GenDataModule {
	res := generic.GenDataModule{
		HashNum: comp.Columns.GetHandle("rlptxn.USER_TXN_NUMBER"),
		Index:   comp.Columns.GetHandle("rlptxn.INDEX_LX"),
		NBytes:  comp.Columns.GetHandle("rlptxn.cmpLIMB_SIZE"),
		ToHash:  comp.Columns.GetHandle("rlptxn.TO_HASH_BY_PROVER"),
	}

	for i := 0; i < common.NbLimbU128; i++ {
		// We need to check how corset will name that column because
		// pre-small-field, the column is mixed and may contain chain-ID or
		// other stuffs and that makes the name hard to predict.
		utils.Panic("Clarify the name of the column once for rlptxn.cmpLIMB")
		res.Limbs = append(res.Limbs, comp.Columns.GetHandle(ifaces.ColIDf("rlptxn.cmpLIMB_%d", i)))
	}

	return res
}
