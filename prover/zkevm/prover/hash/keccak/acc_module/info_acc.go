package gen_acc

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	projection "github.com/consensys/linea-monorepo/prover/protocol/dedicated/projection"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	commonconstraints "github.com/consensys/linea-monorepo/prover/zkevm/prover/common/common_constraints"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
)

// The sub-module GenericInfoAccumulator filters the data from different [generic.GenInfoModule],
//
//	and stitch them together to build a single module.
type GenericInfoAccumulator struct {
	Inputs *GenericAccumulatorInputs
	// stitching of modules together
	// HashHi and HashLo are over the same row
	// isHashHi = isHashLo = IsActive
	Provider generic.GenInfoModule

	// filter indicating where each original module is located over the stitched one
	sFilters []ifaces.Column

	// the active part of the stitching module
	IsActive ifaces.Column

	// max number of rows for the stitched module
	size int
}

func NewGenericInfoAccumulator(comp *wizard.CompiledIOP, inp GenericAccumulatorInputs) *GenericInfoAccumulator {
	info := &GenericInfoAccumulator{
		size:   utils.NextPowerOfTwo(inp.MaxNumKeccakF),
		Inputs: &inp,
	}

	// declare columns
	info.declareColumns(comp, len(inp.ProvidersInfo))

	// sFilter[i] starts immediately after sFilters[i-1].
	s := sym.NewConstant(0)
	for i := 0; i < len(info.sFilters); i++ {
		commonconstraints.MustBeActivationColumns(comp, info.sFilters[i], sym.Sub(1, s))
		s = sym.Add(s, info.sFilters[i])
	}

	comp.InsertGlobal(0, ifaces.QueryIDf("ADDs_UP_TO_IS_ACTIVE_Info"),
		sym.Sub(s, info.IsActive))

	// by the constraints over sFilter, and the following, we have that isActive is an Activation column.
	commonconstraints.MustBeBinary(comp, info.IsActive)

	// projection among providers and stitched module
	for i, gbm := range info.Inputs.ProvidersInfo {

		projection.InsertProjection(comp, ifaces.QueryIDf("Stitch_Modules_Hi_%v", i),
			[]ifaces.Column{gbm.HashHi},
			[]ifaces.Column{info.Provider.HashHi},
			gbm.IsHashHi,
			info.sFilters[i],
		)

		projection.InsertProjection(comp, ifaces.QueryIDf("Stitch_Modules_Lo_%v", i),
			[]ifaces.Column{gbm.HashLo},
			[]ifaces.Column{info.Provider.HashLo},
			gbm.IsHashLo,
			info.sFilters[i],
		)
	}
	return info
}

// declare columns
func (info *GenericInfoAccumulator) declareColumns(comp *wizard.CompiledIOP, nbProviders int) {
	createCol := common.CreateColFn(comp, GENERIC_ACCUMULATOR, info.size)

	info.IsActive = createCol("IsActive_Info")

	info.sFilters = make([]ifaces.Column, nbProviders)
	for i := 0; i < nbProviders; i++ {
		info.sFilters[i] = createCol("sFilterOut_%v", i)
	}

	info.Provider.HashHi = createCol("Hash_Hi")
	info.Provider.HashLo = createCol("Hash_Lo")

	info.Provider.IsHashHi = info.IsActive
	info.Provider.IsHashLo = info.IsActive
}

func (info *GenericInfoAccumulator) Run(run *wizard.ProverRuntime) {
	// fetch the witnesses of gbm

	providers := info.Inputs.ProvidersInfo
	asb := make([]infoAssignmentBuilder, len(providers))
	for i := range providers {
		asb[i].hashHi = providers[i].HashHi.GetColAssignment(run).IntoRegVecSaveAlloc()
		asb[i].hashLo = providers[i].HashLo.GetColAssignment(run).IntoRegVecSaveAlloc()
		asb[i].isHashHi = providers[i].IsHashHi.GetColAssignment(run).IntoRegVecSaveAlloc()
		asb[i].isHashLo = providers[i].IsHashLo.GetColAssignment(run).IntoRegVecSaveAlloc()
	}

	sFilters := make([][]field.Element, len(providers))
	for i := range providers {

		filter := asb[i].isHashHi
		// populate sFilters
		for j := range sFilters {
			for k := range filter {
				if filter[k] == field.One() {
					if j == i {
						sFilters[j] = append(sFilters[j], field.One())
					} else {
						sFilters[j] = append(sFilters[j], field.Zero())
					}
				}
			}

		}

	}

	//assign sFilters
	for i := range providers {
		run.AssignColumn(info.sFilters[i].GetColID(), smartvectors.RightZeroPadded(sFilters[i], info.size))
	}

	// populate and assign isActive
	isActive := vector.Repeat(field.One(), len(sFilters[0]))
	run.AssignColumn(info.IsActive.GetColID(), smartvectors.RightZeroPadded(isActive, info.size))

	// populate Provider
	var sHashHi, sHashLo []field.Element
	for i := range providers {
		filterHi := asb[i].isHashHi
		filterLo := asb[i].isHashLo
		hashHi := asb[i].hashHi
		hashLo := asb[i].hashLo
		for j := range filterHi {
			if filterHi[j] == field.One() {
				sHashHi = append(sHashHi, hashHi[j])
			}
			if filterLo[j] == field.One() {
				sHashLo = append(sHashLo, hashLo[j])
			}
		}
	}

	run.AssignColumn(info.Provider.HashHi.GetColID(), smartvectors.RightZeroPadded(sHashHi, info.size))
	run.AssignColumn(info.Provider.HashLo.GetColID(), smartvectors.RightZeroPadded(sHashLo, info.size))

}

type infoAssignmentBuilder struct {
	hashHi, hashLo     []field.Element
	isHashHi, isHashLo []field.Element
}
