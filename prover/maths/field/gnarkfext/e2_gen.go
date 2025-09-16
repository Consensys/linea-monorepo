package gnarkfext

import (
	"math/big"

	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
)

type Ext2[T zk.Element] struct {
	mixedAPI zk.APIGen[T]
}

func NewExt2[t zk.Element](api frontend.API) *Ext2[t] {
	mixedAPI, err := zk.NewApi[t](api)
	if err != nil {
		panic(err)
	}
	return &Ext2[t]{
		mixedAPI: mixedAPI,
	}
}

// 𝔽r²[u] = 𝔽r/u²-3

// E2Gen[T zk.FType] element in a quadratic extension
type E2Gen[T zk.Element] struct {
	A0, A1 T
}

func NewE2Gen[T zk.Element](v extensions.E2) E2Gen[T] {
	return E2Gen[T]{
		A0: zk.ValueOf[T](v.A0),
		A1: zk.ValueOf[T](v.A1),
	}
}

// SetZero returns a newly allocated element equal to 0
func (ext2 *Ext2[T]) Zero() *E2Gen[T] {
	return &E2Gen[T]{
		A0: *ext2.mixedAPI.FromUint(0),
		A1: *ext2.mixedAPI.FromUint(0),
	}
}

// SetOne returns a newly allocated element equal to 1
func (ext2 *Ext2[T]) One() *E2Gen[T] {
	return &E2Gen[T]{
		A0: *ext2.mixedAPI.FromUint(1),
		A1: *ext2.mixedAPI.FromUint(0),
	}
}

// IsZero returns 1 if the element is equal to 0 and 0 otherwise
func (ext2 *Ext2[T]) IsZero(e *E2Gen[T]) frontend.Variable {
	return ext2.mixedAPI.And(
		ext2.mixedAPI.IsZero(&e.A0),
		ext2.mixedAPI.IsZero(&e.A1),
	)
}

// Neg negates a E2Gen[T zk.FType] elmt
func (ext2 *Ext2[T]) Neg(e1 *E2Gen[T]) *E2Gen[T] {
	zero := ext2.mixedAPI.FromUint(0)
	return &E2Gen[T]{
		A0: *ext2.mixedAPI.Sub(zero, &e1.A0),
		A1: *ext2.mixedAPI.Sub(zero, &e1.A1),
	}
}

// Add E2Gen[T zk.FType] elmts
func (ext2 *Ext2[T]) Add(e1, e2 *E2Gen[T]) *E2Gen[T] {
	return &E2Gen[T]{
		A0: *ext2.mixedAPI.Add(&e1.A0, &e2.A0),
		A1: *ext2.mixedAPI.Add(&e1.A1, &e2.A1),
	}
}

// Double E2Gen[T zk.FType] elmt
func (ext2 *Ext2[T]) Double(e1 *E2Gen[T]) *E2Gen[T] {
	two := ext2.mixedAPI.FromUint(2)
	return &E2Gen[T]{
		A0: *ext2.mixedAPI.Mul(&e1.A0, two),
		A1: *ext2.mixedAPI.Mul(&e1.A1, two),
	}
}

// Sub E2Gen[T zk.FType] elmts
func (ext2 *Ext2[T]) Sub(e1, e2 *E2Gen[T]) *E2Gen[T] {
	return &E2Gen[T]{
		A0: *ext2.mixedAPI.Sub(&e1.A0, &e2.A0),
		A1: *ext2.mixedAPI.Sub(&e1.A1, &e2.A1),
	}
}

// Mul E2Gen[T zk.FType] elmts
func (ext2 *Ext2[T]) Mul(e1, e2 *E2Gen[T]) *E2Gen[T] {

	l1 := ext2.mixedAPI.Add(&e1.A0, &e1.A1)
	l2 := ext2.mixedAPI.Add(&e2.A0, &e2.A1)

	u := ext2.mixedAPI.Mul(l1, l2)

	ac := ext2.mixedAPI.Mul(&e1.A0, &e2.A0)
	bd := ext2.mixedAPI.Mul(&e1.A1, &e2.A1)

	l31 := ext2.mixedAPI.Add(ac, bd)

	l41 := ext2.mixedAPI.Mul(bd, ext2.mixedAPI.FromKoalabear(ext.qnrE2))

	return &E2Gen[T]{
		A0: *ext2.mixedAPI.Add(ac, l41),
		A1: *ext2.mixedAPI.Sub(u, l31),
	}
}

func (ext2 *Ext2[T]) Square(x *E2Gen[T]) *E2Gen[T] {
	// Square sets z to the E2-product of x,x returns z
	var res E2Gen[T]
	a := ext2.mixedAPI.Mul(&x.A0, &x.A1)
	c := ext2.mixedAPI.Mul(&x.A0, &x.A0)
	b := ext2.mixedAPI.Mul(&x.A1, &x.A1)
	b = ext2.mixedAPI.MulConst(b, big.NewInt(3))
	res.A0 = *ext2.mixedAPI.Add(c, b)
	res.A1 = *ext2.mixedAPI.Add(a, a)
	return &res

}

// MulByFp multiplies an fp2 elmt by an fp elmt
func (ext2 *Ext2[T]) MulByFp(e1 *E2Gen[T], c *T) *E2Gen[T] {
	return &E2Gen[T]{
		A0: *ext2.mixedAPI.Mul(&e1.A0, c),
		A1: *ext2.mixedAPI.Mul(&e1.A1, c),
	}
}

// MulByNonResidue multiplies an fp2 elmt by the imaginary elmt
// ext.uSquare is the square of the imaginary root
func (ext2 *Ext2[T]) MulByNonResidue(e1 *E2Gen[T]) *E2Gen[T] {
	return &E2Gen[T]{
		A0: *ext2.mixedAPI.Mul(&e1.A1, ext2.mixedAPI.FromKoalabear(ext.qnrE2)),
		A1: e1.A0,
	}
}

// Conjugate conjugation of an E2Gen[T zk.FType] elmt
func (ext2 *Ext2[T]) Conjugate(e1 *E2Gen[T]) *E2Gen[T] {
	return &E2Gen[T]{
		A0: e1.A0,
		A1: *ext2.mixedAPI.Neg(&e1.A1),
	}
}

func (ext2 *Ext2[T]) assign(e1 []*T) *E2Gen[T] {
	return &E2Gen[T]{
		A0: *e1[0],
		A1: *e1[1],
	}
}

// Inverse E2Gen[T zk.FType] elmts
func (ext2 *Ext2[T]) Inverse(e1 *E2Gen[T]) *E2Gen[T] {

	res, err := ext2.mixedAPI.NewHint(inverseE2Hint, 2, &e1.A0, &e1.A1)
	if err != nil {
		// err is non-nil only for invalid number of inputs
		panic(err)
	}

	e3 := ext2.assign(res[:2])
	one := ext2.One()

	// 1 == e3 * e1
	_res := ext2.Mul(e3, e1)
	ext2.AssertIsEqual(_res, one)
	return e3
}

// // Assign a value to self (witness assignment)
// func (ext2 *Ext2[T]) Assign(a extensions.E2) {
// 	e.A0 = *api.FromKoalabear(a.A0)
// 	e.A1 = *api.FromKoalabear(a.A1)
// }

// AssertIsEqual constraint self to be equal to other into the given constraint system
func (ext2 *Ext2[T]) AssertIsEqual(e, other *E2Gen[T]) {
	ext2.mixedAPI.AssertIsEqual(&e.A0, &other.A0)
	ext2.mixedAPI.AssertIsEqual(&e.A1, &other.A1)
}

// Select sets e to r1 if b=1, r2 otherwise
func (ext2 *Ext2[T]) Select(b frontend.Variable, r1, r2 *E2Gen[T]) *E2Gen[T] {
	return &E2Gen[T]{
		A0: *ext2.mixedAPI.Select(b, &r1.A0, &r2.A0),
		A1: *ext2.mixedAPI.Select(b, &r1.A1, &r2.A1),
	}
}

// Div e2 elmts
// func (ext2 *Ext2[T]) Div(e1, e2 *E2Gen[T]) *E2Gen[T] {
// 	return &E2Gen[T]{
// 		A0: *ext2.mixedAPI.Div(&e1.A0, &e2.A0),
// 		A1: *ext2.mixedAPI.Div(&e1.A1, &e2.A1),
// 	}
// }

// DivByBase  Div e2 Element by a base elmt
func (ext2 *Ext2[T]) DivByBase(e1 *E2Gen[T], e2 *T) *E2Gen[T] {
	return &E2Gen[T]{
		A0: *ext2.mixedAPI.Div(&e1.A0, e2),
		A1: *ext2.mixedAPI.Div(&e1.A1, e2),
	}
}
