package serialization_test

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"
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
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestSerdeValue(t *testing.T) {

	serialization.RegisterImplementation(string(""))
	serialization.RegisterImplementation(ifaces.ColID(""))
	serialization.RegisterImplementation(column.Natural{})
	serialization.RegisterImplementation(column.Shifted{})
	serialization.RegisterImplementation(verifiercol.ConstCol{})
	serialization.RegisterImplementation(verifiercol.FromYs{})
	serialization.RegisterImplementation(verifiercol.FromAccessors{})
	serialization.RegisterImplementation(accessors.FromPublicColumn{})
	serialization.RegisterImplementation(accessors.FromConstAccessor{})
	serialization.RegisterImplementation(query.UnivariateEval{})

	testCases := []struct {
		V    any
		Name string
	}{
		{
			Name: "random-string",
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
			V:    verifiercol.NewConstantCol(field.Element{}, 16),
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
	}

	for i := range testCases {
		t.Run(fmt.Sprintf("test-case-%v/%v", i, testCases[i].Name), func(t *testing.T) {

			msg, err := serialization.Serialize(testCases[i].V)
			require.NoError(t, err)

			fmt.Printf("testcase=%v, msg=%v\n", i, string(msg))

			unmarshaled := reflect.New(reflect.TypeOf(testCases[i].V)).Interface()
			err = serialization.Deserialize(msg, unmarshaled)
			require.NoError(t, err)

			unmarshalledDereferenced := reflect.ValueOf(unmarshaled).Elem().Interface()
			if !compareExportedFields(testCases[i].V, unmarshalledDereferenced, true) {
				t.Errorf("Mismatch in exported fields after full serde value")
			}

		})
	}
}

// compareExportedFields checks if two values are equal, ignoring unexported fields, including in nested structs.
// It logs mismatched fields with their paths and values. If failFast is true, it returns after the first mismatch.
func compareExportedFields(a, b interface{}, failFast bool) bool {
	cachedPtrs := make(map[uintptr]struct{})
	return CompareExportedFieldsWithPath(cachedPtrs, reflect.ValueOf(a), reflect.ValueOf(b), "", failFast)
}

func CompareExportedFieldsWithPath(cachedPtrs map[uintptr]struct{}, a, b reflect.Value, path string, failFast bool) bool {
	// Handle invalid values
	if !a.IsValid() || !b.IsValid() {
		// Treat nil and zero values as equivalent
		if !a.IsValid() && !b.IsValid() {
			return true
		}
		if !a.IsValid() {
			a = reflect.Zero(b.Type())
		}
		if !b.IsValid() {
			b = reflect.Zero(a.Type())
		}
	}

	// Type check after normalization of invalid values
	if a.Type() != b.Type() {
		logrus.Printf("Mismatch at %s: types differ (v1: %v, v2: %v, types: %v, %v)\n", path, a.Interface(), b.Interface(), a.Type(), b.Type())
		return false
	}

	// Specialized handlers
	switch a.Type() {
	case reflect.TypeFor[*symbolic.Expression]():
		return compareSymbolicExpressions(a, b, path)
	case reflect.TypeFor[frontend.Variable]():
		return true
	}

	switch a.Kind() {
	case reflect.Func:
		// Ignore Func
		return true
	case reflect.Interface:
		return CompareExportedFieldsWithPath(cachedPtrs, a.Elem(), b.Elem(), path+".(interface)", failFast)
	case reflect.Map:
		return compareMaps(cachedPtrs, a, b, path, failFast)
	case reflect.Ptr:
		return comparePointers(cachedPtrs, a, b, path, failFast)
	case reflect.Struct:
		return compareStructs(cachedPtrs, a, b, path, failFast)
	case reflect.Slice, reflect.Array:
		return compareSlices(cachedPtrs, a, b, path, failFast)
	default:
		if !reflect.DeepEqual(a.Interface(), b.Interface()) {
			logrus.Printf("Mismatch at %s: values differ (v1: %v, v2: %v, type_v1: %v type_v2: %v)\n", path, a.Interface(), b.Interface(), a.Type(), b.Type())
			return false
		}
		return true
	}
}

func compareSymbolicExpressions(a, b reflect.Value, path string) bool {
	ae := a.Interface().(*symbolic.Expression)
	be := b.Interface().(*symbolic.Expression)

	// If both nil, they are equal
	if ae == nil && be == nil {
		return true
	}

	// If only one is nil, they differ
	if (ae == nil) != (be == nil) {
		logrus.Errorf("Mismatch at %s: one value is nil, the other is not\n", path)
		return false
	}

	// Both non-nil, validate and compare
	errA, errB := ae.Validate(), be.Validate()
	if errA != nil || errB != nil {
		logrus.Errorf("One of the expressions is invalid: path=%s errA=%v, errB=%v\n", path, errA, errB)
		return false
	}

	if ae.ESHash != be.ESHash {
		logrus.Errorf("Mismatch at %s: hashes differ (v1: %v, v2: %v)\n", path, ae.ESHash.Text(16), be.ESHash.Text(16))
		return false
	}

	return true
}

func comparePointers(cachedPtrs map[uintptr]struct{}, a, b reflect.Value, path string, failFast bool) bool {
	if a.IsNil() && b.IsNil() {
		return true
	}

	if a.IsNil() != b.IsNil() {
		logrus.Printf("Mismatch at %s: nil status differs (v1: %v, v2: %v, type: %v)\n", path, a, b, a.Type())
		return false
	}

	if _, seen := cachedPtrs[a.Pointer()]; seen {
		return true
	}

	cachedPtrs[a.Pointer()] = struct{}{}
	return CompareExportedFieldsWithPath(cachedPtrs, a.Elem(), b.Elem(), path, failFast)
}

func compareMaps(cachedPtrs map[uintptr]struct{}, a, b reflect.Value, path string, failFast bool) bool {
	if a.Len() != b.Len() {
		logrus.Printf("Mismatch at %s: map lengths differ (v1: %v, v2: %v, type: %v)\n", path, a.Len(), b.Len(), a.Type())
		return false
	}

	// The module discoverer uses map[ifaces.Column] and map[column.Natural]
	// These use pointers
	switch a.Type().Key() {
	case serialization.TypeOfColumnNatural, reflect.TypeFor[ifaces.Column]():
		return true
	}

	for _, key := range a.MapKeys() {
		valA := a.MapIndex(key)
		valB := b.MapIndex(key)
		if !valB.IsValid() {
			logrus.Printf("Mismatch at %s: key %v is missing in second map\n", path, key)
			return false
		}
		keyPath := fmt.Sprintf("%s[%v]", path, key)
		if !CompareExportedFieldsWithPath(cachedPtrs, valA, valB, keyPath, failFast) {
			if failFast {
				return false
			}
		}
	}
	return true
}

func compareStructs(cachedPtrs map[uintptr]struct{}, a, b reflect.Value, path string, failFast bool) bool {
	equal := true
	for i := 0; i < a.NumField(); i++ {
		structField := a.Type().Field(i)

		// When the field has the omitted tag, we skip it there without any warning.
		if tag, hasTag := structField.Tag.Lookup(serialization.SerdeStructTag); hasTag {
			if strings.Contains(tag, serialization.SerdeStructTagOmit) ||
				strings.Contains(tag, serialization.SerdeStructTagTestOmit) {
				continue
			}
		}

		// Skip unexported fields
		if !structField.IsExported() {
			continue
		}

		fieldA := a.Field(i)
		fieldB := b.Field(i)
		fieldName := structField.Name
		fieldPath := fieldName
		if path != "" {
			fieldPath = path + "." + fieldName
		}

		if !CompareExportedFieldsWithPath(cachedPtrs, fieldA, fieldB, fieldPath, failFast) {
			equal = false
			if failFast {
				return false
			}
		}
	}
	return equal
}

func compareSlices(cachedPtrs map[uintptr]struct{}, a, b reflect.Value, path string, failFast bool) bool {
	if a.Len() != b.Len() {
		logrus.Printf("Mismatch at %s: slice lengths differ (v1: %v, v2: %v, type: %v)\n", path, a.Len(), b.Len(), a.Type())
		return false
	}

	equal := true
	for i := 0; i < a.Len(); i++ {
		elemPath := fmt.Sprintf("%s[%d]", path, i)
		if !CompareExportedFieldsWithPath(cachedPtrs, a.Index(i), b.Index(i), elemPath, failFast) {
			equal = false
			if failFast {
				return false
			}
		}
	}
	return equal
}
