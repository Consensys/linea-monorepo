package fetchers_arithmetization

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/byte32cmp"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	commonconstraints "github.com/consensys/linea-monorepo/prover/zkevm/prover/common/common_constraints"
	arith "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/arith_struct"
	util "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/utilities"
)

const (
	// TimestampOffset is the corresponding offset position for the timestamp
	// since it is a shift, -1 means no offset.
	TimestampOffset = -12
	// dataLoPartStart is the starting position of the dataLo part where the timestamp data is stored
	// limbBitSize is the bit size of each limb
	limbBitSize = 16
)

// BlockDataFetcher is a struct used to fetch the timestamps from the arithmetization's BlockDataCols
type BlockDataFetcher struct {
	// RelBlock is the relative block number, ranging from 1 to the total number of blocks
	RelBlock ifaces.Column
	// timestamp data for the first and last blocks in the conflation, columns of size 1
	First, Last [common.NbLimbU128]ifaces.Column
	// FirstArith and LastArith are identical to First and Last but are used in the constraints
	// involving arithmetization columns. They are constrained to be constant and via the
	// projection query between the fetcher and the
	FirstArith, LastArith [common.NbLimbU128]ifaces.Column
	// Data contains all the timestamps in the conflation, ordered by block
	Data ifaces.Column
	// filter on the BlockDataFetcher.Data column
	FilterFetched ifaces.Column
	// filter on the Arithmetization's columns
	FilterArith ifaces.Column
	// filter that selects only timestamp rows from the arithmetization
	SelectorTimestamp ifaces.Column
	// prover action to compute SelectorTimestamp
	ComputeSelectorTimestamp wizard.ProverAction
	// since there are two timestamp columns, we need to use Ct in order to select only one
	SelectorCt ifaces.Column
	// prover action to compute SelectorCt
	ComputeSelectorCt wizard.ProverAction
	// the absolute ID of the first block number
	FirstBlockID [common.NbLimbU48]ifaces.Column
	// the absolute ID of the last block number
	LastBlockID [common.NbLimbU48]ifaces.Column
	// the absolute ID of the first block number
	FirstBlockIDArith [common.NbLimbU48]ifaces.Column
	// the absolute ID of the last block number
	LastBlockIDArith [common.NbLimbU48]ifaces.Column
	// the last block ID minus the first block ID, used to compute the difference
	// between the first and last blocks and compare it to the RelBlock-1
	LastMinusFirstBlock       byte32cmp.LimbColumns
	LastMinusFirstBlockAction wizard.ProverAction
	// a constant columns that contains -1 at every position
	minusOne ifaces.Column
}

// NewBlockDataFetcher returns a new BlockDataFetcher with initialized columns that are not constrained.
func NewBlockDataFetcher(comp *wizard.CompiledIOP, name string, bdc *arith.BlockDataCols) *BlockDataFetcher {

	size := bdc.Ct.Size()

	res := &BlockDataFetcher{
		RelBlock:      util.CreateColBase(name, "REL_BLOCK", size, comp),
		FilterFetched: util.CreateColBase(name, "FILTER_FETCHED", size, comp),
		FilterArith:   util.CreateColBase(name, "FILTER_ARITHMETIZATION", size, comp),
	}

	for i := range res.Data {
		res.Data[i] = util.CreateColBase(name, fmt.Sprintf("DATA_%d", i), size, comp)
		res.First[i] = util.CreateColBase(name, fmt.Sprintf("FIRST_%d", i), size, comp)
		res.Last[i] = util.CreateColBase(name, fmt.Sprintf("LAST_%d", i), size, comp)
		res.FirstArith[i] = util.CreateColBase(name, fmt.Sprintf("FIRST_ARITHMETIZATION_%d", i), size, comp)
		res.LastArith[i] = util.CreateColBase(name, fmt.Sprintf("LAST_ARITHMETIZATION_%d", i), size, comp)
	}

	for i := range res.FirstBlockID {
		res.FirstBlockID[i] = util.CreateColBase(name, fmt.Sprintf("FIRST_BLOCK_ID_%d", i), size, comp)
		res.LastBlockID[i] = util.CreateColBase(name, fmt.Sprintf("LAST_BLOCK_ID_%d", i), size, comp)
		res.FirstBlockIDArith[i] = util.CreateColBase(name, fmt.Sprintf("FIRST_BLOCK_ID_ARITHMETIZATION_%d", i), size, comp)
		res.LastBlockIDArith[i] = util.CreateColBase(name, fmt.Sprintf("LAST_BLOCK_ID_ARITHMETIZATION_%d", i), size, comp)

	}

	return res
}

// ConstrainFirstAndLastBlockID constraing the values of FirstBlockID and LastBlockID
func ConstrainFirstAndLastBlockID(comp *wizard.CompiledIOP, fetcher *BlockDataFetcher, name string, bdc *arith.BlockDataCols) {
	fetcher.LastMinusFirstBlock, fetcher.LastMinusFirstBlockAction = byte32cmp.NewMultiLimbAdd(comp,
		&byte32cmp.MultiLimbAddIn{
			Name: fmt.Sprintf("%s_LAST_BLOCK_ID_GLOBAL_INTERM_%s", name, fetcher.LastBlockID[0].GetColID()),
			ALimbs: byte32cmp.LimbColumns{
				Limbs:       fetcher.LastBlockID[:],
				LimbBitSize: limbBitSize,
				IsBigEndian: true,
			},
			BLimbs: byte32cmp.LimbColumns{
				Limbs:       fetcher.FirstBlockID[:],
				LimbBitSize: limbBitSize,
				IsBigEndian: true,
			},
			Mask: sym.NewVariable(fetcher.FilterFetched),
		},
		false, // we want a substraction
	)

	for i := range fetcher.FirstBlockID {
		commonconstraints.MustBeConstant(comp, fetcher.FirstBlockID[i])
		commonconstraints.MustBeConstant(comp, fetcher.LastBlockID[i])
		commonconstraints.MustBeConstant(comp, fetcher.FirstBlockIDArith[i])
		commonconstraints.MustBeConstant(comp, fetcher.LastBlockIDArith[i])

		// Constrain the First Block ID
		comp.InsertGlobal(
			0,
			ifaces.QueryIDf("%s_FIRST_BLOCK_ID_GLOBAL_%s_%d", name, fetcher.FirstBlockID[i].GetColID(), i),
			sym.Mul(
				fetcher.FilterArith, // select only non-padding, valid rows.
				sym.Sub(
					bdc.FirstBlock[i],
					fetcher.FirstBlockIDArith[i],
				),
			),
		)
	}

	// FilterFetched is already constrained in the fetcher, no need to constrain it again
	// two cases: Case 1: FilterFetched is not completely filled with 1s (we have a border between 1s and 0s)
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_LAST_BLOCK_ID_GLOBAL_%d", name, common.NbLimbU48-1),
		sym.Mul(
			fetcher.FilterFetched,
			sym.Sub(1,
				column.Shift(fetcher.FilterFetched, 1),
			),
			sym.Sub(
				fetcher.LastMinusFirstBlock.Limbs[common.NbLimbU48-1],
				sym.Add(
					fetcher.RelBlock,
					-1,
				),
			),
		),
	)

	// Case 2: FilterFetched is completely filled with 1s, in which case we do not have a border between 1s and 0s
	comp.InsertLocal(0,
		ifaces.QueryIDf("%s_LAST_BLOCK_ID_LOCAL_%d", name, common.NbLimbU48-1),
		sym.Mul(
			column.Shift(fetcher.FilterFetched, -1),
			sym.Sub(
				fetcher.LastMinusFirstBlock.Limbs[common.NbLimbU48-1],
				sym.Add(
					column.Shift(fetcher.RelBlock, -1),
					-1,
				),
			),
		),
	)
}

// DefineBlockDataFetcher specifies the constraints of the BlockDataFetcher with respect to the BlockDataCols
func DefineBlockDataFetcher(comp *wizard.CompiledIOP, fetcher *BlockDataFetcher, name string, bdc *arith.BlockDataCols) {
	timestampField := util.GetTimestampField()
	// constrain the fetcher.SelectorTimestamp column, which will be used to compute the filter for the arithmetization's BlockDataCols
	fetcher.SelectorTimestamp, fetcher.ComputeSelectorTimestamp = dedicated.IsZero(
		comp,
		sym.Sub(
			bdc.Inst,
			timestampField, // check that the Inst field indicates a timestamp row
		),
	).GetColumnAndProverAction()
	// constrain the fetcher.SelectorCt column, which will be used to compute the filter for the arithmetization's BlockDataCols
	fetcher.SelectorCt, fetcher.ComputeSelectorCt = dedicated.IsZero(
		comp,
		ifaces.ColumnAsVariable(bdc.Ct), // pick the spots where Ct=0
	).GetColumnAndProverAction()
	// constrain the entire arithmetization filtering column, using SelectorCt and SelectorTimestamp
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%s_CONSTRAINT_ARITHMETIZATION_FILTERING_COLUMN", name),
		sym.Sub(
			fetcher.FilterArith, // fetcher.FilterArith must be 1 if and only if SelectorCt and SelectorTimestamp are both 1
			sym.Mul(
				fetcher.SelectorCt,
				fetcher.SelectorTimestamp,
			),
		),
	)

	for i := range fetcher.First {
		commonconstraints.MustBeConstant(comp, fetcher.First[i])
		commonconstraints.MustBeConstant(comp, fetcher.Last[i])
		commonconstraints.MustBeConstant(comp, fetcher.FirstArith[i])
		commonconstraints.MustBeConstant(comp, fetcher.LastArith[i])

		// constrain fetcher.First to contain the value of the first block's timestamp, using all the timestamps in fetcher.Data
		comp.InsertLocal(
			0,
			ifaces.QueryIDf("%s_FIRST_LOCAL_%d", name, i),
			sym.Sub(
				fetcher.First[i],
				fetcher.Data[i], // fetcher.Data is constrained in the projection query
			),
		)

		// constrain fetcher.Last to contain the value of the last block's timestamp,
		comp.InsertLocal(
			0,
			ifaces.QueryIDf("%s_LAST_LOCAL_%d", name, i),
			sym.Sub(
				column.Shift(fetcher.LastArith[i], -1),
				column.Shift(bdc.Data[common.NbLimbU128+i], TimestampOffset),
			),
		)
	}

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
	fetcherTable := []ifaces.Column{fetcher.RelBlock}
	fetcherTable = append(fetcherTable, fetcher.Data[:]...)
	fetcherTable = append(fetcherTable, fetcher.FirstBlockID[:]...)
	fetcherTable = append(fetcherTable, fetcher.LastBlockID[:]...)
	fetcherTable = append(fetcherTable, fetcher.First[:]...)
	fetcherTable = append(fetcherTable, fetcher.Last[:]...)

	// the BlockDataCols we extract timestamp data from, and which we will use to check for consistency
	arithTable := []ifaces.Column{bdc.RelBlock}
	arithTable = append(arithTable, bdc.Data[common.NbLimbU128:]...)
	arithTable = append(arithTable, fetcher.FirstBlockIDArith[:]...)
	arithTable = append(arithTable, fetcher.LastBlockIDArith[:]...)
	arithTable = append(arithTable, fetcher.FirstArith[:]...)
	arithTable = append(arithTable, fetcher.LastArith[:]...)

	// a projection query to check that the timestamp data is fetched correctly
	comp.InsertProjection(
		ifaces.QueryIDf("%s_TIMESTAMP_PROJECTION", name),
		query.ProjectionInput{
			ColumnA: fetcherTable,
			ColumnB: arithTable,
			FilterA: fetcher.FilterFetched,
			// filter lights up on the arithmetization's BlockDataCols rows that contain timestamp data
			FilterB: fetcher.FilterArith,
		})

	// constrain the First/Last Block ID counters
	ConstrainFirstAndLastBlockID(comp, fetcher, name, bdc)

}

// AssignBlockDataFetcher assigns the data in the BlockDataFetcher using data fetched from the BlockDataCols
func AssignBlockDataFetcher(run *wizard.ProverRuntime, fetcher *BlockDataFetcher, bdc *arith.BlockDataCols) {

	var (
		firstBlockID [common.NbLimbU48]field.Element

		first, last, timestamp [common.NbLimbU128]field.Element
		// get the hardcoded timestamp flag
		timestampField = util.GetTimestampField()

		// inst is the flag that specifies the row type
		inst        = bdc.Inst.GetColAssignment(run)
		ct          = bdc.Ct.GetColAssignment(run)
		start, stop = smartvectors.CoCompactRange(ct, inst)

		// initialize empty fetched data and filter on the fetched data
		size          = ct.Len()
		relBlock      = make([]field.Element, size)
		filterFetched = make([]field.Element, size)
		filterArith   = make([]field.Element, stop-start)

		data [common.NbLimbU128][]field.Element

		// counter is used to populate filter.Data and will increment every
		// time we find a new timestamp
		counter uint64 = 0
	)

	for i := range data {
		data[i] = make([]field.Element, size)
	}

	for i := start; i < stop; i++ {

		var (
			inst = inst.GetPtr(i)
			ct   = ct.GetPtr(i)
		)

		if inst.Equal(&timestampField) && ct.IsZero() {
			// the row type is a timestamp-encoding row
			for j := range timestamp {
				timestamp[j] = bdc.Data[common.NbLimbU128+j].GetColAssignmentAt(run, i)
			}
			// in the arithmetization, relBlock is the relative block number inside the conflation
			fetchedRelBlock := bdc.RelBlock.GetColAssignmentAt(run, i)
			if fetchedRelBlock.IsOne() {
				// the first relative block has code 0x1
				for j := range first {
					first[j].Set(&timestamp[j])
				}
				// set the first absolute block ID
				for j := range firstBlockID {
					firstBlockID[j] = bdc.FirstBlock[j].GetColAssignmentAt(run, i)
				}
			}
			// continuously update the last timestamp value
			for j := range last {
				last[j].Set(&timestamp[j])
			}
			// update counters and timestamp data
			filterFetched[counter].SetOne()
			relBlock[counter].Set(&fetchedRelBlock)
			// update the arithmetization filter
			filterArith[i-start].SetOne()

			for j := range data {
				data[j][counter].Set(&timestamp[j])
			}
			counter++
		}
	}

	// compute the last absolute block ID
	lastBlockID := util.Multi16bitLimbAdd(firstBlockID[:], counter-1)

	// assign the fetcher columns
	run.AssignColumn(fetcher.RelBlock.GetColID(), smartvectors.RightZeroPadded(relBlock, size))
	run.AssignColumn(fetcher.FilterFetched.GetColID(), smartvectors.RightZeroPadded(filterFetched, size))
	run.AssignColumn(fetcher.FilterArith.GetColID(), smartvectors.FromCompactWithRange(filterArith, start, stop, size))

	for i := range common.NbLimbU128 {
		run.AssignColumn(fetcher.First[i].GetColID(), smartvectors.NewConstant(first[i], size))
		run.AssignColumn(fetcher.Last[i].GetColID(), smartvectors.NewConstant(last[i], size))
		run.AssignColumn(fetcher.FirstArith[i].GetColID(), smartvectors.NewConstant(first[i], size))
		run.AssignColumn(fetcher.LastArith[i].GetColID(), smartvectors.NewConstant(last[i], size))
		run.AssignColumn(fetcher.Data[i].GetColID(), smartvectors.RightZeroPadded(data[i], size))
	}

	for i := range firstBlockID {
		run.AssignColumn(fetcher.FirstBlockID[i].GetColID(), smartvectors.NewConstant(firstBlockID[i], size))
		run.AssignColumn(fetcher.FirstBlockIDArith[i].GetColID(), smartvectors.NewConstant(firstBlockID[i], size))
		run.AssignColumn(fetcher.LastBlockID[i].GetColID(), smartvectors.NewConstant(lastBlockID[i], size))
		run.AssignColumn(fetcher.LastBlockIDArith[i].GetColID(), smartvectors.NewConstant(lastBlockID[i], size))
	}

	// assign the SelectorTimestamp using the ComputeSelectorTimestamp prover action
	fetcher.ComputeSelectorTimestamp.Run(run)
	// assign the SelectorCt using the ComputeSelectorCt prover action
	fetcher.ComputeSelectorCt.Run(run)
	// assign the LastMinusFirstBlock using the LastMinusFirstBlockAction
	fetcher.LastMinusFirstBlockAction.Run(run)
}
