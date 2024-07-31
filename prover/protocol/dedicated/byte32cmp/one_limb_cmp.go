package byte32cmp

import (
	"fmt"

	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/dedicated"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

// OneLimbCmpCtx is a wizard subroutine responsible for computing the comparison
// of two columns storing each a small value.
type oneLimbCmpCtx struct {

	// A and B are the two columns being compared. The two columns should be
	// preemptively constrained to be somewhat small enough so that they are
	// lookup-able. There is a hard-limit at 64 bits.
	a, b ifaces.Column

	// NumBits is the maximum number of bits the absolute value of the
	// difference between A and B can use. There is a hard limit at 64. It
	// should be positive also.
	numBits int

	// Greater, Lower and Equal are binary and mutually exclusive columns
	// indicating respectively if:
	//
	// 		- Greater[i]	= 1 => 	A[i] > B[i]
	// 		- Equal[i] 		= 1 => 	A[i] == B[i]
	// 		- Lower[i] 		= 1 => 	A[i] < B[i]
	isGreater, isLower, isEqual ifaces.Column

	// rangeChecked is an internal column which is assigned to:
	//
	//  	- Greater[i]	= 1 => 	rangeChecked[i] = A[i] - B[i]
	// 		- Equal[i] 		= 1 => 	rangeChecked[i] = 0
	// 		- Lower[i] 		= 1 => 	rangeChecked[i] = B[i] - A[i]
	rangeChecked ifaces.Column

	// internalProverActions is a list of prover action coming from stuffs
	// defered from other dedicated wizards. This includes the computation of
	// isEqual which uses the IsZero dedicated wizard.
	internalProverAction []wizard.ProverAction
}

// CmpSmallCols computes the comparison of two columns which are allegedly short
// the function does not include range-checks over a and b and it is the
// responsibility of the caller to ensure that they are small.
//
// The function returns g, e, l. There binary and mutually-exclusive indicator
// columns indicating the rows is which (g: a is greater than b, e: a and b are
// equals, l: a is smaller than b).
func CmpSmallCols(comp *wizard.CompiledIOP, a, b ifaces.Column, numBits int) (g, e, l ifaces.Column, pa wizard.ProverAction) {

	if a.Size() != b.Size() {
		utils.Panic("a and b have inconsistent sizes: %v != %v", a.Size(), b.Size())
	}

	if numBits > 64 {
		utils.Panic("the number of bits is too large (%v > 64)", numBits)
	}

	var (
		size    = a.Size()
		round   = max(a.Round(), b.Round())
		ctxName = func(subName string) string {
			return fmt.Sprintf("CMP_ONE_LIMB_CTX_%v_%v_%v", a.GetColID(), b.GetColID(), subName)
		}

		// Note: that isEqual is already constrained to be correctly formed.
		isEqual, isEqualCtx = dedicated.IsZero(comp, sym.Sub(a, b))
		rangeChecked        = comp.InsertCommit(round, ifaces.ColID(ctxName("RANGE_CHECKED")), size)
		isGreater           = comp.InsertCommit(round, ifaces.ColID(ctxName("IS_GREATER")), size)
		isLower             = comp.InsertCommit(round, ifaces.ColID(ctxName("IS_LOWER")), size)
	)

	res := &oneLimbCmpCtx{
		a: a, b: b, numBits: numBits,
		isGreater: isGreater, isEqual: isEqual, isLower: isLower,
		internalProverAction: []wizard.ProverAction{isEqualCtx},
		rangeChecked:         rangeChecked,
	}

	comp.InsertRange(
		round,
		ifaces.QueryID(ctxName("RANGE_CHECK")),
		rangeChecked,
		1<<numBits,
	)

	comp.InsertGlobal(
		round,
		ifaces.QueryID(ctxName("IS_GREATER_IS_BINARY")),
		sym.Mul(isGreater, sym.Sub(isGreater, 1)),
	)

	comp.InsertGlobal(
		round,
		ifaces.QueryID(ctxName("IS_LOWER_IS_BINARY")),
		sym.Mul(isLower, sym.Sub(isLower, 1)),
	)

	comp.InsertGlobal(
		round,
		ifaces.QueryID(ctxName("FLAGS_ARE_MUTUALLY_EXCLUSIVE")),
		sym.Add(isLower, isEqual, isGreater, -1),
	)

	comp.InsertGlobal(
		round,
		ifaces.QueryID(ctxName("RANGE_CHECKED_CORRECTLY_ASSIGNED")),
		sym.Sub(
			rangeChecked,
			sym.Mul(isLower, sym.Sub(b, a)),
			sym.Mul(isGreater, sym.Sub(a, b)),
		),
	)

	return isGreater, isEqual, isLower, res
}

// Run implements the [wizard.ProverAction] interface.
func (ol *oneLimbCmpCtx) Run(run *wizard.ProverRuntime) {

	// This assigns isEqual. Therefore, only isGreater, isLower and rangeChecked
	// need to be assigned.
	ol.internalProverAction[0].Run(run)

	var (
		a      = ol.a.GetColAssignment(run)
		b      = ol.b.GetColAssignment(run)
		length = a.Len()
		g      = make([]field.Element, length)
		l      = make([]field.Element, length)
		rc     = make([]field.Element, length)
	)

	for i := 0; i < a.Len(); i++ {
		var (
			aiF, biF = a.Get(i), b.Get(i)
			ai, bi   = aiF.Uint64(), biF.Uint64()
		)

		if ai > bi {
			g[i].SetOne()
			rc[i].Sub(&aiF, &biF)
		}

		if ai < bi {
			l[i].SetOne()
			rc[i].Sub(&biF, &aiF)
		}
	}

	run.AssignColumn(ol.isGreater.GetColID(), smartvectors.NewRegular(g))
	run.AssignColumn(ol.isLower.GetColID(), smartvectors.NewRegular(l))
	run.AssignColumn(ol.rangeChecked.GetColID(), smartvectors.NewRegular(rc))
}
