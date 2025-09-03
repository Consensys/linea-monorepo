package invalidity_test

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	inval "github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestInvalidityBadBalance(t *testing.T) {

	var (
		config = &smt.Config{
			HashFunc: hashtypes.MiMC,
			Depth:    10,
		}

		pos        = 1
		tree, _, _ = genShomei(t, tcases, config)
		proof, _   = tree.Prove(pos)
		leaf, _    = tree.GetLeaf(pos)
		root       = tree.Root

		assi = inval.AssigningInputs{
			AccountTrieInputs: inval.AccountTrieInputs{
				Proof:       proof,
				Leaf:        leaf,
				Root:        root,
				Config:      config,
				Account:     tcases[1].Account,
				LeafOpening: tcases[1].Leaf,
			},
			Transaction: types.NewTx(&tcases[1].Tx),
			FuncInputs: public_input.Invalidity{
				StateRootHash: root,
			},
		}

		circuit = inval.CircuitInvalidity{
			SubCircuit: &inval.BadBalanceCircuit{},
		}
	)
	// generate keccak proof for the circuit
	kcomp, kproof := inval.MakeKeccakProofs(assi.Transaction, 256, dummy.Compile)
	// assign the circuit
	circuit.Assign(assi, kcomp, kproof)
	// solve the circuit
	witness, err := frontend.NewWitness(&circuit, ecc.BLS12_377.ScalarField())
	require.NoError(t, err)

	// allocate the circuit
	circuit.Allocate(inval.Config{
		Depth: config.Depth,
	}, kcomp)

	// compile the circuit
	scs, err := frontend.Compile(
		ecc.BLS12_377.ScalarField(),
		scs.NewBuilder,
		&circuit,
	)
	require.NoError(t, err)

	err = scs.IsSolved(witness)
	require.NoError(t, err)

}
