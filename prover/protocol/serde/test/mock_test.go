package serde_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/serde" // Adjust import path if needed
	"github.com/consensys/linea-monorepo/prover/protocol/serde/util"
	"github.com/stretchr/testify/require"
)

// --- PHASE 1: STORE ---
func TestDebug_Store(t *testing.T) {
	testDir := "repro_files"
	_ = os.RemoveAll(testDir)
	require.NoError(t, os.MkdirAll(testDir, 0755))

	path := filepath.Join(testDir, "aliasing_bug.bin")
	obj := util.GetDebugObject()

	t.Logf("Storing Object Graph (DirectPtr == ExtraData['aliased_ref']) to %s", path)

	// This triggers the bug if the Writer doesn't check the ptrMap
	// when linearizing the interface{} value inside the map.
	err := serde.StoreToDisk(path, obj, false)
	require.NoError(t, err, "StoreToDisk failed")
}

// --- PHASE 2: LOAD ---
func TestDebug_Load(t *testing.T) {
	testDir := "repro_files"
	path := filepath.Join(testDir, "aliasing_bug.bin")
	defer os.RemoveAll(testDir)

	// 1. Expected Graph
	expected := util.GetDebugObject()

	// 2. Load
	t.Log("Loading Object Graph...")

	// Create a pointer to the type we want to load
	var loaded *util.RootContainer

	// LoadFromDisk usually takes (path, *TargetType, compression)
	// Adjust signature based on your actual serde.LoadFromDisk definition
	closer, err := serde.LoadFromDisk(path, &loaded, false)
	require.NoError(t, err, "LoadFromDisk failed")
	if closer != nil {
		defer closer.Close()
	}

	// 3. Verify Basic Integrity
	require.NotNil(t, loaded, "Loaded object is nil")
	require.NotNil(t, loaded.DirectPtr, "DirectPtr is nil")
	require.NotNil(t, loaded.ExtraData, "ExtraData is nil")

	// 4. CRITICAL CHECK: POINTER IDENTITY
	// We extract the pointer from the map interface
	valInMap, ok := loaded.ExtraData["aliased_ref"]
	require.True(t, ok, "Key 'aliased_ref' missing from map")

	// Assert that the map value is indeed a pointer to SharedLeaf
	leafFromMap, ok := valInMap.(*util.SharedLeaf)
	require.True(t, ok, "Map value has wrong type %T", valInMap)

	t.Logf("Direct Ptr Address: %p", loaded.DirectPtr)
	t.Logf("Map Ref Ptr Address: %p", leafFromMap)

	if loaded.DirectPtr != leafFromMap {
		t.Errorf("FAIL: POINTER IDENTITY LOST.\n"+
			"DirectPtr (%p) != ExtraData['aliased_ref'] (%p).\n"+
			"The serializer created a COPY instead of a REFERENCE.",
			loaded.DirectPtr, leafFromMap)
	} else {
		t.Log("SUCCESS: Pointer Identity Preserved.")
	}

	// 5. Run DeepCmp
	// Note: DeepCmp might actually PASS if it compares by value content,
	// but the test above proves the topology is broken.
	// If DeepCmp enforces topology (pointer equality), it will fail here.
	t.Log("Running DeepCmp...")
	match := serde.DeepCmp(expected, loaded, false)
	require.True(t, match, "DeepCmp returned false")
}

// --- PHASE 1: STORE ---
func TestComplex_Store(t *testing.T) {
	testDir := "complex_repro_files"
	_ = os.RemoveAll(testDir)
	require.NoError(t, os.MkdirAll(testDir, 0755))

	path := filepath.Join(testDir, "complex_aliasing.bin")
	obj := util.GetComplexReproObject()

	t.Logf("Storing Complex Object (Cycle + Struct Wrapper) to %s", path)
	err := serde.StoreToDisk(path, obj, false)
	require.NoError(t, err, "StoreToDisk failed")
}

// --- PHASE 2: LOAD ---
func TestComplex_Load(t *testing.T) {
	testDir := "complex_repro_files"
	path := filepath.Join(testDir, "complex_aliasing.bin")
	// defer os.RemoveAll(testDir)

	t.Log("Loading Complex Object...")
	var loaded *util.RootComplex
	closer, err := serde.LoadFromDisk(path, &loaded, false)
	require.NoError(t, err, "LoadFromDisk failed")
	if closer != nil {
		defer closer.Close()
	}

	// 1. Verify Topology
	parent := loaded.MainParent
	require.NotNil(t, parent)
	require.Len(t, parent.Children, 1)
	childViaParent := parent.Children[0]

	// 2. Extract from Map
	valInMap, ok := loaded.ExtraData["wrapper_key"]
	require.True(t, ok)

	// Note: serialization of interface{} -> concrete type
	// If the writer wrote "Wrapper", the reader reads "Wrapper".
	wrapper, ok := valInMap.(util.Wrapper)
	require.True(t, ok, "Map value should be Wrapper struct")

	childViaMap := wrapper.Ref

	// 3. CRITICAL CHECK 1: Aliasing of Child
	t.Logf("Child via Parent: %p", childViaParent)
	t.Logf("Child via Map:    %p", childViaMap)

	if childViaParent != childViaMap {
		t.Errorf("FAIL: Child Aliasing Lost! (ViaParent != ViaMap)")
	}

	// 4. CRITICAL CHECK 2: Cycle Integrity (Parent Identity)
	parentViaChild := childViaMap.Parent
	t.Logf("Parent via Root:  %p", parent)
	t.Logf("Parent via Child: %p", parentViaChild)

	if parent != parentViaChild {
		t.Errorf("FAIL: Cycle Broken! Child.Parent != Root.Parent")
	}

	if !t.Failed() {
		t.Log("SUCCESS: Complex Aliasing & Cycles Preserved.")
	}
}
