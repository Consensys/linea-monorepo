package fft

import (
	"math/bits"
	"runtime"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/parallel"
)

var (
	// Tracks the number of available CPUs
	numCPU = runtime.GOMAXPROCS(0)
)

// FFT computes (recursively) the discrete Fourier transform of a and stores the result in a
// if decimation == DIT (decimation in time), the input must be in bit-reversed order
// if decimation == DIF (decimation in frequency), the output will be in bit-reversed order
// if coset if set, the FFT(a) returns the evaluation of a on a coset.
func (domain *Domain) BatchFFT(a [][]field.Element, decimation Decimation, coset ...bool) {

	// Check the size of `a` to the be equals to the cardinality
	if len(a) != int(domain.Cardinality) {
		utils.Panic("Expected matrix with %v col, but got %v cols", domain.Cardinality, len(a))
	}

	_coset := false
	if len(coset) > 0 {
		_coset = coset[0]
	}

	// if coset != 0, scale by coset table
	if _coset {
		scale := func(cosetTable []field.Element) {
			parallel.Execute(len(a), func(start, stop int) {
				for i := start; i < stop; i++ {
					vector.ScalarMul(a[i], a[i], cosetTable[i])
				}
			})
		}
		if decimation == DIT {
			scale(domain.CosetTableReversed)

		} else {
			scale(domain.CosetTable)
		}
	}

	// find the stage where we should stop spawning go routines in our recursive calls
	// (ie when we have as many go routines running as we have available CPUs)
	maxSplits := utils.Log2Ceil(numCPU)
	if numCPU <= 1 {
		maxSplits = -1
	}

	switch decimation {
	case DIF:
		batchDifFFT(a, domain.Twiddles, 0, maxSplits, nil)
	case DIT:
		batchDitFFT(a, domain.Twiddles, 0, maxSplits, nil)
	default:
		panic("not implemented")
	}
}

// FFTInverse computes (recursively) the inverse discrete Fourier transform of a and stores the result in a
// if decimation == DIT (decimation in time), the input must be in bit-reversed order
// if decimation == DIF (decimation in frequency), the output will be in bit-reversed order
// coset sets the shift of the fft (0 = no shift, standard fft)
// len(a) must be a power of 2, and w must be a len(a)th root of unity in field F.
func (domain *Domain) BatchFFTInverse(a [][]field.Element, decimation Decimation, coset ...bool) {

	// Check the size of `a` to the be equals to the cardinality
	if len(a) != int(domain.Cardinality) {
		utils.Panic("Expected matrix with %v col, but got %v cols", domain.Cardinality, len(a))
	}

	_coset := false
	if len(coset) > 0 {
		_coset = coset[0]
	}

	// find the stage where we should stop spawning go routines in our recursive calls
	// (ie when we have as many go routines running as we have available CPUs)
	maxSplits := utils.Log2Ceil(numCPU)
	if numCPU <= 1 {
		maxSplits = -1
	}

	switch decimation {
	case DIF:
		batchDifFFT(a, domain.TwiddlesInv, 0, maxSplits, nil)
	case DIT:
		batchDitFFT(a, domain.TwiddlesInv, 0, maxSplits, nil)
	default:
		panic("not implemented")
	}

	// scale by CardinalityInv
	if !_coset {
		parallel.Execute(len(a), func(start, end int) {
			for i := start; i < end; i++ {
				vector.ScalarMul(a[i], a[i], domain.CardinalityInv)
			}
		}, numCPU)
		return
	}

	scale := func(cosetTable []field.Element) {
		parallel.Execute(len(a), func(start, end int) {
			for i := start; i < end; i++ {
				vector.ScalarMul(a[i], a[i], cosetTable[i])
				vector.ScalarMul(a[i], a[i], domain.CardinalityInv)
			}
		}, numCPU)
	}

	if decimation == DIT {
		scale(domain.CosetTableInv)
		return
	}

	// decimation == DIF
	scale(domain.CosetTableInvReversed)

}

func batchDifFFT(a [][]field.Element, twiddles [][]field.Element, stage int, maxSplits int, chDone chan struct{}) {

	if chDone != nil {
		defer close(chDone)
	}

	n := len(a)
	if n == 1 {
		return
	}

	m := n >> 1

	// if stage < maxSplits, we parallelize this butterfly
	// but we have only numCPU / stage cpus available
	if (m > ButterflyThreshold) && (stage < maxSplits) {
		// 1 << stage == estimated used CPUs
		localNumCpu := numCPU / (1 << (stage))
		parallel.Execute(m, func(start, end int) {
			for i := start; i < end; i++ {
				vector.Butterfly(a[i], a[i+m])
				vector.ScalarMul(a[i+m], a[i+m], twiddles[stage][i])
			}
		}, localNumCpu)
	} else {
		// otherwise, no parallel
		// i == 0
		vector.Butterfly(a[0], a[m])
		for i := 1; i < m; i++ {
			vector.Butterfly(a[i], a[i+m])
			vector.ScalarMul(a[i+m], a[i+m], twiddles[stage][i])
		}
	}

	if m == 1 {
		return
	}

	nextStage := stage + 1
	if stage < maxSplits {
		chDone := make(chan struct{}, 1)
		go batchDifFFT(a[m:n], twiddles, nextStage, maxSplits, chDone)
		batchDifFFT(a[0:m], twiddles, nextStage, maxSplits, nil)
		<-chDone
	} else {
		batchDifFFT(a[0:m], twiddles, nextStage, maxSplits, nil)
		batchDifFFT(a[m:n], twiddles, nextStage, maxSplits, nil)
	}
}

func batchDitFFT(a [][]field.Element, twiddles [][]field.Element, stage int, maxSplits int, chDone chan struct{}) {

	if chDone != nil {
		defer close(chDone)
	}

	n := len(a)
	if n == 1 {
		return
	}

	m := n >> 1

	nextStage := stage + 1

	if stage < maxSplits {
		// that's the only time we fire go routines
		chDone := make(chan struct{}, 1)
		go batchDitFFT(a[m:], twiddles, nextStage, maxSplits, chDone)
		batchDitFFT(a[0:m], twiddles, nextStage, maxSplits, nil)
		<-chDone
	} else {
		batchDitFFT(a[0:m], twiddles, nextStage, maxSplits, nil)
		batchDitFFT(a[m:n], twiddles, nextStage, maxSplits, nil)

	}

	// if stage < maxSplits, we parallelize this butterfly
	// but we have only numCPU / stage cpus available
	if (m > ButterflyThreshold) && (stage < maxSplits) {
		// 1 << stage == estimated used CPUs
		numCPU := numCPU / (1 << (stage))
		parallel.Execute(m, func(start, end int) {
			for k := start; k < end; k++ {
				vector.ScalarMul(a[k+m], a[k+m], twiddles[stage][k])
				vector.Butterfly(a[k], a[k+m])
			}
		}, numCPU)

	} else {
		vector.Butterfly(a[0], a[m])
		for k := 1; k < m; k++ {
			vector.ScalarMul(a[k+m], a[k+m], twiddles[stage][k])
			vector.Butterfly(a[k], a[k+m])
		}
	}
}

// BitReverse applies the bit-reversal permutation to a.
// len(a) must be a power of 2 (as in every single function in this file)
func BatchBitReverse(a [][]field.Element) {
	n := uint64(len(a))
	nn := uint64(64 - bits.TrailingZeros64(n))

	for i := uint64(0); i < n; i++ {
		irev := bits.Reverse64(i) >> nn
		if irev > i {
			a[i], a[irev] = a[irev], a[i]
		}
	}
}
