package testing

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// PermutationTestcase represents a permutation relationship and and its assignment.
// The instances are used to generate testcases.
type PermutationTestcase struct {
	Name         string
	A            [][]smartvectors.SmartVector
	B            [][]smartvectors.SmartVector
	MustFailFlag bool
	Q            query.Permutation
}

// ListOfPermutationTestcasePositive lists standard permutation testcases
// that are supposed to pass.
var ListOfPermutationTestcasePositive = []*PermutationTestcase{

	{
		Name: "positive/1234",
		A: [][]smartvectors.SmartVector{
			{
				smartvectors.ForTest(1, 2, 3, 4),
			},
		},
		B: [][]smartvectors.SmartVector{
			{
				smartvectors.ForTest(4, 2, 3, 1),
			},
		},
	},

	{
		Name: "positive/1234-multi-column",
		A: [][]smartvectors.SmartVector{
			{
				smartvectors.ForTest(1, 2, 3, 4),
				smartvectors.ForTest(5, 6, 7, 8),
			},
		},
		B: [][]smartvectors.SmartVector{
			{
				smartvectors.ForTest(4, 2, 3, 1),
				smartvectors.ForTest(8, 6, 7, 5),
			},
		},
	},

	{
		Name: "positive/1234-split",
		A: [][]smartvectors.SmartVector{
			{
				smartvectors.ForTest(1, 2, 3, 4),
				smartvectors.ForTest(5, 6, 7, 8),
			},
		},
		B: [][]smartvectors.SmartVector{
			{
				smartvectors.ForTest(4, 2, 3, 1, 8, 6, 7, 5),
			},
		},
	},
}

// ListOfPermutationTestcasePositive lists standard permutation testcases
// that are supposed to pass.
var ListOfPermutationTestcaseNegative = []*PermutationTestcase{

	{
		Name: "negative/1234/missing-1",
		A: [][]smartvectors.SmartVector{
			{
				smartvectors.ForTest(2, 2, 3, 4),
			},
		},
		B: [][]smartvectors.SmartVector{
			{
				smartvectors.ForTest(4, 2, 3, 1),
			},
		},
		MustFailFlag: true,
	},

	{
		Name: "negative/1234-multi-column/missing-first-row",
		A: [][]smartvectors.SmartVector{
			{
				smartvectors.ForTest(2, 2, 3, 4),
				smartvectors.ForTest(6, 6, 7, 8),
			},
		},
		B: [][]smartvectors.SmartVector{
			{
				smartvectors.ForTest(4, 2, 3, 1),
				smartvectors.ForTest(8, 6, 7, 5),
			},
		},
		MustFailFlag: true,
	},

	{
		Name: "negative/1234-split/missing-first-pos",
		A: [][]smartvectors.SmartVector{
			{
				smartvectors.ForTest(1, 2, 3, 4),
				smartvectors.ForTest(5, 6, 7, 8),
			},
		},
		B: [][]smartvectors.SmartVector{
			{
				smartvectors.ForTest(2, 2, 3, 1, 8, 6, 7, 5),
			},
		},
		MustFailFlag: true,
	},
}

func (p *PermutationTestcase) Define(comp *wizard.CompiledIOP) {

	var (
		a = make([][]ifaces.Column, len(p.A))
		b = make([][]ifaces.Column, len(p.B))
	)

	for i := range p.A {
		a[i] = make([]ifaces.Column, len(p.A[i]))
		for j := range p.A[i] {
			a[i][j] = comp.InsertCommit(
				0,
				formatName[ifaces.ColID]("Permutation", p.Name, "A", i, j),
				p.A[i][j].Len(),
			)
		}
	}

	for i := range p.B {
		b[i] = make([]ifaces.Column, len(p.B[i]))
		for j := range p.B[i] {
			b[i][j] = comp.InsertCommit(
				0,
				formatName[ifaces.ColID]("Permutation", p.Name, "B", i, j),
				p.B[i][j].Len(),
			)
		}
	}

	p.Q = query.Permutation{
		A:  a,
		B:  b,
		ID: formatName[ifaces.QueryID]("Permutation", p.Name),
	}

	comp.QueriesNoParams.AddToRound(0, p.Q.ID, p.Q)
}

func (p *PermutationTestcase) Assign(run *wizard.ProverRuntime) {

	for i := range p.A {
		for j := range p.A[i] {
			run.AssignColumn(
				formatName[ifaces.ColID]("Permutation", p.Name, "A", i, j),
				p.A[i][j],
			)
		}
	}

	for i := range p.B {
		for j := range p.B[i] {
			run.AssignColumn(
				formatName[ifaces.ColID]("Permutation", p.Name, "B", i, j),
				p.B[i][j],
			)
		}
	}
}
