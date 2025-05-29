package serialization

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestSerdeValue(t *testing.T) {

	backupRegistryAndReset()
	defer restoreRegistryFromBackup()

	RegisterImplementation(string(""))
	RegisterImplementation(ifaces.ColID(""))
	RegisterImplementation(column.Natural{})
	RegisterImplementation(column.Shifted{})
	RegisterImplementation(verifiercol.ConstCol{})
	RegisterImplementation(verifiercol.FromYs{})
	RegisterImplementation(verifiercol.FromAccessors{})
	RegisterImplementation(accessors.FromPublicColumn{})
	RegisterImplementation(accessors.FromConstAccessor{})
	RegisterImplementation(query.UnivariateEval{})

	testCases := []struct {
		V any
	}{
		{
			V: "someRandomString",
		},
		{
			V: func() any {
				var s = ifaces.ColID("someIndirectedString")
				return &s
			}(),
		},
		{
			V: func() any {
				// It's important to not provide an untyped string under
				// the interface because the type cannot be serialized.
				var s any = string("someStringUnderIface")
				return &s
			}(),
		},
		{
			V: func() any {
				var id = ifaces.ColID("newTypeUnderIface")
				var s any = &id
				return &s
			}(),
		},
		{
			V: ifaces.QueryID("QueryID"),
		},
		func() struct {
			V any
		} {

			comp := wizard.NewCompiledIOP()
			nat := comp.InsertColumn(0, "myNaturalColumn", 16, column.Committed)
			var v any = &nat

			return struct {
				V any
			}{
				V: v,
			}
		}(),
		func() struct {
			V any
		} {

			comp := wizard.NewCompiledIOP()
			nat := comp.InsertColumn(0, "myNaturalColumn", 16, column.Committed)
			nat = column.Shift(nat, 2)
			var v any = &nat

			return struct {
				V any
			}{
				V: v,
			}
		}(),
		func() struct {
			V any
		} {

			comp := wizard.NewCompiledIOP()

			col := verifiercol.NewConcatTinyColumns(
				comp,
				8,
				field.Element{},
				comp.InsertColumn(0, "a", 1, column.Proof),
				comp.InsertColumn(0, "b", 1, column.Proof),
				comp.InsertColumn(0, "c", 1, column.Proof),
			)

			return struct {
				V any
			}{
				V: &col,
			}
		}(),
		func() struct {
			V any
		} {

			comp := wizard.NewCompiledIOP()

			var (
				a                   = comp.InsertColumn(0, "a", 16, column.Committed)
				aNext               = column.Shift(a, 2)
				tiny                = comp.InsertColumn(0, "b", 1, column.Proof)
				concat              = verifiercol.NewConcatTinyColumns(comp, 4, field.Element{}, tiny)
				univ   ifaces.Query = comp.InsertUnivariate(0, "univ", []ifaces.Column{a, aNext, concat})
			)

			return struct {
				V any
			}{
				V: &univ,
			}
		}(),
		func() struct {
			V any
		} {

			comp := wizard.NewCompiledIOP()

			var (
				a      = comp.InsertColumn(0, "a", 16, column.Committed)
				aNext  = column.Shift(a, 2)
				tiny   = comp.InsertColumn(0, "b", 1, column.Proof)
				concat = verifiercol.NewConcatTinyColumns(comp, 4, field.Element{}, tiny)
				univ   = comp.InsertUnivariate(0, "univ", []ifaces.Column{a, aNext, tiny, concat})
				fromYs = verifiercol.NewFromYs(comp, univ, []ifaces.ColID{a.GetColID(), aNext.GetColID(), tiny.GetColID(), concat.GetColID()})
			)

			return struct {
				V any
			}{
				V: &fromYs,
			}
		}(),
		{
			V: coin.NewInfo("foo", coin.IntegerVec, 16, 16, 1),
		},
		{
			V: big.NewInt(0),
		},
		{
			V: big.NewInt(1),
		},
		{
			V: big.NewInt(-1),
		},
		{
			V: func() any {
				v, ok := new(big.Int).SetString("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 0)
				if !ok {
					panic("bigint does not work")
				}
				return v
			}(),
		},
	}

	for i := range testCases {
		t.Run(fmt.Sprintf("test-case-%v", i), func(t *testing.T) {

			msg, err := Serialize(testCases[i].V)
			require.NoError(t, err)

			fmt.Printf("testcase=%v, msg=%v\n", i, string(msg))

			unmarshaled := reflect.New(reflect.TypeOf(testCases[i].V)).Interface()
			err = Deserialize(msg, unmarshaled)
			require.NoError(t, err)

			unmarshalledDereferenced := reflect.ValueOf(unmarshaled).Elem().Interface()
			if !CompareExportedFields(testCases[i].V, unmarshalledDereferenced) {
				t.Errorf("Mismatch in exported fields after full serde value")
			}

		})
	}
}

// compareExportedFields checks if two values are equal, ignoring unexported fields, including in nested structs.
// It logs mismatched fields with their paths and values.
func CompareExportedFields(a, b interface{}) bool {
	return CompareExportedFieldsWithPath(a, b, "")
}

func CompareExportedFieldsWithPath(a, b interface{}, path string) bool {
	v1, v2 := reflect.ValueOf(a), reflect.ValueOf(b)
	// Ensure both values are valid
	if !v1.IsValid() || !v2.IsValid() {
		// Treat nil and zero values as equivalent
		if !v1.IsValid() && !v2.IsValid() {
			return true
		}
		if !v1.IsValid() {
			v1 = reflect.Zero(v2.Type())
		}
		if !v2.IsValid() {
			v2 = reflect.Zero(v1.Type())
		}
	}

	// Skip ignorable fields
	if IsIgnoreableType(v1.Type()) {
		logrus.Printf("Skipping comparison of Ignorable type:%s at %s\n", v1.Type().String(), path)
		return true
	}

	// Ensure same type
	if v1.Type() != v2.Type() {
		logrus.Printf("Mismatch at %s: types differ (v1: %v, v2: %v, types: %v, %v)\n", path, a, b, v1.Type(), v2.Type())
		return false
	}

	// Ignore Func
	if v1.Kind() == reflect.Func {
		return true
	}

	// Handle maps
	if v1.Kind() == reflect.Map {
		if v1.Len() != v2.Len() {
			if IsIgnoreableType(v1.Type()) {
				logrus.Printf("Skipping comparison of ignoreable types at %s\n", path)
				return true
			}
			logrus.Printf("Mismatch at %s: map lengths differ (v1: %v, v2: %v, type: %v)\n", path, v1.Len(), v2.Len(), v1.Type())
			return false
		}
		for _, key := range v1.MapKeys() {
			value1 := v1.MapIndex(key)
			value2 := v2.MapIndex(key)
			if !value2.IsValid() {
				logrus.Printf("Mismatch at %s: key %v is missing in second map\n", path, key)
				return false
			}
			keyPath := fmt.Sprintf("%s[%v]", path, key)
			if !CompareExportedFieldsWithPath(value1.Interface(), value2.Interface(), keyPath) {
				return false
			}
		}
		// logrus.Infof("Comparing map at %s: len(v1)=%d, len(v2)=%d", path, v1.Len(), v2.Len())
		return true
	}

	// Handle pointers by dereferencing
	if v1.Kind() == reflect.Ptr {
		if v1.IsNil() && v2.IsNil() {
			return true
		}
		if v1.IsNil() != v2.IsNil() {
			if IsIgnoreableType(v1.Type()) {
				logrus.Printf("Skipping comparison of ignoreable types at %s\n", path)
				return true
			}
			logrus.Printf("Mismatch at %s: nil status differs (v1: %v, v2: %v, type: %v)\n", path, a, b, v1.Type())
			return false
		}
		return CompareExportedFieldsWithPath(v1.Elem().Interface(), v2.Elem().Interface(), path)
	}

	// Handle structs
	if v1.Kind() == reflect.Struct {
		equal := true
		for i := 0; i < v1.NumField(); i++ {
			structField := v1.Type().Field(i)
			// Skip unexported fields
			if !structField.IsExported() {
				continue
			}

			f1 := v1.Field(i)
			f2 := v2.Field(i)
			fieldName := structField.Name
			fieldPath := fieldName
			if path != "" {
				fieldPath = path + "." + fieldName
			}
			if !CompareExportedFieldsWithPath(f1.Interface(), f2.Interface(), fieldPath) {
				equal = false
			}
		}
		return equal
	}

	// Handle slices or arrays
	if v1.Kind() == reflect.Slice || v1.Kind() == reflect.Array {
		if v1.Len() != v2.Len() {
			logrus.Printf("Mismatch at %s: slice lengths differ (v1: %v, v2: %v, type: %v)\n", path, v1, v2, v1.Type())
			return false
		}
		equal := true
		for i := 0; i < v1.Len(); i++ {
			elemPath := fmt.Sprintf("%s[%d]", path, i)
			if !CompareExportedFieldsWithPath(v1.Index(i).Interface(), v2.Index(i).Interface(), elemPath) {
				equal = false
			}
		}
		return equal
	}

	// For other types, use DeepEqual and log if mismatched
	if !reflect.DeepEqual(a, b) {
		logrus.Printf("Mismatch at %s: values differ (v1: %v, v2: %v, type_v1: %v type_v2: %v)\n", path, a, b, v1.Type(), v2.Type())
		return false
	}
	return true
}
