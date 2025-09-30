package accessors

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/utils"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

const (
	LOGDERIVSUM_ACCESSOR = "LOGDERIVSUM_ACCESSOR"
)

// FromLogDerivSumAccessor[T] implements [ifaces.Accessor] and accesses the result of
// a [query.LogDerivativeSum].
type FromLogDerivSumAccessor[T zk.Element] struct {
	// Q is the underlying query whose parameters are accessed by the current
	// [ifaces.Accessor].
	Q query.LogDerivativeSum[T]
}

func (l *FromLogDerivSumAccessor[T]) IsBase() bool {
	return false
}

func (l *FromLogDerivSumAccessor[T]) GetValBase(run ifaces.Runtime) (field.Element, error) {
	//TODO implement me
	panic("implement me")
}

func (l *FromLogDerivSumAccessor[T]) GetValExt(run ifaces.Runtime) fext.Element {
	params := run.GetParams(l.Q.ID).(query.LogDerivSumParams)
	return params.Sum.GetExt()
}

func (l *FromLogDerivSumAccessor[T]) GetFrontendVariableBase(api zk.APIGen[T], c ifaces.GnarkRuntime[T]) (T, error) {
	//TODO implement me
	panic("implement me")
}

func (l *FromLogDerivSumAccessor[T]) GetFrontendVariableExt(api zk.APIGen[T], c ifaces.GnarkRuntime[T]) gnarkfext.E4Gen[T] {
	//TODO implement me
	panic("implement me")
}

// NewLogDerivSumAccessor creates an [ifaces.Accessor] returning the opening
// point of a [query.LogDerivativeSum].
func NewLogDerivSumAccessor[T zk.Element](q query.LogDerivativeSum[T]) ifaces.Accessor[T] {
	return &FromLogDerivSumAccessor[T]{Q: q}
}

// Name implements [ifaces.Accessor]
func (l *FromLogDerivSumAccessor[T]) Name() string {
	return fmt.Sprintf("%v_%v", LOGDERIVSUM_ACCESSOR, l.Q.ID)
}

// String implements [github.com/consensys/linea-monorepo/prover/symbolic.Metadata]
func (l *FromLogDerivSumAccessor[T]) String() string {
	return l.Name()
}

// GetVal implements [ifaces.Accessor]
func (l *FromLogDerivSumAccessor[T]) GetVal(run ifaces.Runtime) field.Element {
	utils.Panic("Called GetVal on a FromLogDerivSumAccessor[T], %v, but it should call GetValExt", l.Name())
	return field.Element{} // not reachable
}

// GetFrontendVariable implements [ifaces.Accessor]
func (l *FromLogDerivSumAccessor[T]) GetFrontendVariable(_ zk.APIGen[T], circ ifaces.GnarkRuntime[T]) T {
	params := circ.GetParams(l.Q.ID).(query.GnarkLogDerivSumParams[T])
	return params.Sum
}

// AsVariable implements the [ifaces.Accessor] interface
func (l *FromLogDerivSumAccessor[T]) AsVariable() *symbolic.Expression[T] {
	return symbolic.NewVariable[T](l)
}

// Round implements the [ifaces.Accessor] interface
func (l *FromLogDerivSumAccessor[T]) Round() int {
	return l.Q.Round
}
