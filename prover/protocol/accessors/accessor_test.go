package accessors_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/stretchr/testify/require"
)

func TestDeepEqual(t *testing.T) {

	a := accessors.NewConstant[zk.NativeElement](field.NewElement(2))
	b := a

	require.Equal(t, a, b)
}

func TestCoinRounds(t *testing.T) {

	var (
		c coin.Info
		a ifaces.Accessor[zk.NativeElement]
	)

	define := func(b *wizard.Builder[zk.NativeElement]) {
		_ = b.RegisterCommit("p", 64)
		c = b.RegisterRandomCoin("c", coin.FieldExt)
		a = accessors.NewFromCoin[zk.NativeElement](c)
	}

	wizard.Compile(define)
	require.Equal(t, c.Round, a.Round())
}
