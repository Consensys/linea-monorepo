package gkrposeidon2

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash"
	gkr_poseidon2 "github.com/consensys/gnark/std/hash/poseidon2/gkr-poseidon2"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/largefield/crypto/poseidon2"
)

type hasherFactory struct {
	api frontend.API
}

func (f *hasherFactory) NewHasher() hash.StateStorer {
	res, err := gkr_poseidon2.New(f.api)
	if err != nil {
		panic(err)
	}
	return res
}

func NewHasherFactory(api frontend.API) poseidon2.HasherFactory {
	return &hasherFactory{api: api}
}
