package accessors_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/largefield/maths/field"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/largefield/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/largefield/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/largefield/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/largefield/protocol/wizard"
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
