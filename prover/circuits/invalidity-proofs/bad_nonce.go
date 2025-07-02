package badnonce

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/accumulator"
)

// BadNonceCircuit defines the circuit for the transaction with a bad nonce.
type BadNonceCircuit struct {
	// Transaction Nonce
	TxNonce frontend.Variable
	// Account for the sender of the transaction
	Account types.GnarkAccount
	// LeafOpening of the Account in the Merkle tree
	LeafOpening accumulator.GnarkLeafOpening
	// Merkle proof for the LeafOpening
	MerkleProof MerkleProofCircuit
}

// Define represent the constraints relevant to [BadNonceCircuit]
func (circuit *BadNonceCircuit) Define(api frontend.API) error {

	var (
		diff             = api.Sub(circuit.TxNonce, api.Add(circuit.Account.Nonce, 1))
		accountSlice     = []frontend.Variable{}
		leafOpeningSlice = []frontend.Variable{}
		account          = circuit.Account
		leafOpening      = circuit.LeafOpening
	)

	// check that the FTx.Nonce = Account.Nonce + 1
	api.AssertIsDifferent(diff, 0)

	// Hash (Account) == LeafOpening.HVal
	hashAccount := MimcCircuit{
		PreImage: append(accountSlice,
			account.Nonce,
			account.Balance,
			account.StorageRoot,
			account.MimcCodeHash,
			account.KeccakCodeHashMSB,
			account.KeccakCodeHashLSB,
			account.CodeSize,
		),
		Hash: circuit.LeafOpening.HVal,
	}
	err := hashAccount.Define(api)
	if err != nil {
		return err
	}

	// Hash(LeafOpening)= MerkleProof.Leaf
	hashLeafOpening := MimcCircuit{
		PreImage: append(leafOpeningSlice,
			leafOpening.Prev,
			leafOpening.Next,
			leafOpening.HKey,
			leafOpening.HVal),

		Hash: circuit.MerkleProof.Leaf,
	}
	err = hashLeafOpening.Define(api)
	if err != nil {
		return err
	}

	// check that MerkleProof.Leaf is compatible with the state
	err = circuit.MerkleProof.Define(api)

	if err != nil {
		return err
	}

	//@azam TBD: check that FTx.Nonce is related to  FTx.Hash  and then in the interconnection we show that
	//FTx.Hash is included in the RollingHash
	return nil
}
