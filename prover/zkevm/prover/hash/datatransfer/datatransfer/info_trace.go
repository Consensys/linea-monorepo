package datatransfer

import (
	"github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf"
)

const (
	// number of slices in keccakf for base conversion
	numSlices = 16

	// number of lanes in the output of hash
	// digest is 256 bits that is 4 lanes of 64bits
	numLanesInHashOutPut = 4
)

type HashOutput struct {
	// hash slices in uint4
	HashLoSlices, HashHiSlices [numLanesInHashOutPut / 2][numSlices]ifaces.Column
	HashLo, HashHi             ifaces.Column
	MaxNumRows                 int
}

func (h *HashOutput) newHashOutput(comp *wizard.CompiledIOP, round, maxNumKeccakF int) {
	h.MaxNumRows = utils.NextPowerOfTwo(maxNumKeccakF)
	// declare the columns
	h.DeclareColumns(comp, round)

	// constraints over decomposition hashX to hashXSlices
	h.csDecompose(comp, round)
}
func (h *HashOutput) DeclareColumns(comp *wizard.CompiledIOP, round int) {
	h.HashLo = comp.InsertCommit(round, ifaces.ColIDf("Hash_Lo"), h.MaxNumRows)
	h.HashHi = comp.InsertCommit(round, ifaces.ColIDf("Hash_Hi"), h.MaxNumRows)

	for j := range h.HashLoSlices {
		for k := range h.HashLoSlices[0] {
			h.HashLoSlices[j][k] = comp.InsertCommit(round, ifaces.ColIDf("HashLo_Slices_%v_%v", j, k), h.MaxNumRows)
			h.HashHiSlices[j][k] = comp.InsertCommit(round, ifaces.ColIDf("HashHi_Slices_%v_%v", j, k), h.MaxNumRows)
		}
	}
}

// constraints over decomposition of HashHi and HashLow to slices in uint4.
func (h *HashOutput) csDecompose(comp *wizard.CompiledIOP, round int) {
	a := append(h.HashLoSlices[0][:], h.HashLoSlices[1][:]...)
	slicesLo := SlicesBeToLeHandle(a)

	exprLo := keccakf.BaseRecomposeHandles(slicesLo, 16)
	comp.InsertGlobal(round, ifaces.QueryIDf("Decomposition_HashLo"), symbolic.Sub(exprLo, h.HashLo))

	b := append(h.HashHiSlices[0][:], h.HashHiSlices[1][:]...)
	slicesHi := SlicesBeToLeHandle(b)

	exprHi := keccakf.BaseRecomposeHandles(slicesHi, 16)
	comp.InsertGlobal(round, ifaces.QueryIDf("Decomposition_HashHi"), symbolic.Sub(exprHi, h.HashHi))
}

// It assigns the columns specific to the module.
func (h *HashOutput) AssignHashOutPut(
	run *wizard.ProverRuntime,
	permTrace keccak.PermTraces,
) {

	// assign Info trace
	var v, w field.Element
	var hashLo, hashHi []field.Element
	for _, digest := range permTrace.HashOutPut {
		hi := digest[:maxNByte]
		lo := digest[maxNByte:]

		v.SetBytes(hi[:])
		w.SetBytes(lo[:])

		hashLo = append(hashLo, w)
		hashHi = append(hashHi, v)
	}

	run.AssignColumn(h.HashHi.GetColID(), smartvectors.RightZeroPadded(hashHi, h.MaxNumRows))
	run.AssignColumn(h.HashLo.GetColID(), smartvectors.RightZeroPadded(hashLo, h.MaxNumRows))

	// slices from PermTrace
	var hashHiSlices, hashLoSlices [numLanesInHashOutPut / 2][numSlices][]field.Element
	for i := range hashLo {
		base := 16      // slices of 4 bits, thus base is 2^4
		numchunck := 32 // hashHi is 128 bits and would be decomposed into 32 slices of 4 bits
		decHashLo := keccakf.DecomposeFr(hashLo[i], base, numchunck)
		sliceLo := SlicesBeToLeUint4(decHashLo)

		decHashHi := keccakf.DecomposeFr(hashHi[i], base, numchunck)
		sliceHi := SlicesBeToLeUint4(decHashHi)

		for k := range hashLoSlices[0] {
			// dec[:16] goes to hashSlices[0], and dec[16:] goes to hashSlices[1]
			hashLoSlices[0][k] = append(hashLoSlices[0][k], sliceLo[k])
			hashLoSlices[1][k] = append(hashLoSlices[1][k], sliceLo[k+numSlices])

			hashHiSlices[0][k] = append(hashHiSlices[0][k], sliceHi[k])
			hashHiSlices[1][k] = append(hashHiSlices[1][k], sliceHi[k+numSlices])
		}
	}
	for j := range h.HashLoSlices {
		for k := range h.HashLoSlices[0] {
			run.AssignColumn(h.HashHiSlices[j][k].GetColID(), smartvectors.RightZeroPadded(hashHiSlices[j][k], h.MaxNumRows))
			run.AssignColumn(h.HashLoSlices[j][k].GetColID(), smartvectors.RightZeroPadded(hashLoSlices[j][k], h.MaxNumRows))
		}
	}
}
