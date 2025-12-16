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
	_ kVStore                          = &MockBuilder{}
	_ frontend.Builder[constraint.U32] = &MockBuilder{}
)

type kVStore interface {
	SetKeyValue(key, value any)
	GetKeyValue(key any) (value any)
}

// MockBuilder is a mock builder for testing purposes. It implements
// [frontend.WideCommitter] and [frontend.RangeChecker]. It works purely for
// U32.
type MockBuilder struct {
	// Builder is the underlying frontend builder
	frontend.Builder[constraint.U32]
	Store map[any]any
}

// NewMockBuilder returns a new MockBuilder.
func NewMockBuilder(wrapped frontend.NewBuilderU32) frontend.NewBuilderU32 {
	return func(field *big.Int, config frontend.CompileConfig) (frontend.Builder[constraint.U32], error) {
		builder, err := wrapped(field, config)
		if err != nil {
			return nil, err
		}
		return &MockBuilder{Builder: builder, Store: map[any]any{}}, nil
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

func (m *MockBuilder) SetKeyValue(key, value any) {
	m.Store[key] = value
}

func (m *MockBuilder) GetKeyValue(key any) (value any) {
	return m.Store[key]
}
