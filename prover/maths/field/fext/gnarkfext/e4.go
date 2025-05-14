package gnarkfext

import (
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark/frontend"
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

// E4 element in a quadratic extension
type E4 struct {
	A0, A1 E2
}

// SetZero returns a newly allocated element equal to 0
func (e *E4) SetZero() *E4 {
	e.A0.SetZero()
	e.A1.SetZero()
	return e
}

// SetOne returns a newly allocated element equal to 1
func (e *E4) SetOne() *E4 {
	e.A0.SetOne()
	e.A1.SetZero()
	return e
}

// IsZero returns 1 if the element is equal to 0 and 0 otherwise
func (e *E4) IsZero(api frontend.API) frontend.Variable {
	return api.And(e.A0.IsZero(api), e.A1.IsZero(api))
}

func (e *E4) assign(e1 []frontend.Variable) {
	e.A0.assign(e1[:2])
	e.A1.assign(e1[2:4])
}

// Neg negates a E4 elmt
func (e *E4) Neg(api frontend.API, e1 E4) *E4 {
	e.A0.Neg(api, e1.A0)
	e.A1.Neg(api, e1.A1)
	return e
}

// Add E4 elmts
func (e *E4) Add(api frontend.API, e1, e2 E4) *E4 {
	e.A0.Add(api, e1.A0, e2.A0)
	e.A1.Add(api, e1.A1, e2.A1)
	return e
}

// Double E4 elmt
func (e *E4) Double(api frontend.API, e1 E4) *E4 {
	e.A0.Double(api, e1.A0)
	e.A1.Double(api, e1.A1)
	return e
}

// Sub E4 elmts
func (e *E4) Sub(api frontend.API, e1, e2 E4) *E4 {
	e.A0.Sub(api, e1.A0, e2.A0)
	e.A1.Sub(api, e1.A1, e2.A1)
	return e
}

// Mul e2 elmts
func (e *E4) Mul(api frontend.API, e1, e2 E4) *E4 {

	var l1, l2 E2
	l1.Add(api, e1.A0, e1.A1)
	l2.Add(api, e2.A0, e2.A1)

	var u E2
	u.Mul(api, l1, l2)

	var ac, bd E2
	ac.Mul(api, e1.A0, e2.A0)
	bd.Mul(api, e1.A1, e2.A1)

	var l31 E2
	l31.Add(api, ac, bd)
	e.A1.Sub(api, u, l31)

	var l41 E2
	// l41.Mul(api, bd, ext.qnrE4)
	l41.MulByNonResidue(api, bd)
	e.A0.Add(api, ac, l41)

	return e
}

// Square E4 elt
func (e *E4) Square(api frontend.API, x E4) *E4 {

	var c0, c1 E2
	c0.Mul(api, x.A0, x.A0)     // x0^2
	c1.Mul(api, x.A1, x.A1)     // x1^2
	c1.MulByNonResidue(api, c1) // qnr*x1^2
	c0.Add(api, c0, c1)         // x0^2 + qnr*x1^2
	c1.A0 = 2
	c1.MulByNonResidue(api, c1) // 2*qnr
	c1.Mul(api, c1, x.A0)       // 2*qnr*x0
	c1.Mul(api, c1, x.A1)       // 2*qnr*x0*x1
	e.A0 = c0                   // x0^2 + qnr*x1^2
	e.A1 = c1                   // 2*qnr*x0*x1

	return e
}

// MulByFp multiplies an fp2 elmt by an fp elmt
func (e *E4) MulByE2(api frontend.API, e1 E4, c E2) *E4 {
	e.A0.Mul(api, e1.A0, c)
	e.A1.Mul(api, e1.A1, c)
	return e
}

// MulByNonResidue multiplies an fp2 elmt by the imaginary elmt
// ext.uSquare is the square of the imaginary root
func (e *E4) MulByNonResidue(api frontend.API, e1 E4) *E4 {
	x := e1.A0
	e.A0.MulByNonResidue(api, e.A1)
	e.A1 = x
	return e
}

// Conjugate conjugation of an E4 elmt
func (e *E4) Conjugate(api frontend.API, e1 E4) *E4 {
	e.A0 = e1.A0
	e.A1.Neg(api, e1.A1)
	return e
}

// Inverse E4 elmts
func (e *E4) Inverse(api frontend.API, e1 E4) *E4 {

	res, err := api.NewHint(inverseE4Hint, 4, e1.A0.A0, e1.A0.A1, e1.A1.A0, e1.A1.A1)
	if err != nil {
		// err is non-nil only for invalid number of inputs
		panic(err)
	}

	var e3, one E4
	e3.assign(res[:4])
	one.SetOne()

	// 1 == e3 * e1
	e3.Mul(api, e3, e1)
	e3.AssertIsEqual(api, one)

	e.assign(res[:4])

	return e
}

// Assign a value to self (witness assignment)
func (e *E4) Assign(a extensions.E4) {
	e.A0.Assign(a.B0)
	e.A1.Assign(a.B1)
}

// AssertIsEqual constraint self to be equal to other into the given constraint system
func (e *E4) AssertIsEqual(api frontend.API, other E4) {
	api.AssertIsEqual(e.A0, other.A0)
	api.AssertIsEqual(e.A1, other.A1)
}

// Select sets e to r1 if b=1, r2 otherwise
func (e *E4) Select(api frontend.API, b frontend.Variable, r1, r2 E4) *E4 {

	e.A0.Select(api, b, r1.A0, r2.A0)
	e.A1.Select(api, b, r1.A1, r2.A1)

	return e
}
