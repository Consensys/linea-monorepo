package main

import (
	"math/big"
	"math/rand"
	"testing"
)

func TestInvalidity(t *testing.T) {

	// Create a reproducible RNG
	// #nosec G404 --we don't need a cryptographic RNG for testing purpose
	rng := rand.New(rand.NewSource(seed))

	cfg.AssetsDir = "../../../../prover-assets"

	spec := &InvalidityProofSpec{
		ChainID:             big.NewInt(51),
		ExpectedBlockHeight: 1_000_000_000,
		FtxNumber:           1678,
	}

	ProcessInvaliditySpec(rng, spec, nil, "spec-invalidity.json")

}
