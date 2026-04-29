package sumcheck

import (
	"fmt"
	"math/bits"

	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/polynomials"
	"github.com/consensys/linea-monorepo/prover-ray/utils/parallel"
)

// ProverState drives the sumcheck prover round by round. The caller is
// responsible for Fiat-Shamir: it hashes each [RoundPoly] and feeds the
// resulting challenge back via [FoldAndAdvance].
type ProverState struct {
	// tables[k] is the ext-field copy of input column k, shrunk by half each round.
	tables [][]field.Ext
	// eq is the combined eq table, shrunk by half each round.
	eq   []field.Ext
	cfg  *ProverConfig
	gate Gate
	// claim is the expected sum for the current round.
	claim field.Ext
	round int
	logN  int
}

// NewProverStateWithEqMask initialises the prover for a (multi-)sumcheck.
//
//   - cfg: pre-allocated scratch from [NewProverConfig].
//   - gate: the multivariate polynomial being sumchecked.
//   - tables: base-field evaluation tables (one per gate input); all must have
//     length 2^logN. They are lifted to ext internally and never modified.
//   - qPrimes: evaluation points; len(qPrimes) >= 1; each must have length logN.
//   - mu: recombination challenge (field.Ext) for multi-sumcheck; ignored when
//     len(qPrimes) == 1.
//   - claim: the alleged sum. For multi-sumcheck this must already be the
//     combined claim Σ_j μʲ · claim_j (computed externally by the caller).
//
// It builds the combined eq mask and then delegates to [NewProverStateWithMask].
func NewProverStateWithEqMask(
	cfg *ProverConfig,
	gate Gate,
	tables [][]field.Element,
	qPrimes [][]field.Ext,
	mu field.Ext,
	claim field.Ext,
) (*ProverState, error) {
	if len(qPrimes) == 0 {
		return nil, fmt.Errorf("sumcheck: NewProverState: qPrimes must be non-empty")
	}
	logN := len(qPrimes[0])
	n := 1 << logN

	for j, q := range qPrimes {
		if len(q) != logN {
			return nil, fmt.Errorf("sumcheck: NewProverState: qPrimes[%d] has length %d, want %d", j, len(q), logN)
		}
	}
	if n > len(cfg.EqScratch) {
		return nil, fmt.Errorf("sumcheck: NewProverState: n=%d exceeds cfg.EqScratch capacity %d",
			n, len(cfg.EqScratch))
	}

	// Build the combined eq mask in cfg.EqScratch, then delegate.
	eq := cfg.EqScratch[:n]
	buildEqTable(cfg, eq, qPrimes[0])
	if len(qPrimes) > 1 {
		tmp := cfg.EqTmp[:n]
		muPow := mu // μ¹
		for j := 1; j < len(qPrimes); j++ {
			buildEqTable(cfg, tmp, qPrimes[j], muPow)
			for i := range eq {
				eq[i].Add(&eq[i], &tmp[i])
			}
			muPow.Mul(&muPow, &mu)
		}
	}

	return NewProverStateWithMask(cfg, gate, tables, eq, claim)
}

// NewProverStateWithMask constructs a [ProverState] using a caller-supplied mask
// table instead of deriving one from evaluation-point coordinates.
//
// The prover accumulates Σ_h mask[h]·gate(tables[0][h], tables[1][h], …).
// Passing mask = nil gives the unweighted sum Σ_h gate(tables[h]), equivalent
// to a constant-1 mask: an all-ones table folds back to all-ones under any
// challenge, so the final mask claim is always 1.
//
// len(mask) (when non-nil) and len(tables[k]) must all equal 2^logN and must
// not exceed the capacities pre-allocated in cfg.
func NewProverStateWithMask(
	cfg *ProverConfig,
	gate Gate,
	tables [][]field.Element,
	mask []field.Ext,
	claim field.Ext,
) (*ProverState, error) {
	if len(tables) == 0 {
		return nil, fmt.Errorf("sumcheck: NewProverStateWithMask: tables must be non-empty")
	}
	n := len(tables[0])
	if n == 0 || bits.OnesCount(uint(n)) != 1 {
		return nil, fmt.Errorf("sumcheck: NewProverStateWithMask: table length %d is not a positive power of two", n)
	}
	if len(tables) > len(cfg.Tables) {
		return nil, fmt.Errorf("sumcheck: NewProverStateWithMask: %d tables exceed cfg.Tables capacity %d",
			len(tables), len(cfg.Tables))
	}
	if n > len(cfg.EqScratch) {
		return nil, fmt.Errorf("sumcheck: NewProverStateWithMask: n=%d exceeds cfg.EqScratch capacity %d",
			n, len(cfg.EqScratch))
	}
	for k, t := range tables {
		if len(t) != n {
			return nil, fmt.Errorf("sumcheck: NewProverStateWithMask: table[%d] has length %d, want %d",
				k, len(t), n)
		}
	}
	if mask != nil && len(mask) != n {
		return nil, fmt.Errorf("sumcheck: NewProverStateWithMask: mask length %d != table length %d",
			len(mask), n)
	}

	extTables := liftTables(cfg, tables, n)

	eq := cfg.EqScratch[:n]
	if mask != nil {
		copy(eq, mask)
	} else {
		for i := range eq {
			eq[i].SetOne()
		}
	}

	return finishProverState(cfg, gate, extTables, eq, bits.TrailingZeros(uint(n)), claim), nil
}

// ComputeRoundPoly computes the round polynomial for the current round and
// returns it in Gruen compressed format (evaluations at {0, 2, …, d}).
// Panics if called after all rounds are complete.
func (s *ProverState) ComputeRoundPoly() RoundPoly {
	if s.round >= s.logN {
		panic("sumcheck: ComputeRoundPoly called after all rounds are complete")
	}

	d := s.gate.Degree()
	mid := len(s.eq) / 2
	nInputs := len(s.tables)
	numChunks := (mid + evalSubChunkSize - 1) / evalSubChunkSize

	// Zero out per-thread accumulators.
	for t := range s.cfg.PerThread {
		for i := range s.cfg.PerThread[t].Accum[:d] {
			s.cfg.PerThread[t].Accum[i].SetZero()
		}
	}

	parallel.ExecuteThreadAware(
		numChunks,
		func(_ int) {},
		func(taskID, threadID int) {
			s.computeChunk(taskID, threadID, mid, nInputs, d)
		},
		s.cfg.NumCPU,
	)

	// Reduce per-thread accumulators into the result.
	rp := make(RoundPoly, d)
	for t := range s.cfg.PerThread {
		for i := range rp {
			rp[i].Add(&rp[i], &s.cfg.PerThread[t].Accum[i])
		}
	}

	return rp
}

// computeChunk processes one sub-range of the inner loop for ComputeRoundPoly.
func (s *ProverState) computeChunk(taskID, threadID, mid, nInputs, d int) {
	scratch := &s.cfg.PerThread[threadID]

	start := taskID * evalSubChunkSize
	stop := start + evalSubChunkSize
	if stop > mid {
		stop = mid
	}
	subLen := stop - start
	top := start + mid

	// Resize scratch views to actual subLen (only last chunk may differ).
	gateOut := scratch.GateOut[:subLen]
	eqs := scratch.Eqs[:subLen]
	dEqs := scratch.DEqs[:subLen]
	xs := scratch.Xs[:subLen*nInputs]
	dxs := scratch.DXs[:subLen*nInputs]

	evalBuf := scratch.EvalBuf[:nInputs]

	// -----------------------------------------------------------------------
	// t = 0: gate inputs point directly at the bottom half (no copy).
	// -----------------------------------------------------------------------
	for k := 0; k < nInputs; k++ {
		evalBuf[k] = s.tables[k][start:stop]
	}
	s.gate.EvalBatch(gateOut, evalBuf...)

	eqBottom := s.eq[start:stop]
	for j := 0; j < subLen; j++ {
		var v field.Ext
		v.Mul(&eqBottom[j], &gateOut[j])
		scratch.Accum[0].Add(&scratch.Accum[0], &v)
	}

	// -----------------------------------------------------------------------
	// Initialise incremental scheme at t = 1 for the t ≥ 2 loop.
	// eqs[j] = eq[top+j],  dEqs[j] = eq[top+j] − eq[start+j]
	// xs[k*subLen+j] = table[k][top+j],  dxs[k*subLen+j] = top − bottom
	// -----------------------------------------------------------------------
	eqTop := s.eq[top : top+subLen]
	for j := 0; j < subLen; j++ {
		eqs[j] = eqTop[j]
		dEqs[j].Sub(&eqTop[j], &eqBottom[j])
	}

	for k := 0; k < nInputs; k++ {
		kOff := k * subLen
		colTop := s.tables[k][top : top+subLen]
		colBot := s.tables[k][start:stop]
		for j := 0; j < subLen; j++ {
			xs[kOff+j] = colTop[j]
			dxs[kOff+j].Sub(&colTop[j], &colBot[j])
		}
		scratch.EvalBuf[k] = xs[kOff : kOff+subLen]
	}

	// -----------------------------------------------------------------------
	// t = 2 … d: advance eqs and xs by their deltas, evaluate gate, accumulate.
	// Accum[t-1] collects P(t) (Gruen: index 0 = P(0), 1 = P(2), …).
	// -----------------------------------------------------------------------
	for t := 2; t <= d; t++ {
		// Advance eq values.
		for j := 0; j < subLen; j++ {
			eqs[j].Add(&eqs[j], &dEqs[j])
		}
		// Advance all table columns in a single flat loop (cache-friendly).
		for kj := 0; kj < nInputs*subLen; kj++ {
			xs[kj].Add(&xs[kj], &dxs[kj])
		}

		s.gate.EvalBatch(gateOut, evalBuf...)

		for j := 0; j < subLen; j++ {
			var v field.Ext
			v.Mul(&eqs[j], &gateOut[j])
			scratch.Accum[t-1].Add(&scratch.Accum[t-1], &v)
		}
	}
}

// FoldAndAdvance folds eq and all input tables on the first variable using
// challenge r, updates the current claim to roundPoly.EvalAt(r, claim), and
// advances the round counter.
func (s *ProverState) FoldAndAdvance(roundPoly RoundPoly, challenge field.Ext) {
	mid := len(s.eq) / 2

	// Fold eq table.
	polynomials.FoldInto(
		field.VecFromExt(s.eq[:mid]),
		field.VecFromExt(s.eq),
		field.ElemFromExt(challenge),
	)
	s.eq = s.eq[:mid]

	// Fold each input table.
	for k := range s.tables {
		polynomials.FoldInto(
			field.VecFromExt(s.tables[k][:mid]),
			field.VecFromExt(s.tables[k]),
			field.ElemFromExt(challenge),
		)
		s.tables[k] = s.tables[k][:mid]
	}

	// Update claim.
	s.claim = roundPoly.EvalAt(challenge, s.claim)
	s.round++
}

// FinalClaims returns the evaluation of each table and the eq polynomial at
// the accumulated challenge point (one per variable, in round order).
// Must be called after all logN rounds are complete.
//
// Return order: [eq_claim, table[0]_claim, table[1]_claim, …]
func (s *ProverState) FinalClaims() []field.Ext {
	if s.round != s.logN {
		panic(fmt.Sprintf("sumcheck: FinalClaims: only %d of %d rounds complete", s.round, s.logN))
	}
	out := make([]field.Ext, 1+len(s.tables))
	out[0] = s.eq[0]
	for k, t := range s.tables {
		out[k+1] = t[0]
	}
	return out
}

// Claim returns the current expected sum (useful after FoldAndAdvance to pass
// to the verifier's consistency check).
func (s *ProverState) Claim() field.Ext { return s.claim }

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// liftTables lifts base-field tables into cfg.Tables[:len(tables)] and returns
// the ext-field slice. cfg.Tables[k] must have length ≥ n.
func liftTables(cfg *ProverConfig, tables [][]field.Element, n int) [][]field.Ext {
	ext := cfg.Tables[:len(tables)]
	for k, t := range tables {
		col := ext[k][:n]
		for i, e := range t {
			col[i] = field.Lift(e)
		}
	}
	return ext
}

// finishProverState wires up a ProverState after cfg.Tables and cfg.EqScratch
// have been populated. It ensures per-thread Accum slices are large enough for
// the gate degree.
func finishProverState(cfg *ProverConfig, gate Gate, extTables [][]field.Ext,
	eq []field.Ext, logN int, claim field.Ext) *ProverState {
	d := gate.Degree()
	for t := range cfg.PerThread {
		if len(cfg.PerThread[t].Accum) < d {
			cfg.PerThread[t].Accum = make([]field.Ext, d)
		}
	}
	return &ProverState{
		tables: extTables,
		eq:     eq,
		cfg:    cfg,
		gate:   gate,
		claim:  claim,
		logN:   logN,
	}
}

// buildEqTable fills dst with FoldedEqTableExt(qPrime) using parallelism from cfg.
// An optional multiplier is forwarded to FoldedEqTableExt / ChunkOfEqTableExt.
func buildEqTable(cfg *ProverConfig, dst []field.Ext, qPrime []field.Ext, multiplier ...field.Ext) {
	n := len(dst)
	if n < eqChunkSize {
		polynomials.FoldedEqTableExt(dst, qPrime, multiplier...)
		return
	}
	nChunks := n / eqChunkSize
	parallel.ExecuteThreadAware(
		nChunks,
		func(_ int) {},
		func(taskID, _ int) {
			chunk := dst[taskID*eqChunkSize : (taskID+1)*eqChunkSize]
			polynomials.ChunkOfEqTableExt(chunk, taskID, eqChunkSize, qPrime, multiplier...)
		},
		cfg.NumCPU,
	)
}
