package symbolic_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/symbolic"
	"github.com/stretchr/testify/require"
)

func TestLCConstruction(t *testing.T) {

	x := symbolic.NewDummyVar("x")
	y := symbolic.NewDummyVar("y")

	{
		/*
			Test t a simple case of addition
		*/
		expr1 := x.Add(y)

		require.Equal(t, 2, len(expr1.Children))
		require.Equal(t, expr1.Children[0], x)
		require.Equal(t, expr1.Children[1], y)
		require.Equal(t, 2, len(expr1.Operator.(symbolic.LinComb).Coeffs))
		require.Equal(t, expr1.Operator.(symbolic.LinComb).Coeffs[0], 1)
		require.Equal(t, expr1.Operator.(symbolic.LinComb).Coeffs[1], 1)
	}

	{
		/*
			Adding y then substracting x should give back (y)
		*/
		expr1 := x.Add(y).Sub(x)
		require.Equal(t, expr1, y)
	}

	{
		/*
			Same thing when using Neg
		*/
		expr := x.Neg().Add(x).Add(y)
		require.Equal(t, expr, y)
	}

}

func TestProductConstruction(t *testing.T) {

	x := symbolic.NewDummyVar("x")
	y := symbolic.NewDummyVar("y")

	{
		/*
			Test t a simple case of addition
		*/
		expr1 := x.Mul(y)

		require.Equal(t, 2, len(expr1.Children))
		require.Equal(t, expr1.Children[0], x)
		require.Equal(t, expr1.Children[1], y)
		require.Equal(t, 2, len(expr1.Operator.(symbolic.Product).Exponents))
		require.Equal(t, expr1.Operator.(symbolic.Product).Exponents[0], 1)
		require.Equal(t, expr1.Operator.(symbolic.Product).Exponents[1], 1)
	}

	{
		/*
			Adding y then substracting x should give back (y)
		*/
		expr1 := x.Mul(y).Mul(x)
		require.Equal(t, 2, len(expr1.Children))
		require.Equal(t, expr1.Children[0], x)
		require.Equal(t, expr1.Children[1], y)
		require.Equal(t, 2, len(expr1.Operator.(symbolic.Product).Exponents))
		require.Equal(t, expr1.Operator.(symbolic.Product).Exponents[0], 2)
		require.Equal(t, expr1.Operator.(symbolic.Product).Exponents[1], 1)
	}

	{
		/*
			The Neg should be factored out of the product
		*/
		expr := x.Neg().Mul(x)
		require.Equal(t, 1, len(expr.Children))
		require.Equal(t, 1, len(expr.Operator.(symbolic.LinComb).Coeffs))
		require.Equal(t, expr.Operator.(symbolic.LinComb).Coeffs[0], -1)
	}

	{
		/*
			When we square
		*/
		expr := x.Mul(x)
		require.Equal(t, 1, len(expr.Children))
		require.Equal(t, expr.Children[0], x)
		require.Equal(t, 1, len(expr.Operator.(symbolic.Product).Exponents))
		require.Equal(t, expr.Operator.(symbolic.Product).Exponents[0], 2)
	}

}
