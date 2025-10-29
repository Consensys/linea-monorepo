// Package plonkbuilder defines gnark circuit builder implementing wide committer interface
package plonkbuilder

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
)

var _ StorerBuilderWideCommitter = &widecommitter{}

// PlonkInWizardBuilder is the interface for a builder that can store key-value pairs.
// This is the interface that the builder provided to [From] should implement.
type PlonkInWizardBuilder interface {
	// Builder is the underlying frontend builder
	frontend.Builder[constraint.U32]

	// redefiniton of gnark internal kvstore methods
	SetKeyValue(key, value any)
	GetKeyValue(key any) (value any)

	// GetWireConstraints and GetWiresConstraintExact retrieves the constraints
	// associated to the given wires. The first method deduplicates while the second
	// does not.
	GetWireConstraints(wires []frontend.Variable, addMissing bool) ([][2]int, error)
	GetWiresConstraintExact(wires []frontend.Variable, addMissing bool) ([][2]int, error)
}

type StorerBuilderWideCommitter interface {
	// StorerBuilder wraps the given builder
	PlonkInWizardBuilder
	// WideCommitter allows to commit to variable and obtain extension field element
	frontend.WideCommitter
}

type widecommitter struct {
	PlonkInWizardBuilder
}

// From creates a new builder implementing the [frontend.WideCommitter] interface
// from a base builder implementing [PlonkInWizardBuilder].
func From(newBaseBuilder frontend.NewBuilderU32) frontend.NewBuilderU32 {
	return func(field *big.Int, config frontend.CompileConfig) (frontend.Builder[constraint.U32], error) {
		baseBuilder, err := newBaseBuilder(field, config)
		if err != nil {
			return nil, fmt.Errorf("new base builder: %w", err)
		}
		scb, ok := baseBuilder.(PlonkInWizardBuilder)
		if !ok {
			return nil, fmt.Errorf("base builder doesn't implement StorerWideCommitterBuilder")
		}
		return &widecommitter{
			PlonkInWizardBuilder: scb,
		}, nil
	}
}

func (wc *widecommitter) WideCommit(width int, vars ...frontend.Variable) ([]frontend.Variable, error) {
	fmt.Println("not implemented: widecommitter.WideCommit called")
	ret := make([]frontend.Variable, width)
	for i := range width {
		ret[i] = i
	}
	return ret, nil
}

func PlaceholderWideCommitHint(field *big.Int, inputs []*big.Int, Outputs []*big.Int) error {
	panic("todo")
}
