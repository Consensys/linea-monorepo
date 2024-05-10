package statesummary

import (
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils/types"
)

// AccountPeek provides the columns to store the values of an account that
// we are peeking at.
type AccountPeek struct {
	// Nonce, Balance, MiMCCodeHash and CodeSize store the account field on a
	// single column each.
	Exists, Nonce, Balance, MiMCCodeHash, CodeSize, StorageRoot ifaces.Column
	// KeccakCodeHash stores the keccak code hash of the account.
	KeccakCodeHash HiLoColumns
}

// NewAccountPeek returns a new AccountPeek with initialized and unconstrained
// columns.
func newAccountPeek(comp *wizard.CompiledIOP, size int, name string) AccountPeek {

	createCol := func(subName string) ifaces.Column {
		return comp.InsertCommit(
			0,
			ifaces.ColIDf("STATE_SUMMARY_%v_%v", name, subName),
			size,
		)
	}

	return AccountPeek{
		Exists:         createCol("EXISTS"),
		Nonce:          createCol("NONCE"),
		Balance:        createCol("BALANCE"),
		MiMCCodeHash:   createCol("MIMC_CODEHASH"),
		CodeSize:       createCol("CODESIZE"),
		StorageRoot:    createCol("STORAGE_ROOT"),
		KeccakCodeHash: newHiLoColumns(comp, size, name+"_KECCAK_CODE_HASH"),
	}
}

// accountPeekAssignmentBuilder is a convenience structure storing the column
// builders relating to the an AccountPeek.
type accountPeekAssignmentBuilder struct {
	exists, nonce, balance, miMCCodeHash, codeSize, storageRoot *vectorBuilder
	keccakCodeHash                                              hiLoAssignmentBuilder
}

func newAccountPeekAssignmentBuilder(ap *AccountPeek) accountPeekAssignmentBuilder {
	return accountPeekAssignmentBuilder{
		exists:         newVectorBuilder(ap.Exists),
		nonce:          newVectorBuilder(ap.Nonce),
		balance:        newVectorBuilder(ap.Balance),
		miMCCodeHash:   newVectorBuilder(ap.MiMCCodeHash),
		codeSize:       newVectorBuilder(ap.CodeSize),
		storageRoot:    newVectorBuilder(ap.StorageRoot),
		keccakCodeHash: newHiLoAssignmentBuilder(ap.KeccakCodeHash),
	}
}

func (ss *accountPeekAssignmentBuilder) pushAll(acc types.Account) {
	ss.nonce.PushInt(int(acc.Nonce))
	if acc.Balance != nil {
		ss.balance.PushBytes32(types.LeftPadToBytes32(acc.Balance.Bytes()))
		ss.exists.PushOne()
	} else {
		ss.balance.PushZero()
		ss.exists.PushZero()
	}
	ss.codeSize.PushInt(int(acc.CodeSize))
	ss.miMCCodeHash.PushBytes32(acc.MimcCodeHash)
	ss.keccakCodeHash.push(acc.KeccakCodeHash)
	ss.storageRoot.PushBytes32(acc.StorageRoot)
}

func (ss *accountPeekAssignmentBuilder) pushOverrideStorageRoot(
	acc types.Account,
	storageRoot types.Bytes32,
) {
	ss.nonce.PushInt(int(acc.Nonce))
	if acc.Balance != nil {
		ss.balance.PushBytes32(types.LeftPadToBytes32(acc.Balance.Bytes()))
		ss.exists.PushOne()
	} else {
		ss.balance.PushZero()
		ss.exists.PushZero()
	}
	ss.codeSize.PushInt(int(acc.CodeSize))
	ss.miMCCodeHash.PushBytes32(acc.MimcCodeHash)
	ss.keccakCodeHash.push(acc.KeccakCodeHash)
	ss.storageRoot.PushBytes32(storageRoot)
}

func (ss *accountPeekAssignmentBuilder) PadAndAssign(run *wizard.ProverRuntime) {
	ss.exists.PadAndAssign(run)
	ss.nonce.PadAndAssign(run)
	ss.balance.PadAndAssign(run)
	ss.keccakCodeHash.padAssign(run, types.FullBytes32{})
	ss.miMCCodeHash.PadAndAssign(run)
	ss.storageRoot.PadAndAssign(run)
	ss.codeSize.PadAndAssign(run)
}
