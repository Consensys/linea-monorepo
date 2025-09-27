package symbolic

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"

	"github.com/stretchr/testify/assert"
)

func TestRemoveZeroes(t *testing.T) {

	var (
		a = NewDummyVar[zk.NativeElement]("a")
		b = NewDummyVar[zk.NativeElement]("b")
		c = NewDummyVar[zk.NativeElement]("c")
	)

	testCases := []struct {
		InputVars  []*Expression[zk.NativeElement]
		InputMagn  []int
		OutputVars []*Expression[zk.NativeElement]
		OutputMagn []int
	}{
		{
			InputVars:  []*Expression[zk.NativeElement]{a, a, b, c, b},
			InputMagn:  []int{1, 0, 2, 0, 1},
			OutputVars: []*Expression[zk.NativeElement]{a, b, b},
			OutputMagn: []int{1, 2, 1},
		},
	}

	for i := range testCases {

		oMagn, oVar := removeZeroCoeffs(testCases[i].InputMagn, testCases[i].InputVars)
		assert.Equal(t, testCases[i].OutputVars, oVar)
		assert.Equal(t, testCases[i].OutputMagn, oMagn)
	}
}

func TestExpandTerms(t *testing.T) {

	var (
		a = NewDummyVar[zk.NativeElement]("a")
		b = NewDummyVar[zk.NativeElement]("b")
		// c = NewDummyVar[zk.NativeElement]("c")
	)

	testCases := []struct {
		InputVars  []*Expression[zk.NativeElement]
		InputMagn  []int
		OutputVars []*Expression[zk.NativeElement]
		OutputMagn []int
		Op         Operator[zk.NativeElement]
	}{
		{
			InputVars:  []*Expression[zk.NativeElement]{Add[zk.NativeElement](a, b), Mul[zk.NativeElement](a, b)},
			InputMagn:  []int{2, 2},
			OutputVars: []*Expression[zk.NativeElement]{a, b, Mul[zk.NativeElement](a, b)},
			OutputMagn: []int{2, 2, 2},
			Op:         &LinComb[zk.NativeElement]{},
		},
		{
			InputVars:  []*Expression[zk.NativeElement]{Add[zk.NativeElement](a, b), Mul[zk.NativeElement](a, b)},
			InputMagn:  []int{2, 2},
			OutputVars: []*Expression[zk.NativeElement]{Add[zk.NativeElement](a, b), a, b},
			OutputMagn: []int{2, 2, 2},
			Op:         &Product[zk.NativeElement]{},
		},
		{
			InputVars:  []*Expression[zk.NativeElement]{Add[zk.NativeElement](a, b), Mul[zk.NativeElement](a, b)},
			InputMagn:  []int{2, 2},
			OutputVars: []*Expression[zk.NativeElement]{a, b, Mul[zk.NativeElement](a, b)},
			OutputMagn: []int{2, 2, 2},
			Op:         LinComb[zk.NativeElement]{},
		},
		{
			InputVars:  []*Expression[zk.NativeElement]{Add[zk.NativeElement](a, b), Mul[zk.NativeElement](a, b)},
			InputMagn:  []int{2, 2},
			OutputVars: []*Expression[zk.NativeElement]{Add[zk.NativeElement](a, b), a, b},
			OutputMagn: []int{2, 2, 2},
			Op:         Product[zk.NativeElement]{},
		},
	}

	for _, tc := range testCases {
		oMagn, oVars := expandTerms(tc.Op, tc.InputMagn, tc.InputVars)
		assert.Equal(t, tc.OutputVars, oVars)
		assert.Equal(t, tc.OutputMagn, oMagn)
	}
}

func TestRegroupTerms(t *testing.T) {

	var (
		a    = NewDummyVar[zk.NativeElement]("a")
		b    = NewDummyVar[zk.NativeElement]("b")
		c    = NewDummyVar[zk.NativeElement]("c")
		zero = NewConstant[zk.NativeElement](0)
	)

	testCases := []struct {
		InputVars  []*Expression[zk.NativeElement]
		InputMagn  []int
		OutputVars []*Expression[zk.NativeElement]
		OutputMagn []int
		ConstVars  []fext.GenericFieldElem
		ConstMagn  []int
		Op         Operator[zk.NativeElement]
	}{
		{
			InputVars:  []*Expression[zk.NativeElement]{a, b, c, b, a, c, zero},
			InputMagn:  []int{1, 1, 1, 1, 1, 1, 1},
			OutputVars: []*Expression[zk.NativeElement]{a, b, c},
			OutputMagn: []int{2, 2, 2},
			ConstVars:  []fext.GenericFieldElem{fext.GenericFieldZero()},
			ConstMagn:  []int{1},
		},
	}

	for _, tc := range testCases {
		oMagn, oVars, cMagn, cVars := regroupTerms(tc.InputMagn, tc.InputVars)
		assert.Equal(t, tc.OutputVars, oVars)
		assert.Equal(t, tc.OutputMagn, oMagn)
		assert.Equal(t, tc.ConstVars, cVars)
		assert.Equal(t, tc.ConstMagn, cMagn)
	}
}
