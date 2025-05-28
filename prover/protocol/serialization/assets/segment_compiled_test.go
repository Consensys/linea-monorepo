package assets

import (
	"bytes"
	"testing"
	"time"

	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
	"github.com/sirupsen/logrus"
)

var recurSegComp = dw.CompiledDefault

// var recurSegComp = dw.CompiledGLs[0]

var COMPILED_DEFAULT_PATH = "bin/serialized_dw_compiled_default.bin"

// TestSerdeRecursedSegmentCompilation tests serialization and deserialization of a RecursedSegmentCompilation.
func TestSerdeRecursedSegmentCompilation(t *testing.T) {
	if dw == nil {
		t.Fatal("DistributedWizard is nil")
	}

	if err := ensureAssetsDirectory("bin"); err != nil {
		t.Fatalf("error ensuring assets directory exists: %s\n", err.Error())
	}

	// Serialize the RecursedSegmentCompilation
	serTime := time.Now()
	serializedData, err := SerializeRecursedSegmentCompilation(recurSegComp)
	if err != nil {
		t.Fatalf("Failed to serialize RecursedSegmentCompilation: %v", err)
	}

	logrus.Printf("Serialization of RecursedSegmentCompilation took %vs \n", time.Since(serTime).Seconds())

	writeTime := time.Now()

	// Write the serialized data to a file in the `assets/` directory
	err = utils.WriteToFile(COMPILED_DEFAULT_PATH, bytes.NewReader(serializedData))
	if err != nil {
		t.Fatalf("error writing serialized CompiledIOP to file: %s\n", err.Error())
	}

	logrus.Printf("Serialization of RecursedSegmentCompilation successfully written to file took %vs \n", time.Since(writeTime).Seconds())

	// Deserialize the RecursedSegmentCompilation
	deSerTime := time.Now()
	deserializedSegComp, err := DeserializeRecursedSegmentCompilation(serializedData)
	if err != nil {
		t.Fatalf("Failed to deserialize RecursedSegmentCompilation: %v", err)
	}

	logrus.Printf("Deserialization of RecursedSegmentCompilation successful took %vs \n", time.Since(deSerTime).Seconds())

	// Compare exported fields
	if !test_utils.CompareExportedFields(recurSegComp, deserializedSegComp) {
		t.Errorf("Mismatch in exported fields after RecursedSegmentCompilation serde")
	}
}
