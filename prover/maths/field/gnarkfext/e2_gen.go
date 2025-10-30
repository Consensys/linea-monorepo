package gnarkfext

import (
	"math/big"

	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
)

type Ext2 struct {
	apiGen zk.GenericApi
}

func NewExt2(api frontend.API) (*Ext2, error) {
	var res Ext2
	var err error
	res.apiGen, err = zk.NewGenericApi(api)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// 𝔽r²[u] = 𝔽r/u²-3

// E2Gen element in a quadratic extension
type E2Gen struct {
	A0, A1 zk.WrappedVariable
}

func NewE2Gen(v extensions.E2) E2Gen {
	return E2Gen{
		A0: zk.ValueOf(v.A0.String()),
		A1: zk.ValueOf(v.A1.String()),
	}
}

// SetZero returns a newly allocated element equal to 0
func (ext2 *Ext2) Zero() *E2Gen {
	return &E2Gen{
		A0: zk.ValueOf(0),
		A1: zk.ValueOf(0),
	}
}

// SetOne returns a newly allocated element equal to 1
func (ext2 *Ext2) One() *E2Gen {
	return &E2Gen{
		A0: zk.ValueOf(1),
		A1: zk.ValueOf(0),
	}
}

// IsZero returns 1 if the element is equal to 0 and 0 otherwise
func (ext2 *Ext2) IsZero(e *E2Gen) frontend.Variable {
	return ext2.apiGen.And(
		ext2.apiGen.IsZero(e.A0),
		ext2.apiGen.IsZero(e.A1),
	)
}

// Neg negates a E2Gen[T zk.FType] elmt
func (ext2 *Ext2) Neg(e1 *E2Gen) *E2Gen {
	zero := zk.ValueOf(0)
	return &E2Gen{
		A0: ext2.apiGen.Sub(zero, e1.A0),
		A1: ext2.apiGen.Sub(zero, e1.A1),
	}
}

// Add E2Gen[T zk.FType] elmts
func (ext2 *Ext2) Add(e1, e2 *E2Gen) *E2Gen {
	return &E2Gen{
		A0: ext2.apiGen.Add(e1.A0, e2.A0),
		A1: ext2.apiGen.Add(e1.A1, e2.A1),
	}
}

// Double E2Gen[T zk.FType] elmt
func (ext2 *Ext2) Double(e1 *E2Gen) *E2Gen {
	two := zk.ValueOf(2)
	return &E2Gen{
		A0: ext2.apiGen.Mul(e1.A0, two),
		A1: ext2.apiGen.Mul(e1.A1, two),
	}
}

// Sub E2Gen[T zk.FType] elmts
func (ext2 *Ext2) Sub(e1, e2 *E2Gen) *E2Gen {
	return &E2Gen{
		A0: ext2.apiGen.Sub(e1.A0, e2.A0),
		A1: ext2.apiGen.Sub(e1.A1, e2.A1),
	}
}

// Mul E2Gen[T zk.FType] elmts
func (ext2 *Ext2) Mul(e1, e2 *E2Gen) *E2Gen {

	l1 := ext2.apiGen.Add(e1.A0, e1.A1)
	l2 := ext2.apiGen.Add(e2.A0, e2.A1)

	u := ext2.apiGen.Mul(l1, l2)

	ac := ext2.apiGen.Mul(e1.A0, e2.A0)
	bd := ext2.apiGen.Mul(e1.A1, e2.A1)

	l31 := ext2.apiGen.Add(ac, bd)

	qnrE2 := zk.ValueOf(ext.qnrE2)
	l41 := ext2.apiGen.Mul(bd, qnrE2)

	return &E2Gen{
		A0: ext2.apiGen.Add(ac, l41),
		A1: ext2.apiGen.Sub(u, l31),
	}
}

func (ext2 *Ext2) Square(x *E2Gen) *E2Gen {
	// Square sets z to the E2-product of x,x returns z
	var res E2Gen
	a := ext2.apiGen.Mul(x.A0, x.A1)
	c := ext2.apiGen.Mul(x.A0, x.A0)
	b := ext2.apiGen.Mul(x.A1, x.A1)
	b = ext2.apiGen.MulConst(b, big.NewInt(3))
	res.A0 = ext2.apiGen.Add(c, b)
	res.A1 = ext2.apiGen.Add(a, a)
	return &res

}

// MulByFp multiplies an fp2 elmt by an fp elmt
func (ext2 *Ext2) MulByFp(e1 *E2Gen, c zk.WrappedVariable) *E2Gen {
	return &E2Gen{
		A0: ext2.apiGen.Mul(e1.A0, c),
		A1: ext2.apiGen.Mul(e1.A1, c),
	}
}

// MulByNonResidue multiplies an fp2 elmt by the imaginary elmt
// ext.uSquare is the square of the imaginary root
func (ext2 *Ext2) MulByNonResidue(e1 *E2Gen) *E2Gen {
	qnrE2 := zk.ValueOf(ext.qnrE2)
	return &E2Gen{
		A0: ext2.apiGen.Mul(e1.A1, qnrE2),
		A1: e1.A0,
	}
}

// Conjugate conjugation of an E2Gen[T zk.FType] elmt
func (ext2 *Ext2) Conjugate(e1 *E2Gen) *E2Gen {
	return &E2Gen{
		A0: e1.A0,
		A1: ext2.apiGen.Neg(e1.A1),
	}
}

func (ext2 *Ext2) assign(e1 []zk.WrappedVariable) *E2Gen {
	return &E2Gen{
		A0: e1[0],
		A1: e1[1],
	}
}

// Inverse E2Gen[T zk.FType] elmts
func (ext2 *Ext2) Inverse(e1 *E2Gen) *E2Gen {

	res, err := ext2.apiGen.NewHint(inverseE2Hint, 2, e1.A0, e1.A1)
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
// func (ext2 *Ext2) Assign(a extensions.E2) {
// 	e.A0 = *api.FromKoalabear(a.A0)
// 	e.A1 = *api.FromKoalabear(a.A1)
// }

// AssertIsEqual constraint self to be equal to other into the given constraint system
func (ext2 *Ext2) AssertIsEqual(e, other *E2Gen) {
	ext2.apiGen.AssertIsEqual(e.A0, other.A0)
	ext2.apiGen.AssertIsEqual(e.A1, other.A1)
}

// Select sets e to r1 if b=1, r2 otherwise
func (ext2 *Ext2) Select(b zk.WrappedVariable, r1, r2 *E2Gen) *E2Gen {
	return &E2Gen{
		A0: ext2.apiGen.Select(b, r1.A0, r2.A0),
		A1: ext2.apiGen.Select(b, r1.A1, r2.A1),
	}
}

// Div e2 elmts
func (ext2 *Ext2) Div(e1, e2 *E2Gen) *E2Gen {
	return &E2Gen{
		A0: ext2.apiGen.Div(e1.A0, e2.A0),
		A1: ext2.apiGen.Div(e1.A1, e2.A1),
	}
}

// DivByBase  Div e2 Element by a base elmt
func (ext2 *Ext2) DivByBase(e1 *E2Gen, e2 zk.WrappedVariable) *E2Gen {
	return &E2Gen{
		A0: ext2.apiGen.Div(e1.A0, e2),
		A1: ext2.apiGen.Div(e1.A1, e2),
	}
}
