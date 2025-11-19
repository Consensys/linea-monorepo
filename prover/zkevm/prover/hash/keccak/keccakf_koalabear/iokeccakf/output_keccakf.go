package iokeccakf

import (
	"encoding/binary"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	kcommon "github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf_koalabear/common"
)

const HashSizeBytes = 32 // keccak256

var (
	NbChunksHash = HashSizeBytes // @TODO : change to chunks of two bytes.
)

type KeccakFOutputs struct {
	Hash    []ifaces.Column // hash result represented in chunks of  bytes
	IsHash  ifaces.Column
	hashNum ifaces.Column
}

// it first applies to-basex to get laneX, then a projection query to map lanex to blocks
func NewOutputKeccakF(comp *wizard.CompiledIOP,
	stateCurr [5][5][8]ifaces.Column,
	isBase2 ifaces.Column,
) *KeccakFOutputs {

	var (
		output = &KeccakFOutputs{
			Hash:    make([]ifaces.Column, NbChunksHash),
			hashNum: comp.InsertCommit(0, ifaces.ColIDf("KECCAKF_HASH_NUM"), isBase2.Size(), true),
		}
	)

	// extract the hash result from the state
	j := 0
	for x := 0; x < 4; x++ {
		for z := 0; z < kcommon.NumSlices; z++ {
			output.Hash[j] = stateCurr[x][0][z]
			j++
		}
	}

	output.IsHash = isBase2

	comp.InsertGlobal(0,
		ifaces.QueryIDf("KECCAKF_HASHNUM"),
		sym.Sub(output.hashNum,
			sym.Add(
				column.Shift(output.hashNum, -1),
				output.IsHash,
			),
		),
	)

	comp.InsertLocal(0,
		ifaces.QueryIDf("KECCAKF_HASHNUM_FIRST_ROW"),
		sym.Sub(output.hashNum, output.IsHash),
	)

	return output

}

// assignOutput assigns the values to the columns of output step.
func (o *KeccakFOutputs) Run(run *wizard.ProverRuntime) {
	// assign hash num column
	var (
		hashNum = *common.NewVectorBuilder(o.hashNum)
		isHash  = run.GetColumn(o.IsHash.GetColID()).IntoRegVecSaveAlloc()
	)
	for i := 0; i < o.IsHash.Size(); i++ {
		if isHash[i].IsOne() {
			hashNum.PushInc()
		} else if i != 0 {
			hashNum.PushIncBy(0)
		} else {
			hashNum.PushZero()
		}
	}
	hashNum.PadAndAssign(run)
}

// helper function to extract the hash Digests, ignoring the rows where isHash = 0
func (o *KeccakFOutputs) ExtractHashResult(run *wizard.ProverRuntime) (hashes []keccak.Digest) {

	var (
		hashCol = make([][]field.Element, len(o.Hash))
		size    = o.Hash[0].Size()
		isHash  = o.IsHash.GetColAssignment(run).IntoRegVecSaveAlloc()
	)

	for i := range o.Hash {
		hashCol[i] = make([]field.Element, size)
		hashCol[i] = o.Hash[i].GetColAssignment(run).IntoRegVecSaveAlloc()
	}

	// we just look at row%23 = 0 since the hash results are possibaly stored there.
	k := 0
	for row := 0; row < size; row = k*kcommon.NumRounds - 1 {

		if isHash[row].IsOne() {
			var a keccak.Digest
			for j := range hashCol {
				a[j] = LEBytes(&hashCol[j][row])[0] // take the least significant byte
			}
			hashes = append(hashes, a)
		}
		k++
	}

	return hashes
}

// helper func to convert 32 columns of one byte to  16 columns of 2 bytes
func TwoByTwoCombination() {

}

func LittleEndianBytes(f field.Element) (res [field.Bytes]byte) {
	f.Uint64()
	binary.LittleEndian.PutUint64(res[:], f.Uint64())
	return
}

// Bytes returns the value of z as a little-endian byte array
func LEBytes(z *field.Element) (res [field.Bytes]byte) {
	koalabear.LittleEndian.PutElement(&res, *z)
	return
}
