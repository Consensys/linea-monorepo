package query

import (
	"errors"
	"fmt"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectorsext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext/gnarkfext"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
)

// Represent a batch of inner-product <a, b0>, <a, b1>, <a, b2> ...
type InnerProduct struct {
	A      ifaces.Column
	Bs     []ifaces.Column
	ID     ifaces.QueryID
	isBase bool
}

// Inner product params
type InnerProductParams struct {
	baseYs []field.Element
	extYs  []fext.Element
	isBase bool
}

// Update the fiat-shamir state with inner-product params
func (ipp InnerProductParams) UpdateFS(state *fiatshamir.State) {
	if ipp.isBase {
		state.UpdateVec(ipp.baseYs)
	} else {
		// update this with the proper field extension later
		tempSlice := make([]field.Element, 2*len(ipp.baseYs))
		for i := range ipp.extYs {
			tempSlice[2*i] = ipp.extYs[i].A0
			tempSlice[2*i+1] = ipp.extYs[i].A1
		}
		state.Update(tempSlice...)
	}
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
	return InnerProductParams{
		baseYs: ys,
		extYs:  nil,
		isBase: true,
	}
}

func NewInnerProductParamsExt(ys ...fext.Element) InnerProductParams {
	return InnerProductParams{
		baseYs: nil,
		extYs:  ys,
		isBase: false,
	}
}

// Name implements the [ifaces.Query] interface
func (r InnerProduct) Name() ifaces.QueryID {
	return r.ID
}

// Check the inner-product manually
func (r InnerProduct) Check(run ifaces.Runtime) error {

	wA := r.A.GetColAssignment(run)
	expecteds := run.GetParams(r.ID).(InnerProductParams)

	// Prepare a nice error message in case we need it
	errMsg := fmt.Sprintf("Inner-product %v\n", r.ID)
	errorFlag := false

	if expecteds.isBase {
		for i, b := range r.Bs {
			wB := b.GetColAssignment(run)
			mul := smartvectors.Mul(wA, wB)

			// Compute manually the inner-product of two witnesses
			actualIP := field.Zero()
			for i := 0; i < mul.Len(); i++ {
				tmp := mul.Get(i)
				actualIP.Add(&actualIP, &tmp)
			}

			if expecteds.baseYs[i] != actualIP {
				errorFlag = true
				errMsg = fmt.Sprintf("%v\tFor witness <%v, %v> the alleged value is %v but the correct value is %v\n",
					errMsg, r.A.GetColID(), b.GetColID(), expecteds.baseYs[i].String(), actualIP.String(),
				)
			}
		}
	} else {
		for i, b := range r.Bs {
			wB := b.GetColAssignment(run)
			mul := smartvectorsext.Mul(wA, wB)

			// Compute manually the inner-product of two witnesses
			actualIP := fext.Zero()
			for i := 0; i < mul.Len(); i++ {
				tmp := mul.GetExt(i)
				actualIP.Add(&actualIP, &tmp)
			}

			if expecteds.extYs[i] != actualIP {
				errorFlag = true
				errMsg = fmt.Sprintf("%v\tFor witness <%v, %v> the alleged value is %v but the correct value is %v\n",
					errMsg, r.A.GetColID(), b.GetColID(), expecteds.extYs[i].String(), actualIP.String(),
				)
			}
		}

	}

	if errorFlag {
		return errors.New(errMsg)
	}

	return nil
}

// Check the inner-product manually
func (r InnerProduct) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {

	expecteds := run.GetParams(r.ID).(GnarkInnerProductParams)

	if expecteds.IsBase {
		wA := r.A.GetColAssignmentGnark(run)
		for i, b := range r.Bs {
			wB := b.GetColAssignmentGnark(run)

			// mul <- \sum_j wA * wB
			actualIP := frontend.Variable(0)
			for j := range wA {
				tmp := api.Mul(wA[j], wB[j])
				actualIP = api.Add(actualIP, tmp)
			}

			api.AssertIsEqual(expecteds.BaseYs[i], actualIP)
		}
	} else {
		wA := r.A.GetColAssignmentGnarkExt(run)
		extApi := gnarkfext.API{Inner: api}
		for i, b := range r.Bs {
			wB := b.GetColAssignmentGnarkExt(run)

			// mul <- \sum_j wA * wB
			actualIP := gnarkfext.NewZero()
			for j := range wA {
				tmp := extApi.Mul(wA[j], wB[j])
				actualIP = extApi.Add(actualIP, tmp)
			}

			api.AssertIsEqual(expecteds.ExtYs[i], actualIP)
		}
	}

}
