//go:build cuda

package plonk_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
	"github.com/consensys/linea-monorepo/prover/gpu"
	"github.com/consensys/linea-monorepo/prover/gpu/plonk"
	"github.com/stretchr/testify/require"
)

func fmtSz(n int) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%dM", n>>20)
	case n >= 1<<10:
		return fmt.Sprintf("%dK", n>>10)
	default:
		return fmt.Sprintf("%d", n)
	}
}

func reportTm(b *testing.B, h2d, compute, d2h time.Duration, bytes int64) {
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

func newBenchDev(b *testing.B) *gpu.Device {
	b.Helper()
	dev, err := gpu.New()
	require.NoError(b, err)
	b.Cleanup(func() { dev.Close() })
	return dev
}

func genFrVec(n int) fr.Vector {
	v := make(fr.Vector, n)
	for i := range v {
		v[i].SetRandom()
	}
	return v
}

// ═════════════════════════════════════════════════════════════════════════════
// 1. FrVector: H2D, D2H, D2D transfers (BLS12-377)
// ═════════════════════════════════════════════════════════════════════════════

func BenchmarkFrVecH2D(b *testing.B) {
	dev := newBenchDev(b)
	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20, 1 << 22}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSz(n)), func(b *testing.B) {
			vec, _ := plonk.NewFrVector(dev, n)
			defer vec.Free()
			data := genFrVec(n)
			dev.Sync()

			b.SetBytes(int64(n) * 32)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				vec.CopyFromHost(data)
			}
			dev.Sync()
		})
	}
}

func BenchmarkFrVecD2H(b *testing.B) {
	dev := newBenchDev(b)
	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20, 1 << 22}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSz(n)), func(b *testing.B) {
			vec, _ := plonk.NewFrVector(dev, n)
			defer vec.Free()
			data := genFrVec(n)
			vec.CopyFromHost(data)
			dst := make(fr.Vector, n)
			dev.Sync()

			b.SetBytes(int64(n) * 32)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				vec.CopyToHost(dst)
			}
			dev.Sync()
		})
	}
}

func BenchmarkFrVecD2D(b *testing.B) {
	dev := newBenchDev(b)
	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20, 1 << 22}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSz(n)), func(b *testing.B) {
			src, _ := plonk.NewFrVector(dev, n)
			defer src.Free()
			dst, _ := plonk.NewFrVector(dev, n)
			defer dst.Free()
			data := genFrVec(n)
			src.CopyFromHost(data)
			dev.Sync()

			b.SetBytes(int64(n) * 32)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dst.CopyFromDevice(src)
			}
			dev.Sync()
		})
	}
}

// ═════════════════════════════════════════════════════════════════════════════
// 2. FrVector: arithmetic operations (BLS12-377)
// ═════════════════════════════════════════════════════════════════════════════

func BenchmarkFrVecAdd(b *testing.B) {
	dev := newBenchDev(b)
	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20, 1 << 22}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSz(n)), func(b *testing.B) {
			a, _ := plonk.NewFrVector(dev, n)
			defer a.Free()
			bv, _ := plonk.NewFrVector(dev, n)
			defer bv.Free()
			c, _ := plonk.NewFrVector(dev, n)
			defer c.Free()
			d := genFrVec(n)
			a.CopyFromHost(d)
			bv.CopyFromHost(d)
			dev.Sync()

			b.SetBytes(int64(n) * 32 * 3)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				c.Add(a, bv)
			}
			dev.Sync()
		})
	}
}

func BenchmarkFrVecMul(b *testing.B) {
	dev := newBenchDev(b)
	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20, 1 << 22}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSz(n)), func(b *testing.B) {
			a, _ := plonk.NewFrVector(dev, n)
			defer a.Free()
			bv, _ := plonk.NewFrVector(dev, n)
			defer bv.Free()
			c, _ := plonk.NewFrVector(dev, n)
			defer c.Free()
			d := genFrVec(n)
			a.CopyFromHost(d)
			bv.CopyFromHost(d)
			dev.Sync()

			b.SetBytes(int64(n) * 32 * 3)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				c.Mul(a, bv)
			}
			dev.Sync()
		})
	}
}

func BenchmarkFrVecSub(b *testing.B) {
	dev := newBenchDev(b)
	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20, 1 << 22}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSz(n)), func(b *testing.B) {
			a, _ := plonk.NewFrVector(dev, n)
			defer a.Free()
			bv, _ := plonk.NewFrVector(dev, n)
			defer bv.Free()
			c, _ := plonk.NewFrVector(dev, n)
			defer c.Free()
			d := genFrVec(n)
			a.CopyFromHost(d)
			bv.CopyFromHost(d)
			dev.Sync()

			b.SetBytes(int64(n) * 32 * 3)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				c.Sub(a, bv)
			}
			dev.Sync()
		})
	}
}

func BenchmarkFrVecScaleByPowers(b *testing.B) {
	dev := newBenchDev(b)
	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20, 1 << 22}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSz(n)), func(b *testing.B) {
			a, _ := plonk.NewFrVector(dev, n)
			defer a.Free()
			d := genFrVec(n)
			a.CopyFromHost(d)
			var g fr.Element
			g.SetRandom()
			dev.Sync()

			b.SetBytes(int64(n) * 32 * 2)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				a.ScaleByPowers(g)
			}
			dev.Sync()
		})
	}
}

func BenchmarkFrVecBatchInvert(b *testing.B) {
	dev := newBenchDev(b)
	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSz(n)), func(b *testing.B) {
			a, _ := plonk.NewFrVector(dev, n)
			defer a.Free()
			tmp, _ := plonk.NewFrVector(dev, n)
			defer tmp.Free()
			d := genFrVec(n)
			a.CopyFromHost(d)
			dev.Sync()

			b.SetBytes(int64(n) * 32)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				a.BatchInvert(tmp)
			}
			dev.Sync()
		})
	}
}

// ═════════════════════════════════════════════════════════════════════════════
// 3. BLS12-377 NTT (FFT)
// ═════════════════════════════════════════════════════════════════════════════

func BenchmarkBLSFFTForward(b *testing.B) {
	dev := newBenchDev(b)
	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20, 1 << 22}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSz(n)), func(b *testing.B) {
			dom, _ := plonk.NewFFTDomain(dev, n)
			defer dom.Close()
			vec, _ := plonk.NewFrVector(dev, n)
			defer vec.Free()
			d := genFrVec(n)
			vec.CopyFromHost(d)
			dev.Sync()

			b.SetBytes(int64(n) * 32)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dom.FFT(vec)
			}
			dev.Sync()
		})
	}
}

func BenchmarkBLSFFTInverse(b *testing.B) {
	dev := newBenchDev(b)
	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20, 1 << 22}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSz(n)), func(b *testing.B) {
			dom, _ := plonk.NewFFTDomain(dev, n)
			defer dom.Close()
			vec, _ := plonk.NewFrVector(dev, n)
			defer vec.Free()
			d := genFrVec(n)
			vec.CopyFromHost(d)
			dev.Sync()

			b.SetBytes(int64(n) * 32)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dom.FFTInverse(vec)
			}
			dev.Sync()
		})
	}
}

func BenchmarkBLSFFTCoset(b *testing.B) {
	dev := newBenchDev(b)
	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20, 1 << 22}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSz(n)), func(b *testing.B) {
			dom, _ := plonk.NewFFTDomain(dev, n)
			defer dom.Close()
			vec, _ := plonk.NewFrVector(dev, n)
			defer vec.Free()
			d := genFrVec(n)
			vec.CopyFromHost(d)
			g := fft.NewDomain(uint64(n)).FrMultiplicativeGen
			dev.Sync()

			b.SetBytes(int64(n) * 32)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dom.CosetFFT(vec, g)
			}
			dev.Sync()
		})
	}
}

func BenchmarkBLSFFTE2E(b *testing.B) {
	dev := newBenchDev(b)
	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20, 1 << 22}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSz(n)), func(b *testing.B) {
			dom, _ := plonk.NewFFTDomain(dev, n)
			defer dom.Close()
			vec, _ := plonk.NewFrVector(dev, n)
			defer vec.Free()
			src := genFrVec(n)
			dst := make(fr.Vector, n)
			dev.Sync()

			var h2dTotal, computeTotal, d2hTotal time.Duration
			b.SetBytes(int64(n) * 32 * 2)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				t0 := time.Now()
				vec.CopyFromHost(src)
				dev.Sync()
				t1 := time.Now()
				dom.FFT(vec)
				dev.Sync()
				t2 := time.Now()
				vec.CopyToHost(dst)
				dev.Sync()
				t3 := time.Now()

				h2dTotal += t1.Sub(t0)
				computeTotal += t2.Sub(t1)
				d2hTotal += t3.Sub(t2)
			}
			reportTm(b, h2dTotal/time.Duration(b.N), computeTotal/time.Duration(b.N), d2hTotal/time.Duration(b.N), int64(n)*32*2)
		})
	}
}

func BenchmarkBLSFFTScaling(b *testing.B) {
	dev := newBenchDev(b)
	sizes := []int{1 << 10, 1 << 12, 1 << 14, 1 << 16, 1 << 18, 1 << 20, 1 << 22, 1 << 24}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSz(n)), func(b *testing.B) {
			dom, err := plonk.NewFFTDomain(dev, n)
			if err != nil {
				b.Skip("OOM:", err)
			}
			defer dom.Close()
			vec, err := plonk.NewFrVector(dev, n)
			if err != nil {
				b.Skip("OOM:", err)
			}
			defer vec.Free()
			d := genFrVec(n)
			vec.CopyFromHost(d)
			dev.Sync()

			b.SetBytes(int64(n) * 32)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dom.FFT(vec)
			}
			dev.Sync()
		})
	}
}

// ═════════════════════════════════════════════════════════════════════════════
// 4. BLS12-377 GPU vs CPU FFT
// ═════════════════════════════════════════════════════════════════════════════

func BenchmarkBLSFFTvsCPU(b *testing.B) {
	dev := newBenchDev(b)
	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("GPU/n=%s", fmtSz(n)), func(b *testing.B) {
			dom, _ := plonk.NewFFTDomain(dev, n)
			defer dom.Close()
			vec, _ := plonk.NewFrVector(dev, n)
			defer vec.Free()
			d := genFrVec(n)
			vec.CopyFromHost(d)
			dev.Sync()

			b.SetBytes(int64(n) * 32)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dom.FFT(vec)
			}
			dev.Sync()
		})
		b.Run(fmt.Sprintf("CPU/n=%s", fmtSz(n)), func(b *testing.B) {
			cpuDom := fft.NewDomain(uint64(n))
			d := genFrVec(n)

			b.SetBytes(int64(n) * 32)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cpuDom.FFT(d, fft.DIF)
			}
		})
	}
}

// ═════════════════════════════════════════════════════════════════════════════
// 5. BLS12-377 MSM (uses disk-loaded SRS)
// ═════════════════════════════════════════════════════════════════════════════

func BenchmarkMSMComprehensive(b *testing.B) {
	dev := newBenchDev(b)

	sizes := []int{1 << 14, 1 << 16, 1 << 18, 1 << 20}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSz(n)), func(b *testing.B) {
			tePoints := loadSRSTEPoints(b, n)
			pts, err := plonk.PinG1TEPoints(tePoints)
			require.NoError(b, err)
			defer pts.Free()

			msm, err := plonk.NewG1MSM(dev, pts)
			require.NoError(b, err)
			defer msm.Close()

			scalars := loadTestScalars(n)
			dev.Sync()

			b.SetBytes(int64(n) * 32)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				msm.MultiExp(scalars)
			}
		})
	}
}

func BenchmarkMSMBreakdown(b *testing.B) {
	dev := newBenchDev(b)

	sizes := []int{1 << 16, 1 << 18, 1 << 20}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%s", fmtSz(n)), func(b *testing.B) {
			tePoints := loadSRSTEPoints(b, n)
			pts, err := plonk.PinG1TEPoints(tePoints)
			require.NoError(b, err)
			defer pts.Free()

			msm, err := plonk.NewG1MSM(dev, pts)
			require.NoError(b, err)
			defer msm.Close()

			scalars := loadTestScalars(n)

			var computeTotal time.Duration
			b.SetBytes(int64(n) * 32)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				t0 := time.Now()
				msm.MultiExp(scalars)
				computeTotal += time.Since(t0)
			}
			b.ReportMetric(float64(computeTotal.Microseconds())/float64(b.N), "compute_µs")
		})
	}
}
