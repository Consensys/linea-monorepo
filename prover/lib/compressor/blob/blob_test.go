package blob_test

import (
	"bytes"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/dictionary"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/encode"
	v0 "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v0"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/require"
	"os"
	"testing"

	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob"
	blobv1testing "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestGetVersion(t *testing.T) {
	_blob := blobv1testing.GenTestBlob(t, 1)
	assert.Equal(t, uint32(0x10000), uint32(0xffff)+uint32(blob.GetVersion(_blob)), "version should match the current one")
}

// Once a version of the compressor is released, the encoding/compression scheme must not change
// So we should be able to take a blob from the wild, read it, re-(encode/compress) it and get
// the same blob back
func TestBlobRoundTripV0(t *testing.T) {
	const dictPath = "../compressor_dict.bin"
	dictStore := dictionary.NewStore()
	require.NoError(t, dictStore.Load(dictPath))
	blobData := readHexFile(t, "testdata/sample-blob-01b9918c3f0ceb6a.hex")
	header, _, blocksSerialized, err := v0.DecompressBlob(blobData, dictStore)
	require.NoError(t, err)
	bm, err := v0.NewBlobMaker(v0.MaxUncompressedBytes, "../compressor_dict.bin")
	require.NoError(t, err)
	for i := 0; i < header.NbBatches(); i++ {
		for j := 0; j < header.NbBlocksInBatch(i); j++ {
			dbd, err := v0.DecodeBlockFromUncompressed(bytes.NewReader(blocksSerialized[0]))
			assert.NoError(t, err)

			stdBlockRlp, err := rlp.EncodeToBytes(dbd.ToStd())

			ok, err := bm.Write(stdBlockRlp, false, encode.WithTxAddressGetter(encode.GetAddressFromR))
			assert.NoError(t, err)
			assert.True(t, ok)
			blocksSerialized = blocksSerialized[1:]
		}
		bm.StartNewBatch()
	}
	assert.Empty(t, blocksSerialized)
	test_utils.AssertBytesEqual(t, bm.Bytes(), blobData)
}

func readHexFile(t *testing.T, filename string) []byte {
	hex, err := os.ReadFile(filename)
	require.NoError(t, err)
	res, err := utils.HexDecodeString(string(hex))
	require.NoError(t, err)
	return res
}
