package blob_test

import (
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob"
	blobv1testing "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1/test_utils"
	"github.com/consensys/linea-monorepo/prover/utils"
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

func readHexFile(t *testing.T, filename string) []byte {
	hex, err := os.ReadFile(filename)
	require.NoError(t, err)
	res, err := utils.HexDecodeString(string(hex))
	require.NoError(t, err)
	return res
}
