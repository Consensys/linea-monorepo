package accessors

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

const (
	GRANDPRODUCT_ACCESSOR = "GRANDPRODUCT_ACCESSOR"
)

// FromGrandProductAccessor[T] implements [ifaces.Accessor] and accesses the result of
// a [query.GrandProduct].
type FromGrandProductAccessor[T zk.Element] struct {
	// Q is the underlying query whose parameters are accessed by the current
	// [ifaces.Accessor].
	Q query.GrandProduct[T]
}

// IsBase implements [ifaces.Accessor] and always returns false as grand-product
// always returns an extension field as they depends on randomness in a sound
// setting. For testing it is always possible to mock the FS to return base
// fields to make debugging easier but by default this is not possible.
func (l *FromGrandProductAccessor[T]) IsBase() bool {
	return false
}

func (l *FromGrandProductAccessor[T]) GetValBase(run ifaces.Runtime) (field.Element, error) {
	panic(fmt.Sprintf("Called GetValBase on a FromGrandProductAccessor[T], %v", l.Name()))

}

func (l *FromGrandProductAccessor[T]) GetValExt(run ifaces.Runtime) fext.Element {
	return run.GetParams(l.Q.ID).(query.GrandProductParams).ExtY
}

func (l *FromGrandProductAccessor[T]) GetFrontendVariableBase(api zk.APIGen[T], c ifaces.GnarkRuntime[T]) (T, error) {
	//TODO implement me
	panic("implement me")
}

// GetFrontendVariable implements [ifaces.Accessor]
func (l *FromGrandProductAccessor[T]) GetFrontendVariable(_ zk.APIGen[T], circ ifaces.GnarkRuntime[T]) T {
	params := circ.GetParams(l.Q.ID).(query.GnarkGrandProductParams[T])
	return params.Prod
}

func (l *FromGrandProductAccessor[T]) GetFrontendVariableExt(api zk.APIGen[T], c ifaces.GnarkRuntime[T]) gnarkfext.E4Gen[T] {
	//TODO implement me
	panic("implement me")
}

// NewGrandProductAccessor creates an [ifaces.Accessor] returning the opening
// point of a [query.GrandProduct].
func NewGrandProductAccessor[T zk.Element](q query.GrandProduct[T]) ifaces.Accessor[T] {
	return &FromGrandProductAccessor[T]{Q: q}
}

// Name implements [ifaces.Accessor]
func (l *FromGrandProductAccessor[T]) Name() string {
	return fmt.Sprintf("%v_%v", GRANDPRODUCT_ACCESSOR, l.Q.ID)
}

// String implements [github.com/consensys/linea-monorepo/prover/symbolic.Metadata]
func (l *FromGrandProductAccessor[T]) String() string {
	return l.Name()
}

// GetVal implements [ifaces.Accessor] and returns the opening point of a
// grand product opening.
func (l *FromGrandProductAccessor[T]) GetVal(run ifaces.Runtime) field.Element {
	panic(fmt.Sprintf("Called GetVal on a FromGrandProductAccessor[T], %v", l.Name()))
}

// AsVariable implements the [ifaces.Accessor] interface
func (l *FromGrandProductAccessor[T]) AsVariable() *symbolic.Expression[T] {
	return symbolic.NewVariable[T](l)
}

// Round implements the [ifaces.Accessor] interface
func (l *FromGrandProductAccessor[T]) Round() int {
	return l.Q.Round
}
