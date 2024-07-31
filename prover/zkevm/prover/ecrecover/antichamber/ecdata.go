package antichamber

import (
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/dedicated/projection"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

var (
	SOURCE_ECRECOVER = field.NewElement(0)
	SOURCE_TX        = field.NewElement(1)
)

const (
	NB_ECRECOVER_INPUTS = 24
	NB_TX_INPUTS        = 15
)

type EcRecover struct {
	EcRecoverID     ifaces.Column
	Limb            ifaces.Column
	SuccessBit      ifaces.Column
	EcRecoverIndex  ifaces.Column
	EcRecoverIsData ifaces.Column
	EcRecoverIsRes  ifaces.Column

	// mask for selecting columns in this submodule for projection query
	AuxProjectionMask ifaces.Column

	*Limits
}

type ecDataSource struct {
	CsEcrecover ifaces.Column
	ID          ifaces.Column
	Limb        ifaces.Column
	SuccessBit  ifaces.Column
	Index       ifaces.Column
	IsData      ifaces.Column
	IsRes       ifaces.Column
}

func (ecSrc *ecDataSource) nbActualInstances(run *wizard.ProverRuntime) int {
	var maxId uint64
	csCol := ecSrc.CsEcrecover.GetColAssignment(run)
	idCol := ecSrc.ID.GetColAssignment(run)
	for i := 0; i < csCol.Len(); i++ {
		sel := csCol.Get(i)
		if sel.IsOne() {
			id := idCol.Get(i)
			if !id.IsUint64() {
				utils.Panic("source ecrecover id must be uint64")
			}
			if id.Uint64() > maxId {
				maxId = id.Uint64()
			}
		}
	}
	return int(maxId) + 1
}

func newEcRecover(comp *wizard.CompiledIOP, limits *Limits, src *ecDataSource) *EcRecover {
	createCol := createColFn(comp, NAME_ECRECOVER, limits.sizeAntichamber())
	res := &EcRecover{
		EcRecoverID:       createCol("ECRECOVER_ID"),
		Limb:              createCol("LIMB"),
		SuccessBit:        createCol("SUCCESS_BIT"),
		EcRecoverIndex:    createCol("ECRECOVER_INDEX"),
		EcRecoverIsData:   createCol("ECRECOVER_IS_DATA"),
		EcRecoverIsRes:    createCol("ECRECOVER_IS_RES"),
		AuxProjectionMask: createCol("AUX_PROJECTION_MASK"),

		Limits: limits,
	}
	res.csEcDataProjection(comp, src)
	res.csConstraintAuxProjectionMask(comp)

	return res
}

func (ec *EcRecover) Assign(run *wizard.ProverRuntime, src *ecDataSource) {
	ec.assignFromEcDataSource(run, src)
}

func (ec *EcRecover) assignFromEcDataSource(run *wizard.ProverRuntime, src *ecDataSource) {
	sourceCsEcRecover := run.GetColumn(src.CsEcrecover.GetColID())
	sourceID := run.GetColumn(src.ID.GetColID())
	sourceLimb := run.GetColumn(src.Limb.GetColID())
	sourceSuccessBit := run.GetColumn(src.SuccessBit.GetColID())
	sourceIndex := run.GetColumn(src.Index.GetColID())
	sourceIsData := run.GetColumn(src.IsData.GetColID())
	sourceIsRes := run.GetColumn(src.IsRes.GetColID())

	if sourceCsEcRecover.Len() != sourceID.Len() ||
		sourceID.Len() != sourceLimb.Len() ||
		sourceLimb.Len() != sourceSuccessBit.Len() ||
		sourceSuccessBit.Len() != sourceIndex.Len() ||
		sourceIndex.Len() != sourceIsData.Len() ||
		sourceIsData.Len() != sourceIsRes.Len() {
		panic("all source columns must have the same length")
	}
	var resEcRecoverID, resLimb, resSuccessBit, resEcRecoverIndex, resEcRecoverIsData, resEcRecoverIsRes, resAuxProjectionMask []field.Element
	var rowEcRecoverID, rowLimb, rowSuccessBit, rowEcRecoverIndex, rowEcRecoverIsData, rowEcRecoverIsRes, rowAuxProjectionMask [nbRowsPerEcRec]field.Element
	nbInstances := src.nbActualInstances(run)
	for i := 0; i < nbInstances; i++ {
		if i*nbRowsPerEcRecFetching >= sourceCsEcRecover.Len() {
			break
		}
		selected := sourceCsEcRecover.Get(i * nbRowsPerEcRecFetching)
		if selected.IsZero() {
			continue
		}
		for j := 0; j < nbRowsPerEcRecFetching; j++ {
			sourceIdx := i*nbRowsPerEcRecFetching + j
			rowEcRecoverID[j] = sourceID.Get(sourceIdx)
			rowLimb[j] = sourceLimb.Get(sourceIdx)
			rowSuccessBit[j] = sourceSuccessBit.Get(sourceIdx)
			rowEcRecoverIndex[j] = sourceIndex.Get(sourceIdx)
			rowEcRecoverIsData[j] = sourceIsData.Get(sourceIdx)
			rowEcRecoverIsRes[j] = sourceIsRes.Get(sourceIdx)
			rowAuxProjectionMask[j] = sourceCsEcRecover.Get(sourceIdx)
		}
		resEcRecoverID = append(resEcRecoverID, rowEcRecoverID[:]...)
		resLimb = append(resLimb, rowLimb[:]...)
		resSuccessBit = append(resSuccessBit, rowSuccessBit[:]...)
		resEcRecoverIndex = append(resEcRecoverIndex, rowEcRecoverIndex[:]...)
		resEcRecoverIsData = append(resEcRecoverIsData, rowEcRecoverIsData[:]...)
		resEcRecoverIsRes = append(resEcRecoverIsRes, rowEcRecoverIsRes[:]...)
		resAuxProjectionMask = append(resAuxProjectionMask, rowAuxProjectionMask[:]...)
	}

	// assign this submodule components
	size := ec.sizeAntichamber()
	run.AssignColumn(ec.EcRecoverID.GetColID(), smartvectors.RightZeroPadded(resEcRecoverID, size))
	run.AssignColumn(ec.Limb.GetColID(), smartvectors.RightZeroPadded(resLimb, size))
	run.AssignColumn(ec.SuccessBit.GetColID(), smartvectors.RightZeroPadded(resSuccessBit, size))
	run.AssignColumn(ec.EcRecoverIndex.GetColID(), smartvectors.RightZeroPadded(resEcRecoverIndex, size))
	run.AssignColumn(ec.EcRecoverIsData.GetColID(), smartvectors.RightZeroPadded(resEcRecoverIsData, size))
	run.AssignColumn(ec.EcRecoverIsRes.GetColID(), smartvectors.RightZeroPadded(resEcRecoverIsRes, size))
	run.AssignColumn(ec.AuxProjectionMask.GetColID(), smartvectors.RightZeroPadded(resAuxProjectionMask, size))
}

func (ec *EcRecover) csEcDataProjection(comp *wizard.CompiledIOP, src *ecDataSource) {
	projection.InsertProjection(comp, ifaces.QueryIDf("%v_PROJECT_ECDATA", NAME_ECRECOVER),
		[]ifaces.Column{ec.EcRecoverID, ec.Limb, ec.SuccessBit, ec.EcRecoverIndex, ec.EcRecoverIsData, ec.EcRecoverIsRes},
		[]ifaces.Column{src.ID, src.Limb, src.SuccessBit, src.Index, src.IsData, src.IsRes},
		ec.AuxProjectionMask, src.CsEcrecover,
	)
}

func (ec *EcRecover) csConstraintAuxProjectionMask(comp *wizard.CompiledIOP) {
	mustBeBinary(comp, ec.AuxProjectionMask)
}

// TODO: must be called from the antichamber to ensure that the mask is consistent with the column in the root antichamber
func (ec *EcRecover) csConstrainAuxProjectionMaskConsistency(comp *wizard.CompiledIOP, sourceCol, isFetchingCol ifaces.Column) {
	comp.InsertGlobal(ROUND_NR, ifaces.QueryIDf("%v_%v", NAME_ECRECOVER, "CONSISTENCY_AUX_PROJECTION_MASK"),
		sym.Sub(sym.Mul(isFetchingCol, sym.Sub(1, sourceCol)), ec.AuxProjectionMask),
	)
}
