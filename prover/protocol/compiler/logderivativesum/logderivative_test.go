package logderivativesum_test

import (
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	logderiv "github.com/consensys/linea-monorepo/prover/protocol/compiler/logderivativesum"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/stretchr/testify/require"
)

// It tests that the given expression for the LogDerivativeSum adds up to the given parameter.
func TestLogDerivativeSum(t *testing.T) {

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

		size := 4
		zCat1 := map[int]*query.LogDerivativeSumInput{}
		zCat1[size] = &query.LogDerivativeSumInput{
			Size:        size,
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

	compiled := wizard.Compile(define, logderiv.CompileLogDerivativeSum, dummy.Compile)
	proof := wizard.Prove(compiled, prover)
	valid := wizard.Verify(compiled, proof)
	require.NoError(t, valid)
}

func TestLogDerivativeSumMixed(t *testing.T) {

	define := func(b *wizard.Builder) {
		var (
			comp = b.CompiledIOP
		)

		p0 := b.RegisterCommit("Num_0", 4)
		p1 := b.RegisterCommit("Num_1", 4)
		p2 := b.RegisterCommit("Num_2", 4)
		p3 := b.RegisterCommit("Num_3", 4)

		q0 := b.RegisterCommit("Den_0", 4)
		q1 := b.RegisterCommit("Den_1", 4)
		q2 := b.RegisterCommit("Den_2", 4)
		q3 := b.RegisterCommit("Den_3", 4)

		numerators := []*symbolic.Expression{
			ifaces.ColumnAsVariable(p0),
			symbolic.Mul(p0, p1),
			symbolic.Sub(p2, symbolic.Mul(p0, 4), symbolic.Mul(p0, p1)),
			ifaces.ColumnAsVariable(p3),
		}

		denominators := []*symbolic.Expression{
			ifaces.ColumnAsVariable(q0),
			ifaces.ColumnAsVariable(q1),
			ifaces.ColumnAsVariable(q2),
			ifaces.ColumnAsVariable(q3),
		}

		size := 4
		zCat1 := map[int]*query.LogDerivativeSumInput{}
		zCat1[size] = &query.LogDerivativeSumInput{
			Size:        size,
			Numerator:   numerators,
			Denominator: denominators,
		}
		comp.InsertLogDerivativeSum(0, "LogDerivSum_Test", zCat1)

	}

	prover := func(run *wizard.ProverRuntime) {

		run.AssignColumn("Num_0", smartvectors.ForTest(1, 2, 3, 4))
		run.AssignColumn("Num_1", smartvectors.ForTest(2, 3, 7, 9))
		run.AssignColumn("Num_2", smartvectors.ForTest(3, 6, 9, 12))
		run.AssignColumn("Num_3", smartvectors.ForTest(1, 1, -1, -1))

		run.AssignColumn("Den_0", smartvectors.ForTestFromPairs(1, 2, 3, 4, 5, 6, 7, 8))
		run.AssignColumn("Den_1", smartvectors.ForTestFromPairs(1, 2, 3, 4, 5, 6, 7, 8))
		run.AssignColumn("Den_2", smartvectors.ForTestFromPairs(1, 2, 3, 4, 5, 6, 7, 8))
		run.AssignColumn("Den_3", smartvectors.ForTest(3, 4, 3, 4))

		expectedResult := field.NewElement(0)
		run.AssignLogDerivSumGeneric("LogDerivSum_Test", *fext.NewESHashFromBase(&expectedResult))

	}

	compiled := wizard.Compile(define, logderiv.CompileLogDerivativeSum, dummy.Compile)
	proof := wizard.Prove(compiled, prover)
	valid := wizard.Verify(compiled, proof)
	require.NoError(t, valid)
}

func TestLogDerivativeSumMixed2(t *testing.T) {

	define := func(b *wizard.Builder) {
		var (
			comp = b.CompiledIOP
		)

		p0 := b.RegisterCommit("Num_0", 2)
		p1 := b.RegisterCommit("Num_1", 2)
		p2 := b.RegisterCommit("Num_2", 2)

		q0 := b.RegisterCommit("Den_0", 2)
		q1 := b.RegisterCommit("Den_1", 2)
		q2 := b.RegisterCommit("Den_2", 2)

		numerators := []*symbolic.Expression{
			symbolic.Mul(p0, -1),        // -p0    -2
			ifaces.ColumnAsVariable(p1), // p1      + 2
			symbolic.Mul(p2, p0, 2),     // p2*p0*2			will lead to 2+2
		}

		denominators := []*symbolic.Expression{
			ifaces.ColumnAsVariable(q0),
			ifaces.ColumnAsVariable(q1),
			ifaces.ColumnAsVariable(q2),
		}

		size := 2
		zCat1 := map[int]*query.LogDerivativeSumInput{}
		zCat1[size] = &query.LogDerivativeSumInput{
			Size:        size,
			Numerator:   numerators,
			Denominator: denominators,
		}
		comp.InsertLogDerivativeSum(0, "LogDerivSum_Test", zCat1)

	}

	prover := func(run *wizard.ProverRuntime) {

		run.AssignColumn("Num_0", smartvectors.ForTest(1, 1))
		run.AssignColumn("Num_1", smartvectors.ForTest(2, 3))
		run.AssignColumn("Num_2", smartvectors.ForTest(5, 6))

		run.AssignColumn("Den_0", smartvectors.ForTestFromPairs(1, 0, 1, 0))
		run.AssignColumn("Den_1", smartvectors.ForTestFromPairs(2, 0, 3, 0))
		run.AssignColumn("Den_2", smartvectors.ForTestFromPairs(5, 0, 6, 0))

		expectedResult := field.NewElement(4)
		run.AssignLogDerivSumGeneric("LogDerivSum_Test", *fext.NewESHashFromBase(&expectedResult))

	}

	compiled := wizard.Compile(define, logderiv.CompileLogDerivativeSum, dummy.Compile)
	proof := wizard.Prove(compiled, prover)
	valid := wizard.Verify(compiled, proof)
	require.NoError(t, valid)
}
