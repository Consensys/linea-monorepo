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
	)

	if !isMultiColumn {
		for frag := range a.T {
			tCollapsed[frag] = a.T[frag][0].GetColAssignment(run)
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

	// Partition T
	tBuckets := partitionT(run, a, tCollapsed, numBuckets, mask)

	// Partition S
	sBuckets := partitionS(run, a, sCollapsed, numBuckets, mask)

	// Process Buckets
	mapPool := sync.Pool{
		New: func() any {
			return make(map[fext.Element][2]uint32, 1024)
		},
	}

	parallel.Execute(numBuckets, func(start, end int) {
		mapM := mapPool.Get().(map[fext.Element][2]uint32)
		defer mapPool.Put(mapM)

		for b := start; b < end; b++ {
			// Clear map
			for k := range mapM {
				delete(mapM, k)
			}

			tChunk := tBuckets[b]
			sChunk := sBuckets[b]

			if len(tChunk) == 0 {
				continue
			}

			for _, entry := range tChunk {
				existing, ok := mapM[entry.val]
				if !ok {
					mapM[entry.val] = [2]uint32{entry.frag, entry.row}
				} else {
					if entry.frag < existing[0] || (entry.frag == existing[0] && entry.row < existing[1]) {
						mapM[entry.val] = [2]uint32{entry.frag, entry.row}
					}
				}
			}

			for _, entry := range sChunk {
				pos, ok := mapM[entry.val]
				if !ok {
					utils.Panic("entry %v is not included in the table.", entry.val)
				}

				mFrag, posInFragM := pos[0], pos[1]
				m[mFrag][posInFragM].Add(&m[mFrag][posInFragM], &entry.mk)
			}
		}
	})

	for frag := range m {
		run.AssignColumn(a.M[frag].GetColID(), sv.NewRegular(m[frag]), wizard.DisableAssignmentSizeReduction)
	}

}

type tEntry struct {
	val  fext.Element
	frag uint32
	row  uint32
}

type sEntry struct {
	val fext.Element
	mk  field.Element
}

func hash(v *fext.Element) uint32 {
	h := v.B0.A0.Uint64()
	h = (h * 31) ^ v.B0.A1.Uint64()
	h = (h * 31) ^ v.B1.A0.Uint64()
	h = (h * 31) ^ v.B1.A1.Uint64()
	return uint32(h)
}

// partitionT implements a two-pass parallel partitioning algorithm to avoid
// excessive allocations and synchronization overhead.
//
// Pass 1: Count the number of items for each bucket in parallel.
//
//	Each processor maintains local counts.
//	After the pass, we calculate global offsets for each processor/bucket combination.
//	This allows us to pre-allocate the exact size for each bucket.
//
// Pass 2: Fill the buckets in parallel.
//
//	Each processor writes to the pre-allocated buckets using the computed offsets.
//	Since the offsets are unique per processor, no locks are required during filling.
func partitionT(run *wizard.ProverRuntime, a MAssignmentTask, tCollapsed []sv.SmartVector, numBuckets int, mask uint32) [][]tEntry {
	numProcs := runtime.NumCPU()
	if len(tCollapsed) < numProcs {
		numProcs = max(1, len(tCollapsed))
	}

	counts := make([][]int, numProcs)
	chunkSize := (len(tCollapsed) + numProcs - 1) / numProcs

	var wg sync.WaitGroup
	wg.Add(numProcs)

	for i := 0; i < numProcs; i++ {
		go func(procID int) {
			defer wg.Done()
			start := procID * chunkSize
			end := min(start+chunkSize, len(tCollapsed))
			if start >= end {
				return
			}

			localCounts := make([]int, numBuckets)
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
					localCounts[h&mask]++
				}
			}
			counts[procID] = localCounts
		}(i)
	}
	wg.Wait()

	writeOffsets := make([][]int, numProcs)
	for i := range writeOffsets {
		writeOffsets[i] = make([]int, numBuckets)
	}
	bucketSizes := make([]int, numBuckets)

	for b := 0; b < numBuckets; b++ {
		sum := 0
		for p := 0; p < numProcs; p++ {
			if counts[p] == nil {
				continue
			}
			writeOffsets[p][b] = sum
			sum += counts[p][b]
		}
		bucketSizes[b] = sum
	}

	buckets := make([][]tEntry, numBuckets)
	for b := 0; b < numBuckets; b++ {
		if bucketSizes[b] > 0 {
			buckets[b] = make([]tEntry, bucketSizes[b])
		}
	}

	wg.Add(numProcs)
	for i := 0; i < numProcs; i++ {
		go func(procID int) {
			defer wg.Done()
			start := procID * chunkSize
			end := min(start+chunkSize, len(tCollapsed))
			if start >= end {
				return
			}

			for frag := start; frag < end; frag++ {
				size := tCollapsed[frag].Len()
				rangeStart, rangeEnd := 0, size
				if a.Segmenter != nil {
					root, _ := column.RootsOf(a.T[frag], true)[0].(column.Natural)
					rangeStart, rangeEnd = a.Segmenter.SegmentBoundaryOf(run, root)
				}
				rangeStart = max(0, rangeStart)
				rangeEnd = min(size, rangeEnd)

				for k := rangeStart; k < rangeEnd; k++ {
					v := tCollapsed[frag].GetExt(k)
					h := hash(&v)
					b := h & mask
					offset := writeOffsets[procID][b]
					buckets[b][offset] = tEntry{val: v, frag: uint32(frag), row: uint32(k)}
					writeOffsets[procID][b]++
				}
			}
		}(i)
	}
	wg.Wait()

	return buckets
}

// partitionS follows the same two-pass parallel partitioning strategy as partitionT.
func partitionS(run *wizard.ProverRuntime, a MAssignmentTask, sCollapsed []sv.SmartVector, numBuckets int, mask uint32) [][]sEntry {
	numProcs := runtime.NumCPU()
	if len(sCollapsed) < numProcs {
		numProcs = max(1, len(sCollapsed))
	}

	counts := make([][]int, numProcs)
	chunkSize := (len(sCollapsed) + numProcs - 1) / numProcs

	var wg sync.WaitGroup
	wg.Add(numProcs)

	for i := 0; i < numProcs; i++ {
		go func(procID int) {
			defer wg.Done()
			start := procID * chunkSize
			end := min(start+chunkSize, len(sCollapsed))
			if start >= end {
				return
			}

			localCounts := make([]int, numBuckets)
			for i := start; i < end; i++ {
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
					}
					v := sCollapsed[i].GetExt(k)
					h := hash(&v)
					localCounts[h&mask]++
				}
			}
			counts[procID] = localCounts
		}(i)
	}
	wg.Wait()

	writeOffsets := make([][]int, numProcs)
	for i := range writeOffsets {
		writeOffsets[i] = make([]int, numBuckets)
	}
	bucketSizes := make([]int, numBuckets)

	for b := 0; b < numBuckets; b++ {
		sum := 0
		for p := 0; p < numProcs; p++ {
			if counts[p] == nil {
				continue
			}
			writeOffsets[p][b] = sum
			sum += counts[p][b]
		}
		bucketSizes[b] = sum
	}

	buckets := make([][]sEntry, numBuckets)
	for b := 0; b < numBuckets; b++ {
		if bucketSizes[b] > 0 {
			buckets[b] = make([]sEntry, bucketSizes[b])
		}
	}

	wg.Add(numProcs)
	for i := 0; i < numProcs; i++ {
		go func(procID int) {
			defer wg.Done()
			start := procID * chunkSize
			end := min(start+chunkSize, len(sCollapsed))
			if start >= end {
				return
			}

			for i := start; i < end; i++ {
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
					offset := writeOffsets[procID][b]
					buckets[b][offset] = sEntry{val: v, mk: mk}
					writeOffsets[procID][b]++
				}
			}
		}(i)
	}
	wg.Wait()

	return buckets
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
