package main

import (
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/profile"
	"github.com/consensys/linea-monorepo/prover/circuits/execution"
	blob "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
)

const maxNbBatches = 100

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
	ExpectedSums [maxNbBatches]execution.DataChecksumSnark
}

func (c *circuit) Define(api frontend.API) error {
	return v2.CheckBatchesPartialSums(api, c.NbBatches, c.BlobPayload[:], c.ExpectedSums[:])
}
