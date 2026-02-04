package mock

import (
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils/types"
)

// OperationType specifies the part of an account that is touched by a state
// access. The 0 value corresponds to an invalid OperationType.
type OperationType int

const (
	// Storage denotes an access to the storage of an account
	Storage OperationType = iota + 1
	// Balance denotes an access to the balance of an account
	Balance
	// Nonce denotes an access to the nonce of an account
	Nonce
	// CodeSize denotes an access to the code size of an account. Only for
	// read-only operations. Otherwise, it would be a deploy.
	Codesize
	// CodeHash denotes an access to the code hash of an account. Only for
	// read-only operations. Otherwise, it would be a deploy.
	Codehash
	// SquashedAccount denotes an summary of all the accesses made on an account
	// aside from its storage. This is meant to be used by the Shomei trace
	// generator.
	SquashedAccount
	// AccountErasal indicates that the account deleted from the state. This
	// happens at the end of a transaction calling selfdestruct when the account
	// was existing before the transaction or during a transaction when the
	// selfdestruct is called on an ephemeral account (i.e. an account created
	// by the same transaction).
	AccountErasal
	// AccountInit indicates that the account has been (re)deployed or simply
	// initialized. This can happen when CREATE or CREATE2 are called or when
	// a non-existing account receives a transfer.
	AccountInit
)

// StateAccessLog represents an atomic operation over an account, the
// operation can relate to the account itself or to the storage of the account.
// Stacking StateAccessLog in a slice allows representing a succession of
// operation on the state of Ethereum. However, this is different from a
// sequence of EVM opcode: StateAccessLog have to be understood as "immediately"
// effective while EVM opcodes can have deferred effect on the state.
//
// An example of these two being differents is that self-destruct does not
// immediately remove the account from the state if the account pre-existed
// the transaction.
type StateAccessLog struct {
	// Address indicates the address of the account being accessed by the
	// current state log.
	Address types.EthAddress
	// Block indicates the block in which the access takes place.
	Block int
	// Type adds information about which part of the account is being accessed
	// or the type of access.
	Type OperationType
	// Key is only used for storage accesses and indicates the storage key
	// being accessed.
	Key types.FullBytes32
	// Value contains the "read" value if the log is a read-only log or the
	// new value if the log is write operation.
	Value any
	// OldValue is used only for write operations and indicates the value
	// of the accessed state item prior to being updated.
	OldValue any
	// IsWrite tells whether the logged state access is a write operation or
	// a read-only operation.
	IsWrite bool // false means read
}
