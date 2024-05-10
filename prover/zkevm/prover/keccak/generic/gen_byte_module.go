package generic

import (
	"bytes"

	"github.com/consensys/zkevm-monorepo/prover/crypto/keccak"
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
	// (optional) indicates whether the current limb is finished but either
	// both must be allocated either none of them.
	LC, LX  ifaces.Column
	TO_HASH ifaces.Column
}

// GenInfoModule collects the columns summarizing information about the hash as
// as a whole.
type GenInfoModule struct {
	// Identifier for the hash. Allows joining with the data module
	HashNum ifaces.Column
	HashHi  ifaces.Column
	HashLo  ifaces.Column
}

// the assignment to GenericByteModule
type GenTrace struct {
	HashNum, Index, Limb, NByte []field.Element
	LC, LX, TO_HASH             []field.Element
	//used for the assignment to the leftovers
	CleanLimb []field.Element
}

// Creates a new GenericByteModule from the given definition
func NewGenericByteModule(
	comp *wizard.CompiledIOP,
	definition GenericByteModuleDefinition,
) *GenericByteModule {

	// Load the mandatory fields
	res := &GenericByteModule{
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
			HashNum: comp.Columns.GetHandle(definition.Info.HashNum),
			HashHi:  comp.Columns.GetHandle(definition.Info.HashHi),
			HashLo:  comp.Columns.GetHandle(definition.Info.HashLo),
		}
	}

	// Optionally load the LC and LX columns if they are loaded
	if len(definition.Data.LC) > 0 {
		res.Data.LC = comp.Columns.GetHandle(definition.Data.LC)
		res.Data.LX = comp.Columns.GetHandle(definition.Data.LX)
	}
	if len(definition.Data.TO_HASH) > 0 {
		res.Data.TO_HASH = comp.Columns.GetHandle(definition.Data.TO_HASH)
	}

	return res
}

// Implements the trace providing mechanism for the generic byte module.
func (gen *GenericByteModule) AppendTraces(
	run *wizard.ProverRuntime,
	traces *keccak.PermTraces,
	genTrace *GenTrace,
) {

	data := gen.Data

	// Extract the assignments through a shallow copy.
	hashNum := gen.extractCol(run, gen.Data.HashNum)
	index := gen.extractCol(run, gen.Data.Index)
	limbs := gen.extractCol(run, gen.Data.Limb)
	nBytes := gen.extractCol(run, gen.Data.NBytes)

	// Optionally extract the LC and LX columns
	var lc, lx, toHash []field.Element
	if data.LC != nil {
		lc = gen.extractCol(run, gen.Data.LC)
		lx = gen.extractCol(run, gen.Data.LX)
		genTrace.LC = lc
		genTrace.LX = lx
	}
	if data.TO_HASH != nil {
		toHash = gen.extractCol(run, data.TO_HASH)
		genTrace.TO_HASH = toHash
	}

	stream := bytes.Buffer{}
	limbSerialized := [32]byte{}
	one := field.One()
	cleanLimb := make([]field.Element, len(hashNum))
	for pos := range hashNum {
		// Check if the current limb can be appended
		if lc != nil && (lc[pos] != one || lx[pos] != one) {
			continue
		}
		if toHash != nil && (toHash[pos] != one) {
			continue
		}

		// Sanity-check, if the index is zero must be equivalent to an empty
		// stream
		if index[pos].IsZero() != (stream.Len() == 0) {
			utils.Panic("the index==0 should mean an empty stream, failed at position %v", pos)
		}

		// Extract the limb, which is left aligned to the 16-th byte
		limbSerialized = limbs[pos].Bytes()
		byteSize := nBytes[pos].Uint64()
		res := limbSerialized[GEN_LEFT_ALIGNMENT : GEN_LEFT_ALIGNMENT+byteSize]
		cleanLimb[pos].SetBytes(res)
		stream.Write(
			res,
		)

		// If we are on the last limb or if the hashNum increases, it means
		// that we need to close the hash to start the next one
		if pos+1 == len(hashNum) || hashNum[pos+1].Uint64() > hashNum[pos].Uint64() {
			keccak.Hash(stream.Bytes(), traces)
			stream.Reset()
		}
	}
	genTrace.HashNum = hashNum
	genTrace.Index = index
	genTrace.Limb = limbs
	genTrace.NByte = nBytes
	genTrace.CleanLimb = cleanLimb

}

// Extract a shallow copy of the active zone of a column. Meaning the unpadded
// area where the column encodes actual data.
func (gen *GenericByteModule) extractCol(
	run *wizard.ProverRuntime,
	col ifaces.Column,
) []field.Element {

	// Fetchs the smart-vector and delimitate the active zone. Here we assume
	// that all the columns are zero-prepended. And have the same length. That's
	// why stop - density gives us the starting position for scanning the
	// witness.
	var (
		col_    = run.Columns.MustGet(col.GetColID())
		density = smartvectors.Density(col_)
		stop    = col_.Len()
		start   = stop - smartvectors.Density(col_)
	)

	// Calling subvector would result in an exception. Thus, we treat it as a
	// special case and return an empty vector.
	if density == 0 {
		return []field.Element{}
	}

	// Extract the assignments through a shallow copy.
	return col_.SubVector(start, stop).IntoRegVecSaveAlloc()
}
