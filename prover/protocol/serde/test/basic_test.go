package serde_test

import (
	"os"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/serde"
	"github.com/stretchr/testify/require"
)

// --- Shared Test Types & Data ---

// Define the filename as a constant so both tests access the exact same file
const TestFileName = "integration_test_data.bin"

type InnerData struct {
	ID   int
	Name string
}

type ComplexData struct {
	SinglePtr           *InnerData
	ArrayOfPtrs         [2]*InnerData
	MapOfPtrs           map[string]*InnerData
	RegisteredInterface any
	SharedRef           *InnerData
}

// Helper to generate the exact same data structure for both storing and verification
func getReferenceObject() *ComplexData {
	sharedObj := &InnerData{ID: 101, Name: "SharedObject"}
	knownType := coin.NewInfo("RegisteredCoin", coin.IntegerVec, 16, 16, 1)

	return &ComplexData{
		SinglePtr: sharedObj,
		ArrayOfPtrs: [2]*InnerData{
			{ID: 1, Name: "Array1"},
			{ID: 2, Name: "Array2"},
		},
		MapOfPtrs: map[string]*InnerData{
			"key1": {ID: 10, Name: "MapVal1"},
			"key2": {ID: 20, Name: "MapVal2"},
		},
		RegisteredInterface: &knownType,
		SharedRef:           sharedObj, // Cycle/Shared Reference
	}
}

// --- Test 1: StoreToDisk ---
// This test ONLY persists the file to the current directory.
func TestA_StoreToDisk(t *testing.T) {
	// 1. Get the data
	complexObj := getReferenceObject()

	// 2. Store to Disk (Uncompressed)
	// This creates "integration_test_data.bin" in your current folder
	err := serde.StoreToDisk(TestFileName, complexObj, false)
	require.NoError(t, err, "StoreToDisk failed")

	// 3. Verify it physically exists
	info, err := os.Stat(TestFileName)
	require.NoError(t, err)
	require.Greater(t, info.Size(), int64(0), "File was created but is empty")

	t.Logf("Setup Complete: Persisted %s (Size: %d bytes)", TestFileName, info.Size())
}

// --- Test 2: LoadFromDisk ---
// This test ONLY loads the file created by the previous test.
func TestB_LoadFromDisk(t *testing.T) {
	// 1. Check if the file exists (Dependency check)
	_, err := os.Stat(TestFileName)
	require.NoError(t, err, "Test artifact missing! Did TestA_StoreToDisk run?")

	// 2. Load the artifact
	var loaded ComplexData
	closer, err := serde.LoadFromDisk(TestFileName, &loaded, false)
	require.NoError(t, err, "LoadFromDisk failed")
	defer closer.Close() // Release Mmap

	// 3. Cleanup (Optional: remove file after successful load)
	// defer os.Remove(TestFileName)

	// 4. Comparison
	original := getReferenceObject()

	// A. Deep Equality Check
	// We use the library's own DeepCmp to verify the entire tree structure matches
	isEqual := serde.DeepCmp(original, &loaded, false)
	require.True(t, isEqual, "Loaded object structure differs from original")

	// B. Pointer Deduplication Check
	// Verify that the 'SharedRef' pointer points to the exact same memory address
	// as 'SinglePtr', proving the DAG was preserved.
	if loaded.SinglePtr != loaded.SharedRef {
		t.Fatalf("Pointer Identity Lost! SinglePtr(%p) != SharedRef(%p)", loaded.SinglePtr, loaded.SharedRef)
	}

	// C. Interface Type Safety Check
	// Verify the interface was revived as the correct concrete type (*coin.Info)
	restoredCoin, ok := loaded.RegisteredInterface.(*coin.Info)
	require.True(t, ok, "Interface did not deserialize to *coin.Info")
	require.Equal(t, coin.Name("RegisteredCoin"), restoredCoin.Name, "Interface data corrupted")

	t.Log("Success: Loaded and verified data from disk.")
}

const SliceTestFile = "slice_ptr_fail.bin"

type StructWithSlice struct {
	Name string
	// THIS IS THE PROBLEM FIELD
	SliceOfPtrs []*InnerData
}

// --- Test 3: Store Slice of Ptrs (The Trap) ---
func TestC_SliceOfPtrs_Store(t *testing.T) {
	// Create data
	a := &InnerData{ID: 1, Name: "A"}
	b := &InnerData{ID: 2, Name: "B"}

	obj := &StructWithSlice{
		Name:        "Root",
		SliceOfPtrs: []*InnerData{a, b},
	}

	// Store it
	// This will "work" silently, but it's writing garbage (raw pointers) to the file.
	err := serde.StoreToDisk(SliceTestFile, obj, false)
	require.NoError(t, err)

	info, err := os.Stat(SliceTestFile)
	require.NoError(t, err)
	t.Logf("Stored slice of pointers to %s (Size: %d bytes)", SliceTestFile, info.Size())
}

// --- Test 4: Load Slice of Ptrs (The Crash/Fail) ---
func TestD_SliceOfPtrs_Load(t *testing.T) {
	// 1. Check file exists
	_, err := os.Stat(SliceTestFile)
	require.NoError(t, err, "Test artifact missing")
	defer os.Remove(SliceTestFile)

	// 2. Load
	var loaded StructWithSlice
	closer, err := serde.LoadFromDisk(SliceTestFile, &loaded, false)
	require.NoError(t, err)
	defer closer.Close()

	t.Logf("Loaded object. Slice len: %d", len(loaded.SliceOfPtrs))

	if len(loaded.SliceOfPtrs) > 0 {
		// DANGER ZONE: This pointer is likely 0xc000... from the previous run
		ptr := loaded.SliceOfPtrs[0]

		// Attempt to read data. This is where it dies or returns garbage.
		// If the serializer worked correctly, this would be "A".
		// If it failed (raw ptr), this address is invalid.
		t.Logf("Attempting to read ptr %p...", ptr)
		t.Logf("Value: %v", ptr.Name)
	}
}
