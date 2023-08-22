package symbolic_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/symbolic"
	"github.com/stretchr/testify/require"
)

func TestLinComb(t *testing.T) {

	x := symbolic.NewDummyVar("x")
	y := symbolic.NewDummyVar("y")

	{
		// x + x = 2x
		expr := x.Add(x)
		require.Len(t, expr.Children, 1)
	}

	{
		// x + y
		expr := x.Add(y)
		require.Len(t, expr.Children, 2)
	}

	{
		// (x + x) + (x + x)
		expr := x.Add(x).Add(x.Add(x))
		require.Len(t, expr.Children, 1)
	}

	{
		// ((x + x) + x) + x
		expr := x.Add(x).Add(x).Add(x)
		require.Len(t, expr.Children, 1)
	}

}
