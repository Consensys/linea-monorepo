package wiop

import "github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"

// Proof is the verifier-visible transcript produced by [System.Prove]. It
// carries only the data a verifier is entitled to see:
//
//   - public columns (Visibility == VisibilityPublic); oracle and internal
//     columns are deliberately omitted;
//   - cells (which are always public — see [Cell.Visibility]);
//   - coins (which are always public — see [CoinField.Visibility]);
//   - the per-Runtime size of each dynamic module, so the verifier can
//     reconstruct module domains without re-assigning columns.
//
// A Proof is self-contained: [System.Verify] rebuilds a fresh [Runtime] from it
// and runs every verifier action against that data alone. The fields are
// unexported for now; they can be promoted once a serialized proof format is
// needed.
type Proof struct {
	columns      map[ObjectID]*ConcreteVector
	cells        map[ObjectID]field.Gen
	coins        map[ObjectID]field.Gen
	dynamicSizes map[int]int
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
// then extracts the verifier-visible transcript into the returned Proof.
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
		columns:      make(map[ObjectID]*ConcreteVector),
		cells:        make(map[ObjectID]field.Gen),
		coins:        make(map[ObjectID]field.Gen),
		dynamicSizes: make(map[int]int),
	}
	for _, r := range sys.Rounds {
		for _, col := range r.Columns {
			if col.Visibility == VisibilityPublic && rt.HasColumnAssignment(col) {
				proof.columns[col.Context.ID] = rt.GetColumnAssignment(col)
			}
		}
		for _, cell := range r.Cells {
			// GetCellValue resolves lazily-assigned openings (e.g. endpoint and
			// quotient/evaluation claims) so their values are captured.
			proof.cells[cell.Context.ID] = rt.GetCellValue(cell)
		}
		for _, coin := range r.Coins {
			proof.coins[coin.Context.ID] = rt.GetCoinValue(coin)
		}
	}
	for k, v := range rt.dynamicSizes {
		proof.dynamicSizes[k] = v
	}
	return proof
}

// Verify reconstructs a [Runtime] from the public data in proof and runs every
// verifier action registered on sys. It returns the first failing check, or nil
// if all checks pass.
//
// Verify never needs oracle or internal columns: it loads the proof's public
// columns, cells, and coins directly, and does not advance the Fiat-Shamir
// transcript (which would require the oracle columns it does not have). This
// relies on every [VerifierAction] reading only public columns, cells, and
// coins.
func (sys *System) Verify(proof Proof) error {
	rt := NewRuntime(sys) // preloads the precomputed columns

	// Load the verifier-visible transcript directly into the runtime maps.
	// Pre-loading cells means resolveLazyCell returns the stored value instead
	// of invoking an assigner that would read an absent oracle column.
	for id, v := range proof.columns {
		rt.columns[id] = v
	}
	for id, v := range proof.cells {
		rt.cells[id] = v
	}
	for id, v := range proof.coins {
		rt.coins[id] = v
	}
	for k, v := range proof.dynamicSizes {
		rt.dynamicSizes[k] = v
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
