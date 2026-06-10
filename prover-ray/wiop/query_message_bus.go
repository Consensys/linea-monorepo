package wiop

import "fmt"

// BusDirection is the sign of a [MessageBus] entry's contribution to its
// (segment, handle) accumulator.
type BusDirection int

const (
	// BusSend marks an entry that ADDS its rows to the (segment, handle)
	// accumulator. Built by [System.NewMessageBusSend].
	BusSend BusDirection = iota
	// BusReceive marks an entry that SUBTRACTS its rows (weighted by an optional
	// multiplicity) from the (segment, handle) accumulator. Built by
	// [System.NewMessageBusReceive].
	BusReceive
)

// String returns a human-readable label for d, used in diagnostics.
func (d BusDirection) String() string {
	switch d {
	case BusSend:
		return "Send"
	case BusReceive:
		return "Receive"
	default:
		return fmt.Sprintf("BusDirection(%d)", int(d))
	}
}

// MessageBus is a [Query] declaring that one [Table] participates in a
// (Segment, Handle)-keyed log-up accumulator. Each instance is the unit of
// participation: one Send entry adds its rows to the accumulator; one Receive
// entry subtracts them.
//
// Semantics. Two coins α and β (extension field, shared across every
// participant of every Handle reduced together by the [messagebus] compiler)
// are drawn after every participant column is committed. For an entry with
// column views (c_0, …, c_{w-1}), define the row-folding
//
//	d(row) = β + α^{w-1}·c_0(row) + … + α·c_{w-2}(row) + c_{w-1}(row)
//
// and a per-row filter equal to Tab.Selector when present and 1 otherwise.
// The entry contributes
//
//	Send:    +Σ_row filter(row) · 1 / d(row)
//	Receive: -Σ_row filter(row) · Multiplicity(row) / d(row)         (Multiplicity = 1 if nil)
//
// to the (Segment, Handle) accumulator. The [messagebus] compiler lowers all
// entries with the same Segment/Handle into a single [LogDerivativeSum] (one
// Cell per (Segment, Handle) holding the running sum) and emits a verifier
// action that asserts, for each Handle, the cells across every Segment sum to
// zero. By Schwartz–Zippel over α, β, that identity holds iff the multiset of
// rows sent into the Handle equals the multiset of rows received from it
// (with the multiplicity weighting on the Receive side).
//
// MessageBus does not implement [GnarkCheckableQuery] nor [AssignableQuery]:
// its semantics span multiple queries within a system and are discharged by
// the dedicated compiler pass. [MessageBus.Check] is a no-op for that reason —
// the per-Handle multiset identity is enforced after compilation by the
// LogDerivativeSum and the verifier action.
//
// Use [System.NewMessageBusSend] and [System.NewMessageBusReceive] to
// construct instances.
type MessageBus struct {
	baseQuery
	// Segment is the segment identifier (e.g., a region of the trace) used,
	// together with Handle, to key the accumulator the entry contributes to.
	// Always non-empty.
	Segment string
	// Handle is the bus name. Entries with the same Handle are summed together
	// (across every Segment) by the compiler. Always non-empty.
	Handle string
	// Direction selects the sign of the contribution. See [BusDirection].
	Direction BusDirection
	// Tab is the column tuple (with an optional selector) being sent or
	// received. The Selector field of Tab acts as a per-row filter and may be
	// nil.
	Tab Table
	// Multiplicity is an optional row-weighting expression applied on the
	// Receive side only. Always nil when Direction == BusSend; nil is also
	// permitted on the Receive side and is treated as the constant 1.
	//
	// If non-nil and vector-valued, Multiplicity must share the same module as
	// Tab. A scalar Multiplicity is allowed and applies uniformly to every row.
	Multiplicity Expression
}

// Round implements [Query]. Returns the latest [Round] across every column in
// Tab (including Tab.Selector) and, when present, Multiplicity. The bus's
// semantic check cannot be performed before this round — both α/β need every
// participant column to be committed before being sampled.
func (mb *MessageBus) Round() *Round {
	var best *Round
	if r := mb.Tab.Round(); r != nil {
		best = r
	}
	if mb.Multiplicity != nil {
		if r := maxRoundInExpr(mb.Multiplicity); r != nil && (best == nil || r.ID > best.ID) {
			best = r
		}
	}
	return best
}

// Check implements [Query]. Always returns nil. The per-Handle multiset
// identity is inherently cross-query and is discharged by the [messagebus]
// compiler pass together with the [logderivativesum] compiler that follows.
func (mb *MessageBus) Check(_ Runtime) error { return nil }

// NewMessageBusSend constructs and registers a Send entry on (segment,
// handle). The entry contributes +Σ_row filter(row) / d(row) to the
// (segment, handle) accumulator. There is no multiplicity on the Send side
// (per the message-bus spec).
//
// Invariants enforced at construction:
//   - ctx is non-nil.
//   - segment and handle are non-empty.
//   - tab has at least one column (already enforced by [NewTable]).
//
// Panics on any invariant violation.
func (sys *System) NewMessageBusSend(ctx *ContextFrame, segment, handle string, tab Table) *MessageBus {
	return sys.newMessageBus(ctx, segment, handle, BusSend, tab, nil)
}

// NewMessageBusReceive constructs and registers a Receive entry on (segment,
// handle). The entry contributes -Σ_row filter(row) · multiplicity(row) /
// d(row) to the (segment, handle) accumulator. multiplicity may be nil, in
// which case it is treated as the constant 1.
//
// Invariants enforced at construction:
//   - ctx is non-nil.
//   - segment and handle are non-empty.
//   - tab has at least one column.
//   - If multiplicity is non-nil and vector-valued, its module matches
//     tab.Module().
//
// Panics on any invariant violation.
func (sys *System) NewMessageBusReceive(ctx *ContextFrame, segment, handle string, tab Table, multiplicity Expression) *MessageBus {
	return sys.newMessageBus(ctx, segment, handle, BusReceive, tab, multiplicity)
}

// newMessageBus is the shared constructor for [System.NewMessageBusSend] and
// [System.NewMessageBusReceive]. It validates the invariants and appends the
// query to [System.MessageBuses].
func (sys *System) newMessageBus(
	ctx *ContextFrame,
	segment, handle string,
	dir BusDirection,
	tab Table,
	multiplicity Expression,
) *MessageBus {
	if ctx == nil {
		panic("wiop: System.NewMessageBus* requires a non-nil ContextFrame")
	}
	if segment == "" {
		panic("wiop: System.NewMessageBus*: segment must be non-empty")
	}
	if handle == "" {
		panic("wiop: System.NewMessageBus*: handle must be non-empty")
	}
	if len(tab.Columns) == 0 {
		panic("wiop: System.NewMessageBus*: tab must have at least one column")
	}
	if dir == BusSend && multiplicity != nil {
		panic("wiop: System.NewMessageBusSend: multiplicity must be nil on the Send side")
	}
	if multiplicity != nil {
		if m := multiplicity.Module(); m != nil && m != tab.Module() {
			panic(fmt.Sprintf(
				"wiop: System.NewMessageBusReceive: multiplicity module %q differs from tab module %q; "+
					"a vector-valued multiplicity must share the tab's module",
				m.Context.Path(), tab.Module().Context.Path(),
			))
		}
	}

	mb := &MessageBus{
		baseQuery: baseQuery{
			context:     ctx,
			Annotations: make(Annotations),
		},
		Segment:      segment,
		Handle:       handle,
		Direction:    dir,
		Tab:          tab,
		Multiplicity: multiplicity,
	}
	sys.MessageBuses = append(sys.MessageBuses, mb)
	return mb
}
