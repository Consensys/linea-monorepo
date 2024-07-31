package logs

import (
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/accessors"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/publicInput/utilities"
)

// RollingSelector is used to fetch the last rolling hash and its associated message number
type RollingSelector struct {
	// Exists contains a 1 if there exists at least one rolling hash log
	Exists ifaces.Column
	// the Hi/Lo part of the last Rolling Hash found in the logs
	LastHi, LastLo ifaces.Column
	// the last message number of the last Rolling hash log
	LastMessageNo ifaces.Column
}

// NewRollingSelector returns a new RollingSelector with initialized columns that are not constrained.
func NewRollingSelector(comp *wizard.CompiledIOP, name string) RollingSelector {
	res := RollingSelector{
		Exists:        utilities.CreateCol(name, "EXISTS", 1, comp),
		LastHi:        utilities.CreateCol(name, "LAST_HI", 1, comp),
		LastLo:        utilities.CreateCol(name, "LAST_LO", 1, comp),
		LastMessageNo: utilities.CreateCol(name, "LAST_MESSAGE_NO", 1, comp),
	}
	return res
}

// DefineRollingSelector specifies the constraints of the RollingSelector with respect to the ExtractedData fetched from the arithmetization
func DefineRollingSelector(comp *wizard.CompiledIOP, sel RollingSelector, name string, fetchedHash, fetchedMsg ExtractedData) {
	// fetchedHash.filterFetched is an isActive pattern filter, and is constrained as such in DefineExtractedData
	isActive := fetchedHash.filterFetched

	// set the RollingSelector columns as public in order to get accessors
	comp.Columns.SetStatus(sel.Exists.GetColID(), column.Proof)
	comp.Columns.SetStatus(sel.LastHi.GetColID(), column.Proof)
	comp.Columns.SetStatus(sel.LastLo.GetColID(), column.Proof)
	comp.Columns.SetStatus(sel.LastMessageNo.GetColID(), column.Proof)

	// get accessors for columns of size 1
	accExists := accessors.NewFromPublicColumn(sel.Exists, 0)
	accessLastHi := accessors.NewFromPublicColumn(sel.LastHi, 0)           // to fetch the only field element in the column
	accessLastLo := accessors.NewFromPublicColumn(sel.LastLo, 0)           // to fetch the only field element in the column
	accessLastMsgNo := accessors.NewFromPublicColumn(sel.LastMessageNo, 0) // to fetch the only field element in the column

	// set the exists flag
	comp.InsertLocal(0, ifaces.QueryIDf("%s_%s", name, "EXISTS"),
		sym.Sub(
			isActive,
			accExists,
		),
	)

	// checkLastELemConsistency checks that the last element of the active part of parentCol is present in the field element of acc
	checkLastELemConsistency := func(parentCol ifaces.Column, acc ifaces.Accessor) {
		// active is already constrained in the fetcher, no need to constrain it again
		// two cases: Case 1: isActive is not completely filled with 1s, then parentCol[i] is equal to acc at the last row i where isActive[i] is 1
		comp.InsertGlobal(0, ifaces.QueryIDf("%s_%s_%s", name, "IS_ACTIVE_BORDER_CONSTRAINT", parentCol.GetColID()),
			sym.Mul(
				isActive,
				sym.Sub(1,
					column.Shift(isActive, 1),
				),
				sym.Sub(
					parentCol,
					acc,
				),
			),
		)

		// Case 2: isActive is completely filled with 1s, in which case we ask that isActive[size]*(parentCol[size]-acc) = 0
		// i.e. at the last row, parentCol contains the same element as acc
		comp.InsertLocal(0, ifaces.QueryIDf("%s_%s_%s", name, "IS_ACTIVE_FULL_CONSTRAINT", parentCol.GetColID()),
			sym.Mul(
				column.Shift(isActive, -1),
				sym.Sub(
					column.Shift(parentCol, -1),
					acc,
				),
			),
		)
	}
	// define the consistency constraints
	checkLastELemConsistency(fetchedHash.Hi, accessLastHi)
	checkLastELemConsistency(fetchedHash.Lo, accessLastLo)
	checkLastELemConsistency(fetchedMsg.Lo, accessLastMsgNo)
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

	// assign the RollingSelector columns
	run.AssignColumn(selector.Exists.GetColID(), smartvectors.NewRegular([]field.Element{exists}))
	run.AssignColumn(selector.LastHi.GetColID(), smartvectors.NewRegular([]field.Element{lastHi}))
	run.AssignColumn(selector.LastLo.GetColID(), smartvectors.NewRegular([]field.Element{lastLo}))
	run.AssignColumn(selector.LastMessageNo.GetColID(), smartvectors.NewRegular([]field.Element{lastMsg}))

}
