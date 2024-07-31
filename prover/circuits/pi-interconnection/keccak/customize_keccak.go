package keccak

import (
	"bytes"
	"encoding/binary"

	"github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"

	wKeccak "github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf"
)

type module wKeccak.Module

// DefineCustomizedKeccak declares the columns and the constraints for a customized case,
//
//	where the lanes are provided directly via SliceProviders.
func (m *module) DefineCustomizedKeccak(comp *wizard.CompiledIOP, maxNbKeccakF int) {
	const round = 0
	m.MaxNumKeccakF = maxNbKeccakF

	m.DataTransfer.NewCustomizedDataTransfer(comp, round, maxNbKeccakF)
	m.Keccakf = keccakf.NewModule(comp, round, maxNbKeccakF)

	// assign the blocks from DataTransfer to keccakF,
	// also take the output from keccakF and give it back to DataTransfer
	(*wKeccak.Module)(m).CsConnectDataTransferToKeccakF(comp, round)
}

// AssignCustomizedKeccak customizes keccak module for the case where lanes are directly provided via SliceProviders.
func (m *module) AssignCustomizedKeccak(run *wizard.ProverRuntime) {

	// Construct the traces
	permTrace := keccak.PermTraces{}
	buildPermTrace(m.SliceProviders, &permTrace)

	// And manually assign the module from the content of permTrace.
	m.DataTransfer.AssignCustomizedDataTransfer(run, permTrace)
	m.Keccakf.Assign(run, permTrace)

	//runtime.GC() TODO benchmark and see if this would help

}

// it builds the permutation trace from the input
func buildPermTrace(in [][]byte, permTrace *keccak.PermTraces) {

	stream := bytes.Buffer{}
	for i := range in {
		stream.Write(in[i])
		keccak.Hash(stream.Bytes(), permTrace)
		stream.Reset()
	}
}

// AssignColumns to be used in the interconnection circuit assign function
func AssignColumns(in [][]byte, nbLanes int) (lanes []uint64, isLaneActive, isFirstLaneOfHash []int, hashHi, hashLo []field.Element) {
	hashHi = make([]field.Element, len(in))
	hashLo = make([]field.Element, len(in))
	lanes = make([]uint64, nbLanes)
	isFirstLaneOfHash = make([]int, nbLanes)
	isLaneActive = make([]int, nbLanes)

	laneI := 0
	for i := range in {
		isFirstLaneOfHash[laneI] = 1
		// pad and turn into lanes
		nbBlocks := 1 + len(in[i])/136
		for j := 0; j < nbBlocks; j++ {
			var block [136]byte
			copy(block[:], in[i][j*136:])
			if j == nbBlocks-1 {
				block[len(in[i])-j*136] = 1 // dst
				block[135] |= 0x80          // end marker
			}
			for k := 0; k < 17; k++ {
				isLaneActive[laneI] = 1
				lanes[laneI] = binary.LittleEndian.Uint64(block[k*8 : k*8+8])
				laneI++
			}
		}
		hash := utils.KeccakHash(in[i])
		hashHi[i].SetBytes(hash[:16])
		hashLo[i].SetBytes(hash[16:])

	}

	return
}
