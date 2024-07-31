package sha2

import (
	"crypto/sha256"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	ExpectedHash Digest
	Stream       []byte
}

func TestHash(t *testing.T) {

	var (
		maxSizeByte = 1000
		// #nosec G404 -- we don't need a cryptographic PRNG for testing purposes
		rng = rand.New(rand.NewSource(212678))
	)

	for sizeByte := 0; sizeByte < maxSizeByte; sizeByte++ {

		var (
			testCase                = genTestCase(rng, sizeByte)
			recoveredHashWoTraces   = Hash(testCase.Stream, nil)
			recoveredHashWithTraces = Hash(testCase.Stream, &HashTraces{})
		)

		assert.Equalf(t, testCase.ExpectedHash, recoveredHashWoTraces, "(without trace) for input of size: %v", sizeByte)
		assert.Equalf(t, testCase.ExpectedHash, recoveredHashWithTraces, "(with trace) for input of size: %v", sizeByte)
	}
}

func genTestCase(rng *rand.Rand, sizeByte int) testCase {

	stream := make([]byte, sizeByte)
	rng.Read(stream)

	return testCase{
		Stream:       stream,
		ExpectedHash: sha256.Sum256(stream),
	}
}
