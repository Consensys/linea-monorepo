//go:build cuda

package quotient

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/gpu"
	gpuvortex "github.com/consensys/linea-monorepo/prover/gpu/vortex"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// BenchmarkKBNTT benchmarks single KoalaBear NTT (forward DIF).
func BenchmarkKBNTT(b *testing.B) {
	dev := gpu.GetDevice()
	if dev == nil {
		b.Skip("no GPU")
	}

	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			nttDom, _ := gpuvortex.NewGPUFFTDomain(dev, n)
			defer nttDom.Free()
			dVec, _ := gpuvortex.NewKBVector(dev, n)
			defer dVec.Free()

			h := make([]field.Element, n)
			for i := range h {
				h[i].SetUint64(uint64(i + 1))
			}
			dVec.CopyFromHost(h)
			gpuvortex.Sync(dev)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				nttDom.FFT(dVec)
			}
			gpuvortex.Sync(dev)
		})
	}
}

// BenchmarkKBCosetFFT benchmarks coset forward NTT + bit-reversal.
func BenchmarkKBCosetFFT(b *testing.B) {
	dev := gpu.GetDevice()
	if dev == nil {
		b.Skip("no GPU")
	}

	n := 1 << 18
	nttDom, _ := gpuvortex.NewGPUFFTDomain(dev, n)
	defer nttDom.Free()

	dCoeffs, _ := gpuvortex.NewKBVector(dev, n)
	defer dCoeffs.Free()
	dEval, _ := gpuvortex.NewKBVector(dev, n)
	defer dEval.Free()

	h := make([]field.Element, n)
	for i := range h {
		h[i].SetUint64(uint64(i + 1))
	}
	dCoeffs.CopyFromHost(h)

	shift := computeShift(uint64(n), 2, 0)
	gpuvortex.Sync(dev)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dEval.CopyFromDevice(dCoeffs)
		nttDom.CosetFFT(dEval, shift)
		dEval.BitReverse()
	}
	gpuvortex.Sync(dev)
}

// BenchmarkKBBatchNTT benchmarks batch NTT (all roots in one call).
func BenchmarkKBBatchNTT(b *testing.B) {
	dev := gpu.GetDevice()
	if dev == nil {
		b.Skip("no GPU")
	}

	n := 1 << 18
	batches := []int{10, 100, 500, 1500}
	nttDom, _ := gpuvortex.NewGPUFFTDomain(dev, n)
	defer nttDom.Free()

	shift := computeShift(uint64(n), 2, 0)

	for _, batch := range batches {
		b.Run(fmt.Sprintf("batch=%d", batch), func(b *testing.B) {
			dPacked, _ := gpuvortex.NewKBVector(dev, batch*n)
			defer dPacked.Free()

			h := make([]field.Element, batch*n)
			for i := range h {
				h[i].SetUint64(uint64(i%n + 1))
			}
			dPacked.CopyFromHost(h)
			gpuvortex.Sync(dev)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				nttDom.BatchCosetFFTBitRev(dPacked, batch, shift)
			}
			gpuvortex.Sync(dev)
		})
	}
}
