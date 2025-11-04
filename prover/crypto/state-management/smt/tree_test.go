package smt

import (
	"bytes"
	"math/rand/v2"
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/go-playground/assert/v2"
)

// TestHashNodeLR checks that the hash of a node is the same as the hash of
// the concatenation of the hash of the left and the hash of the right
// child (in bytes).
func TestHashNodeLR(t *testing.T) {

	cfg := &Config{
		HashFunc: poseidon2.Poseidon2,
		Depth:    40,
	}

	rng := rand.New(utils.NewRandSource(0)) // #nosec G404
	l := field.PseudoRandOctuplet(rng)
	r := field.PseudoRandOctuplet(rng)

	buf := &bytes.Buffer{}
	field.WriteOctupletTo(buf, l)
	field.WriteOctupletTo(buf, r)
	lrBytes := buf.Bytes()

	// bytesHasher
	bytesHasher := cfg.HashFunc()
	bytesHasher.Write(lrBytes)
	digestBytesHashing := bytesHasher.Sum(nil)
	digestNodeLR := hashLR(cfg, types.HashToBytes32(l), types.HashToBytes32(r))

	assert.Equal(t, digestBytesHashing, digestNodeLR[:])

}
