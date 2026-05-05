package multilinvortex

import (
	"fmt"
	"math/bits"

	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/sumcheck"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// InsertBootstrapperOpenings is a compiler pass that inserts MultilinearEval
// queries for all round-0 committed columns of comp. Call it on the
// Bootstrapper immediately after DistributeWizard, before the ML Vortex
// compilation pipeline (multilinvortex.Compile / multilineareval.Batch).
//
// Columns are grouped by numVars = log₂(size). For each group of n variables:
//   - n FieldExt coins are inserted at round 1 (the shared evaluation point)
//   - one MultilinearEval query at round 1 covers all columns in the group
//   - a round-1 prover action evaluates every column at the FS challenge and
//     assigns the ML params via AssignMultilinearExt
func InsertBootstrapperOpenings(comp *wizard.CompiledIOP) {
	type nvGroup struct {
		numVars   int
		cols      []ifaces.Column
		coinNames []coin.Name
	}

	groups := make(map[int]*nvGroup)

	for _, col := range comp.Columns.AllHandleCommittedAt(0) {
		size := col.Size()
		if size <= 1 || size&(size-1) != 0 {
			continue // skip non-power-of-two and scalar columns
		}
		nv := bits.TrailingZeros(uint(size))
		g, ok := groups[nv]
		if !ok {
			g = &nvGroup{numVars: nv}
			for d := 0; d < nv; d++ {
				name := coin.Name(fmt.Sprintf("ML_OPEN_POINT_nv%d_d%d", nv, d))
				comp.InsertCoin(1, name, coin.FieldExt)
				g.coinNames = append(g.coinNames, name)
			}
			groups[nv] = g
		}
		g.cols = append(g.cols, col)
	}

	for _, g := range groups {
		qName := ifaces.QueryID(fmt.Sprintf("ML_OPEN_nv%d", g.numVars))
		q := comp.InsertMultilinear(1, qName, g.cols)
		comp.RegisterProverAction(1, &MlOpeningProverAction{
			Q:         q,
			Cols:      g.cols,
			CoinNames: g.coinNames,
		})
	}
}

// InsertBootstrapperOpeningsSharedPoint is a variant of InsertBootstrapperOpenings
// that draws a single master challenge c ∈ F^{n_max} and assigns the prefix
// c[0:n_i] as the evaluation point for each numVars group.  All groups therefore
// share the same underlying FS coins, which means the evaluation points across
// groups are correlated prefixes of the same master challenge rather than
// independently drawn.
//
// This is semantically equivalent to the per-group variant (soundness holds by
// Schwartz-Zippel on the shared prefix), but uses fewer FS coins at compile
// time and makes all initial claims lie on a common sub-space from round 1
// onward — potentially reducing Batch work in later rounds.
func InsertBootstrapperOpeningsSharedPoint(comp *wizard.CompiledIOP) {
	type nvGroup struct {
		numVars int
		cols    []ifaces.Column
	}

	groups := make(map[int]*nvGroup)
	nMax := 0

	for _, col := range comp.Columns.AllHandleCommittedAt(0) {
		size := col.Size()
		if size <= 1 || size&(size-1) != 0 {
			continue
		}
		nv := bits.TrailingZeros(uint(size))
		if nv > nMax {
			nMax = nv
		}
		g, ok := groups[nv]
		if !ok {
			g = &nvGroup{numVars: nv}
			groups[nv] = g
		}
		g.cols = append(g.cols, col)
	}

	if nMax == 0 {
		return
	}

	// Insert nMax shared coins once — all groups borrow a prefix of these.
	masterCoins := make([]coin.Name, nMax)
	for d := 0; d < nMax; d++ {
		name := coin.Name(fmt.Sprintf("ML_OPEN_POINT_SHARED_d%d", d))
		comp.InsertCoin(1, name, coin.FieldExt)
		masterCoins[d] = name
	}

	for _, g := range groups {
		qName := ifaces.QueryID(fmt.Sprintf("ML_OPEN_SHARED_nv%d", g.numVars))
		q := comp.InsertMultilinear(1, qName, g.cols)
		comp.RegisterProverAction(1, &MlOpeningProverAction{
			Q:         q,
			Cols:      g.cols,
			CoinNames: masterCoins[:g.numVars], // prefix of master challenge
		})
	}
}

type MlOpeningProverAction struct {
	Q         query.MultilinearEval
	Cols      []ifaces.Column
	CoinNames []coin.Name
}

func (a *MlOpeningProverAction) Run(run *wizard.ProverRuntime) {
	nv := len(a.CoinNames)
	point := make([]fext.Element, nv)
	for d, name := range a.CoinNames {
		point[d] = run.GetRandomCoinFieldExt(name)
	}

	points := make([][]fext.Element, len(a.Cols))
	ys := make([]fext.Element, len(a.Cols))

	// Evaluate each column at point in parallel. IntoRegVecSaveAllocExt returns
	// a freshly allocated slice, so we fold it in-place (no Clone needed).
	parallel.Execute(len(a.Cols), func(start, stop int) {
		for k := start; k < stop; k++ {
			ml := sumcheck.MultiLin(run.GetColumn(a.Cols[k].GetColID()).IntoRegVecSaveAllocExt())
			for _, r := range point {
				ml.Fold(r)
			}
			ys[k] = ml[0]
			points[k] = point
		}
	})

	run.AssignMultilinearExt(a.Q.Name(), points, ys...)
}
