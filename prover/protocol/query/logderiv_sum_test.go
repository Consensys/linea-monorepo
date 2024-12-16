package query_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/stretchr/testify/require"
)

func TestLogDerivSum(t *testing.T) {

	define := func(b *wizard.Builder) {
		var (
			comp = b.CompiledIOP
		)

		p0 := b.RegisterCommit("Num_0", 4)
		p1 := b.RegisterCommit("Num_1", 4)
		p2 := b.RegisterCommit("Num_2", 4)

		q0 := b.RegisterCommit("Den_0", 4)
		q1 := b.RegisterCommit("Den_1", 4)
		q2 := b.RegisterCommit("Den_2", 4)

		numerators := []*symbolic.Expression{
			symbolic.Mul(p0, -1),
			ifaces.ColumnAsVariable(p1),
			symbolic.Mul(p2, p0, 2),
		}

		denominators := []*symbolic.Expression{
			ifaces.ColumnAsVariable(q0),
			ifaces.ColumnAsVariable(q1),
			ifaces.ColumnAsVariable(q2),
		}

		key := [2]int{0, 0}
		zCat1 := map[[2]int]*query.LogDerivativeSumInput{}
		zCat1[key] = &query.LogDerivativeSumInput{
			Numerator:   numerators,
			Denominator: denominators,
		}
		comp.InsertLogDerivativeSum(0, "LogDerivSum_Test", zCat1)

	}

	prover := func(run *wizard.ProverRuntime) {

		run.AssignColumn("Num_0", smartvectors.ForTest(1, 1, 1, 1))
		run.AssignColumn("Num_1", smartvectors.ForTest(2, 3, 7, 9))
		run.AssignColumn("Num_2", smartvectors.ForTest(5, 6, 1, 1))

		run.AssignColumn("Den_0", smartvectors.ForTest(1, 1, 1, 1))
		run.AssignColumn("Den_1", smartvectors.ForTest(2, 3, 7, 9))
		run.AssignColumn("Den_2", smartvectors.ForTest(5, 6, 1, 1))

		run.AssignLogDerivSum("LogDerivSum_Test", field.NewElement(8))

	}

	compiled := wizard.Compile(define, dummy.Compile)
	proof := wizard.Prove(compiled, prover)
	valid := wizard.Verify(compiled, proof)
	require.NoError(t, valid)
}
