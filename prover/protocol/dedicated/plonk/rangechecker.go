package plonk

import (
	"fmt"
	"math/big"

	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
)

// externalRangeChecker wraps the frontend.Builder. We require that the builder
// also implements [kvstore.Store] (internal to gnark, interface reimplemented
// locally) and [frontend.Committer].
//
// The range checking gadget in gnark works by checking the capabilities of the
// builder and if it provides `native` range checking capabilities (by
// implementing [frontend.Rangechecker]), uses it instead of doing range checks
// using a lookup table. By default the builder doesn't implement the interface
// and thus uses the fallback (using lookup tables). But within Wizard IOP we
// can defer range checking to the Wizard IOP instead of doing in PLONK circuit
// in gnark. For that, the [externalRangeChecker] implements
// [frontend.Rangechecker] by providing [externalRangeChecker.Check] method.
//
// Currently, the implementation is dummy as this wrapped builder doesn't
// actually pass the variables to range check on to Wizard IOP, but ideally it
// should. But this most probably requires that we tag the variables internally
// and then reorder them in the PLONK solver. There, we should probably mark
// these variables using a custom gate which allows later to map these variables
// into Wizard column which can be range checked.
type externalRangeChecker struct {
	storeCommitBuilder
	checked []frontend.Variable
	comp    *wizard.CompiledIOP
	rcCols  chan [2][]int
}

// storeCommitBuilder implements [frontend.Builder], [frontend.Committer] and
// [kvstore.Store].
type storeCommitBuilder interface {
	frontend.Builder
	frontend.Committer
	SetKeyValue(key, value any)
	GetKeyValue(key any) (value any)
	GetWireGates(wires []frontend.Variable) [2][]int
}

func NewExternalRangeChecker(field *big.Int, config frontend.CompileConfig) (frontend.Builder, error) {
	builder, rcGet := newExternalRangeChecker(nil)
	go rcGet()
	return builder(field, config)
}

// newExternalRangeChecker takes compiled IOP and returns [frontend.NewBuilder].
// The returned constructor can be passed to [frontend.Compile] to instantiate a
// new builder of constraint system.
//
// Example usage:
//
//	ccs, err := frontend.Compile(ecc.BN254.ScalarField, circuit, newExternalRangeChecker(comp))
func newExternalRangeChecker(comp *wizard.CompiledIOP) (frontend.NewBuilder, func() [2][]int) {
	rcCols := make(chan [2][]int)
	return func(field *big.Int, config frontend.CompileConfig) (frontend.Builder, error) {
			b, err := scs.NewBuilder(field, config)
			if err != nil {
				return nil, fmt.Errorf("new native builder: %w", err)
			}
			scb, ok := b.(storeCommitBuilder)
			if !ok {
				return nil, fmt.Errorf("native builder doesn't implement committer or kvstore")
			}
			return &externalRangeChecker{
				storeCommitBuilder: scb,
				comp:               comp,
				rcCols:             rcCols,
			}, nil
		}, func() [2][]int {
			return <-rcCols
		}
}

// Check implements [frontend.RangeChecker]
func (builder *externalRangeChecker) Check(v frontend.Variable, bits int) {
	// we store the ID of the wire we want to range check. Later, when calling
	// [Compile], we pass all the wires to the [GetWireGates] function of the
	// underlying builder to get the locations of the constraints
	builder.checked = append(builder.checked, v)
}

// Compile processes range checked variables and then calls Compile method of
// the underlying builder.
func (builder *externalRangeChecker) Compile() (constraint.ConstraintSystem, error) {
	go func() { builder.rcCols <- builder.storeCommitBuilder.GetWireGates(builder.checked) }()
	return builder.storeCommitBuilder.Compile()
}

// Compiler returns the compiler of the underlying builder.
func (builder *externalRangeChecker) Compiler() frontend.Compiler {
	return builder.storeCommitBuilder.Compiler()
}
