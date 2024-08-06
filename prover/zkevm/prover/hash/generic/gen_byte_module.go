package generic

import (
	"bytes"

	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
)

// GenericByteModule encodes the limbs with a left alignment approach
const GEN_LEFT_ALIGNMENT = 16

// Generic byte module as specified by the arithmetization. It contains two set
// of columns, Data and Info. This module can be used for various concrete
// arithmetization modules that call the keccak module.
type GenericByteModule struct {
	// Data module summarizing informations about the data to hash. I
	Data GenDataModule
	// Info module summarizing informations about the hash as a whole
	Info GenInfoModule
}

// GenDataModule collects the columns summarizing the informations about the
// data to hash.
type GenDataModule struct {
	HashNum ifaces.Column // identifier for the hash
	Index   ifaces.Column // identifier for the current limb
	Limb    ifaces.Column // the content of the limb to hash
	NBytes  ifaces.Column // indicates the size of the current limb
	TO_HASH ifaces.Column
}

// GenInfoModule collects the columns summarizing information about the hash as
// as a whole.
type GenInfoModule struct {
	// Identifier for the hash. Allows joining with the data module
	HashNum  ifaces.Column
	HashLo   ifaces.Column
	HashHi   ifaces.Column
	IsHashLo ifaces.Column
	IsHashHi ifaces.Column
}

// the assignment to GenericByteModule
type GenTrace struct {
	HashNum, Index, Limb, NByte []field.Element
	TO_HASH                     []field.Element
	//used for the assignment to the leftovers
	CleanLimb []field.Element

	// info Trace
	HashLo, HashHi, HashNum_Info []field.Element
	IsHashLo, IsHashHi           []field.Element
}

// Creates a new GenericByteModule from the given definition
func NewGenericByteModule(
	comp *wizard.CompiledIOP,
	definition GenericByteModuleDefinition,
) GenericByteModule {

	// Load the mandatory fields
	res := GenericByteModule{
		Data: GenDataModule{
			HashNum: comp.Columns.GetHandle(definition.Data.HashNum),
			Index:   comp.Columns.GetHandle(definition.Data.Index),
			Limb:    comp.Columns.GetHandle(definition.Data.Limb),
			NBytes:  comp.Columns.GetHandle(definition.Data.NBytes),
		},
	}

	// Optionally load the info module. Not every module has an info module
	if definition.Info != (InfoDef{}) {
		res.Info = GenInfoModule{
			HashNum:  comp.Columns.GetHandle(definition.Info.HashNum),
			HashLo:   comp.Columns.GetHandle(definition.Info.HashLo),
			HashHi:   comp.Columns.GetHandle(definition.Info.HashHi),
			IsHashLo: comp.Columns.GetHandle(definition.Info.IsHashLo),
			IsHashHi: comp.Columns.GetHandle(definition.Info.IsHashHi),
		}
	}

	if len(definition.Data.TO_HASH) > 0 {
		res.Data.TO_HASH = comp.Columns.GetHandle(definition.Data.TO_HASH)
	}

	return res
}

// ScanStream scans the receiver GenDataModule's assignment and returns the list
// of the byte stream encoded in the assignment.
func (gdm *GenDataModule) ScanStreams(run *wizard.ProverRuntime) [][]byte {

	var (
		numRow      = gdm.Limb.Size()
		limbs       = gdm.Limb.GetColAssignment(run).IntoRegVecSaveAlloc()
		toHash      = gdm.TO_HASH.GetColAssignment(run).IntoRegVecSaveAlloc()
		hashNum     = gdm.HashNum.GetColAssignment(run).IntoRegVecSaveAlloc()
		nByte       = gdm.NBytes.GetColAssignment(run).IntoRegVecSaveAlloc()
		streams     = [][]byte(nil)
		buffer      = &bytes.Buffer{}
		currHashNum field.Element
	)

	for row := 0; row < numRow; row++ {

		if toHash[row].IsZero() {
			continue
		}

		if hashNum[row] != currHashNum {
			if !currHashNum.IsZero() {
				streams = append(streams, buffer.Bytes())
				buffer = &bytes.Buffer{}
			}
			currHashNum = hashNum[row]
		}

		var (
			currLimbLA  = limbs[row].Bytes() // LA = left-aligned on the 16-th byte
			currNbBytes = nByte[row].Uint64()
			currLimb    = currLimbLA[16 : 16+currNbBytes]
		)

		buffer.Write(currLimb)
	}

	streams = append(streams, buffer.Bytes())
	return streams
}
