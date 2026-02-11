package invalidity_test

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/accumulator"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
	. "github.com/consensys/linea-monorepo/prover/utils/types"
	linTypes "github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/require"
)

// generate a tree for testing (proofs, leafs, root) using smt_koalabear
func getMerkleProof(t *testing.T) (smt_koalabear.Proof, field.Octuplet, field.Octuplet) {

	depth := 10

	// Generate random field elements for leaves
	nbLeaves := 1 << depth
	leaves := make([]field.Octuplet, nbLeaves)
	for i := range leaves {
		leaves[i] = field.RandomOctuplet()
	}

	// And generate the tree using Poseidon2
	tree := smt_koalabear.NewTree(leaves)

	// Get leaf and proof at position 0
	leaf, _ := tree.GetLeaf(0)
	proof, _ := tree.Prove(0)

	// Directly verify the proof
	err := smt_koalabear.Verify(&proof, leaf, tree.Root)
	require.NoErrorf(t, err, "pos #%v, proof #%v", 0, proof)

	return proof, leaf, tree.Root
}

// test [invalidity.MerkleProofCircuit] with Poseidon2
func TestMerkleProofs(t *testing.T) {

	// generate witness using smt_koalabear
	proof, leaf, root := getMerkleProof(t)

	var witness invalidity.MerkleProofCircuit

	// Assign siblings (each is a KoalagnarkOctuplet)
	witness.Proofs.Siblings = make([]poseidon2_koalabear.KoalagnarkOctuplet, len(proof.Siblings))
	for j := 0; j < len(proof.Siblings); j++ {
		for k := 0; k < 8; k++ {
			witness.Proofs.Siblings[j][k] = koalagnark.NewElementFromKoala(proof.Siblings[j][k])
		}
	}
	witness.Proofs.Path = proof.Path

	// Assign leaf and root (each is a KoalagnarkOctuplet)
	for k := 0; k < 8; k++ {
		witness.Leaf[k] = koalagnark.NewElementFromKoala(leaf[k])
		witness.Root[k] = koalagnark.NewElementFromKoala(root[k])
	}

	// compile circuit using KoalaBear field
	var circuit invalidity.MerkleProofCircuit
	circuit.Proofs.Siblings = make([]poseidon2_koalabear.KoalagnarkOctuplet, len(proof.Siblings))

	ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, &circuit, frontend.IgnoreUnconstrainedInputs())
	if err != nil {
		t.Fatal(err)
	}

	// solve the circuit with KoalaBear modulus
	twitness, err := frontend.NewWitness(&witness, koalabear.Modulus())
	if err != nil {
		t.Fatal(err)
	}
	err = ccs.IsSolved(twitness)
	if err != nil {
		t.Fatal(err)
	}
}

// it tests the Poseidon2 Hashing over [types.Account]
func TestHashAccount(t *testing.T) {

	var (
		// generate Poseidon2 witness for Hash(Account)
		a = tcases[1].Account
	)

	// Hash the account natively using Poseidon2
	nativeHasher := poseidon2_koalabear.NewMDHasher()
	a.WriteTo(nativeHasher)
	nativeHashBytes := nativeHasher.Sum(nil)

	var expectedHash linTypes.KoalaOctuplet
	err := expectedHash.SetBytes(nativeHashBytes)
	require.NoError(t, err)

	// Create the circuit for testing account hashing
	circuit := &testAccountHashCircuit{}

	// Create witness with assigned values
	witness := &testAccountHashCircuit{}
	witness.Account.Assign(a)
	witness.ExpectedHash.Assign(field.Octuplet(expectedHash))

	// Compile the circuit using KoalaBear field (for Poseidon2)
	ccs, err := frontend.CompileU32(
		koalabear.Modulus(),
		scs.NewBuilder,
		circuit,
	)
	require.NoError(t, err)

	wit, err := frontend.NewWitness(witness, koalabear.Modulus())
	require.NoError(t, err)

	err = ccs.IsSolved(wit)
	require.NoError(t, err)
}

// testAccountHashCircuit is a test circuit for verifying account hashing
type testAccountHashCircuit struct {
	Account      invalidity.GnarkAccount
	ExpectedHash poseidon2_koalabear.GnarkOctuplet
}

func (c *testAccountHashCircuit) Define(api frontend.API) error {
	// Create Poseidon2 hasher
	hasher := poseidon2_koalabear.NewKoalagnarkMDHasher(api)

	// Use GnarkAccount.Hash from account_trie.go
	hash := c.Account.Hash(hasher)

	// Assert hash matches expected
	for i := 0; i < 8; i++ {
		api.AssertIsEqual(hash[i].Native(), c.ExpectedHash[i])
	}

	return nil
}

// it creates a merkle tree for the given [accumulator.LeafOpening] using smt_koalabear
func genShomei(t *testing.T, tcases []TestCases, depth int) (*smt_koalabear.Tree, []smt_koalabear.Proof, []field.Octuplet) {

	var leaves = []field.Octuplet{}
	for _, c := range tcases {
		// Hash the account using Poseidon2
		accountHash := hashAccountNative(&c.Account)

		// Create leaf opening and hash it
		leafOpening := accumulator.LeafOpening{
			Prev: c.Leaf.Prev,
			Next: c.Leaf.Next,
			HKey: c.Leaf.HKey,
			HVal: accountHash,
		}
		leafHash := leafOpening.Hash()
		leaves = append(leaves, leafHash.ToOctuplet())
	}

	// Build the same tree by adding the leaves one by one
	tree := smt_koalabear.NewEmptyTree(depth)
	for i := range leaves {
		tree.Update(i, leaves[i])
	}

	var (
		leafs  = []field.Octuplet{}
		proofs = []smt_koalabear.Proof{}
	)
	// Make valid proofs
	for i := range leaves {

		leaf, _ := tree.GetLeaf(i)
		proof, _ := tree.Prove(i)

		// Directly verify the proof
		err := smt_koalabear.Verify(&proof, leaf, tree.Root)
		require.NoErrorf(t, err, "pos #%v, proof #%v", i, proof)

		leafs = append(leafs, leaf)
		proofs = append(proofs, proof)
	}

	return tree, proofs, leaves
}

// hashAccountNative hashes an account using Poseidon2
func hashAccountNative(a *Account) linTypes.KoalaOctuplet {
	hasher := poseidon2_koalabear.NewMDHasher()
	a.WriteTo(hasher)
	digest := hasher.Sum(nil)
	var d linTypes.KoalaOctuplet
	if err := d.SetBytes(digest); err != nil {
		panic(err)
	}
	return d
}

// it gets a leaf via its position and check it has the expected value.
func TestShomei(t *testing.T) {

	depth := 10
	tree, _, leaves := genShomei(t, tcases, depth)

	for i := range leaves {
		leaf, _ := tree.GetLeaf(i)
		c := tcases[i]

		// Hash the account using Poseidon2
		accountHash := hashAccountNative(&c.Account)

		// Create expected leaf hash
		expectedLeafOpening := accumulator.LeafOpening{
			Prev: c.Leaf.Prev,
			Next: c.Leaf.Next,
			HKey: c.Leaf.HKey,
			HVal: accountHash,
		}
		expectedLeaf := expectedLeafOpening.Hash().ToOctuplet()

		assert.Equal(t, leaf, expectedLeaf)
	}
}

type TestCases struct {
	Account        Account
	Leaf           accumulator.LeafOpening
	Tx             types.DynamicFeeTx
	FromAddress    common.Address
	TxHash         common.Hash
	InvalidityType invalidity.InvalidityType
}

var tcases = []TestCases{

	{
		Account: Account{
			Balance: big.NewInt(0),
		},
		Leaf: accumulator.LeafOpening{
			Prev: 0,
			Next: 1,
			HKey: hKeyFromAddress(common.HexToAddress("0x00aed6")),
			HVal: hValFromAccount(Account{
				Balance: big.NewInt(0),
			}),
		},
		Tx: types.DynamicFeeTx{
			ChainID:   big.NewInt(59144), // Linea mainnet chain ID
			Nonce:     1,                 // valid nonce
			Value:     big.NewInt(1),     //invalid balance
			Gas:       1,
			GasFeeCap: big.NewInt(1), // gas price
		},
		FromAddress:    common.HexToAddress("0x00aed6"),
		TxHash:         common.HexToHash("0x3f1d2e2b4c3f4e5d6c7b8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d"),
		InvalidityType: 1,
	},
	{
		// EOA
		Account: Account{
			Nonce:          65,
			Balance:        big.NewInt(5690),
			StorageRoot:    MustHexToKoalabearOctuplet("0x0b1dfeef3db4956540da8a5f785917ef1ba432e521368da60a0a1ce430425666"),
			LineaCodeHash:  MustHexToKoalabearOctuplet("0x729aac4455d43f2c69e53bb75f8430193332a4c32cafd9995312fa8346929e73"),
			KeccakCodeHash: FullBytes32FromHex("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"),
			CodeSize:       0,
		},

		Leaf: accumulator.LeafOpening{
			Prev: 0,
			Next: 2,
			HKey: hKeyFromAddress(common.HexToAddress("0x00aed7")),
			HVal: hValFromAccount(Account{
				Nonce:          65,
				Balance:        big.NewInt(5690),
				StorageRoot:    MustHexToKoalabearOctuplet("0x0b1dfeef3db4956540da8a5f785917ef1ba432e521368da60a0a1ce430425666"),
				LineaCodeHash:  MustHexToKoalabearOctuplet("0x729aac4455d43f2c69e53bb75f8430193332a4c32cafd9995312fa8346929e73"),
				KeccakCodeHash: FullBytes32FromHex("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"),
				CodeSize:       0,
			}),
		},
		Tx: types.DynamicFeeTx{
			ChainID:   big.NewInt(59144), // Linea mainnet chain ID
			Nonce:     65,                // invalid nonce
			Value:     big.NewInt(5700),  // invalid value
			Gas:       1,
			GasFeeCap: big.NewInt(1), // gas price
		},
		FromAddress:    common.HexToAddress("0x00aed7"),
		TxHash:         common.HexToHash("0x4f1d2e2b4c3f4e5d6c7b8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9e"),
		InvalidityType: 0, //  0 = BadNonce, 1 = BadBalance
	},
	{
		// Another EOA
		Account: Account{
			Nonce:          65,
			Balance:        big.NewInt(835),
			StorageRoot:    MustHexToKoalabearOctuplet("0x1c41acc261451aae253f621857172d6339919d18059f35921a50aafc69eb5c39"),
			LineaCodeHash:  MustHexToKoalabearOctuplet("0x7b688b215329825e5b00e4aa4e1857bc17afab503a87ecc063614b9b227106b2"),
			KeccakCodeHash: FullBytes32FromHex("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"),
			CodeSize:       0,
		},
		Leaf: accumulator.LeafOpening{
			Prev: 1,
			Next: 3,
			HKey: hKeyFromAddress(common.HexToAddress("0x00aed8")),
			HVal: hValFromAccount(Account{
				Nonce:          65,
				Balance:        big.NewInt(835),
				StorageRoot:    MustHexToKoalabearOctuplet("0x1c41acc261451aae253f621857172d6339919d18059f35921a50aafc69eb5c39"),
				LineaCodeHash:  MustHexToKoalabearOctuplet("0x7b688b215329825e5b00e4aa4e1857bc17afab503a87ecc063614b9b227106b2"),
				KeccakCodeHash: FullBytes32FromHex("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"),
				CodeSize:       0,
			}),
		},
		Tx: types.DynamicFeeTx{
			ChainID:   big.NewInt(59144), // Linea mainnet chain ID
			Nonce:     63,                // invalid nonce
			Value:     big.NewInt(800),   // valid value
			Gas:       1,
			GasFeeCap: big.NewInt(1), // gas price
		},
		FromAddress:    common.HexToAddress("0x00aed8"),
		TxHash:         common.HexToHash("0x5f1d2e2b4c3f4e5d6c7b8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9f"),
		InvalidityType: 0,
	},
}

// hKeyFromAddress computes the HKey using Poseidon2 hash of the address
func hKeyFromAddress(add common.Address) linTypes.KoalaOctuplet {
	hasher := poseidon2_koalabear.NewMDHasher()
	addrBytes := add.Bytes()
	elems := make([]field.Element, 0, 10)
	for i := 0; i < len(addrBytes); i += 2 {
		v := uint64(addrBytes[i])<<8 | uint64(addrBytes[i+1])
		var e field.Element
		e.SetUint64(v)
		elems = append(elems, e)
	}
	hasher.WriteElements(elems...)
	digest := hasher.Sum(nil)
	var d linTypes.KoalaOctuplet
	if err := d.SetBytes(digest); err != nil {
		panic(err)
	}
	return d
}

// hValFromAccount computes the HVal using Poseidon2 hash of the account
func hValFromAccount(a Account) linTypes.KoalaOctuplet {
	hasher := poseidon2_koalabear.NewMDHasher()
	a.WriteTo(hasher)
	digest := hasher.Sum(nil)
	var d linTypes.KoalaOctuplet
	if err := d.SetBytes(digest); err != nil {
		panic(err)
	}
	return d
}
