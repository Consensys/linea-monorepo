package iokeccakf

import (
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
	for x := 0; x < 3; x++ {
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
func (o *KeccakFOutputs) Assign(run *wizard.ProverRuntime) {
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
