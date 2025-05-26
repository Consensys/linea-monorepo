package assets

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/utils"
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

func TestSerdeIOP(t *testing.T) {

	// logrus.Printf("Column exists in recursion input iop:%v\n", testCompInput.Columns.Exists(ifaces.ColID(CHECK_COLUMN_NAME_1)))
	// logrus.Printf("Column exists in recur-segment recur iop:%v\n", testCompRecur.Columns.Exists(ifaces.ColID(CHECK_COLUMN_NAME_1)))
	// logrus.Printf("Column exists in recursion input iop:%v\n", testCompInput.Columns.Exists(ifaces.ColID(CHECK_COLUMN_NAME_2)))
	// logrus.Printf("Column exists in recur-segment recur iop:%v\n", testCompRecur.Columns.Exists(ifaces.ColID(CHECK_COLUMN_NAME_2)))

	startTime := time.Now()
	serComp, err := serialization.SerializeCompiledIOP(testComp)
	if err != nil {
		t.Fatalf("error during ser. recursion input compiled-iop:%s\n", err.Error())
	}

	logrus.Printf("Writing ser. CompiledIOP to file")
	logrus.Printf("Serialization took %vs\n", time.Since(startTime).Seconds())

	deSerComp, err := serialization.DeserializeCompiledIOP(serComp)
	if err != nil {
		t.Fatalf("error during deser. recursion input compiled-iop:%s\n", err.Error())
	}

	logrus.Printf("Deserialization took %vs\n", time.Since(startTime).Seconds())

	if !test_utils.CompareExportedFields(testComp, deSerComp) {
		t.Errorf("Mismatch in exported fields after RecursedCompiledIOP serde")
	}
}

var COMPILED_FILE_PATH = "bin/serialized_compiled_iop.bin"

func ensureAssetsDirectory(dirName string) error {
	err := os.MkdirAll(dirName, 0755)
	if err != nil {
		return fmt.Errorf("failed to create assets directory: %w", err)
	}
	return nil
}

func TestSerIOP(t *testing.T) {
	// Ensure the `assets/` directory exists
	if err := ensureAssetsDirectory("bin"); err != nil {
		t.Fatalf("error ensuring assets directory exists: %s\n", err.Error())
	}

	// Start timing the serialization process
	startTime := time.Now()

	// Serialize the `testComp` object
	serComp, err := serialization.SerializeCompiledIOP(testComp)
	if err != nil {
		t.Fatalf("error during serialization of recursion input compiled-iop: %s\n", err.Error())
	}

	logrus.Printf("Serialization took %vs\n", time.Since(startTime).Seconds())

	writeTime := time.Now()

	// Write the serialized data to a file in the `assets/` directory
	err = utils.WriteToFile(COMPILED_FILE_PATH, bytes.NewReader(serComp))
	if err != nil {
		t.Fatalf("error writing serialized CompiledIOP to file: %s\n", err.Error())
	}

	logrus.Printf("Serialized CompiledIOP written to file: %s and took %vs\n", COMPILED_FILE_PATH, time.Since(writeTime).Seconds())
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
