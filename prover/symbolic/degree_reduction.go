package symbolic

import (
	"fmt"
	"sort"

	"github.com/consensys/linea-monorepo/prover/utils"
)

// eliminatedVarMetadata implements Metadata for variables created during
// degree reduction. Each instance represents a substituted subexpression.
type eliminatedVarMetadata struct {
	id   int
	expr *Expression
}

// String returns a unique identifier for the eliminated variable.
func (e eliminatedVarMetadata) String() string {
	return fmt.Sprintf("_elim_%d", e.id)
}

// IsBase returns true since eliminated variables are treated as base field elements.
func (e eliminatedVarMetadata) IsBase() bool {
	return e.expr.IsBase
}

// ReduceDegreeOfExpressions performs a degree reduction algorithm by iteratively
// eliminating the most frequent common subexpressions whose degree is within
// the bound, from expressions exceeding that bound. The algorithm replaces
// each eliminated subexpression with a fresh variable and continues until
// all expressions satisfy the degree constraint.
//
// If the provided bound is one, the function will panic (as it would not be
// possible to degree-reduce it).
func ReduceDegreeOfExpressions(
	exprs []*Expression,
	degreeBound int,
	degreeGetter GetDegree,
) (degreeReduced []*Expression, eliminatedSubExpressions []*Expression, newVars []Metadata) {

	if degreeBound <= 1 {
		utils.Panic("cannot degree reduce with bound of %v", degreeBound)
	}

	// Working copies that get progressively transformed
	current := copyExprSlice(exprs)
	varCounter := 0

	for {
		// Identify expressions that still exceed the degree bound
		overDegreeIndices := findOverDegreeIndices(current, degreeBound, degreeGetter)
		if len(overDegreeIndices) == 0 {
			break
		}

		// Collect all candidate subexpressions from over-degree expressions
		candidates := collectCandidateSubexprs(current, overDegreeIndices, degreeBound, degreeGetter)
		if len(candidates) == 0 {
			// No valid candidates found; cannot reduce further
			break
		}

		// Select the most frequent candidate
		best := selectMostFrequentCandidate(candidates)

		// Create a new variable to replace this subexpression
		meta := eliminatedVarMetadata{id: varCounter, expr: best}
		varCounter++
		newVar := NewVariable(meta)

		// Substitute the candidate in all expressions
		current = substituteInAll(current, best, newVar)

		eliminatedSubExpressions = append(eliminatedSubExpressions, best)
		newVars = append(newVars, meta)
	}

	degreeReduced = current
	return
}

// copyExprSlice creates a shallow copy of the expression slice.
func copyExprSlice(exprs []*Expression) []*Expression {
	result := make([]*Expression, len(exprs))
	copy(result, exprs)
	return result
}

// findOverDegreeIndices returns indices of expressions exceeding the degree bound.
func findOverDegreeIndices(exprs []*Expression, bound int, degreeGetter GetDegree) []int {
	var indices []int
	for i, expr := range exprs {
		if expr.Degree(degreeGetter) > bound {
			indices = append(indices, i)
		}
	}
	return indices
}

// candidateInfo tracks a candidate subexpression and its occurrence count.
type candidateInfo struct {
	expr  *Expression
	count int
}

// collectCandidateSubexprs gathers all subexpressions from over-degree expressions
// that have degree within the bound. Includes both direct subexpressions and
// sub-products derived from Product nodes.
func collectCandidateSubexprs(
	exprs []*Expression,
	overDegreeIndices []int,
	degreeBound int,
	degreeGetter GetDegree,
) []candidateInfo {

	// Map from ESHash to candidate info for deduplication
	candidateMap := make(map[esHash]*candidateInfo)

	for _, idx := range overDegreeIndices {
		collectFromExpr(exprs[idx], degreeBound, degreeGetter, candidateMap)
	}

	return mapToSlice(candidateMap)
}

// collectFromExpr recursively collects candidate subexpressions from an expression.
func collectFromExpr(
	expr *Expression,
	degreeBound int,
	degreeGetter GetDegree,
	candidates map[esHash]*candidateInfo,
) {
	if expr == nil {
		return
	}

	degree := expr.Degree(degreeGetter)

	// If this expression is within bound and non-trivial, it's a candidate
	if degree > 0 && degree <= degreeBound && !isTrivialExpr(expr) {
		addCandidate(candidates, expr)
	}

	// Recurse into children
	for _, child := range expr.Children {
		collectFromExpr(child, degreeBound, degreeGetter, candidates)
	}

	// For Product nodes, enumerate sub-products as additional candidates
	if prod, ok := expr.Operator.(Product); ok {
		collectSubProducts(expr, prod, degreeBound, degreeGetter, candidates)
	}
}

// isTrivialExpr returns true for expressions that should not be substituted
// (constants, simple variables, or degree-zero expressions).
func isTrivialExpr(expr *Expression) bool {
	switch expr.Operator.(type) {
	case Constant, Variable:
		return true
	}
	return false
}

// addCandidate increments the count for a candidate or creates a new entry.
func addCandidate(candidates map[esHash]*candidateInfo, expr *Expression) {
	if existing, ok := candidates[expr.ESHash]; ok {
		existing.count++
	} else {
		candidates[expr.ESHash] = &candidateInfo{expr: expr, count: 1}
	}
}

// collectSubProducts enumerates all non-trivial sub-multisets of a Product
// expression and adds those within the degree bound as candidates.
func collectSubProducts(
	expr *Expression,
	prod Product,
	degreeBound int,
	degreeGetter GetDegree,
	candidates map[esHash]*candidateInfo,
) {
	if len(expr.Children) == 0 {
		return
	}

	iter := newSubMultisetIterator(prod.Exponents)
	for iter.next() {

		// Quick filter: skip if total exponent is less than 2
		if iter.currentTotalDegree() < 2 {
			continue
		}

		subExpr := buildSubProductFromExponents(expr.Children, iter.currentSnapshot())
		if subExpr == nil {
			continue
		}

		degree := subExpr.Degree(degreeGetter)
		if degree > 0 && degree <= degreeBound {
			addCandidate(candidates, subExpr)
		}
	}
}

// buildSubProductFromExponents constructs a Product from children with given exponents.
// Factors with zero exponent are excluded from the result.
func buildSubProductFromExponents(children []*Expression, exponents []int) *Expression {
	var subChildren []*Expression
	var subExponents []int

	for i, exp := range exponents {
		if exp > 0 {
			subChildren = append(subChildren, children[i])
			subExponents = append(subExponents, exp)
		}
	}

	if len(subChildren) == 0 {
		return nil
	}

	return NewProduct(subChildren, subExponents)
}

// factorOutSubProduct attempts to factor out target's exponents from expr's exponents.
// Returns the remaining factors if target is a sub-multiset of expr.
func factorOutSubProduct(
	exprChildren []*Expression,
	exprExponents []int,
	targetChildren []*Expression,
	targetExponents []int,
) (remaining []factorPair, found bool) {

	// Build map from ESHash to (index, remaining exponent) for expr
	type exponentEntry struct {
		index    int
		exponent int
	}
	exprMap := make(map[esHash]exponentEntry, len(exprChildren))
	for i, child := range exprChildren {
		exprMap[child.ESHash] = exponentEntry{index: i, exponent: exprExponents[i]}
	}

	// Subtract target exponents from expr exponents
	for i, child := range targetChildren {
		entry, exists := exprMap[child.ESHash]
		if !exists || entry.exponent < targetExponents[i] {
			return nil, false
		}
		entry.exponent -= targetExponents[i]
		exprMap[child.ESHash] = entry
	}

	// Collect remaining factors with positive exponents
	for _, child := range exprChildren {
		entry := exprMap[child.ESHash]
		if entry.exponent > 0 {
			remaining = append(remaining, factorPair{expr: child, exp: entry.exponent})
			// Mark as collected
			entry.exponent = 0
			exprMap[child.ESHash] = entry
		}
	}

	return remaining, true
}

// computeSubsetExponentSum calculates the sum of exponents for selected factors.
func computeSubsetExponentSum(mask uint64, exponents []int) int {
	sum := 0
	for i, exp := range exponents {
		if mask&(1<<i) != 0 {
			sum += exp
		}
	}
	return sum
}

// buildSubProduct constructs a Product expression from a subset of factors.
func buildSubProduct(children []*Expression, exponents []int, mask uint64) *Expression {
	var subChildren []*Expression
	var subExponents []int

	for i := range children {
		if mask&(1<<i) != 0 {
			subChildren = append(subChildren, children[i])
			subExponents = append(subExponents, exponents[i])
		}
	}

	if len(subChildren) == 0 {
		return nil
	}

	return NewProduct(subChildren, subExponents)
}

// mapToSlice converts the candidate map to a slice for sorting.
func mapToSlice(m map[esHash]*candidateInfo) []candidateInfo {
	result := make([]candidateInfo, 0, len(m))
	for _, info := range m {
		result = append(result, *info)
	}
	return result
}

// selectMostFrequentCandidate returns the candidate with the highest occurrence count.
// Ties are broken by selecting the highest-degree candidate (more reduction impact).
func selectMostFrequentCandidate(candidates []candidateInfo) *Expression {
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].count != candidates[j].count {
			return candidates[i].count > candidates[j].count
		}
		// Tie-breaker: prefer higher degree for more impact
		return len(candidates[i].expr.Children) > len(candidates[j].expr.Children)
	})
	return candidates[0].expr
}

// substituteInAll replaces all occurrences of target with replacement in each expression.
func substituteInAll(exprs []*Expression, target, replacement *Expression) []*Expression {
	result := make([]*Expression, len(exprs))
	for i, expr := range exprs {
		result[i] = substituteExpr(expr, target, replacement)
	}
	return result
}

// substituteExpr recursively replaces occurrences of target with replacement.
// Returns a new expression tree with substitutions applied.
func substituteExpr(expr, target, replacement *Expression) *Expression {
	if expr == nil {
		return nil
	}

	// Check if current expression matches the target
	if expr.ESHash == target.ESHash {
		return replacement
	}

	// For Product nodes, check if target is a sub-product
	if prod, ok := expr.Operator.(Product); ok {
		if substituted := substituteSubProduct(expr, prod, target, replacement); substituted != nil {
			return substituted
		}
	}

	// Recurse into children
	if len(expr.Children) == 0 {
		return expr
	}

	newChildren := substituteInChildren(expr.Children, target, replacement)
	if childrenUnchanged(expr.Children, newChildren) {
		return expr
	}

	return rebuildExpression(expr, newChildren)
}

// substituteInChildren applies substitution to each child expression.
func substituteInChildren(children []*Expression, target, replacement *Expression) []*Expression {
	newChildren := make([]*Expression, len(children))
	for i, child := range children {
		newChildren[i] = substituteExpr(child, target, replacement)
	}
	return newChildren
}

// childrenUnchanged returns true if no child was modified during substitution.
func childrenUnchanged(old, new []*Expression) bool {
	for i := range old {
		if old[i] != new[i] {
			return false
		}
	}
	return true
}

// rebuildExpression constructs a new expression with the same operator but new children.
func rebuildExpression(original *Expression, newChildren []*Expression) *Expression {
	switch op := original.Operator.(type) {
	case LinComb:
		return NewLinComb(newChildren, op.Coeffs)
	case Product:
		return NewProduct(newChildren, op.Exponents)
	case PolyEval:
		return NewPolyEval(newChildren[0], newChildren[1:])
	default:
		// For other operators, return original (should not happen for valid expressions)
		return original
	}
}

// substituteSubProduct checks if target is a sub-product of expr and performs
// the substitution by factoring out the target. Returns nil if not applicable.
func substituteSubProduct(
	expr *Expression,
	prod Product,
	target, replacement *Expression,
) *Expression {

	targetProd, ok := target.Operator.(Product)
	if !ok {
		return nil
	}

	// Check if target's factors are a subset of expr's factors
	remaining, found := factorOutSubProduct(expr.Children, prod.Exponents, target.Children, targetProd.Exponents)
	if !found {
		return nil
	}

	// Build the result: replacement * remaining factors
	return buildFactoredProduct(replacement, remaining)
}

// factorPair represents a factor and its exponent in a product.
type factorPair struct {
	expr *Expression
	exp  int
}

// buildFactoredProduct constructs a product of replacement and remaining factors.
func buildFactoredProduct(replacement *Expression, remaining []factorPair) *Expression {
	if len(remaining) == 0 {
		return replacement
	}

	children := []*Expression{replacement}
	exponents := []int{1}

	for _, f := range remaining {
		children = append(children, f.expr)
		exponents = append(exponents, f.exp)
	}

	return NewProduct(children, exponents)
}

// subMultisetIterator provides iteration over all non-trivial sub-multisets
// of a product's factors. A sub-multiset assigns to each factor an exponent
// between 0 and its original exponent (inclusive), excluding the empty
// multiset and the full multiset.
type subMultisetIterator struct {
	maxExponents     []int
	currentExponents []int
	exhausted        bool
}

// newSubMultisetIterator creates an iterator over sub-multisets of a product.
// The iterator excludes the empty multiset (all zeros) and the full multiset
// (all maxExponents).
func newSubMultisetIterator(exponents []int) *subMultisetIterator {
	current := make([]int, len(exponents))
	return &subMultisetIterator{
		maxExponents:     exponents,
		currentExponents: current,
		exhausted:        false,
	}
}

// next advances the iterator and returns false when exhausted.
func (it *subMultisetIterator) next() bool {
	if it.exhausted {
		return false
	}

	// Increment with carry propagation (odometer-style)
	for i := 0; i < len(it.currentExponents); i++ {
		if it.currentExponents[i] < it.maxExponents[i] {
			it.currentExponents[i]++
			return it.skipTrivialMultisets()
		}
		it.currentExponents[i] = 0
	}

	it.exhausted = true
	return false
}

// skipTrivialMultisets advances past empty and full multisets.
func (it *subMultisetIterator) skipTrivialMultisets() bool {
	for it.isEmpty() || it.isFull() {
		if !it.advanceOnce() {
			return false
		}
	}
	return true
}

// advanceOnce performs a single increment step.
func (it *subMultisetIterator) advanceOnce() bool {
	for i := 0; i < len(it.currentExponents); i++ {
		if it.currentExponents[i] < it.maxExponents[i] {
			it.currentExponents[i]++
			return true
		}
		it.currentExponents[i] = 0
	}
	it.exhausted = true
	return false
}

// isEmpty returns true if current multiset has total exponent zero.
func (it *subMultisetIterator) isEmpty() bool {
	for _, e := range it.currentExponents {
		if e != 0 {
			return false
		}
	}
	return true
}

// isFull returns true if current equals maxExponents.
func (it *subMultisetIterator) isFull() bool {
	for i, e := range it.currentExponents {
		if e != it.maxExponents[i] {
			return false
		}
	}
	return true
}

// currentTotalDegree returns the sum of current exponents.
func (it *subMultisetIterator) currentTotalDegree() int {
	sum := 0
	for _, e := range it.currentExponents {
		sum += e
	}
	return sum
}

// currentSnapshot returns a copy of the current exponents.
func (it *subMultisetIterator) currentSnapshot() []int {
	snapshot := make([]int, len(it.currentExponents))
	copy(snapshot, it.currentExponents)
	return snapshot
}
