package datatransfer

import (
	"github.com/consensys/zkevm-monorepo/prover/crypto/keccak"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/keccak/keccakf"
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
	HashHiSlices, HashLowSlices [numLanesInHashOutPut / 2][numSlices]ifaces.Column
	hashHi, hashLow             ifaces.Column
	maxNumRows                  int
}

func (h *HashOutput) newHashOutput(comp *wizard.CompiledIOP, round, maxNumKeccakF int, lu lookUpTables) {
	h.maxNumRows = utils.NextPowerOfTwo(maxNumKeccakF)
	// declare the columns
	h.hashHi = comp.InsertCommit(round, ifaces.ColIDf("Hash_Hi"), h.maxNumRows)
	h.hashLow = comp.InsertCommit(round, ifaces.ColIDf("Hash_Low"), h.maxNumRows)

	for j := range h.HashHiSlices {
		for k := range h.HashHiSlices[0] {
			h.HashHiSlices[j][k] = comp.InsertCommit(round, ifaces.ColIDf("HashHi_Slices_%v_%v", j, k), h.maxNumRows)
			h.HashLowSlices[j][k] = comp.InsertCommit(round, ifaces.ColIDf("HashLow_Slices_%v_%v", j, k), h.maxNumRows)
		}
	}

	// constraints over decomposition hashX to hashXSlices
	h.csDecompose(comp, round)
}

// constraints over decomposition of HashHi and HashLow to slices in uint4.
func (h *HashOutput) csDecompose(comp *wizard.CompiledIOP, round int) {
	slicesHi := append(h.HashHiSlices[0][:], h.HashHiSlices[1][:]...)
	exprHi := keccakf.BaseRecomposeHandles(slicesHi, 16)
	comp.InsertGlobal(round, ifaces.QueryIDf("Decomposition_HashHi"), symbolic.Sub(exprHi, h.hashHi))

	slicesLow := append(h.HashLowSlices[0][:], h.HashLowSlices[1][:]...)
	exprLow := keccakf.BaseRecomposeHandles(slicesLow, 16)
	comp.InsertGlobal(round, ifaces.QueryIDf("Decomposition_HashLow"), symbolic.Sub(exprLow, h.hashLow))
}

// It assigns the columns specific to the module.
func (h *HashOutput) AssignHashOutPut(
	run *wizard.ProverRuntime,
	permTrace keccak.PermTraces,
) {
	//hashLow,hashHi from PermTrace
	var hashLow, hashHi []field.Element
	var u, v field.Element

	power8Fr := field.NewElement(power8)
	for _, output := range permTrace.HashOutPut {
		b := output[:keccak.OutputLen/2]
		c := output[keccak.OutputLen/2:]
		var sliceU, sliceV []field.Element
		for i := range b {
			u.SetBytes([]byte{b[i]})
			sliceU = append(sliceU, u)

			v.SetBytes([]byte{c[i]})
			sliceV = append(sliceV, v)
		}

		hashLow = append(hashLow, keccakf.BaseRecompose(sliceU, &power8Fr))
		hashHi = append(hashHi, keccakf.BaseRecompose(sliceV, &power8Fr))
	}
	run.AssignColumn(h.hashLow.GetColID(), smartvectors.RightZeroPadded(hashLow, h.maxNumRows))
	run.AssignColumn(h.hashHi.GetColID(), smartvectors.RightZeroPadded(hashHi, h.maxNumRows))

	// slices from PermTrace
	var hashLowSlices, hashHiSlices [numLanesInHashOutPut / 2][numSlices][]field.Element
	for i := range hashHi {
		base := 16      // slices of 4 bits, thus base is 2^4
		numchunck := 32 // hashHi is 128 bits and would be decomposed into 32 slices of 4 bits
		decHashHi := keccakf.DecomposeFr(hashHi[i], base, numchunck)
		decHashLow := keccakf.DecomposeFr(hashLow[i], base, numchunck)
		for k := range hashHiSlices[0] {
			// dec[:16] goes to hashSlices[0], and dec[16:] goes to hashSlices[1]
			hashHiSlices[0][k] = append(hashHiSlices[0][k], decHashHi[k])
			hashHiSlices[1][k] = append(hashHiSlices[1][k], decHashHi[k+numSlices])

			hashLowSlices[0][k] = append(hashLowSlices[0][k], decHashLow[k])
			hashLowSlices[1][k] = append(hashLowSlices[1][k], decHashLow[k+numSlices])
		}
	}
	for j := range h.HashHiSlices {
		for k := range h.HashHiSlices[0] {
			run.AssignColumn(h.HashLowSlices[j][k].GetColID(), smartvectors.RightZeroPadded(hashLowSlices[j][k], h.maxNumRows))
			run.AssignColumn(h.HashHiSlices[j][k].GetColID(), smartvectors.RightZeroPadded(hashHiSlices[j][k], h.maxNumRows))
		}
	}
}
