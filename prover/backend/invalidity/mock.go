package invalidity

import (
	"fmt"
	"math/big"

	"github.com/consensys/linea-monorepo/prover/backend/execution/statemanager"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/accumulator"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/ethereum/go-ethereum/common"
)

// CreateMockAccountMerkleProof creates a mock Shomei ReadNonZeroTrace with valid Merkle proof.
// This is useful for testing the invalidity circuit with BadNonce/BadBalance cases.
func CreateMockAccountMerkleProof(address common.Address, account types.Account) statemanager.DecodedTrace {
	config := statemanager.MIMC_CONFIG

	// Compute HKey = MiMC(address)
	hKey := HashWithMiMC(address.Bytes())

	// Compute HVal = MiMC(account)
	hVal := HashAccount(account)

	// Create leaf opening
	leafOpening := accumulator.LeafOpening{
		Prev: 0,
		Next: 1,
		HKey: hKey,
		HVal: hVal,
	}

	// Compute leaf = MiMC(leafOpening)
	leaf := leafOpening.Hash(config)

	// Create a simple SMT with just this leaf
	tree := smt.NewEmptyTree(config)
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
		Value:        account,
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
		StorageRoot:    types.Bytes32{},
		MimcCodeHash:   types.Bytes32{},
		KeccakCodeHash: types.FullBytes32FromHex("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"),
		CodeSize:       0,
	}
}

// HashWithMiMC computes MiMC hash of input bytes
func HashWithMiMC(data []byte) types.Bytes32 {
	hasher := mimc.NewMiMC()
	hasher.Write(data)
	return types.Bytes32(hasher.Sum(nil))
}

// HashAccount computes MiMC hash of account (used for HVal in leaf opening)
func HashAccount(a types.Account) types.Bytes32 {
	hasher := mimc.NewMiMC()
	a.WriteTo(hasher)
	return types.Bytes32(hasher.Sum(nil))
}



