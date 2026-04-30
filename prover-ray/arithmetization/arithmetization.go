// Package arithmetization defines the interface between the tracing /
// arithmetization layer and the prover.
//
// # Ownership and .bin files
//
// The arithmetization team ships N static .bin files, one per segment kind
// (e.g. keccak, ALU, EC precompiles). Each .bin file is compiled independently
// by both the arithmetization layer and the prover layer; the .bin file is the
// shared contract. Neither side tells the other what it contains at runtime.
//
// Compiling a .bin file produces a Fiat-Shamir schedule (FSSchedule): a list
// of column-ID groups, one per interactive round. FSSchedule[0] is the set of
// columns committed before any verifier challenge — these are the "preflight
// columns". Their commitments determine the shared Fiat-Shamir randomness that
// all segment provers consume as their initial state.
//
// # Two-milestone output
//
// Arithmetization.Run signals two milestones via channels:
//
//  1. preflightCh — closed as soon as FSSchedule[0] columns are assigned for
//     all segments. The prover reads these, commits to them, and derives the
//     shared randomness, all while arithmetization continues with the rest of
//     the trace.
//
//  2. resultCh — closed when the full trace is expanded and every segment
//     witness is populated. ModuleWitnessLPP.InitialFiatShamirState is zero in
//     every entry; the prover injects the shared randomness it already holds.
//
// Arithmetization never sees or uses shared randomness. It does not know what
// "LPP" or "Fiat-Shamir" mean; it only knows "emit FSSchedule[0] columns
// early, emit everything else once the full trace is ready".
package arithmetization

import (
	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
)

// Config is derived from compiling the N .bin files. It tells arithmetization
// which columns belong to which segment kind, which ones must be emitted early
// (FSSchedule[0] = LPPColumnIDs in the blueprint), and the VK Merkle root to
// embed in witnesses.
//
// Both arithmetization and the prover can compute the identical Config by
// compiling the same .bin files. No runtime hand-off of circuit internals is
// required.
type Config struct {
	// Blueprints is one entry per segment kind (.bin file), in kind-index order.
	Blueprints []ModuleBlueprint
	// VkMerkleRoot is the Merkle root of the prover's verification-key tree,
	// shared across all segment kinds in one proving job.
	VkMerkleRoot field.Octuplet
}

// ModuleBlueprint carries the column-routing information arithmetization needs
// to process one segment kind. It is a plain value type with no pointers to
// circuit internals, safe to serialise or send across a process boundary.
//
// In the context of Fiat-Shamir: LPPColumnIDs = FSSchedule[0] for this kind.
// Arithmetization uses this only to know "emit these columns first"; it does
// not need to understand the FS protocol.
type ModuleBlueprint struct {
	// ModuleIndex is the position of this segment kind in the Config.Blueprints
	// slice (0-based).
	ModuleIndex int
	// LPPColumnIDs is the set of columns to emit in the preflight pass.
	// These are FSSchedule[0] for this segment kind: the columns committed at
	// round 0, whose commitments seed the shared Fiat-Shamir randomness.
	LPPColumnIDs []ColumnID
	// AllColumnIDs is the full set of columns for this segment kind.
	AllColumnIDs []ColumnID
	// ReceivedValuesAccRoots and ReceivedValuesAccPositions describe the LogUp
	// bus "carry" columns: for each globally received value, the (column, row)
	// from which arithmetization reads the running accumulator to forward to the
	// next segment.
	ReceivedValuesAccRoots     []ColumnID
	ReceivedValuesAccPositions []int
	// N0SelectorIDs is the Horner-query selector column IDs, one inner slice
	// per Horner part. Used to compute N0Values for the LPP witness.
	N0SelectorIDs [][]ColumnID
}

// Arithmetization reads execution traces and produces segmented witnesses.
// The zero value is not usable; call Configure before Run.
type Arithmetization struct {
	cfg Config
}

// Configure stores the column-routing config derived from the .bin files.
// Must be called once before Run. Calling it again replaces the config.
func (a *Arithmetization) Configure(cfg Config) {
	a.cfg = cfg
}

// Run processes a single execution trace end-to-end.
//
// It spawns a goroutine that:
//  1. Expands the trace (reads lt-trace, runs go-corset expansion).
//  2. Runs a fast preflight pass: extracts only the LPPColumnIDs
//     (= FSSchedule[0]) for every segment of every kind, and sends the result
//     to preflightCh. The prover reads this channel and immediately starts
//     computing shared randomness.
//  3. Completes full segmentation and sends the result to resultCh.
//     ModuleWitnessLPP.InitialFiatShamirState is zero throughout; the prover
//     fills it in after step 2.
//
// Both channels carry exactly one value. The caller must read both to avoid
// goroutine leaks.
func (a *Arithmetization) Run(tracePath string) (
	preflightCh <-chan []PreflightSegment,
	resultCh <-chan TracingResult,
) {
	pCh := make(chan []PreflightSegment, 1)
	rCh := make(chan TracingResult, 1)

	go func() {
		// PSEUDO: read and expand the trace (go-corset lt-trace parse + AIR
		// expansion). In the real implementation this is the expensive step.
		expanded := expandTrace(tracePath)

		// --- Preflight pass (fast) ---
		// Extract FSSchedule[0] columns (LPPColumnIDs) for every segment.
		// This is a cheap scan: only a small fraction of total column data.
		preflightSegs := runPreflight(expanded, a.cfg)
		pCh <- preflightSegs // prover can now start computing shared randomness

		// --- Full segmentation (slow) ---
		// Extract all columns for all segments. GL witnesses get
		// ReceivedValuesGlobal accumulated segment-by-segment.
		// LPP witnesses get N0Values. InitialFiatShamirState stays zero.
		result := runFullSegmentation(expanded, a.cfg)
		rCh <- result
	}()

	return pCh, rCh
}

// ---------------------------------------------------------------------------
// Internal types and stubs
// ---------------------------------------------------------------------------

type expandedTrace struct{ /* PSEUDO: expanded column map */ }

func expandTrace(_ string) *expandedTrace {
	panic("pseudo: parse lt-trace + run go-corset AIR expansion")
}

func runPreflight(_ *expandedTrace, cfg Config) []PreflightSegment {
	var segs []PreflightSegment
	for _, bp := range cfg.Blueprints {
		nbSegs := pseudoNbSegments(bp)
		for segIdx := range nbSegs {
			seg := PreflightSegment{
				ModuleIndex:  bp.ModuleIndex,
				SegmentIndex: segIdx,
				Columns:      make(map[ColumnID]ColumnData, len(bp.LPPColumnIDs)),
			}
			for _, colID := range bp.LPPColumnIDs {
				seg.Columns[colID] = pseudoSliceColumn(colID, segIdx)
			}
			segs = append(segs, seg)
		}
	}
	return segs
}

func runFullSegmentation(_ *expandedTrace, cfg Config) TracingResult {
	var glWitnesses []*ModuleWitnessGL
	var lppWitnesses []*ModuleWitnessLPP

	totalSegCount := make([]int, len(cfg.Blueprints))
	for _, bp := range cfg.Blueprints {
		totalSegCount[bp.ModuleIndex] = pseudoNbSegments(bp)
	}

	for _, bp := range cfg.Blueprints {
		nbSegs := totalSegCount[bp.ModuleIndex]
		receivedVals := make([]field.Element, len(bp.ReceivedValuesAccRoots))

		for segIdx := range nbSegs {
			glCols := make(map[ColumnID]ColumnData, len(bp.AllColumnIDs))
			for _, id := range bp.AllColumnIDs {
				glCols[id] = pseudoSliceColumn(id, segIdx)
			}
			gl := &ModuleWitnessGL{
				ModuleIndex:          bp.ModuleIndex,
				SegmentIndex:         segIdx,
				TotalSegmentCount:    totalSegCount,
				Columns:              glCols,
				ReceivedValuesGlobal: receivedVals,
				VkMerkleRoot:         cfg.VkMerkleRoot,
			}
			glWitnesses = append(glWitnesses, gl)
			receivedVals = pseudoNextReceivedValues(gl, bp)

			lppCols := make(map[ColumnID]ColumnData, len(bp.LPPColumnIDs))
			for _, id := range bp.LPPColumnIDs {
				lppCols[id] = pseudoSliceColumn(id, segIdx)
			}
			lppWitnesses = append(lppWitnesses, &ModuleWitnessLPP{
				ModuleIndex:       bp.ModuleIndex,
				SegmentIndex:      segIdx,
				TotalSegmentCount: totalSegCount,
				N0Values:          pseudoComputeN0Values(lppCols, bp),
				Columns:           lppCols,
				VkMerkleRoot:      cfg.VkMerkleRoot,
				// InitialFiatShamirState intentionally zero — prover fills this in.
			})
		}
	}
	return TracingResult{WitnessesGL: glWitnesses, WitnessesLPP: lppWitnesses}
}

func pseudoNbSegments(_ ModuleBlueprint) int                                          { panic("pseudo") }
func pseudoSliceColumn(_ ColumnID, _ int) ColumnData                                 { panic("pseudo") }
func pseudoNextReceivedValues(_ *ModuleWitnessGL, _ ModuleBlueprint) []field.Element { panic("pseudo") }
func pseudoComputeN0Values(_ map[ColumnID]ColumnData, _ ModuleBlueprint) []int       { panic("pseudo") }
