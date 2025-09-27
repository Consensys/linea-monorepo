package simplify

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"

	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsFactored(t *testing.T) {

	testcases := []struct {
		Expr       *sym.Expression[zk.NativeElement]
		By         *sym.Expression[zk.NativeElement]
		IsFactored bool
		Factor     *sym.Expression[zk.NativeElement]
	}{
		{
			Expr:       sym.Mul[zk.NativeElement](a, b, c),
			By:         sym.Mul[zk.NativeElement](a, b),
			IsFactored: true,
			Factor:     c,
		},
		{
			Expr:       sym.Mul[zk.NativeElement](a, b, c),
			By:         sym.Mul[zk.NativeElement](a, c),
			IsFactored: true,
			Factor:     b,
		},
		{
			Expr:       sym.Mul[zk.NativeElement](a, b, c),
			By:         sym.Mul[zk.NativeElement](a, d),
			IsFactored: false,
			Factor:     nil,
		},
		{
			Expr:       sym.Add[zk.NativeElement](a, b, c),
			By:         sym.Add[zk.NativeElement](a, d),
			IsFactored: false,
			Factor:     nil,
		},
		{
			Expr:       sym.Add[zk.NativeElement](a, b, c),
			By:         sym.Add[zk.NativeElement](a, b, b),
			IsFactored: false,
			Factor:     nil,
		},
		{
			Expr:       sym.Mul[zk.NativeElement](a, b),
			By:         sym.Mul[zk.NativeElement](a, b),
			IsFactored: true,
			Factor:     sym.NewConstant[zk.NativeElement](1),
		},
		{
			Expr:       sym.Mul[zk.NativeElement](a, a),
			By:         a,
			IsFactored: true,
			Factor:     a,
		},
	}

	for i, tc := range testcases {

		t.Run(fmt.Sprintf("testcase-%v", i), func(t *testing.T) {
			// Build the group exponent map. If by is a product, we use directly
			// the exponents it contains. Otherwise, we say this is a single
			// term product with an exponent of 1.
			groupedExp := map[fext.GenericFieldElem]int{}
			if byProd, ok := tc.By.Operator.(sym.Product[zk.NativeElement]); ok {
				for i, ex := range byProd.Exponents {
					groupedExp[tc.By.Children[i].ESHash] = ex
				}
			} else {
				groupedExp[tc.By.ESHash] = 1
			}

			factored, isFactored := isFactored(tc.Expr, groupedExp)
			assert.Equalf(t, tc.IsFactored, isFactored, "missed factor identification")

			if isFactored && tc.IsFactored {
				assert.Equalf(t, tc.Factor.ESHash.String(), factored.ESHash.String(), "wrong factor")
			}
		})
	}
}

func TestFactorization(t *testing.T) {

	var (
		a = sym.NewDummyVar[zk.NativeElement]("a")
		b = sym.NewDummyVar[zk.NativeElement]("b")
		c = sym.NewDummyVar[zk.NativeElement]("c")
		d = sym.NewDummyVar[zk.NativeElement]("d")
	)

	testCases := []struct {
		Origin   *sym.Expression[zk.NativeElement]
		Factored *sym.Expression[zk.NativeElement]
	}{
		{
			Origin: sym.Add[zk.NativeElement](
				sym.Mul[zk.NativeElement](a, b),
				sym.Mul[zk.NativeElement](a, c),
				sym.Mul[zk.NativeElement](a, d),
			),
			Factored: sym.Mul[zk.NativeElement](
				a,
				sym.Add[zk.NativeElement](b, c, d),
			),
		},
		{
			Origin: sym.Add[zk.NativeElement](
				sym.Mul[zk.NativeElement](a, b, b),
				sym.Mul[zk.NativeElement](a, b, c),
				sym.Mul[zk.NativeElement](a, b, d),
			),
			Factored: sym.Mul[zk.NativeElement](
				a,
				b,
				sym.Add[zk.NativeElement](b, c, d),
			),
		},
		{
			Origin: sym.Add[zk.NativeElement](
				sym.Mul[zk.NativeElement](a, b),
				sym.Mul[zk.NativeElement](a, b, c),
				sym.Mul[zk.NativeElement](a, b, d),
			),
			Factored: sym.Mul[zk.NativeElement](
				a,
				b,
				sym.Add[zk.NativeElement](1, c, d),
			),
		},
		{
			Origin: sym.Add[zk.NativeElement](
				sym.Mul[zk.NativeElement](a, b),
				sym.Mul[zk.NativeElement](a, c),
				sym.Mul[zk.NativeElement](d),
			),
			Factored: sym.Add[zk.NativeElement](
				sym.Mul[zk.NativeElement](a, sym.Add[zk.NativeElement](b, c)),
				d,
			),
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("test-case-%v", i), func(t *testing.T) {
			factored := factorizeExpression[zk.NativeElement](testCase.Origin, 10)
			require.Equal(t, testCase.Origin.ESHash.String(), factored.ESHash.String())
			require.NoError(t, factored.Validate())
			assert.Equal(t, evaluateCostStat(testCase.Factored), evaluateCostStat(factored))
		})
	}
}

func TestFactorLinCompFromGroup(t *testing.T) {

	testCases := []struct {
		LinComb *sym.Expression[zk.NativeElement]
		Group   []*sym.Expression[zk.NativeElement]
		Res     *sym.Expression[zk.NativeElement]
	}{
		{
			LinComb: sym.Add[zk.NativeElement](
				sym.Mul[zk.NativeElement](a, b),
				sym.Mul[zk.NativeElement](a, b, c),
				sym.Mul[zk.NativeElement](a, b, d),
			),
			Group: []*sym.Expression[zk.NativeElement]{a, b},
			Res: sym.Mul[zk.NativeElement](
				a,
				b,
				sym.Add[zk.NativeElement](1, c, d),
			),
		},
		{
			LinComb: sym.Add[zk.NativeElement](
				sym.Mul[zk.NativeElement](a, c),
				sym.Mul[zk.NativeElement](a, a, b),
				1,
			),
			Group: []*sym.Expression[zk.NativeElement]{a},
			Res: sym.Add[zk.NativeElement](
				sym.Mul[zk.NativeElement](
					a,
					sym.Add[zk.NativeElement](
						sym.Mul[zk.NativeElement](a, b),
						c,
					),
				),
				1,
			),
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("test-case-%v", i), func(t *testing.T) {

			group := map[fext.GenericFieldElem]*sym.Expression[zk.NativeElement]{}
			for _, e := range testCase.Group {
				group[e.ESHash] = e
			}

			factored := factorLinCompFromGroup(testCase.LinComb, group)
			assert.Equal(t, testCase.LinComb.ESHash.String(), factored.ESHash.String())

			if t.Failed() {
				fmt.Printf("res=%v\n", testCase.Res.MarshalJSONString())
				fmt.Printf("factored=%v\n", factored.MarshalJSONString())
				t.Fatal()
			}

			require.NoError(t, factored.Validate())
			assert.Equal(t, evaluateCostStat(testCase.Res), evaluateCostStat(factored))

			if t.Failed() {
				fmt.Printf("res=%v\n", testCase.Res.MarshalJSONString())
				fmt.Printf("factored=%v\n", factored.MarshalJSONString())
			}
		})
	}

}
