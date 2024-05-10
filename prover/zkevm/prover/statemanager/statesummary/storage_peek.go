package statesummary

import (
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils/types"
)

// StoragePeek provides the columns to represent a peek to an account storage
// slot.
type StoragePeek struct {
	// StorageKey stores the storage key of the peeked slot
	Key HiLoColumns
	// OldValue represents the storage value that is being peeked at.
	OldValue HiLoColumns
	// NewValue represents the new value of the storage slot.
	NewValue HiLoColumns
}

// newStoragePeek returns a new StoragePeek object with initialized and
// unconstrained columns.
func newStoragePeek(comp *wizard.CompiledIOP, size int, name string) StoragePeek {
	return StoragePeek{
		Key:      newHiLoColumns(comp, size, name+"_KEY"),
		OldValue: newHiLoColumns(comp, size, name+"_OLD_VALUE"),
		NewValue: newHiLoColumns(comp, size, name+"_NEW_VALUE"),
	}
}

// storagePeekAssignmentBuilder is a convenience structure storing the column
// builders relating to a StoragePeek.
type storagePeekAssignmentBuilder struct {
	key, oldValue, newValue hiLoAssignmentBuilder
}

func newStoragePeekAssignmentBuilder(sp *StoragePeek) storagePeekAssignmentBuilder {
	return storagePeekAssignmentBuilder{
		key:      newHiLoAssignmentBuilder(sp.Key),
		oldValue: newHiLoAssignmentBuilder(sp.OldValue),
		newValue: newHiLoAssignmentBuilder(sp.NewValue),
	}
}

func (sh *storagePeekAssignmentBuilder) pushAllZeroes() {
	sh.key.pushZeroes()
	sh.oldValue.pushZeroes()
	sh.newValue.pushZeroes()
}

func (sh *storagePeekAssignmentBuilder) pushOnlyKey(key types.FullBytes32) {
	sh.key.push(key)
	sh.oldValue.pushZeroes()
	sh.newValue.pushZeroes()
}

func (sh *storagePeekAssignmentBuilder) pushOnlyOld(key, oldVal types.FullBytes32) {
	sh.key.push(key)
	sh.oldValue.push(oldVal)
	sh.newValue.pushZeroes()
}

func (sh *storagePeekAssignmentBuilder) pushOnlyNew(key, newVal types.FullBytes32) {
	sh.key.push(key)
	sh.oldValue.pushZeroes()
	sh.newValue.push(newVal)
}

func (sh *storagePeekAssignmentBuilder) push(key, oldVal, newVal types.FullBytes32) {
	sh.key.push(key)
	sh.oldValue.push(oldVal)
	sh.newValue.push(newVal)
}

func (sh *storagePeekAssignmentBuilder) padAssign(run *wizard.ProverRuntime) {
	sh.key.padAssign(run, types.FullBytes32{})
	sh.oldValue.padAssign(run, types.FullBytes32{})
	sh.newValue.padAssign(run, types.FullBytes32{})
}
