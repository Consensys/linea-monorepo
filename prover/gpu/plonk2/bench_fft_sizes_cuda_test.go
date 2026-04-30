//go:build cuda

package plonk2

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	blsfr "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	blsfft "github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
	bnfr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	bnfft "github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
	bwfr "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	bwfft "github.com/consensys/gnark-crypto/ecc/bw6-761/fr/fft"
	"github.com/stretchr/testify/require"

	"github.com/consensys/linea-monorepo/prover/gpu"
	oldplonk "github.com/consensys/linea-monorepo/prover/gpu/plonk"
)

func BenchmarkFFTForwardSizes_CUDA(b *testing.B) {
	dev, err := gpu.New()
	require.NoError(b, err, "creating CUDA device should succeed")
	defer dev.Close()

	for _, n := range fftBenchSizes(b, "PLONK2_FFT_BENCH_SIZES") {
		b.Run(fmt.Sprintf("bn254/n=%s", benchFormatCount(n)), func(b *testing.B) {
			benchPlonk2FFTForwardBN254Size(b, dev, n)
		})
		b.Run(fmt.Sprintf("bw6-761/n=%s", benchFormatCount(n)), func(b *testing.B) {
			benchPlonk2FFTForwardBW6761Size(b, dev, n)
		})
	}
}

func BenchmarkCosetFFTForwardSizes_CUDA(b *testing.B) {
	dev, err := gpu.New()
	require.NoError(b, err, "creating CUDA device should succeed")
	defer dev.Close()

	for _, n := range fftBenchSizes(b, "PLONK2_FFT_BENCH_SIZES") {
		b.Run(fmt.Sprintf("bn254/n=%s", benchFormatCount(n)), func(b *testing.B) {
			benchPlonk2CosetFFTBN254Size(b, dev, n)
		})
		b.Run(fmt.Sprintf("bw6-761/n=%s", benchFormatCount(n)), func(b *testing.B) {
			benchPlonk2CosetFFTBW6761Size(b, dev, n)
		})
	}
}

func BenchmarkCompareBLS12377FFTForwardPlonkVsPlonk2_CUDA(b *testing.B) {
	dev, err := gpu.New()
	require.NoError(b, err, "creating CUDA device should succeed")
	defer dev.Close()

	for _, n := range fftBenchSizes(b, "PLONK2_BLS_FFT_COMPARE_SIZES") {
		b.Run(fmt.Sprintf("gpu-plonk/n=%s", benchFormatCount(n)), func(b *testing.B) {
			benchOldPlonkFFTForwardBLS12377Size(b, dev, n)
		})
		b.Run(fmt.Sprintf("gpu-plonk2/n=%s", benchFormatCount(n)), func(b *testing.B) {
			benchPlonk2FFTForwardBLS12377Size(b, dev, n)
		})
	}
}

func BenchmarkCompareBLS12377CosetFFTPlonkVsPlonk2_CUDA(b *testing.B) {
	dev, err := gpu.New()
	require.NoError(b, err, "creating CUDA device should succeed")
	defer dev.Close()

	for _, n := range fftBenchSizes(b, "PLONK2_BLS_FFT_COMPARE_SIZES") {
		b.Run(fmt.Sprintf("gpu-plonk/n=%s", benchFormatCount(n)), func(b *testing.B) {
			benchOldPlonkCosetFFTBLS12377Size(b, dev, n)
		})
		b.Run(fmt.Sprintf("gpu-plonk2/n=%s", benchFormatCount(n)), func(b *testing.B) {
			benchPlonk2CosetFFTBLS12377Size(b, dev, n)
		})
	}
}

func benchPlonk2FFTForwardBN254Size(b *testing.B, dev *gpu.Device, n int) {
	input := sequentialBN254FFTInput(n)
	benchPlonk2FFTForwardRaw(b, dev, CurveBN254, fftSpecBN254(n), rawBN254(input), n)
	runtime.KeepAlive(input)
}

func benchPlonk2FFTForwardBLS12377Size(b *testing.B, dev *gpu.Device, n int) {
	input := sequentialBLS12377FFTInput(n)
	benchPlonk2FFTForwardRaw(b, dev, CurveBLS12377, fftSpecBLS12377(n), rawBLS12377(input), n)
	runtime.KeepAlive(input)
}

func benchPlonk2FFTForwardBW6761Size(b *testing.B, dev *gpu.Device, n int) {
	input := sequentialBW6761FFTInput(n)
	benchPlonk2FFTForwardRaw(b, dev, CurveBW6761, fftSpecBW6761(n), rawBW6761(input), n)
	runtime.KeepAlive(input)
}

func benchPlonk2CosetFFTBN254Size(b *testing.B, dev *gpu.Device, n int) {
	input := sequentialBN254FFTInput(n)
	domain := bnfft.NewDomain(uint64(n))
	generatorRaw := cloneRaw(rawBN254([]bnfr.Element{domain.FrMultiplicativeGen}))
	benchPlonk2CosetFFTRaw(b, dev, CurveBN254, fftSpecBN254(n), rawBN254(input), generatorRaw, n)
	runtime.KeepAlive(input)
}

func benchPlonk2CosetFFTBLS12377Size(b *testing.B, dev *gpu.Device, n int) {
	input := sequentialBLS12377FFTInput(n)
	domain := blsfft.NewDomain(uint64(n))
	generatorRaw := cloneRaw(rawBLS12377([]blsfr.Element{domain.FrMultiplicativeGen}))
	benchPlonk2CosetFFTRaw(b, dev, CurveBLS12377, fftSpecBLS12377(n), rawBLS12377(input), generatorRaw, n)
	runtime.KeepAlive(input)
}

func benchPlonk2CosetFFTBW6761Size(b *testing.B, dev *gpu.Device, n int) {
	input := sequentialBW6761FFTInput(n)
	domain := bwfft.NewDomain(uint64(n))
	generatorRaw := cloneRaw(rawBW6761([]bwfr.Element{domain.FrMultiplicativeGen}))
	benchPlonk2CosetFFTRaw(b, dev, CurveBW6761, fftSpecBW6761(n), rawBW6761(input), generatorRaw, n)
	runtime.KeepAlive(input)
}

func benchPlonk2FFTForwardRaw(
	b *testing.B,
	dev *gpu.Device,
	curve Curve,
	spec FFTDomainSpec,
	raw []uint64,
	n int,
) {
	b.Helper()
	info := mustCurveInfo(b, curve)
	domain, err := NewFFTDomain(dev, spec)
	require.NoError(b, err, "creating plonk2 FFT domain should succeed")
	defer domain.Free()
	vec, err := NewFrVector(dev, curve, n)
	require.NoError(b, err, "allocating plonk2 FFT vector should succeed")
	defer vec.Free()
	require.NoError(b, vec.CopyFromHostRaw(raw), "copying plonk2 FFT input should succeed")
	require.NoError(b, dev.Sync(), "device sync before benchmark should succeed")

	b.SetBytes(int64(n) * int64(info.ScalarLimbs*8))
	b.ReportMetric(float64(n), "points")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		require.NoError(b, domain.FFT(vec), "plonk2 FFT should succeed")
		require.NoError(b, dev.Sync(), "plonk2 FFT sync should succeed")
	}
}

func benchPlonk2CosetFFTRaw(
	b *testing.B,
	dev *gpu.Device,
	curve Curve,
	spec FFTDomainSpec,
	raw []uint64,
	generator []uint64,
	n int,
) {
	b.Helper()
	info := mustCurveInfo(b, curve)
	domain, err := NewFFTDomain(dev, spec)
	require.NoError(b, err, "creating plonk2 coset FFT domain should succeed")
	defer domain.Free()
	vec, err := NewFrVector(dev, curve, n)
	require.NoError(b, err, "allocating plonk2 coset FFT vector should succeed")
	defer vec.Free()
	require.NoError(b, vec.CopyFromHostRaw(raw), "copying plonk2 coset FFT input should succeed")
	require.NoError(b, dev.Sync(), "device sync before benchmark should succeed")

	b.SetBytes(int64(n) * int64(info.ScalarLimbs*8))
	b.ReportMetric(float64(n), "points")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		require.NoError(b, domain.CosetFFT(vec, generator), "plonk2 coset FFT should succeed")
		require.NoError(b, dev.Sync(), "plonk2 coset FFT sync should succeed")
	}
}

func benchOldPlonkFFTForwardBLS12377Size(b *testing.B, dev *gpu.Device, n int) {
	input := sequentialBLS12377FFTInput(n)
	domain, err := oldplonk.NewFFTDomain(dev, n)
	require.NoError(b, err, "creating gpu/plonk FFT domain should succeed")
	defer domain.Close()
	vec, err := oldplonk.NewFrVector(dev, n)
	require.NoError(b, err, "allocating gpu/plonk FFT vector should succeed")
	defer vec.Free()
	vec.CopyFromHost(input)
	require.NoError(b, dev.Sync(), "device sync before benchmark should succeed")

	b.SetBytes(int64(n) * int64(blsfr.Bytes))
	b.ReportMetric(float64(n), "points")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		domain.FFT(vec)
		require.NoError(b, dev.Sync(), "gpu/plonk FFT sync should succeed")
	}
	runtime.KeepAlive(input)
}

func benchOldPlonkCosetFFTBLS12377Size(b *testing.B, dev *gpu.Device, n int) {
	input := sequentialBLS12377FFTInput(n)
	domain, err := oldplonk.NewFFTDomain(dev, n)
	require.NoError(b, err, "creating gpu/plonk coset FFT domain should succeed")
	defer domain.Close()
	vec, err := oldplonk.NewFrVector(dev, n)
	require.NoError(b, err, "allocating gpu/plonk coset FFT vector should succeed")
	defer vec.Free()
	vec.CopyFromHost(input)
	cosetGen := blsfft.NewDomain(uint64(n)).FrMultiplicativeGen
	require.NoError(b, dev.Sync(), "device sync before benchmark should succeed")

	b.SetBytes(int64(n) * int64(blsfr.Bytes))
	b.ReportMetric(float64(n), "points")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		domain.CosetFFT(vec, cosetGen)
		require.NoError(b, dev.Sync(), "gpu/plonk coset FFT sync should succeed")
	}
	runtime.KeepAlive(input)
}

func fftBenchSizes(tb testing.TB, envName string) []int {
	tb.Helper()
	raw := strings.TrimSpace(os.Getenv(envName))
	if raw == "" {
		return []int{1 << 20, 1 << 22, 1 << 24, 1 << 25}
	}

	parts := strings.Split(raw, ",")
	sizes := make([]int, 0, len(parts))
	for _, part := range parts {
		n, err := parseBenchSize(strings.TrimSpace(part))
		require.NoError(tb, err, "parsing benchmark size %q should succeed", part)
		require.Positive(tb, n, "benchmark size should be positive")
		require.True(tb, isPowerOfTwo(n), "FFT benchmark size %q must be a power of two", part)
		sizes = append(sizes, n)
	}
	return sizes
}

func sequentialBN254FFTInput(n int) []bnfr.Element {
	out := make([]bnfr.Element, n)
	for i := range out {
		out[i].SetUint64(uint64(i + 1))
	}
	return out
}

func sequentialBLS12377FFTInput(n int) []blsfr.Element {
	out := make([]blsfr.Element, n)
	for i := range out {
		out[i].SetUint64(uint64(i + 1))
	}
	return out
}

func sequentialBW6761FFTInput(n int) []bwfr.Element {
	out := make([]bwfr.Element, n)
	for i := range out {
		out[i].SetUint64(uint64(i + 1))
	}
	return out
}
