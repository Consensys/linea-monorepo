package koalagnark

import (
	"math/big"

	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
)

// Register hints at package initialization
func init() {
	solver.RegisterHint(
		inverseE2Hint,
		inverseExtHintNative, inverseExtHintEmulated,
		divExtHintNative, divExtHintEmulated,
		mulExtHintNative, mulExtHintEmulated)
}

// -----------------------------------------------------------------------------
// Types
// -----------------------------------------------------------------------------

// E2 is a quadratic extension element.
// It represents an element of F_p^2 = F_p[u] / (u^2 - 3).
type E2 struct {
	A0, A1 Element
}

// Ext is a circuit variable over the degree-4 extension field.
// It represents an element of F_p^4 = F_p^2[v] / (v^2 - u).
type Ext struct {
	B0, B1 E2
}

// -----------------------------------------------------------------------------
// Witness Assignment Constructors
// -----------------------------------------------------------------------------
//
// Use these functions to create Ext values for witness assignment (outside Define).
// They use emulated.ValueOf internally, which has lazy limb initialization.
// For in-circuit constants, use the API methods (ExtFrom, ZeroExt, etc.) instead.

// NewE2 creates an E2 from extensions.E2 for witness assignment.
func NewE2(v extensions.E2) E2 {
	return E2{
		A0: NewElementFromBase(v.A0),
		A1: NewElementFromBase(v.A1),
	}
}

// NewExtFromExt creates an Ext for witness assignment from a degree-4
// extension field element.
func NewExtFromExt(v fext.Element) Ext {
	return Ext{B0: NewE2(v.B0), B1: NewE2(v.B1)}
}

// NewExtFromBase creates an Ext for witness assignment from a base field
// element, embedding it in B0.A0 with all other components set to zero.
func NewExtFromBase(v field.Element) Ext {
	return LiftToExt(NewElementFromBase(v))
}

// NewExtFromValue creates an Ext for witness assignment from a numeric
// constant, embedding it in B0.A0 with all other components set to zero.
func NewExtFromValue[T interface {
	int | int64 | uint32 | *big.Int
}](v T) Ext {
	return LiftToExt(NewElementFromValue(v))
}

// LiftToExt promotes an already-constructed Element into the extension field,
// placing it in B0.A0 with all other components set to zero. Unlike the
// NewExtFrom* constructors, this accepts an in-circuit Element and preserves
// it as-is without re-wrapping.
func LiftToExt(v Element) Ext {
	zero := NewElementFromValue(0)
	return Ext{B0: E2{A0: v, A1: zero}, B1: E2{A0: zero, A1: zero}}
}
