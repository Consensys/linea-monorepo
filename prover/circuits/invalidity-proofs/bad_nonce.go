package badnonce

import (
	"github.com/consensys/gnark/frontend"
	gmimc "github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/accumulator"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
)

type MerkleProofCircuit struct {
	Proofs smt.GnarkProof    `gnark:",public"`
	Leafs  frontend.Variable `gnark:",public"`
	Root   frontend.Variable
}

type BadNonceCircuit struct {
	// accountNonce from the last execution state
	AccountNonce frontend.Variable
	// Nonce for the target FTx
	FTxNonce frontend.Variable
	// merkle proof for the account nonce
	MerkleTree MerkleProofCircuit
	// leaf correspond with the account nonce
	// LeafOpening LeafOpeningCircuit
	// RLP encoding of the transaction
	// RLPEncodedFTx
}

func (circuit *BadNonceCircuit) Define(api frontend.API) error {

	// check that the FTx.Nonce is not compatible with the account.nonce
	res := api.Sub(circuit.FTxNonce, api.Add(circuit.AccountNonce, 1))
	api.AssertIsDifferent(res, 0)

	// check that the account nonce belongs to the leaf opening
	// circuit.LeafOpeningCircuit.Define()
	// api.AssertIsEqual(circuit.MerkleTree.Leafs)

	// check that leaf is compatible with the state
	err := circuit.MerkleTree.Define(api)

	if err != nil {
		return err
	}

	// check that FTx.Nonce is related to  FTx.Hash  and then in the interconnection we show that
	//FTx.Hash is included in the RollingHash
	return nil
}

func (circuit *MerkleProofCircuit) Define(api frontend.API) error {

	h, err := gmimc.NewMiMC(api)
	if err != nil {
		return err
	}

	smt.GnarkVerifyMerkleProof(api, circuit.Proofs, circuit.Leafs, circuit.Root, &h)

	return nil
}

type LeafOpening accumulator.LeafOpening
