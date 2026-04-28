// Package arithmetization defines the interface between the tracing/arithmetization
// layer and the prover. Arithmetization owns:
//   - reading and expanding the execution trace
//   - segmenting columns according to a prover-supplied blueprint
//   - emitting a PreflightOutput early (only the looked-up / LPP columns) so the
//     prover can derive shared randomness before the full trace is ready
//   - emitting the full TracingResult once all columns are assigned
//
// Arithmetization never sees or uses shared randomness; that is the prover's
// responsibility.
package arithmetization

import (
	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover-ray/wiop"
)

// ColumnID is an alias for the wiop object identifier restricted to columns.
type ColumnID = wiop.ObjectID

// ColumnData is a concrete vector assignment for one column in one segment.
// It corresponds to a contiguous slice of field elements covering exactly
// [segmentStart, segmentStart+segmentSize).
type ColumnData = field.Vector

// PreflightSegment holds the subset of column assignments for a single
// (module, segment) pair that contribute to the LogUp bus — i.e., the columns
// listed in the corresponding ModuleSegmentationBlueprint.LPPColumnSets.
//
// These are the only columns the prover needs to compute the LPP commitment
// and, from that, the shared Fiat-Shamir randomness. Arithmetization emits
// them as soon as they are assigned, before the full trace is complete.
type PreflightSegment struct {
	// ModuleIndex is the integer index of the module in the blueprint list.
	ModuleIndex int
	// SegmentIndex is the vertical split index within this module.
	SegmentIndex int
	// Columns maps each LPP column ID to its segment-slice assignment.
	// Only columns from ModuleSegmentationBlueprint.LPPColumnSets are present.
	Columns map[ColumnID]ColumnData
}

// TracingResult is produced once arithmetization has finished processing the
// full execution trace. It contains the complete column assignments for every
// module segment, ready to be handed to the GL and LPP provers.
//
// InitialFiatShamirState is intentionally absent from ModuleWitnessLPP entries
// in this struct: setting it is the prover's responsibility, using the shared
// randomness it derived from the earlier PreflightOutput.
type TracingResult struct {
	// WitnessesGL contains one entry per (module, segment) pair for the GL
	// (Global Lookup) prover. Each entry holds all columns assigned to that
	// module segment plus the accumulated global received-values.
	WitnessesGL []*ModuleWitnessGL
	// WitnessesLPP contains one entry per (module, segment) pair for the LPP
	// (Local Polynomial Protocol) prover. Each entry holds only the LPP
	// column set; InitialFiatShamirState is zero-valued and must be filled in
	// by the prover before passing the witness to the LPP circuit.
	WitnessesLPP []*ModuleWitnessLPP
}

// ModuleWitnessGL is the full column assignment for one GL segment.
// Arithmetization produces this; the prover consumes it.
type ModuleWitnessGL struct {
	ModuleIndex  int
	SegmentIndex int
	// TotalSegmentCount[i] is the number of segments produced for module i.
	// Needed by the GL circuit to size its public inputs correctly.
	TotalSegmentCount []int
	// Columns is the complete set of columns assigned to this module segment.
	Columns map[ColumnID]ColumnData
	// ReceivedValuesGlobal are the running accumulated values from the LogUp
	// bus for this segment (forwarded from the previous segment's sent values).
	// Computed sequentially by arithmetization as it iterates segments.
	ReceivedValuesGlobal []field.Element
	// VkMerkleRoot is supplied by the prover at configure time and embedded
	// verbatim so the GL circuit can verify the verification key.
	VkMerkleRoot field.Octuplet
}

// ModuleWitnessLPP is the LPP column assignment for one segment.
// InitialFiatShamirState is left at its zero value here; the prover injects
// the shared randomness before handing this witness to the LPP circuit.
type ModuleWitnessLPP struct {
	ModuleIndex  int
	SegmentIndex int
	// TotalSegmentCount mirrors the GL witness field; same semantics.
	TotalSegmentCount []int
	// InitialFiatShamirState is intentionally zero. The prover sets this
	// field to the shared randomness derived from all LPP commitments before
	// running the LPP circuit.
	InitialFiatShamirState field.Octuplet
	// N0Values are the Horner-query segment boundary counters, computed by
	// arithmetization as part of full segmentation.
	N0Values []int
	// Columns contains only the LPP column set (same as PreflightSegment but
	// may include additional context needed by the LPP prover).
	Columns map[ColumnID]ColumnData
	// VkMerkleRoot is forwarded from the configure-time value.
	VkMerkleRoot field.Octuplet
}
