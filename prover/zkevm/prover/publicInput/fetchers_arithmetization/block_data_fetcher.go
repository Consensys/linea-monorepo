package fetchers_arithmetization

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	commonconstraints "github.com/consensys/linea-monorepo/prover/zkevm/prover/common/common_constraints"
	arith "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/arith_struct"
	util "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/utilities"
)

const (
	// TimestampOffset is the corresponding offset position for the timestamp
	// since it is a shift, -1 means no offset.
	TimestampOffset = -12
)

// BlockDataFetcher is a struct used to fetch the timestamps from the arithmetization's BlockDataCols
type BlockDataFetcher struct {
	// RelBlock is the relative block number, ranging from 1 to the total number of blocks
	RelBlock ifaces.Column
	// timestamp data for the first and last blocks in the conflation, columns of size 1
	FirstTimestamp, LastTimestamp ifaces.Column
	// FirstArith and LastArith are identical to First and Last but are used in the constraints
	// involving arithmetization columns. They are constrained to be constant and via the
	// projection query between the fetcher and the
	FirstArith, LastArith ifaces.Column
	// DataLo contains all the timestamps in the conflation, ordered by block
	DataLo, DataHi ifaces.Column
	// filter on the BlockDataFetcher.Data column
	FilterFetched ifaces.Column

	// filter that selects only timestamp, baseFee and coinbase rows from the
	// arithmetization
	SelTimestampArith, SelBaseFeeArith, SelCoinBaseArith ifaces.Column
	// prover action to compute SelectorTimestamp
	ComputeSelTimestamp, ComputeSelBaseFee, ComputeSelCoinBase wizard.ProverAction

	// the absolute ID of the first block number
	FirstBlockID ifaces.Column
	// the absolute ID of the last block number
	LastBlockID ifaces.Column
	// the absolute ID of the first block number
	FirstBlockIDArith ifaces.Column
	// the absolute ID of the last block number
	LastBlockIDArith ifaces.Column
	// BaseFee contains the base fee for each block
	BaseFee ifaces.Column
	// CoinBase contains the coin base for each block
	CoinBase ifaces.Column
}

// NewBlockDataFetcher returns a new BlockDataFetcher with initialized columns that are not constrained.
func NewBlockDataFetcher(comp *wizard.CompiledIOP, name string, bdc *arith.BlockDataCols) *BlockDataFetcher {

	size := bdc.Ct.Size()

	res := &BlockDataFetcher{
		RelBlock:          util.CreateCol(name, "REL_BLOCK", size, comp),
		DataLo:            util.CreateCol(name, "DATA_LO", size, comp),
		FilterFetched:     util.CreateCol(name, "FILTER_FETCHED", size, comp),
		FirstBlockID:      util.CreateCol(name, "FIRST_BLOCK_ID", size, comp),
		LastBlockID:       util.CreateCol(name, "LAST_BLOCK_ID", size, comp),
		FirstTimestamp:    util.CreateCol(name, "FIRST", size, comp),
		LastTimestamp:     util.CreateCol(name, "LAST", size, comp),
		FirstBlockIDArith: util.CreateCol(name, "FIRST_BLOCK_ID_ARITHMETIZATION", size, comp),
		LastBlockIDArith:  util.CreateCol(name, "LAST_BLOCK_ID_ARITHMETIZATION", size, comp),
		FirstArith:        util.CreateCol(name, "FIRST_ARITHMETIZATION", size, comp),
		LastArith:         util.CreateCol(name, "LAST_ARITHMETIZATION", size, comp),
		BaseFee:           util.CreateCol(name, "BASE_FEE", size, comp),
		CoinBase:          util.CreateCol(name, "COIN_BASE", size, comp),
	}

	return res
}

// ConstrainFirstAndLastBlockID constraing the values of FirstBlockID and LastBlockID
func ConstrainFirstAndLastBlockID(comp *wizard.CompiledIOP, fetcher *BlockDataFetcher, name string, bdc *arith.BlockDataCols) {

	commonconstraints.MustBeConstant(comp, fetcher.FirstBlockID)
	commonconstraints.MustBeConstant(comp, fetcher.LastBlockID)
	commonconstraints.MustBeConstant(comp, fetcher.FirstBlockIDArith)
	commonconstraints.MustBeConstant(comp, fetcher.LastBlockIDArith)

	// Constrain the First Block ID
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%s_%s_%s", name, "FIRST_BLOCK_ID_GLOBAL", fetcher.FirstBlockID.GetColID()),
		sym.Mul(
			// Using the timestamp selector allows trimming out the padding rows
			// from the equality check between the FirstBlock and the
			// FirstBlockIDArith column. The difference between these two
			// columns is that the FirstBlock column is subject to padding while
			// the other one is not and is expected to be fully constant.
			fetcher.SelTimestampArith,
			sym.Sub(
				bdc.FirstBlock,
				fetcher.FirstBlockIDArith,
			),
		),
	)

	// FilterFetched is already constrained in the fetcher, no need to constrain it again
	// two cases: Case 1: FilterFetched is not completely filled with 1s (we have a border between 1s and 0s)
	comp.InsertGlobal(0, ifaces.QueryIDf("%s_%s_%s", name, "LAST_BLOCK_ID_GLOBAL", fetcher.LastBlockID.GetColID()),
		sym.Mul(
			fetcher.FilterFetched,
			sym.Sub(1,
				column.Shift(fetcher.FilterFetched, 1),
			),
			sym.Sub(
				fetcher.LastBlockID,
				sym.Add(
					fetcher.RelBlock,
					fetcher.FirstBlockID,
					-1,
				),
			),
		),
	)

	// Case 2: FilterFetched is completely filled with 1s, in which case we do not have a border between 1s and 0s
	comp.InsertLocal(0, ifaces.QueryIDf("%s_%s_%s", name, "LAST_BLOCK_ID_LOCAL", fetcher.LastBlockID.GetColID()),
		sym.Mul(
			column.Shift(fetcher.FilterFetched, -1),
			sym.Sub(
				column.Shift(fetcher.LastBlockID, -1),
				sym.Add(
					column.Shift(fetcher.RelBlock, -1),
					column.Shift(fetcher.FirstBlockID, -1),
					-1,
				),
			),
		),
	)

}

// DefineBlockDataFetcher specifies the constraints of the BlockDataFetcher with respect to the BlockDataCols
func DefineBlockDataFetcher(comp *wizard.CompiledIOP, fetcher *BlockDataFetcher, name string, bdc *arith.BlockDataCols) {

	var (
		timestampField  = util.GetTimestampField()
		baseFeeField    = util.GetBaseFeeField()
		coinBaseField   = util.GetCoinBaseField()
		selTimestampCtx = makeBtcInstSelector(comp, bdc, timestampField)
		selBaseFeeCtx   = makeBtcInstSelector(comp, bdc, baseFeeField)
		selCoinBaseCtx  = makeBtcInstSelector(comp, bdc, coinBaseField)
	)

	fetcher.SelTimestampArith, fetcher.ComputeSelTimestamp = selTimestampCtx.GetColumnAndProverAction()
	fetcher.SelBaseFeeArith, fetcher.ComputeSelBaseFee = selBaseFeeCtx.GetColumnAndProverAction()
	fetcher.SelCoinBaseArith, fetcher.ComputeSelCoinBase = selCoinBaseCtx.GetColumnAndProverAction()

	commonconstraints.MustBeConstant(comp, fetcher.FirstTimestamp)
	commonconstraints.MustBeConstant(comp, fetcher.LastTimestamp)
	commonconstraints.MustBeConstant(comp, fetcher.FirstArith)
	commonconstraints.MustBeConstant(comp, fetcher.LastArith)
	commonconstraints.MustBeConstant(comp, fetcher.CoinBase)
	commonconstraints.MustBeConstant(comp, fetcher.BaseFee)
	commonconstraints.MustBeActivationColumns(comp, fetcher.FilterFetched)

	// constrain fetcher.First to contain the value of the first block's timestamp, using all the timestamps in fetcher.Data
	comp.InsertLocal(
		0,
		ifaces.QueryIDf("%s_%s", name, "FIRST_LOCAL_LO"),
		sym.Sub(
			fetcher.FirstTimestamp,
			fetcher.DataLo, // fetcher.Data is constrained in the projection query
		),
	)

	// constrain fetcher.Last to contain the value of the last block's timestamp,
	comp.InsertLocal(
		0,
		ifaces.QueryIDf("%s_%s", name, "LAST_LOCAL"),
		sym.Sub(
			column.Shift(fetcher.LastArith, -1),
			column.Shift(bdc.DataLo, TimestampOffset),
		),
	)

	// require that the filter on fetched data is a binary column
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%s_FILTER_ON_FETCHED_CONSTRAINT_MUST_BE_BINARY", name),
		sym.Mul(
			fetcher.FilterFetched,
			sym.Sub(fetcher.FilterFetched, 1),
		),
	)

	// require that the filter on fetched timestamps only contains 1s followed by 0s
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%s_FILTER_ON_FETCHED_CONSTRAINT_NO_0_TO_1", name),
		sym.Sub(
			fetcher.FilterFetched,
			sym.Mul(
				column.Shift(fetcher.FilterFetched, -1),
				fetcher.FilterFetched),
		),
	)

	// a projection query to check that the timestamp data is fetched correctly
	comp.InsertProjection(
		ifaces.QueryIDf("%s_TIMESTAMP_PROJECTION", name),
		query.ProjectionInput{
			ColumnA: []ifaces.Column{
				fetcher.RelBlock,
				fetcher.DataLo,
				fetcher.FirstBlockID,
				fetcher.LastBlockID,
				fetcher.FirstTimestamp,
				fetcher.LastTimestamp,
			},
			ColumnB: []ifaces.Column{
				bdc.RelBlock,
				bdc.DataLo,
				fetcher.FirstBlockIDArith,
				fetcher.LastBlockIDArith,
				fetcher.FirstArith,
				fetcher.LastArith,
			},
			// the filter is structured as an isActive column
			FilterA: fetcher.FilterFetched,
			// filter lights up on the arithmetization's BlockDataCols rows that
			// contain timestamp data
			FilterB: fetcher.SelTimestampArith,
		})

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%s_BASEFEE_FETCHING", name),
		sym.Mul(
			fetcher.SelBaseFeeArith,
			sym.Sub(
				fetcher.BaseFee,
				bdc.DataLo,
			),
		),
	)

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%s_COINBASE_FETCHING", name),
		sym.Mul(
			fetcher.SelCoinBaseArith,
			sym.Sub(
				sym.Add(
					bdc.DataLo,
					sym.Mul(bdc.DataHi, 1<<32),
				),
				fetcher.CoinBase,
			),
		),
	)

	// constrain the First/Last Block ID counters
	ConstrainFirstAndLastBlockID(comp, fetcher, name, bdc)
}

// AssignBlockDataFetcher assigns the data in the BlockDataFetcher using data fetched from the BlockDataCols
func AssignBlockDataFetcher(run *wizard.ProverRuntime, fetcher *BlockDataFetcher, bdc *arith.BlockDataCols) {

	var (
		first, last, firstBlockID field.Element
		// get the hardcoded timestamp flag
		timestampField = util.GetTimestampField()
		baseFeeField   = util.GetBaseFeeField()
		coinBaseField  = util.GetCoinBaseField()

		// inst is the flag that specifies the row type
		inst        = bdc.Inst.GetColAssignment(run)
		ct          = bdc.Ct.GetColAssignment(run)
		start, stop = smartvectors.CoCompactRange(ct, inst)

		// initialize empty fetched data and filter on the fetched data
		size          = ct.Len()
		relBlock      = make([]field.Element, size)
		data          = make([]field.Element, size)
		filterFetched = make([]field.Element, size)
		filterArith   = make([]field.Element, stop-start)

		// counter is used to populate filter.Data and will increment every
		// time we find a new timestamp
		counter = 0

		// baseFee tracks the value of the base and is zero by default
		// coinBase tracks the value of the coin base and is zero if the opcode
		// is never used. (but that is normally impossible since these are
		// needed to argue how the sequencer is rewarded for every block, even
		// the empty ones)
		baseFee, coinBase field.Element
	)

	for i := start; i < stop; i++ {

		var (
			inst = inst.GetPtr(i)
			ct   = ct.GetPtr(i)
		)

		if inst.Equal(&baseFeeField) && ct.IsZero() {
			baseFee = bdc.DataLo.GetColAssignmentAt(run, i)
		}

		if inst.Equal(&coinBaseField) && ct.IsZero() {
			coinBase = bdc.DataLo.GetColAssignmentAt(run, i)
		}

		if inst.Equal(&timestampField) && ct.IsZero() {
			// the row type is a timestamp-encoding row
			timestamp := bdc.DataLo.GetColAssignmentAt(run, i)
			// in the arithmetization, relBlock is the relative block number inside the conflation
			fetchedRelBlock := bdc.RelBlock.GetColAssignmentAt(run, i)
			if fetchedRelBlock.IsOne() {
				// the first relative block has code 0x1
				first.Set(&timestamp)
				// set the first absolute block ID
				firstBlockID = bdc.FirstBlock.GetColAssignmentAt(run, i)
			}
			// continuously update the last timestamp value
			last.Set(&timestamp)
			// update counters and timestamp data
			filterFetched[counter].SetOne()
			relBlock[counter].Set(&fetchedRelBlock)
			// update the arithmetization filter
			filterArith[i-start].SetOne()

			data[counter].Set(&timestamp)
			counter++
		}
	}
	// compute the last absolute block ID
	var lastBlockID field.Element
	fieldCounter := field.NewElement(uint64(counter - 1))
	lastBlockID.Add(&firstBlockID, &fieldCounter)

	// assign the fetcher columns
	run.AssignColumn(fetcher.FirstTimestamp.GetColID(), smartvectors.NewConstant(first, size))
	run.AssignColumn(fetcher.LastTimestamp.GetColID(), smartvectors.NewConstant(last, size))
	run.AssignColumn(fetcher.FirstArith.GetColID(), smartvectors.NewConstant(first, size))
	run.AssignColumn(fetcher.LastArith.GetColID(), smartvectors.NewConstant(last, size))
	run.AssignColumn(fetcher.RelBlock.GetColID(), smartvectors.RightZeroPadded(relBlock, size))
	run.AssignColumn(fetcher.DataLo.GetColID(), smartvectors.RightZeroPadded(data, size))
	run.AssignColumn(fetcher.FilterFetched.GetColID(), smartvectors.RightZeroPadded(filterFetched, size))
	run.AssignColumn(fetcher.FirstBlockID.GetColID(), smartvectors.NewConstant(firstBlockID, size))
	run.AssignColumn(fetcher.LastBlockID.GetColID(), smartvectors.NewConstant(lastBlockID, size))
	run.AssignColumn(fetcher.FirstBlockIDArith.GetColID(), smartvectors.NewConstant(firstBlockID, size))
	run.AssignColumn(fetcher.LastBlockIDArith.GetColID(), smartvectors.NewConstant(lastBlockID, size))
	run.AssignColumn(fetcher.BaseFee.GetColID(), smartvectors.NewConstant(baseFee, size))
	run.AssignColumn(fetcher.CoinBase.GetColID(), smartvectors.NewConstant(coinBase, size))

	// assign the selector columns using the is zero prover instance
	fetcher.ComputeSelTimestamp.Run(run)
	fetcher.ComputeSelCoinBase.Run(run)
	fetcher.ComputeSelBaseFee.Run(run)
}

// makeBtcInstSelector constructs a selector for the BTC module marking the
// first row of every frame corresponding to the target instruction.
//
// Concretely, the selector flags up when bdc.Inst == inst && bdc.Ct == 0
//
// The implementation assumes bdc.Inst < 256 and that Ct is really small, both
// are arguably sound and future-proof assumption; CT is the number of rows that
// the module spend on each instruction and the instruction space is defined to
// be a single byte on the EVM).
//
// Relying on these, the selector is implemented as a check that
// 256*Ct+(inst-INST) == 0.
func makeBtcInstSelector(
	comp *wizard.CompiledIOP, bdc *arith.BlockDataCols, inst field.Element,
) *dedicated.IsZeroCtx {

	return dedicated.IsZero(
		comp,
		sym.Add(
			sym.Sub(bdc.Inst, inst),
			sym.Mul(256, bdc.Ct),
		),
	)
}
