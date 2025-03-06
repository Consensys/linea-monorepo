package dedicated

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// HeartBeatColumn is an implementation of the [ifaces.Column] interface that
// represents a structured column. The structured column pulsates a "1" every
// "period" rows with the provided offset.
type HeartBeatColumn struct {
	column.Natural
	Period   int
	Offset   int
	Counter  column.Natural
	IsActive ifaces.Column
	PAs      []wizard.ProverAction
}

// CreateHeartBeat creates a self-constrained column that repeats of "1",
// followed by "period" zero. The period does not have to be a power of
// two. CreateHeartBeat expands over [column.Natural] and lazily
// self-assign itself when its assignment is required. The function is
// masked by an [IsActive] column which control it is zero-padded.
//
// The function also defines and assign underlying columns
func CreateHeartBeat(comp *wizard.CompiledIOP, period, offset int, isActive ifaces.Column) *HeartBeatColumn {

	size := isActive.Size()
	name := fmt.Sprintf("HEART_BEAT_%v_PERIOD=%v_OFFSET=%v", len(comp.Columns.AllKeys()), period, offset)

	hb := &HeartBeatColumn{
		IsActive: isActive,
		Period:   period,
		Offset:   offset,
		Counter:  comp.InsertCommit(0, ifaces.ColID(name+"_COUNTER"), size).(column.Natural),
	}

	comp.InsertLocal(
		0,
		ifaces.QueryID(name+"_COUNTER_STARTS_AT_ZERO"),
		sym.NewVariable(hb.Counter),
	)

	comp.InsertGlobal(
		0,
		ifaces.QueryID(name+"COUNTER_INCREASES"),
		sym.Mul(
			column.Shift(isActive, 1),
			sym.Sub(hb.Counter, period-1),
			sym.Sub(
				column.Shift(hb.Counter, 1),
				hb.Counter,
				1,
			),
		),
	)

	isEndOfPeriod, cptIsEndOfPeriod := IsZeroMask(comp, sym.Sub(hb.Counter, period-1), isActive)

	comp.InsertGlobal(
		0,
		ifaces.QueryID(name+"_COUNTER_RESET"),
		sym.Mul(
			column.Shift(hb.Counter, 1),
			isEndOfPeriod,
		),
	)

	hb.PAs = append(hb.PAs, cptIsEndOfPeriod)

	if offset == -1 || offset == period-1 {
		hb.Natural = isEndOfPeriod.(column.Natural)
	}

	res, isNat := IsZeroMask(comp, sym.Sub(hb.Counter, offset), isActive)

	hb.PAs = append(hb.PAs, isNat)
	hb.Natural = res.(column.Natural)

	return hb
}

// Assign runs the prover steps and assign the CounterColumn
func (hb HeartBeatColumn) Assign(run *wizard.ProverRuntime) {

	var (
		isActive = hb.IsActive.GetColAssignment(run).IntoRegVecSaveAlloc()
		size     = len(isActive)
		res      = make([]field.Element, size)
	)

	for i := range isActive {

		if isActive[i].IsZero() {
			res = res[:i:i]
			break
		}

		n := utils.PositiveMod(i, hb.Period)
		res[i].SetUint64(uint64(n))
	}

	run.AssignColumn(hb.Counter.ID, smartvectors.RightZeroPadded(res, size))

	for i := range hb.PAs {
		hb.PAs[i].Run(run)
	}
}

// GetColAssignment overrides the main [Natural.GetColAssignment] by
// running [Assign] if needed.
func (hb HeartBeatColumn) GetColAssignment(run ifaces.Runtime) ifaces.ColAssignment {
	if pr, isPR := run.(*wizard.ProverRuntime); isPR && !pr.Columns.Exists(hb.ID) {
		hb.Assign(pr)
	}
	return hb.Natural.GetColAssignment(run)
}

// GetColAssignmentAt overrides the main [Natural.GetColAssignmentAt] by
// running [Assign] if needed.
func (hb HeartBeatColumn) GetColAssignmentAt(run ifaces.Runtime, pos int) field.Element {
	if pr, isPR := run.(*wizard.ProverRuntime); isPR && !pr.Columns.Exists(hb.ID) {
		hb.Assign(pr)
	}
	return hb.Natural.GetColAssignmentAt(run, pos)
}
