package invalidity_test

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/backend/ethereum"
	"github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	linTypes "github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func TestInvalidity(t *testing.T) {
	// t.Skip("skipping invalidity tests until we add the missing constraints for consistency between nonce and rlpencoding")
	const maxRlpByteSize = 1024
	var (
		config = &smt.Config{
			HashFunc: hashtypes.MiMC,
			Depth:    10,
		}
		tree, _, _ = genShomei(t, tcases, config)
		root       = tree.Root
	)
	for pos := range tcases {
		var (
			proof, _ = tree.Prove(pos)
			leaf, _  = tree.GetLeaf(pos)
			tcase    = tcases[pos]

			assi = invalidity.AssigningInputs{
				AccountTrieInputs: invalidity.AccountTrieInputs{
					Proof:       proof,
					Leaf:        leaf,
					Root:        root,
					Config:      config,
					Account:     tcase.Account,
					LeafOpening: tcase.Leaf,
				},
				Transaction:    types.NewTx(&tcase.Tx),
				FromAddress:    tcase.FromAddress,
				MaxRlpByteSize: maxRlpByteSize,
				InvalidityType: tcase.InvalidityType,
			}

			b = ethereum.EncodeTxForSigning(assi.Transaction)
		)

		// RLP encode the transaction
		assi.RlpEncodedTx = make([]byte, len(b[:])) // include the type byte
		copy(assi.RlpEncodedTx, b[:])

		assi.FuncInputs = public_input.Invalidity{
			StateRootHash: root,
			TxHash:        common.Hash(crypto.Keccak256(assi.RlpEncodedTx)),
			FromAddress:   linTypes.EthAddress(assi.FromAddress),
		}

		// generate keccak proof for the circuit
		kcomp, kproof := invalidity.MakeKeccakProofs(assi.Transaction, maxRlpByteSize, dummy.Compile)
		assi.KeccakCompiledIOP = kcomp
		assi.KeccakProof = kproof

		// define the circuit
		circuit := invalidity.CircuitInvalidity{
			SubCircuit: &invalidity.BadNonceBalanceCircuit{},
		}

		// allocate the circuit
		circuit.Allocate(invalidity.Config{
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
		assignment := invalidity.CircuitInvalidity{
			SubCircuit: &invalidity.BadNonceBalanceCircuit{},
		}
		assignment.Assign(assi)

		witness, err := frontend.NewWitness(&assignment, ecc.BLS12_377.ScalarField())
		require.NoError(t, err)

		err = scs.IsSolved(witness)
		require.NoError(t, err)
	}

}
