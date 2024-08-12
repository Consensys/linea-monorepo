package logs

import (
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/accessors"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/zkevm-monorepo/prover/symbolic"
	util "github.com/consensys/zkevm-monorepo/prover/zkevm/prover/publicInput/utilities"
)

// RollingSelector is used to fetch the last rolling hash and its associated message number
type RollingSelector struct {
	// Exists contains a 1 if there exists at least one rolling hash log
	Exists ifaces.Column
	// the Hi/Lo part of the first Rolling Hash found in the logs
	FirstHi, FirstLo ifaces.Column
	// the Hi/Lo part of the last Rolling Hash found in the logs
	LastHi, LastLo ifaces.Column
	// the first/last message number of the last Rolling hash log
	FirstMessageNo, LastMessageNo ifaces.Column
}

// NewRollingSelector returns a new RollingSelector with initialized columns that are not constrained.
func NewRollingSelector(comp *wizard.CompiledIOP, name string) RollingSelector {
	res := RollingSelector{
		Exists:         util.CreateCol(name, "EXISTS", 1, comp),
		FirstHi:        util.CreateCol(name, "FIRST_HI", 1, comp),
		FirstLo:        util.CreateCol(name, "FIRST_LO", 1, comp),
		LastHi:         util.CreateCol(name, "LAST_HI", 1, comp),
		LastLo:         util.CreateCol(name, "LAST_LO", 1, comp),
		FirstMessageNo: util.CreateCol(name, "FIRST_MESSAGE_NO", 1, comp),
		LastMessageNo:  util.CreateCol(name, "LAST_MESSAGE_NO", 1, comp),
	}
	return res
}

// DefineRollingSelector specifies the constraints of the RollingSelector with respect to the ExtractedData fetched from the arithmetization
func DefineRollingSelector(comp *wizard.CompiledIOP, sel RollingSelector, name string, fetchedHash, fetchedMsg ExtractedData) {
	// fetchedHash.filterFetched is an isActive pattern filter, and is constrained as such in DefineExtractedData
	isActive := fetchedHash.filterFetched

	// set the RollingSelector columns as public in order to get accessors
	var allCols = []ifaces.Column{sel.Exists, sel.FirstHi, sel.FirstLo, sel.LastHi, sel.LastLo, sel.FirstMessageNo, sel.LastMessageNo}
	for _, col := range allCols {
		comp.Columns.SetStatus(col.GetColID(), column.Proof)
	}

	// get accessors for columns of size 1
	accExists := accessors.NewFromPublicColumn(sel.Exists, 0)
	accessFirstHi := accessors.NewFromPublicColumn(sel.FirstHi, 0)           // to fetch the only field element in the column
	accessFirstLo := accessors.NewFromPublicColumn(sel.FirstLo, 0)           // to fetch the only field element in the column
	accessFirstMsgNo := accessors.NewFromPublicColumn(sel.FirstMessageNo, 0) // to fetch the only field element in the column
	accessLastHi := accessors.NewFromPublicColumn(sel.LastHi, 0)             // to fetch the only field element in the column
	accessLastLo := accessors.NewFromPublicColumn(sel.LastLo, 0)             // to fetch the only field element in the column
	accessLastMsgNo := accessors.NewFromPublicColumn(sel.LastMessageNo, 0)   // to fetch the only field element in the column

	// set the exists flag
	comp.InsertLocal(0, ifaces.QueryIDf("%s_%s", name, "EXISTS"),
		sym.Sub(
			isActive,
			accExists,
		),
	)
	// local openings for the first values
	comp.InsertLocal(0, ifaces.QueryIDf("%s_%s", name, "FIRST_HI"),
		sym.Sub(
			fetchedHash.Hi,
			accessFirstHi,
		),
	)
	comp.InsertLocal(0, ifaces.QueryIDf("%s_%s", name, "FIRST_LO"),
		sym.Sub(
			fetchedHash.Lo,
			accessFirstLo,
		),
	)
	comp.InsertLocal(0, ifaces.QueryIDf("%s_%s", name, "FIRST_MSG_NO"),
		sym.Sub(
			fetchedMsg.Lo,
			accessFirstMsgNo,
		),
	)

	// define the consistency constraints
	util.CheckLastELemConsistency(comp, isActive, fetchedHash.Hi, accessLastHi, name)
	util.CheckLastELemConsistency(comp, isActive, fetchedHash.Lo, accessLastLo, name)
	util.CheckLastELemConsistency(comp, isActive, fetchedMsg.Lo, accessLastMsgNo, name)
}

// AssignRollingSelector assigns the data in the RollingSelector using the ExtractedData fetched from the arithmetization
func AssignRollingSelector(run *wizard.ProverRuntime, selector RollingSelector, fetchedHash, fetchedMsg ExtractedData) {
	size := fetchedHash.Hi.Size()
	exists := run.GetColumnAt(fetchedHash.filterFetched.GetColID(), 0)
	var lastHi, lastLo, lastMsg field.Element
	for i := 0; i < size; i++ {
		isActive := run.GetColumnAt(fetchedHash.filterFetched.GetColID(), i)
		if isActive.IsOne() {
			lastHi = run.GetColumnAt(fetchedHash.Hi.GetColID(), i)
			lastLo = run.GetColumnAt(fetchedHash.Lo.GetColID(), i)
			lastMsg = run.GetColumnAt(fetchedMsg.Lo.GetColID(), i)
		} else {
			break
		}
	}
	// compute first values
	firstHi := run.GetColumnAt(fetchedHash.Hi.GetColID(), 0)
	firstLo := run.GetColumnAt(fetchedHash.Lo.GetColID(), 0)
	firstMsg := run.GetColumnAt(fetchedMsg.Lo.GetColID(), 0)

	// assign the RollingSelector columns
	run.AssignColumn(selector.Exists.GetColID(), smartvectors.NewRegular([]field.Element{exists}))
	run.AssignColumn(selector.FirstHi.GetColID(), smartvectors.NewRegular([]field.Element{firstHi}))
	run.AssignColumn(selector.FirstLo.GetColID(), smartvectors.NewRegular([]field.Element{firstLo}))
	run.AssignColumn(selector.FirstMessageNo.GetColID(), smartvectors.NewRegular([]field.Element{firstMsg}))
	run.AssignColumn(selector.LastHi.GetColID(), smartvectors.NewRegular([]field.Element{lastHi}))
	run.AssignColumn(selector.LastLo.GetColID(), smartvectors.NewRegular([]field.Element{lastLo}))
	run.AssignColumn(selector.LastMessageNo.GetColID(), smartvectors.NewRegular([]field.Element{lastMsg}))

}
