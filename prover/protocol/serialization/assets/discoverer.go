package assets

import (
	"fmt"
	"reflect"

	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
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

	// fmt.Printf("DeserilizedVal:%v\n", deserializedDiscVal)
	// Type assertion to ensure the deserialized value is of the correct type
	deserializedDisc, ok := deserializedDiscVal.Interface().(distributed.ModuleDiscoverer)
	if !ok {
		return nil, fmt.Errorf("deserialized value is not distributed.ModuleDiscoverer: got %T", deserializedDiscVal.Interface())
	}

	return deserializedDisc, nil
}
