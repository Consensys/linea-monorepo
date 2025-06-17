package bls

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

type unalignedCurveMembershipData struct {
	*unalignedCurveMembershipDataSource

	// IsActive is a constructed column which indicates if the circuit is active. Set when selector is on or when we provide the input data.
	IsActive ifaces.Column
	// IsFirstLineOfInput is a constructed column which indicates if the row is
	// the first line of the input.
	IsFirstLineOfInput    ifaces.Column
	IsFirstLineOfInputAct wizard.ProverAction

	GnarkData ifaces.Column
}

type unalignedCurveMembershipDataSource struct {
	Limb              ifaces.Column
	Counter           ifaces.Column
	CsCurveMembership ifaces.Column
}

func newUnalignedCurveMembershipData(comp *wizard.CompiledIOP, g group, size int, src *unalignedCurveMembershipDataSource) *unalignedCurveMembershipData {
	createCol := createColFn(comp, fmt.Sprintf("UNALIGNED_%s_CURVE_MEMBERSHIP", g.StringCurve()), size)
	res := &unalignedCurveMembershipData{
		unalignedCurveMembershipDataSource: src,
		IsActive:                           createCol("IS_ACTIVE"),
		GnarkData:                          createCol("GNARK_DATA"),
	}
	res.IsFirstLineOfInput, res.IsFirstLineOfInputAct = dedicated.IsZero(comp, src.Counter)
	return res
}

func (d *unalignedCurveMembershipData) Assign(run *wizard.ProverRuntime) {
	d.IsFirstLineOfInputAct.Run(run)

	var (
		srcLimb    = d.Limb.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcCounter = d.Counter.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcCs      = d.CsCurveMembership.GetColAssignment(run).IntoRegVecSaveAlloc()
	)
	var (
		dstIsActive  = common.NewVectorBuilder(d.IsActive)
		dstGnarkData = common.NewVectorBuilder(d.GnarkData)
	)

	for i := 0; i < len(srcLimb); i++ {
		if srcCs[i].IsZero() {
			continue
		}
		// for the first line of input, we push the expected success bit
		if srcCounter[i].IsZero() {
			dstIsActive.PushBoolean(true)
			dstGnarkData.PushBoolean(false) // we push additional input to gnark input to indicate curve non-membership
		}
		dstIsActive.PushBoolean(true)
		dstGnarkData.PushField(srcLimb[i])
	}
	dstIsActive.PadAndAssign(run, field.Zero())
	dstGnarkData.PadAndAssign(run, field.Zero())
}
