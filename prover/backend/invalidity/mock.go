package invalidity

import (
	"fmt"
	"math/big"

	"github.com/consensys/linea-monorepo/prover/backend/execution/statemanager"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/accumulator"
	smt "github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_koalabear"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/ethereum/go-ethereum/common"
)

// CreateMockAccountMerkleProof creates a mock Shomei ReadNonZeroTrace with valid Merkle proof.
// This is useful for testing the invalidity circuit with BadNonce/BadBalance cases.
func CreateMockAccountMerkleProof(address common.Address, account types.Account) statemanager.DecodedTrace {

	// Compute HKey = Poseidon2(address) using WriteTo for canonical serialization
	// (EthAddress.WriteTo left-pads each 2-byte pair to 4 bytes, matching the circuit's
	// 16-bit limb decomposition and the accumulator's hash function)
	hKey := HashAddress(types.EthAddress(address))

	// Compute HVal = Poseidon2(account)
	hVal := HashAccount(account)

	// Create leaf opening
	leafOpening := accumulator.LeafOpening{
		Prev: 0,
		Next: 1,
		HKey: hKey,
		HVal: hVal,
	}

	// Compute leaf = Poseidon2(leafOpening)
	leaf := leafOpening.Hash()

	// Create a simple SMT with just this leaf
	tree := smt.NewEmptyTree()
	tree.Update(0, leaf)
	root := tree.Root

	// Get the Merkle proof
	proof, err := tree.Prove(0)
	if err != nil {
		panic(fmt.Sprintf("failed to create Merkle proof: %v", err))
	}

	// Create the ReadNonZeroTrace
	readTrace := statemanager.ReadNonZeroTraceWS{
		Type:         0, // ReadNonZero
		Location:     "0x",
		NextFreeNode: 2,
		Key:          types.EthAddress(address),
		Value:        account.WrappedForShomeiTraces(),
		SubRoot:      root,
		LeafOpening:  leafOpening,
		Proof:        proof,
	}

	// Wrap in DecodedTrace
	return statemanager.DecodedTrace{
		Type:       0,
		Location:   "0x",
		Underlying: readTrace,
	}
}

// CreateMockEOAAccount creates a mock EOA (Externally Owned Account) with given nonce and balance.
func CreateMockEOAAccount(nonce int64, balance *big.Int) types.Account {
	return types.Account{
		Nonce:          nonce,
		Balance:        balance,
		StorageRoot:    types.KoalaOctuplet{},
		LineaCodeHash:  types.KoalaOctuplet{},
		KeccakCodeHash: types.FullBytes32FromHex("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"),
		CodeSize:       0,
	}
}

// HashWithPoseidon2 computes Poseidon2 hash of input bytes, returning a KoalaOctuplet.
func HashAddress(add types.EthAddress) types.KoalaOctuplet {
	hasher := poseidon2_koalabear.NewMDHasher()
	add.WriteTo(hasher)
	digest := hasher.Sum(nil)
	var d types.KoalaOctuplet
	if err := d.SetBytes(digest); err != nil {
		panic(err)
	}
	return d
}

// HashAccount computes Poseidon2 hash of account (used for HVal in leaf opening)
func HashAccount(a types.Account) types.KoalaOctuplet {
	hasher := poseidon2_koalabear.NewMDHasher()
	a.WriteTo(hasher)
	digest := hasher.Sum(nil)
	var d types.KoalaOctuplet
	if err := d.SetBytes(digest); err != nil {
		panic(err)
	}
	return d
}
