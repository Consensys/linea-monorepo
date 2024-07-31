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
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/datatransfer/datatransfer"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
)

type InfoModule struct {
	// filtering outputs for the corresponding module
	sFilters []ifaces.Column

	// ID of the hash
	hashNumOut ifaces.Column

	isActive ifaces.Column

	maxNumRows int
}

func (info *InfoModule) NewInfoModule(
	comp *wizard.CompiledIOP,
	round, maxNumKeccakF int,
	gbms []generic.GenericByteModule,
	hashOutput datatransfer.HashOutput,
	data DataModule,
) {
	info.maxNumRows = utils.NextPowerOfTwo(maxNumKeccakF)
	size := utils.NextPowerOfTwo(maxNumKeccakF)
	// declare columns
	info.declareColumns(comp, round, size, gbms)

	// declare constraints
	// constraints over sFilters
	//
	// 1. they are binary
	//
	// 2. they do'nt overlap isActive =  \sum sFilters
	//
	// 3.  sFilter[i] starts immediately after sFilters[i-1].
	isActive := symbolic.NewConstant(0)
	for i := range info.sFilters {
		comp.InsertGlobal(round, ifaces.QueryIDf("sFilter_IsBinary_Info_%v", i),
			symbolic.Mul(info.sFilters[i], symbolic.Sub(1, info.sFilters[i])))
		isActive = symbolic.Add(info.sFilters[i], isActive)
	}
	comp.InsertGlobal(round, ifaces.QueryIDf("sFilters_NoOverlap_Info"), symbolic.Sub(info.isActive, isActive))

	// for constraint 3; over (1-\sum_{j<i} sFilters[j])*isActive we need that,
	// sFilters[i] have the form (oneThenZeros) namely, it start from ones followed by zeroes.
	s := symbolic.NewConstant(0)
	for i := range info.sFilters {
		// over (1-s)*isActive, sFilter[i] is oneThenZero
		// sFilter[i] is oneThenZero is equivalent with b (in the following) is binary
		b := symbolic.Sub(info.sFilters[i], column.Shift(info.sFilters[i], 1)) // should be binary
		comp.InsertGlobal(round, ifaces.QueryIDf("IsOne_ThenZero_Info%v", i),
			symbolic.Mul(symbolic.Sub(1, s), info.isActive, symbolic.Mul(symbolic.Sub(1, b), b)))
		s = symbolic.Add(s, info.sFilters[i])
	}

	// constraints over isActive
	// 1. It is Binary
	// 2. It has ones followed by zeroes
	comp.InsertGlobal(round, ifaces.QueryIDf("IsActive_IsBinary_InfoTrace"),
		symbolic.Mul(info.isActive, symbolic.Sub(1, isActive)))

	col := symbolic.Sub(column.Shift(info.isActive, -1), info.isActive) // should be binary
	comp.InsertGlobal(round, ifaces.QueryIDf("IsOneThenZero_InfoTrace"),
		symbolic.Mul(col, symbolic.Sub(1, col)))

	// Projection between hashOutputs
	for i := range gbms {
		projection.InsertProjection(comp, ifaces.QueryIDf("Project_HashLo_%v", i),
			[]ifaces.Column{hashOutput.HashLo},
			[]ifaces.Column{gbms[i].Info.HashLo},
			info.sFilters[i],
			gbms[i].Info.IsHashLo)

		projection.InsertProjection(comp, ifaces.QueryIDf("Project_HashHi_%v", i),
			[]ifaces.Column{hashOutput.HashHi},
			[]ifaces.Column{gbms[i].Info.HashHi},
			info.sFilters[i],
			gbms[i].Info.IsHashHi)
	}

}

// declare columns
func (info *InfoModule) declareColumns(
	comp *wizard.CompiledIOP,
	round, size int,
	gbms []generic.GenericByteModule,
) {
	info.isActive = comp.InsertCommit(round, ifaces.ColIDf("IsActive_Info"), size)

	info.sFilters = make([]ifaces.Column, len(gbms))
	for i := range gbms {
		info.sFilters[i] = comp.InsertCommit(round, ifaces.ColIDf("sFilterOut_%v", i), size)
	}
	info.hashNumOut = comp.InsertCommit(round, ifaces.ColIDf("hashNumOut"), size)
}

func (info *InfoModule) AssignInfoModule(
	run *wizard.ProverRuntime,
	gbms []generic.GenericByteModule,
) {
	// fetch the witnesses of gbm
	gt := make([]generic.GenTrace, len(gbms))
	for i := range gbms {
		gt[i].HashNum = extractColLeftPadded(run, gbms[i].Data.HashNum)
		gt[i].Limb = extractColLeftPadded(run, gbms[i].Data.Limb)
		gt[i].NByte = extractColLeftPadded(run, gbms[i].Data.NBytes)
		gt[i].Index = extractColLeftPadded(run, gbms[i].Data.Index)
		gt[i].TO_HASH = extractColLeftPadded(run, gbms[i].Data.TO_HASH)
	}

	// populate hashNumOut and sFilters
	var hashNumOut []field.Element
	sFilters := make([][]field.Element, len(gt))
	for i := range gt {
		for k := range gt[i].Index {
			if gt[i].Index[k] == field.Zero() && gt[i].TO_HASH[k] == field.One() {
				hashNumOut = append(hashNumOut, gt[i].HashNum[k])
				for j := range gt {
					if j == i {
						sFilters[j] = append(sFilters[j], field.One())
					} else {
						sFilters[j] = append(sFilters[j], field.Zero())
					}
				}
			}
		}
	}

	run.AssignColumn(info.hashNumOut.GetColID(), smartvectors.RightZeroPadded(hashNumOut, info.maxNumRows))

	for i := range gt {
		run.AssignColumn(info.sFilters[i].GetColID(), smartvectors.RightZeroPadded(sFilters[i], info.maxNumRows))
	}

	// populate and assign isActive
	isActive := vector.Repeat(field.One(), len(sFilters[0]))
	run.AssignColumn(info.isActive.GetColID(), smartvectors.RightZeroPadded(isActive, info.maxNumRows))

}
