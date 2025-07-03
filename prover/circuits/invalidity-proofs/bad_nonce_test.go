package badnonce_test

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	badnonce "github.com/consensys/linea-monorepo/prover/circuits/invalidity-proofs"
	ac "github.com/consensys/linea-monorepo/prover/crypto/state-management/accumulator"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/accumulator"
	"github.com/stretchr/testify/require"
)

// test [badnonce.BadNonceCircuit]
func TestBadNonce(t *testing.T) {

	// generate witness
	config := &smt.Config{
		HashFunc: hashtypes.MiMC,
		Depth:    10,
	}

	assignment := genWitness(t, tcases, config)

	// allocate the circuit for merkleProof
	var circuit badnonce.BadNonceCircuit
	circuit.MerkleProof.Proofs.Siblings = make([]frontend.Variable, config.Depth)

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

	witMerkle.Proofs.Siblings = make([]frontend.Variable, len(proof.Siblings))
	for j := 0; j < len(proof.Siblings); j++ {

		// witMerkle.Proofs.Siblings[j] = *buf.SetBytes(proof.Siblings[j][:])
		witMerkle.Proofs.Siblings[j] = proof.Siblings[j][:]
	}
	witMerkle.Proofs.Path = proof.Path
	witMerkle.Leaf = leaf[:]

	witMerkle.Root = root[:]

	//generate witness for account and leafOpening
	a := tcases[1].Account

	account := types.GnarkAccount{
		Nonce:    a.Nonce,
		Balance:  a.Balance,
		CodeSize: a.CodeSize,
	}

	account.StorageRoot = a.StorageRoot[:]
	account.MimcCodeHash = a.MimcCodeHash[:]
	account.KeccakCodeHashMSB = a.KeccakCodeHash[16:]
	account.KeccakCodeHashLSB = a.KeccakCodeHash[:16]

	hval := ac.Hash(config, a)

	l := tcases[1].Leaf
	leafOpening := accumulator.GnarkLeafOpening{
		Prev: l.Prev,
		Next: l.Next,
	}

	leafOpening.HKey = l.HKey[:]
	leafOpening.HVal = hval[:]

	return badnonce.BadNonceCircuit{
		TxNonce:     tcases[1].TxNonce,
		MerkleProof: witMerkle,
		LeafOpening: leafOpening,
		Account:     account,
	}

}
