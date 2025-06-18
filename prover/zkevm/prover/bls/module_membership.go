package bls

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	ccs "github.com/consensys/linea-monorepo/prover/zkevm/prover/common/common_constraints"
)

const (
	NAME_UNALIGNED_CURVE_MEMBERSHIP = "UNALIGNED_CURVE_MEMBERSHIP"
)

type unalignedCurveMembershipDataSource struct {
	Limb              ifaces.Column
	Counter           ifaces.Column
	CsCurveMembership ifaces.Column
}

type unalignedCurveMembershipData struct {
	*unalignedCurveMembershipDataSource

	// IsActive is a constructed column which indicates if the circuit is
	// active. Set when selector is on or when we provide the input data. It is IsActive+IsComputed
	IsActive ifaces.Column
	// IsFetching indicates that the input comes from the external source and we
	// need to check projection correctness.
	IsFetching ifaces.Column
	// IsComputed indicates that the input is computed.
	IsComputed ifaces.Column

	GnarkIndex ifaces.Column
	GnarkData  ifaces.Column

	IsFirstLine    ifaces.Column
	IsFirstLineAct wizard.ProverAction

	group
}

func newUnalignedCurveMembershipData(comp *wizard.CompiledIOP, g group, size int, src *unalignedCurveMembershipDataSource) *unalignedCurveMembershipData {
	createCol := createColFn(comp, fmt.Sprintf("UNALIGNED_%s_CURVE_MEMBERSHIP", g.StringCurve()), size)
	res := &unalignedCurveMembershipData{
		unalignedCurveMembershipDataSource: src,
		IsActive:                           createCol("IS_ACTIVE"),
		IsFetching:                         createCol("IS_FETCHING"),
		IsComputed:                         createCol("IS_COMPUTED"),
		GnarkIndex:                         createCol("GNARK_INDEX"),
		GnarkData:                          createCol("GNARK_DATA"),
		group:                              g,
	}
	res.IsFirstLine, res.IsFirstLineAct = dedicated.IsZero(comp, res.GnarkIndex)
	res.csProjection(comp)
	res.csActivationColumns(comp)
	res.csIndex(comp)
	res.csIsOnCurve(comp)
	return res
}

func (d *unalignedCurveMembershipData) Assign(run *wizard.ProverRuntime) {

	var (
		srcLimb    = d.Limb.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcCounter = d.Counter.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcCs      = d.CsCurveMembership.GetColAssignment(run).IntoRegVecSaveAlloc()
	)
	var (
		dstIsActive   = common.NewVectorBuilder(d.IsActive)
		dstGnarkData  = common.NewVectorBuilder(d.GnarkData)
		dstIsFetching = common.NewVectorBuilder(d.IsFetching)
		dstIsComputed = common.NewVectorBuilder(d.IsComputed)
		dstGnarkIndex = common.NewVectorBuilder(d.GnarkIndex)
	)

	var counter, nbLimbs int
	switch d.group {
	case G1:
		nbLimbs = nbG1Limbs
	case G2:
		nbLimbs = nbG2Limbs
	}

	for i := range len(srcLimb) {
		if srcCs[i].IsZero() {
			continue
		}
		dstGnarkIndex.PushInt(counter)
		counter = (counter + 1) % (nbLimbs + 1)
		// for the first line of input, we push the expected success bit
		if srcCounter[i].IsZero() {
			dstIsActive.PushBoolean(true)
			dstIsFetching.PushBoolean(false)
			dstIsComputed.PushBoolean(true)
			// we push additional input to gnark input to indicate curve non-membership
			dstGnarkData.PushBoolean(false)
		}
		dstIsActive.PushBoolean(true)
		dstIsFetching.PushBoolean(true)
		dstIsComputed.PushBoolean(false)
		dstGnarkData.PushField(srcLimb[i])
	}
	dstIsActive.PadAndAssign(run, field.Zero())
	dstIsFetching.PadAndAssign(run, field.Zero())
	dstIsComputed.PadAndAssign(run, field.Zero())
	dstGnarkIndex.PadAndAssign(run, field.Zero())
	dstGnarkData.PadAndAssign(run, field.Zero())
	d.IsFirstLineAct.Run(run)
}

func (d *unalignedCurveMembershipData) csProjection(comp *wizard.CompiledIOP) {
	comp.InsertProjection(
		ifaces.QueryIDf("%s_PROJECTION", NAME_UNALIGNED_CURVE_MEMBERSHIP),
		query.ProjectionInput{
			ColumnA: []ifaces.Column{d.unalignedCurveMembershipDataSource.Limb},
			ColumnB: []ifaces.Column{d.GnarkData},
			FilterA: d.unalignedCurveMembershipDataSource.CsCurveMembership,
			FilterB: d.IsFetching,
		},
	)
}

func (d *unalignedCurveMembershipData) csActivationColumns(comp *wizard.CompiledIOP) {
	ccs.MustBeMutuallyExclusiveBinaryFlags(
		comp,
		d.IsActive,
		[]ifaces.Column{d.IsComputed, d.IsFetching},
	)
}

func (d *unalignedCurveMembershipData) csIndex(comp *wizard.CompiledIOP) {
	// runs sequentially
	comp.InsertGlobal(
		ROUND_NR,
		ifaces.QueryIDf("%s_INDEX_INCREMENT", NAME_UNALIGNED_CURVE_MEMBERSHIP),
		sym.Mul(
			d.IsActive,
			d.GnarkIndex,
			sym.Sub(
				d.GnarkIndex, column.Shift(d.GnarkIndex, -1), 1,
			),
		),
	)
	// when the index is nbLimbs, then switches to zero
	comp.InsertGlobal(
		ROUND_NR,
		ifaces.QueryIDf("%s_INDEX_RESET", NAME_UNALIGNED_CURVE_MEMBERSHIP),
		sym.Mul(
			d.IsActive,
			d.IsFirstLine,
			d.GnarkIndex,
		),
	)

	// when the index is 0, then IsComputed is active
	comp.InsertGlobal(
		ROUND_NR,
		ifaces.QueryIDf("%s_IS_ON_CURVE", NAME_UNALIGNED_CURVE_MEMBERSHIP),
		sym.Mul(
			d.IsComputed,
			d.GnarkIndex,
		),
	)
}

func (d *unalignedCurveMembershipData) csIsOnCurve(comp *wizard.CompiledIOP) {
	// when computed, then isOnCurve is false. we can conditionally make the
	// check to both support IsOnCurve=false and true. Then we need to store the
	// expected value instead.
	comp.InsertGlobal(
		ROUND_NR,
		ifaces.QueryIDf("%s_ON_CURVE_CORRECNTESS", NAME_UNALIGNED_CURVE_MEMBERSHIP),
		sym.Mul(
			d.IsComputed,
			d.GnarkData,
		),
	)
}
