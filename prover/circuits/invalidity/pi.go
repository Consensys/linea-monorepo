package invalidity

import (
	"github.com/consensys/gnark/frontend"
	gnarkHash "github.com/consensys/gnark/std/hash"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
)

// FunctionalPublicInputsGnark represents the gnark version of [public_input.Invalidity]
type FunctionalPublicInputsGnark struct {
	TxHash              [2]frontend.Variable // keccak hash needs 2 field elements
	TxNumber            frontend.Variable
	FromAddress         frontend.Variable
	StateRootHash       frontend.Variable
	ExpectedBlockNumber frontend.Variable
	FtxRollingHash      frontend.Variable
}

// Assign the functional public inputs
func (gpi *FunctionalPublicInputsGnark) Assign(pi public_input.Invalidity) {
	gpi.TxHash[0] = pi.TxHash[:16]
	gpi.TxHash[1] = pi.TxHash[16:]
	gpi.FromAddress = pi.FromAddress[:]
	gpi.ExpectedBlockNumber = pi.ExpectedBlockHeight
	gpi.StateRootHash = pi.StateRootHash[:]
	gpi.TxNumber = pi.TxNumber
	gpi.FtxRollingHash = pi.FtxRollingHash[:]
}

// Sum computes the hash over the functional inputs
func (spi *FunctionalPublicInputsGnark) Sum(api frontend.API, hsh gnarkHash.FieldHasher) frontend.Variable {

	hsh.Reset()
	hsh.Write(
		spi.TxHash[0],
		spi.TxHash[1],
		spi.TxNumber,
		spi.FromAddress,
		spi.ExpectedBlockNumber,
		spi.StateRootHash,
		spi.FtxRollingHash,
	)

	return hsh.Sum()
}

func (f FunctionalPublicInputsGnark) ExecutionCtxFor(c SubCircuit) []frontend.Variable {
	switch c.(type) {
	case *BadNonceBalanceCircuit:
		return []frontend.Variable{f.StateRootHash}
	default:
		panic("unknown or unsupported subcircuit")
	}
}
