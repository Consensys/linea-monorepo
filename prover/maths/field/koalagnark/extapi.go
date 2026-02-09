package koalagnark

import (
	"math/big"

	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
)

// -----------------------------------------------------------------------------
// Ext Constants (in-circuit)
// -----------------------------------------------------------------------------

// ZeroExt returns the additive identity in the extension field.
func (a *API) ZeroExt() Ext {
	z := a.Zero()
	return Ext{B0: E2{A0: z, A1: z}, B1: E2{A0: z, A1: z}}
}

// OneExt returns the multiplicative identity in the extension field.
func (a *API) OneExt() Ext {
	z, o := a.Zero(), a.One()
	return Ext{B0: E2{A0: o, A1: z}, B1: E2{A0: z, A1: z}}
}

// ConstExt creates an Ext element from an fext.Element constant.
func (a *API) ConstExt(v fext.Element) Ext {
	return Ext{
		B0: E2{
			A0: a.Const(v.B0.A0),
			A1: a.Const(v.B0.A1),
		},
		B1: E2{
			A0: a.Const(v.B1.A0),
			A1: a.Const(v.B1.A1),
		},
	}
}

// ConstExtFromBase creates an Ext element from a base field constant, lifting it to the extension.
func (a *API) ConstExtFromBase(v field.Element) Ext {
	return a.LiftToExt(a.Const(v))
}

// ConstExtFromUint64 creates an Ext element from a uint64 constant, lifting it to the extension.
func (a *API) ConstExtFromUint64(v uint64) Ext {
	return a.LiftToExt(a.ConstFromUint64(v))
}

// LiftToExt promotes an already-constructed Element into the extension field
func (a *API) LiftToExt(v Element) Ext {
	z := a.Zero()
	zE2 := E2{A0: z, A1: z}
	return Ext{B0: E2{A0: v, A1: z}, B1: zE2}
}

// -----------------------------------------------------------------------------
// Ext Arithmetic Operations
// -----------------------------------------------------------------------------

// AddExt returns x + y in the extension field.
func (a *API) AddExt(x, y Ext) Ext {
	return Ext{
		B0: E2{A0: a.Add(x.B0.A0, y.B0.A0), A1: a.Add(x.B0.A1, y.B0.A1)},
		B1: E2{A0: a.Add(x.B1.A0, y.B1.A0), A1: a.Add(x.B1.A1, y.B1.A1)},
	}
}

// SubExt returns x - y in the extension field.
func (a *API) SubExt(x, y Ext) Ext {
	return Ext{
		B0: E2{A0: a.Sub(x.B0.A0, y.B0.A0), A1: a.Sub(x.B0.A1, y.B0.A1)},
		B1: E2{A0: a.Sub(x.B1.A0, y.B1.A0), A1: a.Sub(x.B1.A1, y.B1.A1)},
	}
}

// NegExt returns -x in the extension field.
func (a *API) NegExt(x Ext) Ext {
	z := a.Zero()
	return Ext{
		B0: E2{A0: a.Sub(z, x.B0.A0), A1: a.Sub(z, x.B0.A1)},
		B1: E2{A0: a.Sub(z, x.B1.A0), A1: a.Sub(z, x.B1.A1)},
	}
}

// DoubleExt returns 2*x in the extension field.
func (a *API) DoubleExt(x Ext) Ext {
	two := big.NewInt(2)
	return Ext{
		B0: E2{A0: a.MulConst(x.B0.A0, two), A1: a.MulConst(x.B0.A1, two)},
		B1: E2{A0: a.MulConst(x.B1.A0, two), A1: a.MulConst(x.B1.A1, two)},
	}
}

// qnrE2 is the non-residue constant for E2 extension (value 3).
// Used with MulConst to avoid unnecessary range checks.
var qnrE2 = big.NewInt(3)

// e2MulByNonResidue multiplies an E2 by the non-residue u (where u^2 = 3).
// Returns (3*a1, a0).
func (a *API) e2MulByNonResidue(x E2) E2 {
	return E2{
		A0: a.MulConst(x.A1, qnrE2),
		A1: x.A0,
	}
}

// e2Add returns x + y in E2.
func (a *API) e2Add(x, y E2) E2 {
	return E2{A0: a.Add(x.A0, y.A0), A1: a.Add(x.A1, y.A1)}
}

// e2Sub returns x - y in E2.
func (a *API) e2Sub(x, y E2) E2 {
	return E2{A0: a.Sub(x.A0, y.A0), A1: a.Sub(x.A1, y.A1)}
}

// e2Mul returns x * y in E2 using Karatsuba.
// (a0 + a1*u) * (b0 + b1*u) where u^2 = 3
func (a *API) e2Mul(x, y E2) E2 {
	l1 := a.Add(x.A0, x.A1)
	l2 := a.Add(y.A0, y.A1)
	u := a.Mul(l1, l2)      // (a0+a1)(b0+b1)
	ac := a.Mul(x.A0, y.A0) // a0*b0
	bd := a.Mul(x.A1, y.A1) // a1*b1

	sum := a.Add(ac, bd)
	a1 := a.Sub(u, sum) // (a0+a1)(b0+b1) - a0*b0 - a1*b1

	bd3 := a.MulConst(bd, qnrE2) // 3*a1*b1 (since u^2 = 3)
	a0 := a.Add(ac, bd3)

	return E2{A0: a0, A1: a1}
}

// e2Square returns x^2 in E2.
func (a *API) e2Square(x E2) E2 {
	a0sq := a.Mul(x.A0, x.A0)
	a1sq := a.Mul(x.A1, x.A1)
	a1sq3 := a.MulConst(a1sq, qnrE2)
	cross := a.Mul(x.A0, x.A1)
	return E2{
		A0: a.Add(a0sq, a1sq3),
		A1: a.MulConst(cross, big.NewInt(2)),
	}
}

// e2MulByFp multiplies an E2 by a base field element.
func (a *API) e2MulByFp(x E2, c Element) E2 {
	return E2{
		A0: a.Mul(x.A0, c),
		A1: a.Mul(x.A1, c),
	}
}

// e2MulConst multiplies an E2 by a constant.
func (a *API) e2MulConst(x E2, c *big.Int) E2 {
	return E2{
		A0: a.MulConst(x.A0, c),
		A1: a.MulConst(x.A1, c),
	}
}

// MulExt returns x * y in the extension field using Karatsuba.
// (B0 + B1*v) * (C0 + C1*v) where v^2 = u
func (a *API) MulExt(x, y Ext, more ...*Ext) Ext {
	l1 := a.e2Add(x.B0, x.B1)
	l2 := a.e2Add(y.B0, y.B1)
	u := a.e2Mul(l1, l2)      // (B0+B1)(C0+C1)
	ac := a.e2Mul(x.B0, y.B0) // B0*C0
	bd := a.e2Mul(x.B1, y.B1) // B1*C1

	sum := a.e2Add(ac, bd)
	b1 := a.e2Sub(u, sum) // (B0+B1)(C0+C1) - B0*C0 - B1*C1

	bdNR := a.e2MulByNonResidue(bd)
	b0 := a.e2Add(ac, bdNR)

	result := Ext{B0: b0, B1: b1}

	if len(more) > 0 {
		return a.MulExt(result, *more[0], more[1:]...)
	}
	return result
}

// SquareExt returns x^2 in the extension field.
func (a *API) SquareExt(x Ext) Ext {
	sum := a.e2Add(x.B0, x.B1)
	d := a.e2Square(x.B0)
	c := a.e2Square(x.B1)
	sum = a.e2Square(sum)
	bc := a.e2Add(d, c)
	b1 := a.e2Sub(sum, bc)
	cNR := a.e2MulByNonResidue(c)
	b0 := a.e2Add(cNR, d)
	return Ext{B0: b0, B1: b1}
}

// MulByE2Ext multiplies an Ext by an E2 element.
func (a *API) MulByE2Ext(x Ext, c E2) Ext {
	return Ext{
		B0: a.e2Mul(x.B0, c),
		B1: a.e2Mul(x.B1, c),
	}
}

// MulByFpExt multiplies an Ext by a base field element.
func (a *API) MulByFpExt(x Ext, c Element) Ext {
	return Ext{
		B0: a.e2MulByFp(x.B0, c),
		B1: a.e2MulByFp(x.B1, c),
	}
}

// MulConstExt multiplies an Ext by a constant.
func (a *API) MulConstExt(x Ext, c *big.Int) Ext {
	return Ext{
		B0: a.e2MulConst(x.B0, c),
		B1: a.e2MulConst(x.B1, c),
	}
}

// ModReduceExt reduces an Ext element (no-op in native mode).
func (a *API) ModReduceExt(x Ext) Ext {
	if a.IsNative() {
		// in native mode, no reduction is necessary
		return x
	}
	return Ext{
		B0: E2{
			A0: a.ModReduce(x.B0.A0),
			A1: a.ModReduce(x.B0.A1),
		},
		B1: E2{
			A0: a.ModReduce(x.B1.A0),
			A1: a.ModReduce(x.B1.A1),
		},
	}
}

// AddByBaseExt adds a base field element to the constant term.
func (a *API) AddByBaseExt(x Ext, y Element) Ext {
	return Ext{
		B0: E2{A0: a.Add(x.B0.A0, y), A1: x.B0.A1},
		B1: x.B1,
	}
}

// SumExt returns the sum of all input Ext elements.
func (a *API) SumExt(xs ...Ext) Ext {
	n := len(xs)
	b0a0 := make([]Element, n)
	b0a1 := make([]Element, n)
	b1a0 := make([]Element, n)
	b1a1 := make([]Element, n)

	for i, x := range xs {
		b0a0[i] = x.B0.A0
		b0a1[i] = x.B0.A1
		b1a0[i] = x.B1.A0
		b1a1[i] = x.B1.A1
	}

	return Ext{
		B0: E2{A0: a.Sum(b0a0...), A1: a.Sum(b0a1...)},
		B1: E2{A0: a.Sum(b1a0...), A1: a.Sum(b1a1...)},
	}
}

// MulByNonResidueExt multiplies by the non-residue v (where v^2 = u).
func (a *API) MulByNonResidueExt(x Ext) Ext {
	return Ext{
		B0: a.e2MulByNonResidue(x.B1),
		B1: x.B0,
	}
}

// ConjugateExt returns the conjugate of x.
func (a *API) ConjugateExt(x Ext) Ext {
	return Ext{
		B0: x.B0,
		B1: E2{A0: a.Neg(x.B1.A0), A1: a.Neg(x.B1.A1)},
	}
}

// -----------------------------------------------------------------------------
// Ext Comparison and Selection
// -----------------------------------------------------------------------------

// IsConstantZeroExt returns true if e is a constant zero element.
func (api *API) IsConstantZeroExt(e Ext) bool {
	return api.IsConstantZero(e.B0.A0) &&
		api.IsConstantZero(e.B0.A1) &&
		api.IsConstantZero(e.B1.A0) &&
		api.IsConstantZero(e.B1.A1)
}

// SelectExt returns x if sel=1, y otherwise.
func (a *API) SelectExt(sel frontend.Variable, x, y Ext) Ext {
	return Ext{
		B0: E2{
			A0: a.Select(sel, x.B0.A0, y.B0.A0),
			A1: a.Select(sel, x.B0.A1, y.B0.A1),
		},
		B1: E2{
			A0: a.Select(sel, x.B1.A0, y.B1.A0),
			A1: a.Select(sel, x.B1.A1, y.B1.A1),
		},
	}
}

// AssertIsEqualExt constrains x == y.
func (a *API) AssertIsEqualExt(x, y Ext) {
	a.AssertIsEqual(x.B0.A0, y.B0.A0)
	a.AssertIsEqual(x.B0.A1, y.B0.A1)
	a.AssertIsEqual(x.B1.A0, y.B1.A0)
	a.AssertIsEqual(x.B1.A1, y.B1.A1)
}

// -----------------------------------------------------------------------------
// Ext Division and Inverse
// -----------------------------------------------------------------------------

// InverseExt returns 1/x in the extension field.
func (a *API) InverseExt(x Ext) Ext {
	hint := a.inverseExtHint()
	res, err := a.NewHint(hint, 4, x.B0.A0, x.B0.A1, x.B1.A0, x.B1.A1)
	if err != nil {
		panic(err)
	}
	inv := a.extFromVars(res)

	// Verify: x * inv == 1
	product := a.MulExt(x, inv)
	a.AssertIsEqualExt(product, a.OneExt())
	return inv
}

// DivExt returns x / y in the extension field.
func (a *API) DivExt(x, y Ext) Ext {
	hint := a.divExtHint()
	res, err := a.NewHint(hint, 4,
		x.B0.A0, x.B0.A1, x.B1.A0, x.B1.A1,
		y.B0.A0, y.B0.A1, y.B1.A0, y.B1.A1)
	if err != nil {
		panic(err)
	}
	quot := a.extFromVars(res)

	// Verify: y * quot == x
	product := a.MulExt(y, quot)
	a.AssertIsEqualExt(product, x)
	return quot
}

// DivByBaseExt divides an Ext by a base field element.
func (a *API) DivByBaseExt(x Ext, y Element) Ext {
	return Ext{
		B0: E2{A0: a.Div(x.B0.A0, y), A1: a.Div(x.B0.A1, y)},
		B1: E2{A0: a.Div(x.B1.A0, y), A1: a.Div(x.B1.A1, y)},
	}
}

// extFromVars creates an Ext from 4 Vars.
func (a *API) extFromVars(v []Element) Ext {
	return Ext{
		B0: E2{A0: v[0], A1: v[1]},
		B1: E2{A0: v[2], A1: v[3]},
	}
}

// -----------------------------------------------------------------------------
// Ext Exponentiation
// -----------------------------------------------------------------------------

// ExpExt computes x^n using square-and-multiply.
// Optimized for power-of-two exponents.
func (a *API) ExpExt(x Ext, n *big.Int) Ext {
	if n.Sign() == 0 {
		return a.OneExt()
	}
	if n.Cmp(big.NewInt(1)) == 0 {
		return x
	}

	// Fast path: power of two (only squaring needed)
	if n.BitLen() > 0 && new(big.Int).And(n, new(big.Int).Sub(n, big.NewInt(1))).Sign() == 0 {
		res := x
		for i := 0; i < n.BitLen()-1; i++ {
			res = a.SquareExt(res)
		}
		return res
	}

	// General case: square-and-multiply
	res := a.OneExt()
	nBytes := n.Bytes()
	for _, b := range nBytes {
		for j := 0; j < 8; j++ {
			c := (b >> (7 - j)) & 1
			res = a.SquareExt(res)
			if c == 1 {
				res = a.MulExt(res, x)
			}
		}
	}
	return res
}

// ExpVariableExponentExt computes x^exp where exp is a circuit variable.
func (a *API) ExpVariableExponentExt(x Ext, exp frontend.Variable, expNumBits int) Ext {
	expBits := a.nativeAPI.ToBinary(exp, expNumBits)
	res := a.OneExt()

	for i := len(expBits) - 1; i >= 0; i-- {
		if i != len(expBits)-1 {
			res = a.SquareExt(res)
		}
		tmp := a.MulExt(res, x)
		res = a.SelectExt(expBits[i], tmp, res)
	}
	return res
}

// -----------------------------------------------------------------------------
// Ext Debug
// -----------------------------------------------------------------------------

// PrintlnExt prints Ext variables for debugging.
func (a *API) PrintlnExt(vars ...Ext) {
	for i := range vars {
		a.Println(vars[i].B0.A0, vars[i].B0.A1, vars[i].B1.A0, vars[i].B1.A1)
	}
}

// -----------------------------------------------------------------------------
// Ext Hints
// -----------------------------------------------------------------------------

// NewHintExt calls a hint function with Ext inputs and outputs.
func (a *API) NewHintExt(f solver.Hint, nbOutputs int, inputs ...Ext) ([]Ext, error) {
	if a.IsNative() {
		flatInputs := make([]frontend.Variable, 4*len(inputs))
		for i, r := range inputs {
			flatInputs[4*i] = r.B0.A0.Native()
			flatInputs[4*i+1] = r.B0.A1.Native()
			flatInputs[4*i+2] = r.B1.A0.Native()
			flatInputs[4*i+3] = r.B1.A1.Native()
		}
		flatRes, err := a.nativeAPI.NewHint(f, 4*nbOutputs, flatInputs...)
		if err != nil {
			return nil, err
		}
		res := make([]Ext, nbOutputs)
		for i := range res {
			res[i] = Ext{
				B0: E2{A0: Element{V: flatRes[4*i]}, A1: Element{V: flatRes[4*i+1]}},
				B1: E2{A0: Element{V: flatRes[4*i+2]}, A1: Element{V: flatRes[4*i+3]}},
			}
		}
		return res, nil
	}

	flatInputs := make([]*emulated.Element[emulated.KoalaBear], 4*len(inputs))
	for i, r := range inputs {
		flatInputs[4*i] = r.B0.A0.Emulated()
		flatInputs[4*i+1] = r.B0.A1.Emulated()
		flatInputs[4*i+2] = r.B1.A0.Emulated()
		flatInputs[4*i+3] = r.B1.A1.Emulated()
	}
	flatRes, err := a.emulatedAPI.NewHint(f, 4*nbOutputs, flatInputs...)
	if err != nil {
		return nil, err
	}
	res := make([]Ext, nbOutputs)
	for i := range res {
		res[i] = Ext{
			B0: E2{A0: Element{EV: *flatRes[4*i]}, A1: Element{EV: *flatRes[4*i+1]}},
			B1: E2{A0: Element{EV: *flatRes[4*i+2]}, A1: Element{EV: *flatRes[4*i+3]}},
		}
	}
	return res, nil
}

// -----------------------------------------------------------------------------
// Hint Implementations (private)
// -----------------------------------------------------------------------------

func inverseE2Hint(_ *big.Int, inputs []*big.Int, res []*big.Int) error {
	var a, c extensions.E2
	a.A0.SetBigInt(inputs[0])
	a.A1.SetBigInt(inputs[1])
	c.Inverse(&a)
	c.A0.BigInt(res[0])
	c.A1.BigInt(res[1])
	return nil
}

func inverseExtHintNative(_ *big.Int, inputs []*big.Int, res []*big.Int) error {
	var a, c fext.Element
	a.B0.A0.SetBigInt(inputs[0])
	a.B0.A1.SetBigInt(inputs[1])
	a.B1.A0.SetBigInt(inputs[2])
	a.B1.A1.SetBigInt(inputs[3])
	c.Inverse(&a)
	c.B0.A0.BigInt(res[0])
	c.B0.A1.BigInt(res[1])
	c.B1.A0.BigInt(res[2])
	c.B1.A1.BigInt(res[3])
	return nil
}

func inverseExtHintEmulated(_ *big.Int, inputs []*big.Int, output []*big.Int) error {
	return emulated.UnwrapHint(inputs, output, inverseExtHintNative)
}

func (a *API) inverseExtHint() solver.Hint {
	if a.IsNative() {
		return inverseExtHintNative
	}
	return inverseExtHintEmulated
}

func divExtHintNative(_ *big.Int, inputs []*big.Int, res []*big.Int) error {
	var x, y, c fext.Element
	x.B0.A0.SetBigInt(inputs[0])
	x.B0.A1.SetBigInt(inputs[1])
	x.B1.A0.SetBigInt(inputs[2])
	x.B1.A1.SetBigInt(inputs[3])
	y.B0.A0.SetBigInt(inputs[4])
	y.B0.A1.SetBigInt(inputs[5])
	y.B1.A0.SetBigInt(inputs[6])
	y.B1.A1.SetBigInt(inputs[7])
	c.Div(&x, &y)
	c.B0.A0.BigInt(res[0])
	c.B0.A1.BigInt(res[1])
	c.B1.A0.BigInt(res[2])
	c.B1.A1.BigInt(res[3])
	return nil
}

func divExtHintEmulated(_ *big.Int, inputs []*big.Int, output []*big.Int) error {
	return emulated.UnwrapHint(inputs, output, divExtHintNative)
}

func (a *API) divExtHint() solver.Hint {
	if a.IsNative() {
		return divExtHintNative
	}
	return divExtHintEmulated
}

func mulExtHintNative(_ *big.Int, inputs []*big.Int, res []*big.Int) error {
	var x, y, c fext.Element
	x.B0.A0.SetBigInt(inputs[0])
	x.B0.A1.SetBigInt(inputs[1])
	x.B1.A0.SetBigInt(inputs[2])
	x.B1.A1.SetBigInt(inputs[3])
	y.B0.A0.SetBigInt(inputs[4])
	y.B0.A1.SetBigInt(inputs[5])
	y.B1.A0.SetBigInt(inputs[6])
	y.B1.A1.SetBigInt(inputs[7])
	c.Mul(&x, &y)
	c.B0.A0.BigInt(res[0])
	c.B0.A1.BigInt(res[1])
	c.B1.A0.BigInt(res[2])
	c.B1.A1.BigInt(res[3])
	return nil
}

func mulExtHintEmulated(_ *big.Int, inputs []*big.Int, output []*big.Int) error {
	return emulated.UnwrapHint(inputs, output, mulExtHintNative)
}

// -----------------------------------------------------------------------------
// Ext Utility Functions
// -----------------------------------------------------------------------------

// BaseValueOfElement returns the base field element if e represents one.
// Returns (element, true) if B0.A1, B1.A0, and B1.A1 are all constant zeros,
// otherwise returns (nil, false).
func (api *API) BaseValueOfElement(e Ext) (*Element, bool) {
	if !api.IsConstantZero(e.B0.A1) ||
		!api.IsConstantZero(e.B1.A0) ||
		!api.IsConstantZero(e.B1.A1) {
		return nil, false
	}
	return &e.B0.A0, true
}
