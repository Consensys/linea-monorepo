package ecdsa

import (
	"fmt"
	"github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	commonconstraints "github.com/consensys/linea-monorepo/prover/zkevm/prover/common/common_constraints"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
)

// TxSignature is responsible for assigning the relevant columns for transaction-Hash,
// and checking their consistency with the data coming from rlp_txn.
//
// columns for transaction-hash are native columns,
//
// columns for rlp-txn lives on the arithmetization side.
type TxSignature struct {
	Inputs   *txSignatureInputs
	TxHash   [common.NbLimbU256]ifaces.Column
	IsTxHash ifaces.Column

	// Provider for keccak, Provider contains the inputs and outputs of keccak hash.
	Provider generic.GenericByteModule
}

type txSignatureInputs struct {
	RlpTxn generic.GenDataModule
	Ac     *antichamber
}

func newTxSignatures(comp *wizard.CompiledIOP, inp txSignatureInputs) *TxSignature {
	var createCol = createColFn(comp, NAME_TXSIGNATURE, inp.Ac.Size)

	var res = &TxSignature{
		IsTxHash: createCol("TX_IS_HASH_HI"),
		Inputs:   &inp,
	}

	for i := 0; i < common.NbLimbU256; i++ {
		res.TxHash[i] = createCol(fmt.Sprintf("TX_HASH_%d", i))
	}

	commonconstraints.MustBeBinary(comp, res.IsTxHash)

	// isTxHash = 1 if isFeching = 1 and Source = 1
	comp.InsertGlobal(0, ifaces.QueryIDf("IS_TX_HASH"),
		sym.Mul(inp.Ac.IsFetching, inp.Ac.Source,
			sym.Sub(1, res.IsTxHash),
		),
	)

	// txHashHi remains the same between two fetchings.
	for i := 0; i < common.NbLimbU256; i++ {
		comp.InsertGlobal(0, ifaces.QueryIDf("txHash_REMAIN_SAME_%d", i),
			sym.Mul(inp.Ac.IsActive,
				sym.Sub(1, inp.Ac.IsFetching),
				sym.Sub(res.TxHash[i], column.Shift(res.TxHash[i], -1))),
		)
	}

	res.Provider = res.GetProvider(comp, inp.RlpTxn)

	return res
}

// It builds a provider from rlp-txn (as hash input) and native columns of TxSignature (as hash output)
// the consistency check is then deferred to the keccak module.
func (txn *TxSignature) GetProvider(comp *wizard.CompiledIOP, rlpTxn generic.GenDataModule) generic.GenericByteModule {
	provider := generic.GenericByteModule{}

	// pass rlp-txn as DataModule.
	provider.Data = rlpTxn

	// generate infoModule from native columns
	provider.Info = txn.buildInfoModule()

	return provider
}

// it builds an infoModule from native columns
func (txn *TxSignature) buildInfoModule() generic.GenInfoModule {
	info := generic.GenInfoModule{
		HashHi:   txn.TxHash[:common.NbLimbU128],
		HashLo:   txn.TxHash[common.NbLimbU128:],
		IsHashHi: txn.IsTxHash,
		IsHashLo: txn.IsTxHash,
	}
	return info
}

// it assign the native columns
func (txn *TxSignature) assignTxSignature(run *wizard.ProverRuntime, nbActualEcRecover int) {

	var (
		nbEcRecover = nbActualEcRecover
		n           = startAt(nbEcRecover)
		isTxHash    = vector.Repeat(field.Zero(), n)
		size        = txn.Inputs.Ac.Size
		permTrace   = keccak.GenerateTrace(txn.Inputs.RlpTxn.ScanStreams(run))

		hashColumns [common.NbLimbU256][]field.Element
	)

	for i := 0; i < common.NbLimbU256; i++ {
		hashColumns[i] = vector.Repeat(field.Zero(), n)
	}

	for _, digest := range permTrace.HashOutPut {
		hashLimbs := common.SplitBytes(digest[:])

		// Initialize limb values for each column of txHash
		for j, limb := range hashLimbs {
			var element field.Element
			element.SetBytes(limb[:])

			repeat := vector.Repeat(element, nbRowsPerTxSign)
			hashColumns[j] = append(hashColumns[j], repeat...)
		}

		repeatIsTxHash := vector.Repeat(field.Zero(), nbRowsPerTxSign-1)

		isTxHash = append(isTxHash, field.One())
		isTxHash = append(isTxHash, repeatIsTxHash...)
	}

	for i := 0; i < common.NbLimbU256; i++ {
		run.AssignColumn(txn.TxHash[i].GetColID(), smartvectors.RightZeroPadded(hashColumns[i], size))
	}

	run.AssignColumn(txn.IsTxHash.GetColID(), smartvectors.RightZeroPadded(isTxHash, size))
}

func startAt(nbEcRecover int) int {
	return nbEcRecover * nbRowsPerEcRec
}
