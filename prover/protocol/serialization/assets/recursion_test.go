package assets

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
	"github.com/sirupsen/logrus"
)

var (
	CHECK_COLUMN_NAME_1 = "wizard-recursion_PI_0"
	CHECK_COLUMN_NAME_2 = ".PRECOMPUTED_2_I_65536_SUBSLICE_0_OVER_8"
)

var (
	testRec       = recurSegComp.Recursion
	testCompInput = testRec.InputCompiledIOP
	testCompRecur = recurSegComp.RecursionComp
	testComp      = testCompRecur
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

func TestSerdeRecurIOP(t *testing.T) {

	// logrus.Printf("Column exists in recursion input iop:%v\n", testCompInput.Columns.Exists(ifaces.ColID(CHECK_COLUMN_NAME_1)))
	// logrus.Printf("Column exists in recur-segment recur iop:%v\n", testCompRecur.Columns.Exists(ifaces.ColID(CHECK_COLUMN_NAME_1)))
	// logrus.Printf("Column exists in recursion input iop:%v\n", testCompInput.Columns.Exists(ifaces.ColID(CHECK_COLUMN_NAME_2)))
	// logrus.Printf("Column exists in recur-segment recur iop:%v\n", testCompRecur.Columns.Exists(ifaces.ColID(CHECK_COLUMN_NAME_2)))

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

/*
func TestSerdePlonkCkt(t *testing.T) {

	// fmt.Println("Given circuit")
	// fmt.Println(testCkt)

	serCkt, err := SerializePlonkCktInWizard(testCkt)
	if err != nil {
		t.Fatalf("error during ser. plonk circuit")
	}

	logrus.Println("Succesfully ser. plonk circuit in wizard")

	deSerCkt, err := DeSerializePlonkCktInWizard(serCkt, testComp)
	if err != nil {
		t.Fatalf("error during ser. plonk circuit")
	}

	logrus.Println("Succesfully deser. plonk circuit in wizard")

	if !test_utils.CompareExportedFields(testCkt, deSerCkt) {
		t.Fatalf("Mistach in serde. plonk circuit exported fields")
	}
} */
