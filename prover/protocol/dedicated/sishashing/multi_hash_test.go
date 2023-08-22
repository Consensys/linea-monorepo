package sishashing

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/ringsis"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/dummy"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/stretchr/testify/require"
)

func TestMultiSisHash(t *testing.T) {

	// predeclare the commitment names
	var (
		ALLEGED_HASH ifaces.ColID = "ALLEGED_HASH"
		NUM_INPUTS   int          = 16
	)

	key := ringsis.StdParams.GenerateKey(NUM_INPUTS)

	// the test-vector we use has
	numHashes := 2
	testvecs := make([][]field.Element, numHashes)
	for i := range testvecs {
		testvecs[i] = vector.Rand(NUM_INPUTS)
	}

	digest := []field.Element{}
	for _, testvec := range testvecs {
		h := key.Hash(testvec)
		digest = append(digest, h...)
	}

	compiled := wizard.Compile(func(build *wizard.Builder) {
		// the wizard simply calls RingSISCheck over prover-defined columns
		preimages := make([]ifaces.Column, len(testvecs))
		for i := range testvecs {
			preimages[i] = build.RegisterCommit(limbPreimageName(i), NUM_INPUTS*key.NumLimbs())
		}
		allegedSisHash := build.InsertPublicInput(0, "ALLEGED_HASH", len(digest))
		MultiRingSISCheck("MULTI_SIS", build.CompiledIOP, &key, preimages, allegedSisHash)
	}, dummy.Compile)

	proof := wizard.Prove(compiled, func(assi *wizard.ProverRuntime) {
		// inject the precomputed values into it
		for i := range testvecs {
			assi.AssignColumn(limbPreimageName(i), smartvectors.NewRegular(key.LimbSplit(testvecs[i])))
		}
		assi.AssignColumn(ALLEGED_HASH, smartvectors.NewRegular(digest))
	})

	valid := wizard.Verify(compiled, proof)
	require.NoError(t, valid)

}
