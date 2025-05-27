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

func TestCompiledV2(t *testing.T) {
	comp := serialization.NewEmptyCompiledIOP()
	foo := comp.Columns.AddToRound(0, "foo", 16, column.Committed)
	bar := comp.Columns.AddToRound(1, "bar", 16, column.Committed)
	foO := comp.Columns.AddToRound(2, "fooo", 16, column.Committed)
	barr := comp.Columns.AddToRound(3, "barrr", 16, column.Committed)
	_ = comp.InsertCoin(1, "coin", coin.IntegerVec, 16, 16)
	coiz := comp.InsertCoin(1, "coiz", coin.Field)
	coizz := comp.InsertCoin(1, "coizz", coin.Field)
	_ = comp.InsertCoin(1, "coinnn", coin.Field)

	comp.InsertGlobal(
		3,
		"global",
		func() *symbolic.Expression {
			var (
				foo   = ifaces.ColumnAsVariable(foo)
				bar   = ifaces.ColumnAsVariable(bar)
				foO   = ifaces.ColumnAsVariable(foO)
				barr  = ifaces.ColumnAsVariable(barr)
				coiz  = symbolic.NewVariable(coiz)
				coizz = symbolic.NewVariable(coizz)
			)
			return foo.Add(bar).Add(foO).Add(barr).Mul(coiz).Mul(coizz)
		}(),
	)

	encoded, err := serialization.SerializeCompiledIOPV2(comp)
	if err != nil {
		t.Fatalf("could not encode: %v", err.Error())
	}

	deSerComp, err := serialization.DeserializeCompiledIOPV2(encoded)
	if err != nil {
		t.Fatalf("could not encode: %v", err.Error())
	}

	fmt.Printf("comp: %v \n", comp)
	fmt.Printf("DeserComp: %v \n", deSerComp)

	if !test_utils.StrictCompareExportedFields(comp.Coins, deSerComp.Coins) {
		t.Errorf("Mismatch in exported fields after V2 serde")
	}
}
