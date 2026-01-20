package gnarkutil

import (
	"math/big"

	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
)

var (
	// These are tests, that the interface implements frontend.WideCommitter and
	// frontend.RangeChecker.
	_ frontend.WideCommitter = &MockBuilder{}
	_ frontend.Rangechecker  = &MockBuilder{}
	_ minimalBuilder         = &MockBuilder{}
)

// MockBuilder is a mock builder for testing purposes. It implements
// [frontend.WideCommitter] and [frontend.RangeChecker]. It works purely for
// U32.
type MockBuilder struct {
	minimalBuilder
}

type minimalBuilder interface {
	// Builder is the underlying frontend builder
	frontend.Builder[constraint.U32]

	// redefinition of gnark internal kvstore methods
	SetKeyValue(key, value any)
	GetKeyValue(key any) (value any)
}

// NewMockBuilder returns a new MockBuilder.
func NewMockBuilder(wrapped frontend.NewBuilderU32) frontend.NewBuilderU32 {
	return func(field *big.Int, config frontend.CompileConfig) (frontend.Builder[constraint.U32], error) {
		builder, err := wrapped(field, config)
		if err != nil {
			return nil, err
		}
		mb, ok := builder.(minimalBuilder)
		if !ok {
			panic("wrapped builder does not implement minimalBuilder")
		}
		return &MockBuilder{minimalBuilder: mb}, nil
	}
}

// WideCommit returns a dummy static value for testing.
func (mb *MockBuilder) WideCommit(width int, toCommit ...frontend.Variable) (commitment []frontend.Variable, err error) {
	return mb.Compiler().NewHint(MockedWideCommiHint, width, toCommit...)
}

// Check implements the range-checker and does not do anything.
func (*MockBuilder) Check(v frontend.Variable, bits int) {}

func MockedWideCommiHint(mod *big.Int, inputs []*big.Int, outputs []*big.Int) error {
	for i := range outputs {
		outputs[i].Sub(mod, big.NewInt(int64(i))) // dummy value
	}
	return nil
}
