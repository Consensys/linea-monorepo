package statesummary

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/byte32cmp"
	dedicatedmimc "github.com/consensys/linea-monorepo/prover/protocol/dedicated/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

// StoragePeek provides the columns to represent a peek to an account storage
// slot.
type StoragePeek struct {
	// StorageKey stores the storage key of the peeked slot
	Key common.HiLoColumns
	// OldValue represents the storage value that is being peeked at.
	OldValue common.HiLoColumns
	// NewValue represents the new value of the storage slot.
	NewValue common.HiLoColumns

	// OldValueIsZero and NewValueIsZero are indicator columns set to 1 when the
	// old/new value is set to zero and 0 otherwise.
	OldValueIsZero, NewValueIsZero [common.NbLimbU256]ifaces.Column

	// ComputeOldValueIsZero and ComputeNewValueIsZero are respectively
	// responsible for assigning OldValueIsZero and NewValueIsZero.
	ComputeOldValueIsZero, ComputeNewValueIsZero [common.NbLimbU256]wizard.ProverAction

	// KeyLimbs represents the key in limb decomposition.
	KeyLimbs [common.NbLimbU256]byte32cmp.LimbColumns

	// ComputeKeyLimbs and ComputeKeyLimbLo are responsible for computing the
	// "hi" and the "lo" limbs of the KeyLimbs.
	ComputeKeyLimbs [common.NbLimbU256]wizard.ProverAction

	// KeyIncreased is a column indicating whether the current storage
	// key is strictly greater than the previous one.
	KeyIncreased ifaces.Column

	// ComputeKeyIncreased computes KeyIncreased.
	ComputeKeyIncreased wizard.ProverAction

	// OldValueHash and NewValueHash store the hash of the storage
	// peek. It is not passed to the accumulator statement directly as we have
	// special handling for the case where the storage value is zero.
	OldValueHash, NewValueHash [common.NbLimbU256]ifaces.Column

	// ComputeOldValueHash and ComputeNewStorageValue hash compute
	// respectively OldStorageValueHash and NewStorageValueHash
	ComputeOldValueHash, ComputeNewValueHash *dedicatedmimc.HashingCtx

	// KeyHash stores the hash of the storage keys
	KeyHash [common.NbLimbU256]ifaces.Column

	// ComputeStorageKeyHash computes the KeyHash column.
	ComputeKeyHash *dedicatedmimc.HashingCtx

	// OldAndNewAreEqual is an indicator column telling whether the old and the
	// new storage value are equal with a 1 and 0 else.
	OldAndNewValuesAreEqual [common.NbLimbU256]ifaces.Column

	// ComputeOldAndNewValuesAreEqual computes OldAndNewValuesAreEqual
	ComputeOldAndNewValuesAreEqual [common.NbLimbU256]wizard.ProverAction
}

// newStoragePeek returns a new StoragePeek object with initialized and
// unconstrained columns.
func newStoragePeek(comp *wizard.CompiledIOP, size int, name string) StoragePeek {
	res := StoragePeek{
		Key:      common.NewHiLoColumns(comp, size, name+"_KEY"),
		OldValue: common.NewHiLoColumns(comp, size, name+"_OLD_VALUE"),
		NewValue: common.NewHiLoColumns(comp, size, name+"_NEW_VALUE"),
	}

	res.ComputeOldValueHash = dedicatedmimc.HashOf(
		comp,
		[][]ifaces.Column{
			res.OldValue.Lo[:],
			res.OldValue.Hi[:],
		},
	)

	res.OldValueHash = res.ComputeOldValueHash.Result()

	panic("add Poseidon query here")
	res.ComputeNewValueHash = dedicatedmimc.HashOf(
		comp,
		[][]ifaces.Column{
			res.NewValue.Lo[:],
			res.NewValue.Hi[:],
		},
	)

	res.NewValueHash = res.ComputeNewValueHash.Result()

	for i := range common.NbLimbU256 {
		res.OldAndNewValuesAreEqual[i], res.ComputeOldAndNewValuesAreEqual[i] = dedicated.IsZero(
			comp,
			sym.Sub(res.OldValueHash[i], res.NewValueHash[i]),
		).GetColumnAndProverAction()
	}

	res.ComputeKeyHash = dedicatedmimc.HashOf(
		comp,
		[][]ifaces.Column{
			res.Key.Lo[:],
			res.Key.Hi[:],
		},
	)

	res.KeyHash = res.ComputeKeyHash.Result()
	zeroStorageHash := hashOfZeroStorage()

	for i := range common.NbLimbU256 {
		res.OldValueIsZero[i], res.ComputeOldValueIsZero[i] = dedicated.IsZero(
			comp,
			sym.Sub(res.OldValueHash[i], zeroStorageHash[i]),
		).GetColumnAndProverAction()

		res.NewValueIsZero[i], res.ComputeNewValueIsZero[i] = dedicated.IsZero(
			comp,
			sym.Sub(res.NewValueHash[i], zeroStorageHash[i]),
		).GetColumnAndProverAction()
	}

	keyLimbColumbs := byte32cmp.LimbColumns{LimbBitSize: common.LimbBytes * 8, IsBigEndian: true}
	shiftedLimbColumbs := byte32cmp.LimbColumns{LimbBitSize: common.LimbBytes * 8, IsBigEndian: true}
	for i := range common.NbLimbU256 {
		res.KeyLimbs[i], res.ComputeKeyLimbs[i] = byte32cmp.Decompose(comp, res.KeyHash[i], 1, common.LimbBytes*8)

		keyLimbColumbs.Limbs = append(keyLimbColumbs.Limbs, res.KeyLimbs[i].Limbs...)
		shiftedLimbColumbs.Limbs = append(shiftedLimbColumbs.Limbs, res.KeyLimbs[i].Shift(-1).Limbs...)
	}

	res.KeyIncreased, _, _, res.ComputeKeyIncreased = byte32cmp.CmpMultiLimbs(
		comp,
		keyLimbColumbs, shiftedLimbColumbs,
	)

	return res
}

// storagePeekAssignmentBuilder is a convenience structure storing the column
// builders relating to a StoragePeek.
type storagePeekAssignmentBuilder struct {
	key, oldValue, newValue common.HiLoAssignmentBuilder
}

// newStoragePeekAssignmentBuilder constructs a fresh [storagePeekAssignmentBuilder]
// with empty columns.
func newStoragePeekAssignmentBuilder(sp *StoragePeek) storagePeekAssignmentBuilder {
	return storagePeekAssignmentBuilder{
		key:      common.NewHiLoAssignmentBuilder(sp.Key),
		oldValue: common.NewHiLoAssignmentBuilder(sp.OldValue),
		newValue: common.NewHiLoAssignmentBuilder(sp.NewValue),
	}
}

// pushAllZeroes pushes a zero row onto the receiver
func (sh *storagePeekAssignmentBuilder) pushAllZeroes() {
	sh.key.PushZeroes()
	sh.oldValue.PushZeroes()
	sh.newValue.PushZeroes()
}

// pushOnlyKey pushes the key onto the the "key" columns builder and zero on
// the others
func (sh *storagePeekAssignmentBuilder) pushOnlyKey(key types.FullBytes32) {
	var keyLimbs [common.NbLimbU256][]byte
	copy(keyLimbs[:], common.SplitBytes(key[:]))

	sh.key.Push(keyLimbs)
	sh.oldValue.PushZeroes()
	sh.newValue.PushZeroes()
}

// pushOnlyOld pushes a row where the keys and the old value are the one provided
// by the caller and the new value is zero.
func (sh *storagePeekAssignmentBuilder) pushOnlyOld(key, oldVal types.FullBytes32) {
	var keyLimbs [common.NbLimbU256][]byte
	copy(keyLimbs[:], common.SplitBytes(key[:]))

	var oldValueLimbs [common.NbLimbU256][]byte
	copy(oldValueLimbs[:], common.SplitBytes(oldVal[:]))

	sh.key.Push(keyLimbs)
	sh.oldValue.Push(oldValueLimbs)
	sh.newValue.PushZeroes()
}

// pushOnlyNew pushes a row where the key and the new value are the one provided
// by the caller and the old value is set to zero.
func (sh *storagePeekAssignmentBuilder) pushOnlyNew(key, newVal types.FullBytes32) {
	var keyLimbs [common.NbLimbU256][]byte
	copy(keyLimbs[:], common.SplitBytes(key[:]))

	var newValueLimbs [common.NbLimbU256][]byte
	copy(newValueLimbs[:], common.SplitBytes(newVal[:]))

	sh.key.Push(keyLimbs)
	sh.oldValue.PushZeroes()
	sh.newValue.Push(newValueLimbs)
}

// push pushes a row where the key, the old value and the new value are the
// one provided by the caller.
func (sh *storagePeekAssignmentBuilder) push(key, oldVal, newVal types.FullBytes32) {
	var keyLimbs [common.NbLimbU256][]byte
	copy(keyLimbs[:], common.SplitBytes(key[:]))

	var oldValueLimbs [common.NbLimbU256][]byte
	copy(oldValueLimbs[:], common.SplitBytes(oldVal[:]))

	var newValueLimbs [common.NbLimbU256][]byte
	copy(newValueLimbs[:], common.SplitBytes(newVal[:]))

	sh.key.Push(keyLimbs)
	sh.oldValue.Push(oldValueLimbs)
	sh.newValue.Push(newValueLimbs)
}

// padAssign pads and assigns the columns of the storage peek into `run`.
func (sh *storagePeekAssignmentBuilder) padAssign(run *wizard.ProverRuntime) {
	sh.key.PadAssign(run, [common.NbLimbU256][]byte{})
	sh.oldValue.PadAssign(run, [common.NbLimbU256][]byte{})
	sh.newValue.PadAssign(run, [common.NbLimbU256][]byte{})
}

// hashOfZeroStorage returns the hash of (0, 0) which is what we use for empty
// storage slots.
func hashOfZeroStorage() []field.Element {
	res := []field.Element{field.Zero()}
	res = common.BlockCompression(res, []field.Element{field.Zero()})
	return common.BlockCompression(res, []field.Element{field.Zero()})
}
