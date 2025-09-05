package gnarkfext

import (
	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
)

type Ext2[T zk.Element] struct {
	mixedAPI zk.FieldOps[T]
	api      frontend.API
}

func NewExt2[t zk.Element](api frontend.API) *Ext2[t] {
	mixedAPI, err := zk.NewApi[t](api)
	if err != nil {
		panic(err)
	}
	return &Ext2[t]{
		mixedAPI: mixedAPI,
		api:      api,
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
	return ext2.api.And(
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

// // Square E2Gen[T zk.FType] elt
func (ext2 *Ext2[T]) Square(x *E2Gen[T]) *E2Gen[T] {

	extqnrE2 := ext2.mixedAPI.FromKoalabear(ext.qnrE2)
	two := ext2.mixedAPI.FromUint(2)

	c0 := ext2.mixedAPI.Mul(&x.A0, &x.A0) // x0^2
	c1 := ext2.mixedAPI.Mul(&x.A1, &x.A1) // x1^2
	c1 = ext2.mixedAPI.Mul(c1, extqnrE2)  // qnr*x1^2
	c0 = ext2.mixedAPI.Add(c0, c1)        // x0^2 + qnr*x1^2
	c1 = ext2.mixedAPI.Mul(extqnrE2, two) // 2*qnr
	c1 = ext2.mixedAPI.Mul(c1, &x.A0)     // 2*qnr*x0
	c1 = ext2.mixedAPI.Mul(c1, &x.A1)     // 2*qnr*x0*x1
	return &E2Gen[T]{
		A0: *c0, // x0^2 + qnr*x1^2
		A1: *c1, // 2*qnr*x0*x1
	}
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
func (ext2 *Ext2[T]) Div(e1, e2 *E2Gen[T]) *E2Gen[T] {
	return &E2Gen[T]{
		A0: *ext2.mixedAPI.Div(&e1.A0, &e2.A0),
		A1: *ext2.mixedAPI.Div(&e1.A1, &e2.A1),
	}
}

// DivByBase  Div e2 Element by a base elmt
func (ext2 *Ext2[T]) DivByBase(e1 *E2Gen[T], e2 *T) *E2Gen[T] {
	return &E2Gen[T]{
		A0: *ext2.mixedAPI.Div(&e1.A0, e2),
		A1: *ext2.mixedAPI.Div(&e1.A1, e2),
	}
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
