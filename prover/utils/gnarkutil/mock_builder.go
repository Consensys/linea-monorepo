package gnarkutil

import (
	"math/big"

	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
)

var (
	// These are tests, that the interface implements frontend.WideCommitter and
	// frontend.RangeChecker.
	_ frontend.WideCommitter           = &MockBuilder{}
	_ frontend.Rangechecker            = &MockBuilder{}
	_ frontend.Builder[constraint.U32] = &MockBuilder{}
)

// MockBuilder is a mock builder for testing purposes. It implements
// [frontend.WideCommitter] and [frontend.RangeChecker]. It works purely for
// U32.
type MockBuilder struct {
	// Builder is the underlying frontend builder
	frontend.Builder[constraint.U32]
}

// NewMockBuilder returns a new MockBuilder.
func NewMockBuilder(wrapped frontend.NewBuilderU32) frontend.NewBuilderU32 {

	return func(field *big.Int, config frontend.CompileConfig) (frontend.Builder[constraint.U32], error) {
		builder, err := wrapped(field, config)
		if err != nil {
			return nil, err
		}
		return &MockBuilder{Builder: builder}, nil
	}
}

// WideCommit returns a dummy static value for testing.
func (*MockBuilder) WideCommit(width int, toCommit ...frontend.Variable) (commitment []frontend.Variable, err error) {
	res := make([]frontend.Variable, width)
	for i := range res {
		if i < len(toCommit) {
			res[i] = toCommit[i]
			continue
		}
		res[i] = i
	}
	return res, nil
}

// Check implements the range-checker and does not do anything.
func (*MockBuilder) Check(v frontend.Variable, bits int) {}
