package assets

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
	"github.com/sirupsen/logrus"
)

func TestSerdeRecursion(t *testing.T) {
	testRec := recurSegComp.Recursion

	if testRec == nil {
		t.Skip("Skipping TestSerdeRecursion due to nil recursion struct")
	}

	// Serialize
	serialized, err := SerializeRecursion(testRec)
	if err != nil {
		t.Fatalf("Serialization failed: %v", err)
	}

	logrus.Println("Serialization recursion successful")

	// Deserialize
	deserialized, err := DeserializeRecursion(serialized)
	if err != nil {
		t.Fatalf("Deserialization failed: %v", err)
	}

	// Compare
	if !test_utils.CompareExportedFields(testRec, deserialized) {
		t.Errorf("Original and deserialized Recursion structs don't match")
	}

}
