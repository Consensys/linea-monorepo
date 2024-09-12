package packing

import (
	"math/big"

	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

type lookUpTables struct {
	//colNumber:=(1,,..,16) and colPowers:=(2^(8*1),...,2^(8*16))
	colNumber ifaces.Column
	colPowers ifaces.Column
}

// It commits to the lookUp tables used by dataTransfer module.
func NewLookupTables(comp *wizard.CompiledIOP) lookUpTables {
	res := lookUpTables{}
	// table for powers of numbers (used for decomposition of clean limbs)
	colNum, colPower2 := numToPower2(MAXNBYTE)
	res.colNumber = comp.InsertPrecomputed(ifaces.ColIDf("LookUp_Num"), colNum)
	res.colPowers = comp.InsertPrecomputed(ifaces.ColIDf("LookUp_Powers"), colPower2)

	return res
}

func numToPower2(n int) (colNum, colPower2 smartvectors.SmartVector) {
	var num, power2 []field.Element
	var res field.Element
	for i := 0; i <= n; i++ {
		num = append(num, field.NewElement(uint64(i)))
		res.Exp(field.NewElement(POWER8), big.NewInt(int64(i)))
		power2 = append(power2, res)
	}
	size := utils.NextPowerOfTwo(n + 1)
	return smartvectors.RightZeroPadded(num, size),
		smartvectors.RightPadded(power2, field.One(), size)
}
