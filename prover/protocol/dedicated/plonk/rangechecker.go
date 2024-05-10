package plonk

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
)

// Compile-time sanity check the satisfaction of the interface RangeChecker by
// externalRangeChecker
var _ frontend.Rangechecker = &externalRangeChecker{}

// externalRangeChecker wraps the frontend.Builder. We require that the builder
// also implements [frontend.Committer].
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
	checked              []frontend.Variable
	comp                 *wizard.CompiledIOP
	rcCols               chan [][2]int
	addGateForRangeCheck bool
}

// storeCommitBuilder implements [frontend.Builder], [frontend.Committer] and
// [kvstore.Store].
type storeCommitBuilder interface {
	frontend.Builder
	frontend.Committer
	SetKeyValue(key, value any)
	GetKeyValue(key any) (value any)
	GetWireConstraints(wires []frontend.Variable, addMissing bool) ([][2]int, error)
}

// newExternalRangeChecker takes compiled IOP and returns [frontend.NewBuilder].
// The returned constructor can be passed to [frontend.Compile] to instantiate a
// new builder of constraint system.
//
// The function also returns an rcGetter which is in substance a function to be
// called after the compilation to return the position of the wires that are
// range-checked in the circuit.
//
// Example usage:
//
//	```
//	builder, rcGetter := newExternalRangeChecker(comp)
//	ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField, circuit, builder)
//	if err != nil {
//		return fmt.Errorf("could not compile because: %w", err)
//	}
//
//	// This returns the position of the wires to range-check.
//	checkedWires := rcGetter()
//	```
func newExternalRangeChecker(comp *wizard.CompiledIOP, addGateForRangeCheck bool) (frontend.NewBuilder, func() [][2]int) {
	rcCols := make(chan [][2]int)
	return func(field *big.Int, config frontend.CompileConfig) (frontend.Builder, error) {
			b, err := scs.NewBuilder(field, config)
			if err != nil {
				return nil, fmt.Errorf("could not create new native builder: %w", err)
			}
			scb, ok := b.(storeCommitBuilder)
			if !ok {
				return nil, fmt.Errorf("native builder doesn't implement committer or kvstore")
			}
			return &externalRangeChecker{
				storeCommitBuilder:   scb,
				comp:                 comp,
				rcCols:               rcCols,
				addGateForRangeCheck: addGateForRangeCheck,
			}, nil
		}, func() [][2]int {
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
	// GetWireGates may add gates if [addGateForRangeCheck] is true. Call it
	// synchronously before calling compile on the circuit.
	cols, err := builder.storeCommitBuilder.GetWireConstraints(builder.checked, builder.addGateForRangeCheck)
	if err != nil {
		return nil, fmt.Errorf("get wire gates: %w", err)
	}
	// we pass the result in a goroutine until the wizard compiler is ready to receive it
	go func() {
		builder.rcCols <- cols
	}()
	return builder.storeCommitBuilder.Compile()
}

// Compiler returns the compiler of the underlying builder.
func (builder *externalRangeChecker) Compiler() frontend.Compiler {
	return builder.storeCommitBuilder.Compiler()
}
