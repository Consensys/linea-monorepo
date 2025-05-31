package assets

import (
	"reflect"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
	"github.com/sirupsen/logrus"
)

var (
	z   = test_utils.GetZkEVM()
	iop = z.WizardIOP
)

// Helper function for serialization and deserialization tests
func runSerdeTest(t *testing.T, input interface{}, name string) {

	var output = reflect.New(reflect.TypeOf(input)).Interface()
	var b []byte
	var err error

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

	// Sanity check: Compare exported fields
	t.Logf("Running sanity checks on deserialized object: Comparing if the values matched before and after serialization")

	outputDeref := reflect.ValueOf(output).Elem().Interface()
	if !test_utils.CompareExportedFields(input, outputDeref) {
		t.Fatalf("Mismatch in exported fields of %s during serde", name)
	}
}

func TestSerdeZkEVM(t *testing.T) {
	runSerdeTest(t, z, "ZkEVM")
}

func TestSerdeIOP(t *testing.T) {
	runSerdeTest(t, iop, "CompiledIOP")
}
