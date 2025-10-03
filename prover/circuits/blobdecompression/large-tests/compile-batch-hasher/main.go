package main

import (
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/profile"
	v1 "github.com/consensys/linea-monorepo/prover/circuits/blobdecompression/v1"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc/gkrmimc"
	blob "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1"
)

func main() {
	p := profile.Start()
	_, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &circuit{}, frontend.WithCapacity(1<<29))
	if err != nil {
		panic(err)
	}
	p.Stop()
}

type circuit struct {
	NbBatches    T
	BlobPayload  [blob.MaxUncompressedBytes]T
	BatchEnds    [v1.MaxNbBatches]T
	ExpectedSums [v1.MaxNbBatches]T
}

func (c *circuit) Define(api frontend.API) error {
	hsh := gkrmimc.NewHasherFactory(api).NewHasher()
	return v1.CheckBatchesSums(api, hsh, c.NbBatches, c.BlobPayload[:], c.BatchEnds[:], c.ExpectedSums[:])
}
