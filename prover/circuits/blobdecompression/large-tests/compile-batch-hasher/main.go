package main

import (
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/profile"
	gkr_mimc "github.com/consensys/gnark/std/hash/mimc/gkr-mimc"
	v1 "github.com/consensys/linea-monorepo/prover/circuits/blobdecompression/v1"
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
	NbBatches    frontend.Variable
	BlobPayload  [blob.MaxUncompressedBytes]frontend.Variable
	BatchEnds    [v1.MaxNbBatches]frontend.Variable
	ExpectedSums [v1.MaxNbBatches]frontend.Variable
}

func (c *circuit) Define(api frontend.API) error {
	hsh, err := gkr_mimc.New(api)
	if err != nil {
		return err
	}
	return v1.CheckBatchesSums(api, hsh, c.NbBatches, c.BlobPayload[:], c.BatchEnds[:], c.ExpectedSums[:])
}
