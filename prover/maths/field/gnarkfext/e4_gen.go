package gnarkfext

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
)

// Element element in a quadratic extension
type E4Gen[T zk.FType] struct {
	B0, B1 E2Gen[T]
}

// SetZero returns a newly allocated element equal to 0
func (e *E4Gen[T]) SetZero(api zk.FieldOps[T]) *E4Gen[T] {
	e.B0.SetZero(api)
	e.B1.SetZero(api)
	return e
}

// SetOne returns a newly allocated element equal to 1
func (e *E4Gen[T]) SetOne(api zk.FieldOps[T]) *E4Gen[T] {
	e.B0.SetOne(api)
	e.B1.SetZero(api)
	return e
}

// IsZero returns 1 if the element is equal to 0 and 0 otherwise
func (e *E4Gen[T]) IsZero(api zk.FieldOps[T]) frontend.Variable {
	return api.And(e.B0.IsZero(api), e.B1.IsZero(api))
}

func (e *E4Gen[T]) assign(e1 []*T) {
	e.B0.assign(e1[:2])
	e.B1.assign(e1[2:4])
}

// Neg negates a Element elmt
func (e *E4Gen[T]) Neg(api zk.FieldOps[T], e1 *E4Gen[T]) *E4Gen[T] {
	e.B0.Neg(api, &e1.B0)
	e.B1.Neg(api, &e1.B1)
	return e
}

// Add Element elmts
func (e *E4Gen[T]) Add(api zk.FieldOps[T], e1, e2 *E4Gen[T]) *E4Gen[T] {
	e.B0.Add(api, &e1.B0, &e2.B0)
	e.B1.Add(api, &e1.B1, &e2.B1)
	return e
}

// e = e1+e2
func (e *E4Gen[T]) AddByBase(api zk.FieldOps[T], e1 *E4Gen[T], e2 *T) *E4Gen[T] {
	e.B0.A0 = *api.Add(&e1.B0.A0, e2)
	return e
}

// Double Element elmt
func (e *E4Gen[T]) Double(api zk.FieldOps[T], e1 *E4Gen[T]) *E4Gen[T] {
	e.B0.Double(api, &e1.B0)
	e.B1.Double(api, &e1.B1)
	return e
}

// Sub Element elmts
func (e *E4Gen[T]) Sub(api zk.FieldOps[T], e1, e2 E4Gen[T]) *E4Gen[T] {
	e.B0.Sub(api, &e1.B0, &e2.B0)
	e.B1.Sub(api, &e1.B1, &e2.B1)
	return e
}

// Mul e2 elmts, e = e1*e2
func (e *E4Gen[T]) Mul(api zk.FieldOps[T], e1, e2 *E4Gen[T], in ...*E4Gen[T]) *E4Gen[T] {

	api.Println(e1.B0.A0)

	var l1, l2 E2Gen[T]
	l1.Add(api, &e1.B0, &e1.B1)
	l2.Add(api, &e2.B0, &e2.B1)

	var u E2Gen[T]
	u.Mul(api, &l1, &l2)

	var ac, bd E2Gen[T]
	ac.Mul(api, &e1.B0, &e2.B0)
	bd.Mul(api, &e1.B1, &e2.B1)

	var l31 E2Gen[T]
	l31.Add(api, &ac, &bd)
	e.B1.Sub(api, &u, &l31)

	var l41 E2Gen[T]
	// l41.Mul(api, bd, ext.qnrElement)
	l41.MulByNonResidue(api, &bd)
	e.B0.Add(api, &ac, &l41)

	if len(in) > 0 {
		return e.Mul(api, e, in[0], in[1:]...)
	} else {
		return e
	}
}

// Square Element elt
func (e *E4Gen[T]) Square(api zk.FieldOps[T], x *E4Gen[T]) *E4Gen[T] {

	var c0, c2, c3 E2Gen[T]
	c0.Sub(api, &x.B0, &x.B1)
	c3.MulByNonResidue(api, &x.B1).Sub(api, &x.B0, &c3)
	c2.Mul(api, &x.B0, &x.B1)
	c0.Mul(api, &c0, &c3).Add(api, &c0, &c2)
	e.B1.Double(api, &c2)
	c2.MulByNonResidue(api, &c2)
	e.B0.Add(api, &c0, &c2)

	return e
}

// MulByFp multiplies an fp2 elmt by an fp elmt
func (e *E4Gen[T]) MulByE2(api zk.FieldOps[T], e1 *E4Gen[T], c *E2Gen[T]) *E4Gen[T] {
	e.B0.Mul(api, &e1.B0, c)
	e.B1.Mul(api, &e1.B1, c)
	return e
}

// MulByFp multiplies an Fp4 elmt by an fp elmt
func (e *E4Gen[T]) MulByFp(api zk.FieldOps[T], e1 *E4Gen[T], c *T) *E4Gen[T] {
	e.B0.MulByFp(api, &e1.B0, c)
	e.B1.MulByFp(api, &e1.B1, c)
	return e
}

// Sum sets e = e1 + e2 + e3...
func (e *E4Gen[T]) Sum(api zk.FieldOps[T], e1 *E4Gen[T], e2 *E4Gen[T], e3 ...*E4Gen[T]) *E4Gen[T] {
	e.Add(api, e1, e2)
	for i := 0; i < len(e3); i++ {
		e.Add(api, e, e3[i])
	}
	return e
}

// AssertIsEqual asserts that e==e1
func (e *E4Gen[T]) AssertIsEqual(api zk.FieldOps[T], e1 *E4Gen[T]) {
	e.B0.AssertIsEqual(api, &e1.B0)
	e.B1.AssertIsEqual(api, &e1.B1)
}

// MulByNonResidue multiplies an fp2 elmt by the imaginary elmt
// ext.uSquare is the square of the imaginary root
func (e *E4Gen[T]) MulByNonResidue(api zk.FieldOps[T], e1 *E4Gen[T]) *E4Gen[T] {
	x := e1.B0
	e.B0.MulByNonResidue(api, &e.B1)
	e.B1 = x
	return e
}

// Conjugate conjugation of an Element elmt
func (e *E4Gen[T]) Conjugate(api zk.FieldOps[T], e1 E4Gen[T]) *E4Gen[T] {
	e.B0 = e1.B0
	e.B1.Neg(api, &e1.B1)
	return e
}

// Inverse Element elmts
func (e *E4Gen[T]) Inverse(api zk.FieldOps[T], e1 *E4Gen[T]) *E4Gen[T] {

	res, err := api.NewHint(inverseE4Hint, 4, &e1.B0.A0, &e1.B0.A1, &e1.B1.A0, &e1.B1.A1)
	if err != nil {
		// err is non-nil only for invalid number of inputs
		panic(err)
	}

	var e3, one E4Gen[T]
	e3.assign(res[:4])
	one.SetOne(api)

	// 1 == e3 * e1
	e3.Mul(api, &e3, e1)
	e3.AssertIsEqual(api, &one)

	e.assign(res[:4])

	return e
}

// Assign a value to self (witness assignment)
func (e *E4Gen[T]) Assign(api zk.FieldOps[T], a fext.Element) {
	e.B0.Assign(api, a.B0)
	e.B1.Assign(api, a.B1)
}

// Select sets e to r1 if b=1, r2 otherwise
func (e *E4Gen[T]) Select(api zk.FieldOps[T], b frontend.Variable, r1, r2 *E4Gen[T]) *E4Gen[T] {

	e.B0.Select(api, b, &r1.B0, &r2.B0)
	e.B1.Select(api, b, &r1.B1, &r2.B1)

	return e
}

// Sub Element elmts
func (e *E4Gen[T]) Div(api zk.FieldOps[T], e1, e2 *E4Gen[T]) *E4Gen[T] {
	e.B0.Div(api, &e1.B0, &e2.B0)
	e.B1.Div(api, &e1.B1, &e2.B1)
	return e
}

// Sub Element elmts
func (e *E4Gen[T]) DivByBase(api zk.FieldOps[T], e1 *E4Gen[T], e2 *T) *E4Gen[T] {
	e.B0.DivByBase(api, &e1.B0, e2)
	e.B1.DivByBase(api, &e1.B1, e2)
	return e
}

// Exp exponentiation in gnark circuit, using the fast exponentiation
func ExpGen[T zk.FType](api zk.FieldOps[T], x *E4Gen[T], n int) *E4Gen[T] {

	if n < 0 {
		x.Inverse(api, x)
		return ExpGen(api, x, -n)
	}

	if n == 0 {
		var one E4Gen[T]
		one.SetOne(api)
		return &one
	}

	if n == 1 {
		return x
	}

	var x2 E4Gen[T]
	x2.Mul(api, x, x)
	if n%2 == 0 {
		return ExpGen(api, &x2, n/2)
	} else {
		res := ExpGen(api, &x2, (n-1)/2)
		return res.Mul(api, res, x)
	}

}

func NewE4Gen[T zk.FType](v fext.Element) E4Gen[T] {
	var a T
	if _, ok := any(a).(emulated.Element[emulated.KoalaBear]); ok {
		var tv E4Gen[emulated.Element[emulated.KoalaBear]]
		tv.B0.A0 = emulated.ValueOf[emulated.KoalaBear](v.B0.A0)
		tv.B0.A1 = emulated.ValueOf[emulated.KoalaBear](v.B0.A1)
		tv.B1.A0 = emulated.ValueOf[emulated.KoalaBear](v.B1.A0)
		tv.B1.A1 = emulated.ValueOf[emulated.KoalaBear](v.B1.A1)
		return any(tv).(E4Gen[T])
	}
	var tv E4Gen[frontend.Variable]
	tv.B0.A0 = v.B0.A0
	tv.B0.A1 = v.B0.A1
	tv.B1.A0 = v.B1.A0
	tv.B1.A1 = v.B1.A1
	return any(tv).(E4Gen[T])
}
