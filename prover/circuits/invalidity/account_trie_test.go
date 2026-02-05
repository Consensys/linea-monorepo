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
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

// TestAccountTrie tests the full AccountTrie circuit which verifies:
// 1. Hash(Account) == LeafOpening.HVal
// 2. Hash(LeafOpening) == MerkleProof.Leaf
// 3. MerkleProof is valid for the given root
func TestAccountTrie(t *testing.T) {
	const depth = 10

	// Create test accounts
	accounts := []types.Account{
		{
			Nonce:          1,
			Balance:        big.NewInt(1000),
			StorageRoot:    types.MustHexToKoalabearOctuplet("0x0b1dfeef3db4956540da8a5f785917ef1ba432e521368da60a0a1ce430425666"),
			LineaCodeHash:  types.MustHexToKoalabearOctuplet("0x729aac4455d43f2c69e53bb75f8430193332a4c32cafd9995312fa8346929e73"),
			KeccakCodeHash: types.FullBytes32FromHex("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"),
			CodeSize:       0,
		},
		{
			Nonce:          65,
			Balance:        big.NewInt(5690),
			StorageRoot:    types.MustHexToKoalabearOctuplet("0x1c41acc261451aae253f621857172d6339919d18059f35921a50aafc69eb5c39"),
			LineaCodeHash:  types.MustHexToKoalabearOctuplet("0x7b688b215329825e5b00e4aa4e1857bc17afab503a87ecc063614b9b227106b2"),
			KeccakCodeHash: types.FullBytes32FromHex("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"),
			CodeSize:       100,
		},
	}

	// Create leaf openings with account hashes
	addresses := []common.Address{
		common.HexToAddress("0x00aed6"),
		common.HexToAddress("0x00aed7"),
	}

	leafOpenings := make([]accumulator.LeafOpening, len(accounts))
	leafHashes := make([]field.Octuplet, len(accounts))

	for i, account := range accounts {
		// Hash the account
		accountHash := hashAccount(&account)

		// Hash the address to get HKey
		hKey := hashAddress(addresses[i])

		// Create leaf opening
		leafOpenings[i] = accumulator.LeafOpening{
			Prev: int64(i),
			Next: int64(i + 1),
			HKey: hKey,
			HVal: accountHash,
		}

		// Hash the leaf opening
		leafHashes[i] = leafOpenings[i].Hash().ToOctuplet()
	}

	// Build the Merkle tree
	tree := smt_koalabear.NewEmptyTree(depth)
	for i, leafHash := range leafHashes {
		tree.Update(i, leafHash)
	}

	// Test each account
	for i := range accounts {
		t.Run(addresses[i].Hex(), func(t *testing.T) {
			// Get proof for this leaf
			proof, err := tree.Prove(i)
			require.NoError(t, err)

			// Verify proof natively first
			leaf, err := tree.GetLeaf(i)
			require.NoError(t, err)
			err = smt_koalabear.Verify(&proof, leaf, tree.Root)
			require.NoError(t, err, "native Merkle proof verification failed")

			// Create circuit inputs
			inputs := invalidity.AccountTrieInputs{
				Account:     accounts[i],
				LeafOpening: leafOpenings[i],
				Leaf:        leafHashes[i],
				Proof:       proof,
				Root:        tree.Root,
			}

			// Create and allocate circuit
			circuit := &invalidity.AccountTrie{}
			circuit.Allocate(invalidity.Config{Depth: depth})

			// Create and assign witness
			witness := &invalidity.AccountTrie{}
			witness.Assign(inputs)

			// Compile circuit
			ccs, err := frontend.CompileU32(
				koalabear.Modulus(),
				scs.NewBuilder,
				circuit,
				frontend.IgnoreUnconstrainedInputs(),
			)
			require.NoError(t, err)

			// Create witness and verify
			wit, err := frontend.NewWitness(witness, koalabear.Modulus())
			require.NoError(t, err)

			err = ccs.IsSolved(wit)
			require.NoError(t, err)
		})
	}
}

// hashAccount hashes an account using Poseidon2
func hashAccount(a *types.Account) types.KoalaOctuplet {
	hasher := poseidon2_koalabear.NewMDHasher()
	a.WriteTo(hasher)
	digest := hasher.Sum(nil)
	var d types.KoalaOctuplet
	if err := d.SetBytes(digest); err != nil {
		panic(err)
	}
	return d
}

// hashAddress hashes an Ethereum address using Poseidon2
func hashAddress(addr common.Address) types.KoalaOctuplet {
	hasher := poseidon2_koalabear.NewMDHasher()
	hasher.Write(addr.Bytes())
	digest := hasher.Sum(nil)
	var d types.KoalaOctuplet
	if err := d.SetBytes(digest); err != nil {
		panic(err)
	}
	return d
}
