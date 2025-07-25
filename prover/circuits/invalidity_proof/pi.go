package invalidity_proof

import (
	"github.com/consensys/gnark/frontend"
	gnarkHash "github.com/consensys/gnark/std/hash"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
)

// FunctionalPublicInputsGnark represents the gnark version of [public_input.Invalidity]
type FunctionalPublicInputsGnark struct {
	TxHash              frontend.Variable
	TxNumber            frontend.Variable
	FromAddress         frontend.Variable
	SateRootHash        frontend.Variable
	ExpectedBlockNumber frontend.Variable
}

// Assign the functional public inputs
func (gpi *FunctionalPublicInputsGnark) Assign(pi public_input.Invalidity) {
	gpi.TxHash = pi.TxHash[:]
	gpi.FromAddress = pi.FromAddress[:]
	gpi.ExpectedBlockNumber = pi.ExpectedBlockHeight
	gpi.SateRootHash = pi.StateRootHash[:]
	gpi.TxNumber = pi.TxNumber
}

// Sum computes the hash over the functional inputs
func (spi *FunctionalPublicInputsGnark) Sum(api frontend.API, hsh gnarkHash.FieldHasher) frontend.Variable {

	hsh.Reset()
	hsh.Write(
		spi.TxHash,
		spi.TxNumber,
		spi.FromAddress,
		spi.ExpectedBlockNumber,
		spi.SateRootHash,
	)

	return hsh.Sum()
}
