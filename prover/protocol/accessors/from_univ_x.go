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

// var _ ifaces.Accessor = &FromUnivXAccessor{}

// FromUnivXAccessor implements [ifaces.Accessor]. It represents the "X" of a
// univariate evaluation query (see [query.UnivariateEval]).
type FromUnivXAccessor[T zk.Element] struct {
	// Q is the underlying univariate evaluation query
	Q query.UnivariateEval[T]
	// Round is the declaration round of Q
	QRound int
}

func (u *FromUnivXAccessor[T]) IsBase() bool {
	//TODO implement me
	return u.Q.Pols[0].IsBase()

}

func (u *FromUnivXAccessor[T]) GetValBase(run ifaces.Runtime) (field.Element, error) {
	//TODO implement me
	panic("implement me")
}

func (u *FromUnivXAccessor[T]) GetValExt(run ifaces.Runtime) fext.Element {
	params := run.GetParams(u.Q.QueryID).(query.UnivariateEvalParams[T])
	return params.ExtX
}

func (u *FromUnivXAccessor[T]) GetFrontendVariableBase(api zk.APIGen[T], c ifaces.GnarkRuntime[T]) (T, error) {
	//TODO implement me
	panic("implement me")
}

func (u *FromUnivXAccessor[T]) GetFrontendVariableExt(api zk.APIGen[T], c ifaces.GnarkRuntime[T]) gnarkfext.E4Gen[T] {
	//TODO implement me
	panic("implement me")
}

// NewUnivariateX returns an [ifaces.Accessor] object symbolizing the evaluation
// point (the "X" value) of a [query.UnivariateEval]. `qRound` is must be the
// underlying declaration round of the query object.
func NewUnivariateX[T zk.Element](q query.UnivariateEval[T], qround int) ifaces.Accessor[T] {
	return &FromUnivXAccessor[T]{
		Q:      q,
		QRound: qround,
	}
}

// Name implements [ifaces.Accessor]
func (u *FromUnivXAccessor[T]) Name() string {
	return fmt.Sprintf("UNIV_X_ACCESSOR_%v", u.Q.QueryID)
}

// String implements [github.com/consensys/linea-monorepo/prover/symbolic.Metadata]
func (u *FromUnivXAccessor[T]) String() string {
	return u.Name()
}

// GetVal implements [ifaces.Accessor]
func (u *FromUnivXAccessor[T]) GetVal(run ifaces.Runtime) field.Element {
	//TODO implement me
	panic("implement me")
}

// GetFrontendVariable implements [ifaces.Accessor]
func (u *FromUnivXAccessor[T]) GetFrontendVariable(_ zk.APIGen[T], circ ifaces.GnarkRuntime[T]) T {
	params := circ.GetParams(u.Q.QueryID).(query.GnarkUnivariateEvalParams[T])
	return params.X
}

// AsVariable implements the [ifaces.Accessor] interface
func (u *FromUnivXAccessor[T]) AsVariable() *symbolic.Expression[T] {
	return symbolic.NewVariable[T](u)
}

// Round implements the [ifaces.Accessor] interface
func (u *FromUnivXAccessor[T]) Round() int {
	return u.QRound
}
