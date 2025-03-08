package testtools

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// ProjectionTestcase represents a test-case for a projection query
type ProjectionTestcase struct {
	NameStr          string
	FilterA, FilterB smartvectors.SmartVector
	As, Bs           []smartvectors.SmartVector
	ShouldFail       bool
}

var ListOfProjectionTestcasePositive = []*ProjectionTestcase{

	{
		NameStr: "positive/selector-full-zeroes",
		FilterA: smartvectors.NewConstant(field.Zero(), 16),
		FilterB: smartvectors.NewConstant(field.Zero(), 8),
		As: []smartvectors.SmartVector{
			smartvectors.PseudoRand(rng, 16),
		},
		Bs: []smartvectors.SmartVector{
			smartvectors.PseudoRand(rng, 8),
		},
	},

	{
		NameStr: "positive/counting-values",
		FilterA: OnesAt(16, []int{2, 4, 6, 8, 10}),
		FilterB: OnesAt(8, []int{1, 2, 3, 4, 5}),
		As: []smartvectors.SmartVector{
			CountingAt(16, 0, []int{2, 4, 6, 8, 10}),
		},
		Bs: []smartvectors.SmartVector{
			CountingAt(8, 0, []int{1, 2, 3, 4, 5}),
		},
	},
	{
		NameStr: "positive/selector-full-zeroes-multicolumn",
		FilterA: smartvectors.NewConstant(field.Zero(), 16),
		FilterB: smartvectors.NewConstant(field.Zero(), 8),
		As: []smartvectors.SmartVector{
			smartvectors.PseudoRand(rng, 16),
			smartvectors.PseudoRand(rng, 16),
		},
		Bs: []smartvectors.SmartVector{
			smartvectors.PseudoRand(rng, 8),
			smartvectors.PseudoRand(rng, 8),
		},
	},

	{
		NameStr: "positive/counting-values-multicolumn",
		FilterA: OnesAt(16, []int{2, 4, 6, 8, 10}),
		FilterB: OnesAt(8, []int{1, 2, 3, 4, 5}),
		As: []smartvectors.SmartVector{
			CountingAt(16, 0, []int{2, 4, 6, 8, 10}),
			CountingAt(16, 5, []int{2, 4, 6, 8, 10}),
		},
		Bs: []smartvectors.SmartVector{
			CountingAt(8, 0, []int{1, 2, 3, 4, 5}),
			CountingAt(8, 5, []int{1, 2, 3, 4, 5}),
		},
	},
}

var ListOfProjectionTestcaseNegative = []*ProjectionTestcase{

	{
		NameStr: "negative/full-random-with-full-ones-selectors",
		FilterA: smartvectors.NewConstant(field.One(), 16),
		FilterB: smartvectors.NewConstant(field.One(), 8),
		As: []smartvectors.SmartVector{
			smartvectors.PseudoRand(rng, 16),
		},
		Bs: []smartvectors.SmartVector{
			smartvectors.PseudoRand(rng, 8),
		},
		ShouldFail: true,
	},

	{
		NameStr: "negative/counting-too-many",
		FilterA: OnesAt(16, []int{2, 4, 6, 8, 10}),
		FilterB: OnesAt(8, []int{1, 2, 3, 4, 5, 6}),
		As: []smartvectors.SmartVector{
			CountingAt(16, 0, []int{2, 4, 6, 8, 10}),
		},
		Bs: []smartvectors.SmartVector{
			CountingAt(8, 0, []int{1, 2, 3, 4, 5, 6}),
		},
		ShouldFail: true,
	},

	{
		NameStr: "negative/counting-misaligned",
		FilterA: OnesAt(16, []int{2, 4, 6, 8, 10}),
		FilterB: OnesAt(8, []int{1, 2, 3, 4, 5}),
		As: []smartvectors.SmartVector{
			CountingAt(16, 0, []int{2, 4, 6, 8, 10}),
		},
		Bs: []smartvectors.SmartVector{
			CountingAt(8, 0, []int{1, 2, 3, 4, 6}),
		},
		ShouldFail: true,
	},
}

// Define returns a [wizard.DefineFunc] constructing
// columns and a [query.Projection] as specified by the testcase.
func (tc *ProjectionTestcase) Define(comp *wizard.CompiledIOP) {

	inp := query.ProjectionInput{

		FilterA: comp.InsertCommit(
			0,
			formatName[ifaces.ColID]("Projection", tc.Name, "filterA"),
			tc.FilterA.Len(),
		),

		FilterB: comp.InsertCommit(
			0,
			formatName[ifaces.ColID]("Projection", tc.Name, "filterB"),
			tc.FilterB.Len(),
		),

		ColumnA: make([]ifaces.Column, len(tc.As)),
		ColumnB: make([]ifaces.Column, len(tc.Bs)),
	}

	for i := range inp.ColumnA {
		inp.ColumnA[i] = comp.InsertCommit(
			0,
			formatName[ifaces.ColID]("Projection", tc.Name, "A", i),
			tc.As[i].Len(),
		)
	}

	for i := range inp.ColumnB {
		inp.ColumnB[i] = comp.InsertCommit(
			0,
			formatName[ifaces.ColID]("Projection", tc.Name, "B", i),
			tc.Bs[i].Len(),
		)
	}

	comp.InsertProjection(
		formatName[ifaces.QueryID]("Projection", tc.Name, "Query"),
		inp,
	)
}

// proverTestcaseWithProjection returns a prover function assigning the
// columns taking place in the [query.Projection] query.
func (tc *ProjectionTestcase) Assign(run *wizard.ProverRuntime) {

	run.AssignColumn(
		formatName[ifaces.ColID]("Projection", tc.Name, "filterA"),
		tc.FilterA,
	)

	run.AssignColumn(
		formatName[ifaces.ColID]("Projection", tc.Name, "filterB"),
		tc.FilterB,
	)

	for i := range tc.As {
		run.AssignColumn(
			formatName[ifaces.ColID]("Projection", tc.Name, "A", i),
			tc.As[i],
		)
	}

	for i := range tc.Bs {
		run.AssignColumn(
			formatName[ifaces.ColID]("Projection", tc.Name, "B", i),
			tc.Bs[i],
		)
	}
}

func (tc *ProjectionTestcase) MustFail() bool {
	return tc.ShouldFail
}

func (tc *ProjectionTestcase) Name() string {
	return tc.NameStr
}
