package execution_data_collector

import (
	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	util "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/utilities"
)

// GenericPadderPacker is used to MiMC-hash the data in LogMessages. Using a zero initial stata,
// the data in L2L1 logs must be hashed as follows: msg1Hash, msg2Hash, and so on.
// The final value of the chained hash can be retrieved as hash[ctMax[any index]]
type GenericPadderPacker struct {
	InputLimbs [common.NbLimbU128]ifaces.Column
	// The number of bytes in the limbs.
	InputNoBytes   ifaces.Column
	InputIsActive  ifaces.Column
	OutputData     ifaces.Column
	OutputIsActive ifaces.Column
}

// NewPoseidonPadderPacker returns a new GenericPadderPacker with initialized columns that are not constrained.
func NewGenericPadderPacker(comp *wizard.CompiledIOP, inputLimbs [common.NbLimbU128]ifaces.Column, inputNoBytes, inputIsActive ifaces.Column, name string) GenericPadderPacker {
	var (
		res     GenericPadderPacker
		newSize int
	)
	res.InputLimbs = inputLimbs
	res.InputNoBytes = inputNoBytes
	res.InputIsActive = inputIsActive

	newSize = res.InputLimbs[0].Size() * common.NbLimbU128
	res.OutputData = util.CreateCol(name, "OUTPUT_DATA", newSize, comp)
	res.OutputIsActive = util.CreateCol(name, "OUTPUT_IS_ACTIVE", newSize, comp)

	return res
}

// DefineHasher specifies the constraints of the GenericPadderPacker with respect to the ExtractedData fetched from the arithmetization
func DefineGenericPadderPacker(comp *wizard.CompiledIOP, ppp GenericPadderPacker, name string) {
}

// AssignHasher assigns the data in the GenericPadderPacker using the ExtractedData fetched from the arithmetization
func AssignGenericPadderPacker(run *wizard.ProverRuntime, ppp GenericPadderPacker) {
	outputData := make([]field.Element, ppp.OutputData.Size())
	outputIsActive := make([]field.Element, ppp.OutputData.Size())
	bytesVectpr := make([]byte, ppp.OutputData.Size()*common.NbLimbU128)
	counterNoBytes := 0
	for i := 0; i < ppp.InputLimbs[0].Size(); i++ {
		isActive := ppp.InputIsActive.GetColAssignmentAt(run, i)
		nBytesLimb := ppp.InputNoBytes.GetColAssignmentAt(run, i)
		nBytesLimbInt := int(nBytesLimb.Uint64())
		remainingNBytes := nBytesLimbInt
		if isActive.IsOne() {
			for j := 0; j < common.NbLimbU128; j++ {
				limbValue := ppp.InputLimbs[j].GetColAssignmentAt(run, i)
				limbBytes := limbValue.Bytes()
				for b := 0; b < min(remainingNBytes, 4); b++ {
					bytesVectpr[counterNoBytes] = limbBytes[b]
					counterNoBytes++
				}
				remainingNBytes -= 4
			}
		}
	}
	if counterNoBytes%2 == 1 {
		bytesVectpr[counterNoBytes] = 0
		counterNoBytes++
	}
	// now pack the bytes into field elements of OUTPUT_DATA
	outputIndex := 0
	nbFieldElements := counterNoBytes / 2
	for i := 0; i < nbFieldElements; i++ {
		var fe field.Element
		fe.SetBytes([]byte{bytesVectpr[2*i], bytesVectpr[2*i+1]})
		outputData[outputIndex] = fe
		outputIsActive[outputIndex].SetOne()
		outputIndex++
	}
	run.AssignColumn(ppp.OutputData.GetColID(), sv.NewRegular(outputData))
	run.AssignColumn(ppp.OutputIsActive.GetColID(), sv.NewRegular(outputIsActive))

}
