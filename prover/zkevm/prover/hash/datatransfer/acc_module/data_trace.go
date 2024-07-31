// The accumulator package is responsible for accumulating the data from different arithmetization module
// The accumulated data is then set to the datatransfer module to be prepared for keccak hash.
package acc_module

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
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
)

const (
	blockSize = 136 // number of bytes in the block
)

// The sub-module DataModule filters the data from different arithmetization module,
//
//	and stitch them together to build a single module.
type DataModule struct {
	// stitching of modules together
	Provider generic.GenericByteModule

	// filter indicating where each original module is located over the stitched one
	sFilters []ifaces.Column

	// the active part of the stitching module
	isActive ifaces.Column

	// max number of rows for the stitched module
	MaxNumRows int
}

// It declares the new columns and the constraints among them
func (d *DataModule) NewDataModule(
	comp *wizard.CompiledIOP,
	round int,
	maxNumKeccakf int,
	gbms []generic.GenericByteModule,
) {
	d.MaxNumRows = utils.NextPowerOfTwo(maxNumKeccakf * blockSize)
	// declare the new columns per gbm
	d.declareColumns(comp, round, gbms)

	// constraints over sFilters
	//
	// 1. they are binary
	//
	// 2. they do'nt overlap isActive =  \sum sFilters
	//
	// 3. sFilter[i] starts immediately after sFilters[i-1].
	isActive := symbolic.NewConstant(0)
	for i := range d.sFilters {
		comp.InsertGlobal(round, ifaces.QueryIDf("sFilter_IsBinary_%v", i),
			symbolic.Mul(d.sFilters[i], symbolic.Sub(1, d.sFilters[i])))
		isActive = symbolic.Add(d.sFilters[i], isActive)
	}
	comp.InsertGlobal(round, ifaces.QueryIDf("sFilters_NoOverlap"), symbolic.Sub(d.isActive, isActive))

	// for constraint 3; over (1-\sum_{j<i} sFilters[j])*isActive we need that,
	// sFilters[i] have the form (oneThenZeros) namely, it start from ones followed by zeroes.
	s := symbolic.NewConstant(0)
	for i := range d.sFilters {
		// over (1-s)*isActive, sFilter[i] is oneThenZero
		// sFilter[i] is oneThenZero is equivalent with b (in the following) is binary
		b := symbolic.Sub(d.sFilters[i], column.Shift(d.sFilters[i], 1)) // should be binary
		comp.InsertGlobal(round, ifaces.QueryIDf("IsOne_ThenZero_%v", i),
			symbolic.Mul(symbolic.Sub(1, s), d.isActive, symbolic.Mul(symbolic.Sub(1, b), b)))
		s = symbolic.Add(s, d.sFilters[i])
	}

	// projection among gbms and stitched module
	for i, gbm := range gbms {

		projection.InsertProjection(comp, ifaces.QueryIDf("Stitch_Modules_%v", i),
			[]ifaces.Column{gbm.Data.HashNum, gbm.Data.Limb, gbm.Data.NBytes, gbm.Data.Index},
			[]ifaces.Column{d.Provider.Data.HashNum, d.Provider.Data.Limb, d.Provider.Data.NBytes, d.Provider.Data.Index},
			gbm.Data.TO_HASH,
			d.sFilters[i],
		)
	}

	// constraints over isActive
	// 1. it is binary
	// 2. it is zero followed by ones// constraints over isActive
	comp.InsertGlobal(round, ifaces.QueryIDf("IsActive_IsBinary_DataTrace"),
		symbolic.Mul(d.isActive, symbolic.Sub(1, isActive)))

	col := symbolic.Sub(column.Shift(d.isActive, 1), d.isActive) // should be binary
	comp.InsertGlobal(round, ifaces.QueryIDf("IsOneThenZero_DataTrace"),
		symbolic.Mul(col, symbolic.Sub(1, col)))

}

// It declares the columns specific to the DataModule
func (d *DataModule) declareColumns(comp *wizard.CompiledIOP, round int, gbms []generic.GenericByteModule) {
	d.sFilters = make([]ifaces.Column, len(gbms))
	for i := range gbms {
		d.sFilters[i] = comp.InsertCommit(round, ifaces.ColIDf("sFilter_%v", i), d.MaxNumRows)
	}

	d.isActive = comp.InsertCommit(round, ifaces.ColIDf("IsActive"), d.MaxNumRows)
	d.Provider.Data.HashNum = comp.InsertCommit(round, ifaces.ColIDf("sHashNum"), d.MaxNumRows)
	d.Provider.Data.Limb = comp.InsertCommit(round, ifaces.ColIDf("sLimb"), d.MaxNumRows)
	d.Provider.Data.NBytes = comp.InsertCommit(round, ifaces.ColIDf("sNBytes"), d.MaxNumRows)
	d.Provider.Data.Index = comp.InsertCommit(round, ifaces.ColIDf("sIndex"), d.MaxNumRows)
	d.Provider.Data.TO_HASH = d.isActive
}

// It assigns the columns specific to the submodule.
func (d *DataModule) AssignDataModule(
	run *wizard.ProverRuntime,
	gbms []generic.GenericByteModule) {
	// fetch the gbm witnesses
	gt := make([]generic.GenTrace, len(gbms))
	for i := range gbms {
		gt[i].HashNum = extractColLeftPadded(run, gbms[i].Data.HashNum)
		gt[i].Limb = extractColLeftPadded(run, gbms[i].Data.Limb)
		gt[i].NByte = extractColLeftPadded(run, gbms[i].Data.NBytes)
		gt[i].Index = extractColLeftPadded(run, gbms[i].Data.Index)
		gt[i].TO_HASH = extractColLeftPadded(run, gbms[i].Data.TO_HASH)
	}

	sFilters := make([][]field.Element, len(gbms))
	for i := range gbms {
		// remember that gt is the gbms assignment removing the padded part
		filter := gt[i].TO_HASH

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
	for i := range gbms {
		run.AssignColumn(d.sFilters[i].GetColID(), smartvectors.LeftZeroPadded(sFilters[i], d.MaxNumRows))
	}

	// populate and assign isActive
	isActive := vector.Repeat(field.One(), len(sFilters[0]))
	run.AssignColumn(d.isActive.GetColID(), smartvectors.LeftZeroPadded(isActive, d.MaxNumRows))

	// populate sModule
	var sHashNum, sLimb, sNBytes, sIndex []field.Element
	for i := range gbms {
		filter := gt[i].TO_HASH
		hashNum := gt[i].HashNum
		limb := gt[i].Limb
		nBytes := gt[i].NByte
		index := gt[i].Index
		for j := range filter {
			if filter[j] == field.One() {
				sHashNum = append(sHashNum, hashNum[j])
				sLimb = append(sLimb, limb[j])
				sNBytes = append(sNBytes, nBytes[j])
				sIndex = append(sIndex, index[j])

			}
		}
	}

	run.AssignColumn(d.Provider.Data.HashNum.GetColID(), smartvectors.LeftZeroPadded(sHashNum, d.MaxNumRows))
	run.AssignColumn(d.Provider.Data.Limb.GetColID(), smartvectors.LeftZeroPadded(sLimb, d.MaxNumRows))
	run.AssignColumn(d.Provider.Data.NBytes.GetColID(), smartvectors.LeftZeroPadded(sNBytes, d.MaxNumRows))
	run.AssignColumn(d.Provider.Data.Index.GetColID(), smartvectors.LeftZeroPadded(sIndex, d.MaxNumRows))

}
