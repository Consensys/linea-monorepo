package arith_struct

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/csvtraces"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

// BlockDataCols models the arithmetization's BlockData module
type BlockDataCols struct {
	// RelBlock is the relative block number, ranging from 1 to the total number of blocks
	RelBlock ifaces.Column
	// Inst encodes the type of the row
	Inst ifaces.Column
	// Ct is a counter column
	Ct ifaces.Column
	// DataHi/DataLo encode the data, for example the timestamps.
	// It's divided into 16 16-bit limb columns. 256 bits in total.
	Data [common.NbLimbU256]ifaces.Column
	// FirstBlock contains the absolute ID of the first block
	// It's divided into 3 16-bit limb columns. 48 bits in total.
	FirstBlock [common.NbLimbU48]ifaces.Column
}

// TxnData models the arithmetization's TxnData module
type TxnData struct {
	// Absolute number of the transaction (starts from 1 and acts as an Active Filter), and the maximum number of
	// transactions
	AbsTxNum, AbsTxNumMax ifaces.Column
	// Relative TxNum inside the block,
	RelTxNum, RelTxNumMax ifaces.Column
	// Sender address. It's divided into 10 16-bit limb columns. 160 bits in total.
	From [common.NbLimbEthAddress]ifaces.Column
	// 1 if this is the last transaction inside the block
	IsLastTxOfBlock ifaces.Column
	// Relative Block number inside the batch
	RelBlock ifaces.Column
	Ct       ifaces.Column
	USER     ifaces.Column // 1 if this is a user transaction, 0 otherwise
	Selector ifaces.Column // we require an additional selector to identify which data to fetch
	SYSI     ifaces.Column
	SYSF     ifaces.Column
}

// RlpTxn models the arithmetization's RlpTxn module
type RlpTxn struct {
	// Absolute number of the transaction (starts from 1 and acts as an Active Filter), and the maximum number of transactions
	AbsTxNum, AbsTxNumMax ifaces.Column
	// Relative TxNum inside the block,
	ToHashByProver ifaces.Column
	// Limbs are columns that is used to store the RLP data.
	// It represents a single 128-bit limb, which is divided into 8 16-bit columns.
	Limbs [common.NbLimbU128]ifaces.Column
	// the number of bytes to load from the limb
	NBytes         ifaces.Column
	TxnPerspective ifaces.Column // indicator column for the transaction perspective, which we will use to obtain the ChainID
	ChainID        ifaces.Column // dedicated column for the ChainID
}

// DefineTestingArithModules defines the BlockDataCols, TxnData and RlpTxn modules based on csv traces.
// if a csvTrace is nil, the corresponding module will be missing in the return data.
func DefineTestingArithModules(b *wizard.Builder, ctBlockData, ctTxnData, ctRlpTxn *csvtraces.CsvTrace) (*BlockDataCols, *TxnData, *RlpTxn) {
	var (
		blockDataCols *BlockDataCols
		txnDataCols   *TxnData
		rlpTxn        *RlpTxn
	)

	if ctBlockData != nil {
		blockDataCols = &BlockDataCols{
			RelBlock: ctBlockData.GetCommit(b, "REL_BLOCK"),
			Inst:     ctBlockData.GetCommit(b, "INST"),
			Ct:       ctBlockData.GetCommit(b, "CT"),
		}

		for i := range blockDataCols.FirstBlock {
			blockDataCols.FirstBlock[i] = ctBlockData.GetCommit(b, fmt.Sprintf("FIRST_BLOCK_NUMBER_%d", i))
		}

		for i := range blockDataCols.Data {
			blockDataCols.Data[i] = ctBlockData.GetCommit(b, fmt.Sprintf("DATA_%d", i))
		}
	}
	if ctTxnData != nil {
		txnDataCols = &TxnData{
			AbsTxNum:        ctTxnData.GetCommit(b, "TD.ABS_TX_NUM"),
			AbsTxNumMax:     ctTxnData.GetCommit(b, "TD.ABS_TX_NUM_MAX"),
			RelTxNum:        ctTxnData.GetCommit(b, "TD.REL_TX_NUM"),
			RelTxNumMax:     ctTxnData.GetCommit(b, "TD.REL_TX_NUM_MAX"),
			Ct:              ctTxnData.GetCommit(b, "TD.CT"),
			IsLastTxOfBlock: ctTxnData.GetCommit(b, "TD.IS_LAST_TX_OF_BLOCK"),
			RelBlock:        ctTxnData.GetCommit(b, "TD.REL_BLOCK"),
			USER:            ctTxnData.GetCommit(b, "TD.USER"),
			Selector:        ctTxnData.GetCommit(b, "TD.SELECTOR"),
			SYSI:            ctTxnData.GetCommit(b, "TD.SYSI"),
			SYSF:            ctTxnData.GetCommit(b, "TD.SYSF"),
		}

		for i := range txnDataCols.From {
			txnDataCols.From[i] = ctTxnData.GetCommit(b, fmt.Sprintf("TD.FROM_%d", i))
		}
	}
	if ctRlpTxn != nil {
		rlpTxn = &RlpTxn{
			AbsTxNum:       ctRlpTxn.GetCommit(b, "RT.ABS_TX_NUM"),
			AbsTxNumMax:    ctRlpTxn.GetCommit(b, "RT.ABS_TX_NUM_MAX"),
			ToHashByProver: ctRlpTxn.GetCommit(b, "RL.TO_HASH_BY_PROVER"),
			NBytes:         ctRlpTxn.GetCommit(b, "RL.NBYTES"),
			TxnPerspective: ctRlpTxn.GetCommit(b, "RL.TXN"),
			ChainID:        ctRlpTxn.GetCommit(b, "RL.CHAIN_ID"),
		}

		for i := range rlpTxn.Limbs {
			rlpTxn.Limbs[i] = ctRlpTxn.GetCommit(b, fmt.Sprintf("RL.LIMB_%d", i))
		}
	}

	return blockDataCols, txnDataCols, rlpTxn
}

// AssignTestingArithModules assigns the BlockDataCols, TxnData and RlpTxn modules based on csv traces.
// if a module is missing,the corresponding assignment is skipped
func AssignTestingArithModules(run *wizard.ProverRuntime, ctBlockData, ctTxnData, ctRlpTxn *csvtraces.CsvTrace) {
	// assign the CSV data for the mock BlockData, TxnData and RlpTxn arithmetization modules
	if ctBlockData != nil {
		toAssign := []string{"REL_BLOCK", "INST", "CT"}

		for i := range common.NbLimbU256 {
			toAssign = append(toAssign, fmt.Sprintf("DATA_%d", i))
		}

		for i := range common.NbLimbU48 {
			toAssign = append(toAssign, fmt.Sprintf("FIRST_BLOCK_NUMBER_%d", i))
		}

		ctBlockData.Assign(run, toAssign...)
	}
	if ctTxnData != nil {
		toAssign := []string{
			"TD.ABS_TX_NUM",
			"TD.ABS_TX_NUM_MAX",
			"TD.REL_TX_NUM",
			"TD.REL_TX_NUM_MAX",
			"TD.CT",
			"TD.IS_LAST_TX_OF_BLOCK",
			"TD.REL_BLOCK",
			"TD.USER",
			"TD.SELECTOR",
			"TD.SYSI",
			"TD.SYSF",
		}

		for i := range common.NbLimbEthAddress {
			toAssign = append(toAssign, fmt.Sprintf("TD.FROM_%d", i))
		}

		ctTxnData.Assign(run, toAssign...)
	}
	if ctRlpTxn != nil {
		toAssign := []string{
			"RT.ABS_TX_NUM",
			"RT.ABS_TX_NUM_MAX",
			"RL.TO_HASH_BY_PROVER",
			"RL.NBYTES",
			"RL.TXN",
			"RL.CHAIN_ID",
		}

		for i := range common.NbLimbU128 {
			toAssign = append(toAssign, fmt.Sprintf("RL.LIMB_%d", i))
		}

		ctRlpTxn.Assign(run, toAssign...)
	}

}
