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
	commonconstraints "github.com/consensys/linea-monorepo/prover/zkevm/prover/common/common_constraints"
)

// HeartBeatColumn is an implementation of the [ifaces.Column] interface that
// represents a structured column. The structured column pulsates a "1" every
// "period" rows with the provided offset.
type HeartBeatColumn struct {
	column.Natural
	Counter CyclicCounter
	Offset  int
	PAs     []wizard.ProverAction
}

// CyclicCounter is a column counting and periodically resetting. The
// column type is implemented as a composite of [column.Natural]. The
// column can auto-assign if [GetColAssignment] is called before the
// column is explictly assigned.
//
// The counter is furthermore controlled by an 'isActive' column: the values
// of the column are forced to zero if isActive=0.
type CyclicCounter struct {
	column.Natural
	Reset    ifaces.Column
	Period   int
	IsActive ifaces.Column
	PAs      []wizard.ProverAction
}

// RepeatedPattern is a column populated with an ever-repeated pattern.
// The pattern may have a non-zero power of two size. The column is
// subjected to an "is-active" column.
type RepeatedPattern struct {
	column.Natural
	Pattern           []field.Element
	Counter           CyclicCounter
	PatternPrecomp    ifaces.Column
	PatternPosPrecomp ifaces.Column
}

// CreateHeartBeat creates a self-constrained column that repeats of "1",
// followed by "period" zero. The period does not have to be a power of
// two. CreateHeartBeat expands over [column.Natural] and lazily
// self-assign itself when its assignment is required. The function is
// masked by an [IsActive] column which control it is zero-padded.
//
// The function also defines and assign underlying columns
func CreateHeartBeat(comp *wizard.CompiledIOP, period, offset int, isActive ifaces.Column) *HeartBeatColumn {

	hb := &HeartBeatColumn{
		Offset:  offset,
		Counter: *NewRepetionCounter(comp, period, isActive),
	}

	if offset == -1 || offset == period-1 {
		hb.Natural = hb.Counter.Reset.(column.Natural)
		return hb
	}

	res, isNat := IsZeroMask(comp, sym.Sub(hb.Counter, offset), isActive)

	hb.PAs = append(hb.PAs, isNat)
	hb.Natural = res.(column.Natural)

	return hb
}

func (hb HeartBeatColumn) Assign(run *wizard.ProverRuntime) {
	hb.Counter.Assign(run)
	for _, pa := range hb.PAs {
		pa.Run(run)
	}
}

// NewRepetionCounter creates a structured [CyclicCounter]
func NewRepetionCounter(comp *wizard.CompiledIOP, period int, isActive ifaces.Column) *CyclicCounter {

	size := isActive.Size()
	name := fmt.Sprintf("REPETITION_COUNTER_%v_%v", len(comp.Columns.AllKeys()), period)

	rc := &CyclicCounter{
		IsActive: isActive,
		Period:   period,
		Natural:  comp.InsertCommit(0, ifaces.ColID(name+"_COUNTER"), size).(column.Natural),
	}

	commonconstraints.MustZeroWhenInactive(comp, isActive, rc.Natural)

	comp.InsertLocal(
		0,
		ifaces.QueryID(name+"_COUNTER_STARTS_AT_ZERO"),
		sym.NewVariable(rc.Natural),
	)

	comp.InsertGlobal(
		0,
		ifaces.QueryID(name+"COUNTER_INCREASES"),
		sym.Mul(
			column.Shift(isActive, 1),
			sym.Sub(rc.Natural, period-1),
			sym.Sub(
				column.Shift(rc.Natural, 1),
				rc.Natural,
				1,
			),
		),
	)

	isEndOfPeriod, cptIsEndOfPeriod := IsZeroMask(comp, sym.Sub(rc.Natural, period-1), isActive)

	comp.InsertGlobal(
		0,
		ifaces.QueryID(name+"_COUNTER_RESET"),
		sym.Mul(
			column.Shift(rc.Natural, 1),
			isEndOfPeriod,
		),
	)

	rc.PAs = append(rc.PAs, cptIsEndOfPeriod)
	rc.Reset = isEndOfPeriod

	return rc
}

// Assign runs the prover steps and assign the CounterColumn
func (rc CyclicCounter) Assign(run *wizard.ProverRuntime) {

	var (
		isActive = rc.IsActive.GetColAssignment(run).IntoRegVecSaveAlloc()
		size     = len(isActive)
		res      = make([]field.Element, size)
	)

	for i := range isActive {

		if isActive[i].IsZero() {
			res = res[:i:i]
			break
		}

		n := utils.PositiveMod(i, rc.Period)
		res[i].SetUint64(uint64(n))
	}

	run.AssignColumn(rc.ID, smartvectors.RightZeroPadded(res, size))

	for i := range rc.PAs {
		rc.PAs[i].Run(run)
	}
}

// NewRepeatedPattern creates a new [RepeatedPattern] column
func NewRepeatedPattern(comp *wizard.CompiledIOP, pattern []field.Element, isActive ifaces.Column) *RepeatedPattern {

	var (
		size              = isActive.Size()
		period            = len(pattern)
		name              = fmt.Sprintf("REPEATED_PATTERN_%v_%v", len(comp.Columns.AllKeys()), period)
		patternSizePadded = utils.NextPowerOfTwo(period)
		patternPos        = make([]field.Element, period)
	)

	for i := range period {
		patternPos[i] = field.NewElement(uint64(i))
	}

	res := &RepeatedPattern{
		Natural: comp.InsertCommit(0, ifaces.ColID(name)+"_NATURAL", size).(column.Natural),
		Pattern: pattern,
		PatternPrecomp: comp.InsertPrecomputed(
			ifaces.ColID(name)+"_PATTERN",
			smartvectors.RightZeroPadded(pattern, patternSizePadded),
		),
		PatternPosPrecomp: comp.InsertPrecomputed(
			ifaces.ColID(name)+"_PATTERNPOS",
			smartvectors.RightPadded(patternPos, field.NewFromString("-1"), patternSizePadded),
		),
		Counter: *NewRepetionCounter(comp, period, isActive),
	}

	commonconstraints.MustZeroWhenInactive(comp, isActive, res.Natural)

	comp.InsertInclusionConditionalOnIncluded(
		0,
		ifaces.QueryID(name)+"_LOOKUP",
		[]ifaces.Column{
			res.PatternPosPrecomp,
			res.PatternPrecomp,
		},
		[]ifaces.Column{
			res.Counter.Natural,
			res.Natural,
		},
		isActive,
	)

	return res
}

func (rp RepeatedPattern) Assign(run *wizard.ProverRuntime) {

	var (
		isActive = rp.Counter.IsActive.GetColAssignment(run).IntoRegVecSaveAlloc()
		size     = len(isActive)
		res      = make([]field.Element, 0, size)
		period   = len(rp.Pattern)
	)

	for i := 0; i < size && isActive[i].IsOne(); i += period {
		for j := range rp.Pattern {
			if i+j >= size || isActive[i+j].IsZero() {
				break
			}

			res = append(res, rp.Pattern[j])
		}
	}

	run.AssignColumn(rp.ID, smartvectors.RightZeroPadded(res, size))
	rp.Counter.Assign(run)
}
