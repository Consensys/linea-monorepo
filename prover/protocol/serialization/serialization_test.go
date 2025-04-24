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
		Mode        Mode
		CompiledIOP *wizard.CompiledIOP
	}{
		{
			V:        "someRandomString",
			Expected: "psomeRandomString",
			Mode:     DeclarationMode,
		},
		{
			V:        "someRandomString",
			Expected: "psomeRandomString",
			Mode:     DeclarationMode,
		},
		{
			V: func() any {
				var s = ifaces.ColID("someIndirectedString")
				return &s
			}(),
			Expected: "tsomeIndirectedString",
			Mode:     DeclarationMode,
		},
		{
			V: func() any {
				var s any = string("someStringUnderIface")
				return &s
			}(),
			Expected: "\xa2dtypei#string#0evalueUtsomeStringUnderIface",
			Mode:     DeclarationMode,
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
			Mode:     DeclarationMode,
		},
		func() struct {
			V           any
			Expected    string
			Mode        Mode
			CompiledIOP *wizard.CompiledIOP
		} {
			comp := newEmptyCompiledIOP()
			nat := comp.InsertColumn(0, "myNaturalColumn", 16, column.Committed)
			var v any = &nat
			return struct {
				V           any
				Expected    string
				Mode        Mode
				CompiledIOP *wizard.CompiledIOP
			}{
				V:           v,
				Expected:    "\xa2dtypex\x1a/protocol/column#Natural#0evaluePomyNaturalColumn",
				Mode:        ReferenceMode,
				CompiledIOP: comp,
			}
		}(),
		func() struct {
			V           any
			Expected    string
			Mode        Mode
			CompiledIOP *wizard.CompiledIOP
		} {
			comp := newEmptyCompiledIOP()
			nat := comp.InsertColumn(0, "myNaturalColumn", 16, column.Committed)
			nat = column.Shift(nat, 2)
			var v any = &nat
			return struct {
				V           any
				Expected    string
				Mode        Mode
				CompiledIOP *wizard.CompiledIOP
			}{
				V:           v,
				Expected:    "\xa2dtypex\x1a/protocol/column#Shifted#0evalueXL\xa2foffsetA\x02fparentX9\xa2dtypex\x1a/protocol/column#Natural#0evaluePomyNaturalColumn",
				Mode:        ReferenceMode,
				CompiledIOP: comp,
			}
		}(),
		func() struct {
			V           any
			Expected    string
			Mode        Mode
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
				Mode        Mode
				CompiledIOP *wizard.CompiledIOP
			}{
				V:           &col,
				Expected:    "\xa2dtypex,/protocol/column/verifiercol#FromAccessors#0evalueX\xfe\xa4iaccessorsXЃXC\xa2dtypex&/protocol/accessors#FromPublicColumn#1evalueN\xa2ccolBaacposA\x00XC\xa2dtypex&/protocol/accessors#FromPublicColumn#1evalueN\xa2ccolBabcposA\x00XC\xa2dtypex&/protocol/accessors#FromPublicColumn#1evalueN\xa2ccolBaccposA\x00gpaddingI\x84A\x00A\x00A\x00A\x00eroundA\x00dsizeA\b",
				Mode:        ReferenceMode,
				CompiledIOP: comp,
			}
		}(),
		func() struct {
			V           any
			Expected    string
			Mode        Mode
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
				Mode        Mode
				CompiledIOP *wizard.CompiledIOP
			}{
				V:           &univ,
				Expected:    "\xa2dtypex /protocol/query#UnivariateEval#0evalueY\x01a\xa2dpolsY\x01J\x83X+\xa2dtypex\x1a/protocol/column#Natural#0evalueBaaXh\xa2dtypex\x1a/protocol/column#Shifted#0evalueX>\xa2foffsetA\x02fparentX+\xa2dtypex\x1a/protocol/column#Natural#0evalueBaaX\xb0\xa2dtypex,/protocol/column/verifiercol#FromAccessors#0evalueXt\xa4iaccessorsXF\x81XC\xa2dtypex&/protocol/accessors#FromPublicColumn#1evalueN\xa2ccolBabcposA\x00gpaddingI\x84A\x00A\x00A\x00A\x00eroundA\x00dsizeA\x04gqueryIdEduniv",
				Mode:        DeclarationMode,
				CompiledIOP: comp,
			}
		}(),
		func() struct {
			V           any
			Expected    string
			Mode        Mode
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
				Mode        Mode
				CompiledIOP *wizard.CompiledIOP
			}{
				V:           &fromYs,
				Expected:    "\xa2dtypex%/protocol/column/verifiercol#FromYs#0evalueXx\xa3equeryEdunivfrangesXZ\x84BaaMlSHIFT_2_16_aBabXCxAFROM_ACCESSORS_FROM_COLUMN_POSITION_ACCESSOR_b_0_PADDING=0_SIZE=4eroundA\x00",
				Mode:        ReferenceMode,
				CompiledIOP: comp,
			}
		}(),
		{
			V:           coin.Info{Type: coin.IntegerVec, Size: 16, UpperBound: 16, Name: "foo", Round: 1},
			Expected:    "\xa5dnameDcfooeroundA\x01dsizeA\x10dtypeA\x01jupperBoundA\x10",
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
