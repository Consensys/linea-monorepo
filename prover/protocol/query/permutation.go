package query

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/gnark/frontend"
	"github.com/sirupsen/logrus"
)

// Test that two table are row-permutations of one another
type Permutation struct {
	A, B []ifaces.Column
	ID   ifaces.QueryID
}

/*
Construct a permutation. Will panic if it is mal-formed
*/
func NewPermutation(id ifaces.QueryID, a, b []ifaces.Column) Permutation {
	/*
		Both side of the permutation must have the same number of columns
	*/
	if len(a) != len(b) {
		utils.Panic("a and b_ don't have the same number of commitments %v %v", len(a), len(b))
	}

	// All polynomials must have the same MaxSize
	_, err := utils.AllReturnEqual(
		ifaces.Column.Size,
		append(a, b...),
	)

	if err != nil {

		for i := range a {
			logrus.Errorf("size of the column %v of a : %v\n", i, a[i].Size())
		}

		for i := range b {
			logrus.Errorf("size of the column %v of b : %v\n", i, b[i].Size())
		}

		utils.Panic("Permutation (%v) requires that all columns have the same size %v", id, err)
	}

	// recall that a and b have the same size
	for i := range a {
		a[i].MustExists()
		b[i].MustExists()
	}

	return Permutation{A: a, B: b, ID: id}
}

/*
Test that the polynomial evaluation holds
It's a probabilistic check - good enough for testing
*/
func (r Permutation) Check(run ifaces.Runtime) error {
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

	// Populate the b
	for i, pol := range r.B {
		b[i] = pol.GetColAssignment(run)
	}

	return CheckPermutation(a, b)
}

/*
Checks a permutation query manually. The test is probabilistic.
The soundness is only appropriate for testing purposes.
*/
func CheckPermutation(a, b []ifaces.ColAssignment) error {
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
func (p Permutation) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {
	panic("UNSUPPORTED : can't check an inclusion query directly into the circuit")
}
