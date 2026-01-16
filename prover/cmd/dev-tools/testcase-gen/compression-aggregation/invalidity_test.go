package main

import (
	"math/big"
	"math/rand"
	"testing"
)

func TestInvalidity(t *testing.T) {

	rng := rand.New(rand.NewSource(seed))

	cfg.AssetsDir = "../../../../prover-assets"

	spec := &InvalidityProofSpec{
		ChainID:             big.NewInt(51),
		ExpectedBlockHeight: 1_000_000_000,
		PrevFtxRollingHash:  "0x000000000000000000000000000000000000000000000000000000000000000a",
		FtxNumber:           1678,
	}

	ProcessInvaliditySpec(rng, spec, nil, "spec-invalidity.json")

}
