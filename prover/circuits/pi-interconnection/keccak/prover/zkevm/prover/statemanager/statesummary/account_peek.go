package statesummary

import (
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/dedicated/byte32cmp"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils/types"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/zkevm/prover/common"
	types2 "github.com/ethereum/go-ethereum/core/types"
)

// AccountPeek contains the view of the State-summary module regarding accounts.
// Namely, it stores all the account-related columns: the peeked address, the
// initial account and the final account.
type AccountPeek struct {
	// Initial and Final stores the view of the account at the beginning of an
	// account sub-segmenet and the at the current row.
	Initial, Final Account

	// HashInitial, HashFinal stores the hash of the initial account and the
	// hash of the final account
	HashInitial, HashFinal ifaces.Column

	// ComputeHashInitial and ComputeHashFinal are [wizard.ProverAction]
	// responsible for hashing the accounts.
	ComputeHashInitial, ComputeHashFinal wizard.ProverAction

	// InitialAndFinalAreSame is an indicator column set to 1 when the
	// initial and final account share the same hash and 0 otherwise.
	InitialAndFinalAreSame ifaces.Column

	// ComputeInitialAndFinalAreSame is a [wizard.ProverAction] responsible for
	// computing the column InitialAndFinalAreSame
	ComputeInitialAndFinalAreSame wizard.ProverAction

	// Address represents which account is being peeked by the module.
	// It is assigned by providing
	Address ifaces.Column

	// AddressHash is the hash of the account address
	AddressHash ifaces.Column

	// ComputeAddressHash is responsible for computing the AddressHash
	ComputeAddressHash wizard.ProverAction

	// AddressHashLimbs stores the limbs of the address
	AddressHashLimbs byte32cmp.LimbColumns

	// ComputeAddressLimbs computes the [AddressLimbs] column.
	ComputeAddressLimbs wizard.ProverAction

	// HasSameAddressAsPrev is an indicator column telling whether the previous
	// row has the same AccountAddress value as the current one.
	//
	// HasGreaterAddressAsPrev tells of the current address represents a larger
	// number than the previous one.
	HasSameAddressAsPrev, HasGreaterAddressAsPrev ifaces.Column

	// ComputeAddressComparison computes the HashSameAddressAsPrev and
	// HasGreaterAddressAsPrev.
	ComputeAddressComparison wizard.ProverAction
}

// newAccountPeek initializes all the columns related to the account and returns
// an [AccountPeek] object containing all of them. It does not generate
// constraints beyond the one coming from the dedicated wizard.
//
// The function also instantiates the dedicated columns for hashing the account,
// and operating limb-based comparisons.

// Account provides the columns to store the values of an account that
// we are peeking at.
type Account struct {
	// Nonce, Balance, MiMCCodeHash and CodeSize store the account field on a
	// single column each.
	Exists, Nonce, Balance, MiMCCodeHash, CodeSize, StorageRoot ifaces.Column
	// KeccakCodeHash stores the keccak code hash of the account.
	KeccakCodeHash common.HiLoColumns
	// ExpectedHubCodeHash is almost the same as the KeccakCodeHash, with the difference
	// than when the account does not exist, it contains the keccak hash of the empty string
	ExpectedHubCodeHash common.HiLoColumns
	// HasEmptyCodeHash is an indicator column indicating whether the current
	// account has an empty codehash
	HasEmptyCodeHash             ifaces.Column
	CptHasEmptyCodeHash          wizard.ProverAction
	ExistsAndHasNonEmptyCodeHash ifaces.Column
}

// newAccount returns a new AccountPeek with initialized and unconstrained
// columns.

// newAccountPeekAssignmentBuilder initializes a fresh accountPeekAssignmentBuilder

// accountAssignmentBuilder is a convenience structure storing the column
// builders relating to the an Account.
type accountAssignmentBuilder struct {
	exists, nonce, balance, miMCCodeHash, codeSize, storageRoot *common.VectorBuilder
	keccakCodeHash                                              common.HiLoAssignmentBuilder
	expectedHubCodeHash                                         common.HiLoAssignmentBuilder
	existsAndHasNonEmptyCodeHash                                *common.VectorBuilder
}

// newAccountAssignmentBuilder returns a new [accountAssignmentBuilder] bound
// to an [Account].
func newAccountAssignmentBuilder(ap *Account) accountAssignmentBuilder {
	return accountAssignmentBuilder{
		exists:                       common.NewVectorBuilder(ap.Exists),
		nonce:                        common.NewVectorBuilder(ap.Nonce),
		balance:                      common.NewVectorBuilder(ap.Balance),
		miMCCodeHash:                 common.NewVectorBuilder(ap.MiMCCodeHash),
		codeSize:                     common.NewVectorBuilder(ap.CodeSize),
		storageRoot:                  common.NewVectorBuilder(ap.StorageRoot),
		existsAndHasNonEmptyCodeHash: common.NewVectorBuilder(ap.ExistsAndHasNonEmptyCodeHash),
		keccakCodeHash:               common.NewHiLoAssignmentBuilder(ap.KeccakCodeHash),
		expectedHubCodeHash:          common.NewHiLoAssignmentBuilder(ap.ExpectedHubCodeHash),
	}
}

// pushAll stacks the value of a [types.Account] as a new row on the receiver.
func (ss *accountAssignmentBuilder) pushAll(acc types.Account) {
	// accountExists is telling whether the intent is to push an empty account
	accountExists := acc.Balance != nil

	ss.nonce.PushInt(int(acc.Nonce))

	// This is telling us whether the intent is to push an empty account
	if accountExists {
		ss.balance.PushBytes32(types.LeftPadToBytes32(acc.Balance.Bytes()))
		ss.exists.PushOne()
		ss.keccakCodeHash.Push(acc.KeccakCodeHash)
		// if account exists push the same Keccak code hash
		ss.expectedHubCodeHash.Push(acc.KeccakCodeHash)
	} else {
		ss.balance.PushZero()
		ss.exists.PushZero()
		ss.keccakCodeHash.PushZeroes()
		// if account does not exist push empty codehash
		ss.expectedHubCodeHash.Push(types.FullBytes32(types2.EmptyCodeHash))
	}

	ss.codeSize.PushInt(int(acc.CodeSize))
	ss.miMCCodeHash.PushBytes32(acc.MimcCodeHash)
	ss.storageRoot.PushBytes32(acc.StorageRoot)
	ss.existsAndHasNonEmptyCodeHash.PushBoolean(accountExists && acc.CodeSize > 0)
}

// pushOverrideStorageRoot is as [accountAssignmentBuilder.pushAll] but allows
// the caller to override the StorageRoot field with the provided one.
func (ss *accountAssignmentBuilder) pushOverrideStorageRoot(
	acc types.Account,
	storageRoot types.Bytes32,
) {
	// accountExists is telling whether the intent is to push an empty account
	accountExists := acc.Balance != nil

	ss.nonce.PushInt(int(acc.Nonce))

	// This is telling us whether the intent is to push an empty account
	if accountExists {
		ss.balance.PushBytes32(types.LeftPadToBytes32(acc.Balance.Bytes()))
		ss.exists.PushOne()
		ss.keccakCodeHash.Push(acc.KeccakCodeHash)
		// if account exists push the same codehash
		ss.expectedHubCodeHash.Push(acc.KeccakCodeHash)
	} else {
		ss.balance.PushZero()
		ss.exists.PushZero()
		ss.keccakCodeHash.PushZeroes()
		// if account does not exist push empty codehash
		ss.expectedHubCodeHash.Push(types.FullBytes32(types2.EmptyCodeHash))
	}

	ss.codeSize.PushInt(int(acc.CodeSize))
	ss.miMCCodeHash.PushBytes32(acc.MimcCodeHash)
	ss.storageRoot.PushBytes32(storageRoot)
	ss.existsAndHasNonEmptyCodeHash.PushBoolean(accountExists && acc.CodeSize > 0)
}

// PadAndAssign terminates the receiver by padding all the columns representing
// the account with "zeroes" rows up to the target size of the column and then
// assigning the underlying [ifaces.Column] object with it.
func (ss *accountAssignmentBuilder) PadAndAssign(run *wizard.ProverRuntime) {
	ss.exists.PadAndAssign(run)
	ss.nonce.PadAndAssign(run)
	ss.balance.PadAndAssign(run)
	ss.keccakCodeHash.PadAssign(run, types.FullBytes32{})
	ss.expectedHubCodeHash.PadAssign(run, types.FullBytes32{})
	ss.miMCCodeHash.PadAndAssign(run)
	ss.storageRoot.PadAndAssign(run)
	ss.codeSize.PadAndAssign(run)
	ss.existsAndHasNonEmptyCodeHash.PadAndAssign(run)
}
