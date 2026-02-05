package statesummary

import (
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/dedicated/byte32cmp"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/zkevm/prover/common"
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
