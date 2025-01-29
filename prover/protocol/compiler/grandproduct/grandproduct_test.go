package grandproduct_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/grandproduct"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/stretchr/testify/require"
)

// It tests that the given expressions for the GrandProduct quotients out to the given parameter.
func TestGrandProduct(t *testing.T) {

	define := func(b *wizard.Builder) {
		var (
			comp = b.CompiledIOP

			n0 = b.RegisterCommit("Num_0", 2)
			n1 = b.RegisterCommit("Num_1", 4)
			n2 = b.RegisterCommit("Num_2", 4)

			d0 = b.RegisterCommit("Den_0", 2)
			d1 = b.RegisterCommit("Den_1", 4)
			d2 = b.RegisterCommit("Den_2", 4)
		)

		numerators := []*symbolic.Expression{
			symbolic.Mul(n0, 2),
			ifaces.ColumnAsVariable(n1),
			symbolic.Add(n1, n2),
		}

		denominators := []*symbolic.Expression{
			symbolic.Mul(d0, 2),
			ifaces.ColumnAsVariable(d1),
			symbolic.Add(d1, d2),
		}
		var (
			size1 = 2
			size2 = 4
			zCat  = map[int]*query.GrandProductInput{}
		)
		zCat[size1] = &query.GrandProductInput{
			Size:         size1,
			Numerators:   numerators[:1],
			Denominators: denominators[:1],
		}
		zCat[size2] = &query.GrandProductInput{
			Size:         size2,
			Numerators:   numerators[1:],
			Denominators: denominators[1:],
		}
		comp.InsertGrandProduct(0, "GrandProduct_Test", zCat)
	}

	prover := func(run *wizard.ProverRuntime) {

		run.AssignColumn("Num_0", smartvectors.ForTest(2, 4))
		run.AssignColumn("Num_1", smartvectors.ForTest(1, 2, 3, 4))
		run.AssignColumn("Num_2", smartvectors.ForTest(1, 2, 3, 4))

		run.AssignColumn("Den_0", smartvectors.ForTest(1, 2))
		run.AssignColumn("Den_1", smartvectors.ForTest(1, 2, 3, 4))
		run.AssignColumn("Den_2", smartvectors.ForTest(1, 2, 3, 4))

		run.AssignGrandProduct("GrandProduct_Test", field.NewElement(4))

	}

	compiled := wizard.Compile(define, grandproduct.CompileGrandProductDist, dummy.CompileAtProverLvl)
	proof := wizard.Prove(compiled, prover)
	valid := wizard.Verify(compiled, proof)
	require.NoError(t, valid)
}
