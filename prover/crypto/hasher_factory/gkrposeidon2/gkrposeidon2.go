package gkrposeidon2

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash"
	gkrposeidon2 "github.com/consensys/gnark/std/hash/poseidon2/gkr-poseidon2"
)

type HasherFactory struct {
	api frontend.API
}

func (f *HasherFactory) NewHasher() hash.StateStorer {
	res, err := gkrposeidon2.New(f.api)
	if err != nil {
		panic(err)
	}
	return res
}

func NewHasherFactory(api frontend.API) *HasherFactory {
	return &HasherFactory{api: api}
}
