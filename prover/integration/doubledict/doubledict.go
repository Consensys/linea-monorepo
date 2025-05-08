package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"os"

	"github.com/consensys/linea-monorepo/prover/backend/blobsubmission"
	"github.com/consensys/linea-monorepo/prover/cmd/prover/cmd"
	blob "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

func main() {
	//makeBlobProofRequests()
	setup()
	//prove()
}


func setup() {
	assertNoError(cmd.Setup(context.TODO(), cmd.SetupArgs{
		Circuits:   "aggregation",
		DictPath:   "",
		DictSize:   65536,
		AssetsDir:  "/home/ubuntu/linea-monorepo/prover/prover-assets",
		ConfigFile: configFile,
	}))
}

func prove() {
	assertNoError(cmd.Prove(cmd.ProverArgs{
		Input:      "/home/ubuntu/linea-monorepo/prover/integration/doubledict/11383347-11384215-bcv0.0-ccv0.0-7f7b2c7fcdd136111a6acdaf9c34c278dc56e306365b8141d55e0f8fd9f418cb-getZkBlobCompressionProof.json",
		Output:     "",
		ConfigFile: configFile,
	}))
}

const (
	oldDict  = "../../lib/compressor/compressor_dict.bin"
	newDict  = "../../lib/compressor/dict/25-04-21.bin"
	hexOne   = "0x0000000000000000000000000000000000000000000000000000000000000001"
	hexTwo   = "0x0000000000000000000000000000000000000000000000000000000000000002"
	hexThree = "0x0000000000000000000000000000000000000000000000000000000000000003"
)

func makeBlobProofRequests() {
	req := blobProofRequest(hexOne, hexTwo, hexThree, oldDict)
	saveJson(req, "getZkBlobCompressionProof-olddict.json")

	req = blobProofRequest(req.FinalStateRootHash, hexThree, req.ExpectedShnarf, newDict)
	saveJson(req, "getZkBlobCompressionProof-newdict.json")
}

func Close(f interface{ Close() error }) {
	assertNoError(f.Close())
}

func saveJson(o any, path string) {
	f, err := os.Create(path)
	assertNoError(err)
	defer Close(f)

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	assertNoError(enc.Encode(o))
}

func blobProofRequest(parentStateRootHash, stateRootHash, prevShnarf, dictPath string) *blobsubmission.Response {
	res, err := blobsubmission.CraftResponse(&blobsubmission.Request{
		Eip4844Enabled:      true,
		CompressedData:      base64.StdEncoding.EncodeToString(createBlob(dictPath)),
		DataParentHash:      "0x1234567890123456789012345678901234567890123456789012345678901234",
		ConflationOrder:     blobsubmission.ConflationOrder{},
		ParentStateRootHash: parentStateRootHash,
		FinalStateRootHash:  stateRootHash,
		PrevShnarf:          prevShnarf,
	})
	assertNoError(err)
	return res
}

func assertNoError(err error) {
	if err != nil {
		panic(err)
	}
}

func createBlob(dictPath string) []byte {
	const blocksFilePath = "18528963-18529608.blocks"
	bm, err := blob.NewBlobMaker(blob.MaxUsableBytes, dictPath)
	assertNoError(err)
	in, err := os.ReadFile(blocksFilePath)
	assertNoError(err)

	var (
		b  ethtypes.Block
		bb bytes.Buffer
	)

	for {
		bb.Reset()
		assertNoError(rlp.Decode(bytes.NewReader(in), &b))
		assertNoError(rlp.Encode(&bb, &b))
		in = in[bb.Len():]

		ok, err := bm.Write(bb.Bytes(), false)
		assertNoError(err)
		if !ok { // blob is full
			break
		}
	}
	return bm.Bytes()
}
