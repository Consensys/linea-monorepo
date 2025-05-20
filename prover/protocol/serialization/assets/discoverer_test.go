package assets

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
)

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

func TestSerdeDWModDisc(t *testing.T) {
	disc := dw.Disc
	if disc == nil {
		t.Skipf("Module serializer is nil")
	}

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
