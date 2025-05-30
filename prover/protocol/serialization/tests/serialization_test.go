package serialization_test

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
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
		{
			V: field.NewElement(0),
		},
		{
			V: field.NewElement(1),
		},
		{
			V: func() any {
				v, err := new(field.Element).SetString("0x00ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
				if err != nil {
					utils.Panic("field does not work: %v", err)
				}
				return v
			}(),
		},
		{
			V: vector.ForTest(0, 1, 2, 3, 4, 5, 5, 6, 7),
		},
	}

	for i := range testCases {
		t.Run(fmt.Sprintf("test-case-%v", i), func(t *testing.T) {

			msg, err := serialization.Serialize(testCases[i].V)
			require.NoError(t, err)

			fmt.Printf("testcase=%v, msg=%v\n", i, string(msg))

			unmarshaled := reflect.New(reflect.TypeOf(testCases[i].V)).Interface()
			err = serialization.Deserialize(msg, unmarshaled)
			require.NoError(t, err)

			unmarshalledDereferenced := reflect.ValueOf(unmarshaled).Elem().Interface()
			if !test_utils.CompareExportedFields(testCases[i].V, unmarshalledDereferenced) {
				t.Errorf("Mismatch in exported fields after full serde value")
			}

		})
	}
}

type Team struct {
	Name      string
	CreatedAt time.Time
	Members   []*Person
	Metadata  map[string]interface{}
	privateID string // unexported field
}

type Person struct {
	Name       string
	Age        int
	Attributes Attributes
}

type Attributes struct {
	Nickname string
	Score    int
	private  string // unexported field
}

func TestSerdeSampleStruct(t *testing.T) {
	p1 := &Person{
		Name: "Alice",
		Age:  28,
		Attributes: Attributes{
			Nickname: "Ace",
			Score:    95,
			private:  "secret-1",
		},
	}
	p2 := &Person{
		Name: "Bob",
		Age:  35,
		Attributes: Attributes{
			Nickname: "Builder",
			Score:    88,
			private:  "secret-2",
		},
	}

	team := Team{
		Name:      "DevTeam",
		CreatedAt: time.Now().Truncate(time.Second),
		Members:   []*Person{p1, p2},
		Metadata: map[string]interface{}{
			"department": "Engineering",
			"active":     true,
			"head":       p1.Name,
		},
		privateID: "internal-uuid-1234",
	}

	// Serialize
	teamBytes, err := serialization.Serialize(team)
	require.NoError(t, err)

	// Deserialize
	var deserializedTeam Team
	err = serialization.Deserialize(teamBytes, &deserializedTeam)
	require.NoError(t, err)

	if !test_utils.CompareExportedFields(team, deserializedTeam) {
		t.Errorf("expected team and deserializedTeam to be equal")
	}
}
