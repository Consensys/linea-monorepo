package common

import (
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils/types"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/zkevm/prover/common"
)

// StateDiff is a collection of column that appears in several of the modules
// of the state-manager of Linea.
//
// In state summary, we have a unique tuple of (hKey, initialHVal, finalHVal,
// initialRoot, finalRoot) for each of the state operations (e.g. INSERT,
// DELETE, UPDATE, READZERO, and READNONZERO). We want to check that this
// unique tuple is the same for the state summary and the accumulator module.
// In the accumulator module, we have six rows for INSERT and DELETE, and two
// rows for UPDATE, READZERO and READNONZERO. The bridge we call the
// accumulatorSummary constructs the tuple (hKey, initialHVal, finalHVal,
// initialRoot, finalRoot) from various rows of the accumulator module.
// (To exemplify further, depending on the type of state operation on the sparse
// Merkle tree, hkey will appear on a different row in the segment corresponding
// to the state operation—and similarly with the other components in the constructed tuple.)
type StateDiff struct {
	// HKey stores the initial and final accumulator's key hashes.
	HKey ifaces.Column

	// InitialHVal and FinalHVal store the initial and final accumulator's
	// hash of values.
	InitialHVal, FinalHVal ifaces.Column

	// InitialRoot and FinalRoot store the accumulator's initial and final
	// root hashes.
	InitialRoot, FinalRoot ifaces.Column
}

// NewStateDiff declares all the columns adding up to a [StateDiff] and returns
// the corresponding object.
func NewStateDiff(comp *wizard.CompiledIOP, size int, moduleName, name string) StateDiff {

	createCol := func(moduleName, subName string) ifaces.Column {
		return comp.InsertCommit(
			0,
			ifaces.ColIDf("%v_%v_%v", moduleName, name, subName),
			size,
		)
	}

	return StateDiff{
		HKey:        createCol(moduleName, "HKEY"),
		InitialHVal: createCol(moduleName, "INITIAL_HVAL"),
		FinalHVal:   createCol(moduleName, "FINAL_HVAL"),
		InitialRoot: createCol(moduleName, "INITIAL_ROOT"),
		FinalRoot:   createCol(moduleName, "FINAL_ROOT"),
	}
}

// StateDiffAssignmentBuilder is a convenience structure storing the column
// builders relating to an AccumulatorSummary.
type StateDiffAssignmentBuilder struct {
	HKey                   *common.VectorBuilder
	InitialHVal, FinalHVal *common.VectorBuilder
	InitialRoot, FinalRoot *common.VectorBuilder
}

// NewStateDiffAssignmentBuilder initializes a fresh
// [StateDiffAssignmentBuilder]
func NewStateDiffAssignmentBuilder(as StateDiff) StateDiffAssignmentBuilder {
	return StateDiffAssignmentBuilder{
		HKey:        common.NewVectorBuilder(as.HKey),
		InitialHVal: common.NewVectorBuilder(as.InitialHVal),
		FinalHVal:   common.NewVectorBuilder(as.FinalHVal),
		InitialRoot: common.NewVectorBuilder(as.InitialRoot),
		FinalRoot:   common.NewVectorBuilder(as.FinalRoot),
	}
}

// PushReadZero pushes the relevant row when a ReadZero occurs on the
// accumulator side.
func (as *StateDiffAssignmentBuilder) PushReadZero(root, hkey types.Bytes32) {
	as.HKey.PushBytes32(hkey)
	as.InitialHVal.PushZero()
	as.FinalHVal.PushZero()
	as.InitialRoot.PushBytes32(root)
	as.FinalRoot.PushBytes32(root)
}

// PushReadNonZero pushes a row onto `as` for a read-non-zero operation.
func (as *StateDiffAssignmentBuilder) PushReadNonZero(root, hKey, hVal types.Bytes32) {
	as.HKey.PushBytes32(hKey)
	as.InitialHVal.PushBytes32(hVal)
	as.FinalHVal.PushBytes32(hVal)
	as.InitialRoot.PushBytes32(root)
	as.FinalRoot.PushBytes32(root)
}

// PushInsert pushes a row representing an insertion onto `as`.
func (as *StateDiffAssignmentBuilder) PushInsert(oldRoot, newRoot, hKey, newHVal types.Bytes32) {
	as.HKey.PushBytes32(hKey)
	as.InitialHVal.PushZero()
	as.FinalHVal.PushBytes32(newHVal)
	as.InitialRoot.PushBytes32(oldRoot)
	as.FinalRoot.PushBytes32(newRoot)
}

// PushUpdate pushes a row representing an update onto `as`.
func (as *StateDiffAssignmentBuilder) PushUpdate(oldRoot, newRoot, hKey, oldHVal, newHVal types.Bytes32) {
	as.HKey.PushBytes32(hKey)
	as.InitialHVal.PushBytes32(oldHVal)
	as.FinalHVal.PushBytes32(newHVal)
	as.InitialRoot.PushBytes32(oldRoot)
	as.FinalRoot.PushBytes32(newRoot)
}

// PushDelete pushes a row representing a deletion onto `as`.
func (as *StateDiffAssignmentBuilder) PushDelete(oldRoot, newRoot, hKey, oldHVal types.Bytes32) {
	as.HKey.PushBytes32(hKey)
	as.InitialHVal.PushBytes32(oldHVal)
	as.FinalHVal.PushZero()
	as.InitialRoot.PushBytes32(oldRoot)
	as.FinalRoot.PushBytes32(newRoot)
}

// PadAndAssign pads all the column in `as` and assign them into `run`
func (as *StateDiffAssignmentBuilder) PadAndAssign(run *wizard.ProverRuntime) {
	as.HKey.PadAndAssign(run)
	as.InitialHVal.PadAndAssign(run)
	as.FinalHVal.PadAndAssign(run)
	as.InitialRoot.PadAndAssign(run)
	as.FinalRoot.PadAndAssign(run)
}

// addRows add rows to the builder that is used to construct an AccumulatorSummary
func (builder *StateDiffAssignmentBuilder) AddRows(numRowsAccSegment int, hKey, initialHVal, finalHVal, initialRoot, finalRoot field.Element) {
	for i := 1; i <= numRowsAccSegment; i++ {
		builder.HKey.PushField(hKey)
		builder.InitialHVal.PushField(initialHVal)
		builder.FinalHVal.PushField(finalHVal)
		builder.InitialRoot.PushField(initialRoot)
		builder.FinalRoot.PushField(finalRoot)
	}
}
