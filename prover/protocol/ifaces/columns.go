package ifaces

import (
	"fmt"
	"reflect"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/symbolic"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/gnark/frontend"
)

// ID we can give to a commitment. The custom type is here
// to prevent mixing up the queries and the commitment.
type ColID string

// Witness specifies a value we can give to a commitment
// Implicitly all of our commitments are vectorial/polynomials/
type ColAssignment = smartvectors.SmartVector

// Utility function to format ColIDs
func ColIDf(s string, args ...interface{}) ColID {
	return ColID(fmt.Sprintf(s, args...))
}

// A column represents a commitment (or derivative) that is stored
// in the store. By derivative, we mean "commitment X but we shifted the value"
// (see Shifted) by one or "only the values occuring at some regular intervals"
// etc..
type Column interface {
	// Allows using a column as a symbolic variable
	symbolic.Metadata
	// Provides basic utilities for a protocol object
	Round() int
	// Returns the size of the referenced commitment
	Size() int
	// String representation of the column
	GetColID() ColID
	// Returns true if the column is registered. This is trivial
	// by design (because Natural column objects are built by the
	// function that registers it). The goal of this function is to
	// assert this fact. Precisely, it will check if a corresponding
	// entry in the store exists. If it does not, it panics.
	MustExists()
	// Fetches a ColAssignment from the store
	GetColAssignment(run Runtime) ColAssignment
	// Fetches a ColAssignment from the store
	GetColAssignmentAt(run Runtime, pos int) field.Element
	// Fetches a ColAssignment from the circuit. This will panic if the column
	// depends on a private column.
	GetColAssignmentGnark(run GnarkRuntime) []frontend.Variable
	// Fetches a ColAssignment from the circuit. This will panic if the column
	// depends on a private column.
	GetColAssignmentGnarkAt(run GnarkRuntime, pos int) frontend.Variable
	// Is composite
	IsComposite() bool
}

/*
Instantiate a symbolic variable from a handle
*/
func ColumnAsVariable(h Column) *symbolic.Expression {
	return symbolic.NewVariable(h)
}

/*
Assert the round of registration of the commitment
*/
func MustBeInRound(h Column, round int) {
	if round != h.Round() {
		utils.Panic("wrong round assertion for %v, expected %v but was %v", h.GetColID(), round, h.Round())
	}
}

/*
Assert that the handle is natural
*/
func AssertNotComposite(h Column) {
	if h.IsComposite() {
		utils.Panic(
			"expected a natural handle, but got %v. You probably forgot to add"+
				"a `univariates.Naturalize` step the compilation process",
			reflect.TypeOf(h),
		)
	}
}
