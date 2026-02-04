package execution_data_collector

import (
	"fmt"
	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	util "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/utilities"
)

// GenericPadderPacker is used to MiMC-hash the data in LogMessages. Using a zero initial stata,
// the data in L2L1 logs must be hashed as follows: msg1Hash, msg2Hash, and so on.
// The final value of the chained hash can be retrieved as hash[ctMax[any index]]
type PadderPacker struct {
	InputLimbs [common.NbLimbU128]ifaces.Column
	// The number of bytes in the limbs.
	InputNoBytes          ifaces.Column
	InputIsActive         ifaces.Column
	OneColumn             ifaces.Column
	OneColumnFilter       ifaces.Column
	OneColumnBytes        ifaces.Column
	OneColumnBytesSum     ifaces.Column
	PeriodicFilter        [8]ifaces.Column
	TrimmedPeriodicFilter [8]ifaces.Column

	OneColumnWithoutGaps    ifaces.Column
	FilterWithoutGaps       ifaces.Column
	FilterWithoutGapsPadded ifaces.Column
	OuterColumns            [8]ifaces.Column
	OuterIsActive           ifaces.Column
	CounterColumn           ifaces.Column
	CounterColumnFilled     ifaces.Column
	CounterColumnAdded      ifaces.Column

	CounterColumnAddedIndicator ifaces.Column
}

// NewPoseidonPadderPacker returns a new GenericPadderPacker with initialized columns that are not constrained.
func NewPadderPacker(comp *wizard.CompiledIOP, inputLimbs [common.NbLimbU128]ifaces.Column, inputNoBytes, inputIsActive ifaces.Column, name string) PadderPacker {
	var (
		res     PadderPacker
		newSize int
	)
	res.InputLimbs = inputLimbs
	res.InputNoBytes = inputNoBytes
	res.InputIsActive = inputIsActive

	oldSize := res.InputLimbs[0].Size()
	newSize = oldSize * common.NbLimbU128
	res.OneColumn = util.CreateCol(name, "ONE_COLUMN", newSize, comp)
	res.OneColumnFilter = util.CreateCol(name, "ONE_COLUMN_FILTER", newSize, comp)
	res.OneColumnBytes = util.CreateCol(name, "ONE_COLUMN_BYTES", newSize, comp)
	res.OneColumnBytesSum = util.CreateCol(name, "ONE_COLUMN_BYTE_SUM", newSize, comp)
	res.OneColumnWithoutGaps = util.CreateCol(name, "ONE_COLUMN_WITHOUT_GAPS", newSize, comp)
	res.FilterWithoutGaps = util.CreateCol(name, "FILTER_WITHOUT_GAPS", newSize, comp)
	res.FilterWithoutGapsPadded = util.CreateCol(name, "FILTER_WITHOUT_GAPS_PADDED", newSize, comp)
	res.CounterColumn = util.CreateCol(name, "COUNTER_COLUMN", newSize, comp)
	res.CounterColumnFilled = util.CreateCol(name, "COUNTER_COLUMN_FILLED", newSize, comp)
	res.CounterColumnAdded = util.CreateCol(name, "COUNTER_COLUMN_ADDED", newSize, comp)

	res.CounterColumnAddedIndicator = util.CreateCol(name, "COUNTER_COLUMN_ADDED_INDICATOR", newSize, comp)

	for i := range res.PeriodicFilter {
		res.PeriodicFilter[i] = util.CreateCol(name, fmt.Sprintf("PERIODIC_FILTER_%d", i), newSize, comp)
		res.TrimmedPeriodicFilter[i] = util.CreateCol(name, fmt.Sprintf("TRIMMED_PERIODIC_FILTER_%d", i), newSize, comp)
	}

	for i := range res.OuterColumns {
		res.OuterColumns[i] = util.CreateCol(name, fmt.Sprintf("INTER_COLUMN_%d", i), oldSize, comp)
	}
	res.OuterIsActive = util.CreateCol(name, "OUTPUT_IS_ACTIVE", oldSize, comp)
	return res
}

// DefineHasher specifies the constraints of the GenericPadderPacker with respect to the ExtractedData fetched from the arithmetization
func DefinePadderPacker(comp *wizard.CompiledIOP, ppp PadderPacker, name string) {

	/*
		util.MustBeBinary(comp, ppp.OuterIsActive)
		comp.InsertGlobal(0,
			ifaces.QueryIDf("%s_INTER_IS_ACTIVE_SHAPE", name),
			sym.Sub(
				ppp.OuterIsActive,
				sym.Mul(
					column.Shift(ppp.OuterIsActive, -1),
					ppp.OuterIsActive,
				),
			),
		)
		// 3. OutputData is zero when OutputIsActive is zero.
		for j := range ppp.OuterColumns {
			comp.InsertGlobal(0,
				ifaces.QueryIDf("%s_INTER_DATA_ZERO_ON_INACTIVE_%d", name, j),
				sym.Sub(
					ppp.OuterColumns[j],
					sym.Mul(ppp.OuterColumns[j], ppp.OuterIsActive),
				),
			)
		}

	*/
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_PERIODIC_PATTERN", name),
		sym.Mul(
			ppp.PeriodicFilter[0],
			sym.Sub(
				ppp.PeriodicFilter[0],
				column.Shift(ppp.PeriodicFilter[0], -8),
			),
		),
	)

	comp.InsertLocal(0,
		ifaces.QueryIDf("%s_PERIODIC_PATTERN_INIT_ONE", name),
		sym.Sub(
			ppp.PeriodicFilter[0],
			1),
	)
	for i := 1; i < 8; i++ {
		comp.InsertLocal(0,
			ifaces.QueryIDf("%s_PERIODIC_PATTERN_INIT_ZERO_%d", name, i),
			ifaces.ColumnAsVariable(column.Shift(ppp.PeriodicFilter[0], i)),
		)
	}

	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_ONE_COLUMN_NBYTES_SUM", name),
		sym.Mul(
			ppp.PeriodicFilter[0],
			sym.Sub(
				ppp.OneColumnBytesSum,
				ppp.OneColumnBytes,
				column.Shift(ppp.OneColumnBytes, 1),
				column.Shift(ppp.OneColumnBytes, 2),
				column.Shift(ppp.OneColumnBytes, 3),
				column.Shift(ppp.OneColumnBytes, 4),
				column.Shift(ppp.OneColumnBytes, 5),
				column.Shift(ppp.OneColumnBytes, 6),
				column.Shift(ppp.OneColumnBytes, 7),
			),
		),
	)

	for j := 0; j < 8-1; j++ {
		// for all the periodic filters except the last one
		// the last one being excluded ensures that the blocks get reset
		comp.InsertGlobal(0,
			ifaces.QueryIDf("%s_NBYTES_ZERO_PROPAGATION_%d", name, j),
			sym.Mul(
				ppp.PeriodicFilter[j],
				sym.Mul( // (2-OneColumnBytes[i])*OneColumnBytes(i+1)=0
					sym.Sub(
						field.NewElement(2),
						ppp.OneColumnBytes,
					),
					column.Shift(ppp.OneColumnBytes, 1),
				),
			),
		)
	}

	comp.InsertProjection(
		ifaces.QueryIDf("%s_PADDER_PACKER_ONE_COLUMN_PROJECTION_BYTES_SUM", name),
		query.ProjectionInput{
			// the table with the data we fetch from the arithmetization's TxnData columns
			ColumnA: []ifaces.Column{ppp.InputNoBytes},
			// the TxnData we extract sender addresses from, and which we will use to check for consistency
			ColumnB: []ifaces.Column{ppp.OneColumnBytesSum},
			FilterA: ppp.InputIsActive,
			// filter lights up on the arithmetization's TxnData rows that contain sender address data
			FilterB: ppp.TrimmedPeriodicFilter[0]})

	comp.InsertProjection(
		ifaces.QueryIDf("MULTIARY_PROJECTION_STEP_1"),
		query.ProjectionMultiAryInput{
			ColumnsA: [][]ifaces.Column{
				[]ifaces.Column{ppp.InputLimbs[0]},
				[]ifaces.Column{ppp.InputLimbs[1]},
				[]ifaces.Column{ppp.InputLimbs[2]},
				[]ifaces.Column{ppp.InputLimbs[3]},
				[]ifaces.Column{ppp.InputLimbs[4]},
				[]ifaces.Column{ppp.InputLimbs[5]},
				[]ifaces.Column{ppp.InputLimbs[6]},
				[]ifaces.Column{ppp.InputLimbs[7]},
			},
			ColumnsB: [][]ifaces.Column{
				[]ifaces.Column{ppp.OneColumn},
			},
			FiltersA: []ifaces.Column{
				ppp.InputIsActive,
				ppp.InputIsActive,
				ppp.InputIsActive,
				ppp.InputIsActive,
				ppp.InputIsActive,
				ppp.InputIsActive,
				ppp.InputIsActive,
				ppp.InputIsActive,
			},
			FiltersB: []ifaces.Column{
				ppp.OneColumnFilter,
			},
		},
	)

	for j := range ppp.InputLimbs {
		comp.InsertProjection(
			ifaces.QueryIDf("%s_PADDER_PACKER_ONE_COLUMN_PROJECTION_%d", name, j),
			query.ProjectionInput{
				// the table with the data we fetch from the arithmetization's TxnData columns
				ColumnA: []ifaces.Column{ppp.InputLimbs[j]},
				// the TxnData we extract sender addresses from, and which we will use to check for consistency
				ColumnB: []ifaces.Column{ppp.OneColumn},
				FilterA: ppp.InputIsActive,
				// filter lights up on the arithmetization's TxnData rows that contain sender address data
				FilterB: ppp.TrimmedPeriodicFilter[j]})
	}

	comp.InsertProjection(
		ifaces.QueryIDf("%s_PADDER_PACKER_ONE_COLUMN_NO_GAPS_PROJECTION", name),
		query.ProjectionInput{
			// the table with the data we fetch from the arithmetization's TxnData columns
			ColumnA: []ifaces.Column{ppp.OneColumn},
			// the TxnData we extract sender addresses from, and which we will use to check for consistency
			ColumnB: []ifaces.Column{ppp.OneColumnWithoutGaps},
			FilterA: ppp.OneColumnBytes,
			// filter lights up on the arithmetization's TxnData rows that contain sender address data
			FilterB: ppp.FilterWithoutGaps})

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%s_IS_ACTIVE_CONSTRAINT_NO_0_TO_1_FILTER_WITHOUT_GAPS", name),
		sym.Sub(
			ppp.FilterWithoutGaps,
			sym.Mul(
				column.Shift(ppp.FilterWithoutGaps, -1),
				ppp.FilterWithoutGaps,
			),
		),
	)

	comp.InsertProjection(
		ifaces.QueryIDf("MULTIARY_PROJECTION_SIMPLE"),
		query.ProjectionMultiAryInput{
			ColumnsA: [][]ifaces.Column{
				[]ifaces.Column{ppp.OneColumnWithoutGaps},
			},
			ColumnsB: [][]ifaces.Column{
				//[]ifaces.Column{ppp.OuterColumns[0], ppp.OuterColumns[1], ppp.OuterColumns[2], ppp.OuterColumns[3], ppp.OuterColumns[4], ppp.OuterColumns[5], ppp.OuterColumns[6], ppp.OuterColumns[7]},
				[]ifaces.Column{ppp.OuterColumns[0]},
				[]ifaces.Column{ppp.OuterColumns[1]},
				[]ifaces.Column{ppp.OuterColumns[2]},
				[]ifaces.Column{ppp.OuterColumns[3]},
				[]ifaces.Column{ppp.OuterColumns[4]},
				[]ifaces.Column{ppp.OuterColumns[5]},
				[]ifaces.Column{ppp.OuterColumns[6]},
				[]ifaces.Column{ppp.OuterColumns[7]},
			},
			FiltersA: []ifaces.Column{
				verifiercol.NewConstantCol(field.One(), ppp.OneColumnWithoutGaps.Size(), string(ifaces.ColIDf("PADDER_PACKER_FULL_ONES_FILTER_%s", name))),
			},
			FiltersB: []ifaces.Column{
				verifiercol.NewConstantCol(field.One(), ppp.OuterColumns[0].Size(), string(ifaces.ColIDf("PADDER_PACKER_FULL_ONES_FILTER_%s_%d", name, 0))),
				verifiercol.NewConstantCol(field.One(), ppp.OuterColumns[0].Size(), string(ifaces.ColIDf("PADDER_PACKER_FULL_ONES_FILTER_%s_%d", name, 1))),
				verifiercol.NewConstantCol(field.One(), ppp.OuterColumns[0].Size(), string(ifaces.ColIDf("PADDER_PACKER_FULL_ONES_FILTER_%s_%d", name, 2))),
				verifiercol.NewConstantCol(field.One(), ppp.OuterColumns[0].Size(), string(ifaces.ColIDf("PADDER_PACKER_FULL_ONES_FILTER_%s_%d", name, 3))),
				verifiercol.NewConstantCol(field.One(), ppp.OuterColumns[0].Size(), string(ifaces.ColIDf("PADDER_PACKER_FULL_ONES_FILTER_%s_%d", name, 4))),
				verifiercol.NewConstantCol(field.One(), ppp.OuterColumns[0].Size(), string(ifaces.ColIDf("PADDER_PACKER_FULL_ONES_FILTER_%s_%d", name, 5))),
				verifiercol.NewConstantCol(field.One(), ppp.OuterColumns[0].Size(), string(ifaces.ColIDf("PADDER_PACKER_FULL_ONES_FILTER_%s_%d", name, 6))),
				verifiercol.NewConstantCol(field.One(), ppp.OuterColumns[0].Size(), string(ifaces.ColIDf("PADDER_PACKER_FULL_ONES_FILTER_%s_%d", name, 7))),
			},
		},
	)

	comp.InsertProjection(
		ifaces.QueryIDf("MULTIARY_PROJECTION_SIMPLE_TRY_2"),
		query.ProjectionMultiAryInput{
			ColumnsA: [][]ifaces.Column{
				[]ifaces.Column{ppp.OneColumnWithoutGaps},
			},
			ColumnsB: [][]ifaces.Column{
				//[]ifaces.Column{ppp.OuterColumns[0], ppp.OuterColumns[1], ppp.OuterColumns[2], ppp.OuterColumns[3], ppp.OuterColumns[4], ppp.OuterColumns[5], ppp.OuterColumns[6], ppp.OuterColumns[7]},
				[]ifaces.Column{ppp.OuterColumns[0]},
				[]ifaces.Column{ppp.OuterColumns[1]},
				[]ifaces.Column{ppp.OuterColumns[2]},
				[]ifaces.Column{ppp.OuterColumns[3]},
				[]ifaces.Column{ppp.OuterColumns[4]},
				[]ifaces.Column{ppp.OuterColumns[5]},
				[]ifaces.Column{ppp.OuterColumns[6]},
				[]ifaces.Column{ppp.OuterColumns[7]},
			},
			FiltersA: []ifaces.Column{
				ppp.CounterColumnAdded,
			},
			FiltersB: []ifaces.Column{
				ppp.OuterIsActive,
				ppp.OuterIsActive,
				ppp.OuterIsActive,
				ppp.OuterIsActive,
				ppp.OuterIsActive,
				ppp.OuterIsActive,
				ppp.OuterIsActive,
				ppp.OuterIsActive,
			},
		},
	)

	// Constrain CounterColumn
	for i := 0; i < 8; i++ {
		comp.InsertGlobal(0,
			ifaces.QueryIDf("%s_COUNTER_VALUE_%d", name, i),
			sym.Mul(
				ppp.FilterWithoutGaps,
				ppp.PeriodicFilter[i],
				sym.Sub(ppp.CounterColumn, field.NewElement(uint64(i))),
			),
		)
	}

	// Constraint: After last active row, counter continues until reaching 0 (multiple of 8)
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_COUNTER_PADDING_BORDER", name),
		sym.Mul(
			sym.Sub(field.NewElement(1), column.Shift(ppp.FilterWithoutGaps, 1)),
			ppp.FilterWithoutGaps, // at the border of FilterWithoutGaps
			ppp.CounterColumn,     // must be > 0
			sym.Sub(
				// CounterColumnFilled increases by 1 compared to CounterColimn
				// at the border
				column.Shift(ppp.CounterColumnFilled, 1),
				ppp.CounterColumn,
				1,
			),
		),
	)

	comp.InsertGlobal(0,
		ifaces.QueryIDf("%s_COUNTER_PADDING_FILLING_CORRECTNESS", name),
		sym.Mul(
			ppp.CounterColumnFilled,           // CounterColumnFilled > 0
			sym.Sub(1, ppp.PeriodicFilter[7]), // we are not at the end of the size 8 block
			sym.Sub(
				// CounterColumnFilled increases by 1 compared to CounterColimn
				// at the border
				column.Shift(ppp.CounterColumnFilled, 1),
				ppp.CounterColumnFilled,
				1,
			),
		),
	)
}

// AssignHasher assigns the data in the GenericPadderPacker using the ExtractedData fetched from the arithmetization
func AssignPadderPacker(run *wizard.ProverRuntime, ppp PadderPacker) {
	oneColumn := make([]field.Element, ppp.OneColumn.Size())
	oneColumnWithoutGaps := make([]field.Element, ppp.OneColumnWithoutGaps.Size())
	filterWithoutGaps := make([]field.Element, ppp.FilterWithoutGaps.Size())
	filterWithoutGapsPadded := make([]field.Element, ppp.FilterWithoutGaps.Size())
	oneColumnFilter := make([]field.Element, ppp.OneColumn.Size())
	periodicFilter := make([][]field.Element, len(ppp.PeriodicFilter))
	oneColumnBytes := make([]field.Element, ppp.OneColumn.Size())
	oneColumnBytesSum := make([]field.Element, ppp.OneColumn.Size())
	counterColumn := make([]field.Element, ppp.CounterColumn.Size())
	counterColumnFilled := make([]field.Element, ppp.CounterColumnFilled.Size())
	for j := range ppp.PeriodicFilter {
		periodicFilter[j] = make([]field.Element, ppp.PeriodicFilter[j].Size())
	}
	outer := make([][]field.Element, len(ppp.OuterColumns))
	for j := range ppp.OuterColumns {
		outer[j] = make([]field.Element, ppp.OuterColumns[j].Size())
	}
	outerIsActive := make([]field.Element, ppp.OuterIsActive.Size())

	counterRow := 0
	for i := 0; i < ppp.InputLimbs[0].Size(); i++ {
		for j := 0; j < common.NbLimbU128; j++ {
			periodicFilter[j][counterRow].SetOne()
			counterRow++
		}
	}

	counterRow = 0

	for i := 0; i < ppp.InputLimbs[0].Size(); i++ {
		isActive := ppp.InputIsActive.GetColAssignmentAt(run, i)
		nBytesLimb := ppp.InputNoBytes.GetColAssignmentAt(run, i)
		nBytesLimbInt := int(nBytesLimb.Uint64())
		remainingNBytes := nBytesLimbInt
		if isActive.IsOne() {
			for j := 0; j < common.NbLimbU128; j++ {
				if j == 0 {
					oneColumnBytesSum[counterRow].SetUint64(uint64(nBytesLimbInt))
				}
				// periodicFilter[j][counterRow].SetOne()
				limbValue := ppp.InputLimbs[j].GetColAssignmentAt(run, i)
				limbBytes := limbValue.Bytes()
				oneColumnFilter[counterRow].SetOne()
				if remainingNBytes > 0 {
					oneColumnBytes[counterRow].SetUint64(2)
					oneColumn[counterRow].SetBytes(limbBytes[2:4])
					// oneColumnFilter[counterRow].SetOne()
					remainingNBytes -= 2
				}
				counterRow++
			}
		}
	}

	counterRow = 0
	for i := 0; i < len(oneColumn); i++ {
		if !oneColumnBytes[i].IsZero() {
			oneColumnWithoutGaps[counterRow].Set(&oneColumn[i])
			filterWithoutGaps[counterRow].SetOne()
			counterRow++
		}
	}

	lastRow := 0
	for i := 0; i < len(oneColumnWithoutGaps); i++ {
		if filterWithoutGaps[i].IsOne() {
			outer[i%common.NbLimbU128][i/common.NbLimbU128].Set(&oneColumnWithoutGaps[i])
			// outerIsActive[i%common.NbLimbU128][i/common.NbLimbU128].SetOne()
			outerIsActive[i/common.NbLimbU128].SetOne()
			counterColumn[i].SetUint64(uint64(i % common.NbLimbU128))
			// counterColumnFilled[i].SetUint64(uint64(i % common.NbLimbU128))
			lastRow = i
		}
	}
	if lastRow%common.NbLimbU128 != 0 {
		for i := lastRow + 1; i%common.NbLimbU128 != 0; i++ {
			counterColumnFilled[i].SetUint64(uint64(i % common.NbLimbU128))
		}
	}

	counterColumnAddedIndicator := make([]field.Element, ppp.CounterColumnAddedIndicator.Size())
	for i := 0; i < ppp.CounterColumnAddedIndicator.Size(); i++ {
		if counterColumn[i].Uint64()+counterColumnFilled[i].Uint64() > 0 {
			counterColumnAddedIndicator[i].SetOne()
		}
	}

	run.AssignColumn(ppp.OneColumn.GetColID(), sv.NewRegular(oneColumn))
	run.AssignColumn(ppp.OneColumnFilter.GetColID(), sv.NewRegular(oneColumnFilter))
	run.AssignColumn(ppp.OneColumnWithoutGaps.GetColID(), sv.NewRegular(oneColumnWithoutGaps))
	run.AssignColumn(ppp.FilterWithoutGaps.GetColID(), sv.NewRegular(filterWithoutGaps))
	run.AssignColumn(ppp.FilterWithoutGapsPadded.GetColID(), sv.NewRegular(filterWithoutGapsPadded))
	run.AssignColumn(ppp.OneColumnBytes.GetColID(), sv.NewRegular(oneColumnBytes))
	run.AssignColumn(ppp.OneColumnBytesSum.GetColID(), sv.NewRegular(oneColumnBytesSum))
	run.AssignColumn(ppp.CounterColumn.GetColID(), sv.NewRegular(counterColumn))
	run.AssignColumn(ppp.CounterColumnFilled.GetColID(), sv.NewRegular(counterColumnFilled))
	run.AssignColumn(ppp.CounterColumnAdded.GetColID(), sv.Add(sv.NewRegular(filterWithoutGaps), sv.NewRegular(counterColumnFilled)))
	run.AssignColumn(ppp.CounterColumnAddedIndicator.GetColID(), sv.NewRegular(counterColumnAddedIndicator))
	for j := range ppp.PeriodicFilter {
		run.AssignColumn(ppp.PeriodicFilter[j].GetColID(), sv.NewRegular(periodicFilter[j]))
		run.AssignColumn(ppp.TrimmedPeriodicFilter[j].GetColID(), sv.Mul(sv.NewRegular(periodicFilter[j]), sv.NewRegular(oneColumnFilter)))
	}
	for j := range ppp.OuterColumns {
		run.AssignColumn(ppp.OuterColumns[j].GetColID(), sv.NewRegular(outer[j]))
	}
	run.AssignColumn(ppp.OuterIsActive.GetColID(), sv.NewRegular(outerIsActive))

	for i := 0; i < ppp.OuterColumns[0].Size(); i++ {
		for j := range ppp.OuterColumns {
			fetchedValue := ppp.OuterColumns[j].GetColAssignmentAt(run, i)
			bytes := fetchedValue.Bytes()
			fmt.Println(utils.HexEncodeToString(bytes[:]))
		}
	}

}
