package assets

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/sirupsen/logrus"
)

var (
	z   = test_utils.GetZkEVM()
	iop = z.WizardIOP
)

// Helper function for serialization and deserialization tests
func runSerdeTest(t *testing.T, input interface{}, output interface{}, name string) {
	var b []byte
	var err error

	/*
		// Measure serialization time
		serTime := profiling.TimeIt(func() {
			logrus.Printf("Starting to serialize:%s \n", name)
			b, err = serialization.Serialize(input)
			if err != nil {
				t.Fatalf("Error during serialization of %s: %v", name, err)
			}
		})

		// Measure deserialization time
		desTime := profiling.TimeIt(func() {
			logrus.Printf("Starting to deserialize:%s\n", name)
			err = serialization.Deserialize(b, output)
			if err != nil {
				t.Fatalf("Error during deserialization of %s: %v", name, err)
			}
		})

		// Log results
		t.Logf("%s serialization=%v deserialization=%v buffer-size=%v", name, serTime, desTime, len(b))

	*/

	logrus.Printf("Starting to serialize:%s \n", name)
	b, err = serialization.Serialize(input)
	if err != nil {
		t.Fatalf("Error during serialization of %s: %v", name, err)
	}

	logrus.Printf("Starting to deserialize:%s\n", name)
	err = serialization.Deserialize(b, output)
	if err != nil {
		t.Fatalf("Error during deserialization of %s: %v", name, err)
	}

	// Sanity check: Compare exported fields
	t.Logf("Running sanity checks on deserialized object: Comparing if the values matched before and after serialization")
	if !test_utils.CompareExportedFields(input, output) {
		t.Fatalf("Mismatch in exported fields of %s during serde", name)
	}
}

func TestSerdeZkEVM(t *testing.T) {
	runSerdeTest(t, z.Ecadd, &zkevm.ZkEvm{}, "ZkEVM")
}

func TestSerdeIOP(t *testing.T) {
	runSerdeTest(t, iop, &wizard.CompiledIOP{}, "CompiledIOP")
}
