package statesummary

import (
	"github.com/consensys/zkevm-monorepo/prover/crypto/mimc"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

// AccountHashing provides a compilation context for the hashing of an account
type AccountHashing struct {
	// NewAccountHashing and OldAccountHashing store the result of the successive
	// calls to the MiMC permutation function up to the final result.
	NewAccountHashing, OldAccountHashing [7]ifaces.Column
	// AddressHashing stores the hash of the address of the account being peeked
	// at.
	AddressHashing ifaces.Column
}

// newAccountHashing returns a new AccountHashing with initialized and
// unconstrained columns.
func newAccountHashing(comp *wizard.CompiledIOP, size int, name string) AccountHashing {

	createCol := func(subName string) ifaces.Column {
		return comp.InsertCommit(
			0,
			ifaces.ColIDf("STATE_SUMMARY_%v_%v", name, subName),
			size,
		)
	}

	return AccountHashing{
		OldAccountHashing: [7]ifaces.Column{
			createCol("OLD_ACCOUNT_0"),
			createCol("OLD_ACCOUNT_1"),
			createCol("OLD_ACCOUNT_2"),
			createCol("OLD_ACCOUNT_3"),
			createCol("OLD_ACCOUNT_4"),
			createCol("OLD_ACCOUNT_5"),
			createCol("OLD_ACCOUNT_6"),
		},
		NewAccountHashing: [7]ifaces.Column{
			createCol("NEW_ACCOUNT_0"),
			createCol("NEW_ACCOUNT_1"),
			createCol("NEW_ACCOUNT_2"),
			createCol("NEW_ACCOUNT_3"),
			createCol("NEW_ACCOUNT_4"),
			createCol("NEW_ACCOUNT_5"),
			createCol("NEW_ACCOUNT_6"),
		},
		AddressHashing: createCol("ADDRESS"),
	}
}

func (ah *AccountHashing) OldAccountResult() ifaces.Column {
	return ah.OldAccountHashing[len(ah.OldAccountHashing)-1]
}

func (ah *AccountHashing) NewAccountResult() ifaces.Column {
	return ah.NewAccountHashing[len(ah.NewAccountHashing)-1]
}

func (ah *AccountHashing) AddressResult() ifaces.Column {
	return ah.AddressHashing
}

// csIntermediatHashesAreWellComputed ensures the hashing of the account is
// correctly computed. This includes both the account and the address of the
// account.
func (ah *AccountHashing) csIntermediateHashesAreWellComputed(
	comp *wizard.CompiledIOP,
	oldAcc *AccountPeek,
	newAcc *AccountPeek,
	address ifaces.Column,
) {

	cols := []ifaces.Column{
		oldAcc.Nonce,
		oldAcc.Balance,
		oldAcc.StorageRoot,
		oldAcc.MiMCCodeHash,
		oldAcc.KeccakCodeHash.Lo,
		oldAcc.KeccakCodeHash.Hi,
		oldAcc.CodeSize,
	}

	prevState := verifiercol.NewConstantCol(field.Zero(), oldAcc.Nonce.Size())

	for i := range cols {
		comp.InsertMiMC(
			0,
			ifaces.QueryIDf("STATE_SUMMARY_OLD_ACCOUNT_HASHING_%v", i),
			cols[i], prevState, ah.OldAccountHashing[i],
		)

		prevState = ah.OldAccountHashing[i]
	}

	cols = []ifaces.Column{
		newAcc.Nonce,
		newAcc.Balance,
		newAcc.StorageRoot,
		newAcc.MiMCCodeHash,
		newAcc.KeccakCodeHash.Lo,
		newAcc.KeccakCodeHash.Hi,
		newAcc.CodeSize,
	}

	prevState = verifiercol.NewConstantCol(field.Zero(), oldAcc.Nonce.Size())

	for i := range cols {
		comp.InsertMiMC(
			0,
			ifaces.QueryIDf("STATE_SUMMARY_NEW_ACCOUNT_HASHING_%v", i),
			cols[i], prevState, ah.NewAccountHashing[i],
		)

		prevState = ah.NewAccountHashing[i]
	}

	prevState = verifiercol.NewConstantCol(field.Zero(), oldAcc.Nonce.Size())

	comp.InsertMiMC(
		0,
		ifaces.QueryIDf("STATE_SUMMARY_ADDRESS_HASHING"),
		address, prevState, ah.AddressHashing,
	)
}

// StorageHashing represents the hashing a storage slot and its storage key.
type StorageHashing struct {
	NewValueHashing [2]ifaces.Column
	OldValueHashing [2]ifaces.Column
	KeyHashing      [2]ifaces.Column
}

// newStorageHashing returns a new StorageHashing instance with initialized
// and unconstrained columns
func newStorageHashing(comp *wizard.CompiledIOP, size int, name string) StorageHashing {

	createCol := func(subName string) ifaces.Column {
		return comp.InsertCommit(
			0,
			ifaces.ColIDf("STATE_SUMMARY_%v_%v", name, subName),
			size,
		)
	}

	return StorageHashing{
		NewValueHashing: [2]ifaces.Column{
			createCol("NEW_VALUE_0"),
			createCol("NEW_VALUE_1"),
		},
		OldValueHashing: [2]ifaces.Column{
			createCol("OLD_VALUE_0"),
			createCol("OLD_VALUE_1"),
		},
		KeyHashing: [2]ifaces.Column{
			createCol("KEY_0"),
			createCol("KEY_1"),
		},
	}

}

func (sh *StorageHashing) NewValueResult() ifaces.Column {
	return sh.NewValueHashing[len(sh.NewValueHashing)-1]
}

func (sh *StorageHashing) OldValueResult() ifaces.Column {
	return sh.OldValueHashing[len(sh.OldValueHashing)-1]
}

func (sh *StorageHashing) KeyResult() ifaces.Column {
	return sh.KeyHashing[len(sh.KeyHashing)-1]
}

func (sh *StorageHashing) csHashCorrectness(comp *wizard.CompiledIOP, sp *StoragePeek) {

	iVState := verifiercol.NewConstantCol(field.Zero(), sp.NewValue.Hi.Size())

	comp.InsertMiMC(
		0,
		ifaces.QueryIDf("STATE_SUMMARY_STORAGE_KEY_HASHING_LO"),
		sp.Key.Lo, iVState, sh.KeyHashing[0],
	)

	comp.InsertMiMC(
		0,
		ifaces.QueryIDf("STATE_SUMMARY_STORAGE_NEWVAL_HASHING_LO"),
		sp.NewValue.Lo, iVState, sh.NewValueHashing[0],
	)

	comp.InsertMiMC(
		0,
		ifaces.QueryIDf("STATE_SUMMARY_STORAGE_OLDVAL_HASHING_LO"),
		sp.OldValue.Lo, iVState, sh.OldValueHashing[0],
	)

	comp.InsertMiMC(
		0,
		ifaces.QueryIDf("STATE_SUMMARY_STORAGE_KEY_HASHING_HI"),
		sp.Key.Hi, sh.KeyHashing[0], sh.KeyHashing[1],
	)

	comp.InsertMiMC(
		0,
		ifaces.QueryIDf("STATE_SUMMARY_STORAGE_NEWVAL_HASHING_HI"),
		sp.NewValue.Hi, sh.NewValueHashing[0], sh.NewValueHashing[1],
	)

	comp.InsertMiMC(
		0,
		ifaces.QueryIDf("STATE_SUMMARY_STORAGE_OLDVAL_HASHING_HI"),
		sp.OldValue.Hi, sh.OldValueHashing[0], sh.OldValueHashing[1],
	)
}

// accountHashingAssignmentBuilder is a convenience structure storing the column
// builders relating to an AccountHashing.
type accountHashingAssignmentBuilder struct {
	newAccountHashing, oldAccountHashing [7]*vectorBuilder
	addressHashing                       *vectorBuilder
}

// newAccountHashingAssignmentBuilder
func newAccountHashingAssignmentBuilder(ah AccountHashing) accountHashingAssignmentBuilder {

	return accountHashingAssignmentBuilder{
		oldAccountHashing: [7]*vectorBuilder{
			newVectorBuilder(ah.OldAccountHashing[0]),
			newVectorBuilder(ah.OldAccountHashing[1]),
			newVectorBuilder(ah.OldAccountHashing[2]),
			newVectorBuilder(ah.OldAccountHashing[3]),
			newVectorBuilder(ah.OldAccountHashing[4]),
			newVectorBuilder(ah.OldAccountHashing[5]),
			newVectorBuilder(ah.OldAccountHashing[6]),
		},
		newAccountHashing: [7]*vectorBuilder{
			newVectorBuilder(ah.NewAccountHashing[0]),
			newVectorBuilder(ah.NewAccountHashing[1]),
			newVectorBuilder(ah.NewAccountHashing[2]),
			newVectorBuilder(ah.NewAccountHashing[3]),
			newVectorBuilder(ah.NewAccountHashing[4]),
			newVectorBuilder(ah.NewAccountHashing[5]),
			newVectorBuilder(ah.NewAccountHashing[6]),
		},
		addressHashing: newVectorBuilder(ah.AddressHashing),
	}
}

func (ah *accountHashingAssignmentBuilder) Resize(newLen int) {
	for i := range ah.oldAccountHashing {
		ah.oldAccountHashing[i].Resize(newLen)
		ah.newAccountHashing[i].Resize(newLen)
	}

	ah.addressHashing.Resize(newLen)
}

func (ah *accountHashingAssignmentBuilder) assignChunk(
	addr *vectorBuilder,
	initAp, finalAp *accountPeekAssignmentBuilder,
	start, stop int,
) {

	// Old account hashing
	stateSegment := make([]field.Element, stop-start)

	mimcVecCompression(stateSegment, initAp.nonce.slice[start:stop], stateSegment)
	copy(ah.oldAccountHashing[0].slice[start:stop], stateSegment)
	mimcVecCompression(stateSegment, initAp.balance.slice[start:stop], stateSegment)
	copy(ah.oldAccountHashing[1].slice[start:stop], stateSegment)
	mimcVecCompression(stateSegment, initAp.storageRoot.slice[start:stop], stateSegment)
	copy(ah.oldAccountHashing[2].slice[start:stop], stateSegment)
	mimcVecCompression(stateSegment, initAp.miMCCodeHash.slice[start:stop], stateSegment)
	copy(ah.oldAccountHashing[3].slice[start:stop], stateSegment)
	mimcVecCompression(stateSegment, initAp.keccakCodeHash.lo.slice[start:stop], stateSegment)
	copy(ah.oldAccountHashing[4].slice[start:stop], stateSegment)
	mimcVecCompression(stateSegment, initAp.keccakCodeHash.hi.slice[start:stop], stateSegment)
	copy(ah.oldAccountHashing[5].slice[start:stop], stateSegment)
	mimcVecCompression(stateSegment, initAp.codeSize.slice[start:stop], stateSegment)
	copy(ah.oldAccountHashing[6].slice[start:stop], stateSegment)

	// New account hashing
	stateSegment = make([]field.Element, stop-start)

	mimcVecCompression(stateSegment, finalAp.nonce.slice[start:stop], stateSegment)
	copy(ah.newAccountHashing[0].slice[start:stop], stateSegment)
	mimcVecCompression(stateSegment, finalAp.balance.slice[start:stop], stateSegment)
	copy(ah.newAccountHashing[1].slice[start:stop], stateSegment)
	mimcVecCompression(stateSegment, finalAp.storageRoot.slice[start:stop], stateSegment)
	copy(ah.newAccountHashing[2].slice[start:stop], stateSegment)
	mimcVecCompression(stateSegment, finalAp.miMCCodeHash.slice[start:stop], stateSegment)
	copy(ah.newAccountHashing[3].slice[start:stop], stateSegment)
	mimcVecCompression(stateSegment, finalAp.keccakCodeHash.lo.slice[start:stop], stateSegment)
	copy(ah.newAccountHashing[4].slice[start:stop], stateSegment)
	mimcVecCompression(stateSegment, finalAp.keccakCodeHash.hi.slice[start:stop], stateSegment)
	copy(ah.newAccountHashing[5].slice[start:stop], stateSegment)
	mimcVecCompression(stateSegment, finalAp.codeSize.slice[start:stop], stateSegment)
	copy(ah.newAccountHashing[6].slice[start:stop], stateSegment)

	// Address hashing
	stateSegment = make([]field.Element, stop-start)

	mimcVecCompression(stateSegment, addr.slice[start:stop], stateSegment)
	copy(ah.addressHashing.slice[start:stop], stateSegment)

}

func (ah *accountHashingAssignmentBuilder) PadAndAssign(run *wizard.ProverRuntime) {

	var (
		state = field.Zero()
	)

	for i := range ah.newAccountHashing {
		state = mimc.BlockCompression(state, field.Zero())

		if i == 0 {
			ah.addressHashing.PadAndAssign(run, state)
		}

		ah.newAccountHashing[i].PadAndAssign(run, state)
		ah.oldAccountHashing[i].PadAndAssign(run, state)
	}
}

// storagePeekAssignmentBuilder is a convenience structure storing the column
// builders relating to a StorageHashing.
type storageHashingAssignmentBuilder struct {
	newValueHashing [2]*vectorBuilder
	oldValueHashing [2]*vectorBuilder
	keyHashing      [2]*vectorBuilder
}

func newStorageHashingAssignmentBuilder(sh StorageHashing) storageHashingAssignmentBuilder {
	return storageHashingAssignmentBuilder{
		newValueHashing: [2]*vectorBuilder{
			newVectorBuilder(sh.NewValueHashing[0]),
			newVectorBuilder(sh.NewValueHashing[1]),
		},
		oldValueHashing: [2]*vectorBuilder{
			newVectorBuilder(sh.OldValueHashing[0]),
			newVectorBuilder(sh.OldValueHashing[1]),
		},
		keyHashing: [2]*vectorBuilder{
			newVectorBuilder(sh.KeyHashing[0]),
			newVectorBuilder(sh.KeyHashing[1]),
		},
	}
}

func (sh *storageHashingAssignmentBuilder) Resize(newLen int) {
	for i := range sh.keyHashing {
		sh.keyHashing[i].Resize(newLen)
		sh.oldValueHashing[i].Resize(newLen)
		sh.newValueHashing[i].Resize(newLen)
	}
}

func (sh *storageHashingAssignmentBuilder) assignChunk(
	sp *storagePeekAssignmentBuilder,
	start, stop int,
) {

	// Old storage value hashing
	stateSegment := make([]field.Element, stop-start)

	mimcVecCompression(stateSegment, sp.oldValue.lo.slice[start:stop], stateSegment)
	copy(sh.oldValueHashing[0].slice[start:stop], stateSegment)
	mimcVecCompression(stateSegment, sp.oldValue.hi.slice[start:stop], stateSegment)
	copy(sh.oldValueHashing[1].slice[start:stop], stateSegment)

	// New storage value hashing
	stateSegment = make([]field.Element, stop-start)

	mimcVecCompression(stateSegment, sp.newValue.lo.slice[start:stop], stateSegment)
	copy(sh.newValueHashing[0].slice[start:stop], stateSegment)
	mimcVecCompression(stateSegment, sp.newValue.hi.slice[start:stop], stateSegment)
	copy(sh.newValueHashing[1].slice[start:stop], stateSegment)

	// New account address
	stateSegment = make([]field.Element, stop-start)

	mimcVecCompression(stateSegment, sp.key.lo.slice[start:stop], stateSegment)
	copy(sh.keyHashing[0].slice[start:stop], stateSegment)
	mimcVecCompression(stateSegment, sp.key.hi.slice[start:stop], stateSegment)
	copy(sh.keyHashing[1].slice[start:stop], stateSegment)

}

func (sh *storageHashingAssignmentBuilder) PadAndAssign(run *wizard.ProverRuntime) {

	var (
		state = field.Zero()
	)

	for i := range sh.keyHashing {
		state = mimc.BlockCompression(state, field.Zero())
		sh.keyHashing[i].PadAndAssign(run, state)
		sh.newValueHashing[i].PadAndAssign(run, state)
		sh.oldValueHashing[i].PadAndAssign(run, state)
	}
}

func mimcVecCompression(oldState, block, newState []field.Element) {

	if len(oldState) != len(block) || len(block) != len(newState) {
		utils.Panic("the lengths are inconsistent: %v %v %v", len(oldState), len(block), len(newState))
	}

	if len(oldState) == 0 {
		return
	}

	for i := range oldState {
		newState[i] = mimc.BlockCompression(oldState[i], block[i])
	}
}
