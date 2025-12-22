package serde_test

import (
	"crypto/rand"
	"fmt"
	"math/big"
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
	"github.com/sirupsen/logrus"

	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSerdeValue(t *testing.T) {

	testCases := []struct {
		V    any
		Name string
	}{
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
				// It's important to not provide an untyped string under
				// the interface because the type cannot be serialized.
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

				logrus.Printf("recursion=%+v\n", rec)
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
				G *query.Horner // G is left as a nil by default
			}{},
		},
	}

	for i := 0; i < len(testCases); i++ {

		if i != 40 {
			continue
		}

		t.Run(fmt.Sprintf("test-case-%v/%v", i, testCases[i].Name), func(t *testing.T) {

			msg, err := serde.Serialize(testCases[i].V)
			require.NoError(t, err)

			t.Logf("testcase=%v, msg=%v\n", i, string(msg))

			unmarshaled := reflect.New(reflect.TypeOf(testCases[i].V)).Interface()
			err = serde.Deserialize(msg, unmarshaled)
			require.NoError(t, err)

			unmarshalledDereferenced := reflect.ValueOf(unmarshaled).Elem().Interface()
			if !serde.DeepCmp(testCases[i].V, unmarshalledDereferenced, false) {
				t.Errorf("Mismatch in exported fields after full serde value")
			}

		})
	}
}

func TestSerdeFE(t *testing.T) {
	newFieldElement := func(n int64) field.Element {
		var f field.Element
		f.SetBigInt(big.NewInt(n))
		return f
	}

	// Test cases for single field.Element
	singleTests := []struct {
		name string
		val  field.Element
	}{
		{"Zero", field.Element{0, 0, 0, 0}},
		{"Small", newFieldElement(42)},
		{"Large", newFieldElement(1 << 60)},
		{
			"ModulusMinusOne",
			func() field.Element {
				var f field.Element
				f.SetOne()
				f.Neg(&f) // modulus - 1
				return f
			}(),
		},
		{
			"Random",
			func() field.Element {
				var f field.Element
				b := make([]byte, 32)
				_, _ = rand.Read(b)
				f.SetBytes(b)
				return f
			}(),
		},
	}

	for _, tt := range singleTests {
		t.Run("Single_"+tt.name, func(t *testing.T) {
			bytes, err := serde.Serialize(tt.val)
			if err != nil {
				t.Fatalf("serialize error: %v", err)
			}
			var result field.Element
			if err := serde.Deserialize(bytes, &result); err != nil {
				t.Fatalf("deserialize error: %v", err)
			}
			if result != tt.val {
				t.Errorf("mismatch: got %v, want %v", result, tt.val)
			}
		})
	}

	// Test cases for []field.Element
	arrayTests := []struct {
		name string
		val  []field.Element
	}{
		{"Empty", []field.Element{}},
		{"Single", []field.Element{newFieldElement(42)}},
		{"Multiple", []field.Element{
			newFieldElement(0),
			newFieldElement(42),
			newFieldElement(1 << 60),
			func() field.Element {
				var f field.Element
				f.SetOne()
				f.Neg(&f)
				return f
			}(),
		}},
		{
			"Repeated",
			[]field.Element{
				newFieldElement(123),
				newFieldElement(123),
				newFieldElement(123),
			},
		},
		{
			"BoundaryNearModulus",
			[]field.Element{
				func() field.Element {
					var f field.Element
					f.SetOne()
					f.Neg(&f)
					return f
				}(),
				newFieldElement(1),
			},
		},
		{
			"LargeArray",
			func() []field.Element {
				arr := make([]field.Element, 1000)
				for i := 0; i < len(arr); i++ {
					arr[i] = newFieldElement(int64(i * i))
				}
				return arr
			}(),
		},
		{
			"RandomElements",
			func() []field.Element {
				arr := make([]field.Element, 100)
				for i := range arr {
					var f field.Element
					b := make([]byte, 32)
					_, _ = rand.Read(b)
					f.SetBytes(b)
					arr[i] = f
				}
				return arr
			}(),
		},
	}

	for _, tt := range arrayTests {
		t.Run("Array_"+tt.name, func(t *testing.T) {
			bytes, err := serde.Serialize(tt.val)
			if err != nil {
				t.Fatalf("serialize error: %v", err)
			}
			var result []field.Element
			if err := serde.Deserialize(bytes, &result); err != nil {
				t.Fatalf("deserialize error: %v", err)
			}
			if len(result) != len(tt.val) {
				t.Fatalf("length mismatch: got %d, want %d", len(result), len(tt.val))
			}
			for i := range tt.val {
				if result[i] != tt.val[i] {
					t.Errorf("element %d mismatch: got %v, want %v", i, result[i], tt.val[i])
				}
			}
		})
	}
}

func TestFieldElementLimbMismatch(t *testing.T) {
	// Construct a field.Element with known limb values
	original := [4]uint64{
		4432961018360255618, // limb 0
		1234567890123456789, // limb 1
		9876543210987654321, // limb 2
		1111111111111111111, // limb 3
	}

	// Serialize it
	bytes, err := serde.Serialize(original)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Deserialize it
	var result [4]uint64
	if err := serde.Deserialize(bytes, &result); err != nil {
		t.Fatalf("failed to deserialize: %v", err)
	}

	// Compare limbs manually and print diff
	for i := 0; i < 4; i++ {
		if result[i] != original[i] {
			t.Errorf("field.Element limb mismatch at [%d]: got %d, want %d", i, result[i], original[i])
		}
	}
}

func TestStoreAndColumnIntegrity(t *testing.T) {
	// 1. Setup Original State
	originalStore := column.NewStore()

	// Add a column - this creates the internal maps and slices
	// We cast to column.Natural to access the unexported 'store' pointer later
	natValue := originalStore.AddToRound(0, "column_a", 4, column.Committed).(column.Natural)

	// Create a shared Coin instance
	sharedCoin := coin.NewInfo("beta", coin.Field, 1)

	// Build a container where multiple pointers point to the EXACT same RAM addresses
	type Container struct {
		Store *column.Store
		Col1  *column.Natural
		Col2  *column.Natural
		Coin1 *coin.Info
		Coin2 *coin.Info
	}

	input := &Container{
		Store: originalStore,
		Col1:  &natValue,
		Col2:  &natValue, // Point to the same handle
		Coin1: &sharedCoin,
		Coin2: &sharedCoin, // Point to the same coin
	}

	// 2. Serialize
	data, err := serde.Serialize(input)
	require.NoError(t, err, "Serialization failed")

	// 3. Deserialize
	var output Container
	err = serde.Deserialize(data, &output)
	require.NoError(t, err, "Deserialization failed")

	// 4. THE VALIDATION

	// A. Logical Field Validation
	// Proves data was reconstructed correctly despite the Pack/Unpack cycle.
	require.True(t, serde.DeepCmp(input, &output, false), "Logical data mismatch")

	// B. Store Deduplication (Handle -> Root)
	// Both handles must point to the same Store instance found at the root.
	ptrRootStore := reflect.ValueOf(output.Store).Pointer()
	ptrCol1InternalStore := reflect.ValueOf(*output.Col1).FieldByName("store").Pointer()
	ptrCol2InternalStore := reflect.ValueOf(*output.Col2).FieldByName("store").Pointer()

	assert.Equal(t, ptrRootStore, ptrCol1InternalStore, "Col1 handle should share the root Store instance")
	assert.Equal(t, ptrCol1InternalStore, ptrCol2InternalStore, "Col1 and Col2 handles must share the same Store instance")

	// C. Handle Deduplication (Handle -> Handle)
	// This proves the library didn't create two Natural handles for the same offset.
	assert.Equal(t, reflect.ValueOf(output.Col1).Pointer(), reflect.ValueOf(output.Col2).Pointer(),
		"Natural handles themselves were not deduplicated in the ptrMap")

	// D. Coin Deduplication (Structural Type)
	// Proves that standard pointers (no custom handlers) are also correctly deduplicated.
	assert.Equal(t, reflect.ValueOf(output.Coin1).Pointer(), reflect.ValueOf(output.Coin2).Pointer(),
		"Coin pointers should point to the exact same heap address")

	// E. Functional Verification
	// Final check that the private internal maps of the Store are actually working.
	assert.True(t, output.Store.Exists("column_a"), "Private indicesByNames map was not restored")
	assert.Equal(t, 4, output.Store.GetSize("column_a"), "Private byRounds slice was not restored")
}

func TestStoreMutationIntegrity(t *testing.T) {
	// This test ensures that because pointers are shared, a mutation in the Store
	// is visible through the Column handle, just like in real Go RAM.

	store := column.NewStore()
	col := store.AddToRound(0, "col", 8, column.Committed).(column.Natural)

	type Root struct {
		Store *column.Store
		Col   *column.Natural
	}

	data, _ := serde.Serialize(&Root{store, &col})

	var decoded Root
	serde.Deserialize(data, &decoded)

	// Mutate the state in the Store
	decoded.Store.MarkAsIgnored("col")

	// Check if the handle (which points to the same Store) sees the change
	assert.True(t, decoded.Col.Status() == column.Ignored,
		"Mutation in shared Store should be visible through the Natural handle")
}
