package assets

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
)

var testRec = recurSegComp.Recursion

func TestSerdeRecursion(t *testing.T) {

	if testRec == nil {
		t.Skip("Skipping TestSerdeRecursion due to nil recursion struct")
	}

	// Serialize
	serialized, err := SerializeRecursion(testRec)
	if err != nil {
		t.Fatalf("Serialization failed: %v", err)
	}

	fmt.Println("Serialization recursion successful")

	// Deserialize
	deserialized, err := DeserializeRecursion(serialized)
	if err != nil {
		t.Fatalf("Deserialization failed: %v", err)
	}

	fmt.Println("Deserialization recursion successful")

	// Compare
	if !test_utils.CompareExportedFields(testRec, deserialized) {
		t.Errorf("Original and deserialized Recursion structs don't match")
	}

}

var testComp = testRec.InputCompiledIOP

func TestSerdeRecurIOP(t *testing.T) {

	serComp, err := serialization.SerializeCompiledIOP(testComp)
	if err != nil {
		t.Fatalf("error during ser. recursion input compiled-iop:%s\n", err.Error())
	}

	deSerComp, err := serialization.DeserializeCompiledIOP(serComp)
	if err != nil {
		t.Fatalf("error during deser. recursion input compiled-iop:%s\n", err.Error())
	}

	if !test_utils.CompareExportedFields(testComp, deSerComp) {
		t.Errorf("Mismatch in exported fields after RecursedCompiledIOP serde")
	}
}
