package koalagnark

import (
	"math/big"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// koalabearModulus caches koalabear.Modulus() to avoid repeated allocations.
var koalabearModulus = koalabear.Modulus()

// -----------------------------------------------------------------------------
// Types
// -----------------------------------------------------------------------------

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

// Octuplet is an array of 8 Element values.
type Octuplet [8]Element

// -----------------------------------------------------------------------------
// Constructors
// -----------------------------------------------------------------------------
// NewElementFromBase creates an Element from a koalabear.Element for witness assignment.
func NewElementFromBase(v field.Element) Element {
	return NewElementFromValue(v.Uint64())
}

// NewElementFromValue creates an Element from an integer or big.Int for witness assignment.
func NewElementFromValue[T interface {
	int | int64 | uint32 | uint64 | *big.Int
}](v T) Element {
	var res Element
	res.EV = emulated.ValueOf[emulated.KoalaBear](v) // For constants (witness assignment) - uses ValueOf
	res.V = v
	return res
}

// -----------------------------------------------------------------------------
// Element Methods
// -----------------------------------------------------------------------------

// Native returns the native frontend.Variable representation.
// For uninitialized Elements (V=nil and EV.Limbs empty), returns 0.
// This allows circuit definition to proceed when elements are allocated but not yet assigned.
func (v *Element) Native() frontend.Variable {
	if v.V != nil {
		return v.V
	}
	if len(v.EV.Limbs) == 1 {
		return v.EV.Limbs[0]
	}
	// For uninitialized Elements (from make([]Element, n)), return 0 as placeholder.
	// Actual values are provided during witness assignment.
	if len(v.EV.Limbs) == 0 {
		return 0
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

// -----------------------------------------------------------------------------
// Octuplet Methods
// -----------------------------------------------------------------------------

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
