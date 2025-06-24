package badnonce

import (
	"reflect"

	"github.com/consensys/gnark/frontend"
	gmimc "github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
)

// BadNonceCircuit define the circuit for the transaction with a bad nonce.
type BadNonceCircuit struct {
	// Transaction Nonce
	TxNonce frontend.Variable
	// Account for the sender of the transaction
	Account GnarkAccount
	// LeafOpening of the Account in the Merkle tree
	LeafOpening GnarkLeafOpening
	// Merkle proof for the LeafOpening
	MerkleTree MerkleProofCircuit
}

func (circuit *BadNonceCircuit) Define(api frontend.API) error {
	// check that the FTx.Nonce = Account.Nonce + 1
	res := api.Sub(circuit.TxNonce, api.Add(circuit.Account.Nonce, 1))
	api.AssertIsDifferent(res, 0)

	// Hash (Account) == LeafOpening.HVal
	hashAccount := MimcCircuit{
		PreImage: SliceFromStruct(circuit.Account),
		Hash:     circuit.LeafOpening.HVal,
	}

	err := hashAccount.Define(api)
	if err != nil {
		return err
	}

	// Hash(LeafOpening)= MerkleTree.Leaf
	hashLeafOpening := MimcCircuit{
		PreImage: SliceFromStruct(circuit.LeafOpening),
		Hash:     circuit.MerkleTree.Leaf,
	}
	err = hashLeafOpening.Define(api)
	if err != nil {
		return err
	}

	// check that MerkleTree.Leaf is compatible with the state
	err = circuit.MerkleTree.Define(api)

	if err != nil {
		return err
	}

	// TBD: check that FTx.Nonce is related to  FTx.Hash  and then in the interconnection we show that
	//FTx.Hash is included in the RollingHash
	return nil
}

// MerkleProofCircuit defines the circuit for validating the Merkle proofs
type MerkleProofCircuit struct {
	Proofs smt.GnarkProof    `gnark:",public"`
	Leaf   frontend.Variable `gnark:",public"`
	Root   frontend.Variable
}

func (circuit *MerkleProofCircuit) Define(api frontend.API) error {
	h, err := gmimc.NewMiMC(api)
	if err != nil {
		return err
	}

	smt.GnarkVerifyMerkleProof(api, circuit.Proofs, circuit.Leaf, circuit.Root, &h)

	return nil
}

// GnarkAccount represent [types.Account] in gnark
type GnarkAccount struct {
	Nonce             frontend.Variable
	Balance           frontend.Variable
	StorageRoot       frontend.Variable
	MimcCodeHash      frontend.Variable
	KeccakCodeHashMSB frontend.Variable
	KeccakCodeHashLSB frontend.Variable
	CodeSize          frontend.Variable
}

// GnarkLeafOpening represent [accumulator.LeafOpening] in gnark
type GnarkLeafOpening struct {
	Prev frontend.Variable
	Next frontend.Variable
	HKey frontend.Variable
	HVal frontend.Variable
}

// Circuit defines a pre-image knowledge proof
// mimc( preImage) = public hash
type MimcCircuit struct {
	PreImage []frontend.Variable
	Hash     frontend.Variable `gnark:",public"`
}

// Define declares the circuit's constraints
// Hash = mimc(PreImage)
func (circuit *MimcCircuit) Define(api frontend.API) error {
	// hash function
	mimc, _ := gmimc.NewMiMC(api)

	// mimc(preImage) == hash
	for _, toHash := range circuit.PreImage {
		mimc.Write(toHash)
	}

	api.AssertIsEqual(circuit.Hash, mimc.Sum())

	return nil
}

// SliceFromStruct creates a slice of [frontend.Variable] from the struct.
func SliceFromStruct(input interface{}) []frontend.Variable {
	var v []frontend.Variable

	val := reflect.ValueOf(input)

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if variable, ok := field.Interface().(frontend.Variable); ok {
			v = append(v, variable)
		} else {
			panic("field is not of type frontend.Variable")
		}
	}
	return v
}
