package main

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	fr381 "github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/consensys/gnark/test"
	"github.com/consensys/zkevm-monorepo/prover/backend/blobsubmission"
	"github.com/consensys/zkevm-monorepo/prover/circuits/blobdecompression"
	v0 "github.com/consensys/zkevm-monorepo/prover/circuits/blobdecompression/v0"
	blob "github.com/consensys/zkevm-monorepo/prover/lib/compressor/blob/v0"
	"github.com/consensys/zkevm-monorepo/prover/lib/compressor/blob/v0/compress/lzss"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TODO @tabaie @gbotrel @alexandre.belling decide definitively if we prefer integration tests as tests or main functions
func TestBlobV0(t *testing.T) {
	resp, blobBytes, dict := mustGetTestCompressedData2()
	circ := v0.Allocate(dict)

	logrus.Infof("Building the constraint system")

	logrus.Infof("Creating an assignment")

	var x, y fr381.Element

	b, err := utils.HexDecodeString(resp.ExpectedX)
	assert.NoError(t, err)

	x.SetBytes(b)

	b, err = utils.HexDecodeString(resp.ExpectedY)
	assert.NoError(t, err)
	y.SetBytes(b)

	givenSnarkHash, err := utils.HexDecodeString(resp.SnarkHash)
	assert.NoError(t, err)

	a, _, snarkHash, err := blobdecompression.Assign(blobBytes, dict, true, x, y)
	assert.NoError(t, err)
	_, ok := a.(*v0.Circuit)
	assert.True(t, ok)

	assert.Equal(t, snarkHash[:], givenSnarkHash)

	blobPackedBytes := a.(*v0.Circuit).BlobPackedBytes
	assert.Equal(t, len(blobPackedBytes), len(circ.BlobPackedBytes))

	logrus.Infof("Calling the solver")

	err = test.IsSolved(&circ, a, ecc.BLS12_377.ScalarField())
	assert.NoError(t, err)
}

// mustGetTestCompressedData is a test utility function that we use to get
// actual compressed data from the
func mustGetTestCompressedData2() (resp blobsubmission.Response, blobBytes, dict []byte) {
	dict, err := os.ReadFile("../../../../lib/compressor/compressor_dict.bin")
	if err != nil {
		panic(err)
	}
	dict = lzss.AugmentDict(dict)

	respJson, err := os.ReadFile("../../../../integration/circuit-testing/compression/v0/blob.json") // TODO read json
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(respJson, &resp); err != nil {
		panic(err)
	}

	if blobBytes, err = base64.StdEncoding.DecodeString(resp.CompressedData); err != nil {
		panic(err)
	}

	if _, _, _, err = blob.DecompressBlob(blobBytes, dict); err != nil { // just for sanity checking
		panic(err)
	}

	return

}
