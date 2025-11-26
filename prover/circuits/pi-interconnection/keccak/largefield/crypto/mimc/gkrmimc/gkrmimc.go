package gkrmimc

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash"
	gkr_mimc "github.com/consensys/gnark/std/hash/mimc/gkr-mimc"
)

type HasherFactory interface {
	NewHasher() hash.StateStorer
}

type hasherFactory struct {
	api frontend.API
}

func (f *hasherFactory) NewHasher() hash.StateStorer {
	res, err := gkr_mimc.New(f.api)
	if err != nil {
		panic(err)
	}
	return res
}

func NewHasherFactory(api frontend.API) HasherFactory {
	return &hasherFactory{api: api}
}
