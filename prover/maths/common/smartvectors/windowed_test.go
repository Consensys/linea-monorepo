package smartvectors

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
)

func TestProcessWindowed(t *testing.T) {

	a := NewPaddedCircularWindow(
		vector.Rand(5),
		field.Zero(),
		0,
		16,
	)

	b := NewPaddedCircularWindow(
		vector.Rand(12),
		field.Zero(),
		4,
		16,
	)

	_, _ = processWindowedOnly(
		linCombOp{},
		[]SmartVector{b, a},
		[]int{1, 1},
	)
}
