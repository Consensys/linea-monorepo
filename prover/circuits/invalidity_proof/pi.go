package invalidity_proof

import (
	"github.com/consensys/gnark/frontend"
	gnarkHash "github.com/consensys/gnark/std/hash"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
)

// FunctionalPublicInputsGnark represents the gnark version of [public_input.Invalidity]
type FunctionalPublicInputsGnark struct {
	TxHashMSB            frontend.Variable
	TxHashLSB            frontend.Variable
	FromAddress          frontend.Variable
	BlockHeight          frontend.Variable
	InitialStateRootHash frontend.Variable
	TimeStamp            frontend.Variable
}

// Assign the functional public inputs
func (gpi *FunctionalPublicInputsGnark) Assign(pi public_input.Invalidity) {
	gpi.TxHashMSB = pi.TxHash[:16]
	gpi.TxHashLSB = pi.TxHash[16:]
	gpi.FromAddress = pi.FromAddress[:]
	gpi.BlockHeight = pi.BlockHeight
	gpi.InitialStateRootHash = pi.InitialStateRootHash[:]
	gpi.TimeStamp = pi.TimeStamp
}

// Sum computes the hash over the functional inputs
func (spi *FunctionalPublicInputsGnark) Sum(api frontend.API, hsh gnarkHash.FieldHasher) frontend.Variable {

	hsh.Reset()
	hsh.Write(
		spi.TxHashMSB,
		spi.TxHashLSB,
		spi.FromAddress,
		spi.BlockHeight,
		spi.InitialStateRootHash,
		spi.TimeStamp,
	)

	return hsh.Sum()
}
