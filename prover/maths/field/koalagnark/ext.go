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
		A0: NewElement(v.A0),
		A1: NewElement(v.A1),
	}
}

// NewExt creates an Ext for witness assignment from various input types:
//   - fext.Element: full 4-component extension element
//   - extensions.E2: quadratic extension in B0, B1 is zero
//   - Element, field.Element: base field element in B0.A0, all others zero
//   - int, int64, uint32, *big.Int: numeric constants in B0.A0
//   - string: decimal representation of a field element in B0.A0
func NewExt(v any) Ext {
	// Pre-compute zero values to avoid repeated allocations
	zero := NewElement(0)
	zE2 := E2{A0: zero, A1: zero}

	switch v := v.(type) {
	case Ext:
		return v
	case *Ext:
		return *v
	case fext.Element:
		return Ext{B0: NewE2(v.B0), B1: NewE2(v.B1)}
	case *fext.Element:
		return Ext{B0: NewE2(v.B0), B1: NewE2(v.B1)}
	case extensions.E2:
		return Ext{B0: NewE2(v), B1: zE2}
	case *extensions.E2:
		return Ext{B0: NewE2(*v), B1: zE2}
	case Element:
		return Ext{B0: E2{A0: v, A1: zero}, B1: zE2}
	case *Element:
		return Ext{B0: E2{A0: *v, A1: zero}, B1: zE2}

	// Lift base field elements and numeric types to extension by placing them in B0.A0
	case field.Element:
		return Ext{B0: E2{A0: NewElement(v), A1: zero}, B1: zE2}
	case *field.Element:
		return Ext{B0: E2{A0: NewElement(v), A1: zero}, B1: zE2}
	case int:
		return Ext{B0: E2{A0: NewElement(v), A1: zero}, B1: zE2}
	case int64:
		return Ext{B0: E2{A0: NewElement(v), A1: zero}, B1: zE2}
	case uint32:
		return Ext{B0: E2{A0: NewElement(v), A1: zero}, B1: zE2}
	case *big.Int:
		return Ext{B0: E2{A0: NewElement(v), A1: zero}, B1: zE2}
	case string:
		return Ext{B0: E2{A0: NewElement(v), A1: zero}, B1: zE2}

	default:
		panic("NewExt: unsupported type")
	}
}
