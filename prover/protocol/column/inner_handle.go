package column

import (
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/gnark/frontend"
)

const (
	NATURAL     string = "NATURAL"
	REPEAT      string = "REPEAT"
	SHIFT       string = "SHIFT"
	INTERLEAVED string = "INTERLEAVED"
	// Generalizes the concept of Natural to
	// the case of verifier defined columns
	NONCOMPOSITE string = "NONCOMPOSITE"
)

// A handle represents a commitment (or derivative) that is stored
// in the store. By derivative, we mean "commitment X but we shifted
// the value" (see Shifted) by one or "only the values occuring at
// some regular intervals" etc..
type InnerHandle interface {
	// Returns the size of the referenced commitment
	Size() int
	// String representation of the handle
	GetColID() ifaces.ColID
	// Returns true if the handle is registered. This is trivial
	// by design (because Natural handle objects are built by the
	// function that registers it). The goal of this function is to
	// assert this fact. Precisely, it will check if a corresponding
	// entry in the store exists. If it does not, it panics.
	MustExists()
	// Returns the round of registration of the handle
	Round() int
	// Fetches a witness from the store
	GetColAssignment(run ifaces.Runtime) ifaces.ColAssignment
	// Fetches a witness from the circuit. This will panic if the
	// handle depends on a private column.
	GetWitnessGnark(run ifaces.Runtime) []frontend.Variable
	// Is composite
	IsComposite() bool
}
