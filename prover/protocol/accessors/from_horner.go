package accessors

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

// var _ ifaces.Accessor = &FromHornerAccessorFinalValue[T]{}

// FromHornerAccessorFinalValue[T] implements [ifaces.Accessor] and accesses the
// final value of a [Horner] computation.
type FromHornerAccessorFinalValue[T zk.Element] struct {
	Q *query.Horner
}

// NewFromHornerAccessorFinalValue[T] returns a new [FromHornerAccessorFinalValue[T]].
func NewFromHornerAccessorFinalValue[T zk.Element](q *query.Horner) *FromHornerAccessorFinalValue[T] {
	return &FromHornerAccessorFinalValue[T]{Q: q}
}

// Name implements [ifaces.Accessor]
func (l *FromHornerAccessorFinalValue[T]) Name() string {
	return "HORNER_ACCESSOR_FINAL_VALUE_" + string(l.Q.ID)
}

// String implements [github.com/consensys/linea-monorepo/prover/symbolic.Metadata]
func (l *FromHornerAccessorFinalValue[T]) String() string {
	return l.Name()
}

// GetVal implements [ifaces.Accessor]. It is not implemented for this accessor
// as it should always return an extension field due to its dependency on
// randomness.
func (l *FromHornerAccessorFinalValue[T]) GetVal(run ifaces.Runtime) field.Element {
	panic("should not be called as the result is an extension field")
}

// GetVal implements [ifaces.Accessor]
func (l *FromHornerAccessorFinalValue[T]) GetValExt(run ifaces.Runtime) fext.Element {
	params := run.GetParams(l.Q.ID).(query.HornerParams)
	return params.FinalResult
}

// GetFrontendVariable implements [ifaces.Accessor]
func (l *FromHornerAccessorFinalValue[T]) GetFrontendVariable(_ zk.APIGen[T], circ ifaces.GnarkRuntime[T]) T {
	params := circ.GetParams(l.Q.ID).(query.GnarkHornerParams)
	return params.FinalResult
}

// AsVariable implements the [ifaces.Accessor] interface
func (l *FromHornerAccessorFinalValue[T]) AsVariable() *symbolic.Expression[T] {
	return symbolic.NewVariable[T](l)
}

// Round implements the [ifaces.Accessor] interface
func (l *FromHornerAccessorFinalValue[T]) Round() int {
	return l.Q.Round
}

// GetValBase implements [ifaces.Accessor]. It panics as it should never be called
// since the result is always an extension field.
func (l *FromHornerAccessorFinalValue[T]) GetValBase(run ifaces.Runtime) (field.Element, error) {
	//TODO implement me
	panic("should not be called as the result is an extension field")
}

func (l *FromHornerAccessorFinalValue[T]) GetFrontendVariableBase(api zk.APIGen[T], c ifaces.GnarkRuntime[T]) (T, error) {
	//TODO implement me
	panic("implement me")
}

func (l *FromHornerAccessorFinalValue[T]) GetFrontendVariableExt(api zk.APIGen[T], c ifaces.GnarkRuntime[T]) gnarkfext.E4Gen[T] {
	//TODO implement me
	panic("implement me")
}

func (l *FromHornerAccessorFinalValue[T]) IsBase() bool {
	return false
}
