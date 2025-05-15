package serdetests

import (
	"reflect"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
)

func TestSerdeDiscoverer(t *testing.T) {
	var (
		zkevm = distributed.GetZkEVM()
		disc  = &distributed.StandardModuleDiscoverer{
			TargetWeight: 1 << 28,
			Affinities:   distributed.GetAffinities(zkevm),
			Predivision:  1,
		}
	)

	discSer, err := serialization.SerializeValue(reflect.ValueOf(disc), serialization.DeclarationMode)
	if err != nil {
		t.Fatalf("error during serializing discover:%s", err.Error())
	}

	// Create a new empty CompiledIOP for deserialization
	comp := serialization.NewEmptyCompiledIOP()

	// Deserialize into a new Modexp
	deserializedDiscVal, err := serialization.DeserializeValue(discSer, serialization.DeclarationMode, reflect.TypeOf(&distributed.StandardModuleDiscoverer{}), comp)
	if err != nil {
		t.Fatalf("Failed to deserialize discoverer: %v", err)
	}
	deserializedDisc, ok := deserializedDiscVal.Interface().(*distributed.StandardModuleDiscoverer)
	if !ok {
		t.Fatalf("Deserialized value is not *modexp.Module: got %T", deserializedDiscVal.Interface())
	}

	// Compare structs while ignoring unexported fields
	if !test_utils.CompareExportedFields(disc, deserializedDisc) {
		t.Fatalf("Mis-matched fields after serde discoverer (ignoring unexported fields)")
	}

}
