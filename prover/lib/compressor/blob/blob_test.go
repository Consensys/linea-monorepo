package blob

import (
	v1 "github.com/consensys/zkevm-monorepo/prover/lib/compressor/blob/v1"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetVersion(t *testing.T) {
	blob := v1.GenTestBlob(t, 1)
	assert.Equal(t, uint32(0x10000), uint32(0xffff)+uint32(GetVersion(blob)), "version should match the current one")
}
