package wiop

import "github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"

// Proof is the transcript produced by [System.Prove] and consumed by
// [System.Verify]. It carries:
//
//   - every committed column (Visibility >= VisibilityOracle, i.e. oracle and
//     public); internal columns are prover-only and omitted;
//   - cells (always public — see [Cell.Visibility]);
//   - the per-Runtime size of each dynamic module, so the verifier can
//     reconstruct module domains.
//
// It deliberately does NOT carry the verifier coins: [System.Verify] re-derives
// every Fiat-Shamir challenge itself by replaying the transcript, so a prover
// cannot influence the challenges by supplying forged coin values.
//
// Including the oracle columns is a testing convenience: it lets the verifier
// reconstruct the exact Fiat-Shamir state without a commitment scheme (the raw
// oracle values stand in for their commitments). In a production proof an oracle
// column is sent only as its commitment, never in full; the pipeline should
// ultimately guarantee that no VisibilityOracle column survives compilation.
//
// The fields are unexported for now; they can be promoted once a serialized
// proof format is needed.
type Proof struct {
	Columns      map[ObjectID]*ConcreteVector
	Cells        map[ObjectID]field.Gen
	DynamicSizes map[int]int
}

// Prove runs the prover over every interactive round of sys and returns the
// resulting [Proof].
//
// assign is the witness hook: it is called once on a fresh [Runtime] before any
// round is processed and is responsible for assigning the first round's oracle
// columns (and any other prover inputs). This is the seam used by the zkcdriver
// (driver.AssignWithPreRead) and by the test scenarios (AssignHonest /
// AssignWitness).
//
// Prove drives the same prover loop as [wioptest.RunAndVerify]: it runs each
// round's [ProverAction]s, advancing the Fiat-Shamir transcript between rounds,
// then captures the committed columns and cells into the returned Proof. The
// verifier coins are not captured; [System.Verify] re-derives them.
//
// The caller is responsible for running the compiler passes (and, optionally,
// [Materialize]) on sys before calling Prove.
func (sys *System) Prove(assign func(rt *Runtime)) Proof {
	rt := NewRuntime(sys)
	assign(&rt)

	// Run the first round's prover actions before any AdvanceRound, then walk
	// the remaining rounds. Mirrors wioptest.RunAndVerify's prover half.
	for _, a := range rt.CurrentRound().ProverActions {
		a.Run(rt)
	}
	for rt.currentRound.ID < len(sys.Rounds)-1 {
		rt.AdvanceRound()
		for _, a := range rt.CurrentRound().ProverActions {
			a.Run(rt)
		}
	}

	proof := Proof{
		Columns:      make(map[ObjectID]*ConcreteVector),
		Cells:        make(map[ObjectID]field.Gen),
		DynamicSizes: make(map[int]int),
	}
	for _, r := range sys.Rounds {
		for _, col := range r.Columns {
			// Capture every committed column (oracle + public). These are
			// exactly the columns AdvanceRound feeds into Fiat-Shamir, so the
			// verifier needs them to re-derive the coins.
			if col.Visibility >= VisibilityOracle && rt.HasColumnAssignment(col) {
				proof.Columns[col.Context.ID] = rt.GetColumnAssignment(col)
			}
		}
		for _, cell := range r.Cells {
			// GetCellValue resolves lazily-assigned openings (e.g. endpoint and
			// quotient/evaluation claims) so their values are captured.
			proof.Cells[cell.Context.ID] = rt.GetCellValue(cell)
		}
	}
	for k, v := range rt.dynamicSizes {
		proof.DynamicSizes[k] = v
	}
	return proof
}

// Verify reconstructs a [Runtime] from proof and runs every verifier action
// registered on sys. It returns the first failing check, or nil if all checks
// pass.
//
// Crucially, Verify does not trust the coins: it replays the transcript round by
// round, re-deriving every Fiat-Shamir challenge with [Runtime.AdvanceRound]
// from the committed columns and cells. The prover therefore cannot forge a
// challenge. Verifier actions read the re-derived coins, the cells, and the
// public columns.
func (sys *System) Verify(proof Proof) error {
	rt := NewRuntime(sys) // currentRound = r0, preloads precomputed columns

	// Restore dynamic-module sizes so Module.RuntimeSize resolves during the
	// replay and the subsequent verifier checks.
	for k, v := range proof.DynamicSizes {
		rt.dynamicSizes[k] = v
	}

	// assignRound loads the proof's committed columns and cells for r into the
	// runtime. AssignColumn / AssignCell require r to be the current round, so
	// this is always called on rt.CurrentRound().
	assignRound := func(r *Round) {
		for _, col := range r.Columns {
			if v, ok := proof.Columns[col.Context.ID]; ok {
				rt.AssignColumn(col, v)
			}
		}
		for _, cell := range r.Cells {
			if v, ok := proof.Cells[cell.Context.ID]; ok {
				rt.AssignCell(cell, v)
			}
		}
	}

	// Replay the transcript: assign each round's committed data, then advance
	// (which feeds that data into Fiat-Shamir and re-derives the next round's
	// coins). This reproduces the prover's exact challenge sequence.
	assignRound(rt.CurrentRound())
	for rt.currentRound.ID < len(sys.Rounds)-1 {
		rt.AdvanceRound()
		assignRound(rt.CurrentRound())
	}

	for _, r := range sys.Rounds {
		for _, va := range r.VerifierActions {
			if err := va.Check(rt); err != nil {
				return err
			}
		}
	}
	return nil
}
