package public_input

import (
	"hash"
	"log"

	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
)

// Invalidity represents the functional public inputs for the invalidity circuit
// The mimc hash over functional inputs is set as the public input of the circuit.
type Invalidity struct {
	TxHash              common.Hash // hash of the transaction
	TxNumber            uint64
	FromAddress         types.EthAddress // address of the sender
	ExpectedBlockHeight uint64           //  the max expected block number for the transaction to be executed.
	StateRootHash       types.Bytes32    // state-root-hash on which the invalidity is based
	FtxRollingHash      types.Bytes32    // the streamHash of the forced transaction
}

// Sum compute the mimc hash over the functional public inputs
func (pi *Invalidity) Sum(hsh hash.Hash) []byte {
	if hsh == nil {
		hsh = mimc.NewMiMC()
	}

	hsh.Reset()
	hsh.Write(pi.TxHash[:16])
	hsh.Write(pi.TxHash[16:])
	types.WriteInt64On32Bytes(hsh, int64(pi.TxNumber))
	hsh.Write(pi.FromAddress[:])
	types.WriteInt64On32Bytes(hsh, int64(pi.ExpectedBlockHeight))
	hsh.Write(pi.StateRootHash[:])
	hsh.Write(pi.FtxRollingHash[:])

	return hsh.Sum(nil)
}

func (pi *Invalidity) SumAsField() field.Element {

	var (
		sumBytes = pi.Sum(nil)
		sum      = new(field.Element).SetBytes(sumBytes)
	)

	return *sum
}

// TxAbiEncode encode the payload of the transaction in a ABI way.
func TxAbiEncode(tx *ethTypes.Transaction) []byte {
	bigInt, err := abi.NewType("uint256", "", nil)
	if err != nil {
		log.Fatal(err)
	}
	abiUint64, err := abi.NewType("uint64", "", nil)
	if err != nil {
		log.Fatal(err)
	}
	bytes, err := abi.NewType("bytes", "", nil)
	if err != nil {
		log.Fatal(err)
	}
	address, err := abi.NewType("address", "", nil)
	if err != nil {
		log.Fatal(err)
	}

	// Arguments type for the AccessList (dynamic array of tuple)
	accessListType, err := abi.NewType("tuple[]", "", []abi.ArgumentMarshaling{
		{Name: "address", Type: "address"},
		{Name: "storageKeys", Type: "bytes32[]"},
	})
	if err != nil {
		log.Fatal(err)
	}

	myAbi := abi.Arguments{
		{Type: bigInt},
		{Type: abiUint64},
		{Type: bigInt},
		{Type: bigInt},
		{Type: abiUint64},
		{Type: address},
		{Type: bigInt},
		{Type: bytes},
		{Type: accessListType},
	}
	txEncoded, err := myAbi.Pack(tx.ChainId(), tx.Nonce(), tx.GasTipCap(),
		tx.GasFeeCap(), tx.Gas(), tx.To(), tx.Value(), tx.Data(), tx.AccessList())

	if err != nil {
		log.Fatal(err)
	}
	return txEncoded
}
