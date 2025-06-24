package badnonce_test

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	badnonce "github.com/consensys/linea-monorepo/prover/circuits/invalidity-proofs"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/accumulator"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/stretchr/testify/require"
)

// test [badnonce.BadNonceCircuit]
func TestBadNonce(t *testing.T) {

	// generate witness
	config := &smt.Config{
		HashFunc: hashtypes.MiMC,
		Depth:    10,
	}

	// generate witness
	assignment := genWitness(t, tcases, config)

	// allocate the circuit for merkle tree
	var circuit badnonce.BadNonceCircuit
	lenProof := len(assignment.MerkleTree.Proofs.Siblings)
	circuit.MerkleTree.Proofs.Siblings = make([]frontend.Variable, lenProof)

	// compile the circuit
	scs, err := frontend.Compile(
		ecc.BLS12_377.ScalarField(),
		scs.NewBuilder,
		&circuit,
	)

	require.NoError(t, err)
	// solve the circuit
	witness, err := frontend.NewWitness(&assignment, ecc.BLS12_377.ScalarField())
	require.NoError(t, err)

	err = scs.IsSolved(witness)
	require.NoError(t, err)

}

// generates the witness for [badnonce.BadNonceCircuit]
func genWitness(t *testing.T, tcases []TestCases, config *smt.Config) badnonce.BadNonceCircuit {

	tree, proofs, leaves := genShomei(t, tcases, config)

	// test just for one leaf
	proof := proofs[1]
	leaf := leaves[1]
	root := tree.Root

	//generate merkle tree witMerkle
	var witMerkle badnonce.MerkleProofCircuit
	var buf fr.Element

	witMerkle.Proofs.Siblings = make([]frontend.Variable, len(proof.Siblings))
	for j := 0; j < len(proof.Siblings); j++ {
		buf.SetBytes(proof.Siblings[j][:])
		witMerkle.Proofs.Siblings[j] = buf.String()
	}
	witMerkle.Proofs.Path = proof.Path
	buf.SetBytes(leaf[:])
	witMerkle.Leaf = buf.String()

	buf.SetBytes(root[:])
	witMerkle.Root = buf.String()

	//generate witness for account and leafOpening
	a := tcases[1].Account

	account := badnonce.GnarkAccount{
		Nonce:    a.Nonce,
		Balance:  a.Balance,
		CodeSize: a.CodeSize,
	}

	account.StorageRoot = *buf.SetBytes(a.StorageRoot[:])
	account.MimcCodeHash = *buf.SetBytes(a.MimcCodeHash[:])
	account.KeccakCodeHashMSB = *buf.SetBytes(a.KeccakCodeHash[16:])
	account.KeccakCodeHashLSB = *buf.SetBytes(a.KeccakCodeHash[:16])

	hval := accumulator.Hash(config, a)

	l := tcases[1].Leaf
	leafOpening := badnonce.GnarkLeafOpening{
		Prev: l.Prev,
		Next: l.Next,
	}

	leafOpening.HKey = *buf.SetBytes(l.HKey[:])
	leafOpening.HVal = *buf.SetBytes(hval[:])

	return badnonce.BadNonceCircuit{
		TxNonce:     tcases[1].TxNonce,
		MerkleTree:  witMerkle,
		LeafOpening: leafOpening,
		Account:     account,
	}

}
