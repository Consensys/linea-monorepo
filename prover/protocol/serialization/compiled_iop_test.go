package serialization

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

func TestCompiled(t *testing.T) {
	comp := newEmptyCompiledIOP()
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
			return foo.Add(bar).Mul(coiz)
		}(),
	)

	encoded, err := SerializeCompiledIOP(comp)

	if err != nil {
		t.Fatalf("could not encode: %v", err.Error())
	}

	_, err = DeserializeCompiledIOP(encoded)

	if err != nil {
		t.Fatalf("could not encode: %v", err.Error())
	}

}
