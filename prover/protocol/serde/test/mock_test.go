package serde_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/serde" // Adjust import path if needed
	util "github.com/consensys/linea-monorepo/prover/protocol/serde/examples"
	"github.com/stretchr/testify/require"
)

// --- PHASE 1: STORE ---
func TestDebug_Store(t *testing.T) {
	t.Skipf("the test is a development/debug/integration test. It is not needed for CI")
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
	t.Skipf("the test is a development/debug/integration test. It is not needed for CI")
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
	t.Skipf("the test is a development/debug/integration test. It is not needed for CI")
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
	t.Skipf("the test is a development/debug/integration test. It is not needed for CI")
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
	wrapper, ok := valInMap.(util.WrapperC)
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

func makeFieldMatrix() [][]field.Element {
	m := make([][]field.Element, 2)
	for i := range m {
		m[i] = make([]field.Element, 2)
		for j := range m[i] {
			var e field.Element
			e.SetUint64(uint64(10*i + j + 1)) // 1,2,11,12
			m[i][j] = e
		}
	}
	return m
}

func toUint64Limbs(mat [][]field.Element) [][]uint64 {
	out := make([][]uint64, len(mat))
	for i := range mat {
		out[i] = make([]uint64, len(mat[i])*4)
		for j, e := range mat[i] {
			var limbs [4]uint64
			limbs = [4]uint64(e)
			copy(out[i][j*4:(j+1)*4], limbs[:])
		}
	}
	return out
}

func cmatrixPath(t *testing.T) string {
	t.Helper()
	// Use a stable location under the repo, similar to files/iop3.bin.
	return filepath.Join("testdata", "cmatrix.bin")
}

// 1) Writer test: produces the asset once.
func TestStore_CommittedMatrix(t *testing.T) {
	t.Skipf("the test is a development/debug/integration test. It is not needed for CI")
	path := cmatrixPath(t)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir testdata failed: %v", err)
	}

	fmat := makeFieldMatrix()
	limbs := toUint64Limbs(fmat)

	orig := &util.Wrapper{
		Committed: &util.FakeCommittedMatrix{Limbs: limbs},
	}

	if err := serde.StoreToDisk(path, orig, false); err != nil {
		t.Fatalf("StoreToDisk failed: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stored file not found: %v", err)
	}
}

//  2. Reader test: assumes cmatrix.bin already exists (written by old code)
//     and checks whether new LoadFromDisk reproduces the limbs exactly.
func TestLoad_CommittedMatrix(t *testing.T) {
	t.Skipf("the test is a development/debug/integration test. It is not needed for CI")
	path := cmatrixPath(t)

	// Rebuild expected logical matrix deterministically.
	fmat := makeFieldMatrix()
	expected := toUint64Limbs(fmat)

	var loaded util.Wrapper
	closer, err := serde.LoadFromDisk(path, &loaded, false)
	if err != nil {
		t.Fatalf("LoadFromDisk failed: %v", err)
	}
	defer closer.Close()

	got, ok := loaded.Committed.(*util.FakeCommittedMatrix)
	if !ok {
		t.Fatalf("loaded.Committed has type %T, want *FakeCommittedMatrix", loaded.Committed)
	}

	if len(got.Limbs) != len(expected) {
		t.Fatalf("row count mismatch: exp=%d got=%d", len(expected), len(got.Limbs))
	}
	for i := range expected {
		if len(got.Limbs[i]) != len(expected[i]) {
			t.Fatalf("col count mismatch at row %d: exp=%d got=%d",
				i, len(expected[i]), len(got.Limbs[i]))
		}
		for j := range expected[i] {
			if expected[i][j] != got.Limbs[i][j] {
				t.Fatalf("mismatch at [%d][%d]: exp=%d got=%d",
					i, j, expected[i][j], got.Limbs[i][j])
			}
		}
	}
}
