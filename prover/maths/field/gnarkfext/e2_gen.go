package gnarkfext

import (
	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
)

// 𝔽r²[u] = 𝔽r/u²-3

// E2Gen[T zk.FType] element in a quadratic extension
type E2Gen[T zk.FType] struct {
	A0, A1 T
}

// SetZero returns a newly allocated element equal to 0
func (e *E2Gen[T]) SetZero(api zk.FieldOps[T]) *E2Gen[T] {
	e.A0 = *api.FromUint(0)
	e.A1 = *api.FromUint(0)
	return e
}

// SetOne returns a newly allocated element equal to 1
func (e *E2Gen[T]) SetOne(api zk.FieldOps[T]) *E2Gen[T] {
	e.A0 = *api.FromUint(1)
	e.A1 = *api.FromUint(0)
	return e
}

// IsZero returns 1 if the element is equal to 0 and 0 otherwise
func (e *E2Gen[T]) IsZero(api zk.FieldOps[T]) frontend.Variable {
	return api.And(api.IsZero(&e.A0), api.IsZero(&e.A1))
}

// Neg negates a E2Gen[T zk.FType] elmt
func (e *E2Gen[T]) Neg(api zk.FieldOps[T], e1 *E2Gen[T]) *E2Gen[T] {
	zero := api.FromUint(0)
	e.A0 = *api.Sub(zero, &e1.A0)
	e.A1 = *api.Sub(zero, &e1.A1)
	return e
}

// Add E2Gen[T zk.FType] elmts
func (e *E2Gen[T]) Add(api zk.FieldOps[T], e1, e2 *E2Gen[T]) *E2Gen[T] {
	e.A0 = *api.Add(&e1.A0, &e2.A0)
	e.A1 = *api.Add(&e1.A1, &e2.A1)
	return e
}

// Double E2Gen[T zk.FType] elmt
func (e *E2Gen[T]) Double(api zk.FieldOps[T], e1 *E2Gen[T]) *E2Gen[T] {
	two := api.FromUint(2)
	e.A0 = *api.Mul(&e1.A0, two)
	e.A1 = *api.Mul(&e1.A1, two)
	return e
}

// Sub E2Gen[T zk.FType] elmts
func (e *E2Gen[T]) Sub(api zk.FieldOps[T], e1, e2 *E2Gen[T]) *E2Gen[T] {
	e.A0 = *api.Sub(&e1.A0, &e2.A0)
	e.A1 = *api.Sub(&e1.A1, &e2.A1)
	return e
}

// Mul E2Gen[T zk.FType] elmts
func (e *E2Gen[T]) Mul(api zk.FieldOps[T], e1, e2 *E2Gen[T]) *E2Gen[T] {

	l1 := api.Add(&e1.A0, &e1.A1)
	l2 := api.Add(&e2.A0, &e2.A1)

	u := api.Mul(l1, l2)

	ac := api.Mul(&e1.A0, &e2.A0)
	bd := api.Mul(&e1.A1, &e2.A1)

	l31 := api.Add(ac, bd)
	e.A1 = *api.Sub(u, l31)

	l41 := api.Mul(bd, api.FromKoalabear(ext.qnrE2))
	e.A0 = *api.Add(ac, l41)

	return e
}

// // Square E2Gen[T zk.FType] elt
func (e *E2Gen[T]) Square(api zk.FieldOps[T], x *E2Gen[T]) *E2Gen[T] {

	extqnrE2 := api.FromKoalabear(ext.qnrE2)
	two := api.FromUint(2)

	c0 := api.Mul(&x.A0, &x.A0) // x0^2
	c1 := api.Mul(&x.A1, &x.A1) // x1^2
	c1 = api.Mul(c1, extqnrE2)  // qnr*x1^2
	c0 = api.Add(c0, c1)        // x0^2 + qnr*x1^2
	c1 = api.Mul(extqnrE2, two) // 2*qnr
	c1 = api.Mul(c1, &x.A0)     // 2*qnr*x0
	c1 = api.Mul(c1, &x.A1)     // 2*qnr*x0*x1
	e.A0 = *c0                  // x0^2 + qnr*x1^2
	e.A1 = *c1                  // 2*qnr*x0*x1

	return e
}

// MulByFp multiplies an fp2 elmt by an fp elmt
func (e *E2Gen[T]) MulByFp(api zk.FieldOps[T], e1 *E2Gen[T], c *T) *E2Gen[T] {
	e.A0 = *api.Mul(&e1.A0, c)
	e.A1 = *api.Mul(&e1.A1, c)
	return e
}

// MulByNonResidue multiplies an fp2 elmt by the imaginary elmt
// ext.uSquare is the square of the imaginary root
func (e *E2Gen[T]) MulByNonResidue(api zk.FieldOps[T], e1 *E2Gen[T]) *E2Gen[T] {
	x := e1.A0
	e.A0 = *api.Mul(&e1.A1, api.FromKoalabear(ext.qnrE2))
	e.A1 = x
	return e
}

// Conjugate conjugation of an E2Gen[T zk.FType] elmt
func (e *E2Gen[T]) Conjugate(api zk.FieldOps[T], e1 *E2Gen[T]) *E2Gen[T] {
	e.A0 = e1.A0
	e.A1 = *api.Neg(&e1.A1)
	return e
}

func (e *E2Gen[T]) assign(e1 []*T) {
	e.A0 = *e1[0]
	e.A1 = *e1[1]
}

// Inverse E2Gen[T zk.FType] elmts
func (e *E2Gen[T]) Inverse(api zk.FieldOps[T], e1 *E2Gen[T]) *E2Gen[T] {

	res, err := api.NewHint(inverseE2Hint, 2, &e1.A0, &e1.A1)
	if err != nil {
		// err is non-nil only for invalid number of inputs
		panic(err)
	}

	var e3, one E2Gen[T]
	e3.assign(res[:2])
	one.SetOne(api)

	// 1 == e3 * e1
	e3.Mul(api, &e3, e1)
	e3.AssertIsEqual(api, &one)

	e.assign(res[:2])

	return e
}

// Assign a value to self (witness assignment)
func (e *E2Gen[T]) Assign(api zk.FieldOps[T], a extensions.E2) {
	e.A0 = *api.FromKoalabear(a.A0)
	e.A1 = *api.FromKoalabear(a.A1)
}

// AssertIsEqual constraint self to be equal to other into the given constraint system
func (e *E2Gen[T]) AssertIsEqual(api zk.FieldOps[T], other *E2Gen[T]) {
	api.AssertIsEqual(&e.A0, &other.A0)
	api.AssertIsEqual(&e.A1, &other.A1)
}

// Select sets e to r1 if b=1, r2 otherwise
func (e *E2Gen[T]) Select(api zk.FieldOps[T], b frontend.Variable, r1, r2 *E2Gen[T]) **E2Gen[T] {

	e.A0 = *api.Select(b, &r1.A0, &r2.A0)
	e.A1 = *api.Select(b, &r1.A1, &r2.A1)

	return &e
}

// Div e2 elmts
func (e *E2Gen[T]) Div(api zk.FieldOps[T], e1, e2 *E2Gen[T]) *E2Gen[T] {
	e.A0 = *api.Div(&e1.A0, &e2.A0)
	e.A1 = *api.Div(&e1.A1, &e2.A1)
	return e
}

// DivByBase  Div e2 Element by a base elmt
func (e *E2Gen[T]) DivByBase(api zk.FieldOps[T], e1 *E2Gen[T], e2 *T) *E2Gen[T] {
	e.A0 = *api.Div(&e1.A0, e2)
	e.A1 = *api.Div(&e1.A1, e2)
	return e
}

// // Assign a value to self (witness assignment)
// func (e *E2Gen[T zk.FType]) Assign(a extensions.E2Gen[T zk.FType]) {
// 	e.A0 = a.A0
// 	e.A1 = a.A1
// }

// // AssertIsEqual constraint self to be equal to other into the given constraint system
// func (e *E2Gen[T zk.FType]) AssertIsEqual(api frontend.API, other E2Gen[T zk.FType]) {
// 	api.AssertIsEqual(e.A0, other.A0)
// 	api.AssertIsEqual(e.A1, other.A1)
// }

// // Select sets e to r1 if b=1, r2 otherwise
// func (e *E2Gen[T zk.FType]) Select(api frontend.API, b frontend.Variable, r1, r2 E2Gen[T zk.FType]) *E2Gen[T zk.FType] {

// 	e.A0 = api.Select(b, r1.A0, r2.A0)
// 	e.A1 = api.Select(b, r1.A1, r2.A1)

// 	return e
// }

// // Sub E2Gen[T zk.FType] elmts
// func (e *E2Gen[T zk.FType]) Div(api frontend.API, e1, E2Gen[T zk.FType] E2Gen[T zk.FType]) *E2Gen[T zk.FType] {
// 	e.A0 = api.Div(e1.A0, E2Gen[T zk.FType].A0)
// 	e.A1 = api.Div(e1.A1, E2Gen[T zk.FType].A1)
// 	return e
// }

// // Sub Element elmts
// func (e *E2Gen[T zk.FType]) DivByBase(api frontend.API, e1 E2Gen[T zk.FType], E2Gen[T zk.FType] frontend.Variable) *E2Gen[T zk.FType] {
// 	e.A0 = api.Div(e1.A0, E2Gen[T zk.FType])
// 	e.A1 = api.Div(e1.A1, E2Gen[T zk.FType])
// 	return e
// }
