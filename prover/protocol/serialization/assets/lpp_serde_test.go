package assets

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
)

var lpp = dw.LPPs[0]

// TestSerdeLPP tests full serialization and deserialization of a ModuleLPP.
func TestSerdeLPP(t *testing.T) {
	serializedData, err := SerializeModuleLPP(lpp)
	if err != nil {
		t.Fatalf("Failed to serialize ModuleLPP: %v", err)
	}

	deserializedLPP, err := DeserializeModuleLPP(serializedData)
	if err != nil {
		t.Fatalf("Failed to deserialize ModuleLPP: %v", err)
	}

	if !test_utils.CompareExportedFields(lpp, deserializedLPP) {
		t.Errorf("Mismatch in exported fields after full ModuleLPP serde")
	}
}
