package logderivativesum

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"

	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"

	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/pragmas"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/exit"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// proverTaskAtRound implements the [wizard.ProverAction] interface. It gathers
// all the operations related to all compiled tables altogether that have to be
// done at a particular round.
//
// Namely, if applied to the round N. The action will be responsible for
// assigning the M column for tables compiled on round N and the SigmaS/SigmaT
// and their respective LocalOpening for the tables compiled at round N-1.
//
// All these actions are performed in parallel.
type ProverTaskAtRound struct {

	// MAssignmentTasks lists all the tasks consisting of assigning the column
	// M related to table that are scheduled in the current interaction round.
	MAssignmentTasks []MAssignmentTask

	// ZAssignmentTasks lists all the tasks consisting of assigning the
	// columns SigmaS and SigmaT for the given round.
	ZAssignmentTasks []ZAssignmentTask
}

// Run implements the [wizard.ProverAction interface]. The tasks will spawn
// a goroutine for each tasks and wait for all of them to finish. The approach
// for parallelization can be justified if the number of go-routines stays low
// (e.g. less than 1000s).
func (p ProverTaskAtRound) Run(run *wizard.ProverRuntime) {

	wg := &sync.WaitGroup{}
	wg.Add(p.numTasks())

	var (
		panicTrace []byte
		panicMsg   any
		panicOnce  = &sync.Once{}
	)

	for i := range p.MAssignmentTasks {
		go func(i int) {

			// In case the subtask panics, we recover so that we can repanic in
			// the main goroutine. Simplifying the process of tracing back the
			// error and allowing to test the panics.
			defer func() {
				if r := recover(); r != nil {
					panicOnce.Do(func() {
						panicMsg = r
						panicTrace = debug.Stack()

						inspectWiop(run)
					})
				}

				wg.Done()
			}()

			p.MAssignmentTasks[i].Run(run)
		}(i)
	}

	for i := range p.ZAssignmentTasks {
		go func(i int) {

			// In case the subtask panics, we recover so that we can repanic in
			// the main goroutine. Simplifying the process of tracing back the
			// error and allowing to test the panics.
			defer func() {
				if r := recover(); r != nil {
					panicOnce.Do(func() {
						panicMsg = r
						panicTrace = debug.Stack()
					})
				}

				wg.Done()
			}()
			p.ZAssignmentTasks[i].Run(run)
		}(i)
	}

	wg.Wait()

	if len(panicTrace) > 0 {
		utils.Panic("Had a panic: %v\nStack: %v\n", panicMsg, string(panicTrace))
	}
}

// pushMAssignment appends an [mAssignmentTask] to the list of tasks
func (p *ProverTaskAtRound) pushMAssignment(m MAssignmentTask) {
	p.MAssignmentTasks = append(p.MAssignmentTasks, m)
}

// pushZAssignment appends an [sigmaAssignmentTask] to the list of tasks
func (p *ProverTaskAtRound) pushZAssignment(s ZAssignmentTask) {
	p.ZAssignmentTasks = append(p.ZAssignmentTasks, s)
}

// numTasks returns the total number of tasks that are scheduled in the
// [proverTaskAtRound].
func (p *ProverTaskAtRound) numTasks() int {
	return len(p.MAssignmentTasks) + len(p.ZAssignmentTasks)
}

// mAssignmentWork specifically represent the prover task of computing and
// assigning the [singleTableCtx.M] for a particular table. M is computing the
// appearance of the rows of T in the rows of S.
type MAssignmentTask struct {

	// M is the column that the assignMWork
	M []ifaces.Column

	// T the lookup table to which the task is related
	T []table

	// S is the list of checked tables for which inclusion within T is enforced
	// by a compiled query.
	S []table

	// SFilter stores the filters that are applied for each table S.
	SFilter []ifaces.Column

	// Segmenter is the segmenter that will be used to omit some of the values
	// from the "S" column.
	Segmenter ColumnSegmenter
}

// Run executes the task represented by the receiver of the method. Namely, it
// actually computes the value of M.
//
// In the case where the table has a single column, the execution path is
// straightforward: simply counting values in a hashmap. In the multi-column
// case there is a trick going on. The prover samples a random value and does
// the counting over a linear combination of the rows using the powers of the
// randomness as coefficient. We refer to this step as "collapsing" the columns.
//
// This crucially relies on the actual randomness of the sampled randomness.
// Without this, a malicious actor may send a proof request which will
// invalidate the counting (and thus, the whole proof later on).
//
// Note that this trick is completely distinct from the sampling of the coin
// Alpha and is purely internal to the prover's work. Therefore, it cannot be
// a soundness concern.
//
// In case one of the Ss contains an entry that does not appear in T, the
// function panics. This aims at early detecting that the lookup query is not
// satisfied.
func (a MAssignmentTask) Run(run *wizard.ProverRuntime) {

	var (
		// isMultiColumn flags whether the table have multiple column and
		// whether the "collapsing" trick is needed.
		isMultiColumn = len(a.T[0]) > 1

		// tCollapsed contains either the assignment of T if the table is a
		// single column (e.g. isMultiColumn=false) or its collapsed version
		// otherwise.
		tCollapsed = make([]sv.SmartVector, len(a.T))

		// sCollapsed contains either the assignment of the Ss if the table is a
		// single column (e.g. isMultiColumn=false) or their collapsed version
		// otherwise.
		sCollapsed = make([]sv.SmartVector, len(a.S))

		// fragmentUnionSize contains the total number of rows contained in all
		// the fragments of T combined.
		fragmentUnionSize int
	)

	if !isMultiColumn {
		for frag := range a.T {
			tCollapsed[frag] = a.T[frag][0].GetColAssignment(run)
			fragmentUnionSize += a.T[frag][0].Size()
		}

		for i := range a.S {
			sCollapsed[i] = a.S[i][0].GetColAssignment(run)
		}
	}

	if isMultiColumn {
		// collapsingRandomness is the randomness used in the collapsing trick.
		// It is sampled via `crypto/rand` internally to ensure it cannot be
		// predicted ahead of time by an adversary.
		var collapsingRandomness fext.Element
		collapsingRandomness.MustSetRandom()

		for frag := range a.T {
			tCollapsed[frag] = wizardutils.RandLinCombColAssignment(run, collapsingRandomness, a.T[frag])
			fragmentUnionSize += tCollapsed[frag].Len()
		}

		for i := range a.S {
			sCollapsed[i] = wizardutils.RandLinCombColAssignment(run, collapsingRandomness, a.S[i])
		}
	}

	var (
		// m  is associated with tCollapsed
		// m stores the assignment to the column M as we build it.
		m = make([][]field.Element, len(a.T))
	)

	for frag := range a.T {
		m[frag] = make([]field.Element, tCollapsed[frag].Len())
	}

	// Partitioning configuration
	numCPU := runtime.NumCPU()
	numBuckets := 1
	for numBuckets < numCPU*4 {
		numBuckets *= 2
	}
	mask := uint32(numBuckets - 1)

	type tEntry struct {
		val  fext.Element
		frag uint32
		row  uint32
	}

	type sEntry struct {
		val fext.Element
		mk  field.Element
	}

	tBuckets := make([][][]tEntry, numBuckets)
	sBuckets := make([][][]sEntry, numBuckets)
	tBucketLocks := make([]sync.Mutex, numBuckets)
	sBucketLocks := make([]sync.Mutex, numBuckets)

	hash := func(v *fext.Element) uint32 {
		h := v.B0.A0.Uint64()
		h = (h * 31) ^ v.B0.A1.Uint64()
		h = (h * 31) ^ v.B1.A0.Uint64()
		h = (h * 31) ^ v.B1.A1.Uint64()
		return uint32(h)
	}

	// Partition T
	parallel.Execute(len(a.T), func(start, end int) {
		localTBuckets := make([][]tEntry, numBuckets)

		for frag := start; frag < end; frag++ {
			size := tCollapsed[frag].Len()
			rangeStart, rangeEnd := 0, size
			if a.Segmenter != nil {
				root, ok := column.RootsOf(a.T[frag], true)[0].(column.Natural)
				if !ok {
					utils.Panic("col %v should be a column.Natural %++v", root.ID, root)
				}
				rangeStart, rangeEnd = a.Segmenter.SegmentBoundaryOf(run, root)
			}
			rangeStart = max(0, rangeStart)
			rangeEnd = min(size, rangeEnd)

			for k := rangeStart; k < rangeEnd; k++ {
				v := tCollapsed[frag].GetExt(k)
				h := hash(&v)
				b := h & mask
				localTBuckets[b] = append(localTBuckets[b], tEntry{val: v, frag: uint32(frag), row: uint32(k)})
			}
		}

		for b := 0; b < numBuckets; b++ {
			if len(localTBuckets[b]) > 0 {
				tBucketLocks[b].Lock()
				tBuckets[b] = append(tBuckets[b], localTBuckets[b])
				tBucketLocks[b].Unlock()
			}
		}
	})

	// Partition S
	parallel.Execute(len(sCollapsed), func(startIdx, endIdx int) {
		localSBuckets := make([][]sEntry, numBuckets)

		for i := startIdx; i < endIdx; i++ {
			size := sCollapsed[i].Len()
			start, stop := 0, size
			hasFilter := a.SFilter[i] != nil
			var filter []field.Element

			if a.Segmenter != nil {
				sCol := column.RootsOf(a.S[i], true)[0].(column.Natural)
				start, stop = a.Segmenter.SegmentBoundaryOf(run, sCol)
			}

			if hasFilter {
				filter = a.SFilter[i].GetColAssignment(run).IntoRegVecSaveAlloc()
			}

			for k := max(0, start); k < min(stop, size); k++ {
				if hasFilter {
					if filter[k].IsZero() {
						continue
					}
					if !filter[k].IsOne() {
						err := fmt.Errorf(
							"the filter column `%v` has a non-binary value at position `%v`: (%v)",
							a.SFilter[i].GetColID(),
							k,
							filter[k].String(),
						)
						exit.OnUnsatisfiedConstraints(err)
					}
				}

				v := sCollapsed[i].GetExt(k)

				mk := field.One()
				switch {
				case k == 0 && start < 0 && !hasFilter:
					mk = field.NewElement(uint64(-start + 1))
				case k == size-1 && stop > size && !hasFilter:
					mk = field.NewElement(uint64(stop - size + 1))
				}

				h := hash(&v)
				b := h & mask
				localSBuckets[b] = append(localSBuckets[b], sEntry{val: v, mk: mk})
			}
		}

		for b := 0; b < numBuckets; b++ {
			if len(localSBuckets[b]) > 0 {
				sBucketLocks[b].Lock()
				sBuckets[b] = append(sBuckets[b], localSBuckets[b])
				sBucketLocks[b].Unlock()
			}
		}
	})

	// Process Buckets
	parallel.Execute(numBuckets, func(start, end int) {
		for b := start; b < end; b++ {
			count := 0
			for _, chunk := range tBuckets[b] {
				count += len(chunk)
			}

			mapM := make(map[fext.Element][2]uint32, count)

			for _, chunk := range tBuckets[b] {
				for _, entry := range chunk {
					existing, ok := mapM[entry.val]
					if !ok {
						mapM[entry.val] = [2]uint32{entry.frag, entry.row}
					} else {
						if entry.frag < existing[0] || (entry.frag == existing[0] && entry.row < existing[1]) {
							mapM[entry.val] = [2]uint32{entry.frag, entry.row}
						}
					}
				}
			}

			for _, chunk := range sBuckets[b] {
				for _, entry := range chunk {
					pos, ok := mapM[entry.val]
					if !ok {
						utils.Panic("entry %v is not included in the table.", entry.val)
					}

					mFrag, posInFragM := pos[0], pos[1]
					m[mFrag][posInFragM].Add(&m[mFrag][posInFragM], &entry.mk)
				}
			}
		}
	})

	for frag := range m {
		run.AssignColumn(a.M[frag].GetColID(), sv.NewRegular(m[frag]), wizard.DisableAssignmentSizeReduction)
	}

}

// ZAssignmentTask represents a prover task of assignming the columns
// SigmaS and SigmaT for a specific lookup table.
// sigmaAssignment
type ZAssignmentTask ZCtx

func (z ZAssignmentTask) Run(run *wizard.ProverRuntime) {
	parallel.Execute(len(z.ZDenominatorBoarded), func(start, stop int) {

		sb0 := make(field.Vector, z.Size)
		se0 := make(extensions.Vector, z.Size)

		for frag := start; frag < stop; frag++ {

			numeratorMetadata := z.ZNumeratorBoarded[frag].ListVariableMetadata()

			svDenominator := column.EvalExprColumn(run, z.ZDenominatorBoarded[frag])

			// This case does not corresponds to an actual production case
			// because log-derivative sums are always defined in such a way that
			// the denominator depends on a randomness. The case is still here
			// for completeness but we don't optimize for it.
			if sv.IsBase(svDenominator) {
				numerator := sb0
				denominator := svDenominator.IntoRegVecSaveAlloc()
				packedZ := field.BatchInvert(denominator)
				if len(numeratorMetadata) == 0 {
					for i := range numerator {
						numerator[i].SetOne()
					}
				} else {
					evalResult := column.EvalExprColumn(run, z.ZNumeratorBoarded[frag])
					evalResult.WriteInSlice(numerator)
				}
				vp := field.Vector(packedZ)
				vp.Mul(vp, numerator)
				for k := 1; k < len(packedZ); k++ {
					packedZ[k].Add(&packedZ[k], &packedZ[k-1])
				}

				run.AssignColumn(z.Zs[frag].GetColID(), sv.NewRegular(packedZ))
				run.AssignLocalPointExt(z.ZOpenings[frag].ID, fext.Lift(packedZ[len(packedZ)-1]))
				continue
			}
			// we are dealing with extension denominators
			numerator := se0
			// denominator := se1
			denominator := svDenominator.IntoRegVecSaveAllocExt()
			packedZ := fext.ParBatchInvert(denominator, 2)

			if len(numeratorMetadata) == 0 {
				for i := range numerator {
					numerator[i].SetOne()
				}
			} else {
				evalResult := column.EvalExprColumn(run, z.ZNumeratorBoarded[frag])
				if vr, ok := evalResult.(*sv.RegularExt); ok {
					// no need to copy here.
					numerator = extensions.Vector(*vr)
				} else {
					evalResult.WriteInSliceExt(numerator)
				}
			}

			vp := extensions.Vector(packedZ)
			vp.Mul(vp, numerator)

			for k := 1; k < len(packedZ); k++ {
				packedZ[k].Add(&packedZ[k], &packedZ[k-1])
			}

			run.AssignColumn(z.Zs[frag].GetColID(), sv.NewRegularExt(packedZ))
			run.AssignLocalPointExt(z.ZOpenings[frag].ID, packedZ[len(packedZ)-1])

		}
	})

}

func inspectWiop(run *wizard.ProverRuntime) {

	if true {
		return
	}

	columns := run.Spec.Columns.AllKeys()

	fmt.Printf("Name; HasPragmaFullCol; HasPragmaLeftPadded; HasPragmaRightPadded; RangeStart; RangeEnd; Size")

	for _, colID := range columns {

		var (
			col                     = run.Spec.Columns.GetHandle(colID).(column.Natural)
			_, hasPragmaFullCol     = col.GetPragma(pragmas.FullColumnPragma)
			_, hasPragmaLeftPadded  = col.GetPragma(pragmas.LeftPadded)
			_, hasPragmaRightPadded = col.GetPragma(pragmas.RightPadded)
		)

		if !run.Columns.Exists(colID) {
			fmt.Printf("%v; %v; %v; %v; %v; %v; %v\n", colID, hasPragmaFullCol, hasPragmaLeftPadded, hasPragmaRightPadded, "N/A", "N/A", col.Size())
			continue
		}

		var (
			v                    = col.GetColAssignment(run)
			rangeStart, rangeEnd = sv.CoCompactRange(v)
			size                 = v.Len()
		)

		fmt.Printf("%v; %v; %v; %v; %v; %v; %v\n", colID, hasPragmaFullCol, hasPragmaLeftPadded, hasPragmaRightPadded, rangeStart, rangeEnd, size)
	}
}
