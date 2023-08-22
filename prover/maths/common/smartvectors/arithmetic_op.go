package smartvectors

import (
	"math/big"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
)

// Represent either a linear combination with integer coefficients or a product with exponents
type operator interface {
	// res += x * coeff or res *= x ^coeff where both res and x are constants
	ConstIntoConst(res, x *field.Element, coeff int)
	// res += x * coeff or res *= x ^coeff where both res and x are vectors
	VecIntoVec(res, x []field.Element, coeff int)
	// res += x * coeff or res *= x ^coeff where res is a vector and x is a constant
	ConstIntoVec(res []field.Element, x *field.Element, coeff int)
	// res = x * coeff or res = x ^ coeff where x is a constant
	ConstIntoTerm(res, x *field.Element, coeff int)
	// res = x * coeff or res = x ^ coeff where x is a vector
	VecIntoTerm(res, x []field.Element, coeff int)
	// res += term or res *= coeff for constants
	ConstTermIntoConst(res, term *field.Element)
	// res += term for vectors
	VecTermIntoVec(res, term []field.Element)
	// res += term where res is a vector and term is a constant
	ConstTermIntoVec(res []field.Element, term *field.Element)
}

type linCombOp struct{}

func (linCombOp) ConstIntoConst(res, x *field.Element, coeff int) {
	switch coeff {
	case 1:
		res.Add(res, x)
	case -1:
		res.Sub(res, x)
	case 2:
		res.Add(res, x).Add(res, x)
	case -2:
		res.Sub(res, x).Sub(res, x)
	default:
		var c field.Element
		c.SetInt64(int64(coeff))
		c.Mul(&c, x)
		res.Add(res, &c)
	}
}

func (linCombOp) VecIntoVec(res, x []field.Element, coeff int) {
	// Sanity-check
	assertHasLength(len(res), len(x))
	switch coeff {
	case 1:
		vector.Add(res, res, x)
	case -1:
		vector.Sub(res, res, x)
	case 2:
		for i := range res {
			res[i].Add(&res[i], &x[i]).Add(&res[i], &x[i])
		}
	case -2:
		for i := range res {
			res[i].Sub(&res[i], &x[i]).Sub(&res[i], &x[i])
		}
	default:
		var c, tmp field.Element
		c.SetInt64(int64(coeff))
		for i := range res {
			tmp.Mul(&c, &x[i])
			res[i].Add(&res[i], &tmp)
		}
	}
}

func (linCombOp) ConstIntoVec(res []field.Element, val *field.Element, coeff int) {
	var term field.Element
	linCombOp.ConstIntoTerm(linCombOp{}, &term, val, coeff)
	linCombOp.ConstTermIntoVec(linCombOp{}, res, &term)
}

func (linCombOp) VecIntoTerm(term, x []field.Element, coeff int) {
	switch coeff {
	case 1:
		copy(term, x)
	case -1:
		for i := range term {
			term[i].Neg(&x[i])
		}
	case 2:
		vector.Add(term, x, x)
	case -2:
		for i := range term {
			term[i].Add(&x[i], &x[i]).Neg(&term[i])
		}
	default:
		var c field.Element
		c.SetInt64(int64(coeff))
		for i := range term {
			term[i].Mul(&c, &x[i])
		}
	}
}

func (linCombOp) ConstIntoTerm(term, x *field.Element, coeff int) {
	switch coeff {
	case 1:
		term.Set(x)
	case -1:
		term.Neg(x)
	case 2:
		term.Add(x, x)
	case -2:
		term.Add(x, x).Neg(term)
	default:
		var c field.Element
		c.SetInt64(int64(coeff))
		term.Mul(&c, x)
	}
}

func (linCombOp) ConstTermIntoConst(res, term *field.Element) {
	res.Add(res, term)
}

func (linCombOp) VecTermIntoVec(res, term []field.Element) {
	vector.Add(res, res, term)
}

func (linCombOp) ConstTermIntoVec(res []field.Element, term *field.Element) {
	for i := range res {
		res[i].Add(&res[i], term)
	}
}

type productOp struct{}

// res *= x ^coeff where both res and x are constants
func (productOp) ConstIntoConst(res, x *field.Element, coeff int) {
	switch coeff {
	case 0:
		// Nothing to do
	case 1:
		res.Mul(res, x)
	case 2:
		res.Mul(res, x).Mul(res, x)
	case 3:
		var tmp field.Element
		tmp.Square(x)
		tmp.Mul(&tmp, x)
		res.Mul(res, &tmp)
	default:
		var tmp field.Element
		tmp.Exp(*x, big.NewInt(int64(coeff)))
		res.Mul(res, &tmp)
	}
}

// res *= x ^coeff where both res and x are vectors
func (productOp) VecIntoVec(res, x []field.Element, coeff int) {

	// Sanity-check
	assertHasLength(len(res), len(x))

	switch coeff {
	case 0:
		// Nothing to do
	case 1:
		vector.MulElementWise(res, res, x)
	case 2:
		for i := range res {
			res[i].Mul(&res[i], &x[i]).Mul(&res[i], &x[i])
		}
	case 3:
		for i := range res {
			var tmp field.Element
			tmp.Square(&x[i])
			tmp.Mul(&tmp, &x[i])
			res[i].Mul(&res[i], &tmp)
		}
	default:
		var tmp field.Element
		for i := range res {
			tmp.Exp(x[i], big.NewInt(int64(coeff)))
			res[i].Mul(&res[i], &tmp)
		}
	}
}

// res *= x ^coeff where res is a vector and x is a constant
func (productOp) ConstIntoVec(res []field.Element, x *field.Element, coeff int) {
	var term field.Element
	productOp.ConstIntoTerm(productOp{}, &term, x, coeff)
	productOp.ConstTermIntoVec(productOp{}, res, &term)
}

// res = x ^ coeff where x is a constant
func (productOp) ConstIntoTerm(res, x *field.Element, coeff int) {
	switch coeff {
	case 0:
		res.SetOne()
	case 1:
		res.Set(x)
	case 2:
		res.Square(x)
	case 3:
		var tmp field.Element
		tmp.Square(x)
		res.Mul(&tmp, x)
	default:
		res.Exp(*x, big.NewInt(int64(coeff)))
	}
}

// res = x * coeff or res = x ^ coeff where x is a vector
func (productOp) VecIntoTerm(res, x []field.Element, coeff int) {
	switch coeff {
	case 0:
		vector.Fill(res, field.One())
	case 1:
		copy(res, x)
	case 2:
		vector.MulElementWise(res, x, x)
	case 3:
		for i := range res {
			// Creating a new variable for the case where res and x are the same variable
			var tmp field.Element
			tmp.Square(&x[i])
			res[i].Mul(&tmp, &x[i])
		}
	default:
		c := big.NewInt(int64(coeff))
		for i := range res {
			res[i].Exp(x[i], c)
		}
	}
}

// res += term or res *= coeff for constants
func (productOp) ConstTermIntoConst(res, term *field.Element) {
	res.Mul(res, term)
}

// res += term for vectors
func (productOp) VecTermIntoVec(res, term []field.Element) {
	vector.MulElementWise(res, res, term)

}

// res += term where res is a vector and term is a constant
func (productOp) ConstTermIntoVec(res []field.Element, term *field.Element) {
	vector.ScalarMul(res, res, *term)
}
