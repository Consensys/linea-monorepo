package zkevm

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/keccak_wizard/zkevm_keccak"
	"github.com/consensys/accelerated-crypto-monorepo/glue"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/query"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/symbolic"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/zkevm/statemanager"
	"github.com/sirupsen/logrus"
)

/*
Builder is an adapter for wizard.Builder. It is responsible for filtering
out commitment whose size is non-power of two.

It is especially helpful to be able to filter out
*/
type Builder struct {
	internal      *wizard.Builder
	deduplicator  int
	NumColLimit   int
	ColDepthLimit int
	keccak        struct {
		enabled bool // will be zero if the option is not called
	}
	statemanager struct {
		enabled      bool
		RegisterFunc func(comp *wizard.CompiledIOP, round int)
	}
	ecdsa struct {
		// numSigs is upper bound on the number of signatures which can be
		// validated. ECDSA precompiles not proved when numSigs == 0; otherwise,
		// initialize a static circuit allowing to prove correctness of numSigs
		// signatures. When the actual number of signatures is less, then padded
		// with valid dummy signatures.
		numSigs int
		// txExtractor allows to extract signatures from the transaction.
		txExtractor glue.TxSignatureExtractor
	}
}

// Wraps a zk-evm DefineFunc into a wizard one
func WrapDefine(f func(*Builder), ops ...Option) func(*wizard.Builder) {
	return func(b *wizard.Builder) {
		// Wraps the builder
		builder := &Builder{internal: b}

		// Applies the options
		for _, op := range ops {
			op(builder)
		}

		f(builder)

		// Declares the keccaks (optional)
		if builder.keccak.enabled {
			// implicitly, the keccak round is always zero
			zkevm_keccak.RegisterKeccak(b.CompiledIOP, 0, NUM_KECCAKF)
		}

		// Declares the state manager module (optional)
		if builder.statemanager.enabled {
			statemanager.RegisterStateManagerMerkleProof(b.CompiledIOP, 0)
		}

		// Prove the ECDSA signature validities if the corresponding option has
		// been set.
		if builder.ecdsa.numSigs > 0 {
			glue.RegisterECDSA(b.CompiledIOP, 0, builder.ecdsa.numSigs, builder.ecdsa.txExtractor, glue.DefaultTraceExtractor)
		}
	}
}

/*
Registers a new commitment in the protocol
*/
func (b *Builder) RegisterCommit(name ifaces.ColID, size int) Handle {

	/*
		Limit the number of commitments
	*/
	if b.NumColLimit > 0 && len(b.internal.ListCommitments()) >= b.NumColLimit {
		logrus.Tracef("ignoring %v because there is already more than %v commitments", name, b.NumColLimit)
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
			// logrus.Tracef("skipped inclusion query %v because it relates to a poisoned handle\n", name)
			return
		}
		including_ = append(including_, h.Unwrap())

	}

	included_ := []ifaces.Column{}

	// Guard against the included
	for _, h := range included {
		if h.IsNil() {
			// logrus.Tracef("skipped inclusion query %v because it relates to a poisoned handle\n", name)
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

	a1, b1 := []ifaces.Column{}, []ifaces.Column{}

	// Guard against poisoned handles
	for _, h := range a {
		if h.IsNil() {
			// logrus.Tracef("skipped permutation query %v because it relates to a poisoned handle\n", name)
			return
		}
		a1 = append(a1, h.Unwrap())
	}

	for _, h := range b_ {
		if h.IsNil() {
			// logrus.Tracef("skipped permutation query %v because it relates to a poisoned handle\n", name)
			return
		}
		b1 = append(b1, h.Unwrap())
	}

	b.internal.Permutation(name, a1, b1)
}

/*
Create an GlobalConstraint query, returns the global constraint

We prefer not returning anything from this query to avoid creating a shallow
GlobalConstraint{}
*/
func (b *Builder) GlobalConstraint(name ifaces.QueryID, cs_ *symbolic.Expression) {

	if cs_ == nil {
		// logrus.Tracef("skipped global query %v because it relates to a poisoned handle\n", name)
		return
	}

	if ok, val := cs_.IsConstant(); ok {
		if !val.IsZero() {
			// Panic on impossible constraints that can be reduced to 0 = 1
			utils.Panic("the expression of query %v is constant and is non-zero.", name)
		}
		// Skip trivial constraints
		logrus.Tracef("skipped global query %v because it is a constant expression evaluated to zero\n", name)
		return
	}

	if err := cs_.Validate(); err != nil {
		// logrus.Tracef("poisoned global query %v, skipping", name)
		return
	}

	// If it passes, then we cancel the expression (accounting for the offsets)
	q := query.NewGlobalConstraint(name, cs_)

	// Check if a query with the same name was already registered
	// else append a random number to it
	if b.internal.QueriesNoParams.Exists(q.ID) {
		q.ID += ifaces.QueryID(fmt.Sprint(b.deduplicator))
		b.deduplicator++
		logrus.Tracef("renamed %v to %v to avoid duplication of query names\n", name, q.ID)
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
		// logrus.Tracef("skipped local query %v because it relates to a poisoned handle\n", name)
		return
	}

	if err := cs_.Validate(); err != nil {
		// logrus.Tracef("poisoned global query %v, skipping", name)
		return
	}

	// Check if a query with the same name was already registered
	// else append a random number to it
	if b.internal.QueriesNoParams.Exists(name) {
		oldName := name
		name += ifaces.QueryID(fmt.Sprint(b.deduplicator))
		b.deduplicator++
		logrus.Tracef("renamed %v to %v to avoid duplication of query names\n", oldName, name)
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
		// logrus.Tracef("skipping range query %v because it relates to a poisoned handle\n", name)
		return
	}
	// Finally registers the query. This will perform all the checks
	b.internal.Range(name, h.Unwrap(), max)
}
