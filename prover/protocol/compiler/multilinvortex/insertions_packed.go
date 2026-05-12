package multilinvortex

import (
	"fmt"
	"math/bits"
	"sort"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// packedColInfo carries the column handle + nv for the bootstrapper pack step.
type packedColInfo struct {
	col ifaces.Column
	nv  int
}

// InsertBootstrapperOpeningsPacked is the cross-size packed variant of
// [InsertBootstrapperOpenings].
//
// Instead of inserting one MultilinearEval query per size group on the
// original committed columns, it:
//
//  1. Collects ALL round-0 committed columns regardless of size.
//  2. Computes prefix-exclusive locators b_i for each column (buddy alloc).
//  3. Registers a new Committed column Q of size 2^N at round 0, where
//     N = ⌈log₂(Σᵢ 2^{nv_i})⌉.
//  4. Inserts N FieldExt coins at round 1 for a shared master point.
//  5. Inserts ONE MultilinearEval query on Q with K duplicate polys [Q,…,Q],
//     each at its own locator-extended point (b_i ‖ master[N - nv_i :]).
//  6. Registers a round-1 prover action that assigns the K per-pol points.
//  7. Marks the original committed columns as Ignored — only Q enters the
//     proof transcript; the originals are read once at round 0 to build Q
//     and otherwise carry no commitment cost.
//
// Downstream effect: ONE Context (vs K), ONE α coin (vs K), ONE Vortex commit
// on Q (vs K size-group commits in the unpacked variant), and the K original
// claims are batched in a single sumcheck. With per-pol points all sharing
// the SAME cCol suffix (forced by the prefix-exclusive locator structure when
// nv_i ≥ nRow), the prover further collapses to a single UAlpha/RowEvals
// computation that fans out cheaply across the K outputs.
func InsertBootstrapperOpeningsPacked(comp *wizard.CompiledIOP) {
	// Step 1: collect round-0 committed columns.
	originals := comp.Columns.AllHandleCommittedAt(0)
	infos := make([]packedColInfo, 0, len(originals))
	for _, c := range originals {
		size := c.Size()
		if size <= 1 || size&(size-1) != 0 {
			continue
		}
		nv := bits.TrailingZeros(uint(size))
		infos = append(infos, packedColInfo{col: c, nv: nv})
	}
	if len(infos) == 0 {
		return
	}

	// Step 2: compute N and assign locators via buddy allocation.
	var total int
	for _, info := range infos {
		total += 1 << info.nv
	}
	N := bits.Len(uint(total - 1))
	if N == 0 {
		N = 1
	}
	locators := assignLocators(infos, N)

	// Step 3: register Q as a Committed column at round 0.
	qID := ifaces.ColIDf("MLVORTEX_BOOT_PACKED_Q_n%d", N)
	qCol := comp.InsertCommit(0, qID, 1<<N, false)

	// Step 4: insert N master coins at round 1.
	masterCoins := make([]coin.Name, N)
	for d := 0; d < N; d++ {
		masterCoins[d] = coin.Name(fmt.Sprintf("MLVORTEX_BOOT_PACKED_PT_d%d", d))
		comp.InsertCoin(1, masterCoins[d], coin.FieldExt)
	}

	// Step 5: ONE ML query with K duplicate Q polys, each at its own point.
	// All K claims share the same input column Q and the same Fiat-Shamir α
	// downstream, so the prover can collapse UAlpha/RowEvals computation.
	qPols := make([]ifaces.Column, len(infos))
	for i := range infos {
		qPols[i] = qCol
	}
	mergedQuery := comp.InsertMultilinear(1,
		ifaces.QueryID(fmt.Sprintf("MLVORTEX_BOOT_PACKED_MERGED_n%d", N)),
		qPols,
	)

	// cCol-shared safety: cCol = points[k][nRow:N] is identical across all k
	// iff every locator's prefix length L_k = N - nv_k stays inside [0, nRow).
	// That holds when min(nv_k) ≥ N - nRow = nCol. When the bootstrapper can
	// guarantee this from the original-poly sizes, downstream compileWithNRow
	// can collapse RowEvals to a SINGLE column instead of K, and the verifier
	// runs a single Check 3 / Check 1.
	nRowQ := (N + 1) / 2
	nColQ := N - nRowQ
	allSafe := true
	for _, info := range infos {
		if info.nv < nColQ {
			allSafe = false
			break
		}
	}
	if allSafe {
		if comp.ExtraData == nil {
			comp.ExtraData = make(map[string]any)
		}
		set, _ := comp.ExtraData[sharedSafeQueriesKey].(map[ifaces.QueryID]bool)
		if set == nil {
			set = make(map[ifaces.QueryID]bool)
			comp.ExtraData[sharedSafeQueriesKey] = set
		}
		set[mergedQuery.Name()] = true
	}

	// Step 6: register prover actions.
	colHandles := make([]ifaces.Column, len(infos))
	nvs := make([]int, len(infos))
	for i, info := range infos {
		colHandles[i] = info.col
		nvs[i] = info.nv
	}
	shared := &packedBootstrapperShared{}
	comp.RegisterProverAction(0, &packedBootstrapperRound0Action{
		Originals: colHandles,
		Nv:        nvs,
		Locators:  locators,
		N:         N,
		QCol:      qCol,
		shared:    shared,
	})
	comp.RegisterProverAction(1, &packedBootstrapperRound1Action{
		Nv:          nvs,
		Locators:    locators,
		N:           N,
		Query:       mergedQuery,
		MasterCoins: masterCoins,
	})
	_ = shared // currently unused, reserved for future cross-action data

	// Step 7: originals are no longer part of the proof transcript — only Q
	// is. Drop them to Ignored so downstream passes do not re-process them.
	for _, c := range colHandles {
		comp.Columns.MarkAsIgnored(c.GetColID())
	}
}

// assignLocators runs the buddy-allocator at compile time. Returns the
// locator integer for each colInfo entry. Same algorithm as PackPolys.
func assignLocators(infos []packedColInfo, N int) []int {
	K := len(infos)
	locators := make([]int, K)

	// Sort indices by nv descending (place biggest first).
	order := make([]int, K)
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(i, j int) bool {
		return infos[order[i]].nv > infos[order[j]].nv
	})

	free := make([][]int, N+1)
	free[N] = []int{0}
	var allocate func(level int) int
	allocate = func(level int) int {
		if level > N {
			return -1
		}
		if n := len(free[level]); n > 0 {
			off := free[level][n-1]
			free[level] = free[level][:n-1]
			return off
		}
		parent := allocate(level + 1)
		if parent == -1 {
			return -1
		}
		free[level] = append(free[level], parent+(1<<level))
		return parent
	}

	for _, k := range order {
		off := allocate(infos[k].nv)
		if off == -1 {
			panic(fmt.Sprintf("assignLocators: no space for col %d (nv=%d)", k, infos[k].nv))
		}
		locators[k] = off >> infos[k].nv
	}
	return locators
}

// packedBootstrapperShared is reserved for cross-round state between the
// round-0 and round-1 packed-bootstrapper actions. Currently unused: Q is
// materialised at round 0 and accessed at round 1 via run.GetColumn.
type packedBootstrapperShared struct{}

// packedBootstrapperRound0Action materialises Q from the original committed
// columns and assigns it. Runs at round 0 so the round-0 Vortex commit can
// read Q's data.
type packedBootstrapperRound0Action struct {
	Originals []ifaces.Column
	Nv        []int
	Locators  []int
	N         int
	QCol      ifaces.Column
	shared    *packedBootstrapperShared
}

func (a *packedBootstrapperRound0Action) Run(run *wizard.ProverRuntime) {
	Q := make([]field.Element, 1<<a.N)
	for i, c := range a.Originals {
		sv := run.GetColumn(c.GetColID())
		offset := a.Locators[i] << a.Nv[i]
		size := 1 << a.Nv[i]
		base := sv.IntoRegVecSaveAlloc()
		copy(Q[offset:offset+size], base)
	}
	run.AssignColumn(a.QCol.GetColID(), smartvectors.NewRegular(Q))
}

// packedBootstrapperRound1Action assigns the K per-pol locator-extended
// evaluation points for the single merged MultilinearEval query on Q. Runs
// at round 1, after the master coins are drawn.
//
// Each Ys[i] is set to a zero placeholder; the downstream multilinvortex
// ProverAction overwrites it with the actual evaluation Q(b_i ‖ ζ_i) as a
// byproduct of the RowEvals fold.
type packedBootstrapperRound1Action struct {
	Nv          []int
	Locators    []int
	N           int
	Query       query.MultilinearEval
	MasterCoins []coin.Name
}

func (a *packedBootstrapperRound1Action) Run(run *wizard.ProverRuntime) {
	master := make([]fext.Element, a.N)
	for d, name := range a.MasterCoins {
		master[d] = run.GetRandomCoinFieldExt(name)
	}
	K := len(a.Nv)
	points := make([][]fext.Element, K)
	ys := make([]fext.Element, K)
	for i := 0; i < K; i++ {
		L := a.N - a.Nv[i]
		point := make([]fext.Element, a.N)
		// First L variables: locator bits, MSB-first.
		for j := 0; j < L; j++ {
			if (a.Locators[i]>>(L-1-j))&1 == 1 {
				point[j].SetOne()
			}
		}
		// Last nv_i variables: ζ_i = master[L:] (SUFFIX of master point).
		copy(point[L:], master[L:])
		points[i] = point
	}
	run.AssignMultilinearExt(a.Query.Name(), points, ys...)
}
