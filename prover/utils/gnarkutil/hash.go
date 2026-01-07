package gnarkutil

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash"
	gkrposeidon2 "github.com/consensys/gnark/std/hash/poseidon2/gkr-poseidon2"
	mimc "github.com/consensys/linea-monorepo/prover/crypto/mimc_bls12377"
)

// GkrPoseidon2HasherFactory returns a hasher factory implementing mimc.HasherFactory
func GkrPoseidon2HasherFactory(api frontend.API) mimc.HasherFactory {
	return gkrPoseidon2HasherFactory{api}
}

// gkrPoseidon2HasherFactory implements mimc.HasherFactory
type gkrPoseidon2HasherFactory struct{ frontend.API }

func (g gkrPoseidon2HasherFactory) NewHasher() hash.StateStorer {
	hsh, err := gkrposeidon2.New(g.API)
	if err != nil {
		panic(err)
	}
	return hsh
}
