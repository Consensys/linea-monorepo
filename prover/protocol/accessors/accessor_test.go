package accessors_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/accessors"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/symbolic"
	"github.com/stretchr/testify/require"
)

func TestInterface(t *testing.T) {

	// Just some compilation checks, the test does nothing by itself
	var _ symbolic.Metadata = &ifaces.Accessor{}
	var _ ifaces.Runtime = &wizard.ProverRuntime{}
	var _ ifaces.GnarkRuntime = &wizard.WizardVerifierCircuit{}

}

func TestDeepEqual(t *testing.T) {

	a := accessors.AccessorFromConstant(field.NewElement(2))
	b := a

	require.Equal(t, a, b)
}

func TestCoinRounds(t *testing.T) {

	var (
		c coin.Info
		a *ifaces.Accessor
	)

	define := func(b *wizard.Builder) {
		_ = b.RegisterCommit("p", 64)
		c = b.RegisterRandomCoin("c", coin.Field)
		a = accessors.AccessorFromCoin(c)
	}

	wizard.Compile(define)
	require.Equal(t, c.Round, a.Round)
}
