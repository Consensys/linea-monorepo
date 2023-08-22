package sishashing

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/ringsis"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/dummy"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/stretchr/testify/require"
)

func TestSisHash(t *testing.T) {

	// predeclare the commitment names
	var (
		PREIMAGE_LIMBS ifaces.ColID = "PREIMAGE_LIMBS"
		ALLEGED_HASH   ifaces.ColID = "ALLEGED_HASH"
		NUM_INPUTS     int          = 16
	)

	key := ringsis.StdParams.GenerateKey(NUM_INPUTS)

	// the test-vector we use has
	testvec := vector.Rand(NUM_INPUTS)
	digest := key.Hash(testvec)

	compiled := wizard.Compile(func(build *wizard.Builder) {
		// the wizard simply calls RingSISCheck over prover-defined columns
		limbSplitPreimage := build.RegisterCommit("PREIMAGE_LIMBS", len(testvec)*key.NumLimbs())
		allegedSisHash := build.InsertPublicInput(0, "ALLEGED_HASH", len(digest))
		RingSISCheck(build.CompiledIOP, &key, limbSplitPreimage, allegedSisHash)
	}, dummy.Compile)

	proof := wizard.Prove(compiled, func(assi *wizard.ProverRuntime) {
		// inject the precomputed values into it
		limbs := key.LimbSplit(testvec)
		assi.AssignColumn(PREIMAGE_LIMBS, smartvectors.NewRegular(limbs))
		assi.AssignColumn(ALLEGED_HASH, smartvectors.NewRegular(digest))
	})

	valid := wizard.Verify(compiled, proof)
	require.NoError(t, valid)

}
