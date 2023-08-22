package query

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark/frontend"
	"github.com/sirupsen/logrus"
)

/*
	Enforces a fixedPermutation constraint, that  two handles must
	be a fixedpermutation of eachother

*/
//fix permutation over splittings of two vectors
/* here vectors A_i,B_i are splitting of target vectors A*,B*
where a fixed permutation is applied between target vectors
i.e.,A* = A0||...||An, B* = B0||...||Bn and B* = s(A*) for the fixed permutation s */
type FixedPermutation struct {
	ID ifaces.QueryID
	//splittings
	A, B []ifaces.Column
	/*
		the permutation 's' also can be defined by a set of splittings s_i and s_{id,i} where
		s(s_{id,i})=s_i. The columns s_{id,i} are called identity polynomials of permutation 's' and  are known by defult.
		So that the permutation can be determined only by the splittings 's_i'
	*/

	//fixed  permutation
	S []ifaces.ColAssignment
}

/*
Constructor for fixedPermutation constraints also makes the input validation
*/
func NewFixedPermutation(id ifaces.QueryID, S []ifaces.ColAssignment, a, b []ifaces.Column) FixedPermutation {
	/*
		Both side of the permutation must have the same number of columns
	*/
	if len(a) != len(b) || len(a) != len(S) {
		utils.Panic("a , b ,S don't have the same number of splittings %v %v %v", len(a), len(b), len(S))
	}

	// All polynomials must have the same MaxSize
	_, err := utils.AllReturnEqual(
		ifaces.Column.Size,
		append(a, b...))
	if err == nil {
		// a,b should have the same number of rows as S
		if S[0].Len() != a[0].Size() {
			logrus.Errorf("S and 'a' dont have the same number of rows: %v, %v", S[0].Len(), a[0].Size())
		}
	}

	if err != nil {

		for i := range a {
			logrus.Errorf("size of the column %v of a : %v\n", i, a[i].Size())
		}

		for i := range b {
			logrus.Errorf("size of the column %v of b : %v\n", i, b[i].Size())
		}

		utils.Panic("Permutation (%v) requires that all columns have the same size %v", id, err)
	}
	for i := range S {
		if S[i].Len() != S[0].Len() {
			utils.Panic("%v-th splittings of S do not have the right length", i)
		}
	}

	// recall that 'a' and 'b' have the same size
	for i := range a {
		a[i].MustExists()
		b[i].MustExists()
	}

	return FixedPermutation{
		ID: id,
		A:  a,
		B:  b,
		S:  S,
	}
}

/*
 */
func (r FixedPermutation) Check(run ifaces.Runtime) error {
	/*
		They should have the same size and it should be tested
		prior to calling check
	*/
	a := make([]ifaces.ColAssignment, len(r.B))
	b := make([]ifaces.ColAssignment, len(r.A))

	// Populate the `a`
	for i, pol := range r.A {
		a[i] = pol.GetColAssignment(run)
	}

	// Populate the `b`
	for i, pol := range r.B {
		b[i] = pol.GetColAssignment(run)
	}

	return CheckFixedPermutation(a, b, r.S)
}

/*
Checks a fixedpermutation query manually.
*/
func CheckFixedPermutation(a, b []ifaces.ColAssignment, S []ifaces.ColAssignment) error {
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
	S_id := make([]ifaces.ColAssignment, len(S))
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
	var tmp fr.Element
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
func (f FixedPermutation) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {
	panic("UNSUPPORTED : can't check an inclusion query directly into the circuit")
}
