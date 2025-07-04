package invalidity_proof_test

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	inval "github.com/consensys/linea-monorepo/prover/circuits/invalidity_proof"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestInvalidityBadNonce(t *testing.T) {

	var (
		config = &smt.Config{
			HashFunc: hashtypes.MiMC,
			Depth:    10,
		}

		tree, _, _ = genShomei(t, tcases, config)
		tx         = types.DynamicFeeTx{
			Nonce: uint64(tcases[1].TxNonce),
		}

		assi = inval.AssigningInputs{
			AccountTrieInputs: inval.AccountTrieInputs{
				Tree:        tree,
				Pos:         1,
				Account:     tcases[1].Account,
				LeafOpening: tcases[1].Leaf,
			},
			Transaction: types.NewTx(&tx),
		}

		circuit inval.BadNonceCircuit
	)

	// assign the circuit
	circuit.Assign(assi)

	witness, err := frontend.NewWitness(&circuit, ecc.BLS12_377.ScalarField())
	require.NoError(t, err)

	// allocate the circuit
	circuit.Allocate(inval.Config{
		Depth: config.Depth,
	})

	// compile the circuit
	scs, err := frontend.Compile(
		ecc.BLS12_377.ScalarField(),
		scs.NewBuilder,
		&circuit,
	)
	require.NoError(t, err)

	// solve the circuit
	err = scs.IsSolved(witness)
	require.NoError(t, err)

}
