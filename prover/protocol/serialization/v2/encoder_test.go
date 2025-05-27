package v2

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
)

func TestCompiled(t *testing.T) {
	comp := serialization.NewEmptyCompiledIOP()
	foo := comp.Columns.AddToRound(0, "foo", 16, column.Committed)
	bar := comp.Columns.AddToRound(1, "bar", 16, column.Committed)
	_ = comp.InsertCoin(1, "coin", coin.IntegerVec, 16, 16)
	coiz := comp.InsertCoin(1, "coiz", coin.Field)

	comp.InsertGlobal(
		1,
		"global",
		func() *symbolic.Expression {
			var (
				foo  = ifaces.ColumnAsVariable(foo)
				bar  = ifaces.ColumnAsVariable(bar)
				coiz = symbolic.NewVariable(coiz)
			)
			return foo.Add(bar).Add(foo).Mul(coiz)
		}(),
	)

	encoded, err := MarshalCompIOP(comp)
	if err != nil {
		t.Fatalf("could not encode: %v", err.Error())
	}

	deSerComp, err := UnmarshalCompIOP(encoded)
	if err != nil {
		t.Fatalf("could not encode: %v", err.Error())
	}

	fmt.Printf("comp: %v \n", comp)
	fmt.Printf("DeserComp: %v \n", deSerComp)

	if !test_utils.CompareExportedFields(comp, deSerComp) {
		t.Errorf("Mismatch in exported fields after RecursedCompiledIOP serde")
	}
}
