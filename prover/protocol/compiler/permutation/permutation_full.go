package permutation

import (
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

// CompileViaGrandProduct scans `comp`, looking for [query.Permutation] queries and
// compiles them using the GrandProduct argument technique. All the queries are
// compiled independently and the technique relies on computing a column Z
// accumulating the fractions (A[i] + Beta) / (B[i] + Beta)
func CompileViaGrandProduct(comp *wizard.CompiledIOP) {
	CompileIntoGdProduct(comp)
	CompileGrandProduct(comp)
}

// dispatchPermutation applies the grand product argument compilation over
// a specific [query.Permutation]
func dispatchPermutation(
	comp *wizard.CompiledIOP,
	zCatalog map[[2]int]*ZCtx,
	round int,
	q query.Permutation,
) (numPub, denPub []*symbolic.Expression) {

	var (
		isMultiColumn = len(q.A[0]) > 1
		alpha         coin.Info
		beta          = comp.InsertCoin(round+1, deriveName[coin.Name](q, "BETA"), coin.Field)
	)

	if isMultiColumn {
		alpha = comp.InsertCoin(round+1, deriveName[coin.Name](q, "ALPHA"), coin.Field)
	}

	for k, aOrB := range [2][][]ifaces.Column{q.A, q.B} {
		for frag := range aOrB {
			var (
				numRow = aOrB[frag][0].Size()
				factor = symbolic.NewVariable(aOrB[frag][0])
			)

			if isMultiColumn {
				factor = wizardutils.RandLinCombColSymbolic(alpha, aOrB[frag])
			}

			factor = symbolic.Add(factor, beta)

			catalogEntry := [2]int{round + 1, numRow}
			if _, ok := zCatalog[catalogEntry]; !ok {
				zCatalog[catalogEntry] = &ZCtx{
					Size:  numRow,
					Round: round + 1,
				}
			}

			ctx := zCatalog[catalogEntry]

			isPublic := column.IsPublicExpression(factor)

			switch {
			case k == 0 && !isPublic:
				ctx.NumeratorFactors = append(ctx.NumeratorFactors, factor)
			case k == 1 && !isPublic:
				ctx.DenominatorFactors = append(ctx.DenominatorFactors, factor)
			case k == 0 && isPublic:
				numPub = append(numPub, factor)
			case k == 1 && isPublic:
				denPub = append(denPub, factor)
			default:
				panic("invalid k")
			}
		}
	}

	return numPub, denPub
}
