package serialization

import (
	"fmt"
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
		CompiledIOP *wizard.CompiledIOP
	}{
		{
			V:        "someRandomString",
			Expected: "psomeRandomString",
		},
		{
			V: func() any {
				var s = ifaces.ColID("someIndirectedString")
				return &s
			}(),
			Expected: "tsomeIndirectedString",
		},
		{
			V: func() any {
				// It's important to not provide an untyped string under
				// the interface because the type cannot be serialized.
				var s any = string("someStringUnderIface")
				return &s
			}(),
			Expected: "\xa2dtypei#string#0evalueUtsomeStringUnderIface",
		},
		{
			V: func() any {
				var id = ifaces.ColID("newTypeUnderIface")
				var s any = &id
				return &s
			}(),
			Expected: "\xa2dtypex\x18/protocol/ifaces#ColID#1evalueRqnewTypeUnderIface",
		},
		{
			V:        ifaces.QueryID("QueryID"),
			Expected: "gQueryID",
		},
		func() struct {
			V           any
			Expected    string
			CompiledIOP *wizard.CompiledIOP
		} {

			comp := wizard.NewCompiledIOP()
			nat := comp.InsertColumn(0, "myNaturalColumn", 16, column.Committed)
			var v any = &nat

			return struct {
				V           any
				Expected    string
				CompiledIOP *wizard.CompiledIOP
			}{
				V:           v,
				Expected:    "\xa2dtypex\x1a/protocol/column#Natural#0evaluePomyNaturalColumn",
				CompiledIOP: comp,
			}
		}(),
		func() struct {
			V           any
			Expected    string
			CompiledIOP *wizard.CompiledIOP
		} {

			comp := wizard.NewCompiledIOP()
			nat := comp.InsertColumn(0, "myNaturalColumn", 16, column.Committed)
			nat = column.Shift(nat, 2)
			var v any = &nat

			return struct {
				V           any
				Expected    string
				CompiledIOP *wizard.CompiledIOP
			}{
				V:           v,
				Expected:    "\xa2dtypex\x1a/protocol/column#Shifted#0evalueXL\xa2foffsetA\x02fparentX9\xa2dtypex\x1a/protocol/column#Natural#0evaluePomyNaturalColumn",
				CompiledIOP: comp,
			}
		}(),
		func() struct {
			V           any
			Expected    string
			CompiledIOP *wizard.CompiledIOP
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
				V           any
				Expected    string
				CompiledIOP *wizard.CompiledIOP
			}{
				V:           &col,
				Expected:    "\xa2dtypex,/protocol/column/verifiercol#FromAccessors#0evalueX\xfe\xa4dsizeA\beroundA\x00gpaddingI\x84A\x00A\x00A\x00A\x00iaccessorsXЃXC\xa2dtypex&/protocol/accessors#FromPublicColumn#1evalueN\xa2ccolBaacposA\x00XC\xa2dtypex&/protocol/accessors#FromPublicColumn#1evalueN\xa2ccolBabcposA\x00XC\xa2dtypex&/protocol/accessors#FromPublicColumn#1evalueN\xa2ccolBaccposA\x00",
				CompiledIOP: comp,
			}
		}(),
		func() struct {
			V           any
			Expected    string
			CompiledIOP *wizard.CompiledIOP
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
				V           any
				Expected    string
				CompiledIOP *wizard.CompiledIOP
			}{
				V:           &univ,
				Expected:    "\xa2dtypex /protocol/query#UnivariateEval#0evalueY\x01a\xa2dpolsY\x01J\x83X+\xa2dtypex\x1a/protocol/column#Natural#0evalueBaaXh\xa2dtypex\x1a/protocol/column#Shifted#0evalueX>\xa2foffsetA\x02fparentX+\xa2dtypex\x1a/protocol/column#Natural#0evalueBaaX\xb0\xa2dtypex,/protocol/column/verifiercol#FromAccessors#0evalueXt\xa4dsizeA\x04eroundA\x00gpaddingI\x84A\x00A\x00A\x00A\x00iaccessorsXF\x81XC\xa2dtypex&/protocol/accessors#FromPublicColumn#1evalueN\xa2ccolBabcposA\x00gqueryIdEduniv",
				CompiledIOP: comp,
			}
		}(),
		func() struct {
			V           any
			Expected    string
			CompiledIOP *wizard.CompiledIOP
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
				V           any
				Expected    string
				CompiledIOP *wizard.CompiledIOP
			}{
				V:           &fromYs,
				Expected:    "\xa2dtypex%/protocol/column/verifiercol#FromYs#0evalueXx\xa3equeryEduniveroundA\x00frangesXZ\x84BaaMlSHIFT_2_16_aBabXCxAFROM_ACCESSORS_FROM_COLUMN_POSITION_ACCESSOR_b_0_PADDING=0_SIZE=4",
				CompiledIOP: comp,
			}
		}(),
		{
			V:           coin.NewInfo("foo", coin.IntegerVec, 16, 16, 1),
			Expected:    "\xa5dnameDcfoodsizeA\x10dtypeA\x01eroundA\x01jupperBoundA\x10",
			CompiledIOP: nil,
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
			require.Equal(t, testCases[i].V, unmarshalledDereferenced, "wrong deserialization, \n\tleft=%++v\n\tright=%++v", testCases[i].V, unmarshalledDereferenced)
		})
	}
}
