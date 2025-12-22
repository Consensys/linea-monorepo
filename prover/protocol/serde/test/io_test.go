package serde_test

import (
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/serde"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/stretchr/testify/require"
)

// Shared structure for test cases
type serdeTestCase struct {
	Name string
	V    any
}

// getSerdeTestCases generates fresh instances of the test objects.
// We use a function so 'Store' and 'Load' get their own distinct pointers/memory.
func getSerdeTestCases() []serdeTestCase {
	// Ensure types are registered (idempotent)
	serde.RegisterImplementation(string(""))
	serde.RegisterImplementation(ifaces.ColID(""))
	serde.RegisterImplementation(column.Natural{})
	serde.RegisterImplementation(column.Shifted{})
	serde.RegisterImplementation(verifiercol.ConstCol{})
	serde.RegisterImplementation(verifiercol.FromYs{})
	serde.RegisterImplementation(verifiercol.FromAccessors{})
	serde.RegisterImplementation(accessors.FromPublicColumn{})
	serde.RegisterImplementation(accessors.FromConstAccessor{})
	serde.RegisterImplementation(query.UnivariateEval{})

	return []serdeTestCase{
		{
			Name: "random",
			V:    "someRandomString",
		},
		{
			Name: "random-column-id-ptr",
			V: func() any {
				var s = ifaces.ColID("someIndirectedString")
				return &s
			}(),
		},
		{
			Name: "some-string-ptr",
			V: func() any {
				var s any = string("someStringUnderIface")
				return &s
			}(),
		},
		{
			Name: "some-col-id-ptr",
			V: func() any {
				var id = ifaces.ColID("newTypeUnderIface")
				var s any = &id
				return &s
			}(),
		},
		{
			Name: "query-id",
			V:    ifaces.QueryID("QueryID"),
		},
		{
			Name: "natural-column",
			V: func() any {
				comp := wizard.NewCompiledIOP()
				return comp.InsertColumn(0, "myNaturalColumn", 16, column.Committed)
			}(),
		},
		{
			Name: "shifted-column",
			V: func() any {
				comp := wizard.NewCompiledIOP()
				nat := comp.InsertColumn(0, "myNaturalColumn", 16, column.Committed)
				return column.Shift(nat, 2)
			}(),
		},
		{
			Name: "concat-tiny-columns",
			V: func() any {
				comp := wizard.NewCompiledIOP()
				col := verifiercol.NewConcatTinyColumns(
					comp,
					8,
					field.Element{},
					comp.InsertColumn(0, "a", 1, column.Proof),
					comp.InsertColumn(0, "b", 1, column.Proof),
					comp.InsertColumn(0, "c", 1, column.Proof),
				)
				return &col
			}(),
		},
		{
			Name: "from-public-column",
			V: func() any {
				comp := wizard.NewCompiledIOP()
				nat := comp.InsertColumn(0, "myNaturalColumn", 16, column.Proof)
				return accessors.NewFromPublicColumn(nat, 2)
			}(),
		},
		{
			Name: "from-const-accessor",
			V: func() any {
				comp := wizard.NewCompiledIOP()
				c := comp.InsertCoin(0, "myCoin", coin.Field)
				return accessors.NewFromCoin(c)
			}(),
		},
		{
			Name: "univariate-eval",
			V: func() any {
				comp := wizard.NewCompiledIOP()
				a := comp.InsertColumn(0, "a", 16, column.Committed)
				b := comp.InsertColumn(0, "b", 16, column.Committed)
				q := comp.InsertUnivariate(0, "q", []ifaces.Column{a, b})
				return q
			}(),
		},
		{
			Name: "from-ys",
			V: func() any {
				comp := wizard.NewCompiledIOP()
				a := comp.InsertColumn(0, "a", 16, column.Committed)
				b := comp.InsertColumn(0, "b", 16, column.Committed)
				q := comp.InsertUnivariate(0, "q", []ifaces.Column{a, b})
				return verifiercol.NewFromYs(comp, q, []ifaces.ColID{a.GetColID(), b.GetColID()})
			}(),
		},
		{
			Name: "integer-vec-coin",
			V: func() any {
				comp := wizard.NewCompiledIOP()
				c := comp.InsertCoin(0, "myCoin", coin.IntegerVec, 16, 16)
				return c
			}(),
		},
		{
			Name: "from-integer-vec-coin",
			V: func() any {
				comp := wizard.NewCompiledIOP()
				c := comp.InsertCoin(0, "myCoin", coin.IntegerVec, 16, 16)
				return verifiercol.NewFromIntVecCoin(comp, c)
			}(),
		},
		{
			Name: "verifier-col-from-accessors",
			V: func() any {
				comp := wizard.NewCompiledIOP()
				a := comp.InsertColumn(0, "a", 16, column.Proof)
				b := comp.InsertColumn(0, "b", 16, column.Proof)
				return verifiercol.NewFromAccessors(
					[]ifaces.Accessor{
						accessors.NewFromPublicColumn(a, 2),
						accessors.NewFromPublicColumn(b, 2),
					},
					field.Element{}, 16)
			}(),
		},
		{
			Name: "const-col",
			V:    verifiercol.NewConstantCol(field.Element{}, 16, ""),
		},
		{
			Name: "new-coin",
			V:    coin.NewInfo("foo", coin.IntegerVec, 16, 16, 1),
		},
		{
			Name: "bigint-zero",
			V:    big.NewInt(0),
		},
		{
			Name: "bigint-one",
			V:    big.NewInt(1),
		},
		{
			Name: "bigint-minus-one",
			V:    big.NewInt(-1),
		},
		{
			Name: "bigint-max-uint256",
			V: func() any {
				v, ok := new(big.Int).SetString("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 0)
				if !ok {
					panic("bigint does not work")
				}
				return v
			}(),
		},
		{
			Name: "field-zero",
			V:    field.NewElement(0),
		},
		{
			Name: "field-one",
			V:    field.NewElement(1),
		},
		{
			Name: "field-2**248-1",
			V: func() any {
				v, err := new(field.Element).SetString("0x00ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
				if err != nil {
					utils.Panic("field does not work: %v", err)
				}
				return v
			}(),
		},
		{
			Name: "vector-1234",
			V:    vector.ForTest(0, 1, 2, 3, 4, 5, 5, 6, 7),
		},
		{
			Name: "symbolic-add-with-shifted",
			V: func() any {
				comp := wizard.NewCompiledIOP()
				a := comp.InsertColumn(0, "a", 16, column.Committed)
				aNext := column.Shift(a, 2)
				return symbolic.Add(a, aNext)
			}(),
		},
		{
			Name: "symbolic-add",
			V: func() any {
				comp := wizard.NewCompiledIOP()
				a := comp.InsertColumn(0, "a", 16, column.Committed)
				b := comp.InsertColumn(0, "b", 16, column.Committed)
				return symbolic.Add(a, b)
			}(),
		},
		{
			Name: "symbolic-mul",
			V: func() any {
				comp := wizard.NewCompiledIOP()
				a := comp.InsertColumn(0, "a", 16, column.Committed)
				b := comp.InsertColumn(0, "b", 16, column.Committed)
				return symbolic.Mul(a, b)
			}(),
		},
		{
			Name: "symbolic-new-variable",
			V: func() any {
				comp := wizard.NewCompiledIOP()
				a := comp.InsertColumn(0, "a", 16, column.Committed)
				return symbolic.NewVariable(a)
			}(),
		},
		{
			Name: "symbolic-constant-0",
			V:    symbolic.NewConstant(0),
		},
		{
			Name: "symbolic-constant-1",
			V:    symbolic.NewConstant(1),
		},
		{
			Name: "symbolic-constant-minus-1",
			V:    symbolic.NewConstant(-1),
		},
		{
			Name: "symbolic-intricate-expression",
			V: func() any {
				comp := wizard.NewCompiledIOP()
				a := comp.InsertColumn(0, "a", 16, column.Committed)
				b := comp.InsertColumn(0, "b", 16, column.Committed)
				c := comp.InsertColumn(0, "c", 16, column.Committed)
				d := comp.InsertColumn(0, "d", 16, column.Committed)
				return symbolic.Add(symbolic.Mul(symbolic.Add(a, b), symbolic.Add(c, d)), symbolic.NewConstant(1))
			}(),
		},
		{
			Name: "mimc-query",
			V: func() any {
				comp := wizard.NewCompiledIOP()
				a := comp.InsertColumn(0, "a", 16, column.Committed)
				b := comp.InsertColumn(0, "b", 16, column.Committed)
				c := comp.InsertColumn(0, "c", 16, column.Committed)
				return comp.InsertMiMC(0, "mimc", a, b, c, nil)
			}(),
		},
		{
			Name: "nil-expression",
			V: func() any {
				return struct {
					E *symbolic.Expression
				}{}
			}(),
		},
		{
			Name: "map-with-column-as-key",
			V: func() any {
				comp := wizard.NewCompiledIOP()
				a := comp.InsertColumn(0, "a", 16, column.Committed).(column.Natural)
				b := comp.InsertColumn(0, "b", 16, column.Committed).(column.Natural)
				return map[column.Natural]string{a: "a", b: "b"}
			}(),
		},
		{
			Name: "frontend-variables",
			V:    frontend.Variable(0),
		},
		{
			Name: "frontend-variables",
			V:    frontend.Variable(12),
		},
		{
			Name: "frontend-variables",
			V:    frontend.Variable(-10),
		},
		{
			Name: "two-wiop-in-a-struct",
			V: func() any {
				res := struct {
					A, B *wizard.CompiledIOP
				}{
					A: wizard.NewCompiledIOP(),
					B: wizard.NewCompiledIOP(),
				}
				res.A.InsertColumn(0, "a", 16, column.Committed)
				res.B.InsertColumn(0, "b", 16, column.Committed)
				return res
			}(),
		},
		{
			Name: "recursion",
			V: func() any {
				wiop := wizard.NewCompiledIOP()
				a := wiop.InsertCommit(0, "a", 1<<10)
				wiop.InsertUnivariate(0, "u", []ifaces.Column{a})

				wizard.ContinueCompilation(wiop,
					vortex.Compile(
						2,
						vortex.WithOptionalSISHashingThreshold(0),
						vortex.ForceNumOpenedColumns(2),
						vortex.PremarkAsSelfRecursed(),
					),
				)

				rec := wizard.NewCompiledIOP()
				recursion.DefineRecursionOf(rec, wiop, recursion.Parameters{
					MaxNumProof: 1,
					WithoutGkr:  true,
					Name:        "recursion",
				})

				return rec
			}(),
		},
		{
			Name: "mutex",
			V:    []*sync.Mutex{{}, {}},
		},
		{
			Name: "nil-horner-query",
			V:    (*query.Horner)(nil),
		},
		{
			Name: "nil-horner-query-in-a-struct",
			V: struct {
				G *query.Horner
			}{},
		},
	}
}

// PHASE 1: STORE
// Iterates through all test cases and writes them to the "files/" directory.
func TestSerdeValue_Store(t *testing.T) {
	testDir := "files"
	// Cleanup and recreate directory
	_ = os.RemoveAll(testDir)
	require.NoError(t, os.MkdirAll(testDir, 0755))

	testCases := getSerdeTestCases()

	for i, tc := range testCases {

		if i != 40 {
			continue
		}

		t.Run(fmt.Sprintf("Store/%d_%s", i, tc.Name), func(t *testing.T) {
			path := filepath.Join(testDir, fmt.Sprintf("%d_%s.bin", i, tc.Name))

			// Write to disk using standard API (no compression to test raw serde)
			err := serde.StoreToDisk(path, tc.V, false)
			require.NoError(t, err, "StoreToDisk failed")
		})
	}
}

// PHASE 2: LOAD
// Reads the files created by Store, deserializes them, verifies correctness, and deletes them.
func TestSerdeValue_Load(t *testing.T) {
	testDir := "files"
	testCases := getSerdeTestCases()

	for i, tc := range testCases {

		if i != 40 {
			continue
		}

		t.Run(fmt.Sprintf("Load/%d_%s", i, tc.Name), func(t *testing.T) {
			path := filepath.Join(testDir, fmt.Sprintf("%d_%s.bin", i, tc.Name))

			// 1. Prepare destination pointer
			// LoadFromDisk requires a pointer to the type being loaded.
			// Since tc.V is 'any', we use reflection to create a new pointer of that type.
			destPtr := reflect.New(reflect.TypeOf(tc.V))

			// 2. Load
			closer, err := serde.LoadFromDisk(path, destPtr.Interface(), false)
			require.NoError(t, err, "LoadFromDisk failed")

			// 3. Compare (DeepCmp)
			// We check destPtr.Elem() because destPtr is *T, but tc.V is T.
			isMatch := serde.DeepCmp(tc.V, destPtr.Elem().Interface(), false)
			if !isMatch {
				t.Errorf("DeepCmp failed: loaded object does not match original")
			}

			// 4. Cleanup
			// CRITICAL: Close the mmap before attempting to delete the file.
			// On Windows, deleting an open file is forbidden. On Linux, it works but leaks resources.
			require.NoError(t, closer.Close(), "Failed to close mmap")
			require.NoError(t, os.Remove(path), "Failed to delete test file")
		})
	}

	// Final cleanup of directory
	_ = os.RemoveAll(testDir)
}

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

// Helper struct for the test
type StructWith2DSlice struct {
	Name string
	Grid [][]*InnerData
}

const MatrixTestFile = "test_matrix.bin"

// --- Test 5a: 2D Slice of Ptrs (Store) ---
func TestE_2DSliceOfPtrs_Store(t *testing.T) {
	// 1. Setup Data
	// We reuse pointers to ensure referential integrity is preserved.
	// (i.e., 'a' appears in multiple places)
	a := &InnerData{ID: 10, Name: "Origin"}
	b := &InnerData{ID: 20, Name: "Right"}
	c := &InnerData{ID: 30, Name: "Down"}

	// Construct a jagged grid (2D slice)
	// Row 0: [Origin, Right]
	// Row 1: [Down, Origin, nil]
	matrix := [][]*InnerData{
		{a, b},
		{c, a, nil},
	}

	obj := &StructWith2DSlice{
		Name: "MatrixRoot",
		Grid: matrix,
	}

	// 2. Store to Disk
	err := serde.StoreToDisk(MatrixTestFile, obj, false)
	require.NoError(t, err)

	info, err := os.Stat(MatrixTestFile)
	require.NoError(t, err)
	t.Logf("Stored 2D matrix to %s (Size: %d bytes)", MatrixTestFile, info.Size())
}

// --- Test 5b: 2D Slice of Ptrs (Load & Verify) ---
func TestE_2DSliceOfPtrs_Load(t *testing.T) {
	// 1. Check artifact exists
	_, err := os.Stat(MatrixTestFile)
	require.NoError(t, err, "Test artifact missing. Run Store test first.")
	defer os.Remove(MatrixTestFile)

	// 2. Load from Disk
	var loaded StructWith2DSlice
	closer, err := serde.LoadFromDisk(MatrixTestFile, &loaded, false)
	require.NoError(t, err)
	defer closer.Close()

	// 3. Verification
	t.Logf("Loaded matrix. Rows: %d", len(loaded.Grid))

	require.Equal(t, "MatrixRoot", loaded.Name)
	require.Len(t, loaded.Grid, 2)

	// Row 0 checks
	require.Len(t, loaded.Grid[0], 2)
	require.Equal(t, "Origin", loaded.Grid[0][0].Name)
	require.Equal(t, "Right", loaded.Grid[0][1].Name)

	// Row 1 checks
	require.Len(t, loaded.Grid[1], 3)
	require.Equal(t, "Down", loaded.Grid[1][0].Name)

	// 4. Verify Referential Integrity (Critical)
	// Grid[0][0] and Grid[1][1] should be the EXACT same pointer address.
	// If the serializer naively created new copies, this would fail.
	ptr1 := loaded.Grid[0][0]
	ptr2 := loaded.Grid[1][1]

	require.True(t, ptr1 == ptr2,
		"Referential integrity failed: Expected pointers to 'Origin' to be identical, got %p and %p", ptr1, ptr2)

	// 5. Verify Nil Safety
	require.Nil(t, loaded.Grid[1][2], "Expected nil element to remain nil")

	t.Log("2D Slice (Matrix) verification passed.")
}
