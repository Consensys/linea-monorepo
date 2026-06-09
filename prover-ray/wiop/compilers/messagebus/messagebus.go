// Package messagebus implements the LogUp message-bus compiler pass for the
// wiop protocol framework.
//
// The pass consumes every unreduced [wiop.MessageBus] entry and emits, for
// each (segment, handle) pair, a [wiop.LogDerivativeSum] holding the per-pair
// running sum. A single verifier action per handle then asserts that the
// per-segment cells for that handle sum to zero. By Schwartz–Zippel over two
// freshly-sampled extension-field coins α, β (shared across all participants
// of a handle), the resulting identity holds iff
//
//	∑_{Send entries on h}  Σ_row filter(row) ·  1               / d_h(row)
//	    =
//	∑_{Recv entries on h}  Σ_row filter(row) ·  Multiplicity(row) / d_h(row)
//
// where d_h(row) = β_h + α_h^{w-1}·c_0(row) + … + c_{w-1}(row). Equivalently,
// the multiset of rows sent into h equals the multiset of rows received from
// h, weighted by the receiver-side multiplicities. See [wiop.MessageBus] for
// the per-entry semantics.
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
// The pass appends up to two fresh interactive rounds to sys.Rounds: one for
// the per-handle (α, β) coins, one for the LogDerivativeSum result cells. If
// the participating tables include multi-column tuples for some handle, an α
// coin is allocated for that handle; otherwise only β is allocated.
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

	// Locate the latest participant round across every entry. Coins must be
	// sampled strictly later, so we allocate two fresh rounds (coin + result)
	// at the tail of sys.Rounds — that guarantees the ordering regardless of
	// what existed before.
	maxParticipantRound := latestParticipantRound(byHandle)
	coinRound, resultRound := ensureTwoRoundsAfter(sys, maxParticipantRound)

	type handlePlan struct {
		alpha *wiop.CoinField // nil when no multi-column tab participates in this handle
		beta  *wiop.CoinField
		width int // shared column width of every participant in this handle
	}
	plans := make(map[string]handlePlan, len(handles))

	// Per-handle: validate uniform width, then allocate (α?, β).
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

		hCtx := compCtx.Childf("handle-%s", h)
		plan := handlePlan{width: width}
		if width > 1 {
			plan.alpha = coinRound.NewCoinField(hCtx.Childf("alpha"))
		}
		plan.beta = coinRound.NewCoinField(hCtx.Childf("beta"))
		plans[h] = plan
	}

	// Per (handle, segment): collect Fractions, build one LogDerivativeSum,
	// remember the resulting cell so we can sum them per handle.
	cellsByHandle := make(map[string][]*wiop.Cell, len(handles))
	for _, h := range handles {
		entries := byHandle[h]
		plan := plans[h]

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
			fractions := buildFractionsForSegment(plan.alpha, plan.beta, bySegment[seg])
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

// ensureTwoRoundsAfter returns two rounds (coin, result) with coin.ID >
// after.ID and result.ID == coin.ID + 1. Existing rounds at the tail of
// sys.Rounds are reused when present; otherwise fresh ones are appended via
// sys.NewRound. after may be nil, in which case coin = sys.Rounds[0] if it
// exists or a freshly-appended round.
func ensureTwoRoundsAfter(sys *wiop.System, after *wiop.Round) (coin, result *wiop.Round) {
	startID := -1
	if after != nil {
		startID = after.ID
	}
	// Need rounds at indices startID+1 and startID+2.
	for len(sys.Rounds) <= startID+2 {
		sys.NewRound()
	}
	coin = sys.Rounds[startID+1]
	result = sys.Rounds[startID+2]
	return
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
// alpha is nil and the fold collapses to β + c_0(row). A nil Multiplicity on
// the Receive side becomes the constant 1 (so the numerator is just -1).
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
// c_{w-1}, evaluated as a Horner pass over cols. When alpha is nil (width-1
// table), the result is β + c_0.
func foldDenominator(alpha, beta *wiop.CoinField, cols []*wiop.ColumnView) wiop.Expression {
	if alpha == nil {
		// width == 1: β + c_0
		return wiop.Add(beta, cols[0])
	}
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
