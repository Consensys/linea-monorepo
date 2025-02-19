//go:build !fuzzlight

package v0

import (
	"bytes"
	"math/rand/v2"
	"testing"

	"github.com/consensys/linea-monorepo/prover/utils"

	"github.com/stretchr/testify/require"
)

func TestPack(t *testing.T) {
	assert := require.New(t)
	var buf bytes.Buffer
	var rng = rand.New(rand.NewChaCha8([32]byte{}))

	for i := 0; i < 100; i++ {
		// create 2 random slices
		n1 := rng.IntN(100) + 1 // #nosec G404 -- false positive
		n2 := rng.IntN(100) + 1 // #nosec G404 -- false positive

		s1 := make([]byte, n1)
		s2 := make([]byte, n2)

		utils.ReadPseudoRand(rng, s1)
		utils.ReadPseudoRand(rng, s2)

		// pack them
		buf.Reset()
		written, err := PackAlign(&buf, s1, s2)
		assert.NoError(err, "pack should not generate an error")
		assert.Equal(PackAlignSize(s1, s2), int(written), "written bytes should match expected PackAlignSize")
		original, err := UnpackAlign(buf.Bytes())
		assert.NoError(err, "unpack should not generate an error")

		assert.Equal(s1, original[:n1], "slices should match")
		assert.Equal(s2, original[n1:], "slices should match")
	}
}
