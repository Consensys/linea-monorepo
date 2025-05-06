package accessors

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

var _ ifaces.Accessor = &FromHornerAccessorFinalValue{}

// FromHornerAccessorFinalValue implements [ifaces.Accessor] and accesses the
// final value of a [Horner] computation.
type FromHornerAccessorFinalValue struct {
	Q *query.Horner
}

// NewFromHornerAccessorFinalValue returns a new [FromHornerAccessorFinalValue].
func NewFromHornerAccessorFinalValue(q *query.Horner) *FromHornerAccessorFinalValue {
	return &FromHornerAccessorFinalValue{Q: q}
}

// Name implements [ifaces.Accessor]
func (l *FromHornerAccessorFinalValue) Name() string {
	return "HORNER_ACCESSOR_FINAL_VALUE_" + string(l.Q.ID)
}

// String implements [github.com/consensys/linea-monorepo/prover/symbolic.Metadata]
func (l *FromHornerAccessorFinalValue) String() string {
	return l.Name()
}

// GetVal implements [ifaces.Accessor]
func (l *FromHornerAccessorFinalValue) GetVal(run ifaces.Runtime) field.Element {
	params := run.GetParams(l.Q.ID).(query.HornerParams)
	return params.FinalResult
}

// GetFrontendVariable implements [ifaces.Accessor]
func (l *FromHornerAccessorFinalValue) GetFrontendVariable(_ frontend.API, circ ifaces.GnarkRuntime) frontend.Variable {
	params := circ.GetParams(l.Q.ID).(query.GnarkHornerParams)
	return params.FinalResult
}

// AsVariable implements the [ifaces.Accessor] interface
func (l *FromHornerAccessorFinalValue) AsVariable() *symbolic.Expression {
	return symbolic.NewVariable(l)
}

// Round implements the [ifaces.Accessor] interface
func (l *FromHornerAccessorFinalValue) Round() int {
	return l.Q.Round
}

func (l *FromHornerAccessorFinalValue) GetValBase(run ifaces.Runtime) (field.Element, error) {
	//TODO implement me
	panic("implement me")
}

func (l *FromHornerAccessorFinalValue) GetValExt(run ifaces.Runtime) fext.Element {
	//TODO implement me
	panic("implement me")
}

func (l *FromHornerAccessorFinalValue) GetFrontendVariableBase(api frontend.API, c ifaces.GnarkRuntime) (frontend.Variable, error) {
	//TODO implement me
	panic("implement me")
}

func (l *FromHornerAccessorFinalValue) GetFrontendVariableExt(api frontend.API, c ifaces.GnarkRuntime) gnarkfext.Variable {
	//TODO implement me
	panic("implement me")
}

func (l *FromHornerAccessorFinalValue) IsBase() bool {
	//TODO implement me
	panic("implement me")
}
