package query

import (
	"errors"
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors_mixed"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
	"github.com/google/uuid"
)

// Represent a batch of inner-product <a, b0>, <a, b1>, <a, b2> ...
type InnerProductGen[T zk.Element] struct {
	A    ifaces.Column[T]
	Bs   []ifaces.Column[T]
	ID   ifaces.QueryID
	uuid uuid.UUID `serde:"omit"`
}

// Constructor for inner-product.
// The list of polynomial Bs must be deduplicated.
func NewInnerProductGen[T zk.Element](id ifaces.QueryID, a ifaces.Column[T], bs ...ifaces.Column[T]) InnerProductGen[T] {
	// Panics if there is a duplicate
	bsSet := collection.NewSet[ifaces.ColID]()

	if len(bs) == 0 {
		utils.Panic("Inner-product %v declared without bs", id)
	}

	a.MustExists()
	length := a.Size()

	for _, b := range bs {

		b.MustExists()
		if b.Size() != length {
			utils.Panic("bad size for %v, expected %v but got %v", b.GetColID(), b.Size(), length)
		}

		if len(b.GetColID()) == 0 {
			utils.Panic("Assigned a polynomial ifaces.QueryID with an empty length")
		}

		// Note: we used to check and panic for duplicates in the Horner
		// query.
		bsSet.Insert(b.GetColID())
	}

	return InnerProductGen[T]{ID: id, A: a, Bs: bs, uuid: uuid.New()}
}

// Name implements the [ifaces.Query] interface
func (r InnerProductGen[T]) Name() ifaces.QueryID {
	return r.ID
}

// Check the inner-product manually
func (r InnerProductGen[T]) Check(run ifaces.Runtime) error {

	expecteds := run.GetParams(r.ID).(InnerProductParams)
	computed := r.Compute(run)

	// Prepare a nice error message in case we need it
	errMsg := fmt.Sprintf("Inner-product %v\n", r.ID)
	errorFlag := false

	for i := range computed {

		if expecteds.Ys[i] != computed[i] {
			errorFlag = true
			errMsg = fmt.Sprintf("%v\tFor witness <%v, %v> the alleged value is %v but the correct value is %v\n",
				errMsg, r.A.GetColID(), r.Bs[i].GetColID(), expecteds.Ys[i].String(), computed[i].String(),
			)
		}
	}

	if errorFlag {
		return errors.New(errMsg)
	}

	return nil
}

func (r InnerProductGen[T]) Compute(run ifaces.Runtime) []fext.Element {

	res := make([]fext.Element, len(r.Bs))
	a := r.A.GetColAssignment(run)
	a = smartvectors_mixed.LiftToExt(a)

	for i := range r.Bs {

		b := r.Bs[i].GetColAssignment(run)
		b = smartvectors_mixed.LiftToExt(b)
		ab := smartvectors_mixed.Mul(a, b)
		res[i] = smartvectors.SumExt(ab)
	}

	return res
}

// Check the inner-product manually
func (r InnerProductGen[T]) CheckGnark(api frontend.API, run ifaces.GnarkRuntime[T]) {

	apiGen, err := zk.NewApi[T](api)
	if err != nil {
		panic(err)
	}

	wA := r.A.GetColAssignmentGnark(run)
	expected := run.GetParams(r.ID).(GnarkInnerProductParamsGen[T])

	for i, b := range r.Bs {
		wB := b.GetColAssignmentGnark(run)

		// mul <- \sum_j wA * wB
		actualIP := apiGen.FromUint(0)
		for j := range wA {
			tmp := apiGen.Mul(&wA[j], &wB[j])
			actualIP = apiGen.Add(actualIP, tmp)
		}

		apiFext, err := gnarkfext.NewExt4[T](api)
		if err != nil {
			panic(err)
		}

		var actualIPExt gnarkfext.E4Gen[T]
		actualIPExt.B0.A0 = *zk.ValueOf[T](actualIP)
		apiFext.AssertIsEqual(&expected.Ys[i], &actualIPExt)
	}
}

func (r InnerProductGen[T]) UUID() uuid.UUID {
	return r.uuid
}
