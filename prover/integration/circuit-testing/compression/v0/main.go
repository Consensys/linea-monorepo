package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"os"

	"github.com/consensys/gnark/test"
	"github.com/consensys/zkevm-monorepo/prover/circuits/blobdecompression"
	v0 "github.com/consensys/zkevm-monorepo/prover/circuits/blobdecompression/v0"
	blob "github.com/consensys/zkevm-monorepo/prover/lib/compressor/blob/v0"

	"github.com/consensys/compress/lzss"
	"github.com/consensys/gnark-crypto/ecc"
	fr381 "github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/consensys/zkevm-monorepo/prover/backend/blobsubmission"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

const (
	blobDataPath = "./integration/circuit-testing/compression/v0/blob.json"
	dictPath     = "./lib/compressor/compressor_dict.bin"
)

// This function generates a compressed blob
func main() {

	resp, blobBytes, dict := mustGetTestCompressedData()
	circ := v0.Allocate(dict)

	logrus.Infof("Building the constraint system")

	logrus.Infof("Creating an assignment")

	var x, y fr381.Element

	if b, err := utils.HexDecodeString(resp.ExpectedX); err != nil {
		panic(err)
	} else {
		x.SetBytes(b)
	}

	if b, err := utils.HexDecodeString(resp.ExpectedY); err != nil {
		panic(err)
	} else {
		y.SetBytes(b)
	}

	givenSnarkHash, err := utils.HexDecodeString(resp.SnarkHash)
	if err != nil {
		panic(err)
	}

	a, _, snarkHash, err := blobdecompression.Assign(blobBytes, dict, true, x, y)
	if err != nil {
		panic(err)
	}
	if _, ok := a.(*v0.Circuit); !ok {
		utils.Panic("unexpected circuit type")
	}

	if !bytes.Equal(snarkHash[:], givenSnarkHash) {
		utils.Panic("unexpected snark hash")
	}

	blobPackedBytes := a.(*v0.Circuit).BlobPackedBytes
	if len(blobPackedBytes) != len(circ.BlobPackedBytes) {
		utils.Panic("inconsistent size for the compressed bytes: %v != %v", len(blobPackedBytes), len(circ.BlobPackedBytes))
	}

	logrus.Infof("Calling the solver")

	if err = test.IsSolved(&circ, a, ecc.BLS12_377.ScalarField()); err != nil {
		panic(err)
	}
}

// mustGetTestCompressedData is a test utility function that we use to get
// actual compressed data from the
func mustGetTestCompressedData() (resp blobsubmission.Response, blobBytes, dict []byte) {
	dict, err := os.ReadFile(dictPath)
	if err != nil {
		panic(err)
	}
	dict = lzss.AugmentDict(dict)

	respJson, err := os.ReadFile(blobDataPath) // TODO read json
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
