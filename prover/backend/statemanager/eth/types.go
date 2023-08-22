package eth

import (
	"io"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/accumulator"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/hashtypes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type (
	// Aliases for the account tree
	AccountTrie = accumulator.ProverState[Address, Account]
	StorageTrie = accumulator.ProverState[FullBytes32, FullBytes32]

	// Account VS
	AccountVerifier = accumulator.VerifierState[Address, Account]
	StorageVerifier = accumulator.VerifierState[FullBytes32, FullBytes32]

	// ReadNonZeroTrace
	ReadNonZeroTraceWS = accumulator.ReadNonZeroTrace[Address, Account]
	ReadNonZeroTraceST = accumulator.ReadNonZeroTrace[FullBytes32, FullBytes32]

	// ReadZeroTrace
	ReadZeroTraceWS = accumulator.ReadZeroTrace[Address, Account]
	ReadZeroTraceST = accumulator.ReadZeroTrace[FullBytes32, FullBytes32]

	// InsertionTrace
	InsertionTraceWS = accumulator.InsertionTrace[Address, Account]
	InsertionTraceST = accumulator.InsertionTrace[FullBytes32, FullBytes32]

	// UpdateTrace
	UpdateTraceWS = accumulator.UpdateTrace[Address, Account]
	UpdateTraceST = accumulator.UpdateTrace[FullBytes32, FullBytes32]

	// DeletionTrace
	DeletionTraceWS = accumulator.DeletionTrace[Address, Account]
	DeletionTraceST = accumulator.DeletionTrace[FullBytes32, FullBytes32]

	// convenience alias : represent a digest that fits on a single field element
	Digest = hashtypes.Digest
)

// new types for 256 bits values that cannot be represented with a full field element
type FullBytes32 hashtypes.Digest

// wraps ethereum addresses
type Address common.Address

// return iff the string is ab 0x prefixed hex string with length 40
func IsHexAddress(s string) bool {
	// check the length (20 bytes * 2 + prefix (=2) = 42)
	if len(s) != 42 {
		return false
	}
	// then attempt to decode
	_, err := hexutil.Decode(s)
	return err == nil
}

// Specifies how "FullBytes32" are hashed/encoded before hashing
func (f FullBytes32) WriteTo(w io.Writer) (int64, error) {
	var lsb, msb Digest
	// Put the first half of f into the second half of msb
	// and the seconf half of f into the second half of lsb
	// The rest is zero.
	copy(msb[16:], f[:16])
	copy(lsb[16:], f[16:])
	n1, _ := lsb.WriteTo(w)
	n2, _ := msb.WriteTo(w)
	return n1 + n2, nil
}

// The rule for encoding an address is that it is zero-padded on the left
// to 32 bytes/ Under the hood the address is guaranteed to have exactly
// 20 bytes (bc it's an array)
func (a Address) WriteTo(w io.Writer) (int64, error) {
	slice := make([]byte, 32)
	copy(slice[12:], a[:])
	n, err := w.Write(slice)
	return int64(n), err
}

// Print in hex a digest, fullbytes or address
func (a Address) Hex() string {
	return common.Address(a).Hex()
}

// Print the hex address
func (b FullBytes32) Hex() string {
	return hexutil.Encode(b[:])
}
