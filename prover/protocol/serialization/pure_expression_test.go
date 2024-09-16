package serialization

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/column"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
)

func TestPureExpr(t *testing.T) {

	var (
		stringA      = sym.NewDummyVar("a")
		stringB      = sym.NewDummyVar("b")
		comp         = newEmptyCompiledIOP()
		colA         = comp.InsertColumn(0, "a", 16, column.Committed)
		colB         = comp.InsertColumn(0, "b", 16, column.Committed)
		exprBuilders = []func(a, b any) *sym.Expression{
			func(a, b any) *sym.Expression {
				return sym.Mul(a, b)
			},
			func(a, b any) *sym.Expression {
				return sym.Mul(a, b, 1)
			},
			func(a, b any) *sym.Expression {
				return sym.Add(a, b)
			},
			func(a, b any) *sym.Expression {
				return sym.Add(a, b, 1)
			},
			func(a, b any) *sym.Expression {
				return sym.Sub(a, b)
			},
			func(a, b any) *sym.Expression {
				return sym.Sub(a, b, 1)
			},
		}
	)

	t.Run("without-comp", func(t *testing.T) {
		for i := range exprBuilders {
			t.Run(fmt.Sprintf("expression-%v", i), func(t *testing.T) {

				var (
					e       = exprBuilders[i](stringA, stringB)
					buf     = &bytes.Buffer{}
					_, errW = MarshalExprCBOR(buf, e)
				)

				if errW != nil {
					t.Fatalf("failed writing the expression: %v", errW.Error())
				}

				_ = UnmarshalExprCBOR(buf)
			})
		}
	})

	t.Run("with-comp", func(t *testing.T) {
		for i := range exprBuilders {
			t.Run(fmt.Sprintf("expression-%v", i), func(t *testing.T) {

				var (
					e       = exprBuilders[i](colA, colB)
					buf     = &bytes.Buffer{}
					_, errW = MarshalExprCBOR(buf, e)
				)

				if errW != nil {
					t.Fatalf("failed writing the expression: %v", errW.Error())
				}

				_ = UnmarshalExprCBOR(buf)
			})
		}
	})
}
