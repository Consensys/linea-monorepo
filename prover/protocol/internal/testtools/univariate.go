package testtools

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// UnivariateTestcase represents a univariate relationship and and its assignment.
// The instances are used to generate testcases.
type UnivariateTestcase struct {
	NameStr string
	Polys   []smartvectors.SmartVector
	QueryXs []field.Element
	// QueryPols[i] indexes the polynomials in [Polys] that are queried. In the i-th query.
	QueryPols [][]int
	// This parameter is optionally set. If not set, the test case computes the
	// correct value.
	QueryYs [][]field.Element
	// IsPrecomputed of poly optionally control which polys are precomputed.
	IsPrecomputed []bool
}

// ListOfUnivariateTestcasesPositive lists standard univariate testcases
// that are supposed to pass.
var ListOfUnivariateTestcasesPositive = []*UnivariateTestcase{
	{
		NameStr: "constant-poly",
		Polys: []smartvectors.SmartVector{
			smartvectors.ForTest(10, 10, 10, 10, 10, 10, 10, 10),
		},
		QueryXs: []field.Element{
			field.PseudoRand(rng),
		},
		QueryPols: [][]int{
			{0},
		},
	},
	{
		NameStr: "one-poly-one-point-non-simple-values",
		Polys: []smartvectors.SmartVector{
			smartvectors.ForTest(0, 0, 0, 0, 0, 1, 0, 0),
		},
		QueryXs: []field.Element{
			field.NewElement(0),
		},
		QueryPols: [][]int{
			{0},
		},
	},
	{
		NameStr: "one-poly-one-point",
		Polys: []smartvectors.SmartVector{
			RandomVec(8),
		},
		QueryXs: []field.Element{
			field.NewElement(0),
		},
		QueryPols: [][]int{
			{0},
		},
	},
	{
		NameStr: "two-poly-one-point-same-simple-values",
		Polys: []smartvectors.SmartVector{
			smartvectors.ForTest(0, 0, 0, 0, 0, 1, 0, 0),
			smartvectors.ForTest(0, 0, 0, 0, 0, 1, 0, 0),
		},
		QueryXs: []field.Element{
			field.Zero(),
		},
		QueryPols: [][]int{
			{0, 1},
		},
	},
	{
		NameStr: "two-poly-one-point-same-simple-values-one-precomputed",
		Polys: []smartvectors.SmartVector{
			smartvectors.ForTest(0, 0, 0, 0, 0, 1, 0, 0),
			smartvectors.ForTest(0, 0, 0, 0, 0, 1, 0, 0),
		},
		QueryXs: []field.Element{
			field.Zero(),
		},
		QueryPols: [][]int{
			{0, 1},
		},
		IsPrecomputed: []bool{
			true,
			false,
		},
	},
	{
		NameStr: "two-poly-one-point-different-sizes-simple-values",
		Polys: []smartvectors.SmartVector{
			smartvectors.ForTest(0, 0, 0, 0, 0, 1, 0, 0),
			smartvectors.ForTest(0, 1, 0, 0),
		},
		QueryXs: []field.Element{
			field.Zero(),
		},
		QueryPols: [][]int{
			{0, 1},
		},
	},
	{
		NameStr: "two-poly-one-point",
		Polys: []smartvectors.SmartVector{
			RandomVec(8),
			RandomVec(8),
		},
		QueryXs: []field.Element{
			field.PseudoRand(rng),
		},
		QueryPols: [][]int{
			{0, 1},
		},
	},
	{
		NameStr: "one-poly-two-points",
		Polys: []smartvectors.SmartVector{
			RandomVec(8),
		},
		QueryXs: []field.Element{
			field.PseudoRand(rng),
			field.PseudoRand(rng),
		},
		QueryPols: [][]int{
			{0},
			{0},
		},
	},
	{
		NameStr: "two-poly-two-points",
		Polys: []smartvectors.SmartVector{
			RandomVec(8),
			RandomVec(8),
		},
		QueryXs: []field.Element{
			field.PseudoRand(rng),
			field.PseudoRand(rng),
		},
		QueryPols: [][]int{
			{0, 1},
			{0, 1},
		},
	},
}

func (u *UnivariateTestcase) Name() string {
	return u.NameStr
}

func (u *UnivariateTestcase) MustFail() bool { return false }

func (u *UnivariateTestcase) Define(comp *wizard.CompiledIOP) {

	polys := make([]ifaces.Column, len(u.Polys))
	for i := range polys {

		name := formatName[ifaces.ColID]("Univariate", u.NameStr, "Poly", i)

		if len(u.IsPrecomputed) > i && u.IsPrecomputed[i] {
			polys[i] = comp.InsertPrecomputed(name, u.Polys[i])
			continue
		}

		polys[i] = comp.InsertCommit(0, name, u.Polys[i].Len())
	}

	for i := range u.QueryXs {

		queryPols := make([]ifaces.Column, len(u.QueryPols[i]))
		for j := range queryPols {
			queryPols[j] = polys[u.QueryPols[i][j]]
		}

		comp.InsertUnivariate(
			0,
			formatName[ifaces.QueryID]("Univariate", u.NameStr, "Query", i),
			queryPols,
		)
	}
}

func (u *UnivariateTestcase) Assign(run *wizard.ProverRuntime) {

	for i := range u.Polys {

		if len(u.IsPrecomputed) > i && u.IsPrecomputed[i] {
			continue
		}

		run.AssignColumn(
			formatName[ifaces.ColID]("Univariate", u.NameStr, "Poly", i),
			u.Polys[i],
			wizard.DisableAssignmentSizeReduction,
		)
	}

	for i := range u.QueryXs {

		var (
			ys []field.Element
			x  = u.QueryXs[i]
		)

		if len(u.QueryYs) > 0 {
			ys = u.QueryYs[i]
		} else {
			ys = make([]field.Element, len(u.QueryPols[i]))
			for j := range ys {
				p := u.Polys[u.QueryPols[i][j]]
				ys[j] = smartvectors.Interpolate(p, x)
			}
		}

		run.AssignUnivariate(
			formatName[ifaces.QueryID]("Univariate", u.NameStr, "Query", i),
			u.QueryXs[i],
			ys...,
		)
	}
}
