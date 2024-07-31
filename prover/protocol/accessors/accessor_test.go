package accessors_test

import (
	"testing"

	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/accessors"
	"github.com/consensys/zkevm-monorepo/prover/protocol/coin"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/require"
)

func TestDeepEqual(t *testing.T) {

	a := accessors.NewConstant(field.NewElement(2))
	b := a

	require.Equal(t, a, b)
}

func TestCoinRounds(t *testing.T) {

	var (
		c coin.Info
		a ifaces.Accessor
	)

	define := func(b *wizard.Builder) {
		_ = b.RegisterCommit("p", 64)
		c = b.RegisterRandomCoin("c", coin.Field)
		a = accessors.NewFromCoin(c)
	}

	wizard.Compile(define)
	require.Equal(t, c.Round, a.Round())
}
