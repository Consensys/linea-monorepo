package single_round_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/ringsis"
	"github.com/consensys/accelerated-crypto-monorepo/example/single_round"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/dummy"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/selfrecursion"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/vortex"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils/profiling"
	"github.com/consensys/accelerated-crypto-monorepo/zkevm"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestExample(t *testing.T) {

	logrus.SetLevel(logrus.DebugLevel)

	def, prov := single_round.GenWizard(1<<6, 1<<8)

	// profiling.ProfileTrace("test-example", true, true, func() {
	compiled := wizard.Compile(
		zkevm.WrapDefine(def),
		compiler.Arcane(16, 64),
		vortex.Compile(2),
		selfrecursion.SelfRecurse,
		dummy.Compile,
	)

	var proof wizard.Proof

	profiling.ProfileTrace("example-with-self-recursion", true, true, func() {
		proof = wizard.Prove(compiled, prov)
	})

	err := wizard.Verify(compiled, proof)
	require.NoError(t, err)
}

func BenchmarkSingleRound(b *testing.B) {

	logrus.SetLevel(logrus.InfoLevel)

	var (
		size      int = 1 << 15
		smallSize int = 1 << 12
		splitSize int = 1 << 8
	)

	def, prov := single_round.GenWizard(smallSize, size)

	sisInstance := ringsis.Params{
		LogTwoBound_: 8,
		LogTwoDegree: 6,
	}

	// profiling.ProfileTrace("test-example", true, true, func() {
	compiled := wizard.Compile(
		zkevm.WrapDefine(def),
		compiler.Arcane(splitSize/4, splitSize, false),
		vortex.Compile(
			2,
			vortex.WithSISParams(&sisInstance),
			vortex.MerkleMode,
		),
	)

	profiling.ProfileTrace("example-with-self-recursion", true, true, func() {
		_ = wizard.Prove(compiled, prov)
	})
}
