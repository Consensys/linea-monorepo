package serde_test

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"testing"
	"time"

	"github.com/consensys/linea-monorepo/prover/protocol/serde"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

var (
	z = zkevm.GetTestZkEVM()
)

// Helper function for serialization and deserialization tests
func runSerdeTest(t *testing.T, input any, name string, isSanityCheck, failFast bool) {

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
		b, err = serde.Serialize(input)
		if err != nil {
			t.Fatalf("Error during serialization of %s: %v", name, err)
		}
	})

	// Measure deserialization time
	desTime := profiling.TimeIt(func() {
		logrus.Printf("Starting to deserialize:%s\n", name)
		err = serde.Deserialize(b, output)
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
		if !serde.DeepCmp(input, outputDeref, failFast) {
			t.Errorf("Mismatch in exported fields of %s during serde", name)
		} else {
			t.Logf("Sanity checks passed for %s", name)
		}
	}
}

func TestSerdeZkEVM(t *testing.T) {
	runSerdeTest(t, z, "ZKEVM", true, false)
}

func TestSerdeZkEVMIO(t *testing.T) {
	// 1. Get Real ZkEVM
	z := zkevm.GetTestZkEVM()
	require.NotNil(t, z, "ZkEVM instance should not be nil")

	// 2. Setup Temp Dir
	tmpDir, err := os.MkdirTemp("", "serde_zkevm_io")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Logf("Testing IO with ZkEVM instance...")

	// --- Case A: Uncompressed (Zero-Copy Mmap) ---
	t.Run("Uncompressed_Mmap", func(t *testing.T) {
		path := filepath.Join(tmpDir, "zkevm_raw.bin")

		// Measure Store Time
		start := time.Now()
		err := serde.StoreToDisk(path, z, false)
		require.NoError(t, err)
		storeDur := time.Since(start)

		// Check Size
		info, err := os.Stat(path)
		require.NoError(t, err)
		sizeMB := float64(info.Size()) / 1024 / 1024

		t.Logf("[Mmap] Store Time: %v | Size: %.2f MB", storeDur, sizeMB)

		// Measure Load Time
		var zLoaded zkevm.ZkEvm
		start = time.Now()

		// 1. Capture the closer returned by LoadFromDisk.
		// This is critical for Mmap mode to prevent the GC from unmapping memory while we use it.
		closer, err := serde.LoadFromDisk(path, &zLoaded, false)
		require.NoError(t, err)

		// 2. Defer closing the handle. This ensures the memory-mapped data stays valid
		// for the duration of the DeepCmp check below.
		defer closer.Close()

		loadDur := time.Since(start)
		t.Logf("[Mmap] Load Time:  %v (Should be very fast)", loadDur)

		// 3. Sanity Check
		// We pass &zLoaded to ensure we are comparing (*zkevm.ZkEvm) with (*zkevm.ZkEvm).
		if !serde.DeepCmp(z, &zLoaded, true) {
			t.Errorf("Mismatch in exported fields of ZkEVM during serde i/o - Uncompressed (Mmap)")
		} else {
			t.Log("Sanity checks passed")
		}
	})

	// --- Case B: Compressed (Zstd + Heap) ---
	t.Run("Compressed_Zstd", func(t *testing.T) {
		path := filepath.Join(tmpDir, "zkevm_zstd.bin")

		// Measure Store Time
		start := time.Now()
		err := serde.StoreToDisk(path, z, true)
		require.NoError(t, err)
		storeDur := time.Since(start)

		// Check Size
		info, err := os.Stat(path)
		require.NoError(t, err)
		sizeMB := float64(info.Size()) / 1024 / 1024

		t.Logf("[Zstd] Store Time: %v | Size: %.2f MB", storeDur, sizeMB)

		// Measure Load Time
		var zLoaded zkevm.ZkEvm
		start = time.Now()

		// Capture closer (it will be a no-op closer in compressed mode, but signature must match).
		closer, err := serde.LoadFromDisk(path, &zLoaded, true)
		require.NoError(t, err)
		defer closer.Close()

		loadDur := time.Since(start)
		t.Logf("[Zstd] Load Time:  %v (Includes Decompression)", loadDur)

		// Sanity Check
		if !serde.DeepCmp(z, &zLoaded, true) {
			t.Errorf("Mismatch in exported fields of ZkEVM during serde i/o - Compressed (Zstd)")
		} else {
			t.Log("Sanity checks passed")
		}
	})
}

func justserde(t *testing.B, input any, name string) {
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

	b, err = serde.Serialize(input)
	if err != nil {
		t.Fatalf("Error during serialization of %s: %v", name, err)
	}

	err = serde.Deserialize(b, output)
	if err != nil {
		t.Fatalf("Error during deserialization of %s: %v", name, err)
	}
}
