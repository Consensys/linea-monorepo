package testtools

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
)

// HornerTestcase specifies a protocol with a Horner query defined in it
// and the assignment of its columns (which may be valid or not).
type HornerTestcase struct {

	// NameStr is the name of the testcase
	NameStr string

	// SignNegativeParts indicates that parts #i has a reversed sign
	SignNegativeParts []bool

	// Coefficients are the coefficients for each parts of the query
	Coefficients []smartvectors.SmartVector

	// Selectors are the selectors for each parts of the query
	Selectors []smartvectors.SmartVector

	// N0s are the N0 values for each parts of the query
	N0s []int

	// N1s are the N1 values for each parts of the query
	N1s []int

	// Xs are the X values for each parts of the query
	Xs []field.Element

	// FinalResult is the expected final result of the query
	FinalResult field.Element

	// Query is the Horner query
	Query query.Horner

	// MustFailFlag indicates that the testcase must fail
	MustFailFlag bool
}

var ListOfHornerTestcasePositive = []*HornerTestcase{

	{
		NameStr:           "positive/none-selected-single",
		SignNegativeParts: []bool{false},
		Coefficients: []smartvectors.SmartVector{
			smartvectors.NewConstant(field.One(), 8),
		},
		Selectors: []smartvectors.SmartVector{
			smartvectors.NewConstant(field.Zero(), 8),
		},
		N0s:         []int{0},
		N1s:         []int{0},
		Xs:          []field.Element{field.One()},
		FinalResult: field.Zero(),
	},

	{
		NameStr:           "positive/two-parts-cancelling",
		SignNegativeParts: []bool{false, true},
		Coefficients: []smartvectors.SmartVector{
			RandomFromSeed(8, 1),
			RandomFromSeed(8, 1),
		},
		Selectors: []smartvectors.SmartVector{
			smartvectors.NewConstant(field.One(), 8),
			smartvectors.NewConstant(field.One(), 8),
		},
		N0s:         []int{0, 0},
		N1s:         []int{8, 8},
		Xs:          []field.Element{field.One(), field.One()},
		FinalResult: field.Zero(),
	},

	{
		NameStr:           "positive/just-counting",
		SignNegativeParts: []bool{false},
		Coefficients: []smartvectors.SmartVector{
			smartvectors.NewConstant(field.One(), 8),
		},
		Selectors: []smartvectors.SmartVector{
			smartvectors.NewConstant(field.One(), 8),
		},
		N0s:         []int{0},
		N1s:         []int{8},
		Xs:          []field.Element{field.One()},
		FinalResult: field.NewElement(8),
	},

	{
		NameStr:           "positive/just-counting",
		SignNegativeParts: []bool{false},
		Coefficients: []smartvectors.SmartVector{
			smartvectors.NewConstant(field.One(), 8),
		},
		Selectors: []smartvectors.SmartVector{
			smartvectors.NewConstant(field.One(), 8),
		},
		N0s:         []int{0},
		N1s:         []int{8},
		Xs:          []field.Element{field.NewElement(2)},
		FinalResult: field.NewElement(255),
	},

	{
		NameStr:           "positive/12345..7",
		SignNegativeParts: []bool{false},
		Coefficients: []smartvectors.SmartVector{
			smartvectors.ForTest(0, 1, 2, 3, 4, 5, 6, 7),
		},
		Selectors: []smartvectors.SmartVector{
			smartvectors.NewConstant(field.One(), 8),
		},
		N0s:         []int{0},
		N1s:         []int{8},
		Xs:          []field.Element{field.NewElement(2)},
		FinalResult: field.NewElement(1538),
	},
}

var ListOfHornerTestcaseNegative = []*HornerTestcase{

	{
		NameStr:           "negative/none-selected-single/bad-count",
		SignNegativeParts: []bool{false},
		Coefficients: []smartvectors.SmartVector{
			smartvectors.NewConstant(field.One(), 8),
		},
		Selectors: []smartvectors.SmartVector{
			smartvectors.NewConstant(field.Zero(), 8),
		},
		N0s:          []int{0},
		N1s:          []int{1},
		Xs:           []field.Element{field.One()},
		FinalResult:  field.Zero(),
		MustFailFlag: true,
	},

	{
		NameStr:           "negative/none-selected-single/bad-result",
		SignNegativeParts: []bool{false},
		Coefficients: []smartvectors.SmartVector{
			smartvectors.NewConstant(field.One(), 8),
		},
		Selectors: []smartvectors.SmartVector{
			smartvectors.NewConstant(field.Zero(), 8),
		},
		N0s:          []int{0},
		N1s:          []int{0},
		Xs:           []field.Element{field.One()},
		FinalResult:  field.One(),
		MustFailFlag: true,
	},

	{
		NameStr:           "negative/two-parts-should-be-expected-to-cancel",
		SignNegativeParts: []bool{false, true},
		Coefficients: []smartvectors.SmartVector{
			RandomFromSeed(8, 1),
			RandomFromSeed(8, 1),
		},
		Selectors: []smartvectors.SmartVector{
			smartvectors.NewConstant(field.One(), 8),
			smartvectors.NewConstant(field.One(), 8),
		},
		N0s:          []int{0, 0},
		N1s:          []int{8, 8},
		Xs:           []field.Element{field.One(), field.One()},
		FinalResult:  field.One(),
		MustFailFlag: true,
	},

	{
		NameStr:           "negative/two-parts-should-be-expected-to-cancel",
		SignNegativeParts: []bool{false, true},
		Coefficients: []smartvectors.SmartVector{
			RandomFromSeed(8, 1),
			RandomFromSeed(8, 1),
		},
		Selectors: []smartvectors.SmartVector{
			smartvectors.NewConstant(field.One(), 8),
			smartvectors.NewConstant(field.One(), 8),
		},
		N0s:          []int{0, 0},
		N1s:          []int{8, 7},
		Xs:           []field.Element{field.One(), field.One()},
		FinalResult:  field.Zero(),
		MustFailFlag: true,
	},

	{
		NameStr:           "negative/just-counting/bad-n0",
		SignNegativeParts: []bool{false},
		Coefficients: []smartvectors.SmartVector{
			smartvectors.NewConstant(field.One(), 8),
		},
		Selectors: []smartvectors.SmartVector{
			smartvectors.NewConstant(field.One(), 8),
		},
		N0s:          []int{1},
		N1s:          []int{8},
		Xs:           []field.Element{field.One()},
		FinalResult:  field.NewElement(8),
		MustFailFlag: true,
	},

	{
		NameStr:           "negative/just-counting-x=2/bad-result",
		SignNegativeParts: []bool{false},
		Coefficients: []smartvectors.SmartVector{
			smartvectors.NewConstant(field.One(), 8),
		},
		Selectors: []smartvectors.SmartVector{
			smartvectors.NewConstant(field.One(), 8),
		},
		N0s:          []int{0},
		N1s:          []int{8},
		Xs:           []field.Element{field.NewElement(2)},
		FinalResult:  field.NewElement(510),
		MustFailFlag: true,
	},
}

func (t *HornerTestcase) Define(comp *wizard.CompiledIOP) {

	parts := make([]query.HornerPart, len(t.Coefficients))
	for i := range parts {
		parts[i] = query.HornerPart{
			SignNegative: t.SignNegativeParts[i],
			Coefficients: []*sym.Expression{sym.NewVariable(comp.InsertCommit(
				0,
				formatName[ifaces.ColID]("Horner", t.NameStr, "Coefficient", i),
				t.Coefficients[i].Len(),
			))},
			Selectors: []ifaces.Column{comp.InsertCommit(
				0,
				formatName[ifaces.ColID]("Horner", t.NameStr, "Selector", i),
				t.Selectors[i].Len(),
			)},
			X: accessors.NewConstant(t.Xs[i]),
		}
	}

	t.Query = comp.InsertHornerQuery(
		0,
		formatName[ifaces.QueryID]("Horner", t.NameStr, "Query"),
		parts,
	)
}

func (t *HornerTestcase) Assign(run *wizard.ProverRuntime) {

	parts := make([]query.HornerParamsPart, len(t.Coefficients))

	for i := range parts {

		parts[i] = query.HornerParamsPart{
			N0: t.N0s[i],
			N1: t.N1s[i],
		}

		run.AssignColumn(
			formatName[ifaces.ColID]("Horner", t.NameStr, "Coefficient", i),
			t.Coefficients[i],
		)

		run.AssignColumn(
			formatName[ifaces.ColID]("Horner", t.NameStr, "Selector", i),
			t.Selectors[i],
		)
	}

	run.AssignHornerParams(
		formatName[ifaces.QueryID]("Horner", t.NameStr, "Query"),
		query.HornerParams{
			Parts:       parts,
			FinalResult: t.FinalResult,
		},
	)
}

func (t *HornerTestcase) MustFail() bool {
	return t.MustFailFlag
}

func (t *HornerTestcase) Name() string {
	return t.NameStr
}
