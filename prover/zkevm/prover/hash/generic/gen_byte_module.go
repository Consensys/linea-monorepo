package generic

import (
	"bytes"

	"github.com/consensys/zkevm-monorepo/prover/crypto/keccak"
	"github.com/consensys/zkevm-monorepo/prover/crypto/sha2"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils"
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

// Implements the trace providing mechanism for the generic byte module.
// Optionally, generate traces for different hashes that might be applied over the limbs.
func (gen *GenericByteModule) AppendTraces(
	run *wizard.ProverRuntime,
	genTrace *GenTrace,
	trace ...interface{},
) {

	data := gen.Data
	info := gen.Info

	// Extract the assignments through a shallow copy.
	hashNum := gen.extractCol(run, gen.Data.HashNum)
	refCol := gen.Data.HashNum
	index := gen.extractCol(run, gen.Data.Index, refCol)
	limbs := gen.extractCol(run, gen.Data.Limb, refCol)
	nBytes := gen.extractCol(run, gen.Data.NBytes, refCol)
	toHash := gen.extractCol(run, data.TO_HASH, refCol)

	// if info is not empty, extract it
	if info != (GenInfoModule{}) {
		hashLo := gen.extractCol(run, info.HashLo)
		genTrace.HashLo = hashLo

		hashHi := gen.extractCol(run, info.HashHi, info.HashLo)
		genTrace.HashHi = hashHi

		isHashLo := gen.extractCol(run, info.IsHashLo, info.HashLo)
		genTrace.IsHashLo = isHashLo

		isHashHi := gen.extractCol(run, info.IsHashHi, info.HashLo)
		genTrace.IsHashHi = isHashHi

		hashNum_Info := gen.extractCol(run, info.HashNum, info.HashLo)
		genTrace.HashNum_Info = hashNum_Info

	}

	stream := bytes.Buffer{}
	streamSha2 := bytes.Buffer{}

	limbSerialized := [32]byte{}
	one := field.One()
	cleanLimb := make([]field.Element, len(hashNum))
	for pos := range hashNum {
		// Check if the current limb can be appended
		if toHash != nil && (toHash[pos] != one) {
			continue
		}

		// Sanity-check, if the index is zero must be equivalent to an empty
		// stream
		if index[pos].IsZero() != (stream.Len() == 0) && index[pos].IsZero() != (streamSha2.Len() == 0) {
			utils.Panic(
				"the index==0 should mean an empty stream, index %v, stream.Len() %v\n",
				index[pos], stream.Len())
		}

		// Extract the limb, which is left aligned to the 16-th byte
		limbSerialized = limbs[pos].Bytes()
		byteSize := nBytes[pos].Uint64()
		res := limbSerialized[GEN_LEFT_ALIGNMENT : GEN_LEFT_ALIGNMENT+byteSize]
		cleanLimb[pos].SetBytes(res)
		stream.Write(
			res,
		)

		streamSha2.Write(
			res,
		)

		// If we are on the last limb or if the hashNum increases, it means
		// that we need to close the hash to start the next one
		if pos+1 == len(hashNum) || index[pos+1].Uint64() == 0 {
			for i := range trace {
				switch v := trace[i].(type) {
				case *keccak.PermTraces:
					{
						keccak.Hash(stream.Bytes(), v)
						stream.Reset()
					}
				case *sha2.HashTraces:
					{
						sha2.Hash(streamSha2.Bytes(), v)
						streamSha2.Reset()
					}
				default:
					utils.Panic("other hashes are not supported")
				}
			}

		}

	}
	genTrace.HashNum = hashNum
	genTrace.Index = index
	genTrace.Limb = limbs
	genTrace.NByte = nBytes
	genTrace.CleanLimb = cleanLimb
	genTrace.TO_HASH = toHash

}

// Extract a shallow copy of the active zone of a column. Meaning the unpadded
// area where the column encodes actual data.
// The option refCol is used for the sanity check on the length of effective window.
func (gen *GenericByteModule) extractCol(
	run *wizard.ProverRuntime,
	col ifaces.Column, refCol ...ifaces.Column,
) []field.Element {

	// Fetches the smart-vector and delimitate the active zone. Here we assume
	// that all the columns are zero-prepended. And have the same length.

	col_ := run.Columns.MustGet(col.GetColID())
	a := smartvectors.Window(col_)
	m := smartvectors.Density(col_)

	// sanity check on the effective window; the effective part of column is the part without zero-padding.
	if len(refCol) != 0 {
		id := refCol[0].GetColID()
		refColSV := run.Columns.MustGet(id)
		lenRefCol := smartvectors.Density(refColSV)

		if m != lenRefCol {
			utils.Panic(
				"column %v has different effective length from the reference column %v, length %v and %v",
				col.GetColID(), id, m, lenRefCol)
		}
	}

	// return the effective part of the column
	return a
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
