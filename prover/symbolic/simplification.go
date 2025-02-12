package symbolic

import (
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
)

// regroupTerms takes a list of expressions and magnitudes and regroups the
// children sharing the same ESH, adding their magnitudes. This is meant for
// expression simplification. It also returns separately the constants
// term's values and their magnitudes if any are found. The function does not
// attempt to regroup the constants it finds.
//
// The function requires that all the provided children have their ESH set
// correctly. Without it will fall in undetermined behaviors. When, the function
// encounters two children sharing the same ESH but not the same structure, it
// keeps the first one in the list.
func regroupTerms(magnitudes []int, children []*Expression) (
	regroupedMagnitudes []int,
	regroupedChildren []*Expression,
	constantMagnitudes []int,
	constantValues []fext.Element,
) {

	if len(magnitudes) != len(children) {
		panic("magnitudes and children don't have the same length")
	}

	numChildren := len(children)
	foundExpressions := make(map[fext.Element]int, numChildren)
	regroupedChildren = make([]*Expression, 0, numChildren)
	regroupedMagnitudes = make([]int, 0, numChildren)
	constantValues = make([]fext.Element, 0, numChildren)
	constantMagnitudes = make([]int, 0, numChildren)

	for i := 0; i < numChildren; i++ {

		var (
			c       *Constant
			isConst bool
		)

		switch o := children[i].Operator.(type) {
		case Constant:
			c = &o
			isConst = true
		case *Constant:
			c = o
			isConst = true
		}

		if isConst {
			constantMagnitudes = append(constantMagnitudes, magnitudes[i])
			constantValues = append(constantValues, c.Val)
			continue
		}

		if pos, found := foundExpressions[children[i].ESHash]; found {
			regroupedMagnitudes[pos] += magnitudes[i]
			continue
		}

		// The current term cannot be regrouped, we appended to the list of
		// matchable expressions
		newPos := len(regroupedChildren)
		regroupedChildren = append(regroupedChildren, children[i])
		regroupedMagnitudes = append(regroupedMagnitudes, magnitudes[i])
		foundExpressions[children[i].ESHash] = newPos
	}

	return regroupedMagnitudes,
		regroupedChildren,
		constantMagnitudes,
		constantValues

}

// removeZeroCoeffs "cleans" by removing the zero coefficients parents terms in
// the linear combination. This function is used both for simplifying [LinComb]
// expressions and for simplifying [Product]. "magnitude" denotes either the
// coefficient for LinComb or exponents for Product.
func removeZeroCoeffs(magnitudes []int, children []*Expression) (cleanMagnitudes []int, cleanChildren []*Expression) {

	if len(magnitudes) != len(children) {
		panic("magnitudes and children don't have the same length")
	}

	cleanChildren = make([]*Expression, 0, len(children))
	cleanMagnitudes = make([]int, 0, len(children))

	for i, c := range magnitudes {
		if c != 0 {
			cleanMagnitudes = append(cleanMagnitudes, magnitudes[i])
			cleanChildren = append(cleanChildren, children[i])
		}
	}

	return cleanMagnitudes, cleanChildren
}

// expandLinComb takes a list of inputs [Expression] and magnitudes destined
// to serve as parameter to build either a [LinComb] or a [Product]. It returns
// an expanded list of inputs that builds the same expression "without the
// parenthesis". This is meant to enable later simplifications.
//
// Here, the name "magnitude" is coined to denote either the coefficients of
// a linear combinations or the exponents in a product.
//
// The caller passes a target operator which may be any value of type either
// [LinComb] or [Product]. Any other type yields a panic error.
func expandTerms(op Operator, magnitudes []int, children []*Expression) (
	expandedMagnitudes []int,
	expandedExpression []*Expression,
) {

	var (
		opIsProd    bool
		opIsLinC    bool
		numChildren = len(children)
	)

	switch op.(type) {
	case LinComb, *LinComb:
		opIsLinC = true
	case Product, *Product:
		opIsProd = true
	}

	if !(opIsProd || opIsLinC) {
		panic("wrong operator type")
	}

	if len(magnitudes) != numChildren {
		panic("incompatible number of children and magnitudes")
	}

	// The capacity allocation is purely heuristic
	expandedExpression = make([]*Expression, 0, 2*len(magnitudes))
	expandedMagnitudes = make([]int, 0, 2*len(magnitudes))

	for i := 0; i < numChildren; i++ {

		var (
			child     = children[i]
			magnitude = magnitudes[i]
			cProd     *Product
			cLinC     *LinComb
			cIsProd   bool
			cIsLinC   bool
		)

		switch o := child.Operator.(type) {
		case Product:
			cIsProd = true
			cProd = &o
		case *Product:
			cIsProd = true
			cProd = o
		case LinComb:
			cIsLinC = true
			cLinC = &o
		case *LinComb:
			cIsLinC = true
			cLinC = o
		}

		if cIsProd && opIsProd {
			for k := range child.Children {
				expandedExpression = append(expandedExpression, child.Children[k])
				expandedMagnitudes = append(expandedMagnitudes, magnitude*cProd.Exponents[k])
			}
			continue
		}

		if cIsLinC && opIsLinC {
			for k := range child.Children {
				expandedExpression = append(expandedExpression, child.Children[k])
				expandedMagnitudes = append(expandedMagnitudes, magnitude*cLinC.Coeffs[k])

			}
			continue
		}

		expandedExpression = append(expandedExpression, child)
		expandedMagnitudes = append(expandedMagnitudes, magnitude)
	}

	return expandedMagnitudes, expandedExpression
}
