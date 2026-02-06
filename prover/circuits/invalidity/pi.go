package invalidity

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_bls12377"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
)

// FunctionalPublicInputsGnark represents the gnark version of [public_input.Invalidity]
type FunctionalPublicInputsGnark struct {
	TxHash              [2]frontend.Variable // keccak hash needs 2 field elements (16 bytes each)
	TxNumber            frontend.Variable
	FromAddress         frontend.Variable
	StateRootHash       [2]frontend.Variable // KoalaBear octuplet converted to 2 BLS12-377 field elements (16 bytes each)
	ExpectedBlockNumber frontend.Variable
	FtxRollingHash      frontend.Variable // 32 bytes from mimc_bls12377
}

// Assign the functional public inputs
func (gpi *FunctionalPublicInputsGnark) Assign(pi public_input.Invalidity) {
	gpi.TxHash[0] = pi.TxHash[:16]
	gpi.TxHash[1] = pi.TxHash[16:]
	gpi.FromAddress = pi.FromAddress[:]
	gpi.ExpectedBlockNumber = pi.ExpectedBlockHeight
	gpi.TxNumber = pi.TxNumber

	// Convert octuplet to 32 bytes, then split into two 16-byte chunks
	stateRootBytes := pi.StateRootHash.ToBytes()
	gpi.StateRootHash[0] = stateRootBytes[:16]
	gpi.StateRootHash[1] = stateRootBytes[16:]

	gpi.FtxRollingHash = pi.FtxRollingHash[:]

}

// Sum computes the hash over the functional inputs using Poseidon2
// The hash is computed over all fields in a specific order matching the native version
func (spi *FunctionalPublicInputsGnark) Sum(api frontend.API) frontend.Variable {

	hsh, err := poseidon2_bls12377.NewGnarkMDHasher(api)
	if err != nil {
		panic(err)
	}

	hsh.Write(
		spi.TxHash[0],
		spi.TxHash[1],
		spi.TxNumber,
		spi.FromAddress,
		spi.ExpectedBlockNumber,
		spi.StateRootHash[0],
		spi.StateRootHash[1],
		spi.FtxRollingHash,
	)

	return hsh.Sum()
}

func addressToBytes20(api frontend.API, addr frontend.Variable) []frontend.Variable {
	const (
		addrBits  = 160
		chunkBits = 8
		chunks    = addrBits / chunkBits
	)
	bits := api.ToBinary(addr, addrBits)
	res := make([]frontend.Variable, chunks)
	for i := 0; i < chunks; i++ {
		start := (chunks - 1 - i) * chunkBits
		end := start + chunkBits
		res[i] = api.FromBinary(bits[start:end]...)
	}
	return res
}
