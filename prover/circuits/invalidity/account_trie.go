package invalidity

import (
	"github.com/consensys/gnark/frontend"
	gmimc "github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc/gkrmimc"
	ac "github.com/consensys/linea-monorepo/prover/crypto/state-management/accumulator"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/accumulator"
)

// AccountTrie includes the account and the data
// to prove the membership of the account in the state.
type AccountTrie struct {
	// Account for the sender of the transaction
	Account types.GnarkAccount
	// LeafOpening of the Account in the Merkle tree
	LeafOpening accumulator.GnarkLeafOpening
	// Merkle proof for the LeafOpening
	MerkleProof MerkleProofCircuit
}

// AccountTrieInputs collects the data for assigning the [AccountTrie]
type AccountTrieInputs struct {
	Account     types.Account
	LeafOpening ac.LeafOpening // leaf opening of the account in Merkle tree
	Leaf        types.Bytes32  // hash of the LeafOpening
	Proof       smt.Proof      // Merkle proof associated with the leaf
	Root        types.Bytes32  // root of the Merkle tree.
	Config      *smt.Config    // Merkle tree configuration
}

// Define the constraints for the membership of the account in the state
func (ac *AccountTrie) Define(api frontend.API) error {

	var (
		accountSlice     = []frontend.Variable{}
		leafOpeningSlice = []frontend.Variable{}
		account          = ac.Account
		leafOpening      = ac.LeafOpening
	)

	// Hash (Account) == LeafOpening.HVal
	hashAccount := MimcCircuit{
		PreImage: append(accountSlice,
			account.Nonce,
			account.Balance,
			account.StorageRoot,
			account.MimcCodeHash,
			account.KeccakCodeHashMSB,
			account.KeccakCodeHashLSB,
			account.CodeSize,
		),
		Hash: leafOpening.HVal,
	}
	err := hashAccount.Define(api)
	if err != nil {
		return err
	}

	// Hash(LeafOpening)= MerkleProof.Leaf
	hashLeafOpening := MimcCircuit{
		PreImage: append(leafOpeningSlice,
			leafOpening.Prev,
			leafOpening.Next,
			leafOpening.HKey,
			leafOpening.HVal),

		Hash: ac.MerkleProof.Leaf,
	}
	err = hashLeafOpening.Define(api)
	if err != nil {
		return err
	}

	// check that MerkleProof.Leaf is compatible with the state
	err = ac.MerkleProof.Define(api)

	if err != nil {
		return err
	}
	return nil
}

// Allocate the circuit
func (c *AccountTrie) Allocate(config Config) {
	c.MerkleProof.Proofs.Siblings = make([]frontend.Variable, config.Depth)
}

// Assign the circuit from [AAccountTrieInputs]
func (c *AccountTrie) Assign(assi AccountTrieInputs) {

	// assign the merkle proof
	var (
		witMerkle MerkleProofCircuit
		l         = assi.LeafOpening
	)

	witMerkle.Proofs.Siblings = make([]frontend.Variable, len(assi.Proof.Siblings))
	for j := 0; j < len(assi.Proof.Siblings); j++ {
		witMerkle.Proofs.Siblings[j] = assi.Proof.Siblings[j][:]
	}
	witMerkle.Proofs.Path = assi.Proof.Path
	witMerkle.Leaf = assi.Leaf[:]

	witMerkle.Root = assi.Root[:]

	// assign account and leafOpening
	a := assi.Account

	account := types.GnarkAccount{
		Nonce:    a.Nonce,
		Balance:  a.Balance,
		CodeSize: a.CodeSize,
	}

	account.StorageRoot = a.StorageRoot[:]
	account.MimcCodeHash = a.MimcCodeHash[:]
	account.KeccakCodeHashMSB = a.KeccakCodeHash[LIMB_SIZE:]
	account.KeccakCodeHashLSB = a.KeccakCodeHash[:LIMB_SIZE]

	leafOpening := accumulator.GnarkLeafOpening{
		Prev: l.Prev,
		Next: l.Next,
	}

	leafOpening.HKey = l.HKey[:]
	leafOpening.HVal = l.HVal[:]

	*c = AccountTrie{
		MerkleProof: witMerkle,
		LeafOpening: leafOpening,
		Account:     account,
	}
}

// MerkleProofCircuit defines the circuit for validating the Merkle proofs
type MerkleProofCircuit struct {
	Proofs smt.GnarkProof
	Leaf   frontend.Variable
	Root   frontend.Variable
}

// Define the constraints for a merkle proof
func (circuit *MerkleProofCircuit) Define(api frontend.API) error {
	hFac := gkrmimc.NewHasherFactory(api)
	hshM := hFac.NewHasher()

	smt.GnarkVerifyMerkleProof(api, circuit.Proofs, circuit.Leaf, circuit.Root, hshM)

	return nil
}

// Circuit defines a pre-image knowledge proof
// mimc( preImage) = public hash
type MimcCircuit struct {
	PreImage []frontend.Variable
	Hash     frontend.Variable
}

// Define declares the circuit's constraints
// Hash = mimc(PreImage)
func (circuit *MimcCircuit) Define(api frontend.API) error {
	// hash function
	mimc, _ := gmimc.NewMiMC(api)

	// mimc(preImage) == hash
	for _, toHash := range circuit.PreImage {
		mimc.Write(toHash)
	}

	api.AssertIsEqual(circuit.Hash, mimc.Sum())

	return nil
}
