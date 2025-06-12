package assets

import (
	"reflect"
	"runtime/debug"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
	"github.com/sirupsen/logrus"
)

var (
	z = test_utils.GetZkEVM()
)

// Helper function for serialization and deserialization tests
func runSerdeTest(t *testing.T, input interface{}, name string, isSanityCheck bool) {

	// In case the test panics, log the error but do not let the panic
	// interrupt the test.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic during serialization/deserialization of %s: %v", name, r)
			debug.PrintStack()
		}
	}()

	if input == nil {
		t.Error("test input is nil")
		return
	}

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
	t.Logf("%s serialization=%v deserialization=%v buffer-size=%v \n", name, serTime, desTime, len(b))

	if isSanityCheck {
		// Sanity check: Compare exported fields
		t.Logf("Running sanity checks on deserialized object: Comparing if the values matched before and after serialization")
		outputDeref := reflect.ValueOf(output).Elem().Interface()
		if !test_utils.CompareExportedFields(input, outputDeref) {
			t.Errorf("Mismatch in exported fields of %s during serde", name)
		}
	}
}

func TestSerdeZkEVM(t *testing.T) {
	runSerdeTest(t, z, "ZkEVM", false)
}
