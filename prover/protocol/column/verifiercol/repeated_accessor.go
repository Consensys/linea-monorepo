package verifiercol

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// RepeatedAccessor is a [VerifierCol] implementation that is built by
// repeated Size times the same accessor. It will not work if the provided
// accessor require "api" to be evaluated in a gnark circuit.
type RepeatedAccessor struct {
	Accessor ifaces.Accessor
	ColSize  int
}

// NewRepeatedAccessor instantiates a [RepeatedAccessor] column from an
// [ifaces.Accessor] and a size.
func NewRepeatedAccessor(accessor ifaces.Accessor, size int) ifaces.Column {
	return RepeatedAccessor{Accessor: accessor, ColSize: size}
}

// Round returns the round ID of the column and implements the [ifaces.Column]
// interface.
func (f RepeatedAccessor) Round() int {
	return f.Accessor.Round()
}

// GetColID returns the column ID
func (f RepeatedAccessor) GetColID() ifaces.ColID {
	return ifaces.ColIDf("REPEATED_%v_SIZE=%v", f.Accessor.Name(), f.ColSize)
}

// MustExists implements the [ifaces.Column] interface and always returns true.
func (f RepeatedAccessor) MustExists() {}

// Size returns the size of the colum and implements the [ifaces.Column]
// interface.
func (f RepeatedAccessor) Size() int {
	return f.ColSize
}

// Split implements the [VerifierCol] interface
func (f RepeatedAccessor) Split(_ *wizard.CompiledIOP, from, to int) ifaces.Column {
	return NewRepeatedAccessor(f.Accessor, to-from)
}

// GetColAssignment returns the assignment of the current column
func (f RepeatedAccessor) GetColAssignment(run ifaces.Runtime) ifaces.ColAssignment {
	return smartvectors.NewConstant(f.Accessor.GetVal(run), f.Size())
}

// GetColAssignmentGnark returns a gnark assignment of the current column
func (f RepeatedAccessor) GetColAssignmentGnark(run ifaces.GnarkRuntime) []frontend.Variable {
	res := make([]frontend.Variable, f.Size())
	x := f.Accessor.GetFrontendVariable(nil, run)
	for i := range res {
		res[i] = x
	}
	return res
}

// GetColAssignmentAt returns a particular position of the column
func (f RepeatedAccessor) GetColAssignmentAt(run ifaces.Runtime, pos int) field.Element {
	return f.Accessor.GetVal(run)
}

// GetColAssignmentGnarkAt returns a particular position of the column in a gnark circuit
func (f RepeatedAccessor) GetColAssignmentGnarkAt(run ifaces.GnarkRuntime, pos int) frontend.Variable {
	return f.Accessor.GetFrontendVariable(nil, run)
}

// IsComposite implements the [ifaces.Column] interface
func (f RepeatedAccessor) IsComposite() bool {
	return false
}

// String implements the [symbolic.Metadata] interface
func (f RepeatedAccessor) String() string {
	return f.Accessor.Name()
}
