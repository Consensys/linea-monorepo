package assets

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
)

var recurSegComp = dw.CompiledDefault

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

	fmt.Println("Serialization of RecursedSegmentCompilation successful w/o recursion")

	// Deserialize the RecursedSegmentCompilation
	deserializedSegComp, err := DeserializeRecursedSegmentCompilation(serializedData)
	if err != nil {
		t.Fatalf("Failed to deserialize RecursedSegmentCompilation: %v", err)
	}

	fmt.Println("Deserialization of RecursedSegmentCompilation successful w/o recursion")

	// Compare exported fields
	if !test_utils.CompareExportedFields(recurSegComp, deserializedSegComp) {
		t.Errorf("Mismatch in exported fields after RecursedSegmentCompilation serde")
	}
}
