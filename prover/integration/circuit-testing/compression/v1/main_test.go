package main

import (
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/test"
	blob "github.com/consensys/zkevm-monorepo/prover/lib/compressor/blob/v1"
	"github.com/stretchr/testify/assert"
	"testing"
)

// this tests the consistency between the "snarkHash" computations in blobsubmission.CraftResponse and v1.Assign
func TestPrepareEmpty(t *testing.T) {
	prepare(t, blob.EmptyBlob(t))
}

func TestSmallBlob(t *testing.T) {
	c, a := prepareTestBlob(t)
	assert.NoError(t, test.IsSolved(c, a, ecc.BLS12_377.ScalarField()))
}

func TestEmptyBlob(t *testing.T) {
	c, a := prepare(t, blob.EmptyBlob(t))
	assert.NoError(t, test.IsSolved(c, a, ecc.BLS12_377.ScalarField()))
}

func TestTinyTwoBatchBlob(t *testing.T) {
	c, a := prepare(t, blob.TinyTwoBatchBlob(t))
	assert.NoError(t, test.IsSolved(c, a, ecc.BLS12_377.ScalarField()))
}
