package serialization_test

import (
	"reflect"
	"runtime/debug"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
)

// runSerdeBenchmark is updated to properly benchmark serialization and deserialization separately.
// For serialization benchmark, it serializes b.N times.
// For deserialization benchmark, it serializes once outside the loop, then deserializes b.N times.
func runSerdeBenchmark(b *testing.B, input any, name string, onlySerialize bool) {
	// In case the test panics, log the error but do not let the panic
	// interrupt the test.
	defer func() {
		if r := recover(); r != nil {
			b.Errorf("Panic during serialization/deserialization of %s: %v", name, r)
			debug.PrintStack()
		}
	}()

	if input == nil {
		b.Fatal("test input is nil")
	}

	var output = reflect.New(reflect.TypeOf(input)).Interface()
	var bBytes []byte
	var err error

	// Serialize once to get the bytes for deserialization benchmark
	if !onlySerialize {
		bBytes, err = serialization.Serialize(input)
		if err != nil {
			b.Fatalf("Error during initial serialization of %s: %v", name, err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if onlySerialize {
			// Benchmark only serialization
			_, err = serialization.Serialize(input)
			if err != nil {
				b.Fatalf("Error during serialization of %s: %v", name, err)
			}
		} else {
			// Benchmark only deserialization (using pre-serialized bytes)
			err = serialization.Deserialize(bBytes, output)
			if err != nil {
				b.Fatalf("Error during deserialization of %s: %v", name, err)
			}
		}
	}
}

// BenchmarkSerZkEVM benchmarks serialization of ZkEVM separately.
func BenchmarkSerZkEVM(b *testing.B) {
	runSerdeBenchmark(b, z, "ZkEVM-Serialize", true)
}

// BenchmarkDeserZkEVM benchmarks deserialization of ZkEVM separately.
func BenchmarkDeserZkEVM(b *testing.B) {
	runSerdeBenchmark(b, z, "ZkEVM-Deserialize", false)
}

func makeTestModuleWitnessGL() *distributed.ModuleWitnessGL {
	cols := make(map[ifaces.ColID]smartvectors.SmartVector)

	addColumn := func(name string, size int) {
		r := make([]field.Element, size)
		for i := range r {
			r[i] = field.NewElement(uint64(i + 1))
		}
		reg := smartvectors.Regular(r)
		cols[ifaces.ColID(name)] = &reg
	}

	// Base columns from previous example
	addColumn("CYCLIC_COUNTER_3079_64_0_COUNTER", 64)
	addColumn("ECDSA_ANTICHAMBER_ID", 64)
	addColumn("ECDSA_ANTICHAMBER_IS_ACTIVE", 64)
	addColumn("ECDSA_ANTICHAMBER_GNARK_DATA_PI", 64)
	addColumn("ECDSA_ANTICHAMBER_TXSIGNATURE_TX_HASH_HI", 64)
	addColumn("ECDSA_ANTICHAMBER_TXSIGNATURE_TX_HASH_LO", 64)

	// Columns you provided
	names := []string{
		"ECDSA_ANTICHAMBER_ADDRESSES_ADDRESS_LO",
		"ECDSA_ANTICHAMBER_ADDRESSES_ADRESSHI_UNTRIMMED",
		"ECDSA_ANTICHAMBER_ADDRESSES_HASH_NUM",
		"ECDSA_ANTICHAMBER_ADDRESSES_ISADRESS_FROM_ECREC",
		"ECDSA_ANTICHAMBER_ADDRESSES_ISADRESS_FROM_TXNDATA",
		"ECDSA_ANTICHAMBER_ADDRESSES_IS_ADDRESS",
		"ECDSA_ANTICHAMBER_ECRECOVER_AUX_PROJECTION_MASK",
		"ECDSA_ANTICHAMBER_ECRECOVER_ECRECOVER_ID",
		"ECDSA_ANTICHAMBER_ECRECOVER_ECRECOVER_INDEX",
		"ECDSA_ANTICHAMBER_ECRECOVER_ECRECOVER_IS_DATA",
		"ECDSA_ANTICHAMBER_ECRECOVER_ECRECOVER_IS_RES",
		"ECDSA_ANTICHAMBER_ECRECOVER_LIMB",
		"ECDSA_ANTICHAMBER_ECRECOVER_SUCCESS_BIT",
		"ECDSA_ANTICHAMBER_GNARK_DATA_IS_ACTIVE",
		"ECDSA_ANTICHAMBER_GNARK_DATA_PI",
		"ECDSA_ANTICHAMBER_ID",
		"ECDSA_ANTICHAMBER_IS_ACTIVE",
		"ECDSA_ANTICHAMBER_IS_FETCHING",
		"ECDSA_ANTICHAMBER_IS_PUSHING",
		"ECDSA_ANTICHAMBER_SOURCE",
		"ECDSA_ANTICHAMBER_TXSIGNATURE_TX_HASH_HI",
		"ECDSA_ANTICHAMBER_TXSIGNATURE_TX_HASH_LO",
		"ECDSA_ANTICHAMBER_TXSIGNATURE_TX_IS_HASH_HI",
		"ECDSA_ANTICHAMBER_UNALIGNED_GNARK_DATA_GNARK_DATA",
		"ECDSA_ANTICHAMBER_UNALIGNED_GNARK_DATA_GNARK_INDEX",
		"ECDSA_ANTICHAMBER_UNALIGNED_GNARK_DATA_GNARK_PUBLIC_KEY_INDEX",
	}

	// Use realistic-ish widths
	for _, n := range names {
		addColumn(n, 64)
	}

	// ReceivedValuesGlobal
	globals := make([]field.Element, 13)
	for i := range globals {
		globals[i] = field.NewElement(0)
	}

	// VkMerkleRoot
	var vk field.Element
	vk.SetUint64(123456789)

	return &distributed.ModuleWitnessGL{
		ModuleName:           distributed.ModuleName("ECDSA"),
		ModuleIndex:          5,
		SegmentModuleIndex:   0,
		TotalSegmentCount:    []int{1, 0, 0, 0, 0, 1, 0, 1, 0, 1, 1, 0, 0, 0, 1, 2},
		Columns:              cols,
		ReceivedValuesGlobal: globals,
		VkMerkleRoot:         vk,
	}
}

func BenchmarkModuleWitnessGL(b *testing.B) {
	w := makeTestModuleWitnessGL()

	b.Run("serialize", func(b *testing.B) {
		runSerdeBenchmark(b, w, "ModuleWitnessGL", true)
	})

	b.Run("deserialize", func(b *testing.B) {
		runSerdeBenchmark(b, w, "ModuleWitnessGL", false)
	})
}
