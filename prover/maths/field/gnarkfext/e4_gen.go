package gnarkfext

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
)

// Element element in a quadratic extension
type E4Gen[T zk.Element] struct {
	B0, B1 E2Gen[T]
}

func NewE4Gen[T zk.Element](v fext.Element) E4Gen[T] {
	return E4Gen[T]{
		B0: NewE2Gen[T](v.B0),
		B1: NewE2Gen[T](v.B1),
	}
}

// Ext4 contains  the ext4 koalabear operations
type Ext4[T zk.Element] struct {
	mixedAPI zk.FieldOps[T]
	*Ext2[T]
}

func NewExt4[T zk.Element](api frontend.API) (*Ext4[T], error) {
	mixedAPI, err := zk.NewApi[T](api)
	if err != nil {
		return nil, err
	}
	ext2 := NewExt2[T](api)
	return &Ext4[T]{
		mixedAPI: mixedAPI,
		Ext2:     ext2,
	}, nil
}

// SetZero returns a newly allocated element equal to 0
func (ext4 *Ext4[T]) Zero() *E4Gen[T] {
	return &E4Gen[T]{
		B0: *ext4.Ext2.Zero(),
		B1: *ext4.Ext2.Zero(),
	}
}

// SetOne returns a newly allocated element equal to 1
func (ext4 *Ext4[T]) One() *E4Gen[T] {
	return &E4Gen[T]{
		B0: *ext4.Ext2.One(),
		B1: *ext4.Ext2.Zero(),
	}
}

// IsZero returns 1 if the element is equal to 0 and 0 otherwise
func (ext4 *Ext4[T]) IsZero(e *E4Gen[T]) frontend.Variable {
	return ext4.mixedAPI.And(
		ext4.Ext2.IsZero(&e.B0),
		ext4.Ext2.IsZero(&e.B1),
	)
}

func (ext4 *Ext4[T]) assign(e1 []*T) *E4Gen[T] {
	return &E4Gen[T]{
		B0: *ext4.Ext2.assign(e1[:2]),
		B1: *ext4.Ext2.assign(e1[2:4]),
	}
}

// Neg negates a Element elmt
func (ext4 *Ext4[T]) Neg(e1 *E4Gen[T]) *E4Gen[T] {
	return &E4Gen[T]{
		B0: *ext4.Ext2.Neg(&e1.B0),
		B1: *ext4.Ext2.Neg(&e1.B1),
	}
}

// Add Element elmts
func (ext4 *Ext4[T]) Add(e1, e2 *E4Gen[T]) *E4Gen[T] {
	return &E4Gen[T]{
		B0: *ext4.Ext2.Add(&e1.B0, &e2.B0),
		B1: *ext4.Ext2.Add(&e1.B1, &e2.B1),
	}
}

// e = e1+e2
func (ext4 *Ext4[T]) AddByBase(e1 *E4Gen[T], e2 *T) *E4Gen[T] {
	b0a0 := ext4.mixedAPI.Add(&e1.B0.A0, e2)
	return &E4Gen[T]{
		B0: E2Gen[T]{A0: *b0a0, A1: e1.B0.A1},
		B1: e1.B1,
	}
}

// Double Element elmt
func (ext4 *Ext4[T]) Double(e1 *E4Gen[T]) *E4Gen[T] {
	return &E4Gen[T]{
		B0: *ext4.Ext2.Double(&e1.B0),
		B1: *ext4.Ext2.Double(&e1.B1),
	}
}

// Sub Element elmts
func (ext4 *Ext4[T]) Sub(e1, e2 *E4Gen[T]) *E4Gen[T] {
	return &E4Gen[T]{
		B0: *ext4.Ext2.Sub(&e1.B0, &e2.B0),
		B1: *ext4.Ext2.Sub(&e1.B1, &e2.B1),
	}
}

// Mul e2 elmts, e = e1*e2
func (ext4 *Ext4[T]) Mul(e1, e2 *E4Gen[T], in ...*E4Gen[T]) *E4Gen[T] {

	l1 := ext4.Ext2.Add(&e1.B0, &e1.B1)
	l2 := ext4.Ext2.Add(&e2.B0, &e2.B1)

	u := ext4.Ext2.Mul(l1, l2)

	ac := ext4.Ext2.Mul(&e1.B0, &e2.B0)
	bd := ext4.Ext2.Mul(&e1.B1, &e2.B1)

	l31 := ext4.Ext2.Add(ac, bd)

	// l41.Mul(api, bd, ext.qnrElement)
	l41 := ext4.Ext2.MulByNonResidue(bd)
	e := &E4Gen[T]{
		B0: *ext4.Ext2.Add(ac, l41),
		B1: *ext4.Ext2.Sub(u, l31),
	}

	if len(in) > 0 {
		return ext4.Mul(e, in[0], in[1:]...)
	} else {
		return e
	}
}

// Square Element elt
func (ext4 *Ext4[T]) Square(x *E4Gen[T]) *E4Gen[T] {

	c0 := ext4.Ext2.Sub(&x.B0, &x.B1)
	c3 := ext4.Ext2.MulByNonResidue(&x.B1)
	c3 = ext4.Ext2.Sub(&x.B0, c3)
	c2 := ext4.Ext2.Mul(&x.B0, &x.B1)
	c0 = ext4.Ext2.Mul(c0, c3)
	c0 = ext4.Ext2.Add(c0, c2)
	c2 = ext4.Ext2.MulByNonResidue(c2)
	return &E4Gen[T]{
		B0: *ext4.Ext2.Add(c0, c2),
		B1: *ext4.Ext2.Double(c2),
	}
}

// MulByFp multiplies an fp2 elmt by an fp elmt
func (ext4 *Ext4[T]) MulByE2(e1 *E4Gen[T], c *E2Gen[T]) *E4Gen[T] {
	return &E4Gen[T]{
		B0: *ext4.Ext2.Mul(&e1.B0, c),
		B1: *ext4.Ext2.Mul(&e1.B1, c),
	}
}

// MulByFp multiplies an Fp4 elmt by an fp elmt
func (ext4 *Ext4[T]) MulByFp(e1 *E4Gen[T], c *T) *E4Gen[T] {
	return &E4Gen[T]{
		B0: *ext4.Ext2.MulByFp(&e1.B0, c),
		B1: *ext4.Ext2.MulByFp(&e1.B1, c),
	}
}

// Sum sets e = e1 + e2 + e3...
func (ext4 *Ext4[T]) Sum(e1 *E4Gen[T], e2 *E4Gen[T], e3 ...*E4Gen[T]) *E4Gen[T] {
	e := ext4.Add(e1, e2)
	for i := 0; i < len(e3); i++ {
		e = ext4.Add(e, e3[i])
	}
	return e
}

// AssertIsEqual asserts that e==e1
func (ext4 *Ext4[T]) AssertIsEqual(e, e1 *E4Gen[T]) {
	ext4.Ext2.AssertIsEqual(&e.B0, &e1.B0)
	ext4.Ext2.AssertIsEqual(&e.B1, &e1.B1)
}

// MulByNonResidue multiplies an fp2 elmt by the imaginary elmt
// ext.uSquare is the square of the imaginary root
func (ext4 *Ext4[T]) MulByNonResidue(e1 *E4Gen[T]) *E4Gen[T] {
	return &E4Gen[T]{
		B0: *ext4.Ext2.MulByNonResidue(&e1.B1),
		B1: e1.B0,
	}
}

// Conjugate conjugation of an Element elmt
func (ext4 *Ext4[T]) Conjugate(e1 E4Gen[T]) *E4Gen[T] {
	return &E4Gen[T]{
		B0: e1.B0,
		B1: *ext4.Ext2.Neg(&e1.B1),
	}
}

// Inverse Element elmts
func (ext4 *Ext4[T]) Inverse(e1 *E4Gen[T]) *E4Gen[T] {

	inverseHint := zk.MixedHint[T](_inverseE4)
	res, err := ext4.mixedAPI.NewHint(inverseHint, 4, &e1.B0.A0, &e1.B0.A1, &e1.B1.A0, &e1.B1.A1)
	if err != nil {
		// err is non-nil only for invalid number of inputs
		panic(err)
	}
	e3 := ext4.assign(res[:4])
	one := ext4.One()

	// 1 == e3 * e1
	_res := ext4.Mul(e3, e1)
	ext4.AssertIsEqual(_res, one)
	return e3
}

// Select sets e to r1 if b=1, r2 otherwise
func (ext4 *Ext4[T]) Select(b frontend.Variable, r1, r2 *E4Gen[T]) *E4Gen[T] {
	return &E4Gen[T]{
		B0: *ext4.Ext2.Select(b, &r1.B0, &r2.B0),
		B1: *ext4.Ext2.Select(b, &r1.B1, &r2.B1),
	}
}

// Div Element elmts
func (ext4 *Ext4[T]) Div(e1, e2 *E4Gen[T]) *E4Gen[T] {

	divHint := zk.MixedHint[T](_divE4)
	res, err := ext4.mixedAPI.NewHint(
		divHint, 4,
		&e1.B0.A0, &e1.B0.A1, &e1.B1.A0, &e1.B1.A1,
		&e2.B0.A0, &e2.B0.A1, &e2.B1.A0, &e2.B1.A1)
	if err != nil {
		// err is non nil only for invalid number of inputs
		panic(err)
	}
	e3 := ext4.assign(res[:4])
	_res := ext4.Mul(e3, e2)
	ext4.AssertIsEqual(_res, e1)
	return e3
}

// Sub Element elmts
func (ext4 *Ext4[T]) DivByBase(e1 *E4Gen[T], e2 *T) *E4Gen[T] {
	return &E4Gen[T]{
		B0: *ext4.Ext2.DivByBase(&e1.B0, e2),
		B1: *ext4.Ext2.DivByBase(&e1.B1, e2),
	}
}

// Exp exponentiation in gnark circuit, using the fast exponentiation
func (ext4 *Ext4[T]) ExpGen(x *E4Gen[T], n int) *E4Gen[T] {

	if n < 0 {
		xinv := ext4.Inverse(x)
		return ext4.ExpGen(xinv, -n)
	}

	if n == 0 {
		return ext4.One()
	}

	if n == 1 {
		return x
	}

	x2 := ext4.Mul(x, x)
	if n%2 == 0 {
		return ext4.ExpGen(x2, n/2)
	} else {
		res := ext4.ExpGen(x2, (n-1)/2)
		return ext4.Mul(res, x)
	}

}
