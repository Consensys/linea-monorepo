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
	DISTRIBUTED_PROJECTION_ACCESSOR = "DISTRIBUTED_PROJECTION_ACCESSOR"
)

// FromDistributedProjectionAccessor implements [ifaces.Accessor] and accesses the result of
// a [query.DISTRIBUTED_PROJECTION].
type FromDistributedProjectionAccessor struct {
	// Q is the underlying query whose parameters are accessed by the current
	// [ifaces.Accessor].
	Q query.DistributedProjection
}

// NewDistributedProjectionAccessor creates an [ifaces.Accessor] returning the opening
// point of a [query.DISTRIBUTED_PROJECTION].
func NewDistributedProjectionAccessor(q query.DistributedProjection) ifaces.Accessor {
	return &FromDistributedProjectionAccessor{Q: q}
}

// Name implements [ifaces.Accessor]
func (l *FromDistributedProjectionAccessor) Name() string {
	return fmt.Sprintf("%v_%v", DISTRIBUTED_PROJECTION_ACCESSOR, l.Q.ID)
}

// String implements [github.com/consensys/linea-monorepo/prover/symbolic.Metadata]
func (l *FromDistributedProjectionAccessor) String() string {
	return l.Name()
}

// GetVal implements [ifaces.Accessor]
func (l *FromDistributedProjectionAccessor) GetVal(run ifaces.Runtime) field.Element {
	params := run.GetParams(l.Q.ID).(query.DistributedProjectionParams)
	return params.HornerVal
}

// GetFrontendVariable implements [ifaces.Accessor]
func (l *FromDistributedProjectionAccessor) GetFrontendVariable(_ frontend.API, circ ifaces.GnarkRuntime) frontend.Variable {
	params := circ.GetParams(l.Q.ID).(query.GnarkDistributedProjectionParams)
	return params.Sum
}

// AsVariable implements the [ifaces.Accessor] interface
func (l *FromDistributedProjectionAccessor) AsVariable() *symbolic.Expression {
	return symbolic.NewVariable(l)
}

// Round implements the [ifaces.Accessor] interface
func (l *FromDistributedProjectionAccessor) Round() int {
	return l.Q.Round
}
