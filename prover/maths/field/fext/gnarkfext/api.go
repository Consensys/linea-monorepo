package gnarkfext

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// API is a wrapper of [frontend.API] with methods specialized for field
// extension operations.
type API struct {
	Inner frontend.API
}

// E2 element in a quadratic extension
type E2 struct {
	A0, A1 frontend.Variable
}

func NewZero() E2 {
	return E2{
		A0: 0,
		A1: 0,
	}
}

// SetOne returns a newly allocated element equal to 1
func One() E2 {
	return E2{
		A0: 1,
		A1: 0,
	}
}

// IsZero returns 1 if the element is equal to 0 and 0 otherwise
func (api *API) IsZero(e E2) frontend.Variable {
	return api.Inner.And(api.Inner.IsZero(e.A0), api.Inner.IsZero(e.A1))
}

// Neg negates a e2 elmt
func (api *API) Neg(e1 E2) E2 {
	return E2{
		A0: api.Inner.Sub(0, e1.A0),
		A1: api.Inner.Sub(0, e1.A1),
	}
}

// Add e2 elmts
func (api *API) Add(e1, e2 E2) E2 {
	return E2{
		A0: api.Inner.Add(e1.A0, e2.A0),
		A1: api.Inner.Add(e1.A1, e2.A1),
	}
}

// Double e2 elmt
func (api *API) Double(e1 E2) E2 {
	return E2{
		A0: api.Inner.Mul(e1.A0, 2),
		A1: api.Inner.Mul(e1.A1, 2),
	}
}

// Sub e2 elmts
func (api *API) Sub(e1, e2 E2) E2 {
	return E2{
		A0: api.Inner.Sub(e1.A0, e2.A0),
		A1: api.Inner.Sub(e1.A1, e2.A1),
	}
}

// Mul e2 elmts
func (api *API) Mul(x, y E2) E2 {

	a := api.Inner.Add(x.A0, x.A1)
	b := api.Inner.Add(y.A0, y.A1)
	a = api.Inner.Mul(a, b)

	b = api.Inner.Mul(x.A0, y.A0)
	c := api.Inner.Mul(x.A1, y.A1)

	return E2{
		A0: api.Inner.Sub(b, api.Inner.Mul(11, c)),
		A1: api.Inner.Sub(a, b, c),
	}
}

// Square e2 elt
func (api_ *API) Square(x E2) E2 {

	var a, b, c frontend.Variable
	api := api_.Inner

	a = api.Mul(2, x.A0, x.A1)

	c = api.Mul(x.A0, x.A0)
	b = api.Mul(x.A1, x.A1, 11)

	return E2{
		A0: api.Sub(c, b),
		A1: a,
	}
}

// MulByFp multiplies an fp2 elmt by an fp elmt
func (api *API) MulByBase(e1 E2, c frontend.Variable) E2 {
	return E2{
		A0: api.Inner.Mul(e1.A0, c),
		A1: api.Inner.Mul(e1.A1, c),
	}
}

// MulByNonResidue multiplies an fp2 elmt by the imaginary elmt
// ext.uSquare is the square of the imaginary root
func (api *API) MulByNonResidue(e1 E2) E2 {
	return E2{
		A0: api.Inner.Mul(e1.A1, -11),
		A1: e1.A0,
	}
}

// Conjugate conjugation of an e2 elmt
func (api *API) Conjugate(e1 E2) E2 {
	return E2{
		A0: e1.A0,
		A1: api.Inner.Sub(0, e1.A1),
	}
}

// Inverse e2 elmts
func (api_ *API) Inverse(x E2) E2 {

	api := api_.Inner

	a := x.A0 // creating the buffers a, b is faster than querying &x.A0, &x.A1 in the functions call below
	b := x.A1
	t0 := api.Mul(a, a)
	t1 := api.Mul(b, b)
	tmp := t1
	tmp = api.Mul(tmp, 11)
	t0 = api.Add(t0, tmp)
	t1 = api.Inverse(t0)

	return E2{
		A0: api.Mul(a, t1),
		A1: api.Mul(b, t1, -1),
	}
}

// Assign a value to self (witness assignment)
func Assign(a *fext.Element) E2 {
	return E2{
		A0: (fr.Element)(a.A0),
		A1: (fr.Element)(a.A1),
	}
}

// AssertIsEqual constraint self to be equal to other into the given constraint system
func (api *API) AssertIsEqual(e, other E2) {
	api.Inner.AssertIsEqual(e.A0, other.A0)
	api.Inner.AssertIsEqual(e.A1, other.A1)
}

// Select sets e to r1 if b=1, r2 otherwise
func (api *API) Select(b frontend.Variable, r1, r2 E2) E2 {
	return E2{
		A0: api.Inner.Select(b, r1.A0, r2.A0),
		A1: api.Inner.Select(b, r1.A1, r2.A1),
	}
}

// Lookup2 implements two-bit lookup. It returns:
//   - r1 if b1=0 and b2=0,
//   - r2 if b1=0 and b2=1,
//   - r3 if b1=1 and b2=0,
//   - r3 if b1=1 and b2=1.
func (api *API) Lookup2(b1, b2 frontend.Variable, r1, r2, r3, r4 E2) E2 {
	return E2{
		A0: api.Inner.Lookup2(b1, b2, r1.A0, r2.A0, r3.A0, r4.A0),
		A1: api.Inner.Lookup2(b1, b2, r1.A1, r2.A1, r3.A1, r4.A1),
	}
}
