package define

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

/*
/!\ This should be an internal function of the package. The user should be using
the Arithmetization struct which is a public wrapper that hides all the
internal methods that are onyl relevant to integrating the zkevm-arithmetization
constraints with this repo. We still export it because we need to refer to this
package in the define subpackage. And we cannot merge this package with the
define sub-package because of how massive the define sub-package is.

Builder is an adapter for wizard.Builder. It is responsible for filtering
out commitment whose size is non-power of two.

It is especially helpful to be able to filter out.
*/
type Builder struct {
	internal *wizard.Builder
	// This variable is used internally to ensure that no two constraints have
	// the same name. It does not need to be initialized.
	deduplicator int
	*Settings
}

// Calls the define function over the corresponding to the define traces. E.g,
// instantiate all the columns and constraints in the arithmetization. The
// settings of the builder **must** be already given when this function is
// called. Otherwise, it is a panic.
func (b *Builder) Define(wb *wizard.Builder) {

	if b.Settings == nil {
		panic("called Define, but the settings have not been set")
	}

	// The field is ultimately destroyed as it should not be referenced again
	// after the function exits. This would lead to undefined behavior.
	// Ultimately, only `wb` should be mutated at the end of this call.
	b.internal = wb
	defer func() { b.internal = nil }()

	// Call the corset-generated set of columns and constraints over the builder.
	ZkEVMDefine(b)
}

/*
Registers a new commitment in the protocol
*/
func (b *Builder) RegisterCommit(name ifaces.ColID, size int) Handle {

	/*
		Limit the number of commitments
	*/
	if b.NumColLimit > 0 && len(b.internal.ListCommitments()) >= b.NumColLimit {
		logrus.Tracef("Ignoring %v because there is already more than %v commitments", name, b.NumColLimit)
		return NilHandle()
	}

	/*
		Reject the commitment if the size of the commitment is not right
	*/
	if !utils.IsPowerOfTwo(size) {
		logrus.Tracef("Skipping the registeration of %v because its size is a non-power of two %v\n", name, size)
		/*
			The empty handle is how we detect it will not work. The downside of
			it is that the adapter is forced to detect and check all occurence
			of poisoned handles before passing to the underlying wizard.Builder.
		*/
		return NilHandle()
	}

	/*
		Artificially limit the size of the commitment
	*/
	if b.ColDepthLimit > 0 && size > b.ColDepthLimit {
		size = b.ColDepthLimit
	}

	resInner := b.internal.RegisterCommit(name, size)
	return Handle{inner: &resInner}
}

/*
Creates an inclusion query. Here, `included` and `including` are viewed
as a arrays and the query asserts that `included` contains only rows
that are contained within `includings`, regardless of the multiplicity.
*/
func (b *Builder) Inclusion(name ifaces.QueryID, including, included []Handle) {

	including_ := []ifaces.Column{}

	// Guard against poisoned handles
	for _, h := range including {
		if h.IsNil() {
			// logrus.Tracef("Skipped inclusion query %v because it relates to a poisoned handle\n", name)
			return
		}
		including_ = append(including_, h.Unwrap())

	}

	included_ := []ifaces.Column{}

	// Guard against the included
	for _, h := range included {
		if h.IsNil() {
			// logrus.Tracef("Skipped inclusion query %v because it relates to a poisoned handle\n", name)
			return
		}
		included_ = append(included_, h.Unwrap())

	}
	// Guards against the
	b.internal.Inclusion(name, including_, included_)
}

/*
Creates an permutation query. The query views `a` and `b_` to be lists of
columns and asserts that `a` and `b_` have the same rows (possibly in
a different order) but with the same multiplicity.
*/
func (b *Builder) Permutation(name ifaces.QueryID, a, b_ []Handle) {

	a1, b1 := [][]ifaces.Column{{}}, [][]ifaces.Column{{}}

	if len(a[0].interleavedComponents) > 0 {
		var (
			numFragments = len(a[0].interleavedComponents)
		)
		a1 = make([][]ifaces.Column, numFragments)
		for frag := 0; frag < numFragments; frag++ {
			for col := range a {
				if a[col].IsNil() {
					return
				}
				a1[frag] = append(a1[frag], a[col].interleavedComponents[frag])
			}
		}
	}

	if len(b_[0].interleavedComponents) > 0 {
		var (
			numFragments = len(b_[0].interleavedComponents)
		)
		b1 = make([][]ifaces.Column, numFragments)
		for frag := 0; frag < numFragments; frag++ {
			for col := range b_ {
				if b_[col].IsNil() {
					return
				}
				b1[frag] = append(b1[frag], b_[col].interleavedComponents[frag])
			}
		}
	}

	if len(a[0].interleavedComponents) == 0 {
		// Guard against poisoned handles
		for _, h := range a {
			if h.IsNil() {
				// logrus.Tracef("Skipped permutation query %v because it relates to a poisoned handle\n", name)
				return
			}
			a1[0] = append(a1[0], h.Unwrap())
		}
	}

	if len(b_[0].interleavedComponents) == 0 {
		for _, h := range b_ {
			if h.IsNil() {
				// logrus.Tracef("Skipped permutation query %v because it relates to a poisoned handle\n", name)
				return
			}
			b1[0] = append(b1[0], h.Unwrap())
		}
	}

	b.internal.CompiledIOP.InsertFragmentedPermutation(0, name, a1, b1)
}

/*
Create an GlobalConstraint query, returns the global constraint

We prefer not returning anything from this query to avoid creating a shallow
GlobalConstraint{}
*/
func (b *Builder) GlobalConstraint(name ifaces.QueryID, cs_ *symbolic.Expression) {

	if cs_ == nil {
		// logrus.Tracef("Skipped global query %v because it relates to a poisoned handle\n", name)
		return
	}

	if ok, val := cs_.IsConstant(); ok {
		if !val.IsZero() {
			// Panic on impossible constraints that can be reduced to 0 = 1
			utils.Panic("The expression of query %v is constant and is non-zero.", name)
		}
		// Skip trivial constraints
		logrus.Tracef("Skipped global query %v because it is a constant expression evaluated to zero\n", name)
		return
	}

	if err := cs_.Validate(); err != nil {
		// logrus.Tracef("Poisoned global query %v, skipping", name)
		return
	}

	// If it passes, then we cancel the expression (accounting for the offsets)
	q := query.NewGlobalConstraint(name, cs_)

	// Check if a query with the same name was already registered
	// else append a random number to it
	if b.internal.QueriesNoParams.Exists(q.ID) {
		q.ID += ifaces.QueryID(fmt.Sprint(b.deduplicator))
		b.deduplicator++
		logrus.Tracef("Renamed %v to %v to avoid duplication of query names\n", name, q.ID)
	}

	// Finally registers the query. This will perform all the checks
	b.internal.GlobalConstraint(q.ID, q.Expression)
}

/*
Create an GlobalConstraint query

Contrary to the wizard.Builder interface it does not return the local
constraint object. This is helpful to avoid
*/
func (b *Builder) LocalConstraint(name ifaces.QueryID, cs_ *symbolic.Expression) {

	// The expressions are not using the wrapper, thus we need to check for
	// nil directly.
	if cs_ == nil {
		// logrus.Tracef("Skipped local query %v because it relates to a poisoned handle\n", name)
		return
	}

	if err := cs_.Validate(); err != nil {
		// logrus.Tracef("Poisoned global query %v, skipping", name)
		return
	}

	// Check if a query with the same name was already registered
	// else append a random number to it
	if b.internal.QueriesNoParams.Exists(name) {
		oldName := name
		name += ifaces.QueryID(fmt.Sprint(b.deduplicator))
		b.deduplicator++
		logrus.Tracef("Renamed %v to %v to avoid duplication of query names\n", oldName, name)
	}

	// Finally registers the query. This will perform all the checks
	b.internal.LocalConstraint(name, cs_)
}

/*
Registers a Range query

Contrary to the wizard.Builder interface it does not return the local
constraint object. This is helpful to avoid
*/
func (b *Builder) Range(name ifaces.QueryID, h Handle, max int) {
	if h.IsNil() {
		// logrus.Tracef("Skipping range query %v because it relates to a poisoned handle\n", name)
		return
	}
	// Finally registers the query. This will perform all the checks
	b.internal.Range(name, h.Unwrap(), max)
}
