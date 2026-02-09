package koalagnark

import (
	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark/constraint/solver"
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

// NewExt creates an Ext for witness assignment from a degree-4
// extension field element.
func NewExt(v fext.Element) Ext {
	return Ext{B0: NewE2(v.B0), B1: NewE2(v.B1)}
}
