package query

import (
	"errors"
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/fiatshamir"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/collection"
	"github.com/consensys/gnark/frontend"
)

// Represent a batch of inner-product <a, b0>, <a, b1>, <a, b2> ...
type InnerProduct struct {
	A  ifaces.Column
	Bs []ifaces.Column
	ID ifaces.QueryID
}

// Inner product params
type InnerProductParams struct {
	Ys []field.Element
}

// Update the fiat-shamir state with inner-product params
func (ipp InnerProductParams) UpdateFS(state *fiatshamir.State) {
	state.UpdateVec(ipp.Ys)
}

// Constructor for inner-product.
// The list of polynomial Bs must be deduplicated.
func NewInnerProduct(id ifaces.QueryID, a ifaces.Column, bs ...ifaces.Column) InnerProduct {
	// Panics if there is a duplicate
	bsSet := collection.NewSet[ifaces.ColID]()

	if len(bs) == 0 {
		utils.Panic("Inner-product %v declared without bs", id)
	}

	length := a.Size()

	for _, b := range bs {

		if b.Size() != length {
			utils.Panic("bad size for %v, expected %v but got %v", b.GetColID(), b.Size(), length)
		}

		if len(b.GetColID()) == 0 {
			utils.Panic("Assigned a polynomial ifaces.QueryID with an empty length")
		}

		if bsSet.Insert(b.GetColID()) {
			utils.Panic("(query %v) Got a duplicate entry %v in %v\n", id, b, bs)
		}
	}

	return InnerProduct{ID: id, A: a, Bs: bs}
}

// Constructor for fixed point univariate evaluation query parameters
func NewInnerProductParams(ys ...field.Element) InnerProductParams {
	return InnerProductParams{Ys: ys}
}

// Check the inner-product manually
func (r InnerProduct) Check(run ifaces.Runtime) error {

	wA := r.A.GetColAssignment(run)
	expecteds := run.GetParams(r.ID).(InnerProductParams)

	// Prepare a nice error message in case we need it
	errMsg := fmt.Sprintf("Inner-product %v\n", r.ID)
	errorFlag := false

	for i, b := range r.Bs {
		wB := b.GetColAssignment(run)
		mul := smartvectors.Mul(wA, wB)

		// Compute manually the inner-product of two witnesses
		actualIP := field.Zero()
		for i := 0; i < mul.Len(); i++ {
			tmp := mul.Get(i)
			actualIP.Add(&actualIP, &tmp)
		}

		if expecteds.Ys[i] != actualIP {
			errorFlag = true
			errMsg = fmt.Sprintf("%v\tFor witness <%v, %v> the alleged value is %v but the correct value is %v\n",
				errMsg, r.A.GetColID(), b.GetColID(), expecteds.Ys[i].String(), actualIP.String(),
			)
		}
	}

	if errorFlag {
		return errors.New(errMsg)
	}

	return nil
}

// Check the inner-product manually
func (r InnerProduct) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {

	wA := r.A.GetColAssignmentGnark(run)
	expecteds := run.GetParams(r.ID).(GnarkInnerProductParams)

	for i, b := range r.Bs {
		wB := b.GetColAssignmentGnark(run)

		// mul <- \sum_j wA * wB
		actualIP := frontend.Variable(0)
		for j := range wA {
			tmp := api.Mul(wA[j], wB[j])
			actualIP = api.Add(actualIP, tmp)
		}

		api.AssertIsEqual(expecteds.Ys[i], actualIP)
	}
}
