package fetchers_arithmetization

import (
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/dedicated"
	"github.com/consensys/zkevm-monorepo/prover/protocol/dedicated/projection"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/zkevm-monorepo/prover/symbolic"
	arith "github.com/consensys/zkevm-monorepo/prover/zkevm/prover/publicInput/arith_struct"
	util "github.com/consensys/zkevm-monorepo/prover/zkevm/prover/publicInput/utilities"
)

type BlockTxnMetadata struct {
	BlockID         ifaces.Column
	TotalNoTxnBlock ifaces.Column
	FilterFetched   ifaces.Column
	FilterArith     ifaces.Column
	SelectorCt      ifaces.Column
	// prover action to compute SelectorCt
	ComputeSelectorCt wizard.ProverAction

	FirstAbsTxId ifaces.Column
	LastAbsTxId  ifaces.Column
}

func NewBlockTxnMetadata(comp *wizard.CompiledIOP, name string, td *arith.TxnData) BlockTxnMetadata {
	res := BlockTxnMetadata{
		BlockID:         util.CreateCol(name, "BLOCK_ID", td.Ct.Size(), comp),
		TotalNoTxnBlock: util.CreateCol(name, "TOTAL_NO_TX_BLOCK", td.Ct.Size(), comp),
		FilterFetched:   util.CreateCol(name, "FILTER_FETCHED", td.Ct.Size(), comp),
		FilterArith:     util.CreateCol(name, "FILTER_ARITH", td.Ct.Size(), comp),
		FirstAbsTxId:    util.CreateCol(name, "FIRST_ABS_TX_ID", td.Ct.Size(), comp),
		LastAbsTxId:     util.CreateCol(name, "LAST_ABS_TX_ID", td.Ct.Size(), comp),
	}
	return res
}

func DefineBlockTxnMetaData(comp *wizard.CompiledIOP, btm *BlockTxnMetadata, name string, td *arith.TxnData) {
	btm.SelectorCt, btm.ComputeSelectorCt = dedicated.IsZero(
		comp,
		td.Ct, // select the columns where td.Ct = 0
	)

	// constrain the arithmetization filter
	comp.InsertGlobal(0, ifaces.QueryIDf("%s_%s", name, "FILTER_ARITH"),
		sym.Sub(
			btm.FilterArith,
			sym.Mul(
				td.IsLastTxOfBlock,
				btm.SelectorCt,
			),
		),
	)
	// constrain the fetched filter
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%s_FILTER_ON_FETCHED_CONSTRAINT_MUST_BE_BINARY", name),
		sym.Mul(
			btm.FilterFetched,
			sym.Sub(btm.FilterFetched, 1),
		),
	)

	// require that the filter on fetched data only contains 1s followed by 0s
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%s_FILTER_ON_FETCHED_CONSTRAINT_NO_0_TO_1", name),
		sym.Sub(
			btm.FilterFetched,
			sym.Mul(
				column.Shift(btm.FilterFetched, -1),
				btm.FilterFetched),
		),
	)

	// constrain the FirstAbsTxId
	comp.InsertLocal(
		0,
		ifaces.QueryIDf("%s_FIRST_ABS_TX_ID_LOCAL_CONSTRAINT", name),
		sym.Sub(
			btm.FirstAbsTxId,
			1,
		),
	)

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%s_FIRST_ABS_TX_ID_GLOBAL_CONSTRAINT", name),
		sym.Mul(
			btm.FilterFetched, // filter on the active part
			sym.Sub(
				btm.FirstAbsTxId,
				sym.Add(
					column.Shift(btm.LastAbsTxId, -1),
					1,
				),
			),
		),
	)

	// constrain the LastAbsTxId
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%s_LAST_ABS_TX_ID_GLOBAL_CONSTRAINT", name),
		sym.Mul(
			btm.FilterFetched, // active filter
			sym.Sub(
				btm.LastAbsTxId,
				btm.FirstAbsTxId,
				sym.Sub(
					btm.TotalNoTxnBlock,
					1,
				),
			),
		),
	)

	// the table with the data we fetch from the arithmetization's TxnData columns
	fetcherTable := []ifaces.Column{
		btm.BlockID,
		btm.TotalNoTxnBlock,
	}
	// the TxnData we extract sender addresses from, and which we will use to check for consistency
	arithTable := []ifaces.Column{
		td.RelBlock,
		td.RelTxNumMax,
	}

	// a projection query to check that the sender addresses are fetched correctly
	projection.InsertProjection(comp,
		ifaces.QueryIDf("%s_PROJECTION", name),
		fetcherTable,
		arithTable,
		btm.FilterFetched,
		btm.FilterArith,
	)
}

func AssignBlockTxnMetadata(run *wizard.ProverRuntime, btm BlockTxnMetadata, td *arith.TxnData) {
	blockId := make([]field.Element, td.Ct.Size())
	totalNoTxnBlock := make([]field.Element, td.Ct.Size())
	filterFetched := make([]field.Element, td.Ct.Size())
	filterArith := make([]field.Element, td.Ct.Size())
	firstAbsTxId := make([]field.Element, td.Ct.Size())
	lastAbsTxId := make([]field.Element, td.Ct.Size())
	lastRelBlock := field.Zero()
	counter := 0
	var ctAbsTxNum int64 = 1
	for i := 0; i < td.Ct.Size(); i++ {
		relBlock := td.RelBlock.GetColAssignmentAt(run, i)
		if !relBlock.IsZero() && !relBlock.Equal(&lastRelBlock) {
			// relBlock starts from 1, and is 0 in the padding area
			// if relBlock is different from lastRelBlock, it means we need to register the metadata for this block
			lastRelBlock = relBlock
			blockId[counter].Set(&relBlock)
			fetchTotalNoTxnBlock := td.RelTxNumMax.GetColAssignmentAt(run, i)
			totalNoTxnBlock[counter].Set(&fetchTotalNoTxnBlock)
			filterFetched[counter].SetOne()

			// set the absolute IDs, firstAbsTxId and lastAbsTxId for the block
			firstAbsTxId[counter].SetInt64(ctAbsTxNum)
			lastAbsTxId[counter].Set(&firstAbsTxId[counter])
			integerNoOfTxBlock := int64(fetchTotalNoTxnBlock.Uint64())
			lastAbsTxId[counter].SetInt64(ctAbsTxNum + integerNoOfTxBlock - 1)
			// increas ctAbsTxNum counter
			ctAbsTxNum += integerNoOfTxBlock
			// set the counter
			counter++
		}
		lastTxBlock := td.IsLastTxOfBlock.GetColAssignmentAt(run, i)
		ct := td.Ct.GetColAssignmentAt(run, i)
		if lastTxBlock.IsOne() && ct.IsZero() {
			filterArith[i].SetOne()
		}
	}
	run.AssignColumn(btm.BlockID.GetColID(), smartvectors.NewRegular(blockId))
	run.AssignColumn(btm.TotalNoTxnBlock.GetColID(), smartvectors.NewRegular(totalNoTxnBlock))
	run.AssignColumn(btm.FilterFetched.GetColID(), smartvectors.NewRegular(filterFetched))
	run.AssignColumn(btm.FilterArith.GetColID(), smartvectors.NewRegular(filterArith))
	run.AssignColumn(btm.FirstAbsTxId.GetColID(), smartvectors.NewRegular(firstAbsTxId))
	run.AssignColumn(btm.LastAbsTxId.GetColID(), smartvectors.NewRegular(lastAbsTxId))

	btm.ComputeSelectorCt.Run(run)
}
