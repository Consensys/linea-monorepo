package internal

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash"
	gnarkMiMC "github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
)

type HasherWithCompressionFunction interface {
	hash.FieldHasher
	// Compress statelessly applies the hash compression function to the given blocks
	Compress(frontend.API, frontend.Variable, frontend.Variable) frontend.Variable
}

type hasherCompressorWrapper struct {
	hash.FieldHasher
	f func(frontend.API, frontend.Variable, frontend.Variable) frontend.Variable
}

func (h *hasherCompressorWrapper) Compress(api frontend.API, oldState, block frontend.Variable) frontend.Variable {
	return h.f(api, oldState, block)
}

func WrapHasherWithCompressionFunction(h hash.FieldHasher, f func(frontend.API, frontend.Variable, frontend.Variable) frontend.Variable) HasherWithCompressionFunction {
	return &hasherCompressorWrapper{h, f}
}

func NewMiMCWithCompressionFunction(api frontend.API) (HasherWithCompressionFunction, error) {
	h, err := gnarkMiMC.NewMiMC(api)
	if err != nil {
		return nil, err
	}
	return WrapHasherWithCompressionFunction(&h, mimc.GnarkBlockCompression), nil
}
