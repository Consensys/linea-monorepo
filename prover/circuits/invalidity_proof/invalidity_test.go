package invalidity_proof_test

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	badnonce "github.com/consensys/linea-monorepo/prover/circuits/invalidity_proof"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestInvalidity(t *testing.T) {

	var (
		config = &smt.Config{
			HashFunc: hashtypes.MiMC,
			Depth:    10,
		}

		tree, _, _ = genShomei(t, tcases, config)
		tx         = types.DynamicFeeTx{
			Nonce: uint64(tcases[1].TxNonce),
		}

		assi = badnonce.AssigningInputs{
			Tree:        tree,
			Pos:         1,
			Account:     tcases[1].Account,
			LeafOpening: tcases[1].Leaf,
			Transaction: types.NewTx(&tx),
		}

		circuit badnonce.BadNonceCircuit
	)

	// assign the circuit
	circuitAssi := circuit.Assign(assi)

	// allocate the circuit
	circuit.Allocate(badnonce.Config{
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
	witness, err := frontend.NewWitness(&circuitAssi, ecc.BLS12_377.ScalarField())
	require.NoError(t, err)

	err = scs.IsSolved(witness)
	require.NoError(t, err)

}
