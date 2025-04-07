package splitter

import (
	"fmt"
	"reflect"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/sirupsen/logrus"
)

// Parse the constraints and queries
func SplitColumns(size int) func(*wizard.CompiledIOP) {
	return func(ci *wizard.CompiledIOP) {
		Compile(ci, size)
	}
}

// Placeholder to store "compiler-local" variable
type splitterCtx struct {
	selfrecursionCounter int
	size                 int
	commitmentMap        collection.Mapping[ifaces.ColID, []ifaces.Column]
	comp                 *wizard.CompiledIOP
}

func Compile(comp *wizard.CompiledIOP, size int) {

	ctx := splitterCtx{
		selfrecursionCounter: comp.SelfRecursionCount,
		size:                 size,
		commitmentMap:        collection.NewMapping[ifaces.ColID, []ifaces.Column](),
		comp:                 comp,
	}

	numRound := comp.NumRounds()

	// Summarizes the number of "published" columns and their sizes
	publishedSummary := map[int]int{}

	/*
		Replace the commitments
	*/
	for round := 0; round < numRound; round++ {
		handles := comp.Columns.AllHandlesAtRound(round)
		for _, h := range handles {

			status := comp.Columns.Status(h.GetColID())

			// If the column is ignored, we can just skip it. Also if is is public we can
			// as well ignore it. It would be nonetheless interesting to be able to differentiate
			// between proof objects from the
			if status == column.Ignored || status == column.Proof || status == column.VerifyingKey {
				continue
			}

			// If the handle is really small compared to split index we just compile
			// it out by making it a proof element.
			if h.Size() < ctx.size {

				if status.IsPublic() {
					// Nothing to do : the column is already public and we will ask the
					// verifier to perfrom the operation itself.
					continue
				}

				switch status {
				case column.Precomputed:
					// send it to the verifier directly as part of the verifying key
					comp.Columns.SetStatus(h.GetColID(), column.VerifyingKey)
				case column.Committed:
					// send it to the verifier directly as part of the proof
					comp.Columns.SetStatus(h.GetColID(), column.Proof)
				default:
					utils.Panic("Unknown status : %v", status.String())
				}

				// And log it into the summary
				if _, ok := publishedSummary[h.Size()]; !ok {
					publishedSummary[h.Size()] = 0
				}
				publishedSummary[h.Size()]++

				continue
			}

			if h.Size()%ctx.size != 0 {
				panic("the size does not divide")
			}

			// Mark the handle as ignored
			comp.Columns.MarkAsIgnored(h.GetColID())

			// Create the subslices and give them the same status as their parents
			numSubSlices := h.Size() / ctx.size
			subSlices := make([]ifaces.Column, numSubSlices)

			switch status {
			case column.VerifierDefined:
				// If the status of the column is Verifier defined, then the splitting
				// follows a special process. It does not go through the compiledIOP.
				switch vercol := h.(type) {
				case verifiercol.VerifierCol:
					for i := 0; i < len(subSlices); i++ {
						subSlices[i] = vercol.Split(comp, i*ctx.size, (i+1)*ctx.size)
					}
				default:
					utils.Panic("unexpected type of verifier column %T", vercol)
				}
			case column.Precomputed, column.VerifyingKey:
				// Then, on top of defining the new splitted column. We need to assign it
				// directly.
				precomp := comp.Precomputed.MustGet(h.GetColID())
				for i := 0; i < len(subSlices); i++ {
					subSlices[i] = comp.InsertPrecomputed(
						nameHandleSlice(h, i, h.Size()/ctx.size),
						precomp.SubVector(i*ctx.size, (i+1)*ctx.size),
					)
					// For the case when the original slice is a verifying key.
					// That's because "InsertPrecomputed" automatically assigns the
					// status "Precomputed"
					if status != column.Precomputed {
						comp.Columns.SetStatus(subSlices[i].GetColID(), status)
					}
				}

			default:
				for i := 0; i < len(subSlices); i++ {
					subSlices[i] = comp.InsertColumn(round, nameHandleSlice(h, i, h.Size()/ctx.size), ctx.size, status)
				}
			}

			// And register the subslices in the map for easy access later
			ctx.commitmentMap.InsertNew(h.GetColID(), subSlices)
		}
	}

	// Log the summary
	logrus.Infof("Escalated columns with the following profiles ([size:number of columns]) : %v", publishedSummary)

	/*
		Assign the provers for each rounds. The role of the main prover is
		to assign values to all subvectors
	*/
	for round := 0; round < numRound; round++ {
		comp.RegisterProverAction(round, &compileSplitterProverAction{ctx: &ctx, round: round, lastRound: numRound - 1})
	}

	/*
		Replace the global and local constraints. The global constraints are
		reapplied to all "subvectors" separately and "the link" between the
		subvectors is ensured by one or more local constraints at the junction
	*/
	for round := 0; round < numRound; round++ {
		qNames := comp.QueriesNoParams.AllKeysAt(round)
		for _, qName := range qNames {

			if comp.QueriesNoParams.IsIgnored(qName) {
				// Skip ignored constraints
				continue
			}

			q_ := comp.QueriesNoParams.Data(qName)
			switch q := q_.(type) {
			case query.LocalConstraint:
				ctx.compileLocal(comp, q)
			case query.GlobalConstraint:
				ctx.compileGlobal(comp, q)
			default:
				utils.Panic("unexpected type of query that was not compiled : %T (%v)", q, qName)
			}
		}
	}

	/*
		Replace the local evaluation constraints, by one over
	*/
	for round := 0; round < numRound; round++ {
		qNames := comp.QueriesParams.AllKeysAt(round)
		for _, qName := range qNames {

			if comp.QueriesParams.IsIgnored(qName) {
				// Skip ignored constraints
				continue
			}

			q_ := comp.QueriesParams.Data(qName)
			switch q := q_.(type) {
			case query.LocalOpening:
				ctx.compileLocalOpening(comp, q)
			default:
				utils.Panic("unexpected type of query that was not compiled : %T (%v)", q, qName)
			}
		}
	}

	/*
		At the last round, frees the polynomials that are not used anymore
		Note: This is now handled by the single struct below
	*/
}

// compileSplitterProverAction handles both subvector assignment and polynomial freeing in the splitter.
// It implements the [wizard.ProverAction] interface.
type compileSplitterProverAction struct {
	ctx       *splitterCtx
	round     int
	lastRound int
}

// Run executes the appropriate action based on the round.
func (a *compileSplitterProverAction) Run(run *wizard.ProverRuntime) {
	if a.round == a.lastRound {
		// Free polynomials at the last round
		for col := range a.ctx.commitmentMap.InnerMap() {
			run.Columns.TryDel(col)
		}
	} else {
		// Assign subvector values for all rounds
		a.ctx.Prove(a.round)(run)
	}
}

func nameHandleSlice(h ifaces.Column, num, numSlots int) ifaces.ColID {
	return ifaces.ColIDf("%v_SUBSLICE_%v_OVER_%v", h.GetColID(), num, numSlots)
}

// globalConstraintVerifierAction implements the VerifierAction interface for global constraint verification.
type globalConstraintVerifierAction struct {
	q query.GlobalConstraint
}

// Run executes the native verifier check for global constraint consistency.
func (a *globalConstraintVerifierAction) Run(vr *wizard.VerifierRuntime) error {
	err := a.q.Check(vr)
	if err != nil {
		return fmt.Errorf("failure for query %v, here is why %v", a.q.ID, err)
	}
	return nil
}

// RunGnark executes the gnark circuit verifier check for global constraint consistency.
func (a *globalConstraintVerifierAction) RunGnark(api frontend.API, wvc *wizard.WizardVerifierCircuit) {
	a.q.CheckGnark(api, wvc)
}

// localConstraintVerifierAction implements the VerifierAction interface for local constraint verification.
type localConstraintVerifierAction struct {
	q query.LocalConstraint
}

// Run executes the native verifier check for local constraint consistency.
func (a *localConstraintVerifierAction) Run(vr *wizard.VerifierRuntime) error {
	err := a.q.Check(vr)
	if err != nil {
		return fmt.Errorf("failure for query %v, here is why %v", a.q.ID, err)
	}
	return nil
}

// RunGnark executes the gnark circuit verifier check for local constraint consistency.
func (a *localConstraintVerifierAction) RunGnark(api frontend.API, wvc *wizard.WizardVerifierCircuit) {
	a.q.CheckGnark(api, wvc)
}

// replace a global constraint
func (ctx splitterCtx) compileGlobal(comp *wizard.CompiledIOP, q query.GlobalConstraint) {

	// Mark as ignored
	comp.QueriesNoParams.MarkAsIgnored(q.ID)

	board := q.Board()
	metadatas := board.ListVariableMetadata()
	round := comp.QueriesNoParams.Round(q.ID)

	// Check if there is a "hasInterleaved" handle
	for i := range metadatas {
		if h, ok := metadatas[i].(ifaces.Column); ok && hasInterleaved(h) {
			logrus.Warnf("Squeezing the query %v because it refers to an interleaving", q.ID)
			return
		}
	}

	// Verifier checks the query itself.
	if q.DomainSize < ctx.size {
		// Sanity-check : at this point all the parameters of the query
		// should have a public status. Indeed, prior to compiling the
		// constraints to work
		for _, metadata := range metadatas {
			if h, ok := metadata.(ifaces.Column); ok {
				verifiercol.AssertIsPublicCol(comp, h)
			}
		}

		// Requires the verifier to verify the query itself
		comp.RegisterVerifierAction(round, &globalConstraintVerifierAction{q: q})

		// And skip the compilation consequently : we are done
		return
	}

	// Used for cancelling the replayed constraints
	offsetRange := q.MinMaxOffset()

	numSlots := q.DomainSize / ctx.size
	for slot := 0; slot < numSlots; slot++ {
		translationMap := collection.NewMapping[string, *symbolic.Expression]()
		for _, metadata := range metadatas {

			// For each slot, get the expression obtained by replacing the commitment
			// by the appropriated column.

			switch m := metadata.(type) {
			case ifaces.Column:
				// Pass the same variable
				subHandle := ctx.handleForSubsliceInGlobal(m, slot)
				// Sanity-check : the subHandle should have the target size
				if subHandle.Size() != ctx.size {
					utils.Panic(
						"outgoing column %v should have size %v but has size %v (ingoing column was %v, with size %v)",
						subHandle.GetColID(), ctx.size, subHandle.Size(), m.GetColID(), m.Size(),
					)
				}
				translationMap.InsertNew(m.String(), ifaces.ColumnAsVariable(subHandle))
			case variables.X:
				utils.Panic("unsupported, the value of `x` in the unsplit query and the split would be different. query=%v", q.Name())
			case variables.PeriodicSample:
				// Check that the period is not larger than the domain size. If
				// the period is smaller this is a no-op because the period does
				// not change.
				translated := symbolic.NewVariable(metadata)

				if m.T > ctx.size {

					// Here, there are two possibilities. (1) The current slot is
					// on a portion of the Periodic sample where everything is
					// zero or (2) the current slot matchs a portion of the
					// periodic sampling containing a 1. To determine which is
					// the current situation, we need to find out where the slot
					// is located compared to the period.
					var (
						slotStartAt = (slot * ctx.size) % m.T
						slotStopAt  = slotStartAt + ctx.size
					)

					if m.Offset >= slotStartAt && m.Offset < slotStopAt {
						translated = variables.NewPeriodicSample(ctx.size, m.Offset%ctx.size)
					} else {
						translated = symbolic.NewConstant(0)
					}
				}

				// And we can just pass it over because the period does not change
				translationMap.InsertNew(m.String(), translated)
			default:
				// Repass the same variable (for coins or other types of single-valued variable)
				translationMap.InsertNew(m.String(), symbolic.NewVariable(metadata))
			}
		}

		replayed := q.Replay(translationMap)

		// Now, we need to cancel the expression at the beginning and/or the end
		// For the first one, only cancel the end. For the last one, only cancel
		// the beginning.

		if offsetRange.Min < 0 {
			for i := 0; i < offsetRange.Min; i++ {
				// And fill the gap with a local constraint
				if slot > 0 || q.NoBoundCancel {
					ctx.replaceLocal(
						comp, ifaces.QueryIDf("%v_STICKY_%v", q.ID, slot*ctx.size-i),
						q.Expression,
						round, slot*ctx.size-i,
					)
				}
			}
		}

		if offsetRange.Max > 0 {
			for i := 0; i < offsetRange.Max; i++ {
				point := ctx.size - i - 1 // point at which we want to cancel the constraint
				// And fill the gap with a local constraint
				if slot < numSlots-1 || q.NoBoundCancel {
					shift := slot*ctx.size + point
					ctx.replaceLocal(
						comp, ifaces.QueryIDf("%v_STICKY_%v", q.ID, shift),
						q.Expression,
						round, shift,
					)
				}
			}
		}

		// Implicitly, always cancel the constraint because the overflow is always unjustified.
		comp.InsertGlobal(
			round,
			ifaces.QueryIDf("%v_SPLIT_%v_OVER_%v", q.ID, slot, numSlots),
			replayed,
		)
	}
}

// Compile a local constraint
func (ctx splitterCtx) compileLocal(comp *wizard.CompiledIOP, q query.LocalConstraint) {
	// Mark as ignored
	comp.QueriesNoParams.MarkAsIgnored(q.ID)
	round := comp.QueriesNoParams.Round(q.ID)

	board := q.Expression.Board()
	metadatas := board.ListVariableMetadata()

	// Check if there is a "hasInterleaved" handle
	for i := range metadatas {
		if h, ok := metadatas[i].(ifaces.Column); ok && hasInterleaved(h) {
			logrus.Warnf("Squeezing the query %v because it refers to an interleaving", q.ID)
			return
		}
	}

	// Compute the size of the domain of the local constraint
	domainSize := 0
	for _, metadata := range metadatas {
		if h, ok := metadata.(ifaces.Column); ok {
			domainSize = h.Size()
			break
		}
	}

	// Sanity-check, the domain size should be non-zero. Otherwise,
	// that means that something is odd
	if domainSize == 0 {
		utils.Panic("local query %v has domain size 0", q.ID)
	}

	logrus.Tracef(
		"SPLITTER COMPILE LOCAL %v - domain size %v target size %v",
		q.ID, domainSize, ctx.size,
	)

	// Verifier checks the query itself.
	if domainSize < ctx.size {
		// Sanity-check : at this point all the parameters of the query
		// should have a public status. Indeed, prior to compiling the
		// constraints to work
		for _, metadata := range metadatas {
			if h, ok := metadata.(ifaces.Column); ok {
				// The columns in play within the
				verifiercol.AssertIsPublicCol(comp, h)
			}
		}

		// Requires the verifier to verify the query itself
		comp.RegisterVerifierAction(round, &localConstraintVerifierAction{q: q})

		// Skip the rest of the compilation process : we are done
		return
	}

	ctx.replaceLocal(comp, ifaces.QueryIDf("%v_SPLIT", q.ID), q.Expression, round, 0)
}

// replace a global constraint by a local constraint to
// account for the boundary effect of splitting
// the constraint.
func (ctx splitterCtx) replaceLocal(comp *wizard.CompiledIOP, newName ifaces.QueryID, expr *symbolic.Expression, round, shift int) {

	board := expr.Board()
	metadatas := board.ListVariableMetadata()
	translationMap := collection.NewMapping[string, *symbolic.Expression]()

	for _, metadata := range metadatas {
		// Replace the expression by the one

		switch m := metadata.(type) {
		case ifaces.Column:
			// Pass the same variable
			subHandle := ctx.handleForSubsliceInLocal(column.Shift(m, shift), 0)
			translationMap.InsertNew(m.String(), ifaces.ColumnAsVariable(subHandle))
		case variables.PeriodicSample:
			newSample := variables.NewPeriodicSample(m.T, utils.PositiveMod(m.Offset-shift, m.T))
			translationMap.InsertNew(m.String(), newSample)
		default:
			// Repass the same variable
			translationMap.InsertNew(m.String(), symbolic.NewVariable(metadata))
		}
	}

	replayed := expr.Replay(translationMap)
	comp.InsertLocal(round, newName, replayed)
}

// Checks if the current handle has interleaved
func hasInterleaved(h ifaces.Column) bool {
	switch inner := h.(type) {
	case column.Natural, verifiercol.VerifierCol:
		return false
	case column.Shifted:
		return hasInterleaved(inner.Parent)
	default:
		utils.Panic("unexpected type %v", reflect.TypeOf(inner))
	}
	panic("unreachable")
}

// localOpeningDirectVerifierAction implements the VerifierAction interface for direct local opening verification.
type localOpeningDirectVerifierAction struct {
	q query.LocalOpening
}

// Run executes the native verifier check for direct local opening consistency.
func (a *localOpeningDirectVerifierAction) Run(vr *wizard.VerifierRuntime) error {
	return a.q.Check(vr)
}

// RunGnark executes the gnark circuit verifier check for direct local opening consistency.
func (a *localOpeningDirectVerifierAction) RunGnark(api frontend.API, wvc *wizard.WizardVerifierCircuit) {
	a.q.CheckGnark(api, wvc)
}

// localOpeningSplitVerifierAction implements the VerifierAction interface for split local opening verification.
type localOpeningSplitVerifierAction struct {
	q        query.LocalOpening
	newQName ifaces.QueryID
}

// Run executes the native verifier check for split local opening consistency.
func (a *localOpeningSplitVerifierAction) Run(run *wizard.VerifierRuntime) error {
	oldParams := run.GetLocalPointEvalParams(a.q.ID)
	newParams := run.GetLocalPointEvalParams(a.newQName)

	if oldParams != newParams {
		return fmt.Errorf("splitter verifier failed for local opening %v - %v", a.q.ID, a.newQName)
	}

	return nil
}

// RunGnark executes the gnark circuit verifier check for split local opening consistency.
func (a *localOpeningSplitVerifierAction) RunGnark(api frontend.API, wvc *wizard.WizardVerifierCircuit) {
	oldParams := wvc.GetLocalPointEvalParams(a.q.ID)
	newParams := wvc.GetLocalPointEvalParams(a.newQName)
	api.AssertIsEqual(oldParams.Y, newParams.Y)
}

func (ctx splitterCtx) compileLocalOpening(comp *wizard.CompiledIOP, q query.LocalOpening) {

	round := comp.QueriesParams.Round(q.ID)
	comp.QueriesParams.MarkAsIgnored(q.ID)
	position := utils.PositiveMod(column.StackOffsets(q.Pol), q.Pol.Size())

	// Ask the verifier to directly open the
	if q.Pol.Size() < ctx.size {

		verifiercol.AssertIsPublicCol(comp, q.Pol)

		// Requires the verifier to verify the query itself
		comp.RegisterVerifierAction(round, &localOpeningDirectVerifierAction{q: q})

		// And skip the rest of the compilation : we are done
		return
	}

	subvecID, posInSubvec := position/ctx.size, position%ctx.size
	newQName := ifaces.QueryIDf("%v_SPLIT", q.ID)

	// The subslice is linked to the "root" of q.Pol
	nats := column.RootParents(q.Pol)
	if len(nats) > 1 {
		utils.Panic("%++v has several roots", q.Pol)
	}

	subvec := ctx.commitmentMap.MustGet(nats[0].GetColID())[subvecID]
	newQ := comp.InsertLocalOpening(round, newQName, column.Shift(subvec, posInSubvec))

	logrus.Tracef(
		"SPLITTER for LOCAL OPENING %v (%v) -> into %v (%v)",
		q.ID, q.Pol.GetColID(), newQ.ID, newQ.Pol.GetColID(),
	)

	// The prover assigns the new local opening by reusing the result of the past one
	comp.RegisterProverAction(round, &compileLocalOpeningProverAction{
		q:        q,
		newQName: newQName,
	})

	// The verifier ensures that the old and new queries have the same assignment
	comp.RegisterVerifierAction(round, &localOpeningSplitVerifierAction{
		q:        q,
		newQName: newQName,
	})
}

// compileLocalOpeningProverAction assigns the new local opening point in the splitter context.
// It implements the [wizard.ProverAction] interface.
type compileLocalOpeningProverAction struct {
	q        query.LocalOpening
	newQName ifaces.QueryID
}

// Run executes the assignment of the new local opening point.
func (a *compileLocalOpeningProverAction) Run(run *wizard.ProverRuntime) {
	params := run.GetLocalPointEvalParams(a.q.ID)
	run.AssignLocalPoint(a.newQName, params.Y)
}

// Return a subslice handle
func (ctx splitterCtx) handleForSubsliceInGlobal(h ifaces.Column, num int) ifaces.Column {

	// Sanity-check : only for the edge-case h.Size() < ctx.size
	if h.Size() < ctx.size {
		if num > 0 {
			utils.Panic(
				"tried to get share #%v of column %v, but this is an undersized column %v",
				num, h.GetColID(), h.Size(),
			)
		}

		// Not a split  column : returns the input directly
		return h
	}

	if !h.IsComposite() {
		switch col := h.(type) {
		case verifiercol.VerifierCol:
			// Create the split in live
			return col.Split(ctx.comp, num*ctx.size, (num+1)*ctx.size)
		default:
			// No changes
			return ctx.commitmentMap.MustGet(h.GetColID())[num]
		}
	}

	switch inner := h.(type) {
	// Shift the subparent, if the offset is larger than the subparent
	// we repercute it on the num
	case column.Shifted:
		// This works fine assuming h.Size() > ctx.size
		var (
			offset    = inner.Offset
			newNum    = num + (offset / ctx.size)
			newOffset = utils.PositiveMod(offset, ctx.size)
		)
		// This indicates that the offset is so large
		if newNum < 0 {
			newNum = (h.Size() / ctx.size) + newNum
		}
		// The resulting offset should keep the same sign as the old one. This is
		// because the sign indicates which range of position is touched by
		// bound cancellation.
		if newOffset*offset < 0 {
			newOffset -= ctx.size
		}
		parent := ctx.handleForSubsliceInGlobal(inner.Parent, newNum)
		return column.Shift(parent, newOffset)

	default:
		if !h.IsComposite() {
			// No changes
			return h
		}
		utils.Panic("unexpected type %v", reflect.TypeOf(inner))
	}
	panic("unreachable")
}

// Return a subslice handle (used for local constraints conversion). There is a difference in how the negative offsets
// are handled.
func (ctx splitterCtx) handleForSubsliceInLocal(h ifaces.Column, num int) (res ifaces.Column) {

	// Sanity-check : only for the edge-case h.Size() < ctx.size
	if h.Size() < ctx.size && num != 0 {
		utils.Panic("We have h.Size (%v) < ctx.size (%v) but num (%v) != 0 for %v", h.Size(), ctx.size, num, h.GetColID())
	}

	if !h.IsComposite() {
		switch col := h.(type) {
		case verifiercol.VerifierCol:
			// Create the split in live
			return col.Split(ctx.comp, num*ctx.size, (num+1)*ctx.size)
		default:
			// No changes : it means this is a normal column and
			// we shall take the corresponding slice.
			return ctx.commitmentMap.MustGet(h.GetColID())[num]
		}
	}

	switch inner := h.(type) {
	case column.Shifted:
		// Shift the subparent, if the offset is larger than the subparent
		// we repercute it on the num
		if inner.Offset < -ctx.size {
			utils.Panic("unsupported, the offset is too negative")
		}

		offset := utils.PositiveMod(inner.Offset, h.Size())
		deltaNum := offset / ctx.size // This works fine in the case where h.Size() < ctx.size
		parent := ctx.handleForSubsliceInLocal(inner.Parent, num+deltaNum)
		return column.Shift(parent, offset%ctx.size)

	default:
		utils.Panic("unexpected type %v", reflect.TypeOf(inner))
	}
	panic("unreachable")
}

func (ctx splitterCtx) Prove(round int) wizard.MainProverStep {

	return func(run *wizard.ProverRuntime) {

		stopTimer := profiling.LogTimer("splitter compiler")
		defer stopTimer()

		for _, h := range run.Spec.Columns.AllHandlesAtRound(round) {
			// Skip if the commitment is not registered in the map
			if !ctx.commitmentMap.Exists(h.GetColID()) {
				continue
			}

			subSlices := ctx.commitmentMap.MustGet(h.GetColID())

			if h.Size() < ctx.size {
				// Handle the case where the handle is smaller than the size
				slice := make([]field.Element, ctx.size)
				witness := run.Columns.MustGet(h.GetColID())
				for i := 0; i < ctx.size; i += h.Size() {
					witness.WriteInSlice(slice[i : i+h.Size()])
				}
				run.AssignColumn(subSlices[0].GetColID(), smartvectors.NewRegular(slice))
				continue
			}

			// Sanity-check
			if len(subSlices)*ctx.size != h.Size() {
				utils.Panic("Unexpected sizes %v  * %v != %v", len(subSlices), ctx.size, h.Size())
			}

			// If the column is precomputed, then it means it was already assigned
			if ctx.comp.Precomputed.Exists(h.GetColID()) {
				continue
			}

			witness := run.Columns.MustGet(h.GetColID())
			for i := 0; i < len(subSlices); i++ {
				run.AssignColumn(subSlices[i].GetColID(), witness.SubVector(i*ctx.size, (i+1)*ctx.size))
			}
		}
	}
}
