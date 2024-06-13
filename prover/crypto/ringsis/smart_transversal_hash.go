package ringsis

import (
	"runtime"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/sis"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/vector"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/utils/parallel"
)

const (
	numFieldPerPoly int = 2
	degree          int = 64
)

// smartTransveralHash_SIS_8_64 is an optimized version of [TransversalHash]
// dedicated to the parameters modulusDegree=64 and log2_bound=8. Its
// implementation is parallel and is tailored for massive multicore environment.
func smartTransversalHash_SIS_8_64(
	// the Ag for ring-sis
	ag [][]field.Element,
	// A non-transposed list of columns
	// All of the same length
	pols []smartvectors.SmartVector,
	// The precomputed twiddle cosets for the forward FFT
	twiddleCosets []field.Element,
	// The domain for the final inverse-FFT
	domain *fft.Domain,
) []field.Element {

	// Each field element is encoded in 32 limbs but the degree is 64. So, each
	// polynomial multiplication "hashes" 2 field elements at once. This is
	// important to know for parallelization.
	resultSize := pols[0].Len() * degree

	// To optimize memory usage, we limit ourself to hash only 16 columns per
	// iteration.
	numColumnPerJob := 16
	numWorker := runtime.NumCPU()
	// In theory, it should be a div ceil. But in practice we only process power's
	// of two number of columns. If that's not the case, then the function will panic
	// but we can always change that if this is needed. The rational for the current
	// design is simplicity.
	numJobs := utils.DivExact(pols[0].Len(), numColumnPerJob) // we make blocks of 16 columns

	// Main result of the hashing
	mainResults := make([]field.Element, resultSize)
	// When we encounter a const row, it will have the same additive contribution
	// to the result on every column. So we compute the contribution only once and
	// accumulate it with the other "constant column contributions". And it is only
	// performed by the first thread.
	constResults := make([]field.Element, degree)

	parallel.ExecuteChunky(numJobs, func(start, stop int) {

		// We process the columns per segment of `numColumnPerJob`
		for i := start; i < stop; i++ {

			localResult := make([]field.Element, numColumnPerJob*degree)
			limbs := make([]field.Element, degree)

			// Each segment is processed by packet of `numFieldPerPoly` rows
			startFromCol := start * numColumnPerJob
			stopAtCol := stop * numColumnPerJob

			for row := 0; row < len(pols); row += numFieldPerPoly {

				aIFace := pols[row]
				// This addresses the case where the last polynomial is incomplete
				var bIFace smartvectors.SmartVector = smartvectors.NewConstant(field.Zero(), pols[0].Len())
				if row+1 < len(pols) {
					bIFace = pols[row+1]
				}

				// Try to cast them into either reg or constant
				aReg, aIsReg := aIFace.(*smartvectors.Regular)
				bReg, bIsReg := bIFace.(*smartvectors.Regular)
				aCon, aIsCon := aIFace.(*smartvectors.Constant)
				bCon, bIsCon := bIFace.(*smartvectors.Constant)

				switch {
				case (!aIsReg && !aIsCon) || (!bIsReg && !bIsCon):
					utils.Panic("Forbidden types : a = %T and b = %T", aIFace, bIFace)
				case aIsCon && bIsCon && i == 0:
					// Only the first job cares about constant polynomials
					accumulatePartialSisHashConCon(
						constResults,
						ag[row/numFieldPerPoly],
						twiddleCosets,
						aCon,
						bCon,
					)
				case aIsCon && bIsReg:
					accumulatePartialSisHashConReg(
						localResult,
						ag[row/numFieldPerPoly],
						twiddleCosets,
						limbs,
						aCon,
						(*bReg)[startFromCol:stopAtCol],
					)
				case aIsReg && bIsReg:
					accumulatePartialSisHashRegReg(
						localResult,
						ag[row/numFieldPerPoly],
						twiddleCosets,
						limbs,
						(*aReg)[startFromCol:stopAtCol],
						(*bReg)[startFromCol:stopAtCol],
					)
				case aIsReg && bIsCon:
					accumulatePartialSisHashRegCon(
						localResult,
						ag[row/numFieldPerPoly],
						twiddleCosets,
						limbs,
						(*aReg)[startFromCol:stopAtCol],
						bCon,
					)
				}

			}

			// copy the segment into the main result at the end
			copy(mainResults[startFromCol*degree:stopAtCol*degree], localResult)

		}

	}, numWorker)

	// Now, we need to reconciliate the results of the buffer with
	// the result for each thread
	parallel.Execute(pols[0].Len(), func(start, stop int) {
		for col := start; col < stop; col++ {
			// Accumulate the const
			vector.Add(mainResults[col*degree:(col+1)*degree], mainResults[col*degree:(col+1)*degree], constResults)
			// And run the reverse FFT
			domain.FFTInverse(mainResults[col*degree:(col+1)*degree], fft.DIT, fft.OnCoset(), fft.WithNbTasks(1))
		}
	})

	return mainResults
}

// accumulatePartialSisHashRegCon partially evaluates the ring-SIS hash for only
// a single polynomial multiplication. a and b are two [smartvectors.Smartvector]
// representing two consecutive rows and a range of consecutive columns of the
// matrix to hash. The function each column of this submatrix is interpreted
// as a "ring-SIS" polynomial and is multiplied with the polynomial represented
// by ag. The product is then added to `result`.
//
// `limb` is a slice of field element of size 64 which is used to avoid
// periodic memory reallocation.
//
// The function is tailored specifically to the case where both a and b are
// constants in the sub-matrix and minimizes the CPU and memory usage in this
// situation.
func accumulatePartialSisHashConCon(result, ag, twiddleCosets []field.Element, a, b *smartvectors.Constant) {
	// Compute the partial hash only once and accumulate it in the
	// partial const result
	limbs := make([]field.Element, 64)
	limbDecompose64(limbs, a.Val(), b.Val())
	sis.FFT64(limbs, twiddleCosets)
	mulModAcc(result, limbs, ag)
}

// accumulatePartialSisHashRegCon partially evaluates the ring-SIS hash for only
// a single polynomial multiplication. a and b are two [smartvectors.Smartvector]
// representing two consecutive rows and a range of consecutive columns of the
// matrix to hash. The function each column of this submatrix is interpreted
// as a "ring-SIS" polynomial and is multiplied with the polynomial represented
// by ag. The product is then added to `result`.
//
// `limb` is a slice of field element of size 64 which is used to avoid
// periodic memory reallocation.
//
// The function is tailored specifically to the case where neither a nor b are
// constants in the sub-matrix.
func accumulatePartialSisHashRegReg(result, ag, twiddleCosets, limbs []field.Element, a, b []field.Element) {
	for col := 0; col < len(a); col++ {
		// Compute the partial hash only once and accumulate it in the
		// partial const result
		limbDecompose64(limbs, a[col], b[col])
		sis.FFT64(limbs, twiddleCosets)
		mulModAcc(result[col*degree:(col+1)*degree], limbs, ag)
	}
}

// accumulatePartialSisHashRegCon partially evaluates the ring-SIS hash for only
// a single polynomial multiplication. a and b are two [smartvectors.Smartvector]
// representing two consecutive rows and a range of consecutive columns of the
// matrix to hash. The function each column of this submatrix is interpreted
// as a "ring-SIS" polynomial and is multiplied with the polynomial represented
// by ag. The product is then added to `result`.
//
// `limb` is a slice of field element of size 64 which is used to avoid
// periodic memory reallocation.
//
// The function is tailored specifically to the case where "b" is constant to
// minimize the memory bandwidth usage in this situation.
func accumulatePartialSisHashRegCon(result, ag, twiddleCosets, limbs []field.Element, a []field.Element, b *smartvectors.Constant) {
	for col := 0; col < len(a); col++ {
		// Compute the partial hash only once and accumulate it in the
		// partial const result
		limbDecompose64(limbs, a[col], b.Val())
		sis.FFT64(limbs, twiddleCosets)
		mulModAcc(result[col*degree:(col+1)*degree], limbs, ag)
	}
}

// accumulatePartialSisHashConReg partially evaluates the ring-SIS hash for only
// a single polynomial multiplication. a and b are two [smartvectors.Smartvector]
// representing two consecutive rows and a range of consecutive columns of the
// matrix to hash. The function each column of this submatrix is interpreted
// as a "ring-SIS" polynomial and is multiplied with the polynomial represented
// by ag. The product is then added to `result`.
//
// `limb` is a slice of field element of size 64 which is used to avoid
// periodic memory reallocation.
//
// The function is tailored specifically to the case where "a" is constant to
// optimize the memory bandwidth usage in this situation.
func accumulatePartialSisHashConReg(result, ag, twiddleCosets, limbs []field.Element, a *smartvectors.Constant, b []field.Element) {
	for col := 0; col < len(b); col++ {
		// Compute the partial hash only once and accumulate it in the
		// partial const result
		limbDecompose64(limbs, a.Val(), b[col])
		sis.FFT64(limbs, twiddleCosets)
		mulModAcc(result[col*degree:(col+1)*degree], limbs, ag)
	}
}

var _64Zeroes []field.Element = make([]field.Element, 64)

// zeroize64 fills `buf` with zeroes.
func zeroize64(buf []field.Element) {
	copy(buf, _64Zeroes)
}

// mulModAdd increments each entry `i` of `res` as `res[i] = a[i] * b[i]`. The
// input vectors are trusted to all have the same length.
func mulModAcc(res, a, b []field.Element) {
	var tmp field.Element
	for i := range res {
		tmp.Mul(&a[i], &b[i])
		res[i].Add(&res[i], &tmp)
	}
}

// limbDecompose64 decomposes sequentially a and b in limbs of 8 bits and assigns
// the entries of `result` with the obtained limbs in little-endian order,
// pushing the limbs of a and then the limbs of b. The returned field element
// are not in Montgommery form and this is addressed at the end of the hashing.
func limbDecompose64(result []field.Element, a, b field.Element) {

	zeroize64(result)

	bytesBuffer := a.Bytes()

	// unrolled loop (processing a)
	for i := 0; i < 4; i++ {
		// Since the results are zeroized and that the limbs
		// are small enough to hold over 64bits. We only write
		// in the lowest entry of the result.
		result[31-(8*i+0)][0] = uint64(bytesBuffer[8*i+0])
		result[31-(8*i+1)][0] = uint64(bytesBuffer[8*i+1])
		result[31-(8*i+2)][0] = uint64(bytesBuffer[8*i+2])
		result[31-(8*i+3)][0] = uint64(bytesBuffer[8*i+3])
		result[31-(8*i+4)][0] = uint64(bytesBuffer[8*i+4])
		result[31-(8*i+5)][0] = uint64(bytesBuffer[8*i+5])
		result[31-(8*i+6)][0] = uint64(bytesBuffer[8*i+6])
		result[31-(8*i+7)][0] = uint64(bytesBuffer[8*i+7])
	}

	bytesBuffer = b.Bytes()

	// unrolled loop (processing b)
	for i := 0; i < 4; i++ {
		result[63-(8*i+0)][0] = uint64(bytesBuffer[8*i+0])
		result[63-(8*i+1)][0] = uint64(bytesBuffer[8*i+1])
		result[63-(8*i+2)][0] = uint64(bytesBuffer[8*i+2])
		result[63-(8*i+3)][0] = uint64(bytesBuffer[8*i+3])
		result[63-(8*i+4)][0] = uint64(bytesBuffer[8*i+4])
		result[63-(8*i+5)][0] = uint64(bytesBuffer[8*i+5])
		result[63-(8*i+6)][0] = uint64(bytesBuffer[8*i+6])
		result[63-(8*i+7)][0] = uint64(bytesBuffer[8*i+7])
	}
}
