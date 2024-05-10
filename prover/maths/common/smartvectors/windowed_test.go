package smartvectors

import (
	"testing"

	"github.com/consensys/zkevm-monorepo/prover/maths/common/vector"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
)

// This is a simple error case we have faces in the past, the test ensures that
// it does go through.
func TestProcessWindowed(_ *testing.T) {

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
