package stitchsplit_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/stitchsplit"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	P1, P2           ifaces.ColID   = "P1", "P2"
	GLOBAL1, GLOBAL2 ifaces.QueryID = "GLOBAL1", "GLOBAL2"
	LOCAL1           ifaces.QueryID = "LOCAL1"
)

func TestLocalOpening(t *testing.T) {
	testStitcher(t, 8, 16, localOpening(4))
}

func TestStitcherFibo(t *testing.T) {
	testStitcher(t, 16, 32, singlePolyFibo(4))
	testStitcher(t, 4, 8, singlePolyFibo(8))
	testStitcher(t, 8, 16, singlePolyFibo(8))
}

func TestStitcherGlobalWithPeriodicSample(t *testing.T) {
	testStitcher(t, 16, 64, globalWithPeriodicSample(16, 8, 0))
	testStitcher(t, 64, 256, globalWithPeriodicSample(256, 8, 1))
	testStitcher(t, 64, 128, globalWithPeriodicSample(256, 8, 7))
}

func TestStitcherLocalWithPeriodicSample(t *testing.T) {
	testStitcher(t, 32, 64, localWithPeriodicSample(256, 8, 0))
	testStitcher(t, 16, 128, localWithPeriodicSample(256, 8, 1))
	testStitcher(t, 64, 256, localWithPeriodicSample(256, 8, 7))
}

func TestSplitterGlobalWithVerifColAndPerriodic(t *testing.T) {
	testStitcher(t, 8, 64, globalWithVerifColAndPeriodic(8, 4, 0))
	testStitcher(t, 64, 128, globalWithVerifColAndPeriodic(256, 8, 1))
	testStitcher(t, 8, 16, globalWithVerifColAndPeriodic(256, 8, 7))
}

func TestLocalEvalWithStatus(t *testing.T) {
	// Set log level to Info to capture detailed logs during the test
	logrus.SetLevel(logrus.DebugLevel)

	var a, b, c, d ifaces.Column
	var q1, q2, q3, q4, q5, q6, q7, q8, q9, q10, q11, q12 query.LocalOpening

	define := func(builder *wizard.Builder) {
		logrus.Info("Defining columns and local opening queries")
		// Declare columns of different sizes
		a = builder.RegisterCommitExt("A", 2)
		b = builder.RegisterCommitExt("B", 4)
		c = builder.RegisterCommitExt("C", 8)
		d = builder.RegisterCommitExt("D", 16)

		// Local opening at zero
		q1 = builder.LocalOpening("Q00", a)
		q2 = builder.LocalOpening("Q01", b)
		q3 = builder.LocalOpening("Q02", c)
		q4 = builder.LocalOpening("Q03", d)

		// Local opening but shifted by one
		q5 = builder.LocalOpening("Q10", column.Shift(a, 1))
		q6 = builder.LocalOpening("Q11", column.Shift(b, 1))
		q7 = builder.LocalOpening("Q12", column.Shift(c, 1))
		q8 = builder.LocalOpening("Q13", column.Shift(d, 1))

		// Local opening but shifted by negative one
		q9 = builder.LocalOpening("Q20", column.Shift(a, -1))
		q10 = builder.LocalOpening("Q21", column.Shift(b, -1))
		q11 = builder.LocalOpening("Q22", column.Shift(c, -1))
		q12 = builder.LocalOpening("Q23", column.Shift(d, -1))
	}

	logrus.Info("Starting compilation with Stitcher")
	comp := wizard.Compile(define, stitchsplit.Stitcher(4, 8))
	logrus.Info("Compilation completed")

	// After stitching compilation, check column statuses
	logrus.WithFields(logrus.Fields{
		"A": comp.Columns.Status("A").String(),
		"B": comp.Columns.Status("B").String(),
		"C": comp.Columns.Status("C").String(),
		"D": comp.Columns.Status("D").String(),
	}).Info("Column statuses after stitching")

	assert.Equal(t, column.Proof.String(), comp.Columns.Status("A").String())
	assert.Equal(t, column.Ignored.String(), comp.Columns.Status("B").String())
	assert.Equal(t, column.Committed.String(), comp.Columns.Status("C").String())
	assert.Equal(t, column.Committed.String(), comp.Columns.Status("D").String())

	// Check query ignored statuses
	logrus.WithFields(logrus.Fields{
		"Q00": comp.QueriesParams.IsIgnored(q1.ID),
		"Q01": comp.QueriesParams.IsIgnored(q2.ID),
		"Q02": comp.QueriesParams.IsIgnored(q3.ID),
		"Q03": comp.QueriesParams.IsIgnored(q4.ID),
		"Q10": comp.QueriesParams.IsIgnored(q5.ID),
		"Q11": comp.QueriesParams.IsIgnored(q6.ID),
		"Q12": comp.QueriesParams.IsIgnored(q7.ID),
		"Q13": comp.QueriesParams.IsIgnored(q8.ID),
		"Q20": comp.QueriesParams.IsIgnored(q9.ID),
		"Q21": comp.QueriesParams.IsIgnored(q10.ID),
		"Q22": comp.QueriesParams.IsIgnored(q11.ID),
		"Q23": comp.QueriesParams.IsIgnored(q12.ID),
	}).Info("Query ignored statuses after stitching")

	assert.Equal(t, true, comp.QueriesParams.IsIgnored(q1.ID))
	assert.Equal(t, true, comp.QueriesParams.IsIgnored(q2.ID))
	assert.Equal(t, false, comp.QueriesParams.IsIgnored(q3.ID))
	assert.Equal(t, false, comp.QueriesParams.IsIgnored(q4.ID))
	assert.Equal(t, true, comp.QueriesParams.IsIgnored(q5.ID))
	assert.Equal(t, true, comp.QueriesParams.IsIgnored(q6.ID))
	assert.Equal(t, false, comp.QueriesParams.IsIgnored(q7.ID))
	assert.Equal(t, false, comp.QueriesParams.IsIgnored(q8.ID))
	assert.Equal(t, true, comp.QueriesParams.IsIgnored(q9.ID))
	assert.Equal(t, true, comp.QueriesParams.IsIgnored(q10.ID))
	assert.Equal(t, false, comp.QueriesParams.IsIgnored(q11.ID))
	assert.Equal(t, false, comp.QueriesParams.IsIgnored(q12.ID))

	logrus.Info("Manually compiling with dummy compiler")
	dummy.Compile(comp)
	logrus.Info("Dummy compilation completed")

	logrus.Info("Generating proof")
	proof := wizard.Prove(comp, func(assi *wizard.ProverRuntime) {
		logrus.Info("Assigning columns")
		// Assigns all the columns
		assi.AssignColumn(a.GetColID(), smartvectors.ForTest(0, 1))
		logrus.WithField("A", smartvectors.ForTest(0, 1)).Debug("Assigned column A")
		assi.AssignColumn(b.GetColID(), smartvectors.ForTest(2, 3, 4, 5))
		logrus.WithField("B", smartvectors.ForTest(2, 3, 4, 5)).Debug("Assigned column B")
		assi.AssignColumn(c.GetColID(), smartvectors.ForTest(6, 7, 8, 9, 10, 11, 12, 13))
		logrus.WithField("C", smartvectors.ForTest(6, 7, 8, 9, 10, 11, 12, 13)).Debug("Assigned column C")
		assi.AssignColumn(d.GetColID(), smartvectors.ForTest(15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30))
		logrus.WithField("D", smartvectors.ForTest(15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30)).Debug("Assigned column D")

		logrus.Info("Assigning local points")
		// Assign the alleged results
		assi.AssignLocalPointExt("Q00", fext.NewFromInt(0, 0, 0, 0))
		logrus.WithField("Q00", fext.NewFromInt(0, 0, 0, 0)).Debug("Assigned local point Q00")
		assi.AssignLocalPointExt("Q01", fext.NewFromInt(2, 0, 0, 0))
		logrus.WithField("Q01", fext.NewFromInt(2, 0, 0, 0)).Debug("Assigned local point Q01")
		assi.AssignLocalPointExt("Q02", fext.NewFromInt(6, 0, 0, 0))
		logrus.WithField("Q02", fext.NewFromInt(6, 0, 0, 0)).Debug("Assigned local point Q02")
		assi.AssignLocalPointExt("Q03", fext.NewFromInt(15, 0, 0, 0))
		logrus.WithField("Q03", fext.NewFromInt(15, 0, 0, 0)).Debug("Assigned local point Q03")
		assi.AssignLocalPointExt("Q10", fext.NewFromInt(1, 0, 0, 0))
		logrus.WithField("Q10", fext.NewFromInt(1, 0, 0, 0)).Debug("Assigned local point Q10")
		assi.AssignLocalPointExt("Q11", fext.NewFromInt(3, 0, 0, 0))
		logrus.WithField("Q11", fext.NewFromInt(3, 0, 0, 0)).Debug("Assigned local point Q11")
		assi.AssignLocalPointExt("Q12", fext.NewFromInt(7, 0, 0, 0))
		logrus.WithField("Q12", fext.NewFromInt(7, 0, 0, 0)).Debug("Assigned local point Q12")
		assi.AssignLocalPointExt("Q13", fext.NewFromInt(16, 0, 0, 0))
		logrus.WithField("Q13", fext.NewFromInt(16, 0, 0, 0)).Debug("Assigned local point Q13")
		assi.AssignLocalPointExt("Q20", fext.NewFromInt(1, 0, 0, 0))
		logrus.WithField("Q20", fext.NewFromInt(1, 0, 0, 0)).Debug("Assigned local point Q20")
		assi.AssignLocalPointExt("Q21", fext.NewFromInt(5, 0, 0, 0))
		logrus.WithField("Q21", fext.NewFromInt(5, 0, 0, 0)).Debug("Assigned local point Q21")
		assi.AssignLocalPointExt("Q22", fext.NewFromInt(13, 0, 0, 0))
		logrus.WithField("Q22", fext.NewFromInt(13, 0, 0, 0)).Debug("Assigned local point Q22")
		assi.AssignLocalPointExt("Q23", fext.NewFromInt(30, 0, 0, 0))
		logrus.WithField("Q23", fext.NewFromInt(30, 0, 0, 0)).Debug("Assigned local point Q23")
	})
	logrus.Info("Proof generation completed")

	logrus.Info("Verifying proof")
	err := wizard.Verify(comp, proof)
	if err != nil {
		logrus.WithError(err).Error("Proof verification failed")
	} else {
		logrus.Info("Proof verification succeeded")
	}
	require.NoError(t, err)
}

func testStitcher(t *testing.T, minSize, maxSize int, gen func() (wizard.DefineFunc, wizard.MainProverStep)) {

	// Activates the logs for easy debugging
	logrus.SetLevel(logrus.TraceLevel)

	builder, prover := gen()
	comp := wizard.Compile(builder, stitchsplit.Stitcher(minSize, maxSize), dummy.Compile)
	proof := wizard.Prove(comp, prover)
	err := wizard.Verify(comp, proof)

	require.NoError(t, err)

	for _, qName := range comp.QueriesNoParams.AllKeysAt(0) {

		switch q := comp.QueriesNoParams.Data(qName).(type) {

		case query.GlobalConstraint:
			board := q.Expression.Board()
			metadatas := board.ListVariableMetadata()
			metadataNames := []string{}
			for i := range metadatas {
				metadataNames = append(metadataNames, metadatas[i].String())
			}
			t.Logf("query %v - with metadata %v", q.ID, metadataNames)
		case query.LocalConstraint:
			board := q.Expression.Board()
			metadatas := board.ListVariableMetadata()
			metadataNames := []string{}
			for i := range metadatas {
				metadataNames = append(metadataNames, metadatas[i].String())
			}
			t.Logf("query %v - with metadata %v", q.ID, metadataNames)
		}
	}
}

func localOpening(n int) func() (wizard.DefineFunc, wizard.MainProverStep) {
	return func() (wizard.DefineFunc, wizard.MainProverStep) {
		definer := func(build *wizard.Builder) {
			P1 := build.RegisterCommit(P1, n)
			_ = build.LocalOpening("O1", P1)
			_ = build.LocalOpening("O2", column.Shift(P1, 3))
			_ = build.LocalOpening("O3", column.Shift(P1, 4))
			_ = build.LocalOpening("O4", column.Shift(P1, -1))
		}

		prover := func(run *wizard.ProverRuntime) {
			p1_ := make([]field.Element, n)
			for i := range p1_ {
				p1_[i].SetUint64(uint64(i))
			}
			p1 := smartvectors.NewRegular(p1_)
			run.AssignColumn(P1, p1)
			run.AssignLocalPoint("O1", p1.Get(0%n))
			run.AssignLocalPoint("O2", p1.Get(3%n))
			run.AssignLocalPoint("O3", p1.Get(4%n))
			run.AssignLocalPoint("O4", p1.Get(n-1))
		}

		return definer, prover
	}
}

func singlePolyFibo(size int) func() (wizard.DefineFunc, wizard.MainProverStep) {
	return func() (wizard.DefineFunc, wizard.MainProverStep) {
		builder := func(build *wizard.Builder) {
			// Number of rows
			P1 := build.RegisterCommit(P1, size) // overshadows P
			P2 := build.RegisterCommit(P2, size)

			// P(X) = P(X/w) + P(X/w^2)
			expr1 := sym.Sub(
				sym.Add(column.Shift(P1, 1), P1),
				column.Shift(P1, 2))

			expr2 := sym.Sub(
				sym.Add(column.Shift(P2, 1), P2),
				column.Shift(P2, 2))

			_ = build.GlobalConstraint(GLOBAL1, expr1)
			_ = build.GlobalConstraint(GLOBAL2, expr2)
			// 	_ = build.LocalConstraint(LOCAL1, sym.Sub(P1, 1))
		}

		prover := func(run *wizard.ProverRuntime) {
			x := make([]field.Element, size)
			x[0].SetOne()
			x[1].SetOne()
			for i := 2; i < size; i++ {
				x[i].Add(&x[i-1], &x[i-2])
			}
			run.AssignColumn(P1, smartvectors.NewRegular(x))
			run.AssignColumn(P2, smartvectors.NewRegular(x))
		}

		return builder, prover
	}
}

func globalWithPeriodicSample(size, period, offset int) func() (wizard.DefineFunc, wizard.MainProverStep) {
	return func() (wizard.DefineFunc, wizard.MainProverStep) {

		builder := func(build *wizard.Builder) {
			P1 := build.RegisterCommit(P1, size) // overshadows P
			_ = build.GlobalConstraint(GLOBAL1, variables.NewPeriodicSample(period, offset).Mul(ifaces.ColumnAsVariable(P1)))
		}

		prover := func(run *wizard.ProverRuntime) {
			v := vector.Repeat(field.One(), size)
			for i := 0; i < size; i++ {
				if i%period == offset {
					v[i].SetZero()
				}
			}
			run.AssignColumn(P1, smartvectors.NewRegular(v))
		}

		return builder, prover
	}
}

func localWithPeriodicSample(size, period, offset int) func() (wizard.DefineFunc, wizard.MainProverStep) {
	return func() (wizard.DefineFunc, wizard.MainProverStep) {

		builder := func(build *wizard.Builder) {
			P1 := build.RegisterCommit(P1, size) // overshadows P
			_ = build.LocalConstraint(GLOBAL1, variables.NewPeriodicSample(period, offset).Mul(ifaces.ColumnAsVariable(P1)))
		}

		prover := func(run *wizard.ProverRuntime) {
			v := vector.Repeat(field.One(), size)
			for i := 0; i < size; i++ {
				if i%period == offset {
					v[i].SetZero()
				}
			}
			run.AssignColumn(P1, smartvectors.NewRegular(v))
		}

		return builder, prover
	}
}

func globalWithVerifColAndPeriodic(size, period, offset int) func() (wizard.DefineFunc, wizard.MainProverStep) {
	return func() (wizard.DefineFunc, wizard.MainProverStep) {

		builder := func(build *wizard.Builder) {
			P1 := build.RegisterCommit(P1, size)
			verifcol1 := verifiercol.NewFromAccessors(genAccessors(0, size), fext.Zero(), size)
			verifcol2 := verifiercol.NewFromAccessors(genAccessors(2, size), fext.Zero(), size)
			_ = build.GlobalConstraint(LOCAL1,
				sym.Sub(

					sym.Mul(sym.Sub(1, P1),
						verifcol2),

					sym.Mul(variables.NewPeriodicSample(period, offset),
						sym.Add(2, verifcol1))),
			)
		}

		prover := func(run *wizard.ProverRuntime) {
			v := vector.Repeat(field.One(), size)
			for i := 0; i < size; i++ {
				if i%period == offset {
					v[i].SetZero()
				}
			}
			run.AssignColumn(P1, smartvectors.NewRegular(v))
		}

		return builder, prover
	}
}

func genAccessors(start, size int) (res []ifaces.Accessor) {
	for i := start; i < size+start; i++ {
		t := accessors.NewConstant(field.NewElement(uint64(i)))
		res = append(res, t)
	}
	return res
}
