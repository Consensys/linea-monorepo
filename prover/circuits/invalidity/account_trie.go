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
// to prove the membership/non-membership of the account in the state.
type AccountTrie struct {
	// Account for the sender of the transaction  (if account does not exist, the balance is set to 0)
	Account GnarkAccount
	// LeafOpening of the Account in the Merkle tree (for the non-existing account only HKey is set legitimately to the hash of the address)
	LeafOpening GnarkLeafOpening
	// Merkle proof for the LeafOpening for the existing account (for the non-existing account, the proof is set to Minus to pass the Proof check trivially)
	MerkleProof      MerkleProofCircuit
	LeafOpeningMinus GnarkLeafOpening   // LeafOpening of the Minus Leaf (for non-existing account)
	LeafOpeningPlus  GnarkLeafOpening   // LeafOpening of the Plus Leaf (for non-existing account)
	MerkleProofMinus MerkleProofCircuit // Merkle proof for the Minus Leaf (for non-existing account)
	MerkleProofPlus  MerkleProofCircuit // Merkle proof for the Plus Leaf (for non-existing account)
	AccountExists    frontend.Variable  // 1 if the account exists, 0 otherwise

	NextFreeNode [common.NbLimbU64]koalagnark.Element // the next free node in the Merkle tree
	TopRoot      koalagnark.Octuplet                  // the top root of the Merkle tree
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
	Account          types.Account  // account of the existing account (if account does not exist, the balance and nonce are set to 0)
	LeafOpening      LeafOpening    // leaf opening of the account in Merkle tree (for existing account)
	LeafOpeningMinus LeafOpening    // LeafOpening of the Minus Leaf (for non-existing account)
	LeafOpeningPlus  LeafOpening    // LeafOpening of the Plus Leaf (for non-existing account)
	SubRoot          field.Octuplet // sub root of the Merkle tree (used in the proofs)
	NextFreeNode     int64          // the next free node in the Merkle tree
	TopRoot          field.Octuplet // the top root of the Merkle tree
	AccountExists    bool           // true if account exists, false for non-membership proof
}

// LeafOpening represents a leaf opening in the Merkle tree
// for the non-existing account only HKey is legitimately set to  the hash of the address, Leaf, Proof are set to Minus to pass the Proof check trivially.
type LeafOpening struct {
	LeafOpening ac.LeafOpening // leaf opening of the account in Merkle tree
	Leaf        field.Octuplet // hash of the LeafOpening
	Proof       smt.Proof      // Merkle proof associated with the leaf
}

// Define the constraints for membership (existing account) or non-membership
// (non-existing account) in the state trie.
//
// If AccountExists == 1 (existing):
//   - Hash(Account) == LeafOpening.HVal
//   - Hash(LeafOpening) == MerkleProof.Leaf
//   - MerkleProof verifies against Root
//
// If AccountExists == 0 (non-existing):
//   - Hash(LeafOpeningMinus) == MerkleProofMinus.Leaf
//   - Hash(LeafOpeningPlus) == MerkleProofPlus.Leaf
//   - Both Merkle proofs verify against Root
//   - Minus.Next == Plus.leafIndex (adjacency)
//   - Plus.Prev == Minus.leafIndex (adjacency)
//   - Minus.HKey < LeafOpening.HKey < Plus.HKey (wrapping)
func (ac *AccountTrie) Define(api frontend.API) error {

	koalaAPI := koalagnark.NewAPI(api)
	hasher := poseidon2_koalabear.NewKoalagnarkMDHasher(koalaAPI)

	exists := ac.AccountExists
	notExists := api.Sub(1, exists)
	existsKoala := koalaAPI.FromFrontendVar(exists)
	notExistsKoala := koalaAPI.FromFrontendVar(notExists)
	// exist is boolean
	api.AssertIsBoolean(exists)
	// TopRoot = Poseidon2(NextFreeNode, SubTreeRoot)
	// NextFreeNode is 4 x 16-bit limbs, padded to 16 elements to match WriteInt64On64Bytes
	for i := 0; i < 16-common.NbLimbU64; i++ {
		hasher.Write(koalagnark.NewElement(0))
	}
	hasher.Write(ac.NextFreeNode[:]...)
	hasher.WriteOctuplet(ac.MerkleProof.Root)
	recoveredTopRoot := hasher.Sum()
	koalaAPI.AssertOctupletEqual(ac.TopRoot, recoveredTopRoot)

	// ========== EXISTING ACCOUNT CHECKS (conditional on AccountExists) ==========

	// Hash(Account) == LeafOpening.HVal
	accountHash := ac.Account.Hash(hasher)
	koalaAPI.AssertOctupletEqualIf(existsKoala, accountHash, ac.LeafOpening.HVal)

	// Hash(LeafOpening) == MerkleProof.Leaf
	leafHash := ac.LeafOpening.Hash(hasher)
	koalaAPI.AssertOctupletEqualIf(existsKoala, leafHash, ac.MerkleProof.Leaf)

	// Verify Merkle proof for the existing account's leaf against the given root
	smt.KoalagnarkVerifyMerkleProof(api, ac.MerkleProof.Proofs, ac.MerkleProof.Leaf, ac.MerkleProof.Root)

	// ========== NON-EXISTING ACCOUNT CHECKS (conditional on !AccountExists) ==========

	// Hash(LeafOpeningMinus) == MerkleProofMinus.Leaf
	leafMinusHash := ac.LeafOpeningMinus.Hash(hasher)
	koalaAPI.AssertOctupletEqualIf(notExistsKoala, leafMinusHash, ac.MerkleProofMinus.Leaf)

	// Hash(LeafOpeningPlus) == MerkleProofPlus.Leaf
	leafPlusHash := ac.LeafOpeningPlus.Hash(hasher)
	koalaAPI.AssertOctupletEqualIf(notExistsKoala, leafPlusHash, ac.MerkleProofPlus.Leaf)

	// Verify Merkle proofs for minus and plus leaves against the given root (unconditional since existing-account assignment copies valid proofs)
	smt.KoalagnarkVerifyMerkleProof(api, ac.MerkleProofMinus.Proofs, ac.MerkleProofMinus.Leaf, ac.MerkleProofMinus.Root)
	smt.KoalagnarkVerifyMerkleProof(api, ac.MerkleProofPlus.Proofs, ac.MerkleProofPlus.Leaf, ac.MerkleProofPlus.Root)

	// --- Adjacency: minus and plus point to each other ---
	// minus.Next == plus.leafIndex
	minusNext := combine16BitLimbs(api, toNativeSlice(ac.LeafOpeningMinus.Next[:]))
	api.AssertIsEqual(api.Mul(notExists, api.Sub(minusNext, ac.MerkleProofPlus.Proofs.Path)), 0)

	// plus.Prev == minus.leafIndex
	plusPrev := combine16BitLimbs(api, toNativeSlice(ac.LeafOpeningPlus.Prev[:]))
	api.AssertIsEqual(api.Mul(notExists, api.Sub(plusPrev, ac.MerkleProofMinus.Proofs.Path)), 0)

	// --- Wrapping: minus.HKey < target.HKey < plus.HKey ---
	koalaAPI.AssertOctupletIsLessIf(notExists, ac.LeafOpeningMinus.HKey, ac.LeafOpening.HKey)
	koalaAPI.AssertOctupletIsLessIf(notExists, ac.LeafOpening.HKey, ac.LeafOpeningPlus.HKey)

	// the root is the same for all the proofs
	koalaAPI.AssetOctupletEqual(ac.MerkleProof.Root, ac.MerkleProofMinus.Root)
	koalaAPI.AssertOctupletEqual(ac.MerkleProof.Root, ac.MerkleProofPlus.Root)
	return nil
}

// Allocate the circuit
func (c *AccountTrie) Allocate(config Config) {
	c.MerkleProof.Proofs.Siblings = make([]poseidon2_koalabear.KoalagnarkOctuplet, config.Depth)
	c.MerkleProofMinus.Proofs.Siblings = make([]poseidon2_koalabear.KoalagnarkOctuplet, config.Depth)
	c.MerkleProofPlus.Proofs.Siblings = make([]poseidon2_koalabear.KoalagnarkOctuplet, config.Depth)
}

// Assign the circuit from [AccountTrieInputs]
func (c *AccountTrie) Assign(assi AccountTrieInputs) {

	// Account and leaf opening for the target account
	c.Account.Assign(assi.Account)
	c.LeafOpening.Assign(assi.LeafOpening.LeafOpening)

	// Merkle proofs: existing, minus, plus (all share the same root)
	assignMerkleProof(&c.MerkleProof, assi.LeafOpening, assi.SubRoot)
	assignMerkleProof(&c.MerkleProofMinus, assi.LeafOpeningMinus, assi.SubRoot)
	assignMerkleProof(&c.MerkleProofPlus, assi.LeafOpeningPlus, assi.SubRoot)

	// Leaf openings for minus/plus
	c.LeafOpeningMinus.Assign(assi.LeafOpeningMinus.LeafOpening)
	c.LeafOpeningPlus.Assign(assi.LeafOpeningPlus.LeafOpening)

	if assi.AccountExists {
		c.AccountExists = 1
	} else {
		c.AccountExists = 0
	}

	// NextFreeNode: 64-bit integer split into 4 x 16-bit limbs (big-endian)
	nfnLimbs := common.SplitBigEndianUint64(uint64(assi.NextFreeNode))
	for i := 0; i < common.NbLimbU64; i++ {
		c.NextFreeNode[i] = koalagnark.NewElement(limbBytesToUint64(nfnLimbs[i]))
	}

	// TopRoot: 8 KoalaBear elements
	for i := 0; i < 8; i++ {
		c.TopRoot[i] = koalagnark.NewElementFromKoala(assi.TopRoot[i])
	}
}

func assignMerkleProof(mp *MerkleProofCircuit, lo LeafOpening, root field.Octuplet) {
	mp.Proofs.Siblings = make([]poseidon2_koalabear.KoalagnarkOctuplet, len(lo.Proof.Siblings))
	for j := range lo.Proof.Siblings {
		for i := 0; i < 8; i++ {
			mp.Proofs.Siblings[j][i] = koalagnark.NewElementFromKoala(lo.Proof.Siblings[j][i])
		}
	}
	mp.Proofs.Path = lo.Proof.Path

	for i := 0; i < 8; i++ {
		mp.Leaf[i] = koalagnark.NewElementFromKoala(lo.Leaf[i])
		mp.Root[i] = koalagnark.NewElementFromKoala(root[i])
	}
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

// ComputeTopRoot computes TopRoot = Poseidon2(nextFreeNode || subTreeRoot).
// This matches the on-chain root stored in the Linea rollup contract.
func ComputeTopRoot(nextFreeNode int64, subTreeRoot types.KoalaOctuplet) types.KoalaOctuplet {
	hasher := poseidon2_koalabear.NewMDHasher()
	types.WriteInt64On64Bytes(hasher, nextFreeNode)
	subTreeRoot.WriteTo(hasher)
	digest := hasher.Sum(nil)
	r, err := types.BytesToKoalaOctuplet(digest)
	if err != nil {
		panic(err)
	}
	return r
}
