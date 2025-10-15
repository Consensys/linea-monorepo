package gnarkfext

import (
	"math/big"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
)

// Element element in a quadratic extension
type E4Gen struct {
	B0, B1 E2Gen
}

func NewE4Gen(v fext.Element) E4Gen {
	return E4Gen{
		B0: NewE2Gen(v.B0),
		B1: NewE2Gen(v.B1),
	}
}

func NewE4GenFromBase(v any) E4Gen {
	var res E4Gen
	res.B0.A0 = zk.ValueOf(v)
	res.B0.A1 = zk.ValueOf(0)
	res.B1.A0 = zk.ValueOf(0)
	res.B1.A1 = zk.ValueOf(0)
	return res
}

func FromBase(v zk.WrappedVariable) E4Gen {
	var res E4Gen
	res.B0.A0 = v
	res.B0.A1 = zk.ValueOf(0)
	res.B1.A0 = zk.ValueOf(0)
	res.B1.A1 = zk.ValueOf(0)
	return res
}

// Ext4 contains  the ext4 koalabear operations
type Ext4 struct {
	mixedAPI zk.GenericApi
	*Ext2
}

func NewExt4(api frontend.API) (*Ext4, error) {
	mixedAPI, err := zk.NewGenericApi(api)
	if err != nil {
		return nil, err
	}
	ext2, err := NewExt2(api)
	if err != nil {
		return nil, err
	}
	return &Ext4{
		mixedAPI: mixedAPI,
		Ext2:     ext2,
	}, nil
}

// SetZero returns a newly allocated element equal to 0
func (ext4 *Ext4) Zero() *E4Gen {
	return &E4Gen{
		B0: *ext4.Ext2.Zero(),
		B1: *ext4.Ext2.Zero(),
	}
}

// SetOne returns a newly allocated element equal to 1
func (ext4 *Ext4) One() *E4Gen {
	return &E4Gen{
		B0: *ext4.Ext2.One(),
		B1: *ext4.Ext2.Zero(),
	}
}

// IsZero returns 1 if the element is equal to 0 and 0 otherwise
func (ext4 *Ext4) IsZero(e *E4Gen) frontend.Variable {
	return ext4.mixedAPI.And(
		ext4.Ext2.IsZero(&e.B0),
		ext4.Ext2.IsZero(&e.B1),
	)
}

func (ext4 *Ext4) assign(e1 []*zk.WrappedVariable) *E4Gen {
	return &E4Gen{
		B0: *ext4.Ext2.assign(e1[:2]),
		B1: *ext4.Ext2.assign(e1[2:4]),
	}
}

// Neg negates a Element elmt
func (ext4 *Ext4) Neg(e1 *E4Gen) *E4Gen {
	return &E4Gen{
		B0: *ext4.Ext2.Neg(&e1.B0),
		B1: *ext4.Ext2.Neg(&e1.B1),
	}
}

// Add Element elmts
func (ext4 *Ext4) Add(e1, e2 *E4Gen) *E4Gen {
	return &E4Gen{
		B0: *ext4.Ext2.Add(&e1.B0, &e2.B0),
		B1: *ext4.Ext2.Add(&e1.B1, &e2.B1),
	}
}

// e = e1+e2
func (ext4 *Ext4) AddByBase(e1 *E4Gen, e2 *zk.WrappedVariable) *E4Gen {
	b0a0 := ext4.mixedAPI.Add(&e1.B0.A0, e2)
	return &E4Gen{
		B0: E2Gen{A0: *b0a0, A1: e1.B0.A1},
		B1: e1.B1,
	}
}

// Double Element elmt
func (ext4 *Ext4) Double(e1 *E4Gen) *E4Gen {
	return &E4Gen{
		B0: *ext4.Ext2.Double(&e1.B0),
		B1: *ext4.Ext2.Double(&e1.B1),
	}
}

// Sub Element elmts
func (ext4 *Ext4) Sub(e1, e2 *E4Gen) *E4Gen {
	return &E4Gen{
		B0: *ext4.Ext2.Sub(&e1.B0, &e2.B0),
		B1: *ext4.Ext2.Sub(&e1.B1, &e2.B1),
	}
}

// Mul e2 elmts, e = e1*e2
func (ext4 *Ext4) Mul(e1, e2 *E4Gen, in ...*E4Gen) *E4Gen {

	l1 := ext4.Ext2.Add(&e1.B0, &e1.B1)
	l2 := ext4.Ext2.Add(&e2.B0, &e2.B1)

	u := ext4.Ext2.Mul(l1, l2)

	ac := ext4.Ext2.Mul(&e1.B0, &e2.B0)
	bd := ext4.Ext2.Mul(&e1.B1, &e2.B1)

	l31 := ext4.Ext2.Add(ac, bd)

	// l41.Mul(api, bd, ext.qnrElement)
	l41 := ext4.Ext2.MulByNonResidue(bd)
	e := &E4Gen{
		B0: *ext4.Ext2.Add(ac, l41),
		B1: *ext4.Ext2.Sub(u, l31),
	}

	if len(in) > 0 {
		return ext4.Mul(e, in[0], in[1:]...)
	} else {
		return e
	}
}

// Square sets z=x*x in E4 and returns z
func (ext4 *Ext4) Square(x *E4Gen) *E4Gen {
	// same as mul, but we remove duplicate add and simplify multiplications with squaring
	// note: this is more efficient than Algorithm 22 from https://eprint.iacr.org/2010/354.pdf
	a := ext4.Ext2.Add(&x.B0, &x.B1)
	d := ext4.Ext2.Square(&x.B0)
	c := ext4.Ext2.Square(&x.B1)
	a = ext4.Ext2.Square(a)
	bc := ext4.Ext2.Add(d, c)
	var z E4Gen
	z.B1 = *ext4.Ext2.Sub(a, bc)
	z.B0 = *ext4.Ext2.MulByNonResidue(c)
	z.B0 = *ext4.Ext2.Add(&z.B0, d)

	return &z
}

// MulByFp multiplies an fp2 elmt by an fp elmt
func (ext4 *Ext4) MulByE2(e1 *E4Gen, c *E2Gen) *E4Gen {
	return &E4Gen{
		B0: *ext4.Ext2.Mul(&e1.B0, c),
		B1: *ext4.Ext2.Mul(&e1.B1, c),
	}
}

// MulByFp multiplies an Fp4 elmt by an fp elmt
func (ext4 *Ext4) MulByFp(e1 *E4Gen, c *zk.WrappedVariable) *E4Gen {
	return &E4Gen{
		B0: *ext4.Ext2.MulByFp(&e1.B0, c),
		B1: *ext4.Ext2.MulByFp(&e1.B1, c),
	}
}

// Sum sets e = e1 + e2 + e3...
func (ext4 *Ext4) Sum(e1 *E4Gen, e2 *E4Gen, e3 ...*E4Gen) *E4Gen {
	e := ext4.Add(e1, e2)
	for i := 0; i < len(e3); i++ {
		e = ext4.Add(e, e3[i])
	}
	return e
}

// AssertIsEqual asserts that e==e1
func (ext4 *Ext4) AssertIsEqual(e, e1 *E4Gen) {
	ext4.Ext2.AssertIsEqual(&e.B0, &e1.B0)
	ext4.Ext2.AssertIsEqual(&e.B1, &e1.B1)
}

// MulByNonResidue multiplies an fp2 elmt by the imaginary elmt
// ext.uSquare is the square of the imaginary root
func (ext4 *Ext4) MulByNonResidue(e1 *E4Gen) *E4Gen {
	return &E4Gen{
		B0: *ext4.Ext2.MulByNonResidue(&e1.B1),
		B1: e1.B0,
	}
}

// Conjugate conjugation of an Element elmt
func (ext4 *Ext4) Conjugate(e1 E4Gen) *E4Gen {
	return &E4Gen{
		B0: e1.B0,
		B1: *ext4.Ext2.Neg(&e1.B1),
	}
}

// Select sets e to r1 if b=1, r2 otherwise
func (ext4 *Ext4) Select(b zk.WrappedVariable, r1, r2 *E4Gen) *E4Gen {
	return &E4Gen{
		B0: *ext4.Ext2.Select(b, &r1.B0, &r2.B0),
		B1: *ext4.Ext2.Select(b, &r1.B1, &r2.B1),
	}
}

// Inverse Element elmts
func (ext4 *Ext4) Inverse(e1 *E4Gen) *E4Gen {

	invE4 := inverseE4Hint(ext4.apiGen.Type())

	res, err := ext4.mixedAPI.NewHint(invE4, 4, &e1.B0.A0, &e1.B0.A1, &e1.B1.A0, &e1.B1.A1)
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

// Div Element elmts
func (ext4 *Ext4) Div(e1, e2 *E4Gen) *E4Gen {

	divE4 := divE4Hint(ext4.apiGen.Type())

	res, err := ext4.mixedAPI.NewHint(
		divE4, 4,
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
func (ext4 *Ext4) DivByBase(e1 *E4Gen, e2 *zk.WrappedVariable) *E4Gen {
	return &E4Gen{
		B0: *ext4.Ext2.DivByBase(&e1.B0, e2),
		B1: *ext4.Ext2.DivByBase(&e1.B1, e2),
	}
}

// Exp exponentiation in gnark circuit, using the fast exponentiation
func (ext4 *Ext4) Exp(x *E4Gen, n *big.Int) *E4Gen {

	res := ext4.One()
	nBytes := n.Bytes()

	// TODO handle negative case
	for _, b := range nBytes {
		for j := 0; j < 8; j++ {
			c := (b >> (7 - j)) & 1
			res = ext4.Square(res)
			if c == 1 {
				res = ext4.Mul(res, x)
			}
		}
	}

	return res
}
