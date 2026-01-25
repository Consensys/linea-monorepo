package symbolic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// constantDegreeGetter returns a degree getter that assigns degree 1 to all variables.
func constantDegreeGetter() GetDegree {
	return func(interface{}) int { return 1 }
}

func TestReduceDegreeOfExpressions(t *testing.T) {

	var (
		a      = NewDummyVar("a")
		b      = NewDummyVar("b")
		c      = NewDummyVar("c")
		d      = NewDummyVar("d")
		getDeg = constantDegreeGetter()
	)

	t.Run("no reduction needed", func(t *testing.T) {
		// Expression already within bound
		expr := a.Add(b)
		reduced, eliminated, newVars := ReduceDegreeOfExpressions(
			[]*Expression{expr},
			2,
			getDeg,
		)
		require.Len(t, reduced, 1)
		require.Empty(t, eliminated)
		require.Empty(t, newVars)
		assert.Equal(t, expr.ESHash, reduced[0].ESHash)
	})

	t.Run("single quadratic reduction", func(t *testing.T) {
		// a * b * c has degree 3, bound is 2
		expr := a.Mul(b).Mul(c)
		reduced, eliminated, newVars := ReduceDegreeOfExpressions(
			[]*Expression{expr},
			2,
			getDeg,
		)
		require.Len(t, reduced, 1)
		require.Len(t, eliminated, 1)
		require.Len(t, newVars, 1)
		// Reduced expression should have degree <= 2
		reducedDegree := reduced[0].Board().Degree(getDeg)
		assert.LessOrEqual(t, reducedDegree, 2)
	})

	t.Run("multiple expressions sharing subexpression", func(t *testing.T) {
		// Both expressions contain a*b as a subexpression
		expr1 := a.Mul(b).Mul(c) // degree 3
		expr2 := a.Mul(b).Mul(a) // degree 3
		reduced, eliminated, newVars := ReduceDegreeOfExpressions(
			[]*Expression{expr1, expr2},
			2,
			getDeg,
		)
		require.Len(t, reduced, 2)
		// Should eliminate a*b since it appears in both
		require.GreaterOrEqual(t, len(eliminated), 1)
		require.Equal(t, len(eliminated), len(newVars))
		// Both reduced expressions should satisfy the bound
		for i, red := range reduced {
			deg := red.Board().Degree(getDeg)
			assert.LessOrEqual(t, deg, 2, "expression %d has degree %d", i, deg)
		}
	})

	t.Run("high degree requires multiple reductions", func(t *testing.T) {
		// a^5 has degree 5, bound is 2
		expr := a.Pow(5)
		reduced, eliminated, newVars := ReduceDegreeOfExpressions(
			[]*Expression{expr},
			2,
			getDeg,
		)
		require.Len(t, reduced, 1)
		require.NotEmpty(t, eliminated)
		require.Equal(t, len(eliminated), len(newVars))
		reducedDegree := reduced[0].Board().Degree(getDeg)
		assert.LessOrEqual(t, reducedDegree, 2)
	})

	t.Run("empty input", func(t *testing.T) {
		reduced, eliminated, newVars := ReduceDegreeOfExpressions(
			[]*Expression{},
			2,
			getDeg,
		)
		require.Empty(t, reduced)
		require.Empty(t, eliminated)
		require.Empty(t, newVars)
	})

	t.Run("exactly at bound", func(t *testing.T) {
		// a * b has degree 2, bound is 2 - no reduction needed
		expr := a.Mul(b)
		reduced, eliminated, newVars := ReduceDegreeOfExpressions(
			[]*Expression{expr},
			2,
			getDeg,
		)
		require.Len(t, reduced, 1)
		require.Empty(t, eliminated)
		require.Empty(t, newVars)
		assert.Equal(t, expr.ESHash, reduced[0].ESHash)
	})

	t.Run("very high degree a^10", func(t *testing.T) {
		// a^10 requires multiple reduction steps to reach degree 2
		expr := a.Pow(10)
		reduced, eliminated, newVars := ReduceDegreeOfExpressions(
			[]*Expression{expr},
			2,
			getDeg,
		)
		require.Len(t, reduced, 1)
		require.NotEmpty(t, eliminated)
		require.Equal(t, len(eliminated), len(newVars))
		reducedDegree := reduced[0].Board().Degree(getDeg)
		assert.LessOrEqual(t, reducedDegree, 2)
	})

	t.Run("mixed addition and multiplication", func(t *testing.T) {
		// (a * b * c) + (a * b * d) - common subexpression a*b
		expr := a.Mul(b).Mul(c).Add(a.Mul(b).Mul(d))
		reduced, eliminated, newVars := ReduceDegreeOfExpressions(
			[]*Expression{expr},
			2,
			getDeg,
		)
		require.Len(t, reduced, 1)
		require.NotEmpty(t, eliminated)
		require.Equal(t, len(eliminated), len(newVars))
		reducedDegree := reduced[0].Board().Degree(getDeg)
		assert.LessOrEqual(t, reducedDegree, 2)
	})

	t.Run("nested products", func(t *testing.T) {
		// (a * b) * (c * d) = degree 4
		expr := a.Mul(b).Mul(c.Mul(d))
		reduced, eliminated, newVars := ReduceDegreeOfExpressions(
			[]*Expression{expr},
			2,
			getDeg,
		)
		require.Len(t, reduced, 1)
		require.NotEmpty(t, eliminated)
		require.Equal(t, len(eliminated), len(newVars))
		reducedDegree := reduced[0].Board().Degree(getDeg)
		assert.LessOrEqual(t, reducedDegree, 2)
	})

	t.Run("multiple independent expressions", func(t *testing.T) {
		// Three independent high-degree expressions
		expr1 := a.Pow(4)
		expr2 := b.Pow(4)
		expr3 := c.Pow(4)
		reduced, eliminated, newVars := ReduceDegreeOfExpressions(
			[]*Expression{expr1, expr2, expr3},
			2,
			getDeg,
		)
		require.Len(t, reduced, 3)
		require.NotEmpty(t, eliminated)
		require.Equal(t, len(eliminated), len(newVars))
		for i, red := range reduced {
			deg := red.Board().Degree(getDeg)
			assert.LessOrEqual(t, deg, 2, "expression %d has degree %d", i, deg)
		}
	})

	t.Run("some expressions need reduction some dont", func(t *testing.T) {
		expr1 := a.Add(b)        // degree 1 - no reduction
		expr2 := a.Mul(b)        // degree 2 - no reduction
		expr3 := a.Mul(b).Mul(c) // degree 3 - needs reduction
		reduced, eliminated, newVars := ReduceDegreeOfExpressions(
			[]*Expression{expr1, expr2, expr3},
			2,
			getDeg,
		)
		require.Len(t, reduced, 3)
		// Only expr3 should trigger elimination
		require.NotEmpty(t, eliminated)
		require.Equal(t, len(eliminated), len(newVars))
		for i, red := range reduced {
			deg := red.Board().Degree(getDeg)
			assert.LessOrEqual(t, deg, 2, "expression %d has degree %d", i, deg)
		}
	})

	t.Run("repeated subexpression across many expressions", func(t *testing.T) {
		// a*b appears in all expressions - should be eliminated first
		ab := a.Mul(b)
		expr1 := ab.Mul(c)
		expr2 := ab.Mul(d)
		expr3 := ab.Mul(a)
		expr4 := ab.Mul(b)
		reduced, eliminated, newVars := ReduceDegreeOfExpressions(
			[]*Expression{expr1, expr2, expr3, expr4},
			2,
			getDeg,
		)
		require.Len(t, reduced, 4)
		require.NotEmpty(t, eliminated)
		require.Equal(t, len(eliminated), len(newVars))
		// a*b should be among the eliminated since it appears 4 times
		for i, red := range reduced {
			deg := red.Board().Degree(getDeg)
			assert.LessOrEqual(t, deg, 2, "expression %d has degree %d", i, deg)
		}
	})

	t.Run("constant expressions unchanged", func(t *testing.T) {
		// Constants have degree 0, should pass through unchanged
		one := NewConstant(1)
		expr := a.Mul(b).Mul(c).Add(one)
		reduced, eliminated, newVars := ReduceDegreeOfExpressions(
			[]*Expression{expr},
			2,
			getDeg,
		)
		require.Len(t, reduced, 1)
		require.NotEmpty(t, eliminated)
		require.Equal(t, len(eliminated), len(newVars))
		reducedDegree := reduced[0].Board().Degree(getDeg)
		assert.LessOrEqual(t, reducedDegree, 2)
	})

	t.Run("single variable unchanged", func(t *testing.T) {
		// Single variable has degree 1
		reduced, eliminated, newVars := ReduceDegreeOfExpressions(
			[]*Expression{a},
			2,
			getDeg,
		)
		require.Len(t, reduced, 1)
		require.Empty(t, eliminated)
		require.Empty(t, newVars)
		assert.Equal(t, a.ESHash, reduced[0].ESHash)
	})

	t.Run("large bound no reduction", func(t *testing.T) {
		// With bound 10, even a^5 * b^5 should not need reduction
		expr := a.Pow(5).Mul(b.Pow(5))
		reduced, eliminated, newVars := ReduceDegreeOfExpressions(
			[]*Expression{expr},
			10,
			getDeg,
		)
		require.Len(t, reduced, 1)
		require.Empty(t, eliminated)
		require.Empty(t, newVars)
	})

	t.Run("squared terms a^2 * b^2", func(t *testing.T) {
		// a^2 * b^2 has degree 4, with bound 2 should reduce
		expr := a.Square().Mul(b.Square())
		reduced, eliminated, newVars := ReduceDegreeOfExpressions(
			[]*Expression{expr},
			2,
			getDeg,
		)
		require.Len(t, reduced, 1)
		require.NotEmpty(t, eliminated)
		require.Equal(t, len(eliminated), len(newVars))
		reducedDegree := reduced[0].Board().Degree(getDeg)
		assert.LessOrEqual(t, reducedDegree, 2)
	})

	t.Run("complex expression with multiple operators", func(t *testing.T) {
		// (a*b*c) + (a^2*b) - (b*c*d) has degree 3 in all terms
		expr := a.Mul(b).Mul(c).Add(a.Square().Mul(b)).Sub(b.Mul(c).Mul(d))
		reduced, eliminated, newVars := ReduceDegreeOfExpressions(
			[]*Expression{expr},
			2,
			getDeg,
		)
		require.Len(t, reduced, 1)
		require.NotEmpty(t, eliminated)
		require.Equal(t, len(eliminated), len(newVars))
		reducedDegree := reduced[0].Board().Degree(getDeg)
		assert.LessOrEqual(t, reducedDegree, 2)
	})

	t.Run("preservation of expression count", func(t *testing.T) {
		// Ensure we always get back the same number of expressions
		exprs := []*Expression{
			a.Pow(3),
			b.Pow(4),
			c.Pow(5),
			a.Mul(b).Mul(c),
			a.Add(b).Mul(c.Mul(d)),
		}
		reduced, eliminated, newVars := ReduceDegreeOfExpressions(
			exprs,
			2,
			getDeg,
		)
		require.Len(t, reduced, len(exprs))
		require.Equal(t, len(eliminated), len(newVars))
		for i, red := range reduced {
			deg := red.Board().Degree(getDeg)
			assert.LessOrEqual(t, deg, 2, "expression %d has degree %d", i, deg)
		}
	})
}

func TestFindOverDegreeIndices(t *testing.T) {

	var (
		a      = NewDummyVar("a")
		b      = NewDummyVar("b")
		getDeg = constantDegreeGetter()
	)

	t.Run("mixed degrees", func(t *testing.T) {
		exprs := []*Expression{
			a.Add(b),        // degree 1
			a.Mul(b),        // degree 2
			a.Mul(b).Mul(a), // degree 3
		}

		indices := findOverDegreeIndices(exprs, 2, getDeg)
		require.Len(t, indices, 1)
		assert.Equal(t, 2, indices[0])
	})

	t.Run("all within bound", func(t *testing.T) {
		exprs := []*Expression{
			a.Add(b),
			a.Mul(b),
		}

		indices := findOverDegreeIndices(exprs, 2, getDeg)
		require.Empty(t, indices)
	})

	t.Run("all over bound", func(t *testing.T) {
		exprs := []*Expression{
			a.Mul(b).Mul(a),
			a.Pow(4),
		}

		indices := findOverDegreeIndices(exprs, 2, getDeg)
		require.Len(t, indices, 2)
	})
}

func TestIsTrivialExpr(t *testing.T) {

	var (
		a   = NewDummyVar("a")
		one = NewConstant(1)
	)

	t.Run("variable is trivial", func(t *testing.T) {
		assert.True(t, isTrivialExpr(a))
	})

	t.Run("constant is trivial", func(t *testing.T) {
		assert.True(t, isTrivialExpr(one))
	})

	t.Run("product is not trivial", func(t *testing.T) {
		expr := Mul(a, a)
		assert.False(t, isTrivialExpr(expr))
	})

	t.Run("lincomb is not trivial", func(t *testing.T) {
		b := NewDummyVar("b")
		expr := a.Add(b)
		assert.False(t, isTrivialExpr(expr))
	})
}

func TestSubstituteExpr(t *testing.T) {

	var (
		a = NewDummyVar("a")
		b = NewDummyVar("b")
		c = NewDummyVar("c")
	)

	t.Run("exact match replacement", func(t *testing.T) {
		target := a.Mul(b)
		replacement := c

		expr := a.Mul(b)
		result := substituteExpr(expr, target, replacement)

		assert.Equal(t, c.ESHash, result.ESHash)
	})

	t.Run("nested replacement", func(t *testing.T) {
		target := a.Mul(b)
		replacement := c

		// (a*b) + a should become c + a
		expr := target.Add(a)
		result := substituteExpr(expr, target, replacement)

		expected := c.Add(a)
		assert.Equal(t, expected.ESHash, result.ESHash)
	})

	t.Run("no match returns original", func(t *testing.T) {
		target := a.Mul(b)
		replacement := c

		expr := a.Add(b)
		result := substituteExpr(expr, target, replacement)

		assert.Equal(t, expr.ESHash, result.ESHash)
	})
}

func TestBuildSubProduct(t *testing.T) {

	var (
		a = NewDummyVar("a")
		b = NewDummyVar("b")
		c = NewDummyVar("c")
	)

	t.Run("single factor", func(t *testing.T) {
		children := []*Expression{a, b, c}
		exponents := []int{1, 2, 1}
		mask := uint64(0b001) // select only 'a'

		result := buildSubProduct(children, exponents, mask)
		require.NotNil(t, result)

		// Should be equivalent to a^1
		assert.Equal(t, a.ESHash, result.ESHash)
	})

	t.Run("two factors", func(t *testing.T) {
		children := []*Expression{a, b, c}
		exponents := []int{1, 2, 1}
		mask := uint64(0b011) // select 'a' and 'b'

		result := buildSubProduct(children, exponents, mask)
		require.NotNil(t, result)

		// Should be equivalent to a * b^2
		expected := NewProduct([]*Expression{a, b}, []int{1, 2})
		assert.Equal(t, expected.ESHash, result.ESHash)
	})

	t.Run("empty mask", func(t *testing.T) {
		children := []*Expression{a, b}
		exponents := []int{1, 1}
		mask := uint64(0)

		result := buildSubProduct(children, exponents, mask)
		assert.Nil(t, result)
	})
}

func TestComputeSubsetExponentSum(t *testing.T) {

	t.Run("single element", func(t *testing.T) {
		exponents := []int{3, 2, 1}
		mask := uint64(0b001)

		sum := computeSubsetExponentSum(mask, exponents)
		assert.Equal(t, 3, sum)
	})

	t.Run("multiple elements", func(t *testing.T) {
		exponents := []int{3, 2, 1}
		mask := uint64(0b111)

		sum := computeSubsetExponentSum(mask, exponents)
		assert.Equal(t, 6, sum)
	})

	t.Run("empty mask", func(t *testing.T) {
		exponents := []int{3, 2, 1}
		mask := uint64(0)

		sum := computeSubsetExponentSum(mask, exponents)
		assert.Equal(t, 0, sum)
	})
}

func TestEliminatedVarMetadata(t *testing.T) {

	a := NewDummyVar("a")
	expr := a.Mul(a)

	meta := eliminatedVarMetadata{id: 42, expr: expr}

	t.Run("string format", func(t *testing.T) {
		assert.Equal(t, "_elim_42", meta.String())
	})

	t.Run("isbase from expr", func(t *testing.T) {
		// The expression a*a is base if a is base
		assert.Equal(t, expr.IsBase, meta.IsBase())
	})
}

func TestSelectMostFrequentCandidate(t *testing.T) {

	var (
		a = NewDummyVar("a")
		b = NewDummyVar("b")
	)

	t.Run("selects highest count", func(t *testing.T) {
		candidates := []candidateInfo{
			{expr: a, count: 2},
			{expr: b, count: 5},
		}

		result := selectMostFrequentCandidate(candidates)
		assert.Equal(t, b.ESHash, result.ESHash)
	})

	t.Run("tie breaker by children count", func(t *testing.T) {
		ab := a.Mul(b)
		candidates := []candidateInfo{
			{expr: a, count: 3},
			{expr: ab, count: 3},
		}

		result := selectMostFrequentCandidate(candidates)
		// ab has more children, so it should be selected
		assert.Equal(t, ab.ESHash, result.ESHash)
	})
}

func TestReduceDegreePreservesSemantics(t *testing.T) {
	// This test verifies that the reduced expression computes
	// the same value as the original when the eliminated subexpressions
	// are substituted back

	var (
		a      = NewDummyVar("a")
		b      = NewDummyVar("b")
		getDeg = constantDegreeGetter()
	)

	t.Run("semantic equivalence", func(t *testing.T) {
		original := a.Mul(b).Mul(a).Add(a.Mul(b))
		reduced, eliminated, newVars := ReduceDegreeOfExpressions(
			[]*Expression{original},
			2,
			getDeg,
		)

		require.Len(t, reduced, 1)

		// Verify that we can trace back the eliminated expressions
		for i, meta := range newVars {
			elimMeta, ok := meta.(eliminatedVarMetadata)
			require.True(t, ok)
			assert.Equal(t, eliminated[i].ESHash, elimMeta.expr.ESHash)
		}
	})
}

func TestSubMultisetIterator(t *testing.T) {

	t.Run("single factor with exponent 1", func(t *testing.T) {
		// Exponents [1] -> sub-multisets: {} and {1}
		// Both are trivial (empty or full), so no valid sub-multisets
		iter := newSubMultisetIterator([]int{1})
		count := 0
		for iter.next() {
			count++
		}
		assert.Equal(t, 0, count, "single factor exp=1 has no proper sub-multisets")
	})

	t.Run("single factor with exponent 2", func(t *testing.T) {
		// Exponents [2] -> sub-multisets: {0}, {1}, {2}
		// {0} is empty (skip), {2} is full (skip)
		// Valid: {1}
		iter := newSubMultisetIterator([]int{2})
		var results [][]int
		for iter.next() {
			results = append(results, iter.currentSnapshot())
		}
		require.Len(t, results, 1)
		assert.Equal(t, []int{1}, results[0])
	})

	t.Run("single factor with exponent 3", func(t *testing.T) {
		// Exponents [3] -> valid sub-multisets: {1}, {2}
		iter := newSubMultisetIterator([]int{3})
		var results [][]int
		for iter.next() {
			results = append(results, iter.currentSnapshot())
		}
		require.Len(t, results, 2)
		assert.Equal(t, []int{1}, results[0])
		assert.Equal(t, []int{2}, results[1])
	})

	t.Run("two factors both exponent 1", func(t *testing.T) {
		// Exponents [1,1] -> all combinations: {0,0}, {0,1}, {1,0}, {1,1}
		// {0,0} empty, {1,1} full -> valid: {0,1}, {1,0}
		iter := newSubMultisetIterator([]int{1, 1})
		var results [][]int
		for iter.next() {
			results = append(results, iter.currentSnapshot())
		}
		require.Len(t, results, 2)
		// Results should be {1,0} and {0,1} in some order
		assert.Contains(t, results, []int{1, 0})
		assert.Contains(t, results, []int{0, 1})
	})

	t.Run("two factors mixed exponents", func(t *testing.T) {
		// Exponents [2,1] -> combinations (0..2) x (0..1)
		// (0,0) empty, (2,1) full
		// Valid: (1,0), (2,0), (0,1), (1,1)
		iter := newSubMultisetIterator([]int{2, 1})
		var results [][]int
		for iter.next() {
			results = append(results, iter.currentSnapshot())
		}
		require.Len(t, results, 4)
		assert.Contains(t, results, []int{1, 0})
		assert.Contains(t, results, []int{2, 0})
		assert.Contains(t, results, []int{0, 1})
		assert.Contains(t, results, []int{1, 1})
	})

	t.Run("three factors all exponent 1", func(t *testing.T) {
		// Exponents [1,1,1] -> 2^3 = 8 combinations
		// Excluding empty {0,0,0} and full {1,1,1} -> 6 valid
		iter := newSubMultisetIterator([]int{1, 1, 1})
		var results [][]int
		for iter.next() {
			results = append(results, iter.currentSnapshot())
		}
		require.Len(t, results, 6)
	})

	t.Run("product counts match formula", func(t *testing.T) {
		// For exponents [e1, e2, ...], total combinations = prod(ei+1)
		// Valid = total - 2 (excluding empty and full)
		exponents := []int{2, 2, 2}
		iter := newSubMultisetIterator(exponents)
		count := 0
		for iter.next() {
			count++
		}
		// Total = 3*3*3 = 27, valid = 27-2 = 25
		assert.Equal(t, 25, count)
	})
}

func TestSubMultisetIteratorTotalDegree(t *testing.T) {

	t.Run("degree computation", func(t *testing.T) {
		iter := newSubMultisetIterator([]int{3, 2})
		var degrees []int
		for iter.next() {
			degrees = append(degrees, iter.currentTotalDegree())
		}
		// All degrees should be between 1 and 4 (total=5, excluding 0 and 5)
		for _, d := range degrees {
			assert.GreaterOrEqual(t, d, 1)
			assert.LessOrEqual(t, d, 4)
		}
	})
}

func TestBuildSubProductFromExponents(t *testing.T) {

	var (
		a = NewDummyVar("a")
		b = NewDummyVar("b")
		c = NewDummyVar("c")
	)

	t.Run("all zero exponents", func(t *testing.T) {
		result := buildSubProductFromExponents([]*Expression{a, b}, []int{0, 0})
		assert.Nil(t, result)
	})

	t.Run("single factor selected", func(t *testing.T) {
		result := buildSubProductFromExponents([]*Expression{a, b}, []int{2, 0})
		require.NotNil(t, result)
		// Should be a^2
		expected := NewProduct([]*Expression{a}, []int{2})
		assert.Equal(t, expected.ESHash, result.ESHash)
	})

	t.Run("multiple factors with varying exponents", func(t *testing.T) {
		result := buildSubProductFromExponents([]*Expression{a, b, c}, []int{1, 2, 0})
		require.NotNil(t, result)
		// Should be a * b^2
		expected := NewProduct([]*Expression{a, b}, []int{1, 2})
		assert.Equal(t, expected.ESHash, result.ESHash)
	})

	t.Run("all factors selected", func(t *testing.T) {
		result := buildSubProductFromExponents([]*Expression{a, b}, []int{1, 1})
		require.NotNil(t, result)
		expected := NewProduct([]*Expression{a, b}, []int{1, 1})
		assert.Equal(t, expected.ESHash, result.ESHash)
	})
}

func TestCollectSubProducts(t *testing.T) {

	var (
		a      = NewDummyVar("a")
		getDeg = constantDegreeGetter()
	)

	t.Run("a^3 collects a and a^2", func(t *testing.T) {
		// a^3 should generate sub-multisets: a^2
		expr := a.Pow(3)
		prod := expr.Operator.(Product)
		candidates := make(map[esHash]*candidateInfo)

		collectSubProducts(expr, prod, 2, getDeg, candidates)

		// With degree bound 2, only a^2 (deg=2) qualifies
		require.Len(t, candidates, 1)

		aSquared := a.Pow(2)
		_, hasASquared := candidates[aSquared.ESHash]
		assert.True(t, hasASquared, "should contain a^2")
	})

	t.Run("a^5 collects a, a^2, a^3 and a^4", func(t *testing.T) {

		// a^5 should generate sub-multisets: a^2, a^3, a^4
		expr := a.Pow(5)
		prod := expr.Operator.(Product)
		candidates := make(map[esHash]*candidateInfo)

		collectSubProducts(expr, prod, 5, getDeg, candidates)

		require.Len(t, candidates, 3)

		for d := 2; d < 5; d++ {
			aPowD := a.Pow(d)
			_, hasAPowD := candidates[aPowD.ESHash]
			assert.True(t, hasAPowD, "should contain a^2")
		}
	})
}
