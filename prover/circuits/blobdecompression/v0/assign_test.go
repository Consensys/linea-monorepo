package v0_test

import (
	"encoding/base64"
	"encoding/json"
	v0 "github.com/consensys/zkevm-monorepo/prover/circuits/blobdecompression/v0"
	"github.com/consensys/zkevm-monorepo/prover/lib/compressor/blob/v1/test_utils"
	"os"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	fr381 "github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/consensys/gnark/test"
	"github.com/consensys/zkevm-monorepo/prover/backend/blobsubmission"
	"github.com/consensys/zkevm-monorepo/prover/circuits/blobdecompression"
	blob "github.com/consensys/zkevm-monorepo/prover/lib/compressor/blob/v0"
	"github.com/consensys/zkevm-monorepo/prover/lib/compressor/blob/v0/compress/lzss"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestBlobV0(t *testing.T) {
	resp, blobBytes, dict := mustGetTestCompressedData(t)
	circ := v0.Allocate(dict)

	logrus.Infof("Building the constraint system")

	logrus.Infof("Creating an assignment")

	var (
		x [32]byte
		y fr381.Element
	)

	b, err := utils.HexDecodeString(resp.ExpectedX)
	assert.NoError(t, err)
	copy(x[:], b)

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
func mustGetTestCompressedData(t *testing.T) (resp blobsubmission.Response, blobBytes []byte, dict []byte) {
	dict = lzss.AugmentDict(test_utils.GetDict(t))

	respJson, err := os.ReadFile("sample-blob.json")
	assert.NoError(t, err)

	assert.NoError(t, json.Unmarshal(respJson, &resp))

	blobBytes, err = base64.StdEncoding.DecodeString(resp.CompressedData)
	assert.NoError(t, err)

	_, _, _, err = blob.DecompressBlob(blobBytes, dict)
	assert.NoError(t, err)

	return
}
