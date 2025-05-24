package assets

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
	"github.com/sirupsen/logrus"
)

var recurSegComp = dw.CompiledDefault

// var recurSegComp = dw.CompiledGLs[0]

// TestSerdeRecursedSegmentCompilation tests serialization and deserialization of a RecursedSegmentCompilation.
func TestSerdeRecursedSegmentCompilation(t *testing.T) {
	if dw == nil {
		t.Fatal("DistributedWizard is nil")
	}

	// Serialize the RecursedSegmentCompilation
	serializedData, err := SerializeRecursedSegmentCompilation(recurSegComp)
	if err != nil {
		t.Fatalf("Failed to serialize RecursedSegmentCompilation: %v", err)
	}

	logrus.Println("Serialization of RecursedSegmentCompilation successful with recursion")

	// Deserialize the RecursedSegmentCompilation
	deserializedSegComp, err := DeserializeRecursedSegmentCompilation(serializedData)
	if err != nil {
		t.Fatalf("Failed to deserialize RecursedSegmentCompilation: %v", err)
	}

	logrus.Println("Deserialization of RecursedSegmentCompilation successful with recursion")

	// Compare exported fields
	if !test_utils.CompareExportedFields(recurSegComp, deserializedSegComp) {
		t.Errorf("Mismatch in exported fields after RecursedSegmentCompilation serde")
	}
}
