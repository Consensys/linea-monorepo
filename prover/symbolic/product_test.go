package symbolic_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/symbolic"
	"github.com/stretchr/testify/require"
)

func TestExp(t *testing.T) {

	getdeg := func(interface{}) int { return 1 }
	x := symbolic.NewDummyVar("x")
	y := symbolic.NewDummyVar("y")

	{
		// x time y
		expr := x.Mul(symbolic.NewConstant(1))
		board := expr.Board()
		deg := board.Degree(getdeg)
		require.Equal(t, deg, 1)
	}

	{
		// x time y
		expr := x.Mul(y)
		board := expr.Board()
		deg := board.Degree(getdeg)
		require.Equal(t, deg, 2)
	}

	{
		// x^2
		expr := x.Square()
		board := expr.Board()
		deg := board.Degree(getdeg)
		require.Len(t, board.Nodes, 2)
		require.Len(t, board.Nodes[0], 1)
		require.Len(t, board.Nodes[1], 1)
		require.Len(t, board.Nodes[1][0].Children, 1)
		require.Equal(t, deg, 2)
	}

	{
		// x^3
		expr := x.Square().Mul(x)
		board := expr.Board()
		deg := board.Degree(getdeg)
		require.Equal(t, deg, 3)
	}

	{
		// x^4
		expr := x.Square().Mul(x.Square())
		expr.AssertValid()
		board := expr.Board()
		deg := board.Degree(getdeg)
		require.Equal(t, deg, 4)
	}

	{
		// x^4
		expr := x.Mul(x).Mul(x).Mul(x)
		expr.AssertValid()
		board := expr.Board()
		deg := board.Degree(getdeg)
		require.Equal(t, deg, 4)
	}

	{
		// x^5
		expr := x.Pow(5)
		expr.AssertValid()
		require.Lenf(t, expr.Children, 1, "for input %v\n", 5)
		require.IsType(t, expr.Operator, symbolic.Product{})
		board := expr.Board()
		deg := board.Degree(getdeg)
		require.Equal(t, deg, 5)
	}

	for i := 2; i < 10; i++ {
		expr := x.Pow(i)
		expr.AssertValid()
		require.Lenf(t, expr.Children, 1, "for input %v\n", i)
		require.IsType(t, expr.Operator, symbolic.Product{})
		board := expr.Board()
		deg := board.Degree(getdeg)
		require.Equal(t, deg, i)
	}

}
