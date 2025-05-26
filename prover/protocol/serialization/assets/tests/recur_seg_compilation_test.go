package tests

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization/assets"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

var (
	COMPILED_IOP_FILE_PATH   = "../bin/dw_compiled_def_iop.bin"
	DW_COMPILED_DEFAULT_PATH = "../bin/dw_compiled_def.bin"
)

// Parametrized test function for deserialization using switch case
func testDeserFromFile(t *testing.T, binFile string) {
	// Ensure the file exists
	if _, err := os.Stat(binFile); os.IsNotExist(err) {
		t.Fatalf("serialized file does not exist: %s\n", binFile)
	}

	// Load the serialized data from the file
	loadStartTime := time.Now()
	var serComp bytes.Buffer
	err := utils.ReadFromFile(binFile, &serComp)
	if err != nil {
		t.Fatalf("error reading serialized data from file: %s\n", err.Error())
	}

	logrus.Printf("Loaded serialized data from file: %s and took %vs\n", binFile, time.Since(loadStartTime).Seconds())

	// Start timing the deserialization process
	startTime := time.Now()

	// Use switch case to handle different file paths
	switch binFile {
	case COMPILED_IOP_FILE_PATH:
		_, err = serialization.DeserializeCompiledIOP(serComp.Bytes())
		if err != nil {
			t.Fatalf("error during deserialization of CompiledIOP: %s\n", err.Error())
		}
		logrus.Printf("Deserialization of CompiledIOP took %vs\n", time.Since(startTime).Seconds())

	case DW_COMPILED_DEFAULT_PATH:
		_, err = assets.DeserializeRecursedSegmentCompilation(serComp.Bytes())
		if err != nil {
			t.Fatalf("error during deserialization of RecursedSegmentCompilation: %s\n", err.Error())
		}
		logrus.Printf("Deserialization of RecursedSegmentCompilation took %vs\n", time.Since(startTime).Seconds())

	default:
		t.Fatalf("unknown file path: %s\n", binFile)
	}
}

func TestDeCompIOP(t *testing.T) {
	testDeserFromFile(t, COMPILED_IOP_FILE_PATH)
}

func TestDeRecurSegComp(t *testing.T) {
	testDeserFromFile(t, DW_COMPILED_DEFAULT_PATH)
}
