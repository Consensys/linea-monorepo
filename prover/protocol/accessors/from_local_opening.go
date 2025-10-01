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

// FromLocalOpeningYAccessor[T] implements [ifaces.Accessor] and accesses the result of
// a [query.LocalOpening].
type FromLocalOpeningYAccessor[T zk.Element] struct {
	// Q is the underlying query whose parameters are accessed by the current
	// [ifaces.Accessor].
	Q query.LocalOpening[T]
	// QRound is the declaration round of the query
	QRound int
}

func (l *FromLocalOpeningYAccessor[T]) IsBase() bool {
	return l.Q.Pol.IsBase()
}

func (l *FromLocalOpeningYAccessor[T]) GetValBase(run ifaces.Runtime) (field.Element, error) {
	//TODO implement me
	panic("implement me")
}

func (l *FromLocalOpeningYAccessor[T]) GetValExt(run ifaces.Runtime) fext.Element {
	params := run.GetParams(l.Q.ID).(query.LocalOpeningParams[T])
	return params.ExtY
}

func (l *FromLocalOpeningYAccessor[T]) GetFrontendVariableBase(api zk.APIGen[T], c ifaces.GnarkRuntime[T]) (T, error) {
	//TODO implement me
	panic("implement me")
}

func (l *FromLocalOpeningYAccessor[T]) GetFrontendVariableExt(api zk.APIGen[T], c ifaces.GnarkRuntime[T]) gnarkfext.E4Gen[T] {
	//TODO implement me
	panic("implement me")
}

// NewLocalOpeningAccessor creates an [ifaces.Accessor] returning the opening
// point of a [query.LocalOpening].
func NewLocalOpeningAccessor[T zk.Element](q query.LocalOpening[T], qRound int) ifaces.Accessor[T] {
	return &FromLocalOpeningYAccessor[T]{Q: q, QRound: qRound}
}

// Name implements [ifaces.Accessor]
func (l *FromLocalOpeningYAccessor[T]) Name() string {
	return fmt.Sprintf("LOCAL_OPENING_ACCESSOR_%v", l.Q.ID)
}

// String implements [github.com/consensys/linea-monorepo/prover/symbolic.Metadata]
func (l *FromLocalOpeningYAccessor[T]) String() string {
	return l.Name()
}

// GetVal implements [ifaces.Accessor]
func (l *FromLocalOpeningYAccessor[T]) GetVal(run ifaces.Runtime) field.Element {
	params := run.GetParams(l.Q.ID).(query.LocalOpeningParams[T])
	return params.BaseY
}

// GetFrontendVariable implements [ifaces.Accessor]
func (l *FromLocalOpeningYAccessor[T]) GetFrontendVariable(_ zk.APIGen[T], circ ifaces.GnarkRuntime[T]) T {
	params := circ.GetParams(l.Q.ID).(query.GnarkLocalOpeningParams[T])
	return params.BaseY
}

// AsVariable implements the [ifaces.Accessor] interface
func (l *FromLocalOpeningYAccessor[T]) AsVariable() *symbolic.Expression[T] {
	return symbolic.NewVariable[T](l)
}

// Round implements the [ifaces.Accessor] interface
func (l *FromLocalOpeningYAccessor[T]) Round() int {
	return l.QRound
}
