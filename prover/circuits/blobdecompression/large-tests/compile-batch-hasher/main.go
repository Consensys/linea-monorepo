package main

import (
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/profile"
	v1 "github.com/consensys/linea-monorepo/prover/circuits/blobdecompression/v1"
	"github.com/consensys/linea-monorepo/prover/crypto/hasher_factory/gkrmimc"
	blob "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
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
	NbBatches    zk.WrappedVariable
	BlobPayload  [blob.MaxUncompressedBytes]zk.WrappedVariable
	BatchEnds    [v1.MaxNbBatches]zk.WrappedVariable
	ExpectedSums [v1.MaxNbBatches]zk.WrappedVariable
}

func (c *circuit) Define(api frontend.API) error {
	hsh := gkrmimc.NewHasherFactory(api).NewHasher()
	return v1.CheckBatchesSums(api, hsh, c.NbBatches, c.BlobPayload[:], c.BatchEnds[:], c.ExpectedSums[:])
}
