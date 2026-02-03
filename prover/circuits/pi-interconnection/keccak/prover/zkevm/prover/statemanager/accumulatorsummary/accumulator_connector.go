package accumulatorsummary

import (
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/zkevm/prover/statemanager/accumulator"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/zkevm/prover/statemanager/common"
)

// Constants that are used to compute the connector from AccumulatorSummary to StateSummary
const (
	READ_NON_ZERO      int = 0
	READ_ZERO          int = 1
	INSERT             int = 2
	UPDATE             int = 3
	DELETE             int = 4
	END_OF_ACCUMULATOR int = 5
	rowsReadNonZero    int = 2
	rowsReadZero       int = 2
	rowsInsert         int = 6
	rowsUpdate         int = 2
	rowsDelete         int = 6
)

// getOperationType inspects the indicator columns and returns the operation type
func getOperationType(run *wizard.ProverRuntime, acc *accumulator.Module, index int) int {
	isReadNonZero := acc.Cols.IsReadNonZero.GetColAssignmentAt(run, index)
	if isReadNonZero.IsOne() {
		return READ_NON_ZERO
	}
	isReadZero := acc.Cols.IsReadZero.GetColAssignmentAt(run, index)
	if isReadZero.IsOne() {
		return READ_ZERO
	}
	isInsert := acc.Cols.IsInsert.GetColAssignmentAt(run, index)
	if isInsert.IsOne() {
		return INSERT
	}
	isUpdate := acc.Cols.IsUpdate.GetColAssignmentAt(run, index)
	if isUpdate.IsOne() {
		return UPDATE
	}
	isDelete := acc.Cols.IsDelete.GetColAssignmentAt(run, index)
	if isDelete.IsOne() {
		return DELETE
	}
	// if none of the operations above is present, indicate that we reached the end/inactive area of the accumulator columns
	return END_OF_ACCUMULATOR
}

// Assign assigns the AccumulatorSummary, meaning it assign all the internal columns.
// It works by instantiating an AccumulatorSummaryAssignmentBuilder struct and
// processes the columns of the accumulator sequentially the segments created
// fetch the relevant values from the accumulator and populates segments that
// will be constant
func (accSummary *Module) Assign(run *wizard.ProverRuntime) {

	var (
		// the total number of rows in the corresponding accumulator segment
		numRowsAccSegment            int
		initialRoot, finalRoot, hKey field.Element
		initialHVal, finalHVal       field.Element
		builder                      = common.NewStateDiffAssignmentBuilder(accSummary.StateDiff)
		index                        = 0
		acc                          = accSummary.Inputs.Accumulator
	)

	for index < acc.NumRows() {
		opType := getOperationType(run, &acc, index)

		switch opType {
		case READ_NON_ZERO:
			numRowsAccSegment = rowsReadNonZero
			hKey = acc.Cols.LeafOpenings.HKey.GetColAssignmentAt(run, index)
			initialHVal = acc.Cols.LeafOpenings.HVal.GetColAssignmentAt(run, index)
			finalHVal = initialHVal

		case READ_ZERO:
			numRowsAccSegment = rowsReadZero
			// the sandwich check is enabled for READ_ZERO
			// therefore we can get the HKey from acc.Cols.HKey
			hKey = acc.Cols.HKey.GetColAssignmentAt(run, index)
			initialHVal = field.Zero()
			finalHVal = initialHVal

		case INSERT:
			numRowsAccSegment = rowsInsert
			// we check the 4th row of an accumulator INSERT segment
			hKey = acc.Cols.LeafOpenings.HKey.GetColAssignmentAt(run, index+3)
			initialHVal = field.Zero()
			finalHVal = acc.Cols.LeafOpenings.HVal.GetColAssignmentAt(run, index+3)

		case UPDATE:
			numRowsAccSegment = rowsUpdate
			hKey = acc.Cols.LeafOpenings.HKey.GetColAssignmentAt(run, index)
			initialHVal = acc.Cols.LeafOpenings.HVal.GetColAssignmentAt(run, index)
			finalHVal = acc.Cols.LeafOpenings.HVal.GetColAssignmentAt(run, index+1)

		case DELETE:
			numRowsAccSegment = rowsDelete // 6 rows for DELETE in the accumulator, we need to check the third and the fourth
			hKey = acc.Cols.LeafOpenings.HKey.GetColAssignmentAt(run, index+2)
			initialHVal = acc.Cols.LeafOpenings.HVal.GetColAssignmentAt(run, index+2)
			finalHVal = acc.Cols.LeafOpenings.HVal.GetColAssignmentAt(run, index+3) // empty leaf
		}

		if opType == END_OF_ACCUMULATOR {
			break
		}

		// set the roots
		initialRoot = acc.Cols.TopRoot.GetColAssignmentAt(run, index)
		finalRoot = acc.Cols.TopRoot.GetColAssignmentAt(run, index+numRowsAccSegment-1)

		builder.AddRows(numRowsAccSegment, hKey, initialHVal, finalHVal, initialRoot, finalRoot)
		index += numRowsAccSegment

	}
	builder.PadAndAssign(run)
}

// defineSegmentConstraints constrains the columns of the AccumulatorSummary and SegmentCounters,
// to ensure that AccumulatorSummary has constant values on each segment and that the counters of
// SegmentCounters are computed correctly
func defineSegmentConstraints(comp *wizard.CompiledIOP, mod *accumulator.Module, accSummary Module) {
	// mustBeConstantOnSubsegment defines a template for generating the
	// constraints ensuring that the values remain constant on an accumulator sub-segment.
	mustBeConstantOnSubsegment := func(col ifaces.Column) {
		comp.InsertGlobal(
			0,
			ifaces.QueryIDf("%v_IS_CONSTANT_DURING_SUB_SEGMENT", col.GetColID()),
			sym.Mul(
				mod.Cols.IsActiveAccumulator,
				sym.Sub(1, mod.Cols.IsFirst),
				sym.Sub(col, column.Shift(col, -1)),
			),
		)
	}
	// apply the constraints on each field
	mustBeConstantOnSubsegment(accSummary.HKey)
	mustBeConstantOnSubsegment(accSummary.InitialHVal)
	mustBeConstantOnSubsegment(accSummary.FinalHVal)
	mustBeConstantOnSubsegment(accSummary.InitialRoot)
	mustBeConstantOnSubsegment(accSummary.FinalRoot)
}

// accumulatorDefineHKeyConstraint defines a global constraint that ensures that the HKey is fetched correctly
// from the Accumulator when it is assigned to the AccumulatorSummary
func accumulatorDefineHKeyConstraint(comp *wizard.CompiledIOP, mod *accumulator.Module, accSummary Module) {
	// the following constraint relies on the fact that
	// IsReadZero, IsReadNonZero, IsUpdate, IsInsert, and IsDelete are mutually-exclusive flags
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("ACCUMULATOR_SUMMARY_HKEY_SELECTOR"),
		sym.Add(
			sym.Mul(
				sym.Add(
					mod.Cols.IsReadNonZero,
					mod.Cols.IsUpdate, // READ_NON_ZERO and UPDATE both check first row
				),
				mod.Cols.IsFirst,
				sym.Sub(mod.Cols.LeafOpenings.HKey, accSummary.HKey), // ensure hKey correctness in the Accumulator Summary
			),
			sym.Mul(
				mod.Cols.IsReadZero,                     // READ_ZERO checks first row
				mod.Cols.IsFirst,                        // exclude the 2nd row of a ReadZero segment, as it does not set an HKey
				sym.Sub(mod.Cols.HKey, accSummary.HKey), // we pick HKey differently, as the sandwich check is enabled
			),
			sym.Mul(
				mod.Cols.IsInsert,
				mod.Cols.IsFirst, // use isFirst flag so that we only check the hKey with offset 3
				sym.Sub(column.Shift(mod.Cols.LeafOpenings.HKey, 3), column.Shift(accSummary.HKey, 3)),
			),
			sym.Mul(
				mod.Cols.IsDelete,
				mod.Cols.IsFirst, // use isFirst flag so that we only check the hKey with offset 2
				sym.Sub(column.Shift(mod.Cols.LeafOpenings.HKey, 2), column.Shift(accSummary.HKey, 2)),
			),
		),
	)
}

// accumulatorDefineHValConstraints defines global constraints tbat ensures that the Initial and Final HVal is fetched correctly from the accumulator.Accumulator when it is assigned to the AccumulatorSummary
func accumulatorDefineHValConstraints(comp *wizard.CompiledIOP, mod *accumulator.Module, accSummary Module) {
	// only need to check InitialHVal for ReadNonZero, Update and Delete
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("ACCUMULATOR_SUMMARY_INITIAL_HVAL_SELECTOR"),
		sym.Add(
			sym.Mul(
				sym.Add(
					mod.Cols.IsReadNonZero,
					mod.Cols.IsUpdate, // READ_NON_ZERO and UPDATE check first row
				),
				mod.Cols.IsFirst,
				sym.Sub(mod.Cols.LeafOpenings.HVal, accSummary.InitialHVal),
			),
			sym.Mul(
				mod.Cols.IsDelete,
				mod.Cols.IsFirst, // use isFirst flag so that we only check the InitialHVal with offset 2
				sym.Sub(column.Shift(mod.Cols.LeafOpenings.HVal, 2), column.Shift(accSummary.InitialHVal, 2)),
			),
		),
	)

	// only need to check FinalHVal for ReadNonZero, Insert, and Update
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("ACCUMULATOR_SUMMARY_FINAL_HVAL_SELECTOR"),
		sym.Add(
			sym.Mul(
				sym.Add(
					mod.Cols.IsReadNonZero, // READ_NON_ZERO contains the same row two times, so we can also check the second row
					mod.Cols.IsUpdate,      // READ_NON_ZERO and UPDATE both check first row
				),
				mod.Cols.IsFirst,
				sym.Sub(column.Shift(mod.Cols.LeafOpenings.HVal, 1), column.Shift(accSummary.FinalHVal, 1)),
			),
			sym.Mul(
				mod.Cols.IsInsert,
				mod.Cols.IsFirst, // use isFirst flag so that we only check the FinalHVal with offset 3
				sym.Sub(column.Shift(mod.Cols.LeafOpenings.HVal, 3), column.Shift(accSummary.FinalHVal, 3)),
			),
		),
	)
}

// accumulatorDefineRootConstraints defines global constraints tbat ensures that the Initial and Final Roots are fetched correctly
// from the accumulator.Accumulator when it is assigned to the AccumulatorSummary
func accumulatorDefineRootConstraints(comp *wizard.CompiledIOP, mod *accumulator.Module, accSummary Module) {
	// Initial Root Selector
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("ACCUMULATOR_SUMMARY_INITIAL_ROOT_SELECTOR"),
		sym.Mul(
			mod.Cols.IsFirst,
			sym.Sub(mod.Cols.TopRoot, accSummary.InitialRoot),
		),
	)

	// use that IsReadZero, IsReadNonZero, IsUpdate, IsInsert and IsDelete are mutually-exclusive flags
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("ACCUMULATOR_SUMMARY_FINAL_ROOT_SELECTOR"),
		sym.Add(
			sym.Mul(
				sym.Add(
					mod.Cols.IsReadNonZero,
					mod.Cols.IsReadZero,
					mod.Cols.IsUpdate,
				),
				mod.Cols.IsFirst,
				sym.Sub(column.Shift(mod.Cols.TopRoot, 1), column.Shift(accSummary.FinalRoot, 1)),
			),
			sym.Mul(
				sym.Add(
					mod.Cols.IsInsert,
					mod.Cols.IsDelete,
				),
				mod.Cols.IsFirst,
				sym.Sub(column.Shift(mod.Cols.TopRoot, 5), column.Shift(accSummary.FinalRoot, 5)),
			),
		),
	)
}
