package badnonce

import (
	"github.com/consensys/gnark/frontend"
	ac "github.com/consensys/linea-monorepo/prover/crypto/state-management/accumulator"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/accumulator"
	"github.com/crate-crypto/go-ipa/bandersnatch/fr"
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
func (circuit BadNonceCircuit) Define(api frontend.API) error {

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

func (c BadNonceCircuit) Allocate(config Config) {
	c.MerkleProof.Proofs.Siblings = make([]frontend.Variable, config.Depth)
}

func (c *BadNonceCircuit) Assign(assi AssigningInputs) BadNonceCircuit {

	// assign the merkle proof
	leaf, _ := assi.Tree.GetLeaf(assi.Pos)
	proof, _ := assi.Tree.Prove(assi.Pos)
	root := assi.Tree.Root

	var witMerkle MerkleProofCircuit
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
	a := assi.Account

	account := types.GnarkAccount{
		Nonce:    a.Nonce,
		Balance:  a.Balance,
		CodeSize: a.CodeSize,
	}

	account.StorageRoot = *buf.SetBytes(a.StorageRoot[:])
	account.MimcCodeHash = *buf.SetBytes(a.MimcCodeHash[:])
	account.KeccakCodeHashMSB = *buf.SetBytes(a.KeccakCodeHash[16:])
	account.KeccakCodeHashLSB = *buf.SetBytes(a.KeccakCodeHash[:16])

	hval := ac.Hash(assi.Tree.Config, a)

	l := assi.LeafOpening
	leafOpening := accumulator.GnarkLeafOpening{
		Prev: l.Prev,
		Next: l.Next,
	}

	leafOpening.HKey = *buf.SetBytes(l.HKey[:])
	leafOpening.HVal = *buf.SetBytes(hval[:])

	res := BadNonceCircuit{
		TxNonce:     assi.Transaction.Nonce(),
		MerkleProof: witMerkle,
		LeafOpening: leafOpening,
		Account:     account,
	}
	return res
}
