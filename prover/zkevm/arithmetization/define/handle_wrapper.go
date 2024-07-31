package define

import (
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

/*
Handle wraps commitment.Handle as an Option, the negative case of the
option is represented by a `nil` pointer on inner and is used to -
represent a Handle that is to be ignored by the system.
*/
type Handle struct {
	/*
		The fact that we have a pointer here, ensure it will
		panic before getting inside the wizard.Builder. This is a mechanism
		to protect against propagating invalid `Handle` objects and thus,
		having hard-to-fix bugs.
	*/
	inner *ifaces.Column
	// This field is set when the user calls Interleave
	interleavedComponents []ifaces.Column
}

/*
Returns the nil-case value of a handle
*/
func NilHandle() Handle {
	return Handle{}
}

/*
Returns true if this is the nil case
*/
func (h Handle) IsNil() bool {
	return h.inner == nil
}

/*
Returns the wrapped Handle, and properly panic if that did not work.
*/
func (h Handle) Unwrap() ifaces.Column {
	if h.IsNil() {
		utils.Panic("Poisoned handle")
	}
	return *h.inner
}

/*
Shift the handle or propagate the nil-case
*/
func (h Handle) Shift(offset int) Handle {
	if h.IsNil() {
		// return the same `nil` handle
		return h
	}
	innerNew := column.Shift(*h.inner, offset)
	return Handle{inner: &innerNew}
}

// interleaved is a legacy type. It can only be used as input of a fragmented
// permutation query.
type interleaved struct {
	// this is basically left as empty and make the interleaved satisfy the
	// interface.
	ifaces.Column
	inner []ifaces.Column
}

/*
Create an interleaving of multiple commitment or propagate the nil-case

Will also return a nil-case if the number of
*/
func Interleave(parents ...Handle) Handle {

	if !utils.IsPowerOfTwo(len(parents)) {
		logrus.Warnf("Skipping interleaved because it has %v parents, which is not a power of two", len(parents))
		return NilHandle()
	}

	parentsInner := []ifaces.Column{}
	for _, h := range parents {
		if h.IsNil() {
			return NilHandle()
		}
		parentsInner = append(parentsInner, *h.inner)
	}

	return Handle{interleavedComponents: parentsInner}
}

/*
Converts into a variable. In the nil-case, it becomes a `nil` in-place of
the metadata.
*/
func (h Handle) AsVariable() *symbolic.Expression {
	if h.IsNil() {
		return nil
	}
	return ifaces.ColumnAsVariable(*h.inner)
}
