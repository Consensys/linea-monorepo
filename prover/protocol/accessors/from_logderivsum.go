package accessors

import (
	"fmt"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext/gnarkfext"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

const (
	LOGDERIVSUM_ACCESSOR = "LOGDERIVSUM_ACCESSOR"
)

// FromLogDerivSumAccessor implements [ifaces.Accessor] and accesses the result of
// a [query.LogDerivativeSum].
type FromLogDerivSumAccessor struct {
	// Q is the underlying query whose parameters are accessed by the current
	// [ifaces.Accessor].
	Q query.LogDerivativeSum
}

func (l *FromLogDerivSumAccessor) IsBase() bool {
	//TODO implement me
	panic("implement me")
}

func (l *FromLogDerivSumAccessor) GetValBase(run ifaces.Runtime) (field.Element, error) {
	//TODO implement me
	panic("implement me")
}

func (l *FromLogDerivSumAccessor) GetValExt(run ifaces.Runtime) fext.Element {
	//TODO implement me
	panic("implement me")
}

func (l *FromLogDerivSumAccessor) GetFrontendVariableBase(api frontend.API, c ifaces.GnarkRuntime) (frontend.Variable, error) {
	//TODO implement me
	panic("implement me")
}

func (l *FromLogDerivSumAccessor) GetFrontendVariableExt(api frontend.API, c ifaces.GnarkRuntime) gnarkfext.Variable {
	//TODO implement me
	panic("implement me")
}

// NewLogDerivSumAccessor creates an [ifaces.Accessor] returning the opening
// point of a [query.LogDerivativeSum].
func NewLogDerivSumAccessor(q query.LogDerivativeSum) ifaces.Accessor {
	return &FromLogDerivSumAccessor{Q: q}
}

// Name implements [ifaces.Accessor]
func (l *FromLogDerivSumAccessor) Name() string {
	return fmt.Sprintf("%v_%v", LOGDERIVSUM_ACCESSOR, l.Q.ID)
}

// String implements [github.com/consensys/linea-monorepo/prover/symbolic.Metadata]
func (l *FromLogDerivSumAccessor) String() string {
	return l.Name()
}

// GetVal implements [ifaces.Accessor]
func (l *FromLogDerivSumAccessor) GetVal(run ifaces.Runtime) field.Element {
	params := run.GetParams(l.Q.ID).(query.LogDerivSumParams)
	res, err := params.Sum.GetBase()
	if err != nil {
		panic(err)
	}
	return res
}

// GetFrontendVariable implements [ifaces.Accessor]
func (l *FromLogDerivSumAccessor) GetFrontendVariable(_ frontend.API, circ ifaces.GnarkRuntime) frontend.Variable {
	params := circ.GetParams(l.Q.ID).(query.GnarkLogDerivSumParams)
	return params.Sum
}

// AsVariable implements the [ifaces.Accessor] interface
func (l *FromLogDerivSumAccessor) AsVariable() *symbolic.Expression {
	return symbolic.NewVariable(l)
}

// Round implements the [ifaces.Accessor] interface
func (l *FromLogDerivSumAccessor) Round() int {
	return l.Q.Round
}
