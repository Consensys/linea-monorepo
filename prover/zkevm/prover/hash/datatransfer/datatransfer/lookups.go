package datatransfer

import (
	"math/big"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf"
)

type lookUpTables struct {
	//colNumber:=(1,,..,16) and colPowers:=(2^(8*1),...,2^(8*16))
	colNumber ifaces.Column
	colPowers ifaces.Column

	// columns for base conversion
	colUint16 ifaces.Column
	colBaseA  ifaces.Column
	colBaseB  ifaces.Column

	// a column of single bytes
	colSingleByte ifaces.Column

	// columns for base conversion from baseBDirty to 4bit integers
	ColUint4      ifaces.Column
	ColBaseBDirty ifaces.Column

	// columns for checking the maximum number of padded bytes
	colSha2MaxPadding   ifaces.Column
	colKeccakMaxPadding ifaces.Column
}

// It commits to the lookUp tables used by dataTransfer module.
func newLookupTables(comp *wizard.CompiledIOP) lookUpTables {
	res := lookUpTables{}
	// table for powers of numbers (used for decomposition of clean limbs)
	colNum, colPower2 := numToPower2(maxNByte)
	res.colNumber = comp.InsertPrecomputed(deriveName("LookUp_Num"), colNum)
	res.colPowers = comp.InsertPrecomputed(deriveName("LookUp_Powers"), colPower2)

	// table for base conversion (used for converting blocks to what keccakf expect)
	colUint16, colBaseA, colBaseB := baseConversionKeccakBaseX()
	res.colUint16 = comp.InsertPrecomputed(ifaces.ColIDf("LOOKUP_Uint16"), colUint16)
	res.colBaseA = comp.InsertPrecomputed(ifaces.ColIDf("LOOKUP_BaseA"), colBaseA)
	res.colBaseB = comp.InsertPrecomputed(ifaces.ColIDf("LOOKUP_BaseB"), colBaseB)

	// table for decomposition to single bytes
	res.colSingleByte = comp.InsertPrecomputed(ifaces.ColIDf("LOOKUP_SingleByte"), singleByte())

	// table for base conversion (from BaseBDirty to uint4)
	colUint4, colBaseBDirty := baseConversionKeccakBaseBDirtyToUint4()
	res.ColUint4 = comp.InsertPrecomputed(ifaces.ColIDf("LOOKUP_Uint4"), colUint4)
	res.ColBaseBDirty = comp.InsertPrecomputed(ifaces.ColIDf("LOOKUP_BaseBDirty"), colBaseBDirty)

	// table to check the number of padded bytes don't overflow its maximum.
	res.colSha2MaxPadding = comp.InsertPrecomputed(ifaces.ColIDf("LOOKUP_Max_Pad_Sha2"),
		colBlockSize(maxBlockSizeSha2, 8))

	// table to check the number of padded bytes don't overflow its maximum.
	res.colKeccakMaxPadding = comp.InsertPrecomputed(ifaces.ColIDf("LOOKUP_Max_Pad_Keccak"),
		colBlockSize(maxBlockSize, 1))
	return res

}

func numToPower2(n int) (colNum, colPower2 smartvectors.SmartVector) {
	var num, power2 []field.Element
	var res field.Element
	for i := 0; i < n; i++ {
		num = append(num, field.NewElement(uint64(i)))
		res.Exp(field.NewElement(power8), big.NewInt(int64(i)))
		power2 = append(power2, res)
	}
	size := utils.NextPowerOfTwo(n + 1)
	return smartvectors.RightZeroPadded(num, size),
		smartvectors.RightPadded(power2, field.One(), size)
}

// convert slices of 16bits to keccak.BaseX
func baseConversionKeccakBaseX() (uint16Col, baseACol, baseBCol smartvectors.SmartVector) {
	var u, v, w []field.Element
	for i := 0; i < power16; i++ {
		u = append(u, field.NewElement(uint64(i)))
		v = append(v, uInt16ToBaseX(uint16(i), &keccakf.BaseAFr))
		w = append(w, uInt16ToBaseX(uint16(i), &keccakf.BaseBFr))
	}
	return smartvectors.NewRegular(u), smartvectors.NewRegular(v), smartvectors.NewRegular(w)
}

// it creates a slice of all the single bytes
func singleByte() smartvectors.SmartVector {
	var v []field.Element
	for i := 0; i < power8; i++ {
		v = append(v, field.NewElement(uint64(i)))
	}
	return smartvectors.NewRegular(v)
}

func baseConversionKeccakBaseBDirtyToUint4() (
	uint4Col, baseBDirtyCol smartvectors.SmartVector) {
	var u, v []field.Element
	for j := 0; j < keccakf.BaseBPow4; j++ {
		x := field.NewElement(uint64(j))
		uint4 := BaseBToUint4(x, keccakf.BaseB)
		u = append(u, x)
		v = append(v, field.NewElement(uint4))
	}
	n := utils.NextPowerOfTwo(keccakf.BaseBPow4)
	for i := keccakf.BaseBPow4; i < n; i++ {
		u = append(u, u[len(u)-1])
		v = append(v, v[len(v)-1])
	}
	return smartvectors.NewRegular(v), smartvectors.NewRegular(u)
}

func BaseBToUint4(x field.Element, base int) (res uint64) {
	res = 0
	decomposedF := keccakf.DecomposeFr(x, base, 4)

	bitPos := 1
	for i, limb := range decomposedF {
		bit := (limb.Uint64() >> bitPos) & 1
		res |= bit << i
	}

	return res
}

// it return a column of as (1,2,3,...., BlockSize+len(lastPadding))
// this is the accumulator for the number of padded bytes.
// for sha2 len(lastPadding) = 8 , for keccak en(lastPadding) = 1.
// it is used to check that the number of padded bytes fall in this table.
func colBlockSize(blockSize, lenLastPad int) smartvectors.SmartVector {
	n := blockSize
	var u []field.Element
	m := utils.NextPowerOfTwo(n + lenLastPad)
	for i := 0; i < n+lenLastPad; i++ {
		u = append(u, field.NewElement(uint64(i+1)))
	}
	for i := n + lenLastPad; i < m; i++ {
		u = append(u, field.NewElement(uint64(n)))
	}
	return smartvectors.NewRegular(u)
}
