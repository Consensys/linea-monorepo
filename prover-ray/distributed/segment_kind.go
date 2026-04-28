package distributed

import (
	"github.com/consensys/linea-monorepo/prover-ray/arithmetization"
	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
)

// SegmentKind represents one compiled .bin file.
//
// The arithmetization team produces N static .bin files, one per "kind" of
// segment (e.g. keccak, ALU, EC precompiles). Each .bin file is compiled once,
// at setup time, by both the arithmetization layer and the prover layer
// independently. The .bin file is the shared contract: both sides derive the
// same column structure and the same FSSchedule from it, without any runtime
// communication of circuit internals.
//
// The prover does not decide what a segment kind proves; that is fixed by the
// .bin file the arithmetization team ships.
type SegmentKind struct {
	// Name identifies this segment kind (e.g. "keccak", "alu", "ec").
	Name string

	// FSSchedule is the Fiat-Shamir schedule derived from the .bin compilation.
	//
	// FSSchedule[r] is the list of column IDs committed at round r of the
	// interactive protocol, i.e. the columns that feed the Fiat-Shamir state
	// before the verifier's r-th random challenge is drawn.
	//
	// Round 0 is special: it is the only round whose commitments contribute to
	// the *shared* Fiat-Shamir state (the shared randomness). Columns in
	// FSSchedule[0] are therefore the "preflight columns" — the ones
	// arithmetization must emit early so the prover can compute shared
	// randomness before the full trace is ready.
	//
	// Rounds 1..R are segment-local; their challenges are derived from the
	// per-segment transcript and are not shared across segment kinds.
	FSSchedule [][]arithmetization.ColumnID

	// AllColumns is the full set of columns for this segment kind, in the
	// order they appear in the expanded trace.
	AllColumns []arithmetization.ColumnID

	// ReceivedValuesAccRoots and ReceivedValuesAccPositions describe, for each
	// global received value, the (column, row) pair used to forward the running
	// accumulator from one segment to the next. This encodes the LogUp bus
	// "carry" between consecutive segments of the same kind.
	ReceivedValuesAccRoots     []arithmetization.ColumnID
	ReceivedValuesAccPositions []int

	// N0SelectorIDs carries the Horner-query selector column IDs used to
	// compute the N0Values boundary counters for the LPP witness.
	// One inner slice per Horner part.
	N0SelectorIDs [][]arithmetization.ColumnID

	// VkMerkleRoot is the Merkle root of the prover's verification-key tree
	// for this segment kind. Embedded verbatim in every witness.
	VkMerkleRoot field.Octuplet
}

// PreflightColumnIDs returns the columns that must appear in the preflight
// output for this segment kind.
//
// These are exactly FSSchedule[0]: the columns committed at round 0, whose
// commitments determine the shared Fiat-Shamir randomness. All other rounds
// are segment-local and do not need to be emitted early.
func (sk *SegmentKind) PreflightColumnIDs() []arithmetization.ColumnID {
	if len(sk.FSSchedule) == 0 {
		return nil
	}
	return sk.FSSchedule[0]
}

// CompileSegmentKind compiles a .bin file into a SegmentKind.
//
// Both the arithmetization layer and the prover call this function with the
// same .bin file. The resulting SegmentKind is identical on both sides, so no
// runtime communication of circuit structure is needed.
//
// PSEUDO: real implementation calls go-corset's binfile.Parse + air.Compile,
// then walks the resulting AIR schema to extract the FSSchedule, AllColumns,
// ReceivedValues* and N0Selector* fields.
func CompileSegmentKind(binPath string, vkMerkleRoot field.Octuplet) (*SegmentKind, error) {
	panic("pseudo: parse .bin, compile AIR schema, extract FS schedule and column lists")
}

// compiledProver is the prover circuit compiled from a SegmentKind.
// It is used by the prover to prove individual segment witnesses.
//
// PSEUDO: real type wraps a wiop.System + compiled polynomial commitment scheme.
type compiledProver struct {
	kind *SegmentKind
	// PSEUDO: circuit *wiop.System, etc.
}

// CompileProver produces the prover circuit for one segment kind.
// Called by the prover at setup time, after CompileSegmentKind.
//
// PSEUDO: real implementation runs the wiop compiler pipeline over the
// SegmentKind's AIR schema.
func CompileProver(sk *SegmentKind) (*compiledProver, error) {
	panic("pseudo: compile wiop circuit from SegmentKind AIR schema")
}
