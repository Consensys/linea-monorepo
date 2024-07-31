package serialization

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/accessors"
	"github.com/consensys/zkevm-monorepo/prover/protocol/coin"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/query"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/require"
)

func TestSerializeValue(t *testing.T) {

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
		V           any
		Expected    string
		Mode        mode
		CompiledIOP *wizard.CompiledIOP
	}{
		{
			V:        "someRandomString",
			Expected: `"someRandomString"`,
			Mode:     DeclarationMode,
		},
		{
			V:        "someRandomString",
			Expected: `"someRandomString"`,
			Mode:     DeclarationMode,
		},
		{
			V: func() any {
				var s = ifaces.ColID("someIndirectedString")
				return &s
			}(),
			Expected: `"someIndirectedString"`,
			Mode:     DeclarationMode,
		},
		{
			V: func() any {
				// It's important to not provide an untyped string under
				// the interface because the type cannot be serialized.
				var s any = string("someStringUnderIface")
				return &s
			}(),
			Expected: `{"type":"#string#0","value":"someStringUnderIface"}`,
			Mode:     DeclarationMode,
		},
		{
			V: func() any {
				var id = ifaces.ColID("newTypeUnderIface")
				var s any = &id
				return &s
			}(),
			Expected: `{"type":"/protocol/ifaces#ColID#1","value":"newTypeUnderIface"}`,
		},
		{
			V:        ifaces.QueryID("QueryID"),
			Expected: `"QueryID"`,
			Mode:     DeclarationMode,
		},
		func() struct {
			V           any
			Expected    string
			Mode        mode
			CompiledIOP *wizard.CompiledIOP
		} {

			comp := newEmptyCompiledIOP()
			nat := comp.InsertColumn(0, "myNaturalColumn", 16, column.Committed)
			var v any = &nat

			return struct {
				V           any
				Expected    string
				Mode        mode
				CompiledIOP *wizard.CompiledIOP
			}{
				V:           v,
				Expected:    "{\"type\":\"/protocol/column#Natural#0\",\"value\":\"myNaturalColumn\"}",
				Mode:        ReferenceMode,
				CompiledIOP: comp,
			}
		}(),
		func() struct {
			V           any
			Expected    string
			Mode        mode
			CompiledIOP *wizard.CompiledIOP
		} {

			comp := newEmptyCompiledIOP()
			nat := comp.InsertColumn(0, "myNaturalColumn", 16, column.Committed)
			nat = column.Shift(nat, 2)
			var v any = &nat

			return struct {
				V           any
				Expected    string
				Mode        mode
				CompiledIOP *wizard.CompiledIOP
			}{
				V:           v,
				Expected:    "{\"type\":\"/protocol/column#Shifted#0\",\"value\":{\"offset\":2,\"parent\":{\"type\":\"/protocol/column#Natural#0\",\"value\":\"myNaturalColumn\"}}}",
				Mode:        ReferenceMode,
				CompiledIOP: comp,
			}
		}(),
		func() struct {
			V           any
			Expected    string
			Mode        mode
			CompiledIOP *wizard.CompiledIOP
		} {

			comp := newEmptyCompiledIOP()

			col := verifiercol.NewConcatTinyColumns(
				comp,
				8,
				field.Element{},
				comp.InsertColumn(0, "a", 1, column.Proof),
				comp.InsertColumn(0, "b", 1, column.Proof),
				comp.InsertColumn(0, "c", 1, column.Proof),
			)

			return struct {
				V           any
				Expected    string
				Mode        mode
				CompiledIOP *wizard.CompiledIOP
			}{
				V:           &col,
				Expected:    "{\"type\":\"/protocol/column/verifiercol#FromAccessors#0\",\"value\":{\"accessors\":[{\"type\":\"/protocol/accessors#FromPublicColumn#1\",\"value\":{\"col\":\"a\",\"pos\":0}},{\"type\":\"/protocol/accessors#FromPublicColumn#1\",\"value\":{\"col\":\"b\",\"pos\":0}},{\"type\":\"/protocol/accessors#FromPublicColumn#1\",\"value\":{\"col\":\"c\",\"pos\":0}},{\"type\":\"/protocol/accessors#FromConstAccessor#1\",\"value\":{\"f\":[0,0,0,0]}},{\"type\":\"/protocol/accessors#FromConstAccessor#1\",\"value\":{\"f\":[0,0,0,0]}},{\"type\":\"/protocol/accessors#FromConstAccessor#1\",\"value\":{\"f\":[0,0,0,0]}},{\"type\":\"/protocol/accessors#FromConstAccessor#1\",\"value\":{\"f\":[0,0,0,0]}},{\"type\":\"/protocol/accessors#FromConstAccessor#1\",\"value\":{\"f\":[0,0,0,0]}}],\"round\":0}}",
				Mode:        ReferenceMode,
				CompiledIOP: comp,
			}
		}(),
		func() struct {
			V           any
			Expected    string
			Mode        mode
			CompiledIOP *wizard.CompiledIOP
		} {

			comp := newEmptyCompiledIOP()

			var (
				a                   = comp.InsertColumn(0, "a", 16, column.Committed)
				aNext               = column.Shift(a, 2)
				tiny                = comp.InsertColumn(0, "b", 1, column.Proof)
				concat              = verifiercol.NewConcatTinyColumns(comp, 4, field.Element{}, tiny)
				univ   ifaces.Query = comp.InsertUnivariate(0, "univ", []ifaces.Column{a, aNext, concat})
			)

			return struct {
				V           any
				Expected    string
				Mode        mode
				CompiledIOP *wizard.CompiledIOP
			}{
				V:           &univ,
				Expected:    "{\"type\":\"/protocol/query#UnivariateEval#0\",\"value\":{\"pols\":[{\"type\":\"/protocol/column#Natural#0\",\"value\":\"a\"},{\"type\":\"/protocol/column#Shifted#0\",\"value\":{\"offset\":2,\"parent\":{\"type\":\"/protocol/column#Natural#0\",\"value\":\"a\"}}},{\"type\":\"/protocol/column/verifiercol#FromAccessors#0\",\"value\":{\"accessors\":[{\"type\":\"/protocol/accessors#FromPublicColumn#1\",\"value\":{\"col\":\"b\",\"pos\":0}},{\"type\":\"/protocol/accessors#FromConstAccessor#1\",\"value\":{\"f\":[0,0,0,0]}},{\"type\":\"/protocol/accessors#FromConstAccessor#1\",\"value\":{\"f\":[0,0,0,0]}},{\"type\":\"/protocol/accessors#FromConstAccessor#1\",\"value\":{\"f\":[0,0,0,0]}}],\"round\":0}}],\"queryId\":\"univ\"}}",
				Mode:        DeclarationMode,
				CompiledIOP: comp,
			}
		}(),
		func() struct {
			V           any
			Expected    string
			Mode        mode
			CompiledIOP *wizard.CompiledIOP
		} {

			comp := newEmptyCompiledIOP()

			var (
				a      = comp.InsertColumn(0, "a", 16, column.Committed)
				aNext  = column.Shift(a, 2)
				tiny   = comp.InsertColumn(0, "b", 1, column.Proof)
				concat = verifiercol.NewConcatTinyColumns(comp, 4, field.Element{}, tiny)
				univ   = comp.InsertUnivariate(0, "univ", []ifaces.Column{a, aNext, tiny, concat})
				fromYs = verifiercol.NewFromYs(comp, univ, []ifaces.ColID{a.GetColID(), aNext.GetColID(), tiny.GetColID(), concat.GetColID()})
			)

			return struct {
				V           any
				Expected    string
				Mode        mode
				CompiledIOP *wizard.CompiledIOP
			}{
				V:           &fromYs,
				Expected:    "{\"type\":\"/protocol/column/verifiercol#FromYs#0\",\"value\":{\"query\":\"univ\",\"ranges\":[\"a\",\"SHIFT_2_16_a\",\"b\",\"FROM_ACCESSORS_FROM_COLUMN_POSITION_ACCESSOR_b_0_CONST_ACCESSOR_0_CONST_ACCESSOR_0_CONST_ACCESSOR_0\"],\"round\":0}}",
				Mode:        ReferenceMode,
				CompiledIOP: comp,
			}
		}(),
		{
			V:           coin.Info{Type: coin.IntegerVec, Size: 16, UpperBound: 16, Name: "foo", Round: 1},
			Expected:    "{\"name\":\"foo\",\"round\":1,\"size\":16,\"type\":1,\"upperBound\":16}",
			Mode:        ReferenceMode,
			CompiledIOP: nil,
		},
	}

	for i := range testCases {
		t.Run(fmt.Sprintf("test-case-%v", i), func(t *testing.T) {

			v := reflect.ValueOf(testCases[i].V)
			msg, err := SerializeValue(v, testCases[i].Mode)
			require.NoError(t, err)
			require.Equal(t, testCases[i].Expected, string(msg), "wrong serialization")

			deserialized, err := DeserializeValue(msg, testCases[i].Mode, v.Type(), testCases[i].CompiledIOP)
			require.NoError(t, err)
			require.Equal(t, testCases[i].V, deserialized.Interface(), "wrong deserialization")
		})
	}

}
