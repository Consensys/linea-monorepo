package wizard

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

// Enforces a fixedPermutation constraint, that  two handles must
// be a fixedpermutation of eachother
//
// the permutation 's' also can be defined by a set of splittings s_i and s_{id,i} where
// s(s_{id,i})=s_i. The columns s_{id,i} are called identity polynomials of permutation 's' and  are known by defult.
// So that the permutation can be determined only by the splittings 's_i'
type QueryFixedPermutation struct {
	// A and B are the two side subjected to the fixed-permutation constraint.
	//
	// Here vectors A_i,B_i are splitting of target vectors A*,B*
	// where a fixed permutation is applied between target vectors
	// i.e.,A* = A0||...||An, B* = B0||...||Bn and B* = s(A*) for the fixed permutation s */
	A, B []Column
	// S is the set of vectors encoding the fixed-permutation
	S        []smartvectors.SmartVector
	metadata *metadata
	*subQuery
}

/*
Constructor for fixedPermutation constraints also makes the input validation
*/
func (api *API) NewQueryFixedPermutation(s []smartvectors.SmartVector, a, b []Column) QueryFixedPermutation {

	// Both side of the permutation must have the same number of columns
	if len(a) != len(b) || len(a) != len(s) {
		utils.Panic("a , b ,S don't have the same number of splittings %v %v %v", len(a), len(b), len(s))
	}

	// All polynomials must have the same MaxSize
	_, err := utils.AllReturnEqual(
		Column.Size,
		append(a, b...))
	if err == nil {
		// a,b must have the same number of rows as S
		if s[0].Len() != a[0].Size() {
			utils.Panic("S and 'a' dont have the same number of rows: %v, %v", s[0].Len(), a[0].Size())
		}
	}

	if err != nil {

		for i := range a {
			logrus.Errorf("size of the column %v of a : %v\n", i, a[i].Size())
		}

		for i := range b {
			logrus.Errorf("size of the column %v of b : %v\n", i, b[i].Size())
		}

		utils.Panic("Permutation requires that all columns have the same size %v", err)
	}

	for i := range s {
		if s[i].Len() != s[0].Len() {
			utils.Panic("%v-th splittings of S do not have the right length", i)
		}
	}

	// recall that 'a' and 'b' have the same size
	round := 0
	for i := range a {
		a[i].Round()
		b[i].Round()
	}

	return QueryFixedPermutation{
		A: a,
		B: b,
		S: s,
		subQuery: &subQuery{
			round: round,
		},
		metadata: api.newMetadata(),
	}
}

// Check implements the [Query] interface
func (r QueryFixedPermutation) Check(run Runtime) error {
	/*
		They should have the same size and it should be tested
		prior to calling check
	*/
	a := make([]smartvectors.SmartVector, len(r.B))
	b := make([]smartvectors.SmartVector, len(r.A))

	// Populate the `a`
	for i, pol := range r.A {
		a[i] = pol.GetAssignment(run)
	}

	// Populate the `b`
	for i, pol := range r.B {
		b[i] = pol.GetAssignment(run)
	}

	return CheckQueryFixedPermutation(a, b, r.S)
}

/*
Checks a fixedpermutation query manually.
*/
func CheckQueryFixedPermutation(a, b []smartvectors.SmartVector, S []smartvectors.SmartVector) error {
	/*
		Sample two  random elements alpha, beta ,alpha is used for LC
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
	if len(a) != len(b) || len(a) != len(S) {
		utils.Panic("Not the same number of columns %v %v %v", len(a), len(b), len(S))
	}

	nRow := a[0].Len()
	/*
		Sanity-check, all sample should have the same number of rows.
	*/
	for i := range a {
		if a[i].Len() != nRow {
			utils.Panic("Row %v of a has an inconsistent size. Expected %v but got %v", i, nRow, a[i].Len())
		}
		if b[i].Len() != nRow {
			utils.Panic("Row %v of b has an inconsistent size. Expected %v but got %v", i, nRow, b[i].Len())
		}
		if S[i].Len() != nRow {
			utils.Panic("Row %v of S has an inconsistent size. Expected %v but got %v", i, nRow, S[i].Len())
		}
	}

	//generate identity polys of permutation; S_id
	S_id := make([]smartvectors.SmartVector, len(S))
	n := S[0].Len()
	for j := range S {
		identity := make([]field.Element, n)
		for i := 0; i < n; i++ {
			identity[i] = field.NewElement(uint64(n*j + i))
		}
		S_id[j] = smartvectors.NewRegular(identity)
	}

	prodA := field.One()
	prodB := field.One()
	var tmp field.Element
	for j := range a {
		for i := 0; i < nRow; i++ {
			//for a
			u := a[j].Get(i)
			v := S_id[j].Get(i)
			tmp.Mul(&alpha, &v).Add(&u, &tmp)
			tmp.Add(&tmp, &beta)
			prodA.Mul(&prodA, &tmp)

			//for b
			u = b[j].Get(i)
			v = S[j].Get(i)
			tmp.Mul(&alpha, &v).Add(&u, &tmp)
			tmp.Add(&tmp, &beta)
			prodB.Mul(&prodB, &tmp)
		}
	}

	// At the end, the two product should be equals
	if prodA != prodB {
		return fmt.Errorf("the permutation check rejected")
	}

	return nil
}

// GnarkCheck will panic in this construction because we do not have a good way
// to check the query within a circuit
func (f QueryFixedPermutation) CheckGnark(api frontend.API, run GnarkRuntime) {
	panic("UNSUPPORTED : can't check an inclusion query directly into the circuit")
}
