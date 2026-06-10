// Package messagebus implements the LogUp message-bus compiler pass for the
// wiop protocol framework.
//
// The pass consumes every unreduced [wiop.MessageBus] entry and emits, for
// each (segment, handle) pair, a [wiop.LogDerivativeSum] holding the per-pair
// running sum. A single verifier action per handle then asserts that the
// per-segment cells for that handle sum to zero. By Schwartz–Zippel over two
// extension-field coins α, β (shared across every participant of every
// handle reduced by this pass), the resulting identity holds, for each
// handle h, iff
//
//	∑_{Send entries on h}  Σ_row filter(row) ·  1               / d_h(row)
//	    =
//	∑_{Recv entries on h}  Σ_row filter(row) ·  Multiplicity(row) / d_h(row)
//
// where d_h(row) = β + α^{w_h-1}·c_0(row) + … + c_{w_h-1}(row). Equivalently,
// the multiset of rows sent into h equals the multiset of rows received from
// h, weighted by the receiver-side multiplicities. The same α, β are reused
// across handles (each handle just folds with α raised to powers up to its
// own width); handles remain independent multiset identities because each is
// asserted by its own verifier action. See [wiop.MessageBus] for the
// per-entry semantics.
//
// The coins themselves are NOT allocated by this pass: they must be
// pre-populated on the System as [wiop.System.MessageBusAlpha] and
// [wiop.System.MessageBusBeta] by an external mechanism (typically a
// Fiat–Shamir derivation from round-0 columns) before Compile runs. The pass
// panics if either is nil while there are unreduced MessageBus entries to
// process.
//
// Caller order: invoke messagebus.Compile(sys) BEFORE
// logderivativesum.Compile(sys); the latter discharges the LogDerivativeSums
// this pass emits.
package messagebus

import (
	"fmt"
	"sort"

	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover-ray/wiop"
)

// Compile reduces every unreduced [wiop.MessageBus] entry in sys to a
// collection of [wiop.LogDerivativeSum] queries (one per (segment, handle)
// pair) plus one [wiop.VerifierAction] per handle that asserts the per-segment
// sums total to zero. See the package documentation for the full reduction.
//
// The pass appends up to one fresh interactive round to sys.Rounds: the
// round on which the [wiop.LogDerivativeSum] result cells and the per-handle
// verifier action live. The coin round is NOT allocated here — the supplied
// [wiop.System.MessageBusAlpha] and [wiop.System.MessageBusBeta] are
// expected to already be registered on some earlier round. Both coins must
// be non-nil whenever there is at least one unreduced MessageBus entry.
//
// Already-reduced entries are skipped; remaining unreduced entries are marked
// reduced on return.
func Compile(sys *wiop.System) {
	// Collect every unreduced MessageBus entry in declaration order, indexed by
	// handle. Sort the handles for deterministic round/coin/cell ordering
	// across runs.
	byHandle := map[string][]*wiop.MessageBus{}
	for _, mb := range sys.MessageBuses {
		if mb.IsReduced() {
			continue
		}
		byHandle[mb.Handle] = append(byHandle[mb.Handle], mb)
	}
	if len(byHandle) == 0 {
		return
	}
	handles := make([]string, 0, len(byHandle))
	for h := range byHandle {
		handles = append(handles, h)
	}
	sort.Strings(handles)

	compCtx := sys.Context.Childf("message-bus")

	// Pull the shared (α, β) coins from the System. They are NOT allocated
	// here — the caller (typically a round-0 Fiat–Shamir derivation) must
	// have populated both fields before this pass runs.
	alpha := sys.MessageBusAlpha
	beta := sys.MessageBusBeta
	if alpha == nil {
		panic("wiop/compilers/messagebus: System.MessageBusAlpha is nil; the messagebus pass requires an externally-supplied α coin")
	}
	if beta == nil {
		panic("wiop/compilers/messagebus: System.MessageBusBeta is nil; the messagebus pass requires an externally-supplied β coin")
	}

	// The result round (where LDS cells and the verifier action live) must
	// come strictly after the latest of (a) every participant column's round
	// and (b) the round on which the supplied α/β are sampled. Otherwise the
	// LDS cell could be assigned before its inputs are available.
	maxRound := latestParticipantRound(byHandle)
	maxRound = laterRound(maxRound, alpha.Round())
	maxRound = laterRound(maxRound, beta.Round())
	resultRound := ensureRoundAfter(sys, maxRound)

	// Per-handle: validate uniform width within a handle. Widths may differ
	// across handles (each handle just raises the shared α to powers up to
	// its own width).
	for _, h := range handles {
		entries := byHandle[h]
		width := entries[0].Tab.Width()
		for _, mb := range entries[1:] {
			if mb.Tab.Width() != width {
				panic(fmt.Sprintf(
					"wiop/compilers/messagebus: handle %q has a participant of width %d at %q "+
						"but expected %d (set by the first participant at %q); all participants of a handle must share a width",
					h, mb.Tab.Width(), mb.Context().Path(), width, entries[0].Context().Path(),
				))
			}
		}
	}

	// Per (handle, segment): collect Fractions, build one LogDerivativeSum,
	// remember the resulting cell so we can sum them per handle.
	cellsByHandle := make(map[string][]*wiop.Cell, len(handles))
	for _, h := range handles {
		entries := byHandle[h]

		// Group entries by segment. We iterate in declaration order so that
		// segment order is deterministic; the registration ordering of
		// fractions within a segment is then the input order.
		bySegment := map[string][]*wiop.MessageBus{}
		segmentOrder := []string{}
		for _, mb := range entries {
			if _, seen := bySegment[mb.Segment]; !seen {
				segmentOrder = append(segmentOrder, mb.Segment)
			}
			bySegment[mb.Segment] = append(bySegment[mb.Segment], mb)
		}

		hCtx := compCtx.Childf("handle-%s", h)
		for _, seg := range segmentOrder {
			fractions := buildFractionsForSegment(alpha, beta, bySegment[seg])
			ld := sys.NewLogDerivativeSum(hCtx.Childf("seg-%s", seg), fractions)
			cellsByHandle[h] = append(cellsByHandle[h], ld.Result)
		}
	}

	// One verifier action per handle: the per-segment LogDerivativeSum cells
	// must algebraically sum to zero.
	for _, h := range handles {
		resultRound.RegisterVerifierAction(&handleSumIsZero{
			Handle: h,
			Cells:  cellsByHandle[h],
			Path:   compCtx.Childf("handle-%s", h).Childf("sum-is-zero").Path(),
		})
	}

	// Mark every consumed entry as reduced.
	for _, h := range handles {
		for _, mb := range byHandle[h] {
			mb.MarkAsReduced()
		}
	}
}

// latestParticipantRound returns the [wiop.Round] with the highest ID among
// the participant columns and multiplicity expressions of every unreduced
// MessageBus entry, or nil if no entry references a round-bearing leaf.
func latestParticipantRound(byHandle map[string][]*wiop.MessageBus) *wiop.Round {
	var best *wiop.Round
	update := func(r *wiop.Round) {
		if r != nil && (best == nil || r.ID > best.ID) {
			best = r
		}
	}
	for _, entries := range byHandle {
		for _, mb := range entries {
			update(mb.Round())
		}
	}
	return best
}

// laterRound returns whichever of a, b has the higher ID. nil counts as
// "earlier than every real round"; if both are nil the result is nil.
func laterRound(a, b *wiop.Round) *wiop.Round {
	switch {
	case a == nil:
		return b
	case b == nil:
		return a
	case b.ID > a.ID:
		return b
	default:
		return a
	}
}

// ensureRoundAfter returns a round with ID > after.ID, reusing the existing
// tail round when one already sits in that slot; otherwise appending a fresh
// round via sys.NewRound. after may be nil, in which case the returned round
// is sys.Rounds[0] (allocated if absent).
func ensureRoundAfter(sys *wiop.System, after *wiop.Round) *wiop.Round {
	startID := -1
	if after != nil {
		startID = after.ID
	}
	for len(sys.Rounds) <= startID+1 {
		sys.NewRound()
	}
	return sys.Rounds[startID+1]
}

// buildFractionsForSegment turns every MessageBus entry on one (segment,
// handle) into a [wiop.Fraction] suitable for [wiop.System.NewLogDerivativeSum].
//
// Each entry contributes one fraction:
//
//	Send:    Filter = Tab.Selector, Numerator = +1,             Denominator = d_h(row)
//	Receive: Filter = Tab.Selector, Numerator = -Multiplicity,  Denominator = d_h(row)
//
// where d_h(row) = β + α^{w-1}·c_0(row) + … + c_{w-1}(row). For width-1 tabs
// the Horner loop is empty and the fold reduces to β + c_0(row); α is still
// passed in but never multiplied. A nil Multiplicity on the Receive side
// becomes the constant 1 (so the numerator is just -1).
func buildFractionsForSegment(
	alpha *wiop.CoinField,
	beta *wiop.CoinField,
	entries []*wiop.MessageBus,
) []wiop.Fraction {
	one := wiop.NewConstantField(field.NewFromString("1"))

	fractions := make([]wiop.Fraction, 0, len(entries))
	for _, mb := range entries {
		den := foldDenominator(alpha, beta, mb.Tab.Columns)

		var num wiop.Expression
		switch mb.Direction {
		case wiop.BusSend:
			num = one
		case wiop.BusReceive:
			weight := mb.Multiplicity
			if weight == nil {
				weight = one
			}
			num = wiop.Negate(weight)
		default:
			panic(fmt.Sprintf(
				"wiop/compilers/messagebus: unknown BusDirection %v at %q",
				mb.Direction, mb.Context().Path(),
			))
		}

		var filter wiop.Expression
		if mb.Tab.Selector != nil {
			filter = mb.Tab.Selector
		}

		fractions = append(fractions, wiop.Fraction{
			Filter:      filter,
			Numerator:   num,
			Denominator: den,
		})
	}
	return fractions
}

// foldDenominator returns the expression β + α^{w-1}·c_0 + … + α·c_{w-2} +
// c_{w-1}, evaluated as a Horner pass over cols. For width-1 tabs the loop
// is empty and the result is β + c_0; α is not consulted in that case but
// must still be non-nil so callers don't need to special-case it.
func foldDenominator(alpha, beta *wiop.CoinField, cols []*wiop.ColumnView) wiop.Expression {
	acc := wiop.Expression(cols[0])
	for _, c := range cols[1:] {
		acc = wiop.Add(wiop.Mul(acc, alpha), c)
	}
	return wiop.Add(beta, acc)
}

// handleSumIsZero is the verifier action that closes the message-bus
// reduction: the LogDerivativeSum cells produced for one handle (one per
// participating segment) must algebraically sum to zero.
type handleSumIsZero struct {
	// Handle names the bus this check belongs to. Diagnostic-only.
	Handle string
	// Cells is the per-segment result cells, in registration order.
	Cells []*wiop.Cell
	// Path is the qualified ContextFrame path of the check, used in error
	// messages.
	Path string
}

// Check implements [wiop.VerifierAction]. Sums the values of every per-segment
// cell registered for this handle and returns an error if the total is
// non-zero.
func (h *handleSumIsZero) Check(rt wiop.Runtime) error {
	acc := field.ElemZero()
	for _, c := range h.Cells {
		acc = acc.Add(rt.GetCellValue(c))
	}
	if !acc.IsZero() {
		return fmt.Errorf(
			"wiop/compilers/messagebus: handle %q (%s): per-segment cells sum to %v, expected 0",
			h.Handle, h.Path, acc,
		)
	}
	return nil
}
