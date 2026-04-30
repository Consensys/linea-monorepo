package distributed

import (
	"github.com/consensys/linea-monorepo/prover-ray/arithmetization"
)

// ArithmetizationBlueprint converts a SegmentKind into the value type that
// arithmetization needs to segment the trace for this kind.
//
// This is NOT the prover telling arithmetization what to do. Both sides
// independently compile the same .bin file via CompileSegmentKind and arrive
// at the same column structure. ArithmetizationBlueprint is simply a
// projection of the SegmentKind into the subset of fields arithmetization
// needs, expressed as the cross-boundary value type arithmetization.ModuleBlueprint.
//
// The round-0 columns from the FSSchedule become the LPPColumnIDs in the
// blueprint. Arithmetization uses these to identify which columns to extract
// early (preflight), without knowing anything about Fiat-Shamir or LPP —
// from its perspective they are just "the columns I must emit first".
func (sk *SegmentKind) ArithmetizationBlueprint(kindIndex int) arithmetization.ModuleBlueprint {
	return arithmetization.ModuleBlueprint{
		ModuleIndex: kindIndex,
		// Round-0 columns of the FS schedule = the preflight columns.
		// Arithmetization does not need to know *why* these are special;
		// it just extracts them early.
		LPPColumnIDs:               sk.PreflightColumnIDs(),
		AllColumnIDs:               sk.AllColumns,
		ReceivedValuesAccRoots:     sk.ReceivedValuesAccRoots,
		ReceivedValuesAccPositions: sk.ReceivedValuesAccPositions,
		N0SelectorIDs:              sk.N0SelectorIDs,
	}
}

// ArithmetizationConfig builds the arithmetization.Config from a slice of
// compiled segment kinds. The result is passed to
// arithmetization.Arithmetization.Configure once at setup time.
//
// The prover calls this after compiling all .bin files. Arithmetization
// can also compute the identical Config by compiling the same .bin files
// itself — both routes produce the same output because the source of truth
// is the .bin file, not either layer's internal representation.
func ArithmetizationConfig(kinds []*SegmentKind) arithmetization.Config {
	blueprints := make([]arithmetization.ModuleBlueprint, len(kinds))
	for i, sk := range kinds {
		blueprints[i] = sk.ArithmetizationBlueprint(i)
	}
	// VkMerkleRoot is taken from the first kind; all kinds in one proving job
	// share the same verification-key tree root.
	var vkRoot [8]interface{}
	_ = vkRoot
	return arithmetization.Config{
		Blueprints:   blueprints,
		VkMerkleRoot: kinds[0].VkMerkleRoot,
	}
}
