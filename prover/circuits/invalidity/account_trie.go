package invalidity

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	ac "github.com/consensys/linea-monorepo/prover/crypto/state-management/accumulator"
	smt "github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
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
	Nonce             [common.NbLimbU64]koalagnark.Element  // 4 elements
	Balance           [common.NbLimbU256]koalagnark.Element // 16 elements
	StorageRoot       koalagnark.Octuplet                   // 8 elements (KoalaBear octuplet)
	LineaCodeHash     koalagnark.Octuplet                   // 8 elements (KoalaBear octuplet)
	KeccakCodeHashMSB [common.NbLimbU128]koalagnark.Element // 8 elements (Hi)
	KeccakCodeHashLSB [common.NbLimbU128]koalagnark.Element // 8 elements (Lo)
	CodeSize          [common.NbLimbU64]koalagnark.Element  // 4 elements
}

// GnarkLeafOpening represents a leaf opening in gnark with Poseidon2 octuplet outputs
type GnarkLeafOpening struct {
	Prev [common.NbLimbU64]koalagnark.Element
	Next [common.NbLimbU64]koalagnark.Element
	HKey koalagnark.Octuplet // 8 KoalaBear elements
	HVal koalagnark.Octuplet // 8 KoalaBear elements
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
	hasher := poseidon2_koalabear.NewKoalagnarkMDHasher(api)
	koalaAPI := koalagnark.NewAPI(api)

	// Hash(Account) and verify it equals LeafOpening.HVal
	accountHash := ac.Account.Hash(hasher)
	for i := 0; i < 8; i++ {
		koalaAPI.AssertIsEqual(accountHash[i], ac.LeafOpening.HVal[i])
	}

	// Hash(LeafOpening) and verify it equals MerkleProof.Leaf
	leafHash := ac.LeafOpening.Hash(hasher)
	koalaAPI.AssertOctupletEqual(leafHash, leafHash)

	// Verify Merkle proof: MerkleProof.Leaf is in the tree with MerkleProof.Root
	return smt.KoalagnarkVerifyMerkleProof(api, ac.MerkleProof.Proofs, ac.MerkleProof.Leaf, ac.MerkleProof.Root)
}

// Allocate the circuit
func (c *AccountTrie) Allocate(config Config) {
	c.MerkleProof.Proofs.Siblings = make([]poseidon2_koalabear.KoalagnarkOctuplet, config.Depth)
}

// Assign the circuit from [AccountTrieInputs]
func (c *AccountTrie) Assign(assi AccountTrieInputs) {

	c.MerkleProof.Proofs.Siblings = make([]poseidon2_koalabear.KoalagnarkOctuplet, len(assi.Proof.Siblings))
	for j := range assi.Proof.Siblings {
		for i := 0; i < 8; i++ {
			c.MerkleProof.Proofs.Siblings[j][i] = koalagnark.NewElementFromKoala(assi.Proof.Siblings[j][i])
		}
	}
	c.MerkleProof.Proofs.Path = assi.Proof.Path

	for i := 0; i < 8; i++ {
		c.MerkleProof.Leaf[i] = koalagnark.NewElementFromKoala(assi.Leaf[i])
		c.MerkleProof.Root[i] = koalagnark.NewElementFromKoala(assi.Root[i])
	}

	c.Account.Assign(assi.Account)
	c.LeafOpening.Assign(assi.LeafOpening)
}

// MerkleProofCircuit defines the circuit for validating the Merkle proofs
// using Poseidon2 over KoalaBear
type MerkleProofCircuit struct {
	Proofs smt.KoalagnarkProof
	Leaf   poseidon2_koalabear.KoalagnarkOctuplet
	Root   poseidon2_koalabear.KoalagnarkOctuplet
}

// Define the constraints for a merkle proof using Poseidon2
func (circuit *MerkleProofCircuit) Define(api frontend.API) error {
	return smt.KoalagnarkVerifyMerkleProof(api, circuit.Proofs, circuit.Leaf, circuit.Root)
}

// Hash computes the Poseidon2 hash of the GnarkAccount using the provided hasher.
// The hasher is reset before use. Layout: 80 elements total.
func (a *GnarkAccount) Hash(hasher *poseidon2_koalabear.KoalagnarkMDHasher) poseidon2_koalabear.KoalagnarkOctuplet {
	hasher.Reset()

	// Nonce: 4 elements, padded to 16 (12 zeros prepended)
	for i := 0; i < 16-common.NbLimbU64; i++ {
		hasher.Write(koalagnark.NewElement(0))
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
		hasher.Write(koalagnark.NewElement(0))
	}
	hasher.Write(a.CodeSize[:]...)

	return hasher.Sum()
}

// Hash computes the Poseidon2 hash of the leaf opening using the provided hasher.
// The hasher is reset before use. Layout: 48 elements total.
func (l *GnarkLeafOpening) Hash(hasher *poseidon2_koalabear.KoalagnarkMDHasher) poseidon2_koalabear.KoalagnarkOctuplet {
	hasher.Reset()

	// Prev: 4 elements, padded to 16 (12 zeros prepended)
	for i := 0; i < 16-common.NbLimbU64; i++ {
		hasher.Write(koalagnark.NewElement(0))
	}
	hasher.Write(l.Prev[:]...)

	// Next: 4 elements, padded to 16 (12 zeros prepended)
	for i := 0; i < 16-common.NbLimbU64; i++ {
		hasher.Write(koalagnark.NewElement(0))
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
		gnarkAccount.Nonce[i] = koalagnark.NewElement(limbBytesToUint64(nonceLimbs[i]))
	}

	// Balance: 256-bit integer split into 16 x 16-bit limbs (big-endian order)
	balanceLimbs := common.SplitBigEndianBigInt(a.Balance, 32)
	for i := 0; i < common.NbLimbU256; i++ {
		gnarkAccount.Balance[i] = koalagnark.NewElement(limbBytesToUint64(balanceLimbs[i]))
	}

	// StorageRoot: 8 KoalaBear field elements
	for i := 0; i < 8; i++ {
		gnarkAccount.StorageRoot[i] = koalagnark.NewElementFromKoala(field.Element(a.StorageRoot[i]))
	}

	// LineaCodeHash: 8 KoalaBear field elements
	for i := 0; i < 8; i++ {
		gnarkAccount.LineaCodeHash[i] = koalagnark.NewElementFromKoala(field.Element(a.LineaCodeHash[i]))
	}

	// KeccakCodeHash: 32 bytes split into Hi (16 bytes) and Lo (16 bytes)
	// Each part split into 8 x 16-bit limbs (big-endian order)
	keccakHiLimbs := common.SplitBytes(a.KeccakCodeHash[:16])
	for i := 0; i < common.NbLimbU128; i++ {
		gnarkAccount.KeccakCodeHashMSB[i] = koalagnark.NewElement(limbBytesToUint64(keccakHiLimbs[i]))
	}

	keccakLoLimbs := common.SplitBytes(a.KeccakCodeHash[16:])
	for i := 0; i < common.NbLimbU128; i++ {
		gnarkAccount.KeccakCodeHashLSB[i] = koalagnark.NewElement(limbBytesToUint64(keccakLoLimbs[i]))
	}

	// CodeSize: 64-bit integer split into 4 x 16-bit limbs (big-endian order)
	codeSizeLimbs := common.SplitBigEndianUint64(uint64(a.CodeSize))
	for i := 0; i < common.NbLimbU64; i++ {
		gnarkAccount.CodeSize[i] = koalagnark.NewElement(limbBytesToUint64(codeSizeLimbs[i]))
	}
}

// Assign converts a types.LeafOpening to a GnarkLeafOpening for circuit assignment.
func (leafOpening *GnarkLeafOpening) Assign(l ac.LeafOpening) {
	// Prev: 64-bit integer split into 4 x 16-bit limbs (big-endian order)
	prevLimbs := common.SplitBigEndianUint64(uint64(l.Prev))
	for i := 0; i < common.NbLimbU64; i++ {
		leafOpening.Prev[i] = koalagnark.NewElement(limbBytesToUint64(prevLimbs[i]))
	}

	// Next: 64-bit integer split into 4 x 16-bit limbs (big-endian order)
	nextLimbs := common.SplitBigEndianUint64(uint64(l.Next))
	for i := 0; i < common.NbLimbU64; i++ {
		leafOpening.Next[i] = koalagnark.NewElement(limbBytesToUint64(nextLimbs[i]))
	}

	// HKey: 8 KoalaBear field elements
	for i := 0; i < 8; i++ {
		leafOpening.HKey[i] = koalagnark.NewElementFromKoala(field.Element(l.HKey[i]))
	}

	// HVal: 8 KoalaBear field elements
	for i := 0; i < 8; i++ {
		leafOpening.HVal[i] = koalagnark.NewElementFromKoala(field.Element(l.HVal[i]))
	}
}

func limbBytesToUint64(limb []byte) uint64 {
	var res uint64
	for _, b := range limb {
		res = (res << 8) | uint64(b)
	}
	return res
}
