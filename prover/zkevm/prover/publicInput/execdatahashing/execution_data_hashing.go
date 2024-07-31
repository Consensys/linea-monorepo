package execdatahashing

import (
	"fmt"
	"strings"

	"github.com/consensys/zkevm-monorepo/prover/crypto/mimc"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/query"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/zkevm-monorepo/prover/symbolic"
)

// ExecDataHashingInput collect the inputs of the presently-specified [execDataHashing]
// module.
type ExecDataHashingInput struct {
	// Data is a column storing the execution data bytes in the following format
	// 	- Each field element stores 31 bytes, this is a requirement for the MiMC
	// 		hash function.
	// 	- The input is prepadded.
	Data ifaces.Column
	// Selector is a binary indicator column indicating whether the current row
	// of Data corresponds actually to a 31 bytes limb. The rows marked with a
	// zero are ignored by the present hashing module.
	Selector ifaces.Column
}

// execData stores the prover context for hashing the execution data. Its role
// is to coordinate the MiMC module. The structure is also responsible for
// assigned its internal columns, it implements the [wizard.ProverAction]
// interface.
//
// The structure is to be constructed by the caller of the [HashExecutionData]
// function and all provided columns must have the same length.
type execData struct {
	// Inputs stores the [ExecDataHashingInput] used to generate the module
	Inputs *ExecDataHashingInput
	// InitState stores the initial state of the compression function of MiMC
	InitState ifaces.Column
	// NextState stores the result of the compression function
	NextState ifaces.Column
	// FinalHash stores the latest value of the digest and can be local opened
	// on the last position to obtain de final value.
	FinalHash ifaces.Column

	// FinalOpening opens the last position of FinalHash and returns the
	// resulting value of the digest.
	FinalHashOpening query.LocalOpening
}

// HashExecutionData generates the execution data hashing components: the relevant
// constraints and returns a [wizard.ProverAction] object that can be used to
// assign the internal columns.
func HashExecutionData(comp *wizard.CompiledIOP, edhi *ExecDataHashingInput) (ed *execData) {

	// size denotes the size of the internal columns
	var size = edhi.Data.Size()

	ed = &execData{
		Inputs:    edhi,
		InitState: comp.InsertCommit(0, deriveName[ifaces.ColID]("INIT_STATE"), size),
		NextState: comp.InsertCommit(0, deriveName[ifaces.ColID]("NEW_STATE"), size),
		FinalHash: comp.InsertCommit(0, deriveName[ifaces.ColID]("FINAL_HASH"), size),
	}

	// When the selector is zero, then the newState and the oldState are the
	// same as the previous row.
	comp.InsertGlobal(0,
		deriveName[ifaces.QueryID]("OLD_STATE_ON_UNSELECTED"),
		sym.Sub(
			ed.InitState,
			sym.Mul(
				sym.Sub(1, column.Shift(ed.Inputs.Selector, -1)),
				column.Shift(ed.InitState, -1),
			),
			sym.Mul(
				column.Shift(ed.Inputs.Selector, -1),
				column.Shift(ed.NextState, -1),
			),
		),
	)

	comp.InsertGlobal(0,
		deriveName[ifaces.QueryID]("FINAL_HASH_CORRECTNESS"),
		sym.Sub(
			ed.FinalHash,
			sym.Mul(
				sym.Sub(1, ed.Inputs.Selector),
				column.Shift(ed.FinalHash, -1),
			),
			sym.Mul(
				ed.Inputs.Selector,
				ed.NextState,
			),
		),
	)

	comp.InsertMiMC(0,
		deriveName[ifaces.QueryID]("MIMC_COMPRESSION"),
		ed.Inputs.Data,
		ed.InitState,
		ed.NextState,
	)

	ed.FinalHashOpening = comp.InsertLocalOpening(0,
		deriveName[ifaces.QueryID]("FINAL_OPENING"),
		column.Shift(ed.FinalHash, -1),
	)

	comp.InsertLocal(0,
		deriveName[ifaces.QueryID]("INITIAL_INIT_STATE"),
		sym.NewVariable(ed.InitState),
	)

	return ed
}

// Run implements the [wizard.ProverAction] interface. It is responsible for
// assigning the internal columns of ed.
func (ed *execData) Run(run *wizard.ProverRuntime) {

	var (
		selector  = ed.Inputs.Selector.GetColAssignment(run)
		data      = ed.Inputs.Data.GetColAssignment(run)
		size      = ed.Inputs.Data.Size()
		initState = make([]field.Element, 0, size)
		newState  = make([]field.Element, 0, size)
		finalHash = make([]field.Element, 0, size)
	)

	for row := 0; row < size; row++ {

		if row == 0 {
			initState = append(initState, field.Zero())
		}

		if row > 0 {

			prevSel := selector.Get(row - 1)

			if prevSel.IsOne() {
				initState = append(initState, newState[row-1])
			}

			if prevSel.IsZero() {
				initState = append(initState, initState[row-1])
			}
		}

		newState = append(newState,
			mimc.BlockCompression(
				initState[row],
				data.Get(row),
			),
		)

		currSel := selector.Get(row)

		if currSel.IsZero() && row == 0 {
			// We use zero as initial value this is just a filler, that is
			// unconstrained. It would also work with any other value but this
			// is fine as long execution data is never empty which is our case.
			finalHash = append(finalHash, field.Zero())
		}

		if currSel.IsZero() && row > 0 {
			finalHash = append(finalHash, finalHash[row-1])
		}

		if currSel.IsOne() {
			finalHash = append(finalHash, newState[row])
		}
	}

	run.AssignColumn(ed.InitState.GetColID(), smartvectors.NewRegular(initState))
	run.AssignColumn(ed.NextState.GetColID(), smartvectors.NewRegular(newState))
	run.AssignColumn(ed.FinalHash.GetColID(), smartvectors.NewRegular(finalHash))
	run.AssignLocalPoint(ed.FinalHashOpening.ID, finalHash[size-1])
}

func deriveName[S ~string](args ...any) S {
	fmtArgs := []string{"HASHING_EXECUTION_DATA"}
	for _, arg := range args {
		fmtArgs = append(fmtArgs, fmt.Sprintf("%v", arg))
	}
	return S(strings.Join(fmtArgs, "_"))
}
