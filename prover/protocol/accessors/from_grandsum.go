package accessors

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

const (
	GRANDSUM_ACCESSOR = "GRANDSUM_ACCESSOR"
)

// FromGrandSumAccessor implements [ifaces.Accessor] and accesses the result of
// a [query.GrandSum].
type FromGrandSumAccessor struct {
	// Q is the underlying query whose parameters are accessed by the current
	// [ifaces.Accessor].
	Q query.GrandSum
}

// NewGrandSumAccessor creates an [ifaces.Accessor] returning the opening
// point of a [query.GrandSum].
func NewGrandSumAccessor(q query.GrandSum) ifaces.Accessor {
	return &FromGrandSumAccessor{Q: q}
}

// Name implements [ifaces.Accessor]
func (l *FromGrandSumAccessor) Name() string {
	return fmt.Sprintf("%v_%v", GRANDSUM_ACCESSOR, l.Q.ID)
}

// String implements [github.com/consensys/linea-monorepo/prover/symbolic.Metadata]
func (l *FromGrandSumAccessor) String() string {
	return l.Name()
}

// GetVal implements [ifaces.Accessor]
func (l *FromGrandSumAccessor) GetVal(run ifaces.Runtime) field.Element {
	params := run.GetParams(l.Q.ID).(query.GrandSumParams)
	return params.Y
}

// GetFrontendVariable implements [ifaces.Accessor]
func (l *FromGrandSumAccessor) GetFrontendVariable(_ frontend.API, circ ifaces.GnarkRuntime) frontend.Variable {
	panic("unimplemented")
}

// AsVariable implements the [ifaces.Accessor] interface
func (l *FromGrandSumAccessor) AsVariable() *symbolic.Expression {
	return symbolic.NewVariable(l)
}

// Round implements the [ifaces.Accessor] interface
func (l *FromGrandSumAccessor) Round() int {
	return l.Q.Round
}
