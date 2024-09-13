package blob_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestGetVersion(t *testing.T) {
	_blob := test_utils.GenTestBlob(t, 1)
	assert.Equal(t, uint32(0x10000), uint32(0xffff)+uint32(blob.GetVersion(_blob)), "version should match the current one")
}
