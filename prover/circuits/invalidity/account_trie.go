package invalidity

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	ac "github.com/consensys/linea-monorepo/prover/crypto/state-management/accumulator"
	smt "github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

// AccountTrie includes the account and the data
// to prove the membership of the account in the state.
type AccountTrie struct {
	// Account for the sender of the transaction
	Account GnarkAccount
	// LeafOpening of the Account in the Merkle tree
	LeafOpening GnarkLeafOpening
	// Merkle proof for the LeafOpening
	MerkleProof MerkleProofCircuit
}

// GnarkAccount represent [types.Account] in gnark with Poseidon2-compatible layout
type GnarkAccount struct {
	Nonce             [common.NbLimbU64]frontend.Variable  // 4 elements
	Balance           [common.NbLimbU256]frontend.Variable // 16 elements
	StorageRoot       poseidon2_koalabear.GnarkOctuplet    // 8 elements (KoalaBear octuplet)
	LineaCodeHash     poseidon2_koalabear.GnarkOctuplet    // 8 elements (KoalaBear octuplet)
	KeccakCodeHashMSB [common.NbLimbU128]frontend.Variable // 8 elements (Hi)
	KeccakCodeHashLSB [common.NbLimbU128]frontend.Variable // 8 elements (Lo)
	CodeSize          [common.NbLimbU64]frontend.Variable  // 4 elements
}

// GnarkLeafOpening represents a leaf opening in gnark with Poseidon2 octuplet outputs
type GnarkLeafOpening struct {
	Prev [common.NbLimbU64]frontend.Variable
	Next [common.NbLimbU64]frontend.Variable
	HKey poseidon2_koalabear.GnarkOctuplet // 8 KoalaBear elements
	HVal poseidon2_koalabear.GnarkOctuplet // 8 KoalaBear elements
}

// AccountTrieInputs collects the data for assigning the [AccountTrie]
type AccountTrieInputs struct {
	Account     types.Account
	LeafOpening ac.LeafOpening // leaf opening of the account in Merkle tree
	Leaf        field.Octuplet // hash of the LeafOpening
	Proof       smt.Proof      // Merkle proof associated with the leaf
	Root        field.Octuplet // root of the Merkle tree.
}

// Define the constraints for the membership of the account in the state
func (ac *AccountTrie) Define(api frontend.API) error {

	// Create Poseidon2 hasher
	hasher, err := poseidon2_koalabear.NewGnarkMDHasher(api)
	if err != nil {
		return err
	}

	// Hash(Account) and verify it equals LeafOpening.HVal
	accountHash := ac.Account.Hash(&hasher)
	accountHash.AssertEqual(api, ac.LeafOpening.HVal)

	// Hash(LeafOpening) and verify it equals MerkleProof.Leaf
	leafHash := ac.LeafOpening.Hash(&hasher)
	leafHash.AssertEqual(api, ac.MerkleProof.Leaf)

	// Verify Merkle proof: MerkleProof.Leaf is in the tree with MerkleProof.Root
	return smt.GnarkVerifyMerkleProof(
		api,
		ac.MerkleProof.Proofs,
		ac.MerkleProof.Leaf,
		ac.MerkleProof.Root)
}

// Allocate the circuit
func (c *AccountTrie) Allocate(config Config) {
	c.MerkleProof.Proofs.Siblings = make([]poseidon2_koalabear.GnarkOctuplet, config.Depth)
}

// Assign the circuit from [AccountTrieInputs]
func (c *AccountTrie) Assign(assi AccountTrieInputs) {

	c.MerkleProof.Proofs.Siblings = make([]poseidon2_koalabear.GnarkOctuplet, len(assi.Proof.Siblings))
	for j := range assi.Proof.Siblings {
		c.MerkleProof.Proofs.Siblings[j].Assign(assi.Proof.Siblings[j])
	}
	c.MerkleProof.Proofs.Path = assi.Proof.Path

	c.MerkleProof.Leaf.Assign(assi.Leaf)
	c.MerkleProof.Root.Assign(assi.Root)

	c.Account.Assign(assi.Account)
	c.LeafOpening.Assign(assi.LeafOpening)
}

// MerkleProofCircuit defines the circuit for validating the Merkle proofs
// using Poseidon2 over KoalaBear
type MerkleProofCircuit struct {
	Proofs smt.GnarkProof
	Leaf   poseidon2_koalabear.GnarkOctuplet
	Root   poseidon2_koalabear.GnarkOctuplet
}

// Define the constraints for a merkle proof using Poseidon2
func (circuit *MerkleProofCircuit) Define(api frontend.API) error {
	return smt.GnarkVerifyMerkleProof(api, circuit.Proofs, circuit.Leaf, circuit.Root)
}

// Hash computes the Poseidon2 hash of the GnarkAccount using the provided hasher.
// The hasher is reset before use. Layout: 80 elements total.
func (a *GnarkAccount) Hash(hasher *poseidon2_koalabear.GnarkMDHasher) poseidon2_koalabear.GnarkOctuplet {
	hasher.Reset()

	// Nonce: 4 elements, padded to 16 (12 zeros prepended)
	for i := 0; i < 16-common.NbLimbU64; i++ {
		hasher.Write(frontend.Variable(0))
	}
	hasher.Write(a.Nonce[:]...)

	// Balance: 16 elements (no padding needed)
	hasher.Write(a.Balance[:]...)

	// StorageRoot: 8 elements
	hasher.Write(a.StorageRoot[:]...)

	// LineaCodeHash: 8 elements
	hasher.Write(a.LineaCodeHash[:]...)

	// KeccakCodeHash Hi (MSB): 8 elements
	hasher.Write(a.KeccakCodeHashMSB[:]...)

	// KeccakCodeHash Lo (LSB): 8 elements
	hasher.Write(a.KeccakCodeHashLSB[:]...)

	// CodeSize: 4 elements, padded to 16 (12 zeros prepended)
	for i := 0; i < 16-common.NbLimbU64; i++ {
		hasher.Write(frontend.Variable(0))
	}
	hasher.Write(a.CodeSize[:]...)

	return hasher.Sum()
}

// Hash computes the Poseidon2 hash of the leaf opening using the provided hasher.
// The hasher is reset before use. Layout: 48 elements total.
func (l *GnarkLeafOpening) Hash(hasher *poseidon2_koalabear.GnarkMDHasher) poseidon2_koalabear.GnarkOctuplet {
	hasher.Reset()

	// Prev: 4 elements, padded to 16 (12 zeros prepended)
	for i := 0; i < 16-common.NbLimbU64; i++ {
		hasher.Write(frontend.Variable(0))
	}
	hasher.Write(l.Prev[:]...)

	// Next: 4 elements, padded to 16 (12 zeros prepended)
	for i := 0; i < 16-common.NbLimbU64; i++ {
		hasher.Write(frontend.Variable(0))
	}
	hasher.Write(l.Next[:]...)

	// HKey: 8 elements
	hasher.Write(l.HKey[:]...)

	// HVal: 8 elements
	hasher.Write(l.HVal[:]...)

	return hasher.Sum()
}

// Assign converts a types.Account to a GnarkAccount for circuit assignment.
// The layout matches the native/wizard version with 16-bit limbs in big-endian order.
func (gnarkAccount *GnarkAccount) Assign(a types.Account) {

	// Nonce: 64-bit integer split into 4 x 16-bit limbs (big-endian order)
	nonceLimbs := common.SplitBigEndianUint64(uint64(a.Nonce))
	for i := 0; i < common.NbLimbU64; i++ {
		gnarkAccount.Nonce[i] = nonceLimbs[i]
	}

	// Balance: 256-bit integer split into 16 x 16-bit limbs (big-endian order)
	balanceLimbs := common.SplitBigEndianBigInt(a.Balance, 32)
	for i := 0; i < common.NbLimbU256; i++ {
		gnarkAccount.Balance[i] = balanceLimbs[i]
	}

	// StorageRoot: 8 KoalaBear field elements

	gnarkAccount.StorageRoot.Assign(a.StorageRoot)

	// LineaCodeHash: 8 KoalaBear field elements

	gnarkAccount.LineaCodeHash.Assign(a.LineaCodeHash)

	// KeccakCodeHash: 32 bytes split into Hi (16 bytes) and Lo (16 bytes)
	// Each part split into 8 x 16-bit limbs (big-endian order)
	keccakHiLimbs := common.SplitBytes(a.KeccakCodeHash[:16])
	for i := 0; i < common.NbLimbU128; i++ {
		gnarkAccount.KeccakCodeHashMSB[i] = keccakHiLimbs[i]
	}

	keccakLoLimbs := common.SplitBytes(a.KeccakCodeHash[16:])
	for i := 0; i < common.NbLimbU128; i++ {
		gnarkAccount.KeccakCodeHashLSB[i] = keccakLoLimbs[i]
	}

	// CodeSize: 64-bit integer split into 4 x 16-bit limbs (big-endian order)
	codeSizeLimbs := common.SplitBigEndianUint64(uint64(a.CodeSize))
	for i := 0; i < common.NbLimbU64; i++ {
		gnarkAccount.CodeSize[i] = codeSizeLimbs[i]
	}
}

// Assign converts a types.LeafOpening to a GnarkLeafOpening for circuit assignment.
func (leafOpening *GnarkLeafOpening) Assign(l ac.LeafOpening) {
	// Prev: 64-bit integer split into 4 x 16-bit limbs (big-endian order)
	prevLimbs := common.SplitBigEndianUint64(uint64(l.Prev))
	for i := 0; i < common.NbLimbU64; i++ {
		leafOpening.Prev[i] = prevLimbs[i]
	}

	// Next: 64-bit integer split into 4 x 16-bit limbs (big-endian order)
	nextLimbs := common.SplitBigEndianUint64(uint64(l.Next))
	for i := 0; i < common.NbLimbU64; i++ {
		leafOpening.Next[i] = nextLimbs[i]
	}

	leafOpening.HKey.Assign(field.Octuplet(l.HKey))
	leafOpening.HVal.Assign(field.Octuplet(l.HVal))
}
