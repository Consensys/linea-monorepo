package assets

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
	"github.com/sirupsen/logrus"
)

var (
	CHECK_COLUMN_NAME_1 = "wizard-recursion_PI_0"
	CHECK_COLUMN_NAME_2 = ".PRECOMPUTED_2_I_65536_SUBSLICE_0_OVER_8"
)

var (
	testRec = recurSegComp.Recursion
)

func TestSerdeRecursion(t *testing.T) {

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
	deserialized, err := DeserializeRecursion(serialized, testComp)
	if err != nil {
		t.Fatalf("Deserialization failed: %v", err)
	}

	logrus.Println("Deserialization recursion successful")

	// Compare
	if !test_utils.CompareExportedFields(testRec, deserialized) {
		t.Errorf("Original and deserialized Recursion structs don't match")
	}
}
