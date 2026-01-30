package koalagnark

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// koalabearModulus caches koalabear.Modulus() to avoid repeated allocations.
var koalabearModulus = koalabear.Modulus()

// Element represents a circuit variable over the KoalaBear base field.
// It abstracts over native and emulated representations, allowing the same
// circuit code to work in both native KoalaBear circuits and emulated circuits.
type Element struct {
	// V is used when the circuit's native field is KoalaBear.
	// Exported for gnark witness serialization.
	V frontend.Variable

	// EV is used when KoalaBear is emulated in another field.
	// Exported for gnark witness serialization.
	EV emulated.Element[emulated.KoalaBear]
}

// Octuplet is an array of 8 Var elements.
type Octuplet [8]Element

// NewElement creates a Var for witness assignment from any value.
// Use this when initializing circuit struct fields with witness values.
//
// For in-circuit constants, use [API.Const] instead.
func NewElement(v any) Element {
	switch v := v.(type) {
	case Element:
		return v
	case *Element:
		return *v
	}

	var res Element
	res.EV = emulated.ValueOf[emulated.KoalaBear](v)
	res.V = v
	return res
}

// NewElementFromKoala creates a Var from a field.Element for witness assignment.
func NewElementFromKoala(v field.Element) Element {
	return NewElement(v.Uint64())
}

// WrapFrontendVariable wraps an existing frontend.Variable as a Var.
// Use this for variables that come from other gnark APIs.
func WrapFrontendVariable(v frontend.Variable) Element {
	switch v.(type) {
	case Element, *Element:
		panic("attempted to wrap a koalagnark.Var into a koalagnark.Var")
	}

	var res Element
	res.V = v
	res.EV = emulated.Element[emulated.KoalaBear]{Limbs: []frontend.Variable{v}}
	return res
}

// Native returns the native frontend.Variable representation.
// Panics if the variable cannot be represented as a single native variable.
func (v *Element) Native() frontend.Variable {
	if v.V != nil {
		return v.V
	}
	if len(v.EV.Limbs) == 1 {
		return v.EV.Limbs[0]
	}
	utils.Panic("unexpected shape for Var: %++v", v)
	return nil // unreachable
}

// Emulated returns the emulated element representation.
func (v *Element) Emulated() *emulated.Element[emulated.KoalaBear] {
	return &v.EV
}

// IsEmpty returns true if the Var has not been initialized.
func (v *Element) IsEmpty() bool {
	return v.V == nil && len(v.EV.Limbs) == 0
}

// Initialize prepares the Var for the given field modulus.
func (v *Element) Initialize(modulus *big.Int) {
	if modulus.Cmp(koalabearModulus) == 0 {
		return
	}
	v.EV.Initialize(modulus)
}

// NativeArray converts an Octuplet to an array of native frontend.Variables.
func (o Octuplet) NativeArray() [8]frontend.Variable {
	res := [8]frontend.Variable{}
	for i := range res {
		res[i] = o[i].Native()
		if res[i] == nil {
			panic("Var is nil")
		}
	}
	return res
}

// In case the
var (
	zeroKoalaFr = field.Zero()
	oneKoalaFr  = field.One()
	oneBigInt   = big.NewInt(1)
)

// ConstantValueOf returns true if the variable represent a constant value and
// returns the non-nil value if so.
func (api API) ConstantValueOfElement(v Element) (*field.Element, bool) {

	var res field.Element

	if api.IsNative() {

		if v.V == nil {
			panic("unexpected, api is native but not the field element")
		}

		f, ok := api.nativeAPI.Compiler().ConstantValue(v.V)
		if !ok {
			return nil, false
		}

		if f.Sign() == 0 {
			return &zeroKoalaFr, true
		}

		if f.Cmp(oneBigInt) == 0 {
			return &oneKoalaFr, true
		}

		res.SetBigInt(f)
		return &res, true
	}

	var (
		// @alex: unfortunately we can't get the native field size from the
		// emulated api so we have to hardcode this part. Fortunately, this
		// should be easy enough to maintain in case this changes.
		nbLimb, nbBit = emulated.GetEffectiveFieldParams[emulated.KoalaBear](ecc.BLS12_377.ScalarField())
		fBig          big.Int
		f             field.Element
	)

	if int(nbLimb) != len(v.EV.Limbs) {
		utils.Panic("field contains %v limbs, but the emulated API suggests it contains %v", len(v.EV.Limbs), nbLimb)
	}

	for i := range v.EV.Limbs {
		g, ok := api.nativeAPI.Compiler().ConstantValue(v.EV.Limbs[i])
		if !ok {
			return nil, false
		}

		if i > 0 {
			fBig.Lsh(&fBig, uint(nbBit))
		}
		fBig.Add(&fBig, g)
	}

	f.SetBigInt(&fBig)
	return &f, true
}
