package symbolic

import (
	"reflect"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"

	"github.com/stretchr/testify/assert"
)

func TestConstructorExt(t *testing.T) {

	var (
		a    = NewDummyVar[zk.NativeElement]("a")
		b    = NewDummyVar[zk.NativeElement]("b")
		c    = NewDummyVar[zk.NativeElement]("c")
		zero = NewConstant[zk.NativeElement](fext.NewFromUint(0, 0, 0, 0))
		one  = NewConstant[zk.NativeElement](fext.NewFromUint(1, 0, 0, 0))
	)

	testCases := []struct {
		Actual           *Expression[zk.NativeElement]
		ExpectedOperator Operator[zk.NativeElement]
		ExpectedParent   []*Expression[zk.NativeElement]
		Explainer        string
	}{
		{
			Actual:           a.Add(b),
			ExpectedOperator: LinComb[zk.NativeElement]{Coeffs: []int{1, 1}},
			ExpectedParent:   []*Expression[zk.NativeElement]{a, b},
			Explainer:        "Normally adding two variables",
		},
		{
			Actual:           a.Add(zero),
			ExpectedOperator: a.Operator,
			ExpectedParent:   []*Expression[zk.NativeElement]{},
			Explainer:        "When adding zero, this should be a no-op",
		},
		{
			Actual:           a.Add(a),
			ExpectedOperator: LinComb[zk.NativeElement]{Coeffs: []int{2}},
			ExpectedParent:   []*Expression[zk.NativeElement]{a},
			Explainer:        "When adding twice the same variable, this should resolve in a lincomb with a single term and a coefficient of 2",
		},
		{
			Actual:           a.Add(one).Add(one),
			ExpectedOperator: LinComb[zk.NativeElement]{Coeffs: []int{1, 1}},
			ExpectedParent:   []*Expression[zk.NativeElement]{a, NewConstant[zk.NativeElement](2)},
			Explainer:        "When adding twice a constant, the constants should be regrouped",
		},
		{
			Actual:           a.Add(b).Add(c),
			ExpectedOperator: LinComb[zk.NativeElement]{Coeffs: []int{1, 1, 1}},
			ExpectedParent:   []*Expression[zk.NativeElement]{a, b, c},
			Explainer:        "LinComb should be regrouped by associativity",
		},
		{
			Actual:           Sub[zk.NativeElement](a, b),
			ExpectedOperator: LinComb[zk.NativeElement]{Coeffs: []int{1, -1}},
			ExpectedParent:   []*Expression[zk.NativeElement]{a, b},
			Explainer:        "Normally substracting two variables should create a LinComb object",
		},
		{
			Actual:           a.Sub(zero),
			ExpectedOperator: a.Operator,
			ExpectedParent:   []*Expression[zk.NativeElement]{},
			Explainer:        "When substracting zero, this should be simplified into a no-op",
		},
		{
			Actual:           a.Sub(a),
			ExpectedOperator: zero.Operator,
			ExpectedParent:   []*Expression[zk.NativeElement]{},
			Explainer:        "When substracting a with itself, this should be simplified into zero",
		},
		{
			Actual:           a.Sub(one).Sub(one),
			ExpectedOperator: LinComb[zk.NativeElement]{Coeffs: []int{1, 1}},
			ExpectedParent:   []*Expression[zk.NativeElement]{a, NewConstant[zk.NativeElement](-2)},
			Explainer:        "When substracting twice by a constant, this should be simplified into LinComb with only `1` as coeffs",
		},
		{
			Actual:           a.Sub(b).Sub(c),
			ExpectedOperator: LinComb[zk.NativeElement]{Coeffs: []int{1, -1, -1}},
			ExpectedParent:   []*Expression[zk.NativeElement]{a, b, c},
			Explainer:        "When substracting a with b then c, this should be regrouped into a single linear combination",
		},
		{
			Actual:           a.Mul(b),
			ExpectedOperator: Product[zk.NativeElement]{Exponents: []int{1, 1}},
			ExpectedParent:   []*Expression[zk.NativeElement]{a, b},
			Explainer:        "Normally multiplying a with b should produce a product term",
		},
		{
			Actual:           a.Mul(zero),
			ExpectedOperator: zero.Operator,
			ExpectedParent:   []*Expression[zk.NativeElement]{},
			Explainer:        "Multiplying something by zero automatically returns the zero constant",
		},
		{
			Actual:           a.Mul(one),
			ExpectedOperator: a.Operator,
			ExpectedParent:   []*Expression[zk.NativeElement]{},
			Explainer:        "Multiplying something by 1 is a no-op and returns directly the multiplied expression",
		},
		{
			Actual:           a.Mul(a),
			ExpectedOperator: Product[zk.NativeElement]{Exponents: []int{2}},
			ExpectedParent:   []*Expression[zk.NativeElement]{a},
			Explainer:        "Multiplying a with itself produces a product with a single term with exponent 2",
		},
	}

	for i := range testCases {

		t.Run(testCases[i].Explainer, func(t *testing.T) {
			// This checks that the expression is well-formed and that the GenericFieldELem
			// is correctly computed.
			assert.NoError(t, testCases[i].Actual.Validate(), "the expression is not well-formed")
			assert.Equal(t, testCases[i].ExpectedOperator, testCases[i].Actual.Operator, "the operator does not match")
			assert.Equal(t, testCases[i].ExpectedParent, testCases[i].ExpectedParent, "the expected parents do not match")
		})
	}

}

func TestImmutableConstructorsExt(t *testing.T) {

	var (
		// Importantly, this implements Metadata but is not Variable. We will
		// be expecting that the tested functions perform the casting over them
		a    = StringVarExt("a")
		b    = StringVarExt("b")
		c    = StringVarExt("c")
		aVar = NewVariable[zk.NativeElement](a)
		bVar = NewVariable[zk.NativeElement](b)
		cVar = NewVariable[zk.NativeElement](c)
		one  = NewConstant[zk.NativeElement](fext.NewFromUint(1, 0, 0, 0))
	)

	testCases := []struct {
		Actual    *Expression[zk.NativeElement]
		Expected  *Expression[zk.NativeElement]
		Explainer string
	}{
		{
			Actual:    Add[zk.NativeElement](a, b, 1),
			Expected:  aVar.Add(bVar).Add(one),
			Explainer: "Adding two metadata and a constant",
		},
		{
			Actual:    Add[zk.NativeElement](aVar, bVar, c),
			Expected:  aVar.Add(bVar).Add(cVar),
			Explainer: "Adding two expressions and a metadata",
		},
		{
			Actual:    Sub[zk.NativeElement](a, b, 1),
			Expected:  aVar.Sub(bVar).Sub(one),
			Explainer: "Substracting two metadata and a constant",
		},
		{
			Actual:    Sub[zk.NativeElement](aVar, bVar, c),
			Expected:  aVar.Sub(bVar).Sub(cVar),
			Explainer: "Substracting two expressions and a metadata",
		},
		{
			Actual:    Mul[zk.NativeElement](a, b, 1),
			Expected:  aVar.Mul(bVar).Mul(one),
			Explainer: "Multiplying two metadata and a constant",
		},
		{
			Actual:    Mul[zk.NativeElement](aVar, bVar, c),
			Expected:  aVar.Mul(bVar).Mul(cVar),
			Explainer: "Multiplying two expressions and a metadata",
		},
		{
			Actual:    Square[zk.NativeElement](aVar),
			Expected:  aVar.Square(),
			Explainer: "Squaring an expression",
		},
		{
			Actual:    Square[zk.NativeElement](a),
			Expected:  aVar.Square(),
			Explainer: "Squaring a metadata",
		},
		{
			Actual:    Square[zk.NativeElement](1),
			Expected:  one.Square(),
			Explainer: "Squaring a constant",
		},
		{
			Actual:    Neg[zk.NativeElement](aVar),
			Expected:  aVar.Neg(),
			Explainer: "Negating an expression",
		},
		{
			Actual:    Neg[zk.NativeElement](a),
			Expected:  aVar.Neg(),
			Explainer: "Negating a metadata",
		},
		{
			Actual:    Neg[zk.NativeElement](1),
			Expected:  one.Neg(),
			Explainer: "Negating a constant",
		},
		{
			Actual:    Pow[zk.NativeElement](aVar, 4),
			Expected:  aVar.Pow(4),
			Explainer: "Exponentiating an expression",
		},
		{
			Actual:    Pow[zk.NativeElement](a, 4),
			Expected:  aVar.Pow(4),
			Explainer: "Exponentiating a metadata",
		},
		{
			Actual:    Pow[zk.NativeElement](1, 4),
			Expected:  one.Pow(4),
			Explainer: "Exponentiating a constant",
		},
	}

	for i := range testCases {
		t.Run(testCases[i].Explainer, func(t *testing.T) {
			assert.Equal(t, testCases[i].Expected, testCases[i].Actual)
		})
	}

}

func TestConstructorMixed(t *testing.T) {

	var (
		a = NewDummyVar[zk.NativeElement]("a")
		b = NewDummyVar[zk.NativeElement]("b")
		c = NewDummyVar[zk.NativeElement]("c")
		// extension variables
		extA = NewDummyVar[zk.NativeElement]("extA")
		extB = NewDummyVar[zk.NativeElement]("extB")
		extC = NewDummyVar[zk.NativeElement]("extC")
		// constants
		zero    = NewConstant[zk.NativeElement](fext.NewFromUint(0, 0, 0, 0))
		one     = NewConstant[zk.NativeElement](fext.NewFromUint(1, 0, 0, 0))
		oneBase = NewConstant[zk.NativeElement](field.NewElement(1))
	)

	testCases := []struct {
		Actual           *Expression[zk.NativeElement]
		ExpectedOperator Operator[zk.NativeElement]
		ExpectedChildren []*Expression[zk.NativeElement]
		Explainer        string
	}{
		{
			Actual:           Add[zk.NativeElement](a, b, extA, extB),
			ExpectedOperator: LinComb[zk.NativeElement]{Coeffs: []int{1, 1, 1, 1}},
			ExpectedChildren: []*Expression[zk.NativeElement]{a, b, extA, extB},
			Explainer:        "Normally adding two variables",
		},
		{
			Actual:           a.Add(zero),
			ExpectedOperator: a.Operator,
			ExpectedChildren: []*Expression[zk.NativeElement]{},
			Explainer:        "When adding zero, this should be a no-op",
		},
		{
			Actual:           a.Add(a),
			ExpectedOperator: LinComb[zk.NativeElement]{Coeffs: []int{2}},
			ExpectedChildren: []*Expression[zk.NativeElement]{a},
			Explainer:        "When adding twice the same variable, this should resolve in a lincomb with a single term and a coefficient of 2",
		},
		{
			Actual:           a.Add(one).Add(one),
			ExpectedOperator: LinComb[zk.NativeElement]{Coeffs: []int{1, 1}},
			ExpectedChildren: []*Expression[zk.NativeElement]{a, NewConstant[zk.NativeElement](2)},
			Explainer:        "When adding twice a constant, the constants should be regrouped",
		},
		{
			Actual:           Add[zk.NativeElement](a, b, c, extA, extB, extC),
			ExpectedOperator: LinComb[zk.NativeElement]{Coeffs: []int{1, 1, 1, 1, 1, 1}},
			ExpectedChildren: []*Expression[zk.NativeElement]{a, b, c, extA, extB, extC},
			Explainer:        "LinComb should be regrouped by associativity",
		},
		{
			Actual:           Sub[zk.NativeElement](a, b, extA, extB),
			ExpectedOperator: LinComb[zk.NativeElement]{Coeffs: []int{1, -1, -1, -1}},
			ExpectedChildren: []*Expression[zk.NativeElement]{a, b, extA, extB},
			Explainer:        "Normally substracting two variables should create a LinComb object",
		},
		{
			Actual:           a.Sub(zero),
			ExpectedOperator: a.Operator,
			ExpectedChildren: []*Expression[zk.NativeElement]{},
			Explainer:        "When substracting zero, this should be simplified into a no-op",
		},
		{
			Actual:           a.Sub(a),
			ExpectedOperator: zero.Operator,
			ExpectedChildren: []*Expression[zk.NativeElement]{},
			Explainer:        "When substracting a with itself, this should be simplified into zero",
		},
		{
			Actual:           a.Sub(one).Sub(one),
			ExpectedOperator: LinComb[zk.NativeElement]{Coeffs: []int{1, 1}},
			ExpectedChildren: []*Expression[zk.NativeElement]{a, NewConstant[zk.NativeElement](-2)},
			Explainer:        "When substracting twice by a constant, this should be simplified into LinComb with only `1` as coeffs",
		},
		{
			Actual:           Sub[zk.NativeElement](a, b, c, extA, extB, extC),
			ExpectedOperator: LinComb[zk.NativeElement]{Coeffs: []int{1, -1, -1, -1, -1, -1}},
			ExpectedChildren: []*Expression[zk.NativeElement]{a, b, c, extA, extB, extC},
			Explainer:        "When substracting a with b then c, this should be regrouped into a single linear combination",
		},
		{
			Actual:           Mul[zk.NativeElement](a, b, extA, extB),
			ExpectedOperator: Product[zk.NativeElement]{Exponents: []int{1, 1, 1, 1}},
			ExpectedChildren: []*Expression[zk.NativeElement]{a, b, extA, extB},
			Explainer:        "Normally multiplying a with b should produce a product term",
		},
		{
			Actual:           Mul[zk.NativeElement](a, extA, zero),
			ExpectedOperator: zero.Operator,
			ExpectedChildren: []*Expression[zk.NativeElement]{},
			Explainer:        "Multiplying something by zero automatically returns the zero constant",
		},
		{
			Actual:           Mul[zk.NativeElement](a, one, oneBase),
			ExpectedOperator: a.Operator,
			ExpectedChildren: []*Expression[zk.NativeElement]{},
			Explainer:        "Multiplying something by 1 is a no-op and returns directly the multiplied expression",
		},
		{
			Actual:           Mul[zk.NativeElement](a, a),
			ExpectedOperator: Product[zk.NativeElement]{Exponents: []int{2}},
			ExpectedChildren: []*Expression[zk.NativeElement]{a},
			Explainer:        "Multiplying a with itself produces a product with a single term with exponent 2",
		},
	}

	for i := range testCases {

		t.Run(testCases[i].Explainer, func(t *testing.T) {
			// This checks that the expression is well-formed and that the GenericFieldELem
			// is correctly computed.
			assert.NoError(t, testCases[i].Actual.Validate(), "the expression is not well-formed")
			assert.Equal(t, testCases[i].ExpectedOperator, testCases[i].Actual.Operator, "the operator does not match")
			assert.Equal(t, true, reflect.DeepEqual(testCases[i].ExpectedChildren, testCases[i].Actual.Children), "the expected children do not match")
		})
	}

}

func TestImmutableConstructorsMixed(t *testing.T) {

	var (
		// Importantly, this implements Metadata but is not Variable. We will
		// be expecting that the tested functions perform the casting over them
		a = StringVar("a")
		b = StringVar("b")
		c = StringVar("c")
		// extension variables
		extA = StringVarExt("a")
		extB = StringVarExt("b")
		extC = StringVarExt("c")
		// immutable versions of the base variables
		aVar = NewVariable[zk.NativeElement](a)
		bVar = NewVariable[zk.NativeElement](b)
		cVar = NewVariable[zk.NativeElement](c)
		// immutable versions of the extension variables
		aExtVar = NewVariable[zk.NativeElement](extA)
		bExtVar = NewVariable[zk.NativeElement](extB)
		cExtVar = NewVariable[zk.NativeElement](extC)
		// constants
		one     = NewConstant[zk.NativeElement](fext.NewFromUint(1, 0, 0, 0))
		oneBase = NewConstant[zk.NativeElement](field.NewElement(1))
	)

	testCases := []struct {
		Actual    *Expression[zk.NativeElement]
		Expected  *Expression[zk.NativeElement]
		Explainer string
	}{
		{
			Actual:    Add[zk.NativeElement](a, b, 1, extA, extB),
			Expected:  Add[zk.NativeElement](aVar, bVar, one, aExtVar, bExtVar),
			Explainer: "Adding two metadata and a constant",
		},
		{
			Actual:    Add[zk.NativeElement](aVar, bVar, c, aExtVar, bExtVar, extC),
			Expected:  Add[zk.NativeElement](aVar, bVar, cVar, aExtVar, bExtVar, cExtVar),
			Explainer: "Adding two expressions and a metadata",
		},
		{
			Actual:    Sub[zk.NativeElement](a, b, 1, extA, extB),
			Expected:  Sub[zk.NativeElement](aVar, bVar, one, aExtVar, bExtVar),
			Explainer: "Substracting two metadata and a constant",
		},
		{
			Actual:    Sub[zk.NativeElement](aVar, bVar, c, aExtVar, bExtVar, extC),
			Expected:  Sub[zk.NativeElement](aVar, bVar, cVar, aExtVar, bExtVar, cExtVar),
			Explainer: "Substracting two expressions and a metadata",
		},
		{
			Actual:    Mul[zk.NativeElement](a, b, 1, extA, extB),
			Expected:  Mul[zk.NativeElement](aVar, bVar, one, aExtVar, bExtVar, oneBase),
			Explainer: "Multiplying two metadata and a constant",
		},
		{
			Actual:    Mul[zk.NativeElement](aVar, bVar, c, aExtVar, bExtVar, extC),
			Expected:  Mul[zk.NativeElement](aVar, bVar, cVar, aExtVar, bExtVar, cExtVar),
			Explainer: "Multiplying two expressions and a metadata",
		},
	}

	for i := range testCases {
		t.Run(testCases[i].Explainer, func(t *testing.T) {
			assert.Equal(t, testCases[i].Expected, testCases[i].Actual)
		})
	}

}
