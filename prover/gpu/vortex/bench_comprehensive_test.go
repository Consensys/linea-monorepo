//go:build cuda

package vortex

import (
	"fmt"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/field/koalabear/sis"
	"github.com/consensys/linea-monorepo/prover/gpu"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────────────────────────────────────
// Timing helpers
// ─────────────────────────────────────────────────────────────────────────────

func fmtSize(n int) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%dM", n>>20)
	case n >= 1<<10:
		return fmt.Sprintf("%dK", n>>10)
	default:
		return fmt.Sprintf("%d", n)
	}
}

func reportTiming(b *testing.B, h2d, compute, d2h time.Duration, bytes int64) {
	total := h2d + compute + d2h
	b.ReportMetric(float64(h2d.Microseconds()), "h2d_µs")
	b.ReportMetric(float64(compute.Microseconds()), "compute_µs")
	b.ReportMetric(float64(d2h.Microseconds()), "d2h_µs")
	b.ReportMetric(float64(total.Microseconds()), "total_µs")
	if bytes > 0 && total > 0 {
		gbps := float64(bytes) / total.Seconds() / 1e9
		b.ReportMetric(gbps, "GB/s")
	}
}

func newBenchDevice(b *testing.B) *gpu.Device {
	b.Helper()
	dev, err := gpu.New()
	require.NoError(b, err)
	b.Cleanup(func() { dev.Close() })
	return dev
}

func randSlice(rng *rand.Rand, n int) []koalabear.Element {
	s := make([]koalabear.Element, n)
	for i := range s {
		s[i] = randKB(rng)
	}
	return s
}

// ═════════════════════════════════════════════════════════════════════════════
// 1. KBVector: H2D, D2H, D2D transfers
// ═════════════════════════════════════════════════════════════════════════════

func BenchmarkKBVecH2D(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{1}))

	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20, 1 << 22, 1 << 24}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSize(n)), func(b *testing.B) {
			vec, err := NewKBVector(dev, n)
			require.NoError(b, err)
			defer vec.Free()
			data := randSlice(rng, n)
			Sync(dev)

			b.SetBytes(int64(n) * 4)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				vec.CopyFromHost(data)
			}
			Sync(dev)
		})
	}
}

func BenchmarkKBVecH2DPinned(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{2}))

	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20, 1 << 22, 1 << 24}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSize(n)), func(b *testing.B) {
			vec, err := NewKBVector(dev, n)
			require.NoError(b, err)
			defer vec.Free()
			pinned := AllocPinned(n)
			defer FreePinned(pinned)
			for i := range pinned {
				pinned[i] = randKB(rng)
			}
			Sync(dev)

			b.SetBytes(int64(n) * 4)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				vec.CopyFromHostPinned(pinned)
			}
			Sync(dev)
		})
	}
}

func BenchmarkKBVecD2H(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{3}))

	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20, 1 << 22, 1 << 24}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSize(n)), func(b *testing.B) {
			vec, err := NewKBVector(dev, n)
			require.NoError(b, err)
			defer vec.Free()
			data := randSlice(rng, n)
			vec.CopyFromHost(data)
			dst := make([]koalabear.Element, n)
			Sync(dev)

			b.SetBytes(int64(n) * 4)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				vec.CopyToHost(dst)
			}
			Sync(dev)
		})
	}
}

func BenchmarkKBVecD2D(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{4}))

	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20, 1 << 22, 1 << 24}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSize(n)), func(b *testing.B) {
			src, err := NewKBVector(dev, n)
			require.NoError(b, err)
			defer src.Free()
			dst, err := NewKBVector(dev, n)
			require.NoError(b, err)
			defer dst.Free()
			data := randSlice(rng, n)
			src.CopyFromHost(data)
			Sync(dev)

			b.SetBytes(int64(n) * 4)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dst.CopyFromDevice(src)
			}
			Sync(dev)
		})
	}
}

// ═════════════════════════════════════════════════════════════════════════════
// 2. KBVector: arithmetic operations
// ═════════════════════════════════════════════════════════════════════════════

func BenchmarkKBVecAdd(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{5}))

	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20, 1 << 22, 1 << 24}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSize(n)), func(b *testing.B) {
			a, _ := NewKBVector(dev, n)
			defer a.Free()
			bv, _ := NewKBVector(dev, n)
			defer bv.Free()
			c, _ := NewKBVector(dev, n)
			defer c.Free()
			d := randSlice(rng, n)
			a.CopyFromHost(d)
			bv.CopyFromHost(d)
			Sync(dev)

			b.SetBytes(int64(n) * 4 * 3) // read 2 + write 1
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				c.Add(a, bv)
			}
			Sync(dev)
		})
	}
}

func BenchmarkKBVecMul(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{6}))

	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20, 1 << 22, 1 << 24}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSize(n)), func(b *testing.B) {
			a, _ := NewKBVector(dev, n)
			defer a.Free()
			bv, _ := NewKBVector(dev, n)
			defer bv.Free()
			c, _ := NewKBVector(dev, n)
			defer c.Free()
			d := randSlice(rng, n)
			a.CopyFromHost(d)
			bv.CopyFromHost(d)
			Sync(dev)

			b.SetBytes(int64(n) * 4 * 3) // read 2 + write 1
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				c.Mul(a, bv)
			}
			Sync(dev)
		})
	}
}

func BenchmarkKBVecScale(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{7}))

	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20, 1 << 22, 1 << 24}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSize(n)), func(b *testing.B) {
			a, _ := NewKBVector(dev, n)
			defer a.Free()
			d := randSlice(rng, n)
			a.CopyFromHost(d)
			scalar := randKB(rng)
			Sync(dev)

			b.SetBytes(int64(n) * 4 * 2) // read + write
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				a.Scale(scalar)
			}
			Sync(dev)
		})
	}
}

func BenchmarkKBVecScaleByPowers(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{8}))

	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20, 1 << 22, 1 << 24}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSize(n)), func(b *testing.B) {
			a, _ := NewKBVector(dev, n)
			defer a.Free()
			d := randSlice(rng, n)
			a.CopyFromHost(d)
			g := randKB(rng)
			Sync(dev)

			b.SetBytes(int64(n) * 4 * 2)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				a.ScaleByPowers(g)
			}
			Sync(dev)
		})
	}
}

func BenchmarkKBVecBitReverse(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{9}))

	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20, 1 << 22, 1 << 24}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSize(n)), func(b *testing.B) {
			a, _ := NewKBVector(dev, n)
			defer a.Free()
			d := randSlice(rng, n)
			a.CopyFromHost(d)
			Sync(dev)

			b.SetBytes(int64(n) * 4 * 2)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				a.BitReverse()
			}
			Sync(dev)
		})
	}
}

// ═════════════════════════════════════════════════════════════════════════════
// 3. KoalaBear NTT (FFT forward, inverse, coset)
// ═════════════════════════════════════════════════════════════════════════════

func BenchmarkKBNTTForward(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{10}))

	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20, 1 << 22, 1 << 24}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSize(n)), func(b *testing.B) {
			dom, err := NewGPUFFTDomain(dev, n)
			require.NoError(b, err)
			defer dom.Free()
			vec, _ := NewKBVector(dev, n)
			defer vec.Free()
			d := randSlice(rng, n)
			vec.CopyFromHost(d)
			Sync(dev)

			b.SetBytes(int64(n) * 4)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dom.FFT(vec)
			}
			Sync(dev)
		})
	}
}

func BenchmarkKBNTTInverse(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{11}))

	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20, 1 << 22, 1 << 24}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSize(n)), func(b *testing.B) {
			dom, err := NewGPUFFTDomain(dev, n)
			require.NoError(b, err)
			defer dom.Free()
			vec, _ := NewKBVector(dev, n)
			defer vec.Free()
			d := randSlice(rng, n)
			vec.CopyFromHost(d)
			Sync(dev)

			b.SetBytes(int64(n) * 4)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dom.FFTInverse(vec)
			}
			Sync(dev)
		})
	}
}

func BenchmarkKBNTTCosetFwd(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{12}))

	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20, 1 << 22, 1 << 24}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSize(n)), func(b *testing.B) {
			dom, err := NewGPUFFTDomain(dev, n)
			require.NoError(b, err)
			defer dom.Free()
			vec, _ := NewKBVector(dev, n)
			defer vec.Free()
			d := randSlice(rng, n)
			vec.CopyFromHost(d)
			g := fft.NewDomain(uint64(n)).Generator
			Sync(dev)

			b.SetBytes(int64(n) * 4)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dom.CosetFFT(vec, g)
			}
			Sync(dev)
		})
	}
}

// End-to-end NTT with H2D + compute + D2H breakdown
func BenchmarkKBNTTE2E(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{13}))

	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20, 1 << 22, 1 << 24}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSize(n)), func(b *testing.B) {
			dom, err := NewGPUFFTDomain(dev, n)
			require.NoError(b, err)
			defer dom.Free()
			vec, _ := NewKBVector(dev, n)
			defer vec.Free()
			src := randSlice(rng, n)
			dst := make([]koalabear.Element, n)
			Sync(dev)

			var h2dTotal, computeTotal, d2hTotal time.Duration
			b.SetBytes(int64(n) * 4 * 2) // H2D + D2H
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				t0 := time.Now()
				vec.CopyFromHost(src)
				Sync(dev)
				t1 := time.Now()
				dom.FFT(vec)
				Sync(dev)
				t2 := time.Now()
				vec.CopyToHost(dst)
				Sync(dev)
				t3 := time.Now()

				h2dTotal += t1.Sub(t0)
				computeTotal += t2.Sub(t1)
				d2hTotal += t3.Sub(t2)
			}
			reportTiming(b, h2dTotal/time.Duration(b.N), computeTotal/time.Duration(b.N), d2hTotal/time.Duration(b.N), int64(n)*4*2)
		})
	}
}

// ═════════════════════════════════════════════════════════════════════════════
// 4. Batch NTT
// ═════════════════════════════════════════════════════════════════════════════

func BenchmarkKBBatchCosetFFT(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{14}))

	n := 1 << 18
	batches := []int{1, 10, 100, 500, 1000}
	dom, _ := NewGPUFFTDomain(dev, n)
	defer dom.Free()
	g := fft.NewDomain(uint64(n)).Generator

	for _, batch := range batches {
		b.Run(fmt.Sprintf("n=%s_batch=%d", fmtSize(n), batch), func(b *testing.B) {
			total := batch * n
			vec, _ := NewKBVector(dev, total)
			defer vec.Free()
			d := randSlice(rng, total)
			vec.CopyFromHost(d)
			Sync(dev)

			b.SetBytes(int64(total) * 4)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dom.BatchCosetFFTBitRev(vec, batch, g)
			}
			Sync(dev)
		})
	}
}

func BenchmarkKBBatchIFFTScale(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{15}))

	n := 1 << 18
	batches := []int{1, 10, 100, 500, 1000}
	dom, _ := NewGPUFFTDomain(dev, n)
	defer dom.Free()
	var nInv koalabear.Element
	nInv.SetUint64(uint64(n))
	nInv.Inverse(&nInv)

	for _, batch := range batches {
		b.Run(fmt.Sprintf("n=%s_batch=%d", fmtSize(n), batch), func(b *testing.B) {
			total := batch * n
			vec, _ := NewKBVector(dev, total)
			defer vec.Free()
			d := randSlice(rng, total)
			vec.CopyFromHost(d)
			Sync(dev)

			b.SetBytes(int64(total) * 4)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dom.BatchIFFTScale(vec, batch, nInv)
			}
			Sync(dev)
		})
	}
}

// NTT at various sizes for scaling analysis
func BenchmarkKBNTTScaling(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{16}))

	sizes := []int{1 << 10, 1 << 12, 1 << 14, 1 << 16, 1 << 18, 1 << 20, 1 << 22, 1 << 24}
	// Note: KoalaBear two-adicity is 24, max NTT size is 2^24 = 16M
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSize(n)), func(b *testing.B) {
			dom, err := NewGPUFFTDomain(dev, n)
			if err != nil {
				b.Skip("OOM:", err)
			}
			defer dom.Free()
			vec, err := NewKBVector(dev, n)
			if err != nil {
				b.Skip("OOM:", err)
			}
			defer vec.Free()
			d := randSlice(rng, n)
			vec.CopyFromHost(d)
			Sync(dev)

			b.SetBytes(int64(n) * 4)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dom.FFT(vec)
			}
			Sync(dev)
		})
	}
}

// ═════════════════════════════════════════════════════════════════════════════
// 5. KoalaBear NTT vs CPU
// ═════════════════════════════════════════════════════════════════════════════

func BenchmarkKBNTTvsCPU(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{17}))

	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("GPU/n=%s", fmtSize(n)), func(b *testing.B) {
			dom, _ := NewGPUFFTDomain(dev, n)
			defer dom.Free()
			vec, _ := NewKBVector(dev, n)
			defer vec.Free()
			d := randSlice(rng, n)
			vec.CopyFromHost(d)
			Sync(dev)

			b.SetBytes(int64(n) * 4)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dom.FFT(vec)
			}
			Sync(dev)
		})
		b.Run(fmt.Sprintf("CPU/n=%s", fmtSize(n)), func(b *testing.B) {
			cpuDom := fft.NewDomain(uint64(n))
			d := randSlice(rng, n)

			b.SetBytes(int64(n) * 4)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cpuDom.FFT(d, fft.DIF)
			}
		})
	}
}

// ═════════════════════════════════════════════════════════════════════════════
// 6. Poseidon2 (batch hashing)
// ═════════════════════════════════════════════════════════════════════════════

func BenchmarkPoseidon2Compress(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{20}))

	p2, err := NewGPUPoseidon2(dev, 16)
	require.NoError(b, err)
	defer p2.Free()

	counts := []int{1, 64, 256, 1024, 4096, 16384, 65536}
	for _, count := range counts {
		b.Run(fmt.Sprintf("count=%s", fmtSize(count)), func(b *testing.B) {
			input := make([]koalabear.Element, 16*count)
			for i := range input {
				input[i] = randKB(rng)
			}

			b.SetBytes(int64(16*count) * 4) // input size
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				p2.CompressBatch(input, count)
			}
		})
	}
}

func BenchmarkPoseidon2E2E(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{21}))

	p2, err := NewGPUPoseidon2(dev, 16)
	require.NoError(b, err)
	defer p2.Free()

	counts := []int{256, 4096, 65536}
	for _, count := range counts {
		b.Run(fmt.Sprintf("count=%s", fmtSize(count)), func(b *testing.B) {
			input := make([]koalabear.Element, 16*count)
			for i := range input {
				input[i] = randKB(rng)
			}
			inputBytes := int64(16*count) * 4
			outputBytes := int64(8*count) * 4

			var computeTotal time.Duration
			b.SetBytes(inputBytes + outputBytes)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				t0 := time.Now()
				p2.CompressBatch(input, count)
				computeTotal += time.Since(t0)
			}
			b.ReportMetric(float64(computeTotal.Microseconds())/float64(b.N), "compute_µs")
			hps := float64(count) / (computeTotal.Seconds() / float64(b.N))
			b.ReportMetric(hps/1e6, "Mhash/s")
		})
	}
}

// ═════════════════════════════════════════════════════════════════════════════
// 7. GPU Linear Combination (E4)
// ═════════════════════════════════════════════════════════════════════════════

func BenchmarkLinCombE4(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{25}))

	configs := []struct{ nRows, nCols int }{
		{16, 1024},
		{64, 4096},
		{128, 16384},
		{256, 65536},
		{512, 65536},
		{1024, 65536},
	}

	for _, cfg := range configs {
		b.Run(fmt.Sprintf("rows=%d_cols=%s", cfg.nRows, fmtSize(cfg.nCols)), func(b *testing.B) {
			rows := make([]*KBVector, cfg.nRows)
			for i := range rows {
				v, _ := NewKBVector(dev, cfg.nCols)
				d := randSlice(rng, cfg.nCols)
				v.CopyFromHost(d)
				rows[i] = v
			}
			defer func() {
				for _, r := range rows {
					r.Free()
				}
			}()

			alpha := randE4(rng)
			Sync(dev)

			dataBytes := int64(cfg.nRows) * int64(cfg.nCols) * 4 // input
			dataBytes += int64(cfg.nCols) * 16                    // output (E4 = 4×uint32)
			b.SetBytes(dataBytes)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				GPULinCombE4(dev, rows, alpha, cfg.nCols)
			}
		})
	}
}

// ═════════════════════════════════════════════════════════════════════════════
// 8. Vortex Commit Pipeline (end-to-end)
// ═════════════════════════════════════════════════════════════════════════════

func BenchmarkVortexCommitE2E(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{30}))

	configs := []struct {
		nCols, nRows, rate int
	}{
		{1024, 128, 2},
		{4096, 256, 2},
		{16384, 256, 2},
		{65536, 512, 2},
		{1 << 17, 1 << 10, 2},
		{1 << 18, 1 << 11, 2},
		{1 << 19, 1 << 11, 2},
		{1 << 20, 1 << 12, 2},
	}

	for _, cfg := range configs {
		name := fmt.Sprintf("cols=%s_rows=%s_rate=%d", fmtSize(cfg.nCols), fmtSize(cfg.nRows), cfg.rate)
		b.Run(name, func(b *testing.B) {
			sisParams, err := sis.NewRSis(0, 9, 16, cfg.nRows)
			if err != nil {
				b.Skip("SIS init:", err)
			}
			params, err := NewParams(cfg.nCols, cfg.nRows, sisParams, cfg.rate, 32)
			if err != nil {
				b.Skip("Params:", err)
			}
			m := randMatrix(rng, cfg.nRows, cfg.nCols)

			gv, err := NewGPUVortex(dev, params, cfg.nRows)
			if err != nil {
				b.Skip("GPUVortex:", err)
			}
			defer gv.Free()

			inputBytes := int64(cfg.nCols) * int64(cfg.nRows) * 4
			b.SetBytes(inputBytes)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _, err := gv.Commit(m)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Breakdown: H2D copy, GPU commit, and total.
func BenchmarkVortexCommitBreakdown(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{31}))

	nCols := 16384
	nRows := 256
	rate := 2

	sisParams, _ := sis.NewRSis(0, 9, 16, nRows)
	params, _ := NewParams(nCols, nRows, sisParams, rate, 32)
	m := randMatrix(rng, nRows, nCols)

	gv, err := NewGPUVortex(dev, params, nRows)
	require.NoError(b, err)
	defer gv.Free()

	inputBytes := int64(nCols) * int64(nRows) * 4

	b.Run("full_commit", func(b *testing.B) {
		b.SetBytes(inputBytes)
		var totalTime time.Duration
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			t0 := time.Now()
			_, _, err := gv.Commit(m)
			totalTime += time.Since(t0)
			if err != nil {
				b.Fatal(err)
			}
		}
		avgMs := float64(totalTime.Milliseconds()) / float64(b.N)
		b.ReportMetric(avgMs, "ms/op")
		mbps := float64(inputBytes) / (totalTime.Seconds() / float64(b.N)) / (1024 * 1024)
		b.ReportMetric(mbps, "MB/s")
	})
}

// ═════════════════════════════════════════════════════════════════════════════
// 9. Vortex Prove (lincomb + extract)
// ═════════════════════════════════════════════════════════════════════════════

func BenchmarkVortexProve(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{32}))

	nCols := 16384
	nRows := 256
	rate := 2
	nSelected := 32

	sisParams, _ := sis.NewRSis(0, 9, 16, nRows)
	params, _ := NewParams(nCols, nRows, sisParams, rate, nSelected)
	m := randMatrix(rng, nRows, nCols)

	gv, err := NewGPUVortex(dev, params, nRows)
	require.NoError(b, err)
	defer gv.Free()

	cs, _, err := gv.Commit(m)
	require.NoError(b, err)

	alpha := randE4(rng)
	selectedCols := make([]int, nSelected)
	for i := range selectedCols {
		selectedCols[i] = i * (nCols * rate / nSelected)
	}

	b.Run("lincomb", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := cs.LinComb(alpha)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("extract_columns", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := cs.ExtractColumns(selectedCols)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("full_prove", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := cs.Prove(alpha, selectedCols)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ═════════════════════════════════════════════════════════════════════════════
// 10. Vortex GPU vs CPU comparison
// ═════════════════════════════════════════════════════════════════════════════

func BenchmarkVortexGPUvsCPU(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{33}))

	configs := []struct {
		nCols, nRows int
	}{
		{4096, 128},
		{16384, 256},
		{65536, 512},
	}

	for _, cfg := range configs {
		rate := 2
		nSelected := 32
		name := fmt.Sprintf("cols=%s_rows=%d", fmtSize(cfg.nCols), cfg.nRows)

		sisParams, _ := sis.NewRSis(0, 9, 16, cfg.nRows)
		params, _ := NewParams(cfg.nCols, cfg.nRows, sisParams, rate, nSelected)
		m := randMatrix(rng, cfg.nRows, cfg.nCols)

		gv, err := NewGPUVortex(dev, params, cfg.nRows)
		if err != nil {
			b.Skip("GPUVortex:", err)
		}
		defer gv.Free()

		inputBytes := int64(cfg.nCols) * int64(cfg.nRows) * 4

		b.Run(name+"/GPU", func(b *testing.B) {
			b.SetBytes(inputBytes)
			for i := 0; i < b.N; i++ {
				if _, _, err := gv.Commit(m); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run(name+"/CPU", func(b *testing.B) {
			b.SetBytes(inputBytes)
			for i := 0; i < b.N; i++ {
				if _, _, err := params.Commit(m); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// ═════════════════════════════════════════════════════════════════════════════
// 11. KB NTT — GPU vs CPU comparison at large sizes (base field)
// ═════════════════════════════════════════════════════════════════════════════

func BenchmarkNTTComparison(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{40}))

	sizes := []int{1 << 20, 1 << 21, 1 << 22, 1 << 23, 1 << 24}
	for _, n := range sizes {
		// ── GPU FFT (base field) ──
		b.Run(fmt.Sprintf("GPU_FFT/n=%s", fmtSize(n)), func(b *testing.B) {
			dom, err := NewGPUFFTDomain(dev, n)
			if err != nil {
				b.Skip("domain init:", err)
			}
			defer dom.Free()
			vec, err := NewKBVector(dev, n)
			if err != nil {
				b.Skip("alloc:", err)
			}
			defer vec.Free()
			d := randSlice(rng, n)
			vec.CopyFromHost(d)
			Sync(dev)

			b.SetBytes(int64(n) * 4)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dom.FFT(vec)
			}
			Sync(dev)
		})

		// ── CPU FFT (base field) ──
		b.Run(fmt.Sprintf("CPU_FFT/n=%s", fmtSize(n)), func(b *testing.B) {
			cpuDom := fft.NewDomain(uint64(n))
			d := randSlice(rng, n)

			b.SetBytes(int64(n) * 4)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cpuDom.FFT(d, fft.DIF)
			}
		})

		// ── GPU IFFT (base field) ──
		b.Run(fmt.Sprintf("GPU_IFFT/n=%s", fmtSize(n)), func(b *testing.B) {
			dom, err := NewGPUFFTDomain(dev, n)
			if err != nil {
				b.Skip("domain init:", err)
			}
			defer dom.Free()
			vec, err := NewKBVector(dev, n)
			if err != nil {
				b.Skip("alloc:", err)
			}
			defer vec.Free()
			d := randSlice(rng, n)
			vec.CopyFromHost(d)
			Sync(dev)

			b.SetBytes(int64(n) * 4)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dom.FFTInverse(vec)
			}
			Sync(dev)
		})

		// ── CPU IFFT (base field) ──
		b.Run(fmt.Sprintf("CPU_IFFT/n=%s", fmtSize(n)), func(b *testing.B) {
			cpuDom := fft.NewDomain(uint64(n))
			d := randSlice(rng, n)

			b.SetBytes(int64(n) * 4)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cpuDom.FFTInverse(d, fft.DIT)
			}
		})

		// ── GPU CosetFFT (base field) ──
		b.Run(fmt.Sprintf("GPU_CosetFFT/n=%s", fmtSize(n)), func(b *testing.B) {
			dom, err := NewGPUFFTDomain(dev, n)
			if err != nil {
				b.Skip("domain init:", err)
			}
			defer dom.Free()
			vec, err := NewKBVector(dev, n)
			if err != nil {
				b.Skip("alloc:", err)
			}
			defer vec.Free()
			d := randSlice(rng, n)
			vec.CopyFromHost(d)
			g := fft.NewDomain(uint64(n)).Generator
			Sync(dev)

			b.SetBytes(int64(n) * 4)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dom.CosetFFT(vec, g)
			}
			Sync(dev)
		})

		// ── CPU CosetFFT (base field) ──
		b.Run(fmt.Sprintf("CPU_CosetFFT/n=%s", fmtSize(n)), func(b *testing.B) {
			cpuDom := fft.NewDomain(uint64(n))
			d := randSlice(rng, n)

			b.SetBytes(int64(n) * 4)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cpuDom.FFT(d, fft.DIF, fft.OnCoset())
			}
		})
	}
}

// ═════════════════════════════════════════════════════════════════════════════
// 12. E4 NTT — GPU vs CPU comparison (extension field)
// ═════════════════════════════════════════════════════════════════════════════

func BenchmarkE4NTTComparison(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{41}))

	sizes := []int{1 << 20, 1 << 21, 1 << 22, 1 << 23, 1 << 24}
	for _, n := range sizes {
		// ── GPU FFTE4 (end-to-end: AoS→SoA→4×NTT→SoA→AoS) ──
		b.Run(fmt.Sprintf("GPU_FFTE4/n=%s", fmtSize(n)), func(b *testing.B) {
			dom, err := NewGPUFFTDomain(dev, n)
			if err != nil {
				b.Skip("domain init:", err)
			}
			defer dom.Free()

			data := make([]fext.E4, n)
			for i := range data {
				data[i] = randE4(rng)
			}

			b.SetBytes(int64(n) * 16) // 4 × uint32 per E4
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dom.FFTE4(data)
			}
		})

		// ── CPU FFTExt ──
		b.Run(fmt.Sprintf("CPU_FFTExt/n=%s", fmtSize(n)), func(b *testing.B) {
			cpuDom := fft.NewDomain(uint64(n))
			data := make([]fext.E4, n)
			for i := range data {
				data[i] = randE4(rng)
			}

			b.SetBytes(int64(n) * 16)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cpuDom.FFTExt(data, fft.DIF)
			}
		})

		// ── GPU CosetFFTE4 ──
		b.Run(fmt.Sprintf("GPU_CosetFFTE4/n=%s", fmtSize(n)), func(b *testing.B) {
			dom, err := NewGPUFFTDomain(dev, n)
			if err != nil {
				b.Skip("domain init:", err)
			}
			defer dom.Free()

			cpuDom := fft.NewDomain(uint64(n))
			data := make([]fext.E4, n)
			for i := range data {
				data[i] = randE4(rng)
			}

			b.SetBytes(int64(n) * 16)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dom.CosetFFTE4(data, cpuDom.FrMultiplicativeGen)
			}
		})

		// ── CPU CosetFFTExt ──
		b.Run(fmt.Sprintf("CPU_CosetFFTExt/n=%s", fmtSize(n)), func(b *testing.B) {
			cpuDom := fft.NewDomain(uint64(n))
			data := make([]fext.E4, n)
			for i := range data {
				data[i] = randE4(rng)
			}

			b.SetBytes(int64(n) * 16)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cpuDom.FFTExt(data, fft.DIF, fft.OnCoset())
			}
		})

		// ── GPU BatchCosetFFTE4 (SoA, device-resident, no AoS transpose) ──
		b.Run(fmt.Sprintf("GPU_BatchCosetFFTE4/n=%s", fmtSize(n)), func(b *testing.B) {
			dom, err := NewGPUFFTDomain(dev, n)
			if err != nil {
				b.Skip("domain init:", err)
			}
			defer dom.Free()

			vec, err := NewKBVector(dev, n*4)
			if err != nil {
				b.Skip("alloc:", err)
			}
			defer vec.Free()

			// Fill with random data on GPU
			d := randSlice(rng, n*4)
			vec.CopyFromHost(d)
			g := fft.NewDomain(uint64(n)).FrMultiplicativeGen
			Sync(dev)

			b.SetBytes(int64(n) * 16)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dom.BatchCosetFFTE4BitRev(vec, n, g)
			}
			Sync(dev)
		})
	}
}

// ═════════════════════════════════════════════════════════════════════════════
// 13. E4 IFFT — GPU vs CPU comparison
// ═════════════════════════════════════════════════════════════════════════════

func BenchmarkE4IFFTComparison(b *testing.B) {
	dev := newBenchDevice(b)
	rng := rand.New(rand.NewChaCha8([32]byte{42}))

	sizes := []int{1 << 20, 1 << 21, 1 << 22, 1 << 23, 1 << 24}
	for _, n := range sizes {
		// ── GPU FFTInverseE4 ──
		b.Run(fmt.Sprintf("GPU_IFFE4/n=%s", fmtSize(n)), func(b *testing.B) {
			dom, err := NewGPUFFTDomain(dev, n)
			if err != nil {
				b.Skip("domain init:", err)
			}
			defer dom.Free()

			data := make([]fext.E4, n)
			for i := range data {
				data[i] = randE4(rng)
			}

			b.SetBytes(int64(n) * 16)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dom.FFTInverseE4(data)
			}
		})

		// ── CPU FFTInverseExt ──
		b.Run(fmt.Sprintf("CPU_IFFTExt/n=%s", fmtSize(n)), func(b *testing.B) {
			cpuDom := fft.NewDomain(uint64(n))
			data := make([]fext.E4, n)
			for i := range data {
				data[i] = randE4(rng)
			}

			b.SetBytes(int64(n) * 16)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cpuDom.FFTInverseExt(data, fft.DIT)
			}
		})

		// ── GPU BatchIFFTScaleE4 (SoA, device-resident) ──
		b.Run(fmt.Sprintf("GPU_BatchIFFTE4/n=%s", fmtSize(n)), func(b *testing.B) {
			dom, err := NewGPUFFTDomain(dev, n)
			if err != nil {
				b.Skip("domain init:", err)
			}
			defer dom.Free()

			vec, err := NewKBVector(dev, n*4)
			if err != nil {
				b.Skip("alloc:", err)
			}
			defer vec.Free()
			d := randSlice(rng, n*4)
			vec.CopyFromHost(d)
			var nInv koalabear.Element
			nInv.SetUint64(uint64(n))
			nInv.Inverse(&nInv)
			Sync(dev)

			b.SetBytes(int64(n) * 16)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dom.BatchIFFTScaleE4(vec, n, nInv)
			}
			Sync(dev)
		})
	}
}

// Helpers (randKB, randE4, randMatrix defined in vortex_test.go)
