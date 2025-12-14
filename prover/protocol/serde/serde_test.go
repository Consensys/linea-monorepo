package serde_test

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/consensys/gnark/constraint" // <--- CRITICAL IMPORT
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
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func init() {
	serde.RegisterImplementation(string(""))
	serde.RegisterImplementation(ifaces.ColID(""))
	serde.RegisterImplementation(ifaces.QueryID(""))
	serde.RegisterImplementation(column.Natural{})
	serde.RegisterImplementation(column.Shifted{})
	serde.RegisterImplementation(verifiercol.ConstCol{})
	serde.RegisterImplementation(verifiercol.FromYs{})
	serde.RegisterImplementation(verifiercol.FromAccessors{})

	// FIX: Added extra '1' to satisfy IntegerVec requirement (nbIntegers, upperBound)
	serde.RegisterImplementation(verifiercol.NewFromIntVecCoin(nil, coin.NewInfo("t", coin.IntegerVec, 1, 1, 1)))

	serde.RegisterImplementation(accessors.FromPublicColumn{})
	serde.RegisterImplementation(accessors.FromConstAccessor{})
	serde.RegisterImplementation(accessors.NewFromCoin(coin.NewInfo("t", coin.Field, 1)))
	serde.RegisterImplementation(query.UnivariateEval{})
	serde.RegisterImplementation(query.MiMC{})
	serde.RegisterImplementation(query.Horner{})
	serde.RegisterImplementation(symbolic.Variable{})
	serde.RegisterImplementation(symbolic.Constant{})
	serde.RegisterImplementation(symbolic.Product{})
	serde.RegisterImplementation(symbolic.LinComb{})
	serde.RegisterImplementation(symbolic.Add(nil, nil))
	serde.RegisterImplementation(symbolic.Mul(nil, nil))
	serde.RegisterImplementation(coin.Info{})
	serde.RegisterImplementation(coin.Name(""))

	// Recursion hints
	serde.RegisterImplementation(&constraint.BlueprintGenericHint{})

}

var (
	SerdeStructTag         = "serde"
	SerdeStructTagOmit     = "omit"
	SerdeStructTagTestOmit = "test_omit"
	TypeOfColumnNatural    = reflect.TypeOf(column.Natural{})
)

func deepCmp(a, b interface{}, failFast bool) bool {
	cachedPtrs := make(map[uintptr]struct{})
	return compareExportedFieldsWithPath(cachedPtrs, reflect.ValueOf(a), reflect.ValueOf(b), "", failFast)
}

func compareExportedFieldsWithPath(cachedPtrs map[uintptr]struct{}, a, b reflect.Value, path string, failFast bool) bool {
	if !a.IsValid() || !b.IsValid() {
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

	if a.Type() != b.Type() {
		logrus.Printf("Mismatch at %s: types differ (v1: %v, v2: %v, types: %v, %v)\n", path, a.Interface(), b.Interface(), a.Type(), b.Type())
		return false
	}

	if a.Type() == reflect.TypeOf((*symbolic.Expression)(nil)) {
		return compareSymbolicExpressions(a, b, path)
	}
	if a.Type() == reflect.TypeOf(frontend.Variable(nil)) {
		return true
	}

	switch a.Kind() {
	case reflect.Func:
		return true
	case reflect.Interface:
		return compareExportedFieldsWithPath(cachedPtrs, a.Elem(), b.Elem(), path+".(interface)", failFast)
	case reflect.Map:
		return compareMaps(cachedPtrs, a, b, path, failFast)
	case reflect.Ptr:
		return comparePointers(cachedPtrs, a, b, path, failFast)
	case reflect.Struct:
		return compareStructs(cachedPtrs, a, b, path, failFast)
	case reflect.Slice, reflect.Array:
		return compareSlices(cachedPtrs, a, b, path, failFast)
	default:
		if a.Type() == reflect.TypeOf(big.Int{}) {
			biA := a.Interface().(big.Int)
			biB := b.Interface().(big.Int)
			if biA.Cmp(&biB) != 0 {
				logrus.Printf("Mismatch at %s: BigInts differ (v1: %v, v2: %v)\n", path, biA.String(), biB.String())
				return false
			}
			return true
		}
		if !reflect.DeepEqual(a.Interface(), b.Interface()) {
			logrus.Printf("Mismatch at %s: values differ (v1: %v, v2: %v)\n", path, a.Interface(), b.Interface())
			return false
		}
		return true
	}
}

func compareSymbolicExpressions(a, b reflect.Value, path string) bool {
	if a.IsNil() && b.IsNil() {
		return true
	}
	if a.IsNil() != b.IsNil() {
		return false
	}
	ae := a.Interface().(*symbolic.Expression)
	be := b.Interface().(*symbolic.Expression)
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
		logrus.Printf("Mismatch at %s: nil status differs\n", path)
		return false
	}
	if _, seen := cachedPtrs[a.Pointer()]; seen {
		return true
	}
	cachedPtrs[a.Pointer()] = struct{}{}
	return compareExportedFieldsWithPath(cachedPtrs, a.Elem(), b.Elem(), path, failFast)
}

func compareMaps(cachedPtrs map[uintptr]struct{}, a, b reflect.Value, path string, failFast bool) bool {
	if a.Len() != b.Len() {
		logrus.Printf("Mismatch at %s: map lengths differ\n", path)
		return false
	}
	if a.Type().Key() == TypeOfColumnNatural || a.Type().Key() == reflect.TypeOf((*ifaces.Column)(nil)).Elem() {
		return true
	}
	for _, key := range a.MapKeys() {
		valA := a.MapIndex(key)
		valB := b.MapIndex(key)
		if !valB.IsValid() {
			logrus.Printf("Mismatch at %s: key %v missing\n", path, key)
			return false
		}
		if !compareExportedFieldsWithPath(cachedPtrs, valA, valB, fmt.Sprintf("%s[%v]", path, key), failFast) {
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
		if tag, hasTag := structField.Tag.Lookup(SerdeStructTag); hasTag {
			if strings.Contains(tag, SerdeStructTagOmit) || strings.Contains(tag, SerdeStructTagTestOmit) {
				continue
			}
		}
		if !structField.IsExported() {
			continue
		}
		fieldPath := structField.Name
		if path != "" {
			fieldPath = path + "." + structField.Name
		}
		if !compareExportedFieldsWithPath(cachedPtrs, a.Field(i), b.Field(i), fieldPath, failFast) {
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
		logrus.Printf("Mismatch at %s: slice lengths differ (%d vs %d)\n", path, a.Len(), b.Len())
		return false
	}
	equal := true
	for i := 0; i < a.Len(); i++ {
		elemPath := fmt.Sprintf("%s[%d]", path, i)
		if !compareExportedFieldsWithPath(cachedPtrs, a.Index(i), b.Index(i), elemPath, failFast) {
			equal = false
			if failFast {
				return false
			}
		}
	}
	return equal
}

func TestZeroCopySerdeValue(t *testing.T) {
	testCases := []struct {
		V    any
		Name string
	}{
		{Name: "random-string", V: "someRandomString"},
		{Name: "query-id", V: ifaces.QueryID("QueryID")},
		{Name: "bigint-max-uint256", V: func() any {
			v, _ := new(big.Int).SetString("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 0)
			return v
		}()},
		{Name: "field-2**248-1", V: func() any {
			v, _ := new(field.Element).SetString("0x00ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
			return v
		}()},
		{Name: "vector-1234", V: vector.ForTest(0, 1, 2, 3, 4, 5, 5, 6, 7)},
		{Name: "natural-column", V: func() any {
			comp := wizard.NewCompiledIOP()
			return comp.InsertColumn(0, "myNaturalColumn", 16, column.Committed)
		}()},
		{Name: "map-with-column-as-key", V: func() any {
			comp := wizard.NewCompiledIOP()
			a := comp.InsertColumn(0, "a", 16, column.Committed).(column.Natural)
			b := comp.InsertColumn(0, "b", 16, column.Committed).(column.Natural)
			return map[column.Natural]string{a: "a", b: "b"}
		}()},
		{Name: "recursion", V: func() any {
			wiop := wizard.NewCompiledIOP()
			a := wiop.InsertCommit(0, "a", 1<<10)
			wiop.InsertUnivariate(0, "u", []ifaces.Column{a})
			wizard.ContinueCompilation(wiop, vortex.Compile(2, vortex.WithOptionalSISHashingThreshold(0), vortex.ForceNumOpenedColumns(2), vortex.PremarkAsSelfRecursed()))
			rec := wizard.NewCompiledIOP()
			recursion.DefineRecursionOf(rec, wiop, recursion.Parameters{MaxNumProof: 1, WithoutGkr: true, Name: "recursion"})
			return rec
		}()},
		{Name: "mutex", V: []*sync.Mutex{{}, {}}}, // Should serialize to empty/nil
	}

	for i := range testCases {
		t.Run(fmt.Sprintf("%d/%s", i, testCases[i].Name), func(t *testing.T) {
			flatBytes, err := serde.Serialize(testCases[i].V)
			require.NoError(t, err)

			t.Logf("testcase=%v, size=%v bytes\n", i, len(flatBytes))

			view, err := serde.NewView(flatBytes)
			require.NoError(t, err)

			unmarshaledPtr := reflect.New(reflect.TypeOf(testCases[i].V)).Interface()
			err = view.Deserialize(unmarshaledPtr)
			require.NoError(t, err)

			unmarshalledDereferenced := reflect.ValueOf(unmarshaledPtr).Elem().Interface()
			if !deepCmp(testCases[i].V, unmarshalledDereferenced, true) {
				t.Errorf("Mismatch in exported fields after full zero-copy serde")
			}
		})
	}
}
