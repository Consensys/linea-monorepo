package badnonce

import (
	"reflect"

	"github.com/consensys/gnark/frontend"
	gmimc "github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
)

type MerkleProofCircuit struct {
	Proofs smt.GnarkProof    `gnark:",public"`
	Leaf   frontend.Variable `gnark:",public"`
	Root   frontend.Variable
}

type BadNonceCircuit struct {
	// Nonce for the target FTx
	FTxNonce frontend.Variable
	// merkle proof for the account nonce
	MerkleTree  MerkleProofCircuit
	LeafOpening GnarkLeafOpening
	Account     GnarkAccount
}

func (circuit *BadNonceCircuit) Define(api frontend.API) error {

	// check that the FTx.Nonce = Account.Nonce + 1
	res := api.Sub(circuit.FTxNonce, api.Add(circuit.Account.Nonce, 1))
	api.AssertIsDifferent(res, 0)

	// Hash (Account) == LeafOpening.HVal
	hashAccount := MimcCircuit{
		PreImage: sliceFromStruct(circuit.Account),
		Hash:     circuit.LeafOpening.HVal,
	}

	err := hashAccount.Define(api)
	if err != nil {
		return err
	}

	// Hash(LeafOpening)= MerkleTree.Leaf
	hashLeafOpening := MimcCircuit{
		PreImage: sliceFromStruct(circuit.LeafOpening),
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

func (circuit *MerkleProofCircuit) Define(api frontend.API) error {

	h, err := gmimc.NewMiMC(api)
	if err != nil {
		return err
	}

	smt.GnarkVerifyMerkleProof(api, circuit.Proofs, circuit.Leaf, circuit.Root, &h)

	return nil
}

type GnarkAccount struct {
	Nonce          frontend.Variable
	Balance        frontend.Variable
	StorageRoot    frontend.Variable
	MimcCodeHash   frontend.Variable
	KeccakCodeHash frontend.Variable
	CodeSize       frontend.Variable
}

type GnarkLeafOpening struct {
	Prev frontend.Variable
	Next frontend.Variable
	HKey frontend.Variable
	HVal frontend.Variable
}

// Circuit defines a pre-image knowledge proof
// mimc( preImage) = public hash
type MimcCircuit struct {
	// struct tag on a variable is optional
	// default uses variable name and secret visibility.
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
} // specify constraints

func sliceFromStruct(input interface{}) []frontend.Variable {
	var v []frontend.Variable

	val := reflect.ValueOf(input)

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		v = append(v, field)
	}
	return v
}

// DynamicFeeTx represents an EIP-1559 transaction.
type GnarkTransaction struct {
	ChainID    frontend.Variable
	Nonce      frontend.Variable
	GasTipCap  frontend.Variable // a.k.a. maxPriorityFeePerGas
	GasFeeCap  frontend.Variable // a.k.a. maxFeePerGas
	Gas        frontend.Variable
	To         frontend.Variable `rlp:"nil"` // nil means contract creation
	Value      frontend.Variable
	Data       []frontend.Variable
	AccessList GnarkAccessList
}

type GnarkAccessList []GnarkAccessTuple

type GnarkAccessTuple struct {
	Address     frontend.Variable   `json:"address"     gencodec:"required"`
	StorageKeys []frontend.Variable `json:"storageKeys" gencodec:"required"`
}
