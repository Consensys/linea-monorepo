package gnarkutil

import "github.com/consensys/gnark/frontend"

func RepeatedVariable(x frontend.Variable, n int) []frontend.Variable {
	res := make([]frontend.Variable, n)
	for i := range res {
		res[i] = x
	}
	return res
}

/*
exponentiation in gnark circuit, using the fast exponentiation
*/
func Exp(api frontend.API, x frontend.Variable, n int) frontend.Variable {

	if n < 0 {
		x = api.Inverse(x)
		return Exp(api, x, -n)
	}

	if n == 0 {
		return frontend.Variable(1)
	}

	if n == 1 {
		return x
	}

	if n%2 == 0 {
		x2 := api.Mul(x, x)
		return Exp(api, x2, n/2)
	}

	if n%2 == 1 {
		x2 := api.Mul(x, x)
		res := Exp(api, x2, (n-1)/2)
		return api.Mul(res, x)
	}

	panic("unreachable")
}
