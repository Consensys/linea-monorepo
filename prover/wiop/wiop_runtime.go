package wiop

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/koalabear/field"
)

// Runtime is the execution context for protocol [Action]s. It holds column
// assignments, cell values, coin values, and an arbitrary state bag. A single
// Runtime serves both prover-side (assign) and verifier-side (read) usage
// through the same [Action] interface.
//
// Pass Runtime by value: all mutable storage lives in map fields (reference
// types), so map mutations made inside an [Action] propagate to the caller.
// The sole exception is [Runtime.AdvanceRound], which must be called on a
// pointer to update [Runtime.currentRound].
type Runtime struct {
	// System is the protocol specification this Runtime executes against.
	System *System
	// currentRound is the round currently being processed.
	currentRound *Round
	// columns maps each Column to its concrete vector assignment.
	columns map[*Column]*ConcreteVector
	// cells maps each Cell to its concrete scalar value.
	cells map[*Cell]field.Element
	// coins maps each CoinField to its sampled value.
	coins map[*CoinField]field.Element
	// state is a free-form key-value store for stateful actions.
	state map[string]any
}

// NewRuntime creates a fresh Runtime for sys. currentRound is initialised to
// the first interactive round if one exists, or nil otherwise. Precomputed
// column assignments from [System.PrecomputedRound] are pre-loaded.
func NewRuntime(sys *System) Runtime {
	run := Runtime{
		System:  sys,
		columns: make(map[*Column]*ConcreteVector),
		cells:   make(map[*Cell]field.Element),
		coins:   make(map[*CoinField]field.Element),
		state:   make(map[string]any),
	}
	if len(sys.Rounds) > 0 {
		run.currentRound = sys.Rounds[0]
	}
	pr := sys.PrecomputedRound
	for i, col := range pr.Columns {
		run.columns[col] = pr.PrecomputedValues[i]
	}
	return run
}

// CurrentRound returns the round currently being processed, or nil if the
// system has no interactive rounds.
func (run Runtime) CurrentRound() *Round { return run.currentRound }

// AdvanceRound moves the runtime to the next interactive round. Panics if
// there is no next round.
//
// TODO: perform Fiat-Shamir coin sampling once the hash layer is implemented.
func (run *Runtime) AdvanceRound() {
	if run.currentRound == nil {
		panic("wiop: AdvanceRound: system has no interactive rounds")
	}
	next, ok := run.currentRound.Next()
	if !ok {
		panic(fmt.Sprintf(
			"wiop: AdvanceRound: already at the last round (id=%d)",
			run.currentRound.ID,
		))
	}
	run.currentRound = next
}

// AssignColumn stores a concrete vector assignment for col. Panics if col does
// not belong to the current round or has already been assigned.
func (run Runtime) AssignColumn(col *Column, v *ConcreteVector) {
	if col.round != run.currentRound {
		panic(fmt.Sprintf(
			"wiop: AssignColumn: column %q belongs to round %d but current round is %v",
			col.Context.Path(), col.round.ID, run.currentRound,
		))
	}
	if _, exists := run.columns[col]; exists {
		panic(fmt.Sprintf(
			"wiop: AssignColumn: column %q already assigned",
			col.Context.Path(),
		))
	}
	run.columns[col] = v
}

// GetColumnAssignment returns the concrete assignment of col. Panics if col
// has not been assigned yet.
func (run Runtime) GetColumnAssignment(col *Column) *ConcreteVector {
	v, ok := run.columns[col]
	if !ok {
		panic(fmt.Sprintf(
			"wiop: GetColumnAssignment: column %q is not assigned",
			col.Context.Path(),
		))
	}
	return v
}

// HasColumnAssignment reports whether col has been assigned in this runtime.
func (run Runtime) HasColumnAssignment(col *Column) bool {
	_, ok := run.columns[col]
	return ok
}

// AssignCell stores a concrete scalar value for cell. Panics if cell does not
// belong to the current round or has already been assigned.
func (run Runtime) AssignCell(cell *Cell, v field.Element) {
	if cell.round != run.currentRound {
		panic(fmt.Sprintf(
			"wiop: AssignCell: cell %q belongs to round %d but current round is %v",
			cell.Context.Path(), cell.round.ID, run.currentRound,
		))
	}
	if _, exists := run.cells[cell]; exists {
		panic(fmt.Sprintf(
			"wiop: AssignCell: cell %q already assigned",
			cell.Context.Path(),
		))
	}
	run.cells[cell] = v
}

// GetCellValue returns the concrete scalar value of cell. Panics if cell has
// not been assigned yet.
func (run Runtime) GetCellValue(cell *Cell) field.Element {
	v, ok := run.cells[cell]
	if !ok {
		panic(fmt.Sprintf(
			"wiop: GetCellValue: cell %q is not assigned",
			cell.Context.Path(),
		))
	}
	return v
}

// InsertCoin records a sampled value for coin. Intended for use by the round
// advancement logic; not for direct use inside [Action]s. Panics if the coin
// value has already been set.
func (run Runtime) InsertCoin(coin *CoinField, v field.Element) {
	if _, exists := run.coins[coin]; exists {
		panic(fmt.Sprintf(
			"wiop: InsertCoin: coin %q already set",
			coin.Context.Path(),
		))
	}
	run.coins[coin] = v
}

// GetCoinValue returns the sampled value of coin. Panics if the coin has not
// been inserted yet.
func (run Runtime) GetCoinValue(coin *CoinField) field.Element {
	v, ok := run.coins[coin]
	if !ok {
		panic(fmt.Sprintf(
			"wiop: GetCoinValue: coin %q has not been sampled",
			coin.Context.Path(),
		))
	}
	return v
}

// GetState returns the value stored under key and whether it was present.
func (run Runtime) GetState(key string) (any, bool) {
	v, ok := run.state[key]
	return v, ok
}

// SetState stores value under key in the runtime's state bag.
func (run Runtime) SetState(key string, value any) {
	run.state[key] = value
}
