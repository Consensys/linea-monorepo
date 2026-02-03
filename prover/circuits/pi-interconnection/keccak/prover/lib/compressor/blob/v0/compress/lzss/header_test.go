package lzss

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHeaderRoundTrip(t *testing.T) {
	assert := require.New(t)
	h := Header{
		Version: Version,
		Level:   BestCompression,
	}

	var buf bytes.Buffer
	_, err := h.WriteTo(&buf)
	assert.NoError(err)

	var h2 Header
	_, err = h2.ReadFrom(&buf)
	assert.NoError(err)

	assert.Equal(h, h2)
}
