package statesummary

import (
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/utils/types"
)

var (
	emptyStorageRoot = field.NewFromString("3433815185717552096229990069654548485149817391158829673291533636212160492517")
)

// AccumulatorStatement represents the statement sent to the accumulator
type AccumulatorStatement struct {
	// IsReadZero, IsReadNonZero, IsInsert, IsUpdate, IsDelete are binary
	// flags indicating the type of traces being deferred to the accumulator
	// module.
	IsReadZero, IsReadNonZero, IsInsert, IsUpdate, IsDelete ifaces.Column

	// InitialHKey and FinalHKey store the initial and final accumulator's
	// hash keys.
	HKey ifaces.Column

	// InitialHVal and FinalHVal store the initial and final accumulator's
	// hash values.
	InitialHVal, FinalHVal ifaces.Column

	// InitialRoot and FinalRoot store the initial and final accumulator's
	// root hash.
	InitialRoot, FinalRoot ifaces.Column
}

// newAccumulatorStatement returns a new AccumulatorStatement with initialized
// columns that are not constrained.
func newAccumulatorStatement(comp *wizard.CompiledIOP, size int, name string) AccumulatorStatement {

	createCol := func(subName string) ifaces.Column {
		return comp.InsertCommit(
			0,
			ifaces.ColIDf("STATE_SUMMARY_%v_%v", name, subName),
			size,
		)
	}

	return AccumulatorStatement{
		IsReadZero:    createCol("IS_READ_ZERO"),
		IsReadNonZero: createCol("IS_READ_NON_ZERO"),
		IsInsert:      createCol("IS_INSERT"),
		IsUpdate:      createCol("IS_UPDATE"),
		IsDelete:      createCol("IS_DELETE"),
		HKey:          createCol("HKEY"),
		InitialHVal:   createCol("INITIAL_HVAL"),
		FinalHVal:     createCol("FINAL_HVAL"),
		InitialRoot:   createCol("INITIAL_ROOT"),
		FinalRoot:     createCol("FINAL_ROOT"),
	}
}

// csFlags enforces the IsReadZero, ... IsDelete flags to be mutually exclusive
// and binary. The booleanity is checks independently and the mutual exclusivity
// is done by checking that they sum to one.
func (as *AccumulatorStatement) csAccumulatorStatementFlags(comp *wizard.CompiledIOP, isActive ifaces.Column) {

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_ACCUMULATOR_STATEMENT_ISREADZERO_IS_BOOLEAN",
		sym.Mul(as.IsReadZero, sym.Sub(1, as.IsReadZero)),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_ACCUMULATOR_STATEMENT_ISREADNZ_IS_BOOLEAN",
		sym.Mul(as.IsReadNonZero, sym.Sub(1, as.IsReadNonZero)),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_ACCUMULATOR_STATEMENT_ISINSERT_IS_BOOLEAN",
		sym.Mul(as.IsInsert, sym.Sub(1, as.IsInsert)),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_ACCUMULATOR_STATEMENT_ISUPDATE_IS_BOOLEAN",
		sym.Mul(as.IsUpdate, sym.Sub(1, as.IsUpdate)),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_ACCUMULATOR_STATEMENT_ISDELETE_IS_BOOLEAN",
		sym.Mul(as.IsDelete, sym.Sub(1, as.IsDelete)),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_ACCUMULATOR_STATEMENT_FLAGS_ARE_MUTUALLY_EXCLUSIVE",
		sym.Add(as.IsReadZero, as.IsReadNonZero, as.IsInsert, as.IsUpdate, as.IsDelete, sym.Neg(isActive)),
	)
}

// accumulatorStatementAssignmentBuilder is a convenience structure storing
// the column builders relating to an AccumulatorStatement.
type accumulatorStatementAssignmentBuilder struct {
	isReadZero, isReadNonZero, isInsert, isUpdate, isDelete *vectorBuilder
	hKey                                                    *vectorBuilder
	initialHVal, finalHVal                                  *vectorBuilder
	initialRoot, finalRoot                                  *vectorBuilder
}

func newAccumulatorStatementAssignmentBuilder(as *AccumulatorStatement) accumulatorStatementAssignmentBuilder {
	return accumulatorStatementAssignmentBuilder{
		isReadZero:    newVectorBuilder(as.IsReadZero),
		isReadNonZero: newVectorBuilder(as.IsReadNonZero),
		isInsert:      newVectorBuilder(as.IsInsert),
		isUpdate:      newVectorBuilder(as.IsUpdate),
		isDelete:      newVectorBuilder(as.IsDelete),
		hKey:          newVectorBuilder(as.HKey),
		initialHVal:   newVectorBuilder(as.InitialHVal),
		finalHVal:     newVectorBuilder(as.FinalHVal),
		initialRoot:   newVectorBuilder(as.InitialRoot),
		finalRoot:     newVectorBuilder(as.FinalRoot),
	}
}

func (as *accumulatorStatementAssignmentBuilder) pushReadZero(root, hkey types.Bytes32) {
	as.isReadZero.PushOne()
	as.isReadNonZero.PushZero()
	as.isInsert.PushZero()
	as.isUpdate.PushZero()
	as.isDelete.PushZero()
	as.hKey.PushBytes32(hkey)
	as.initialHVal.PushZero()
	as.finalHVal.PushZero()
	as.initialRoot.PushBytes32(root)
	as.finalRoot.PushBytes32(root)
}

func (as *accumulatorStatementAssignmentBuilder) pushReadNonZero(root, hKey, hVal types.Bytes32) {
	as.isReadZero.PushZero()
	as.isReadNonZero.PushOne()
	as.isInsert.PushZero()
	as.isUpdate.PushZero()
	as.isDelete.PushZero()
	as.hKey.PushBytes32(hKey)
	as.initialHVal.PushBytes32(hVal)
	as.finalHVal.PushBytes32(hVal)
	as.initialRoot.PushBytes32(root)
	as.finalRoot.PushBytes32(root)
}

func (as *accumulatorStatementAssignmentBuilder) pushInsert(oldRoot, newRoot, hKey, newHVal types.Bytes32) {
	as.isReadZero.PushZero()
	as.isReadNonZero.PushZero()
	as.isInsert.PushOne()
	as.isUpdate.PushZero()
	as.isDelete.PushZero()
	as.hKey.PushBytes32(hKey)
	as.initialHVal.PushZero()
	as.finalHVal.PushBytes32(newHVal)
	as.initialRoot.PushBytes32(oldRoot)
	as.finalRoot.PushBytes32(newRoot)
}

func (as *accumulatorStatementAssignmentBuilder) pushUpdate(oldRoot, newRoot, hKey, oldHVal, newHVal types.Bytes32) {
	as.isReadZero.PushZero()
	as.isReadNonZero.PushZero()
	as.isInsert.PushZero()
	as.isUpdate.PushOne()
	as.isDelete.PushZero()
	as.hKey.PushBytes32(hKey)
	as.initialHVal.PushBytes32(oldHVal)
	as.finalHVal.PushBytes32(newHVal)
	as.initialRoot.PushBytes32(oldRoot)
	as.finalRoot.PushBytes32(newRoot)
}

func (as *accumulatorStatementAssignmentBuilder) pushDelete(oldRoot, newRoot, hKey, oldHVal types.Bytes32) {
	as.isReadZero.PushZero()
	as.isReadNonZero.PushZero()
	as.isInsert.PushZero()
	as.isUpdate.PushZero()
	as.isDelete.PushOne()
	as.hKey.PushBytes32(hKey)
	as.initialHVal.PushBytes32(oldHVal)
	as.finalHVal.PushZero()
	as.initialRoot.PushBytes32(oldRoot)
	as.finalRoot.PushBytes32(newRoot)
}

func (as *accumulatorStatementAssignmentBuilder) padAndAssign(run *wizard.ProverRuntime) {
	as.isReadZero.PadAndAssign(run)
	as.isReadNonZero.PadAndAssign(run)
	as.isInsert.PadAndAssign(run)
	as.isUpdate.PadAndAssign(run)
	as.isDelete.PadAndAssign(run)
	as.hKey.PadAndAssign(run)
	as.initialHVal.PadAndAssign(run)
	as.finalHVal.PadAndAssign(run)
	as.initialRoot.PadAndAssign(run)
	as.finalRoot.PadAndAssign(run)
}
