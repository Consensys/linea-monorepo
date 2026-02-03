package main

import (
	"fmt"
	"strings"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/profile"
	v1 "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/circuits/blobdecompression/v1"
	blob "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/lib/compressor/blob/v1"
)

func main() {
	c := v1.Circuit{
		BlobBytes:             make([]frontend.Variable, 32*4096),
		Dict:                  make([]frontend.Variable, 64*1024),
		MaxBlobPayloadNbBytes: blob.MaxUncompressedBytes,
	}

	p := profile.Start(profile.WithPath(fmt.Sprintf("e2e-%sK.pprof", formatFloat(blob.MaxUncompressedBytes/1024.0))))

	if _, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &c, frontend.WithCapacity(1<<27)); err != nil {
		panic(err)
	}

	p.Stop()
}

func formatFloat(f float64) string {
	res := fmt.Sprintf("%f", f)
	if !strings.Contains(res, ".") {
		return res
	}
	for res[len(res)-1] == '0' {
		res = res[:len(res)-1]
	}
	if res[len(res)-1] == '.' {
		res = res[:len(res)-1]
	}
	return res
}
