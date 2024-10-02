package blob_test

import (
	"bytes"
	"fmt"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/dictionary"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/encode"
	v0 "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v0"
	blobv1testing "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1/test_utils"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestGetVersion(t *testing.T) {
	_blob := blobv1testing.GenTestBlob(t, 1)
	assert.Equal(t, uint32(0x10000), uint32(0xffff)+uint32(blob.GetVersion(_blob)), "version should match the current one")
}

const dictPath = "../compressor_dict.bin"

func TestAddToBlob(t *testing.T) {
	dictStore := dictionary.NewStore()
	require.NoError(t, dictStore.Load(dictPath))
	blobData := withNoError(t, os.ReadFile, "testdata/v0/sample-blob-01b9918c3f0ceb6a.bin")
	header, _, blocksSerialized, err := v0.DecompressBlob(blobData, dictStore)
	require.NoError(t, err)

	//blobData = withNoError(t, os.ReadFile, "testdata/v0/sample-blob-0151eda71505187b5.bin")
	//_, _, blocksSerializedNext, err := v0.DecompressBlob(blobData, dictStore)

	blocksSerialized = append(blocksSerialized, blocksSerialized...)

	bm, err := v0.NewBlobMaker(v0.MaxUncompressedBytes, "../compressor_dict.bin")
	require.NoError(t, err)
	var ok bool
	for i := 0; i < header.NbBatches(); i++ {
		for j := 0; j < header.NbBlocksInBatch(i); j++ {
			dbd, err := v0.DecodeBlockFromUncompressed(bytes.NewReader(blocksSerialized[0]))
			assert.NoError(t, err)

			stdBlockRlp, err := rlp.EncodeToBytes(dbd.ToStd())

			ok, err = bm.Write(stdBlockRlp, false, encode.WithTxAddressGetter(encode.GetAddressFromR))
			assert.NoError(t, err)
			assert.True(t, ok)
			blocksSerialized = blocksSerialized[1:]
		}
		bm.StartNewBatch()
	}
	util0 := 100 * bm.Len() / v0.MaxUsableBytes

	for ok { // all in one batch
		ok, err = bm.Write(blocksSerialized[0], false)
		assert.NoError(t, err)
	}

	util1 := 100 * bm.Len() / v0.MaxUsableBytes

	fmt.Println(util0, util1)
}

func withNoError[X, Y any](t *testing.T, f func(X) (Y, error), x X) Y {
	y, err := f(x)
	require.NoError(t, err)
	return y
}
