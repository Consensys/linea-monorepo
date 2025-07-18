package invalidity_proof

import (
	"math/big"

	"github.com/consensys/gnark/frontend"
	gnarkHash "github.com/consensys/gnark/std/hash"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	"github.com/consensys/linea-monorepo/prover/utils/types"
)

// FunctionalPublicInputsGnark represents the gnark version of [public_input.Invalidity]
type FunctionalPublicInputsGnark struct {
	TxHash                       frontend.Variable
	FromAddress                  frontend.Variable
	FunctionalPublicInputsQGnark FunctionalPublicInputsQGnark // the FPI participation in the interconnection.
}

// Assign the functional public inputs
func (gpi *FunctionalPublicInputsGnark) Assign(pi public_input.Invalidity) {
	gpi.TxHash = pi.TxHash[:]
	gpi.FromAddress = pi.FromAddress[:]
	gpi.FunctionalPublicInputsQGnark.BlockNumber = pi.BlockHeight
	gpi.FunctionalPublicInputsQGnark.SateRootHash = pi.StateRootHash[:]
}

// Sum computes the hash over the functional inputs
func (spi *FunctionalPublicInputsGnark) Sum(api frontend.API, hsh gnarkHash.FieldHasher) frontend.Variable {

	hsh.Reset()
	hsh.Write(
		spi.TxHash,
		spi.FromAddress,
		spi.FunctionalPublicInputsQGnark.BlockNumber,
		spi.FunctionalPublicInputsQGnark.SateRootHash,
	)

	return hsh.Sum()
}

// FunctionalPublicInputsQ represents the functional public-inputs  participating in the interconnection.
type FunctionalPublicInputsQ struct {
	SateRootHash types.Bytes32
	BlockNumber  big.Int
}

// FunctionalPublicInputsQGnark represents [FunctionalPublicInputsQ] in the gnark circuit.
type FunctionalPublicInputsQGnark struct {
	SateRootHash frontend.Variable
	BlockNumber  frontend.Variable
}
