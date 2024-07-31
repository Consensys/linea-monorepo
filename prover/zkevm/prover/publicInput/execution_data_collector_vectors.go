package publicInput

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

type ExecutionDataCollectorVectors struct {
	BlockID, AbsTxID, AbsTxIDMax []field.Element
	Limb, NoBytes                []field.Element
	UnalignedLimb, AlignedPow    []field.Element
	TotalNoTxBlock               []field.Element

	IsActive                                                                       []field.Element
	IsNoTx, IsBlockHashHi, IsBlockHashLo, IsTimestamp, IsTxRLP, IsAddrHi, IsAddrLo []field.Element

	FirstAbsTxIDBlock, LastAbsTxIDBlock []field.Element

	EndOfRlpSegment []field.Element
}

func NewExecutionDataCollectorVectors(size int) *ExecutionDataCollectorVectors {
	res := &ExecutionDataCollectorVectors{
		BlockID:           make([]field.Element, size),
		AbsTxID:           make([]field.Element, size),
		AbsTxIDMax:        make([]field.Element, size),
		Limb:              make([]field.Element, size),
		NoBytes:           make([]field.Element, size),
		UnalignedLimb:     make([]field.Element, size),
		AlignedPow:        make([]field.Element, size),
		TotalNoTxBlock:    make([]field.Element, size),
		IsActive:          make([]field.Element, size),
		IsNoTx:            make([]field.Element, size),
		IsBlockHashHi:     make([]field.Element, size),
		IsBlockHashLo:     make([]field.Element, size),
		IsTimestamp:       make([]field.Element, size),
		IsTxRLP:           make([]field.Element, size),
		IsAddrHi:          make([]field.Element, size),
		IsAddrLo:          make([]field.Element, size),
		FirstAbsTxIDBlock: make([]field.Element, size),
		LastAbsTxIDBlock:  make([]field.Element, size),
		EndOfRlpSegment:   make([]field.Element, size),
	}
	return res
}

func (vect *ExecutionDataCollectorVectors) SetLimbAndUnalignedLimb(totalCt int, value field.Element, opType int) {
	var powField, limbValue field.Element
	switch opType {
	case loadNoTxn:
		powField = field.NewFromString(powBytesNoTxn)
	case loadTimestamp:
		powField = field.NewFromString(powTimestamp)
	case loadBlockHashHi:
		powField = field.NewFromString(powBlockHash)
	case loadBlockHashLo:
		powField = field.NewFromString(powBlockHash)
	case loadSenderAddrHi:
		powField = field.NewFromString(powSenderAddrHi)
	case loadSenderAddrLo:
		powField = field.NewFromString(powSenderAddrLo)
	case loadRlp:
		powField = field.NewFromString("1") // Replace this with the right value in the case of RLP
	}
	vect.AlignedPow[totalCt].Set(&powField)
	limbValue.Set(&value)
	limbValue.Mul(&limbValue, &powField)
	vect.Limb[totalCt].Set(&limbValue)
	vect.UnalignedLimb[totalCt].Set(&value)
}

func (vect *ExecutionDataCollectorVectors) SetCounters(totalCt, blockCt, absTxCt, absTxIdMax int) {
	vect.BlockID[totalCt].SetInt64(int64(blockCt + 1))
	vect.AbsTxID[totalCt].SetInt64(int64(absTxCt))
	vect.AbsTxIDMax[totalCt].SetInt64(int64(absTxIdMax))
}

func (vect *ExecutionDataCollectorVectors) SetBlockMetadata(totalCt int, totalTxBlock, firstAbsTxIDBlock, lastAbsTxIDBlock field.Element) {
	vect.TotalNoTxBlock[totalCt].Set(&totalTxBlock)
	vect.FirstAbsTxIDBlock[totalCt].Set(&firstAbsTxIDBlock)
	vect.LastAbsTxIDBlock[totalCt].Set(&lastAbsTxIDBlock)
}
