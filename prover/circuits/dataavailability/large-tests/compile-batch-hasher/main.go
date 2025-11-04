package main

import (
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/profile"
	"github.com/consensys/linea-monorepo/prover/circuits/dataavailability/v2"
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
	NbBatches    frontend.Variable
	BlobPayload  [blob.MaxUncompressedBytes]frontend.Variable
	BatchEnds    [v2.MaxNbBatches]frontend.Variable
	ExpectedSums [v2.MaxNbBatches]frontend.Variable
}

func (c *circuit) Define(api frontend.API) error {
	hsh := gkrmimc.NewHasherFactory(api).NewHasher()
	return v2.CheckBatchesSums(api, hsh, c.NbBatches, c.BlobPayload[:], c.BatchEnds[:], c.ExpectedSums[:])
}
