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
