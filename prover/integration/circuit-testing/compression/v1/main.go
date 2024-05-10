package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/test"
	"github.com/consensys/zkevm-monorepo/prover/backend/blobsubmission"
	"github.com/consensys/zkevm-monorepo/prover/circuits/blobdecompression"
	v1 "github.com/consensys/zkevm-monorepo/prover/circuits/blobdecompression/v1"
	blob "github.com/consensys/zkevm-monorepo/prover/lib/compressor/blob/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	dictPath = "./lib/compressor/compressor_dict.bin"
)

type t struct{}

func (_ t) Errorf(format string, args ...interface{}) {
	panic(fmt.Errorf(format, args...))
}

func (_ t) FailNow() {
	panic("fail now")
}

func main() {
	t := t{}

	c, a := prepareTestBlob(t)
	assert.NoError(t, test.IsSolved(c, a, ecc.BLS12_377.ScalarField()))

	c, a = prepare(t, blob.EmptyBlob(t))
	assert.NoError(t, test.IsSolved(c, a, ecc.BLS12_377.ScalarField()))

	b, err := os.ReadFile("./integration/circuit-testing/compression/v1/tiny-two-batch-blob.bin")
	assert.NoError(t, err)
	c, a = prepare(t, b)
	assert.NoError(t, test.IsSolved(c, a, ecc.BLS12_377.ScalarField()))
}

func prepareTestBlob(t require.TestingT) (c, a frontend.Circuit) {
	return prepare(t, blob.GenTestBlob(t, 1000000))
}

func prepare(t require.TestingT, blobBytes []byte) (c, a frontend.Circuit) {

	resp, err := blobsubmission.CraftResponse(&blobsubmission.Request{
		Eip4844Enabled: true,
		CompressedData: base64.StdEncoding.EncodeToString(blobBytes),
	})
	assert.NoError(t, err)

	b, err := hex.DecodeString(resp.ExpectedX[2:])
	assert.NoError(t, err)
	var x fr.Element
	x.SetBytes(b)

	b, err = hex.DecodeString(resp.ExpectedY[2:])
	assert.NoError(t, err)
	var y fr.Element
	y.SetBytes(b)

	blobBytes = append(blobBytes, make([]byte, blob.MaxUsableBytes-len(blobBytes))...)
	dict := blob.LoadDict(t)
	a, _, snarkHash, err := blobdecompression.Assign(blobBytes, dict, true, x, y)
	assert.NoError(t, err)

	_, ok := a.(*v1.Circuit)
	assert.True(t, ok)

	assert.Equal(t, resp.SnarkHash[2:], hex.EncodeToString(snarkHash))

	return &v1.Circuit{
		Dict:      make([]frontend.Variable, len(dict)),
		BlobBytes: make([]frontend.Variable, blob.MaxUsableBytes),
	}, a
}
