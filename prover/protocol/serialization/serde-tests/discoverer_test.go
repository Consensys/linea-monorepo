package serdetests

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
)

// SerializeDisc serializes a ModuleDiscoverer instance.
func SerializeDisc(disc distributed.ModuleDiscoverer) ([]byte, error) {
	return serialization.SerializeValue(reflect.ValueOf(&disc), serialization.DeclarationMode)
}

// DeserializeDisc deserializes a byte slice into a ModuleDiscoverer instance.
func DeserializeDisc(data []byte) (distributed.ModuleDiscoverer, error) {
	// Create a new empty CompiledIOP for deserialization
	comp := serialization.NewEmptyCompiledIOP()

	// Perform deserialization
	deserializedDiscVal, err := serialization.DeserializeValue(data, serialization.DeclarationMode, reflect.TypeOf((*distributed.ModuleDiscoverer)(nil)).Elem(), comp)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize discoverer: %w", err)
	}

	// Type assertion to ensure the deserialized value is of the correct type
	deserializedDisc, ok := deserializedDiscVal.Interface().(distributed.ModuleDiscoverer)
	if !ok {
		return nil, fmt.Errorf("deserialized value is not distributed.ModuleDiscoverer: got %T", deserializedDiscVal.Interface())
	}

	return deserializedDisc, nil
}

// TestSerdeDisc tests serialization and deserialization of the StandardModuleDiscoverer.
func TestSerdeDisc(t *testing.T) {
	var (
		zkevm = test_utils.GetZkEVM()
		disc  = &distributed.StandardModuleDiscoverer{
			TargetWeight: 1 << 28,
			Affinities:   test_utils.GetAffinities(zkevm),
			Predivision:  1,
		}
	)

	// Serialize the discoverer
	discSer, err := SerializeDisc(disc)
	if err != nil {
		t.Fatalf("error during serializing discoverer: %s", err.Error())
	}

	// Deserialize the discoverer
	deserializedDisc, err := DeserializeDisc(discSer)
	if err != nil {
		t.Fatalf("error during deserializing discoverer: %s", err.Error())
	}

	// Compare structs while ignoring unexported fields
	if !test_utils.CompareExportedFields(disc, deserializedDisc) {
		t.Fatalf("Mis-matched fields after serde discoverer (ignoring unexported fields)")
	}
}
