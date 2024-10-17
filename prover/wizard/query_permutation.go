package wizard

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

// QueryPermutation is a predicate that assess that two tables contains the same rows
// up to a permutation. The tables can contain several columns and they can be
// described in a fractioned way: the table is the union of the rows of several
// tables.
type QueryPermutation struct {
	// A and B represent the tables on both sides of the argument. The
	// permutation can be fractionned (len(A) = len(B) > 1) and it can be
	// multi-column (len(A[*]) = len(B[*]) > 1.
	A, B     [][]Column
	metadata *metadata
	*subQuery
}

// NewPermutation constructs a new permutation query and performs all the
// well-formedness sanity-checks. In case of failure, it will panic.
func (api *API) NewPermutation(a, b [][]Column) *QueryPermutation {

	var (
		nCol     = len(a[0])
		totalRow = [2]int{}
		round    = 0
	)

	for side, aOrB := range [2][][]Column{a, b} {
		for frag := range aOrB {

			if len(aOrB[frag]) != nCol {
				utils.Panic("all tables must have the same number of columns")
			}

			for _, col := range aOrB[frag] {
				round = max(round, col.Round())
			}

			sizeFrag, err := utils.AllReturnEqual(
				Column.Size,
				aOrB[frag],
			)

			if err != nil {
				utils.Panic("all tables must be sets of columns with the same size")
			}

			totalRow[side] += sizeFrag
		}
	}

	if totalRow[0] != totalRow[1] {
		utils.Panic("a and b must have the same total number of rows")
	}

	res := &QueryPermutation{
		A:        a,
		B:        b,
		metadata: api.newMetadata(),
		subQuery: &subQuery{round: round},
	}

	api.queries.addToRound(round, res)
	return res
}

// Check probabilistically checks whether the permutation predicates holds. The
// test works by incrementally computes the products:
//
//	\prod_i (\gamma + \sum_j C{i, j} \alpha^j)
//
// With overhelming probability, if the predicate is wrong then then the
// products will be unequal and this will be equal if the predicate is
// satisfied.
func (r QueryPermutation) Check(run Runtime) error {

	var (
		numCol    = len(r.A[0])
		prods     = []field.Element{field.One(), field.One()}
		randGamma = field.Element{}
		randAlpha = field.Element{}
	)

	randGamma.SetRandom()
	randAlpha.SetRandom()

	for k, aOrB := range [2][][]Column{r.A, r.B} {
		for frag := range aOrB {
			var (
				tab        = make([]smartvectors.SmartVector, numCol)
				numRowFrag = aOrB[frag][0].Size()
				gamma      = smartvectors.NewConstant(randGamma, numRowFrag)
			)

			for col := range aOrB[frag] {
				tab[col] = aOrB[frag][col].GetAssignment(run)
			}

			collapsed := smartvectors.PolyEval(append(tab, gamma), randAlpha)

			for row := 0; row < collapsed.Len(); row++ {
				tmp := collapsed.Get(row)
				prods[k].Mul(&prods[k], &tmp)
			}
		}
	}

	if prods[0] != prods[1] {
		return fmt.Errorf("the permutation query %v is not satisfied", r.String())
	}

	return nil
}

// CheckPermutation manually checks that a permutation argument is satisfied.
func CheckPermutation(a, b []smartvectors.SmartVector) error {
	/*
		Sample a random element alpha, usefull for multivalued inclusion checks
		It allows to reference multiple number through a linear combination
	*/
	var alpha, beta field.Element
	_, err := alpha.SetRandom()
	_, err2 := beta.SetRandom()
	if err != nil || err2 != nil {
		utils.Panic("Could not generate a random number %v %v", err, err2)
	}

	/*
		Sanity-check both sides should have the same number of cols
	*/
	if len(a) != len(b) {
		utils.Panic("Not the same number of columns %v %v", len(a), len(b))
	}

	nRow := a[0].Len()
	/*
		Sanity-check, all sample should have the same number of rows.
		This might become an error later, but this is easy to change.
	*/
	for i := range a {
		if a[i].Len() != nRow {
			utils.Panic("Row %v of a has an inconsistent size. Expected %v but got %v", i, nRow, a[i].Len())
		}
		if b[i].Len() != nRow {
			utils.Panic("Row %v of b has an inconsistent size. Expected %v but got %v", i, nRow, b[i].Len())
		}
	}
	if nRow != b[0].Len() {
		return fmt.Errorf("a and b do not have the same length : %v != %v", a[0].Len(), b[0].Len())
	}

	prodA := field.One()
	prodB := field.One()

	for i := 0; i < nRow; i++ {
		// The product for a
		tmp := rowLinComb(alpha, i, a)
		tmp.Add(&tmp, &beta)
		prodA.Mul(&prodA, &tmp)

		// The product for b
		tmp = rowLinComb(alpha, i, b)
		tmp.Add(&tmp, &beta)
		prodB.Mul(&prodB, &tmp)
	}

	// At the end, the two product should be equals
	if prodA != prodB {
		return fmt.Errorf("the permutation check rejected")
	}

	return nil
}

// GnarkCheck will panic in this construction because we do not have a good way
// to check the query within a circuit
func (p QueryPermutation) CheckGnark(api frontend.API, run RuntimeGnark) {
	panic("UNSUPPORTED : can't check an permutation query directly into the circuit")
}
