package assets

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
)

var gl = dw.GLs[0]

// TestSerdeGL tests full serialization and deserialization of a ModuleGL.
func TestSerdeGL(t *testing.T) {
	serializedData, err := SerializeModuleGL(gl)
	if err != nil {
		t.Fatalf("Failed to serialize ModuleGL: %v", err)
	}

	deserializedGL, err := DeserializeModuleGL(serializedData)
	if err != nil {
		t.Fatalf("Failed to deserialize ModuleGL: %v", err)
	}

	if !test_utils.CompareExportedFields(gl, deserializedGL) {
		t.Errorf("Mismatch in exported fields after full ModuleGL serde")
	}
}
