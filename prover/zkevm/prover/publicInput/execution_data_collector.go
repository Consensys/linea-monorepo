package publicInput

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/projection"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	fetch "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/fetchers_arithmetization"
	util "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/utilities"
)

/*
The ExecutionDataCollector module has as input data that was fetched from various modules of
the arithmetization, namely the BLOCKDATA, TXNDATA and RLPTXN modules. It constructs a series of limbs
that represent the execution data of the public input.

This module, along with its component arithmetization fetchers, is still a work in progress.

For each block with transactions tx_1...tx_n, we want to hash, in the following order:
The number of transactions in the block (2 bytes), the block timestamp (4 bytes), the blockhash (32 bytes),
and then for each transaction tx_i, the sender address (20 bytes) and the transaction RLP.
We then continue analogously for each block.

Due to design choices in the arithmetization and other submodules, we can only load at most 16 bytes
at a time. For this reason, blockhash is divided into two columns: BlockHashHi (16 bytes) and BlockHashLo (16 bytes).
Similarly, the sender address is divided into AddrHi (4 bytes) and AddrLo (16 bytes).


Finally, the RLP data for each transaction is stored in the RLPTXN module. We use an intermediary fetcher
which gives us the following columns:
AbsTxId, RLPLimb, NBytes,
where AbsTxId contains increasing, contiguous segments of the absolute transaction ids, followed by the
limb data and the NBytes column which indicates the number of relevant bytes stored in each row of
the RLPLimb column. To give more details, the AbsTxId is the transaction id inside the entire conflated batch,
across the multiple blocks present in the batch.

All the data to be fetched is ordered in the Limb and UnalignedLimb columns (we need both because different sources
have different data formats). We address the alignment formatting after we describe the structure of the
module. In order to load data from different sources, we use a series of indicator columns:
IsNoTx, IsBlockHashHi, IsBlockHashLo, IsTimestamp, IsTxRLP, IsAddrHi, IsAddrLo, which signal which
type of data we are loading in each row.

An example of this first structural level can be found below:

| BlockID | AbsTxID | Limb | IsNoTx | IsTimestamp | IsBlockHashHi | IsBlockHashLo | IsAddrHi | IsAddrLo | IsTxRLP |
|---------|---------|------|--------|-------------|---------------|---------------|----------|----------|---------|
|    1    |    1    | 0xab |   1    |       0     |       0       |       0       |     0    |     0    |    0    |
|    1    |    1    | 0xa  |   0    |       1     |       0       |       0       |     0    |     0    |    0    |
|    1    |    1    | 0xcc |   0    |       0     |       1       |       0       |     0    |     0    |    0    |
|    1    |    1    | 0xa  |   0    |       0     |       0       |       1       |     0    |     0    |    0    |
|    1    |    1    | 0xdd |   0    |       0     |       0       |       0       |     1    |     0    |    0    |
|    1    |    1    | 0xa  |   0    |       0     |       0       |       0       |     0    |     1    |    0    |
|    1    |    1    | 0xdd |   0    |       0     |       0       |       0       |     0    |     0    |    1    |
|    1    |    1    | 0xe  |   0    |       0     |       0       |       0       |     0    |     0    |    1    |
|    1    |    1    | 0xcf |   0    |       0     |       0       |       0       |     0    |     0    |    1    |
|---------|---------|------|--------|-------------|---------------|---------------|----------|----------|---------|
Note that the IsTxRLP column loads multiple RLP limbs for transaction with AbsTxID = 1. Once these are loaded,
we continue similarly with AbsTxID = 2.
|---------|---------|------|--------|-------------|---------------|---------------|----------|----------|---------|
|    1    |    2    | 0xff |   1    |       0     |       0       |       0       |     0    |     0    |    0    |
|   ...   |   ...   |  ... |   ...  |      ...    |      ...      |      ...      |    ...   |    ...   |   ...   |
|---------|---------|------|--------|-------------|---------------|---------------|----------|----------|---------|
When all the data for BlockID = 1 and its associated transactions has been loaded, we move to BlockID = 2
|---------|---------|------|--------|-------------|---------------|---------------|----------|----------|---------|
|    2    |    1    | 0xa  |   1    |       0     |       0       |       0       |     0    |     0    |    0    |
|   ...   |   ...   |  ... |  ...   |      ...    |      ...      |      ...      |    ...   |    ...   |   ...   |
|---------|---------|------|--------|-------------|---------------|---------------|----------|----------|---------|
And we continue similarly until all the blocks have been processed.
Note that IsNoTx, IsBlockHashHi, IsBlockHashLo, IsTimestamp, IsTxRLP, IsAddrHi, IsAddrLo are mutually exclusive,
and cannot be 1 at the same time.

We distinguish different types of segments inside our module, in three different layers:
1. The BlockId layers groups block data and transaction data belonging to that block. These rows will be
contiguous and have the same BlockId.
2. Each block has different transactions, and the data of which transaction will correspond to a contiguous
segment that has the same AbsTxID.
3. Each transaction has a segment in which its RLP data is loaded.
We need to be able to distinguish when each segment ends, for which we use the following columns:

FirstAbsTxIDBlock and LastAbsTxIDBlock are two columns that store the first and last AbsTxID inside the
corresponding block. These are constrained in two ways:
- via a projection query from the fetcher that computes block metadata out of the TXNDATA module.
- via constancy constraints that require FirstAbsTxIDBlock/LastAbsTxIDBlock to remain constant whenever
the BlockId remains constant.
Using the LastAbsTxIDBlock, we can tell when the block data ends, by checking that AbsTxId = LastAbsTxIDBlock
via a selector column SelectorLastTxBlock, with:
(SelectorLastTxBlock[i] = 1) iff (AbsTxId[i] = LastAbsTxIDBlock[i])
This explains how to check the end of the block layer.

For the transaction level, we need to distinguish when the corresponding RLP segment ends. For this purpose,
we use a EndOfRlpSegment column, which is constrained in two ways:
- via a projection query from the fetcher that processed RLP data from the RLPTXN module.
The projection inspects (abAbsTxID, AbsTxIDMax, limb, EndOfRlp) on filter IsTxRLP = 1
- via a global constraint that ensures that EndOfRlpSegment = 0 whenever IsTxRLP = 1.

Note that BlockId can only increase when EndOfRlpSegment = 1, which ensures that it is
not possible for leftover rlp limbs belonging to another transaction to end up in the next block.

If the projection query passes, we know that all RLP limbs are loaded at IsTxRLP = 1
for rows with the proper transaction number AbsTxId. AbsTxIdMax also gets enforced through this projection query.

In order to distinguish the relevant part of the module from the padding, we use an IsActive column, which
is 1 on every row that loads data and 0 otherwise.


The counters:
BlockID is constrained via a projection query on filter edc.IsNoTx = 1
it is also constrained with a local constraint, forcing it to start at value 1
it increases by 1 when isActive = 1, SelectorLastTxBlock = 1, EndOfRlpSegment = 1
and SelectorEndOfAllTx = 0

| Limb | NoBytes | UnalignedLimb | AlignedPow |

A different projection query ensures that (RelBlock, FirstAbsTxIDBlock, LastAbsTxIDBlock, TotalNoTxBlock) is
correctly fetched when IsAddrHi = 1 and separately, also when IsAddrLo = 1

The timestamp projection query on (RelBlock, Limb) ensures that when IsTimestamp = 1,
the limb contains the timestamp properly associated with that relative block number.

The AbsTxId counter starts from 1. When IsActive = 1, EndOfRlpSegment = 0 and SelectorEndOfAllTx = 0, the
AbsTxId counter is forced to remain constant. The reason to also ask that SelectorEndOfAllTx = 0, is because
the row after SelectorEndOfAllTx = 1 will have AbsTxId = 0 (enforce this)
*/

const (
	loadNoTxn        = 0
	loadTimestamp    = 1
	loadBlockHashHi  = 2
	loadBlockHashLo  = 3
	loadSenderAddrHi = 4
	loadSenderAddrLo = 5
	loadRlp          = 6

	noBytesNoTxn        = 2
	noBytesTimestamp    = 6
	noBytesBlockHash    = 16
	noBytesSenderAddrHi = 4
	noBytesSenderAddrLo = 16

	powBytesNoTxn   = "5192296858534827628530496329220096" // 2 bytes when loading NO_TX, 2^(128-2*8)
	powTimestamp    = "1208925819614629174706176"          // 6 bytes when loading TIMESTAMP, 2^(128-6*8)
	powBlockHash    = "1"                                  // 16 bytes when loading BlockHash, 2^(128-16*8)
	powSenderAddrHi = "79228162514264337593543950336"      // 4 bytes when loading SENDER ADDR HI, 2^(128-4*8)
	powSenderAddrLo = "1"                                  // 16 bytes bytes when loading SenderAddrLo, 2^(128-16*8)
)

type ExecutionDataCollector struct {
	BlockID, AbsTxID, AbsTxIDMax ifaces.Column
	Limb, NoBytes                ifaces.Column
	UnalignedLimb, AlignedPow    ifaces.Column
	TotalNoTxBlock               ifaces.Column

	IsActive                                                                       ifaces.Column
	IsNoTx, IsTimestamp, IsBlockHashHi, IsBlockHashLo, IsAddrHi, IsAddrLo, IsTxRLP ifaces.Column

	SelectorBlockDiff        ifaces.Column
	ComputeSelectorBlockDiff wizard.ProverAction

	SelectorLastTxBlock        ifaces.Column
	ComputeSelectorLastTxBlock wizard.ProverAction

	FirstAbsTxIDBlock, LastAbsTxIDBlock ifaces.Column

	SelectorEndOfAllTx        ifaces.Column
	ComputeSelectorEndOfAllTx wizard.ProverAction

	EndOfRlpSegment ifaces.Column

	SelectorAbsTxIDDiff        ifaces.Column
	ComputeSelectorAbsTxIDDiff wizard.ProverAction
}

func NewLimbSummary(comp *wizard.CompiledIOP, name string, size int) ExecutionDataCollector {
	res := ExecutionDataCollector{
		BlockID:           util.CreateCol(name, "BLOCK_ID", size, comp),
		AbsTxID:           util.CreateCol(name, "ABS_TX_ID", size, comp),
		AbsTxIDMax:        util.CreateCol(name, "ABS_TX_ID_MAX", size, comp),
		FirstAbsTxIDBlock: util.CreateCol(name, "FIRST_ABS_TX_ID_BLOCK", size, comp),
		LastAbsTxIDBlock:  util.CreateCol(name, "LAST_ABS_TX_ID_BLOCK", size, comp),

		Limb:            util.CreateCol(name, "LIMB", size, comp),
		NoBytes:         util.CreateCol(name, "NO_BYTES", size, comp),
		UnalignedLimb:   util.CreateCol(name, "UNALIGNED_LIMB", size, comp),
		AlignedPow:      util.CreateCol(name, "ALIGNED_POW", size, comp),
		TotalNoTxBlock:  util.CreateCol(name, "TOTAL_NO_TX_BLOCK", size, comp),
		IsActive:        util.CreateCol(name, "IS_ACTIVE", size, comp),
		IsNoTx:          util.CreateCol(name, "IS_NO_TX", size, comp),
		IsBlockHashHi:   util.CreateCol(name, "IS_BLOCK_HASH_HI", size, comp),
		IsBlockHashLo:   util.CreateCol(name, "IS_BLOCK_HASH_LO", size, comp),
		IsTimestamp:     util.CreateCol(name, "IS_TIMESTAMP", size, comp),
		IsTxRLP:         util.CreateCol(name, "IS_TX_RLP", size, comp),
		IsAddrHi:        util.CreateCol(name, "IS_ADDR_HI", size, comp),
		IsAddrLo:        util.CreateCol(name, "IS_ADDR_LO", size, comp),
		EndOfRlpSegment: util.CreateCol(name, "END_OF_RLP_SEGMENT", size, comp),
	}
	return res
}

func GetSummarySize(btm *fetch.BlockTxnMetadata, bdc *fetch.BlockDataCols, td *fetch.TxnData, rt *fetch.RlpTxn) int {
	// number of transactions, block timestamp, blockhash + for every transaction, sender address + transaction RLP limbs
	size := td.Ct.Size()
	if size < rt.Limb.Size() {
		size = rt.Limb.Size()
	}
	return size
}

func DefineBlockIdCounterConstraints(comp *wizard.CompiledIOP, edc *ExecutionDataCollector, name string) {
	// Counter constraints
	// First, the counter starts from 0
	comp.InsertLocal(0, ifaces.QueryIDf("%s_%v_COUNTER_START_LOCAL_CONSTRAINT", name, edc.BlockID.GetColID()),
		sym.Mul(
			edc.IsActive,
			sym.Sub(
				edc.BlockID,
				1, // blockIDs start from 1
			),
		),
	)

	comp.InsertGlobal(0, ifaces.QueryIDf("%s_%v_COUNTER_STAYS_THE_SAME_OR_INCREASES_BY_1_GLOBAL_CONSTRAINT", name, edc.BlockID.GetColID()),
		sym.Mul(
			column.Shift(edc.IsActive, 1),
			sym.Mul(
				sym.Sub(
					column.Shift(edc.BlockID, 1),
					edc.BlockID,
					1), // increases by 1
				sym.Sub(
					column.Shift(edc.BlockID, 1),
					edc.BlockID, // or stays the same
				),
			),
		),
	)

	// this constraint might not be needed anymore
	comp.InsertGlobal(0, ifaces.QueryIDf("%s_%v_COUNTER_INCREASES_BY_1_GLOBAL_CONSTRAINT", name, edc.BlockID.GetColID()),
		sym.Mul(
			edc.IsActive,
			edc.SelectorLastTxBlock, // the last transaction inside the block
			edc.EndOfRlpSegment,
			sym.Sub(
				1,
				edc.SelectorEndOfAllTx,
			), // not at the end of all blocks
			sym.Sub(
				column.Shift(edc.BlockID, 1),
				edc.BlockID,
				1,
			),
		),
	)
}

func DefineAbsTxIdCounterConstraints(comp *wizard.CompiledIOP, edc *ExecutionDataCollector, name string) {
	// Counter constraints
	// First, the counter starts from 1
	comp.InsertLocal(0, ifaces.QueryIDf("%s_%v_COUNTER_START_LOCAL_CONSTRAINT", name, edc.AbsTxID.GetColID()),
		sym.Mul(
			edc.IsActive,
			sym.Sub(
				edc.AbsTxID,
				1, // absTxId starts from 1
			),
		),
	)
	// edc.AbsTxID can only stay the same or increase exactly by 1
	comp.InsertGlobal(0, ifaces.QueryIDf("%s_%v_COUNTER_STAYS_THE_SAME_OR_INCREASES_BY_1_GLOBAL_CONSTRAINT", name, edc.AbsTxID.GetColID()),
		sym.Mul(
			column.Shift(edc.IsActive, 1),
			sym.Mul(
				sym.Sub(
					column.Shift(edc.AbsTxID, 1),
					edc.AbsTxID,
					1), // increases by 1
				sym.Sub(
					column.Shift(edc.AbsTxID, 1),
					edc.AbsTxID, // or stays the same
				),
			),
		),
	)
	// fine-grained control over when edc.AbsTxID increases
	comp.InsertGlobal(0, ifaces.QueryIDf("%s_%v_COUNTER_INCREASES_GLOBAL_CONSTRAINT", name, edc.AbsTxID.GetColID()),
		sym.Mul(
			edc.IsActive,
			edc.EndOfRlpSegment, // at the end of the RLP segment, we move to the next transaction
			sym.Sub(
				1,
				edc.SelectorEndOfAllTx,
			), // not at the end of all blocks
			sym.Sub(
				column.Shift(edc.AbsTxID, 1),
				edc.AbsTxID,
				1,
			),
		),
	)

	// fine-grained constraint over when edc.AbsTxID remains constant inside a transaction segment
	comp.InsertGlobal(0, ifaces.QueryIDf("%s_%v_CONSTANT_INSIDE_TRANSACTION_SEGMENTS", name, edc.AbsTxID.GetColID()),
		sym.Mul(
			edc.IsActive,
			sym.Sub(
				1,
				edc.EndOfRlpSegment,
			), // while not at the end of the RLP segment, the transaction number remains constant
			sym.Sub(
				1,
				edc.SelectorEndOfAllTx,
			), // not at the end of all blocks
			sym.Sub(
				column.Shift(edc.AbsTxID, 1),
				edc.AbsTxID,
			),
		),
	)
}

func DefineIndicatorOrder(comp *wizard.CompiledIOP, edc *ExecutionDataCollector, name string) {
	// start with IsNoTx = 1
	comp.InsertLocal(0,
		ifaces.QueryIDf("%s_START_WITH_IS_NO_TX", name),
		sym.Sub(
			edc.IsNoTx,
			1,
		),
	)

	// IsNbTx -> isTimestamp
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_IS_NB_TX_TO_IS_TIMESTAMP", name),
		sym.Sub(
			column.Shift(edc.IsNoTx, -1),
			edc.IsTimestamp,
		),
	)

	// isTimestamp -> isBlockHashHi
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_IS_NB_TX_TO_IS_BLOCKHASH_HI", name),
		sym.Sub(
			column.Shift(edc.IsTimestamp, -1),
			edc.IsBlockHashHi,
		),
	)

	// isBlockHashHi->isBlockHashLo
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_BLOCKHASH_HI_TO_BLOCKHASH_LO", name),
		sym.Sub(
			column.Shift(edc.IsBlockHashHi, -1),
			edc.IsBlockHashLo,
		),
	)

	// isBlockHashLO->IsAddr
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_BLOCKHASH_LO_TO_IS_ADDR_HI", name),
		sym.Mul(
			column.Shift(edc.IsBlockHashLo, -1), // this constraint says if IsBlockHashLo[i-1] =1 then IsAddrHi[i] = 1
			sym.Sub(
				column.Shift(edc.IsBlockHashLo, -1),
				edc.IsAddrHi,
			),
		),
	)

	// IsAddrHi -> IsAddrLo
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_IS_ADDR_HI_TO_IS_ADDR_LO", name),
		sym.Sub(
			column.Shift(edc.IsAddrHi, -1),
			edc.IsAddrLo,
		),
	)

	// IsAddrLo -> IsTxRLP
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_IS_ADDR_LO_TO_IS_TX_RLP", name),
		sym.Mul(
			column.Shift(edc.IsAddrLo, -1),
			sym.Sub(
				column.Shift(edc.IsAddrLo, -1),
				edc.IsTxRLP,
			),
		),
	)

	// IsTxRLP -> IsTxRLP while inside the transaction RLP segment
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_IS_TX_RLP_REMAINS_1_INSIDE_RLP_SEGMENT", name),
		sym.Mul(
			sym.Sub(1,
				edc.EndOfRlpSegment,
			), // we are inside an RLP segment
			edc.IsTxRLP, // if IsTxRLP is 1 then on the next row IsTxRLP will be 1 if we are inside the block
			sym.Sub(
				1,
				column.Shift(edc.IsTxRLP, 1),
			),
		),
	)

	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_IS_TX_RLP_TO_NEXT_BLOCK", name),
		sym.Mul(
			edc.SelectorLastTxBlock, // it is the last transaction in the block
			edc.EndOfRlpSegment,     // it is the end of the RLP segment
			sym.Sub(
				1,
				edc.SelectorEndOfAllTx, // but it is not the end of all transactions
			),
			edc.IsTxRLP, // do not comment this row, due to the dependency between EndOfRlpSegment and IsTxRLP.
			// (if IsTxRLP is 1 then on the next row IsNoTx will be 1 since we move on to the next block)
			sym.Sub(
				edc.IsTxRLP,
				column.Shift(edc.IsNoTx, 1),
			),
		),
	)

	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_IS_TX_RLP_TO_IS_ADDR_HI_GLOBAL_CONSTRAINT", name),
		sym.Mul(
			sym.Sub(
				1,
				edc.SelectorLastTxBlock, // not the end of the transactions in the current block
			),
			edc.EndOfRlpSegment, // end of RLP segment for the current transaction
			edc.IsTxRLP,         // do not comment this row, due to the dependency between EndOfRlpSegment and IsTxRLP.
			// we are inside the RLP segment
			sym.Sub(
				1,
				column.Shift(edc.IsAddrHi, 1), // load the sender address of the next transaction, since we are still inside the block
			),
		),
	)

	// direct transition to inactive part when there are no overflow bytes
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_IS_TX_RLP_DIRECTLY_TO_IS_INACTIVE_GLOBAL_CONSTRAINT", name),
		sym.Mul(
			edc.SelectorEndOfAllTx,        // 1 at the very last transaction, in the last block
			edc.EndOfRlpSegment,           // 1 at the end of RLP segment for the current transaction
			edc.IsTxRLP,                   // 1 inside the RLP segment
			column.Shift(edc.IsActive, 1), // all the above forces isActive to be 0 on the next position
		),
	)

}

func DefineIndicatorExclusion(comp *wizard.CompiledIOP, edc *ExecutionDataCollector, name string) {
	// IsNoTx + IsBlockHashHi + IsBlockHashLo + IsTimestamp + IsTxRLP + IsAddr = 1
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_INDICATOR_EXCLUSION_GLOBAL_CONSTRAINT", name),
		sym.Sub(
			edc.IsActive,
			sym.Add(
				edc.IsNoTx,
				edc.IsBlockHashHi,
				edc.IsBlockHashLo,
				edc.IsTimestamp,
				edc.IsTxRLP,
				edc.IsAddrHi,
				edc.IsAddrLo,
			),
		),
	)

}

func DefineAlignmentPowers(comp *wizard.CompiledIOP, edc *ExecutionDataCollector, name string) {
	// Value of the alignment exponent, isNoTx
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_IS_NO_TX_ALIGNMENT_EXPONENT_VALUE", name),
		sym.Mul(
			edc.IsNoTx,
			sym.Sub(
				edc.AlignedPow,
				field.NewFromString(powBytesNoTxn), // 2 bytes when loading NO_TX, 2^(128-2*8)
			),
		),
	)

	// Value of the alignment exponent, IS_TIMESTAMP
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_IS_TIMESTAMP_ALIGNMENT_EXPONENT_VALUE", name),
		sym.Mul(
			edc.IsTimestamp,
			sym.Sub(
				edc.AlignedPow,
				field.NewFromString(powTimestamp), // 6 bytes when loading TIMESTAMP, 2^(128-6*8)
			),
		),
	)

	// Value of the alignment exponent, IS_BLOCKHASH_HI
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_IS_BLOCKHASH_HI_ALIGNMENT_EXPONENT_VALUE", name),
		sym.Mul(
			edc.IsBlockHashHi,
			sym.Sub(
				edc.AlignedPow,
				field.NewFromString(powBlockHash),
			),
		),
	)

	// Value of the alignment exponent, IS_BLOCKHASH_LO
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_IS_BLOCKHASH_LO_ALIGNMENT_EXPONENT_VALUE", name),
		sym.Mul(
			edc.IsBlockHashLo,
			sym.Sub(
				edc.AlignedPow,
				field.NewFromString(powBlockHash),
			),
		),
	)

	// Value of the alignment exponent, IS_SENDER_HI
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_IS_SENDER_HI_ALIGNMENT_EXPONENT_VALUE", name),
		sym.Mul(
			edc.IsAddrHi,
			sym.Sub(
				edc.AlignedPow,
				field.NewFromString(powSenderAddrHi),
			),
		),
	)

	// Value of the alignment exponent, IS_SENDER_LO
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_IS_SENDER_LO_ALIGNMENT_EXPONENT_VALUE", name),
		sym.Mul(
			edc.IsAddrLo,
			sym.Sub(
				edc.AlignedPow,
				field.NewFromString(powSenderAddrLo),
			),
		),
	)
}

func DefineNumberOfBytesConstraints(comp *wizard.CompiledIOP, edc *ExecutionDataCollector, name string) {
	// noOfBytes isNoTx
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_IS_NO_TX_NO_BYTES", name),
		sym.Mul(
			edc.IsNoTx,
			sym.Sub(
				edc.NoBytes,
				noBytesNoTxn, // 2 bytes when loading NO_TX
			),
		),
	)

	// noOfBytes isTimestamp
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_IS_TIMESTAMP", name),
		sym.Mul(
			edc.IsTimestamp,
			sym.Sub(
				edc.NoBytes,
				noBytesTimestamp, // 6 bytes when loading a TIMESTAMP
			),
		),
	)

	// noOfBytes isBlockhashHi
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_IS_BLOCKHASH_HI_NO_BYTES", name),
		sym.Mul(
			edc.IsBlockHashHi,
			sym.Sub(
				edc.NoBytes,
				noBytesBlockHash, // 6 bytes when loading a blockhashHi
			),
		),
	)
	// noOfBytes isBlockhashLo
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_IS_BLOCKHASH_LO_NO_BYTES", name),
		sym.Mul(
			edc.IsBlockHashLo,
			sym.Sub(
				edc.NoBytes,
				noBytesBlockHash, // 6 bytes when loading a blockhashLo
			),
		),
	)

	// noOfBytes isAddrHi
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_IS_ADDR_HI_NO_BYTES", name),
		sym.Mul(
			edc.IsAddrHi,
			sym.Sub(
				edc.NoBytes,
				noBytesSenderAddrHi, // 4 bytes when loading a isAddrHi
			),
		),
	)

	// noOfBytes isAddrLo
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_IS_ADDR_LO_NO_BYTES", name),
		sym.Mul(
			edc.IsAddrLo,
			sym.Sub(
				edc.NoBytes,
				noBytesSenderAddrLo, // 16 bytes when loading a isAddrLo
			),
		),
	)

	// noOfBytes limbs in the RLP transaction data
	// this must be checked with a projection query
}

func ProjectionQueries(comp *wizard.CompiledIOP,
	edc *ExecutionDataCollector,
	name string,
	timestamps fetch.TimestampFetcher,
	metadata fetch.BlockTxnMetadata,
	txnData fetch.TxnDataFetcher,
	rlp fetch.RlpTxnFetcher) {

	metadataTable := []ifaces.Column{
		metadata.BlockID,
		metadata.TotalNoTxnBlock,
		metadata.FirstAbsTxId,
		metadata.LastAbsTxId,
	}

	lsMetadataTable := []ifaces.Column{
		edc.BlockID,
		edc.TotalNoTxBlock,
		edc.FirstAbsTxIDBlock,
		edc.LastAbsTxIDBlock,
	}

	projection.InsertProjection(comp,
		ifaces.QueryIDf("%s_BLOCK_METADATA_PROJECTION", name),
		lsMetadataTable,
		metadataTable,
		edc.IsNoTx,
		metadata.FilterFetched,
	)

	// !!! need to add segment constancies, since I am only using isNoTx in BLOCK_METADATA_PROJECTION

	timestampTable := []ifaces.Column{
		timestamps.RelBlock,
		timestamps.Data,
	}

	lsTimestamps := []ifaces.Column{
		edc.BlockID,
		edc.UnalignedLimb,
	}

	projection.InsertProjection(comp,
		ifaces.QueryIDf("%s_TIMESTAMP_PROJECTION", name),
		lsTimestamps,
		timestampTable,
		edc.IsTimestamp,
		timestamps.FilterFetched,
	)

	txnDataTableHi := []ifaces.Column{
		txnData.RelBlock,
		txnData.AbsTxNum,
		txnData.FromHi,
	}
	lsTxnSenderAddressTableHi := []ifaces.Column{
		edc.BlockID,
		edc.AbsTxID,
		edc.UnalignedLimb,
	}

	projection.InsertProjection(comp,
		ifaces.QueryIDf("%s_SENDER_ADDRESS_HI_PROJECTION", name),
		lsTxnSenderAddressTableHi,
		txnDataTableHi,
		edc.IsAddrHi,
		txnData.FilterFetched,
	)

	txnDataTableLo := []ifaces.Column{
		txnData.RelBlock,
		txnData.AbsTxNum,
		txnData.FromLo,
	}
	lsTxnSenderAddressTableLo := []ifaces.Column{
		edc.BlockID,
		edc.AbsTxID,
		edc.UnalignedLimb,
	}

	projection.InsertProjection(comp,
		ifaces.QueryIDf("%s_SENDER_ADDRESS_LO_PROJECTION", name),
		lsTxnSenderAddressTableLo,
		txnDataTableLo,
		edc.IsAddrLo,
		txnData.FilterFetched,
	)

	rlpDataTable := []ifaces.Column{
		rlp.AbsTxNum,
		rlp.AbsTxNumMax,
		rlp.Limb,
		rlp.NBytes,
		rlp.EndOfRlpSegment,
	}
	lsRlpDataTable := []ifaces.Column{
		edc.AbsTxID,
		edc.AbsTxIDMax,
		edc.UnalignedLimb,
		edc.NoBytes,
		edc.EndOfRlpSegment, // This constrains EndOfRlpSegment on edc.IsTx.RLP = 1, but we still need to constrain it elsewhere
	}

	projection.InsertProjection(comp,
		ifaces.QueryIDf("%s_RLP_LIMB_DATA_PROJECTION", name),
		lsRlpDataTable,
		rlpDataTable,
		edc.IsTxRLP,
		rlp.FilterFetched,
	)
}

func EnforceZeroOnInactiveFilter(comp *wizard.CompiledIOP, filterExpr *sym.Expression, targetCol ifaces.Column, name, subname string) {
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%s_%s_%v_IS_ZERO_WHEN_INACTIVE_OR_NO_PADDING_GLOBAL_CONSTRAINT", name, subname, targetCol.GetColID()),
		sym.Sub(
			targetCol,
			sym.Mul(
				targetCol,
				filterExpr,
			),
		),
	)
}

func DefineSelectorConstraints(comp *wizard.CompiledIOP, edc *ExecutionDataCollector, name string) {
	edc.SelectorLastTxBlock, edc.ComputeSelectorLastTxBlock = dedicated.IsZero(
		comp,
		sym.Sub(
			edc.AbsTxID,
			edc.LastAbsTxIDBlock,
		),
	)

	edc.SelectorEndOfAllTx, edc.ComputeSelectorEndOfAllTx = dedicated.IsZero(
		comp,
		sym.Sub(
			edc.AbsTxID,
			edc.AbsTxIDMax,
		),
	)

	edc.SelectorBlockDiff, edc.ComputeSelectorBlockDiff = dedicated.IsZero(
		comp,
		sym.Sub(
			edc.BlockID,
			column.Shift(edc.BlockID, 1),
		),
	)

	edc.SelectorAbsTxIDDiff, edc.ComputeSelectorAbsTxIDDiff = dedicated.IsZero(
		comp,
		sym.Sub(
			edc.AbsTxID,
			column.Shift(edc.AbsTxID, 1),
		),
	)

	// edc.EndOfRlpSegment is partially constrained in the projection queries, on areas where edc.IsTxRLP = 1
	// here we require that when edc.IsTxRLP = 0, we have EndOfRlpSegment = 0
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%s_END_OF_RLP_SEGMENT_ZEROIZATION", name),
		sym.Mul(
			sym.Sub(
				1,
				edc.IsTxRLP,
			),
			ifaces.ColumnAsVariable(edc.EndOfRlpSegment),
		),
	)

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%s_END_OF_RLP_SEGMENT_IS_0_WHENEVER_TX_ID_IS_CONSTANT", name),
		sym.Mul(
			edc.SelectorAbsTxIDDiff,                      // Recall that (SelectorAbsTxIDDiff = 1) iff (AbsTxID[i]=AbsTxID[i+1])
			ifaces.ColumnAsVariable(edc.EndOfRlpSegment), // must be 0 in this case
		),
	)

}

func DefineCounterConstraints(comp *wizard.CompiledIOP, edc *ExecutionDataCollector, name string) {
	DefineBlockIdCounterConstraints(comp, edc, name)
	DefineAbsTxIdCounterConstraints(comp, edc, name)
}

func DefineZeroizationConstraints(comp *wizard.CompiledIOP, edc *ExecutionDataCollector, name string) {
	// enforce zero fields when isActive is not set to 1
	var emptyWhenInactive = [...]ifaces.Column{
		edc.BlockID,
		edc.AbsTxID,
		edc.AbsTxIDMax,
		edc.TotalNoTxBlock,
		edc.IsNoTx,
		edc.IsBlockHashHi,
		edc.IsBlockHashLo,
		edc.IsTimestamp,
		edc.IsTxRLP,
		edc.IsAddrHi,
		edc.IsAddrLo,
		edc.FirstAbsTxIDBlock,
		edc.LastAbsTxIDBlock,
		edc.EndOfRlpSegment,
		edc.Limb,
		edc.NoBytes,
		edc.UnalignedLimb,
		edc.AlignedPow,
	}

	for _, col := range emptyWhenInactive {
		// if isActive = 0 then the column becomes 0
		EnforceZeroOnInactiveFilter(comp, ifaces.ColumnAsVariable(edc.IsActive), col, name, "IS_ACTIVE")
	}
}
func DefineIndicatorsMustBeBinary(comp *wizard.CompiledIOP, edc *ExecutionDataCollector) {
	util.MustBeBinary(comp, edc.IsActive)
	util.MustBeBinary(comp, edc.IsNoTx)
	util.MustBeBinary(comp, edc.IsBlockHashHi)
	util.MustBeBinary(comp, edc.IsBlockHashLo)
	util.MustBeBinary(comp, edc.IsTimestamp)
	util.MustBeBinary(comp, edc.IsAddrHi)
	util.MustBeBinary(comp, edc.IsAddrLo)
	util.MustBeBinary(comp, edc.IsTxRLP)
}

func DefineConstantConstraints(comp *wizard.CompiledIOP, edc *ExecutionDataCollector, name string) {
	// in order for SelectorLastTxBlock to be constrained properly, the values of FirstAbsTxIDBlock and LastAbsTxIDBlock
	// must be constant for the entire segment defined by the block (otherwise SelectorLastTxBlock is meaningless).

	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_%v_CONSTANT_FIRST_ABS_TX_ID_BLOCK_INSIDE_THE_BLOCK_SEGMENT", name, edc.FirstAbsTxIDBlock.GetColID()),
		sym.Mul(
			edc.SelectorBlockDiff, // 1 if edc.BlockId[i] = edc.BlockId[i+1]
			sym.Sub(
				edc.FirstAbsTxIDBlock,
				column.Shift(edc.FirstAbsTxIDBlock, 1),
			),
		),
	)

	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_%v_CONSTANT_LAST_ABS_TX_ID_BLOCK_INSIDE_THE_BLOCK_SEGMENT", name, edc.LastAbsTxIDBlock.GetColID()),
		sym.Mul(
			edc.SelectorBlockDiff, // 1 if edc.BlockId[i] = edc.BlockId[i+1]
			sym.Sub(
				edc.LastAbsTxIDBlock,
				column.Shift(edc.LastAbsTxIDBlock, 1),
			),
		),
	)
	// we do not contrain that FirstAbsTxIDBlock/LastAbsTxIDBlock increases only by 1, this is
	// constrained in the corresponding fetcher and enforced by the projection query (in conjunction
	// with the constancy property)

	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_%v_CONSTANT_ABS_TX_ID_MAX", name, edc.AbsTxIDMax.GetColID()),
		sym.Mul(
			column.Shift(edc.IsActive, 1),
			sym.Sub(
				edc.AbsTxIDMax,
				column.Shift(edc.AbsTxIDMax, 1),
			),
		),
	)
}

func DefineLimbSummary(comp *wizard.CompiledIOP,
	edc *ExecutionDataCollector,
	name string,
	timestamps fetch.TimestampFetcher,
	metadata fetch.BlockTxnMetadata,
	txnData fetch.TxnDataFetcher,
	rlp fetch.RlpTxnFetcher) {

	DefineSelectorConstraints(comp, edc, name)
	DefineIndicatorExclusion(comp, edc, name)
	DefineAlignmentPowers(comp, edc, name)
	DefineIndicatorOrder(comp, edc, name)
	DefineIndicatorsMustBeBinary(comp, edc)
	DefineNumberOfBytesConstraints(comp, edc, name)
	DefineZeroizationConstraints(comp, edc, name)
	DefineConstantConstraints(comp, edc, name)

	// isActive constraints
	// require that isActive is binary

	// require that the isActive filter only contains 1s followed by 0s
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%s_IS_ACTIVE_CONSTRAINT_NO_0_TO_1", name),
		sym.Sub(
			edc.IsActive,
			sym.Mul(
				column.Shift(edc.IsActive, -1),
				edc.IsActive,
			),
		),
	)

	// unaligned limb --- aligned limb constraints
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_UNALIGNED_LIMB_AND_ALIGNED_LIMB_CONSTRAINT", name),
		sym.Sub(
			edc.Limb,
			sym.Mul(
				edc.UnalignedLimb,
				edc.AlignedPow,
			),
		),
	)

	DefineCounterConstraints(comp, edc, name)
	ProjectionQueries(comp, edc, name, timestamps, metadata, txnData, rlp)
}

func AssignLimbSummary(run *wizard.ProverRuntime,
	edc ExecutionDataCollector,
	timestamps fetch.TimestampFetcher,
	metadata fetch.BlockTxnMetadata,
	txnData fetch.TxnDataFetcher,
	rlp fetch.RlpTxnFetcher) {
	size := edc.Limb.Size()
	vect := NewExecutionDataCollectorVectors(size)

	fetchedAbsTxIdMax := rlp.AbsTxNumMax.GetColAssignmentAt(run, 0)
	absTxIdMax := int(fetchedAbsTxIdMax.Uint64())

	absTxCt := 1
	rlpCt := 0
	totalCt := 0

	for blockCt := 0; blockCt < timestamps.Data.Size(); blockCt++ {
		isBlockPresent := metadata.FilterFetched.GetColAssignmentAt(run, blockCt)
		if isBlockPresent.IsOne() {
			// block-wide information
			totalTxBlockField := metadata.TotalNoTxnBlock.GetColAssignmentAt(run, blockCt)
			totalTxBlock := totalTxBlockField.Uint64()
			firstAbsTxIDBlock := metadata.FirstAbsTxId.GetColAssignmentAt(run, blockCt)
			lastAbsTxIDBlock := metadata.LastAbsTxId.GetColAssignmentAt(run, blockCt)
			fetchNoTx := metadata.TotalNoTxnBlock.GetColAssignmentAt(run, blockCt)

			genericLoadFunction := func(opType int, value field.Element) {
				vect.IsActive[totalCt].SetOne()
				vect.SetLimbAndUnalignedLimb(totalCt, value, opType)
				vect.SetCounters(totalCt, blockCt, absTxCt, absTxIdMax)
				vect.SetBlockMetadata(totalCt, totalTxBlockField, firstAbsTxIDBlock, lastAbsTxIDBlock)
			}

			// row 0, load the number of transactions
			vect.IsNoTx[totalCt].SetOne()
			vect.NoBytes[totalCt].SetInt64(noBytesNoTxn)
			genericLoadFunction(loadNoTxn, fetchNoTx)
			totalCt++

			// row 1, load the timestamp
			fetchedTimestamp := timestamps.Data.GetColAssignmentAt(run, blockCt)
			vect.IsTimestamp[totalCt].SetOne()
			vect.NoBytes[totalCt].SetInt64(noBytesTimestamp)
			genericLoadFunction(loadTimestamp, fetchedTimestamp)
			totalCt++

			// row 2, load the Hi part of the blockhash
			fetchedBlockhashHi := field.Zero() // TO BE REPLACED LATER
			vect.IsBlockHashHi[totalCt].SetOne()
			vect.NoBytes[totalCt].SetInt64(noBytesBlockHash)
			genericLoadFunction(loadBlockHashHi, fetchedBlockhashHi)
			vect.AbsTxIDMax[totalCt].Set(&fetchedAbsTxIdMax)
			totalCt++

			// row 3, load the Lo part of the blockhash
			fetchedBlockhashLo := field.Zero() // TO BE REPLACED LATER
			vect.IsBlockHashLo[totalCt].SetOne()
			vect.NoBytes[totalCt].SetInt64(noBytesBlockHash)
			genericLoadFunction(loadBlockHashLo, fetchedBlockhashLo)
			totalCt++

			// iterate through transactions
			for txIdInBlock := 1; txIdInBlock <= int(totalTxBlock); txIdInBlock++ {

				// load the sender address Hi
				fetchedAddrHi := txnData.FromHi.GetColAssignmentAt(run, absTxCt-1)
				vect.IsAddrHi[totalCt].SetOne()
				vect.NoBytes[totalCt].SetInt64(noBytesSenderAddrHi)
				genericLoadFunction(loadSenderAddrHi, fetchedAddrHi)
				totalCt++

				// load the sender address Lo
				fetchedAddrLo := txnData.FromLo.GetColAssignmentAt(run, absTxCt-1)
				vect.IsAddrLo[totalCt].SetOne()
				vect.NoBytes[totalCt].SetInt64(noBytesSenderAddrLo)
				genericLoadFunction(loadSenderAddrLo, fetchedAddrLo)
				totalCt++

				// load the RLP limbs
				currentAbsTxId := field.NewElement(uint64(absTxCt))
				rlpPointerAbsTxId := rlp.AbsTxNum.GetColAssignmentAt(run, rlpCt)
				// add RLP limbs (multiple limbs)
				for currentAbsTxId.Equal(&rlpPointerAbsTxId) {
					// while currentAbsTxId is equal to rlpPointerAbsTxId, namely we are parsing the limbs for the same AbsTxID
					rlpLimb := rlp.Limb.GetColAssignmentAt(run, rlpCt)
					rlpNBytes := rlp.NBytes.GetColAssignmentAt(run, rlpCt)
					vect.IsTxRLP[totalCt].SetOne()
					vect.NoBytes[totalCt].Set(&rlpNBytes)
					genericLoadFunction(loadRlp, rlpLimb)
					totalCt++

					rlpCt++
					rlpPointerAbsTxId = rlp.AbsTxNum.GetColAssignmentAt(run, rlpCt)
				}
				vect.EndOfRlpSegment[totalCt-1].SetOne()
				// increase transaction counter
				absTxCt++
			}

		} else {
			// finished processing all the blocks, move to padding
			// we do not set the isActive filter to 1
			// No more blocks to assign
			break
		}
	} // end of the block for loop

	AssignExecutionDataColumns(run, edc, vect)
	edc.ComputeSelectorBlockDiff.Run(run)
	edc.ComputeSelectorLastTxBlock.Run(run)
	edc.ComputeSelectorEndOfAllTx.Run(run)
	edc.ComputeSelectorAbsTxIDDiff.Run(run)
}

func AssignExecutionDataColumns(run *wizard.ProverRuntime, edc ExecutionDataCollector, vect *ExecutionDataCollectorVectors) {
	run.AssignColumn(edc.BlockID.GetColID(), smartvectors.NewRegular(vect.BlockID))
	run.AssignColumn(edc.AbsTxID.GetColID(), smartvectors.NewRegular(vect.AbsTxID))
	run.AssignColumn(edc.Limb.GetColID(), smartvectors.NewRegular(vect.Limb))
	run.AssignColumn(edc.NoBytes.GetColID(), smartvectors.NewRegular(vect.NoBytes))
	run.AssignColumn(edc.UnalignedLimb.GetColID(), smartvectors.NewRegular(vect.UnalignedLimb))
	run.AssignColumn(edc.AlignedPow.GetColID(), smartvectors.NewRegular(vect.AlignedPow))
	run.AssignColumn(edc.TotalNoTxBlock.GetColID(), smartvectors.NewRegular(vect.TotalNoTxBlock))
	run.AssignColumn(edc.IsActive.GetColID(), smartvectors.NewRegular(vect.IsActive))
	run.AssignColumn(edc.IsNoTx.GetColID(), smartvectors.NewRegular(vect.IsNoTx))
	run.AssignColumn(edc.IsBlockHashHi.GetColID(), smartvectors.NewRegular(vect.IsBlockHashHi))
	run.AssignColumn(edc.IsBlockHashLo.GetColID(), smartvectors.NewRegular(vect.IsBlockHashLo))
	run.AssignColumn(edc.IsTimestamp.GetColID(), smartvectors.NewRegular(vect.IsTimestamp))
	run.AssignColumn(edc.IsTxRLP.GetColID(), smartvectors.NewRegular(vect.IsTxRLP))
	run.AssignColumn(edc.IsAddrHi.GetColID(), smartvectors.NewRegular(vect.IsAddrHi))
	run.AssignColumn(edc.IsAddrLo.GetColID(), smartvectors.NewRegular(vect.IsAddrLo))
	run.AssignColumn(edc.AbsTxIDMax.GetColID(), smartvectors.NewRegular(vect.AbsTxIDMax))
	run.AssignColumn(edc.EndOfRlpSegment.GetColID(), smartvectors.NewRegular(vect.EndOfRlpSegment))
	run.AssignColumn(edc.FirstAbsTxIDBlock.GetColID(), smartvectors.NewRegular(vect.FirstAbsTxIDBlock))
	run.AssignColumn(edc.LastAbsTxIDBlock.GetColID(), smartvectors.NewRegular(vect.LastAbsTxIDBlock))
}
