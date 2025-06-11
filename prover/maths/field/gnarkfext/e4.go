package gnarkfext

import (
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
)

//	𝔽r²[u] = 𝔽r/u²-3
//	𝔽r⁴[v] = 𝔽r²/v²-u

type Extension struct {
	qnrE2 koalabear.Element
	qnrE4 extensions.E2
}

func getKoalaBearExtension() Extension {
	res := Extension{}
	res.qnrE2.SetUint64(3)
	res.qnrE4.A1.SetOne()
	return res
}

var ext = getKoalaBearExtension()

//	𝔽r⁴[v] = 𝔽r²/v²-u

// Element element in a quadratic extension
type Element struct {
	B0, B1 E2
}

// SetZero returns a newly allocated element equal to 0
func (e *Element) SetZero() *Element {
	e.B0.SetZero()
	e.B1.SetZero()
	return e
}

// SetOne returns a newly allocated element equal to 1
func (e *Element) SetOne() *Element {
	e.B0.SetOne()
	e.B1.SetZero()
	return e
}

// IsZero returns 1 if the element is equal to 0 and 0 otherwise
func (e *Element) IsZero(api frontend.API) frontend.Variable {
	return api.And(e.B0.IsZero(api), e.B1.IsZero(api))
}

func (e *Element) assign(e1 []frontend.Variable) {
	e.B0.assign(e1[:2])
	e.B1.assign(e1[2:4])
}

// Neg negates a Element elmt
func (e *Element) Neg(api frontend.API, e1 Element) *Element {
	e.B0.Neg(api, e1.B0)
	e.B1.Neg(api, e1.B1)
	return e
}

// Add Element elmts
func (e *Element) Add(api frontend.API, e1, e2 Element) *Element {
	e.B0.Add(api, e1.B0, e2.B0)
	e.B1.Add(api, e1.B1, e2.B1)
	return e
}

// Double Element elmt
func (e *Element) Double(api frontend.API, e1 Element) *Element {
	e.B0.Double(api, e1.B0)
	e.B1.Double(api, e1.B1)
	return e
}

// Sub Element elmts
func (e *Element) Sub(api frontend.API, e1, e2 Element) *Element {
	e.B0.Sub(api, e1.B0, e2.B0)
	e.B1.Sub(api, e1.B1, e2.B1)
	return e
}

// Mul e2 elmts
func (e *Element) Mul(api frontend.API, e1, e2 Element, in ...Element) *Element {

	var l1, l2 E2
	l1.Add(api, e1.B0, e1.B1)
	l2.Add(api, e2.B0, e2.B1)

	var u E2
	u.Mul(api, l1, l2)

	var ac, bd E2
	ac.Mul(api, e1.B0, e2.B0)
	bd.Mul(api, e1.B1, e2.B1)

	var l31 E2
	l31.Add(api, ac, bd)
	e.B1.Sub(api, u, l31)

	var l41 E2
	// l41.Mul(api, bd, ext.qnrElement)
	l41.MulByNonResidue(api, bd)
	e.B0.Add(api, ac, l41)

	if len(in) > 0 {
		return e.Mul(api, *e, in[0], in[1:]...)
	} else {
		return e
	}
}

// Square Element elt
func (e *Element) Square(api frontend.API, x Element) *Element {

	var c0, c2, c3 E2
	c0.Sub(api, x.B0, x.B1)
	c3.MulByNonResidue(api, x.B1).Sub(api, x.B0, c3)
	c2.Mul(api, x.B0, x.B1)
	c0.Mul(api, c0, c3).Add(api, c0, c2)
	e.B1.Double(api, c2)
	c2.MulByNonResidue(api, c2)
	e.B0.Add(api, c0, c2)

	return e
}

// MulByFp multiplies an fp2 elmt by an fp elmt
func (e *Element) MulByE2(api frontend.API, e1 Element, c E2) *Element {
	e.B0.Mul(api, e1.B0, c)
	e.B1.Mul(api, e1.B1, c)
	return e
}

// MulByFp multiplies an Fp4 elmt by an fp elmt
func (e *Element) MulByFp(api frontend.API, e1 Element, c frontend.Variable) *Element {
	e.B0.MulByFp(api, e1.B0, c)
	e.B1.MulByFp(api, e1.B1, c)
	return e
}

// Sum sets e = e1 + e2 + e3...
func (e *Element) Sum(api frontend.API, e1 Element, e2 Element, e3 ...Element) *Element {
	e.Add(api, e1, e2)
	for i := 0; i < len(e3); i++ {
		e.Add(api, *e, e3[i])
	}
	return e
}

// AssertIsEqual asserts that e==e1
func (e *Element) AssertIsEqual(api frontend.API, e1 Element) {
	e.B0.AssertIsEqual(api, e1.B0)
	e.B1.AssertIsEqual(api, e1.B1)
}

// MulByNonResidue multiplies an fp2 elmt by the imaginary elmt
// ext.uSquare is the square of the imaginary root
func (e *Element) MulByNonResidue(api frontend.API, e1 Element) *Element {
	x := e1.B0
	e.B0.MulByNonResidue(api, e.B1)
	e.B1 = x
	return e
}

// Conjugate conjugation of an Element elmt
func (e *Element) Conjugate(api frontend.API, e1 Element) *Element {
	e.B0 = e1.B0
	e.B1.Neg(api, e1.B1)
	return e
}

// Inverse Element elmts
func (e *Element) Inverse(api frontend.API, e1 Element) *Element {

	res, err := api.NewHint(inverseE4Hint, 4, e1.B0.A0, e1.B0.A1, e1.B1.A0, e1.B1.A1)
	if err != nil {
		// err is non-nil only for invalid number of inputs
		panic(err)
	}

	var e3, one Element
	e3.assign(res[:4])
	one.SetOne()

	// 1 == e3 * e1
	e3.Mul(api, e3, e1)
	e3.AssertIsEqual(api, one)

	e.assign(res[:4])

	return e
}

// Assign a value to self (witness assignment)
func (e *Element) Assign(a fext.Element) {
	e.B0.Assign(a.B0)
	e.B1.Assign(a.B1)
}

// Select sets e to r1 if b=1, r2 otherwise
func (e *Element) Select(api frontend.API, b frontend.Variable, r1, r2 Element) *Element {

	e.B0.Select(api, b, r1.B0, r2.B0)
	e.B1.Select(api, b, r1.B1, r2.B1)

	return e
}

func NewFromBase(e frontend.Variable) Element {
	return Element{
		B0: E2{A0: e, A1: 0},

		B1: E2{A0: 0, A1: 0},
	}
}
