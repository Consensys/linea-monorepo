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
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/publicInput/utilities"
)

// RlpTxn models the arithmetization's RlpTxn module
type RlpTxn struct {
	AbsTxNum, AbsTxNumMax ifaces.Column // Absolute number of the transaction (starts from 1 and acts as an Active Filter), and the maximum number of transactions
	ToHashByProver        ifaces.Column // Relative TxNum inside the block,
	Limb                  ifaces.Column //
	NBytes                ifaces.Column //
}

type RlpTxnFetcher struct {
	AbsTxNum, AbsTxNumMax ifaces.Column // Absolute number of the transaction (starts from 1 and acts as an Active Filter), and the maximum number of transactions
	Limb                  ifaces.Column //
	NBytes                ifaces.Column
	FilterFetched         ifaces.Column
	EndOfRlpSegment       ifaces.Column

	SelectorDiffAbsTxId        ifaces.Column
	ComputeSelectorDiffAbsTxId wizard.ProverAction
}

func NewRlpTxnFetcher(comp *wizard.CompiledIOP, name string, rt *RlpTxn) RlpTxnFetcher {
	size := rt.Limb.Size()
	res := RlpTxnFetcher{
		AbsTxNum:        utilities.CreateCol(name, "ABS_TX_NUM", size, comp),
		AbsTxNumMax:     utilities.CreateCol(name, "ABS_TX_NUM_MAX", size, comp),
		Limb:            utilities.CreateCol(name, "LIMB", size, comp),
		NBytes:          utilities.CreateCol(name, "NBYTES", size, comp),
		FilterFetched:   utilities.CreateCol(name, "FILTER_FETCHED", size, comp),
		EndOfRlpSegment: utilities.CreateCol(name, "END_OF_RLP_SEGMENT", size, comp),
	}
	return res
}

func DefineRlpTxnFetcher(comp *wizard.CompiledIOP, fetcher *RlpTxnFetcher, name string, rlpTxnArith *RlpTxn) {
	fetcher.SelectorDiffAbsTxId, fetcher.ComputeSelectorDiffAbsTxId = dedicated.IsZero(
		comp,
		sym.Sub(
			fetcher.AbsTxNum,
			column.Shift(fetcher.AbsTxNum, 1),
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

	// require that the filter on fetched data only contains 1s followed by 0s
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

	utilities.MustBeBinary(comp, fetcher.EndOfRlpSegment)

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%s_GLOBAL_CONSTRAINT_ON_END_RLP_SEGMENT", name),
		sym.Mul(
			fetcher.FilterFetched, // constrain only on the active part of the module
			sym.Sub(
				fetcher.EndOfRlpSegment, // constrain EndOfRlpSegment
				sym.Sub(
					1,
					fetcher.SelectorDiffAbsTxId,
				),
			),
		),
	)

	// the table with the data we fetch from the arithmetization columns RlpTxn
	fetcherTable := []ifaces.Column{
		fetcher.AbsTxNum,
		fetcher.AbsTxNumMax,
		fetcher.Limb,
		fetcher.NBytes,
	}
	// the RlpTxn we extract timestamp data from, and which we will use to check for consistency
	arithTable := []ifaces.Column{
		rlpTxnArith.AbsTxNum,
		rlpTxnArith.AbsTxNumMax,
		rlpTxnArith.Limb,
		rlpTxnArith.NBytes,
	}

	// a projection query to check that the timestamp data is fetched correctly
	projection.InsertProjection(comp,
		ifaces.QueryIDf("%s_RLP_TXN_PROJECTION", name),
		fetcherTable,
		arithTable,
		fetcher.FilterFetched,
		rlpTxnArith.ToHashByProver, // filter lights up on the arithmetization's RlpTxn rows that contain rlp transaction data
	)
}

func AssignRlpTxnFetcher(run *wizard.ProverRuntime, fetcher *RlpTxnFetcher, rlpTxnArith *RlpTxn) {

	absTxNum := make([]field.Element, rlpTxnArith.Limb.Size())
	absTxNumMax := make([]field.Element, rlpTxnArith.Limb.Size())
	limb := make([]field.Element, rlpTxnArith.Limb.Size())
	nBytes := make([]field.Element, rlpTxnArith.Limb.Size())
	filterFetched := make([]field.Element, rlpTxnArith.Limb.Size())
	endOfRlpSegment := make([]field.Element, rlpTxnArith.Limb.Size())

	// counter is used to populate filter.Data and will increment every time we find a new timestamp
	counter := 0

	for i := 0; i < rlpTxnArith.Limb.Size(); i++ {
		toHashByProver := rlpTxnArith.ToHashByProver.GetColAssignmentAt(run, i)
		if toHashByProver.IsOne() {
			arithAbsTxNum := rlpTxnArith.AbsTxNum.GetColAssignmentAt(run, i)
			arithAbsTxNumMax := rlpTxnArith.AbsTxNumMax.GetColAssignmentAt(run, i)
			arithLimb := rlpTxnArith.Limb.GetColAssignmentAt(run, i)
			arithNBytes := rlpTxnArith.NBytes.GetColAssignmentAt(run, i)

			absTxNum[counter].Set(&arithAbsTxNum)
			absTxNumMax[counter].Set(&arithAbsTxNumMax)
			limb[counter].Set(&arithLimb)
			nBytes[counter].Set(&arithNBytes)

			filterFetched[counter].SetOne()
			counter++
		}
	}

	for i := 0; i < rlpTxnArith.Limb.Size()-1; i++ {
		if filterFetched[i].IsOne() {
			// only set end of segments in the active area
			if !absTxNum[i].Equal(&absTxNum[i+1]) {
				endOfRlpSegment[i].SetOne()
			}
		}
	}

	// assign the fetcher columns
	run.AssignColumn(fetcher.AbsTxNum.GetColID(), smartvectors.NewRegular(absTxNum))
	run.AssignColumn(fetcher.AbsTxNumMax.GetColID(), smartvectors.NewRegular(absTxNumMax))
	run.AssignColumn(fetcher.Limb.GetColID(), smartvectors.NewRegular(limb))
	run.AssignColumn(fetcher.NBytes.GetColID(), smartvectors.NewRegular(nBytes))
	run.AssignColumn(fetcher.FilterFetched.GetColID(), smartvectors.NewRegular(filterFetched))
	run.AssignColumn(fetcher.EndOfRlpSegment.GetColID(), smartvectors.NewRegular(endOfRlpSegment))

	fetcher.ComputeSelectorDiffAbsTxId.Run(run)
}
