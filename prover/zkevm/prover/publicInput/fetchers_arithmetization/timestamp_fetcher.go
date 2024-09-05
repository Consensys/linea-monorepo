package fetchers_arithmetization

import (
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/accessors"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/dedicated"
	"github.com/consensys/zkevm-monorepo/prover/protocol/dedicated/projection"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/zkevm-monorepo/prover/symbolic"
	arith "github.com/consensys/zkevm-monorepo/prover/zkevm/prover/publicInput/arith_struct"
	util "github.com/consensys/zkevm-monorepo/prover/zkevm/prover/publicInput/utilities"
)

const (
	TimestampOffset = -6 // the corresponding offset position for the timestamp
)

// TimestampFetcher is a struct used to fetch the timestamps from the arithmetization's BlockDataCols
type TimestampFetcher struct {
	// RelBlock is the relative block number, ranging from 1 to the total number of blocks
	RelBlock ifaces.Column
	// timestamp data for the first and last blocks in the conflation, columns of size 1
	First, Last ifaces.Column
	// Data contains all the timestamps in the conflation, ordered by block
	Data ifaces.Column
	// filter on the TimestampFetcher.Data column
	FilterFetched ifaces.Column
	// filter on the arithmetization's BlockDataCols
	SelectorTimestamp ifaces.Column
	// prover action to compute SelectorTimestamp
	ComputeSelectorTimestamp wizard.ProverAction
	// the absolute ID of the first block number
	FirstBlockID ifaces.Column
	// the absolute ID of the last block number
	LastBlockID ifaces.Column
}

// NewTimestampFetcher returns a new TimestampFetcher with initialized columns that are not constrained.
func NewTimestampFetcher(comp *wizard.CompiledIOP, name string, bdc *arith.BlockDataCols) TimestampFetcher {
	size := bdc.Ct.Size()
	res := TimestampFetcher{
		RelBlock:      util.CreateCol(name, "REL_BLOCK", size, comp),
		First:         util.CreateCol(name, "FIRST", 1, comp),
		Last:          util.CreateCol(name, "LAST", 1, comp),
		Data:          util.CreateCol(name, "DATA", size, comp),
		FilterFetched: util.CreateCol(name, "FILTER_FETCHED", size, comp),
		FirstBlockID:  util.CreateCol(name, "FIRST_BLOCK_ID", 1, comp),
		LastBlockID:   util.CreateCol(name, "LAST_BLOCK_ID", 1, comp),
	}
	return res
}

// ConstrainFirstAndLastBlockID constraing the values of FirstBlockID and LastBlockID
func ConstrainFirstAndLastBlockID(comp *wizard.CompiledIOP, fetcher *TimestampFetcher, name string, bdc *arith.BlockDataCols) {
	// set the fetcher columns as public for accessors
	comp.Columns.SetStatus(fetcher.FirstBlockID.GetColID(), column.Proof)
	comp.Columns.SetStatus(fetcher.LastBlockID.GetColID(), column.Proof)
	// get accessors
	accessFirstBlockID := accessors.NewFromPublicColumn(fetcher.FirstBlockID, 0)
	accessLastBlockID := accessors.NewFromPublicColumn(fetcher.LastBlockID, 0)
	// Constrain the First Block ID
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%s_%s_%s", name, "FIRST_BLOCK_ID_GLOBAL", fetcher.FirstBlockID.GetColID()),
		sym.Mul(
			fetcher.SelectorTimestamp, // select only non-padding, valid rows.
			sym.Sub(
				bdc.FirstBlock,
				accessFirstBlockID,
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
				accessLastBlockID,
				sym.Add(
					fetcher.RelBlock,
					accessFirstBlockID,
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
				accessLastBlockID,
				sym.Add(
					column.Shift(fetcher.RelBlock, -1),
					accessFirstBlockID,
					-1,
				),
			),
		),
	)

}

// DefineTimestampFetcher specifies the constraints of the TimestampFetcher with respect to the BlockDataCols
func DefineTimestampFetcher(comp *wizard.CompiledIOP, fetcher *TimestampFetcher, name string, bdc *arith.BlockDataCols) {
	timestampField := util.GetTimestampField()
	// constrain the fetcher.SelectorTimestamp column, which will be the filter for the arithmetization's BlockDataCols
	fetcher.SelectorTimestamp, fetcher.ComputeSelectorTimestamp = dedicated.IsZero(
		comp,
		sym.Sub(
			bdc.Inst,
			timestampField, // check that the Inst field indicates a timestamp row
		),
	)
	// set the fetcher columns as public for accessors
	comp.Columns.SetStatus(fetcher.First.GetColID(), column.Proof)
	comp.Columns.SetStatus(fetcher.Last.GetColID(), column.Proof)
	// get the accessors
	accessFirst := accessors.NewFromPublicColumn(fetcher.First, 0)
	accessLast := accessors.NewFromPublicColumn(fetcher.Last, 0)

	// constrain fetcher.First to contain the value of the first block's timestamp, using all the timestamps in fetcher.Data
	comp.InsertLocal(
		0,
		ifaces.QueryIDf("%s_%s", name, "FIRST_LOCAL"),
		sym.Sub(
			accessFirst,
			fetcher.Data, // fetcher.Data is constrained in the projection query
		),
	)

	// constrain fetcher.Last to contain the value of the last block's timestamp,
	comp.InsertLocal(
		0,
		ifaces.QueryIDf("%s_%s", name, "LAST_LOCAL"),
		sym.Sub(
			accessLast,
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

	// the table with the data we fetch from the arithmetization columns BlockDataCols
	fetcherTable := []ifaces.Column{
		fetcher.RelBlock,
		fetcher.Data,
	}
	// the BlockDataCols we extract timestamp data from, and which we will use to check for consistency
	arithTable := []ifaces.Column{
		bdc.RelBlock,
		bdc.DataLo,
	}

	// a projection query to check that the timestamp data is fetched correctly
	projection.InsertProjection(comp,
		ifaces.QueryIDf("%s_TIMESTAMP_PROJECTION", name),
		fetcherTable,
		arithTable,
		fetcher.FilterFetched,
		fetcher.SelectorTimestamp, // filter lights up on the arithmetization's BlockDataCols rows that contain timestamp data
	)

	// constrain the First/Last Block ID counters
	ConstrainFirstAndLastBlockID(comp, fetcher, name, bdc)

}

// AssignTimestampFetcher assigns the data in the TimestampFetcher using data fetched from the BlockDataCols
func AssignTimestampFetcher(run *wizard.ProverRuntime, fetcher TimestampFetcher, bdc *arith.BlockDataCols) {

	var first, last, firstBlockID field.Element
	// get the hardcoded timestamp flag
	timestampField := util.GetTimestampField()

	// initialize empty fetched data and filter on the fetched data
	size := bdc.Ct.Size()
	relBlock := make([]field.Element, size)
	data := make([]field.Element, size)
	filterFetched := make([]field.Element, size)

	// counter is used to populate filter.Data and will increment every time we find a new timestamp
	counter := 0

	for i := 0; i < size; i++ {
		// inst is the flag that specifies the row type
		inst := bdc.Inst.GetColAssignmentAt(run, i)
		if inst.Equal(&timestampField) {
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

			data[counter].Set(&timestamp)
			counter++
		}
	}
	// compute the last absolute block ID
	var lastBlockID field.Element
	fieldCounter := field.NewElement(uint64(counter - 1))
	lastBlockID.Add(&firstBlockID, &fieldCounter)

	// assign the fetcher columns
	run.AssignColumn(fetcher.First.GetColID(), smartvectors.NewRegular([]field.Element{first}))
	run.AssignColumn(fetcher.Last.GetColID(), smartvectors.NewRegular([]field.Element{last}))
	run.AssignColumn(fetcher.RelBlock.GetColID(), smartvectors.NewRegular(relBlock))
	run.AssignColumn(fetcher.Data.GetColID(), smartvectors.NewRegular(data))
	run.AssignColumn(fetcher.FilterFetched.GetColID(), smartvectors.NewRegular(filterFetched))
	run.AssignColumn(fetcher.FirstBlockID.GetColID(), smartvectors.NewRegular([]field.Element{firstBlockID}))
	run.AssignColumn(fetcher.LastBlockID.GetColID(), smartvectors.NewRegular([]field.Element{lastBlockID}))
	// assign the SelectorTimestamp using the ComputeSelectorTimestamp prover action
	fetcher.ComputeSelectorTimestamp.Run(run)
}
