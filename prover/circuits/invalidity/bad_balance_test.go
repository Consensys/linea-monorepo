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
		maxRlpByteSize = 256
		config         = &smt.Config{
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
			FromAddress: tcases[1].FromAddress,
		}
		b = inval.PrefixedRLPNoSignature(assi.Transaction)
	)
	assi.RlpEncodedTx = make([]byte, len(b[:])) // include the type byte
	copy(assi.RlpEncodedTx, b[:])

	// generate keccak proof for the circuit
	kcomp, kproof := inval.MakeKeccakProofs(assi.Transaction, maxRlpByteSize, dummy.Compile)
	assi.KeccakCompiledIOP = kcomp
	assi.KeccakProof = kproof
	assi.MaxRlpByteSize = maxRlpByteSize

	// define the circuit
	circuit := inval.CircuitInvalidity{
		SubCircuit: &inval.BadBalanceCircuit{},
	}

	// allocate the circuit
	circuit.Allocate(inval.Config{
		Depth:             config.Depth,
		KeccakCompiledIOP: kcomp,
		MaxRlpByteSize:    maxRlpByteSize,
	})

	// compile the circuit
	scs, err := frontend.Compile(
		ecc.BLS12_377.ScalarField(),
		scs.NewBuilder,
		&circuit,
	)
	require.NoError(t, err)

	// assign the circuit
	assignment := inval.CircuitInvalidity{
		SubCircuit: &inval.BadBalanceCircuit{},
	}
	assignment.Assign(assi)

	witness, err := frontend.NewWitness(&assignment, ecc.BLS12_377.ScalarField())
	require.NoError(t, err)

	// assert the proof is valid
	err = scs.IsSolved(witness)
	require.NoError(t, err)

}
