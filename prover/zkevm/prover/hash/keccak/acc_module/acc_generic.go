// The accumulator package is responsible for accumulating the data from different arithmetization module.
package gen_acc

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	projection "github.com/consensys/linea-monorepo/prover/protocol/dedicated/projection"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
)

const (
	blockSize           = 136 // number of bytes in the block
	GENERIC_ACCUMULATOR = "GENERIC_ACCUMULATOR"
)

type GenericAccumulatorInputs struct {
	MaxNumKeccakF int
	Providers     []generic.GenDataModule
}

// The sub-module GenericAccumulator filters the data from different [generic.GenDataModule],
//
//	and stitch them together to build a single module.
type GenericAccumulator struct {
	Inputs *GenericAccumulatorInputs
	// stitching of modules together
	Provider generic.GenDataModule

	// filter indicating where each original module is located over the stitched one
	sFilters []ifaces.Column

	// the active part of the stitching module
	IsActive ifaces.Column

	// max number of rows for the stitched module
	size int
}

// It declares the new columns and the constraints among them
func NewGenericAccumulator(comp *wizard.CompiledIOP, inp GenericAccumulatorInputs) *GenericAccumulator {

	d := &GenericAccumulator{
		size:   utils.NextPowerOfTwo(inp.MaxNumKeccakF * blockSize),
		Inputs: &inp,
	}
	// declare the new columns per gbm
	d.declareColumns(comp, len(inp.Providers))

	// constraints over sFilters
	//
	// 1. they are binary
	//
	// 2. they do'nt overlap isActive =  \sum sFilters
	//
	// 3. sFilter[i] starts immediately after sFilters[i-1].
	isActive := symbolic.NewConstant(0)
	for i := range d.sFilters {
		comp.InsertGlobal(0, ifaces.QueryIDf("sFilter_IsBinary_%v", i),
			symbolic.Mul(d.sFilters[i], symbolic.Sub(1, d.sFilters[i])))
		isActive = symbolic.Add(d.sFilters[i], isActive)
	}
	comp.InsertGlobal(0, ifaces.QueryIDf("sFilters_NoOverlap"), symbolic.Sub(d.IsActive, isActive))

	// for constraint 3; over (1-\sum_{j<i} sFilters[j])*isActive we need that,
	// sFilters[i] have the form (oneThenZeros) namely, it start from ones followed by zeroes.
	s := symbolic.NewConstant(0)
	for i := range d.sFilters {
		// over (1-s)*isActive, sFilter[i] is oneThenZero
		// sFilter[i] is oneThenZero is equivalent with b (in the following) is binary
		b := symbolic.Sub(d.sFilters[i], column.Shift(d.sFilters[i], 1)) // should be binary
		comp.InsertGlobal(0, ifaces.QueryIDf("IsOne_ThenZero_%v", i),
			symbolic.Mul(symbolic.Sub(1, s), d.IsActive, symbolic.Mul(symbolic.Sub(1, b), b)))
		s = symbolic.Add(s, d.sFilters[i])
	}

	// projection among providers and stitched module
	for i, gbm := range d.Inputs.Providers {

		projection.InsertProjection(comp, ifaces.QueryIDf("Stitch_Modules_%v", i),
			[]ifaces.Column{gbm.HashNum, gbm.Limb, gbm.NBytes, gbm.Index},
			[]ifaces.Column{d.Provider.HashNum, d.Provider.Limb, d.Provider.NBytes, d.Provider.Index},
			gbm.TO_HASH,
			d.sFilters[i],
		)
	}

	// constraints over isActive
	// 1. it is binary
	// 2. it is zero followed by ones// constraints over isActive
	comp.InsertGlobal(0, ifaces.QueryIDf("IsActive_IsBinary_DataTrace"),
		symbolic.Mul(d.IsActive, symbolic.Sub(1, isActive)))

	col := symbolic.Sub(column.Shift(d.IsActive, 1), d.IsActive) // should be binary
	comp.InsertGlobal(0, ifaces.QueryIDf("IsOneThenZero_DataTrace"),
		symbolic.Mul(col, symbolic.Sub(1, col)))
	return d
}

// It declares the columns specific to the DataModule
func (d *GenericAccumulator) declareColumns(comp *wizard.CompiledIOP, nbProviders int) {
	createCol := common.CreateColFn(comp, GENERIC_ACCUMULATOR, d.size)

	d.sFilters = make([]ifaces.Column, nbProviders)
	for i := 0; i < nbProviders; i++ {
		d.sFilters[i] = createCol("sFilter_%v", i)
	}

	d.IsActive = createCol("IsActive")
	d.Provider.HashNum = createCol("sHashNum")
	d.Provider.Limb = createCol("sLimb")
	d.Provider.NBytes = createCol("sNBytes")
	d.Provider.Index = createCol("sIndex")
	d.Provider.TO_HASH = d.IsActive
}

// It assigns the columns specific to the submodule.
func (d *GenericAccumulator) Run(run *wizard.ProverRuntime) {
	// fetch the gbm witnesses
	providers := d.Inputs.Providers
	asb := make([]assignmentBuilder, len(providers))
	for i := range providers {
		asb[i].HashNum = providers[i].HashNum.GetColAssignment(run).IntoRegVecSaveAlloc()
		asb[i].Limb = providers[i].Limb.GetColAssignment(run).IntoRegVecSaveAlloc()
		asb[i].NBytes = providers[i].NBytes.GetColAssignment(run).IntoRegVecSaveAlloc()
		asb[i].Index = providers[i].Index.GetColAssignment(run).IntoRegVecSaveAlloc()
		asb[i].TO_HASH = providers[i].TO_HASH.GetColAssignment(run).IntoRegVecSaveAlloc()
	}

	sFilters := make([][]field.Element, len(providers))
	for i := range providers {
		// remember that gt is the providers assignment removing the padded part
		filter := asb[i].TO_HASH

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
		run.AssignColumn(d.sFilters[i].GetColID(), smartvectors.LeftZeroPadded(sFilters[i], d.size))
	}

	// populate and assign isActive
	isActive := vector.Repeat(field.One(), len(sFilters[0]))
	run.AssignColumn(d.IsActive.GetColID(), smartvectors.LeftZeroPadded(isActive, d.size))

	// populate sModule
	var sHashNum, sLimb, sNBytes, sIndex []field.Element
	for i := range providers {
		filter := asb[i].TO_HASH
		hashNum := asb[i].HashNum
		limb := asb[i].Limb
		nBytes := asb[i].NBytes
		index := asb[i].Index
		for j := range filter {
			if filter[j] == field.One() {
				sHashNum = append(sHashNum, hashNum[j])
				sLimb = append(sLimb, limb[j])
				sNBytes = append(sNBytes, nBytes[j])
				sIndex = append(sIndex, index[j])

			}
		}
	}

	run.AssignColumn(d.Provider.HashNum.GetColID(), smartvectors.LeftZeroPadded(sHashNum, d.size))
	run.AssignColumn(d.Provider.Limb.GetColID(), smartvectors.LeftZeroPadded(sLimb, d.size))
	run.AssignColumn(d.Provider.NBytes.GetColID(), smartvectors.LeftZeroPadded(sNBytes, d.size))
	run.AssignColumn(d.Provider.Index.GetColID(), smartvectors.LeftZeroPadded(sIndex, d.size))

}

// GenDataModule collects the columns summarizing the informations about the
// data to hash.
type assignmentBuilder struct {
	HashNum, Index, Limb []field.Element
	NBytes, TO_HASH      []field.Element
}
