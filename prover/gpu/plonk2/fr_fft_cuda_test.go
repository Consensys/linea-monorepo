//go:build cuda

package plonk2

import (
	"math/big"
	"math/rand"
	"testing"
	"unsafe"

	blsfr "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	blsfft "github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
	bnfr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	bnfft "github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
	bwfr "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	bwfft "github.com/consensys/gnark-crypto/ecc/bw6-761/fr/fft"
	"github.com/stretchr/testify/require"

	"github.com/consensys/linea-monorepo/prover/gpu"
)

func TestFrVectorOps_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	t.Run("bn254", func(t *testing.T) { testFrVectorOpsBN254(t, dev, 1024) })
	t.Run("bls12-377", func(t *testing.T) { testFrVectorOpsBLS12377(t, dev, 1024) })
	t.Run("bw6-761", func(t *testing.T) { testFrVectorOpsBW6761(t, dev, 1024) })
}

func TestFFT_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	t.Run("bn254", func(t *testing.T) { testFFTBN254(t, dev, 1024) })
	t.Run("bls12-377", func(t *testing.T) { testFFTBLS12377(t, dev, 1024) })
	t.Run("bw6-761", func(t *testing.T) { testFFTBW6761(t, dev, 1024) })
}

func TestCosetFFT_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	t.Run("bn254", func(t *testing.T) { testCosetFFTBN254(t, dev, 1024) })
	t.Run("bls12-377", func(t *testing.T) { testCosetFFTBLS12377(t, dev, 1024) })
	t.Run("bw6-761", func(t *testing.T) { testCosetFFTBW6761(t, dev, 1024) })
}

func BenchmarkFrVectorMul_CUDA(b *testing.B) {
	dev, err := gpu.New()
	require.NoError(b, err, "creating CUDA device should succeed")
	defer dev.Close()

	benchFrMulBN254(b, dev, 1<<20)
	benchFrMulBLS12377(b, dev, 1<<20)
	benchFrMulBW6761(b, dev, 1<<20)
}

func BenchmarkFFTForward_CUDA(b *testing.B) {
	dev, err := gpu.New()
	require.NoError(b, err, "creating CUDA device should succeed")
	defer dev.Close()

	benchFFTBN254(b, dev, 1<<20)
	benchFFTBLS12377(b, dev, 1<<20)
	benchFFTBW6761(b, dev, 1<<20)
}

func BenchmarkCosetFFTForward_CUDA(b *testing.B) {
	dev, err := gpu.New()
	require.NoError(b, err, "creating CUDA device should succeed")
	defer dev.Close()

	benchCosetFFTBN254(b, dev, 1<<20)
	benchCosetFFTBLS12377(b, dev, 1<<20)
	benchCosetFFTBW6761(b, dev, 1<<20)
}

func testFrVectorOpsBN254(t *testing.T, dev *gpu.Device, n int) {
	a := deterministicBN254(n, 1)
	b := deterministicBN254(n, 2)

	gpuA, err := NewFrVector(dev, CurveBN254, n)
	require.NoError(t, err, "allocating input A should succeed")
	defer gpuA.Free()
	gpuB, err := NewFrVector(dev, CurveBN254, n)
	require.NoError(t, err, "allocating input B should succeed")
	defer gpuB.Free()
	result, err := NewFrVector(dev, CurveBN254, n)
	require.NoError(t, err, "allocating result should succeed")
	defer result.Free()

	require.NoError(t, gpuA.CopyFromHostRaw(rawBN254(a)), "copying A should succeed")
	require.NoError(t, gpuB.CopyFromHostRaw(rawBN254(b)), "copying B should succeed")

	expectedAdd := make([]bnfr.Element, n)
	expectedSub := make([]bnfr.Element, n)
	expectedMul := make([]bnfr.Element, n)
	expectedAddMul := make([]bnfr.Element, n)
	expectedScalarMul := make([]bnfr.Element, n)
	expectedAddScalarMul := make([]bnfr.Element, n)
	var scalar bnfr.Element
	scalar.SetUint64(7)
	for i := range a {
		expectedAdd[i].Add(&a[i], &b[i])
		expectedSub[i].Sub(&a[i], &b[i])
		expectedMul[i].Mul(&a[i], &b[i])
		expectedAddMul[i].Add(&a[i], &expectedMul[i])
		expectedScalarMul[i].Mul(&a[i], &scalar)
		var scaledB bnfr.Element
		scaledB.Mul(&b[i], &scalar)
		expectedAddScalarMul[i].Add(&a[i], &scaledB)
	}

	out := make([]bnfr.Element, n)
	copyVec, err := NewFrVector(dev, CurveBN254, n)
	require.NoError(t, err, "allocating copy vector should succeed")
	defer copyVec.Free()
	require.NoError(t, copyVec.CopyFromDevice(gpuA), "GPU d2d copy should succeed")
	require.NoError(t, copyVec.CopyToHostRaw(rawBN254(out)), "copying d2d output should succeed")
	requireBN254Equal(t, a, out, "copy d2d")

	require.NoError(t, result.Add(gpuA, gpuB), "GPU add should succeed")
	require.NoError(t, result.CopyToHostRaw(rawBN254(out)), "copying add output should succeed")
	requireBN254Equal(t, expectedAdd, out, "add")

	require.NoError(t, result.Sub(gpuA, gpuB), "GPU sub should succeed")
	require.NoError(t, result.CopyToHostRaw(rawBN254(out)), "copying sub output should succeed")
	requireBN254Equal(t, expectedSub, out, "sub")

	require.NoError(t, result.Mul(gpuA, gpuB), "GPU mul should succeed")
	require.NoError(t, result.CopyToHostRaw(rawBN254(out)), "copying mul output should succeed")
	requireBN254Equal(t, expectedMul, out, "mul")

	require.NoError(t, result.CopyFromHostRaw(rawBN254(a)), "resetting result should succeed")
	require.NoError(t, result.AddMul(gpuA, gpuB), "GPU addmul should succeed")
	require.NoError(t, result.CopyToHostRaw(rawBN254(out)), "copying addmul output should succeed")
	requireBN254Equal(t, expectedAddMul, out, "addmul")

	scalarRaw := cloneRaw(rawBN254([]bnfr.Element{scalar}))
	require.NoError(t, result.CopyFromHostRaw(rawBN254(a)), "resetting result should succeed")
	require.NoError(t, result.ScalarMulRaw(scalarRaw), "GPU scalar mul should succeed")
	require.NoError(t, result.CopyToHostRaw(rawBN254(out)), "copying scalar mul output should succeed")
	requireBN254Equal(t, expectedScalarMul, out, "scalar mul")

	require.NoError(t, result.CopyFromHostRaw(rawBN254(a)), "resetting result should succeed")
	require.NoError(t, result.AddScalarMulRaw(gpuB, scalarRaw), "GPU add scalar mul should succeed")
	require.NoError(t, result.CopyToHostRaw(rawBN254(out)), "copying add scalar mul output should succeed")
	requireBN254Equal(t, expectedAddScalarMul, out, "add scalar mul")

	invInput := nonZeroBN254(a)
	expectedInv := make([]bnfr.Element, n)
	for i := range invInput {
		expectedInv[i].Inverse(&invInput[i])
	}
	temp, err := NewFrVector(dev, CurveBN254, n)
	require.NoError(t, err, "allocating inverse temp should succeed")
	defer temp.Free()
	require.NoError(t, result.CopyFromHostRaw(rawBN254(invInput)), "copying inverse input should succeed")
	require.NoError(t, result.BatchInvert(temp), "GPU batch invert should succeed")
	require.NoError(t, result.CopyToHostRaw(rawBN254(out)), "copying inverse output should succeed")
	requireBN254Equal(t, expectedInv, out, "batch invert")

	require.NoError(t, result.SetZero(), "GPU zero should succeed")
	require.NoError(t, result.CopyToHostRaw(rawBN254(out)), "copying zero output should succeed")
	for i := range out {
		require.True(t, out[i].IsZero(), "zero output at index %d should be zero", i)
	}
}

func testFrVectorOpsBLS12377(t *testing.T, dev *gpu.Device, n int) {
	a := deterministicBLS12377(n, 3)
	b := deterministicBLS12377(n, 4)

	gpuA, err := NewFrVector(dev, CurveBLS12377, n)
	require.NoError(t, err, "allocating input A should succeed")
	defer gpuA.Free()
	gpuB, err := NewFrVector(dev, CurveBLS12377, n)
	require.NoError(t, err, "allocating input B should succeed")
	defer gpuB.Free()
	result, err := NewFrVector(dev, CurveBLS12377, n)
	require.NoError(t, err, "allocating result should succeed")
	defer result.Free()

	require.NoError(t, gpuA.CopyFromHostRaw(rawBLS12377(a)), "copying A should succeed")
	require.NoError(t, gpuB.CopyFromHostRaw(rawBLS12377(b)), "copying B should succeed")

	expectedAdd := make([]blsfr.Element, n)
	expectedSub := make([]blsfr.Element, n)
	expectedMul := make([]blsfr.Element, n)
	expectedAddMul := make([]blsfr.Element, n)
	expectedScalarMul := make([]blsfr.Element, n)
	expectedAddScalarMul := make([]blsfr.Element, n)
	var scalar blsfr.Element
	scalar.SetUint64(7)
	for i := range a {
		expectedAdd[i].Add(&a[i], &b[i])
		expectedSub[i].Sub(&a[i], &b[i])
		expectedMul[i].Mul(&a[i], &b[i])
		expectedAddMul[i].Add(&a[i], &expectedMul[i])
		expectedScalarMul[i].Mul(&a[i], &scalar)
		var scaledB blsfr.Element
		scaledB.Mul(&b[i], &scalar)
		expectedAddScalarMul[i].Add(&a[i], &scaledB)
	}

	out := make([]blsfr.Element, n)
	copyVec, err := NewFrVector(dev, CurveBLS12377, n)
	require.NoError(t, err, "allocating copy vector should succeed")
	defer copyVec.Free()
	require.NoError(t, copyVec.CopyFromDevice(gpuA), "GPU d2d copy should succeed")
	require.NoError(t, copyVec.CopyToHostRaw(rawBLS12377(out)), "copying d2d output should succeed")
	requireBLS12377Equal(t, a, out, "copy d2d")

	require.NoError(t, result.Add(gpuA, gpuB), "GPU add should succeed")
	require.NoError(t, result.CopyToHostRaw(rawBLS12377(out)), "copying add output should succeed")
	requireBLS12377Equal(t, expectedAdd, out, "add")

	require.NoError(t, result.Sub(gpuA, gpuB), "GPU sub should succeed")
	require.NoError(t, result.CopyToHostRaw(rawBLS12377(out)), "copying sub output should succeed")
	requireBLS12377Equal(t, expectedSub, out, "sub")

	require.NoError(t, result.Mul(gpuA, gpuB), "GPU mul should succeed")
	require.NoError(t, result.CopyToHostRaw(rawBLS12377(out)), "copying mul output should succeed")
	requireBLS12377Equal(t, expectedMul, out, "mul")

	require.NoError(t, result.CopyFromHostRaw(rawBLS12377(a)), "resetting result should succeed")
	require.NoError(t, result.AddMul(gpuA, gpuB), "GPU addmul should succeed")
	require.NoError(t, result.CopyToHostRaw(rawBLS12377(out)), "copying addmul output should succeed")
	requireBLS12377Equal(t, expectedAddMul, out, "addmul")

	scalarRaw := cloneRaw(rawBLS12377([]blsfr.Element{scalar}))
	require.NoError(t, result.CopyFromHostRaw(rawBLS12377(a)), "resetting result should succeed")
	require.NoError(t, result.ScalarMulRaw(scalarRaw), "GPU scalar mul should succeed")
	require.NoError(t, result.CopyToHostRaw(rawBLS12377(out)), "copying scalar mul output should succeed")
	requireBLS12377Equal(t, expectedScalarMul, out, "scalar mul")

	require.NoError(t, result.CopyFromHostRaw(rawBLS12377(a)), "resetting result should succeed")
	require.NoError(t, result.AddScalarMulRaw(gpuB, scalarRaw), "GPU add scalar mul should succeed")
	require.NoError(t, result.CopyToHostRaw(rawBLS12377(out)), "copying add scalar mul output should succeed")
	requireBLS12377Equal(t, expectedAddScalarMul, out, "add scalar mul")

	invInput := nonZeroBLS12377(a)
	expectedInv := make([]blsfr.Element, n)
	for i := range invInput {
		expectedInv[i].Inverse(&invInput[i])
	}
	temp, err := NewFrVector(dev, CurveBLS12377, n)
	require.NoError(t, err, "allocating inverse temp should succeed")
	defer temp.Free()
	require.NoError(t, result.CopyFromHostRaw(rawBLS12377(invInput)), "copying inverse input should succeed")
	require.NoError(t, result.BatchInvert(temp), "GPU batch invert should succeed")
	require.NoError(t, result.CopyToHostRaw(rawBLS12377(out)), "copying inverse output should succeed")
	requireBLS12377Equal(t, expectedInv, out, "batch invert")
}

func testFrVectorOpsBW6761(t *testing.T, dev *gpu.Device, n int) {
	a := deterministicBW6761(n, 5)
	b := deterministicBW6761(n, 6)

	gpuA, err := NewFrVector(dev, CurveBW6761, n)
	require.NoError(t, err, "allocating input A should succeed")
	defer gpuA.Free()
	gpuB, err := NewFrVector(dev, CurveBW6761, n)
	require.NoError(t, err, "allocating input B should succeed")
	defer gpuB.Free()
	result, err := NewFrVector(dev, CurveBW6761, n)
	require.NoError(t, err, "allocating result should succeed")
	defer result.Free()

	require.NoError(t, gpuA.CopyFromHostRaw(rawBW6761(a)), "copying A should succeed")
	require.NoError(t, gpuB.CopyFromHostRaw(rawBW6761(b)), "copying B should succeed")

	expectedAdd := make([]bwfr.Element, n)
	expectedSub := make([]bwfr.Element, n)
	expectedMul := make([]bwfr.Element, n)
	expectedAddMul := make([]bwfr.Element, n)
	expectedScalarMul := make([]bwfr.Element, n)
	expectedAddScalarMul := make([]bwfr.Element, n)
	var scalar bwfr.Element
	scalar.SetUint64(7)
	for i := range a {
		expectedAdd[i].Add(&a[i], &b[i])
		expectedSub[i].Sub(&a[i], &b[i])
		expectedMul[i].Mul(&a[i], &b[i])
		expectedAddMul[i].Add(&a[i], &expectedMul[i])
		expectedScalarMul[i].Mul(&a[i], &scalar)
		var scaledB bwfr.Element
		scaledB.Mul(&b[i], &scalar)
		expectedAddScalarMul[i].Add(&a[i], &scaledB)
	}

	out := make([]bwfr.Element, n)
	copyVec, err := NewFrVector(dev, CurveBW6761, n)
	require.NoError(t, err, "allocating copy vector should succeed")
	defer copyVec.Free()
	require.NoError(t, copyVec.CopyFromDevice(gpuA), "GPU d2d copy should succeed")
	require.NoError(t, copyVec.CopyToHostRaw(rawBW6761(out)), "copying d2d output should succeed")
	requireBW6761Equal(t, a, out, "copy d2d")

	require.NoError(t, result.Add(gpuA, gpuB), "GPU add should succeed")
	require.NoError(t, result.CopyToHostRaw(rawBW6761(out)), "copying add output should succeed")
	requireBW6761Equal(t, expectedAdd, out, "add")

	require.NoError(t, result.Sub(gpuA, gpuB), "GPU sub should succeed")
	require.NoError(t, result.CopyToHostRaw(rawBW6761(out)), "copying sub output should succeed")
	requireBW6761Equal(t, expectedSub, out, "sub")

	require.NoError(t, result.Mul(gpuA, gpuB), "GPU mul should succeed")
	require.NoError(t, result.CopyToHostRaw(rawBW6761(out)), "copying mul output should succeed")
	requireBW6761Equal(t, expectedMul, out, "mul")

	require.NoError(t, result.CopyFromHostRaw(rawBW6761(a)), "resetting result should succeed")
	require.NoError(t, result.AddMul(gpuA, gpuB), "GPU addmul should succeed")
	require.NoError(t, result.CopyToHostRaw(rawBW6761(out)), "copying addmul output should succeed")
	requireBW6761Equal(t, expectedAddMul, out, "addmul")

	scalarRaw := cloneRaw(rawBW6761([]bwfr.Element{scalar}))
	require.NoError(t, result.CopyFromHostRaw(rawBW6761(a)), "resetting result should succeed")
	require.NoError(t, result.ScalarMulRaw(scalarRaw), "GPU scalar mul should succeed")
	require.NoError(t, result.CopyToHostRaw(rawBW6761(out)), "copying scalar mul output should succeed")
	requireBW6761Equal(t, expectedScalarMul, out, "scalar mul")

	require.NoError(t, result.CopyFromHostRaw(rawBW6761(a)), "resetting result should succeed")
	require.NoError(t, result.AddScalarMulRaw(gpuB, scalarRaw), "GPU add scalar mul should succeed")
	require.NoError(t, result.CopyToHostRaw(rawBW6761(out)), "copying add scalar mul output should succeed")
	requireBW6761Equal(t, expectedAddScalarMul, out, "add scalar mul")

	invInput := nonZeroBW6761(a)
	expectedInv := make([]bwfr.Element, n)
	for i := range invInput {
		expectedInv[i].Inverse(&invInput[i])
	}
	temp, err := NewFrVector(dev, CurveBW6761, n)
	require.NoError(t, err, "allocating inverse temp should succeed")
	defer temp.Free()
	require.NoError(t, result.CopyFromHostRaw(rawBW6761(invInput)), "copying inverse input should succeed")
	require.NoError(t, result.BatchInvert(temp), "GPU batch invert should succeed")
	require.NoError(t, result.CopyToHostRaw(rawBW6761(out)), "copying inverse output should succeed")
	requireBW6761Equal(t, expectedInv, out, "batch invert")
}

func testFFTBN254(t *testing.T, dev *gpu.Device, n int) {
	input := deterministicBN254(n, 11)
	cpu := append([]bnfr.Element(nil), input...)
	domain := bnfft.NewDomain(uint64(n))
	domain.FFT(cpu, bnfft.DIF)

	gpuDomain, err := NewFFTDomain(dev, fftSpecBN254(n))
	require.NoError(t, err, "creating FFT domain should succeed")
	defer gpuDomain.Free()
	vec, err := NewFrVector(dev, CurveBN254, n)
	require.NoError(t, err, "allocating FFT vector should succeed")
	defer vec.Free()

	require.NoError(t, vec.CopyFromHostRaw(rawBN254(input)), "copying FFT input should succeed")
	require.NoError(t, gpuDomain.FFT(vec), "GPU FFT should succeed")
	out := make([]bnfr.Element, n)
	require.NoError(t, vec.CopyToHostRaw(rawBN254(out)), "copying FFT output should succeed")
	requireBN254Equal(t, cpu, out, "fft")

	require.NoError(t, gpuDomain.FFTInverse(vec), "GPU inverse FFT should succeed")
	require.NoError(t, vec.CopyToHostRaw(rawBN254(out)), "copying inverse FFT output should succeed")
	requireBN254Equal(t, input, out, "roundtrip")
}

func testFFTBLS12377(t *testing.T, dev *gpu.Device, n int) {
	input := deterministicBLS12377(n, 12)
	cpu := append([]blsfr.Element(nil), input...)
	domain := blsfft.NewDomain(uint64(n))
	domain.FFT(cpu, blsfft.DIF)

	gpuDomain, err := NewFFTDomain(dev, fftSpecBLS12377(n))
	require.NoError(t, err, "creating FFT domain should succeed")
	defer gpuDomain.Free()
	vec, err := NewFrVector(dev, CurveBLS12377, n)
	require.NoError(t, err, "allocating FFT vector should succeed")
	defer vec.Free()

	require.NoError(t, vec.CopyFromHostRaw(rawBLS12377(input)), "copying FFT input should succeed")
	require.NoError(t, gpuDomain.FFT(vec), "GPU FFT should succeed")
	out := make([]blsfr.Element, n)
	require.NoError(t, vec.CopyToHostRaw(rawBLS12377(out)), "copying FFT output should succeed")
	requireBLS12377Equal(t, cpu, out, "fft")

	require.NoError(t, gpuDomain.FFTInverse(vec), "GPU inverse FFT should succeed")
	require.NoError(t, vec.CopyToHostRaw(rawBLS12377(out)), "copying inverse FFT output should succeed")
	requireBLS12377Equal(t, input, out, "roundtrip")
}

func testFFTBW6761(t *testing.T, dev *gpu.Device, n int) {
	input := deterministicBW6761(n, 13)
	cpu := append([]bwfr.Element(nil), input...)
	domain := bwfft.NewDomain(uint64(n))
	domain.FFT(cpu, bwfft.DIF)

	gpuDomain, err := NewFFTDomain(dev, fftSpecBW6761(n))
	require.NoError(t, err, "creating FFT domain should succeed")
	defer gpuDomain.Free()
	vec, err := NewFrVector(dev, CurveBW6761, n)
	require.NoError(t, err, "allocating FFT vector should succeed")
	defer vec.Free()

	require.NoError(t, vec.CopyFromHostRaw(rawBW6761(input)), "copying FFT input should succeed")
	require.NoError(t, gpuDomain.FFT(vec), "GPU FFT should succeed")
	out := make([]bwfr.Element, n)
	require.NoError(t, vec.CopyToHostRaw(rawBW6761(out)), "copying FFT output should succeed")
	requireBW6761Equal(t, cpu, out, "fft")

	require.NoError(t, gpuDomain.FFTInverse(vec), "GPU inverse FFT should succeed")
	require.NoError(t, vec.CopyToHostRaw(rawBW6761(out)), "copying inverse FFT output should succeed")
	requireBW6761Equal(t, input, out, "roundtrip")
}

func testCosetFFTBN254(t *testing.T, dev *gpu.Device, n int) {
	input := deterministicBN254(n, 41)
	domain := bnfft.NewDomain(uint64(n))
	cpu := append([]bnfr.Element(nil), input...)
	domain.FFT(cpu, bnfft.DIF, bnfft.OnCoset())
	bitReverseSlice(cpu)

	gpuDomain, err := NewFFTDomain(dev, fftSpecBN254(n))
	require.NoError(t, err, "creating FFT domain should succeed")
	defer gpuDomain.Free()
	vec, err := NewFrVector(dev, CurveBN254, n)
	require.NoError(t, err, "allocating coset FFT vector should succeed")
	defer vec.Free()

	require.NoError(t, vec.CopyFromHostRaw(rawBN254(input)), "copying coset input should succeed")
	generatorRaw := cloneRaw(rawBN254([]bnfr.Element{domain.FrMultiplicativeGen}))
	require.NoError(t, gpuDomain.CosetFFT(vec, generatorRaw))
	out := make([]bnfr.Element, n)
	require.NoError(t, vec.CopyToHostRaw(rawBN254(out)), "copying coset output should succeed")
	requireBN254Equal(t, cpu, out, "coset fft")

	var genInv bnfr.Element
	genInv.Inverse(&domain.FrMultiplicativeGen)
	genInvRaw := cloneRaw(rawBN254([]bnfr.Element{genInv}))
	require.NoError(t, gpuDomain.CosetFFTInverse(vec, genInvRaw))
	require.NoError(t, vec.CopyToHostRaw(rawBN254(out)), "copying inverse coset output should succeed")
	requireBN254Equal(t, input, out, "coset roundtrip")
}

func testCosetFFTBLS12377(t *testing.T, dev *gpu.Device, n int) {
	input := deterministicBLS12377(n, 42)
	domain := blsfft.NewDomain(uint64(n))
	cpu := append([]blsfr.Element(nil), input...)
	domain.FFT(cpu, blsfft.DIF, blsfft.OnCoset())
	bitReverseSlice(cpu)

	gpuDomain, err := NewFFTDomain(dev, fftSpecBLS12377(n))
	require.NoError(t, err, "creating FFT domain should succeed")
	defer gpuDomain.Free()
	vec, err := NewFrVector(dev, CurveBLS12377, n)
	require.NoError(t, err, "allocating coset FFT vector should succeed")
	defer vec.Free()

	require.NoError(t, vec.CopyFromHostRaw(rawBLS12377(input)), "copying coset input should succeed")
	generatorRaw := cloneRaw(rawBLS12377([]blsfr.Element{domain.FrMultiplicativeGen}))
	require.NoError(t, gpuDomain.CosetFFT(vec, generatorRaw))
	out := make([]blsfr.Element, n)
	require.NoError(t, vec.CopyToHostRaw(rawBLS12377(out)), "copying coset output should succeed")
	requireBLS12377Equal(t, cpu, out, "coset fft")

	var genInv blsfr.Element
	genInv.Inverse(&domain.FrMultiplicativeGen)
	genInvRaw := cloneRaw(rawBLS12377([]blsfr.Element{genInv}))
	require.NoError(t, gpuDomain.CosetFFTInverse(vec, genInvRaw))
	require.NoError(t, vec.CopyToHostRaw(rawBLS12377(out)), "copying inverse coset output should succeed")
	requireBLS12377Equal(t, input, out, "coset roundtrip")
}

func testCosetFFTBW6761(t *testing.T, dev *gpu.Device, n int) {
	input := deterministicBW6761(n, 43)
	domain := bwfft.NewDomain(uint64(n))
	cpu := append([]bwfr.Element(nil), input...)
	domain.FFT(cpu, bwfft.DIF, bwfft.OnCoset())
	bitReverseSlice(cpu)

	gpuDomain, err := NewFFTDomain(dev, fftSpecBW6761(n))
	require.NoError(t, err, "creating FFT domain should succeed")
	defer gpuDomain.Free()
	vec, err := NewFrVector(dev, CurveBW6761, n)
	require.NoError(t, err, "allocating coset FFT vector should succeed")
	defer vec.Free()

	require.NoError(t, vec.CopyFromHostRaw(rawBW6761(input)), "copying coset input should succeed")
	generatorRaw := cloneRaw(rawBW6761([]bwfr.Element{domain.FrMultiplicativeGen}))
	require.NoError(t, gpuDomain.CosetFFT(vec, generatorRaw))
	out := make([]bwfr.Element, n)
	require.NoError(t, vec.CopyToHostRaw(rawBW6761(out)), "copying coset output should succeed")
	requireBW6761Equal(t, cpu, out, "coset fft")

	var genInv bwfr.Element
	genInv.Inverse(&domain.FrMultiplicativeGen)
	genInvRaw := cloneRaw(rawBW6761([]bwfr.Element{genInv}))
	require.NoError(t, gpuDomain.CosetFFTInverse(vec, genInvRaw))
	require.NoError(t, vec.CopyToHostRaw(rawBW6761(out)), "copying inverse coset output should succeed")
	requireBW6761Equal(t, input, out, "coset roundtrip")
}

func benchFrMulBN254(b *testing.B, dev *gpu.Device, n int) {
	b.Run("bn254", func(b *testing.B) {
		a := deterministicBN254(n, 21)
		c := deterministicBN254(n, 22)
		gpuA, result := mustBenchMulVectors(b, dev, CurveBN254, n, rawBN254(a), rawBN254(c))
		defer gpuA.Free()
		defer result.Free()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			require.NoError(b, result.Mul(gpuA, result), "GPU mul should succeed")
			require.NoError(b, dev.Sync(), "GPU sync should succeed")
		}
	})
}

func benchFrMulBLS12377(b *testing.B, dev *gpu.Device, n int) {
	b.Run("bls12-377", func(b *testing.B) {
		a := deterministicBLS12377(n, 23)
		c := deterministicBLS12377(n, 24)
		gpuA, result := mustBenchMulVectors(b, dev, CurveBLS12377, n, rawBLS12377(a), rawBLS12377(c))
		defer gpuA.Free()
		defer result.Free()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			require.NoError(b, result.Mul(gpuA, result), "GPU mul should succeed")
			require.NoError(b, dev.Sync(), "GPU sync should succeed")
		}
	})
}

func benchFrMulBW6761(b *testing.B, dev *gpu.Device, n int) {
	b.Run("bw6-761", func(b *testing.B) {
		a := deterministicBW6761(n, 25)
		c := deterministicBW6761(n, 26)
		gpuA, result := mustBenchMulVectors(b, dev, CurveBW6761, n, rawBW6761(a), rawBW6761(c))
		defer gpuA.Free()
		defer result.Free()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			require.NoError(b, result.Mul(gpuA, result), "GPU mul should succeed")
			require.NoError(b, dev.Sync(), "GPU sync should succeed")
		}
	})
}

func benchFFTBN254(b *testing.B, dev *gpu.Device, n int) {
	b.Run("bn254", func(b *testing.B) {
		input := deterministicBN254(n, 31)
		benchFFT(b, dev, CurveBN254, fftSpecBN254(n), rawBN254(input), n)
	})
}

func benchFFTBLS12377(b *testing.B, dev *gpu.Device, n int) {
	b.Run("bls12-377", func(b *testing.B) {
		input := deterministicBLS12377(n, 32)
		benchFFT(b, dev, CurveBLS12377, fftSpecBLS12377(n), rawBLS12377(input), n)
	})
}

func benchFFTBW6761(b *testing.B, dev *gpu.Device, n int) {
	b.Run("bw6-761", func(b *testing.B) {
		input := deterministicBW6761(n, 33)
		benchFFT(b, dev, CurveBW6761, fftSpecBW6761(n), rawBW6761(input), n)
	})
}

func benchCosetFFTBN254(b *testing.B, dev *gpu.Device, n int) {
	b.Run("bn254", func(b *testing.B) {
		input := deterministicBN254(n, 51)
		domain := bnfft.NewDomain(uint64(n))
		generatorRaw := cloneRaw(rawBN254([]bnfr.Element{domain.FrMultiplicativeGen}))
		benchCosetFFT(
			b,
			dev,
			CurveBN254,
			fftSpecBN254(n),
			rawBN254(input),
			generatorRaw,
			n,
		)
	})
}

func benchCosetFFTBLS12377(b *testing.B, dev *gpu.Device, n int) {
	b.Run("bls12-377", func(b *testing.B) {
		input := deterministicBLS12377(n, 52)
		domain := blsfft.NewDomain(uint64(n))
		generatorRaw := cloneRaw(rawBLS12377([]blsfr.Element{domain.FrMultiplicativeGen}))
		benchCosetFFT(
			b,
			dev,
			CurveBLS12377,
			fftSpecBLS12377(n),
			rawBLS12377(input),
			generatorRaw,
			n,
		)
	})
}

func benchCosetFFTBW6761(b *testing.B, dev *gpu.Device, n int) {
	b.Run("bw6-761", func(b *testing.B) {
		input := deterministicBW6761(n, 53)
		domain := bwfft.NewDomain(uint64(n))
		generatorRaw := cloneRaw(rawBW6761([]bwfr.Element{domain.FrMultiplicativeGen}))
		benchCosetFFT(
			b,
			dev,
			CurveBW6761,
			fftSpecBW6761(n),
			rawBW6761(input),
			generatorRaw,
			n,
		)
	})
}

func mustBenchMulVectors(
	b *testing.B,
	dev *gpu.Device,
	curve Curve,
	n int,
	aRaw []uint64,
	resultRaw []uint64,
) (*FrVector, *FrVector) {
	b.Helper()
	gpuA, err := NewFrVector(dev, curve, n)
	require.NoError(b, err, "allocating input should succeed")
	result, err := NewFrVector(dev, curve, n)
	require.NoError(b, err, "allocating result should succeed")
	require.NoError(b, gpuA.CopyFromHostRaw(aRaw), "copying input should succeed")
	require.NoError(b, result.CopyFromHostRaw(resultRaw), "copying result should succeed")
	require.NoError(b, dev.Sync(), "syncing setup should succeed")
	return gpuA, result
}

func benchFFT(b *testing.B, dev *gpu.Device, curve Curve, spec FFTDomainSpec, raw []uint64, n int) {
	b.Helper()
	domain, err := NewFFTDomain(dev, spec)
	require.NoError(b, err, "creating FFT domain should succeed")
	defer domain.Free()
	vec, err := NewFrVector(dev, curve, n)
	require.NoError(b, err, "allocating FFT vector should succeed")
	defer vec.Free()
	require.NoError(b, vec.CopyFromHostRaw(raw), "copying FFT input should succeed")
	require.NoError(b, dev.Sync(), "syncing setup should succeed")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		require.NoError(b, domain.FFT(vec), "GPU FFT should succeed")
		require.NoError(b, dev.Sync(), "GPU sync should succeed")
	}
}

func benchCosetFFT(
	b *testing.B,
	dev *gpu.Device,
	curve Curve,
	spec FFTDomainSpec,
	raw []uint64,
	generator []uint64,
	n int,
) {
	b.Helper()
	domain, err := NewFFTDomain(dev, spec)
	require.NoError(b, err, "creating FFT domain should succeed")
	defer domain.Free()
	vec, err := NewFrVector(dev, curve, n)
	require.NoError(b, err, "allocating coset FFT vector should succeed")
	defer vec.Free()
	require.NoError(b, vec.CopyFromHostRaw(raw), "copying coset FFT input should succeed")
	require.NoError(b, dev.Sync(), "syncing setup should succeed")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		require.NoError(b, domain.CosetFFT(vec, generator), "GPU coset FFT should succeed")
		require.NoError(b, dev.Sync(), "GPU sync should succeed")
	}
}

func deterministicBN254(n int, seed int64) []bnfr.Element {
	out := make([]bnfr.Element, n)
	fillDeterministic(n, seed, bnfr.Modulus(), func(i int, v *big.Int) {
		out[i].SetBigInt(v)
	})
	return out
}

func deterministicBLS12377(n int, seed int64) []blsfr.Element {
	out := make([]blsfr.Element, n)
	fillDeterministic(n, seed, blsfr.Modulus(), func(i int, v *big.Int) {
		out[i].SetBigInt(v)
	})
	return out
}

func deterministicBW6761(n int, seed int64) []bwfr.Element {
	out := make([]bwfr.Element, n)
	fillDeterministic(n, seed, bwfr.Modulus(), func(i int, v *big.Int) {
		out[i].SetBigInt(v)
	})
	return out
}

func nonZeroBN254(in []bnfr.Element) []bnfr.Element {
	out := append([]bnfr.Element(nil), in...)
	for i := range out {
		if out[i].IsZero() {
			out[i].SetUint64(uint64(i + 1))
		}
	}
	return out
}

func nonZeroBLS12377(in []blsfr.Element) []blsfr.Element {
	out := append([]blsfr.Element(nil), in...)
	for i := range out {
		if out[i].IsZero() {
			out[i].SetUint64(uint64(i + 1))
		}
	}
	return out
}

func nonZeroBW6761(in []bwfr.Element) []bwfr.Element {
	out := append([]bwfr.Element(nil), in...)
	for i := range out {
		if out[i].IsZero() {
			out[i].SetUint64(uint64(i + 1))
		}
	}
	return out
}

func fillDeterministic(n int, seed int64, modulus *big.Int, set func(i int, v *big.Int)) {
	// #nosec G404 - deterministic test vectors are required for reproducibility.
	rng := rand.New(rand.NewSource(seed))
	for i := 0; i < n; i++ {
		set(i, new(big.Int).Rand(rng, modulus))
	}
	if n == 0 {
		return
	}
	set(0, new(big.Int))
	if n > 1 {
		set(1, big.NewInt(1))
	}
	if n > 2 {
		set(2, new(big.Int).Sub(modulus, big.NewInt(1)))
	}
	if n > 3 {
		set(3, new(big.Int).Sub(modulus, big.NewInt(2)))
	}
}

func fftSpecBN254(n int) FFTDomainSpec {
	domain := bnfft.NewDomain(uint64(n))
	fwd, inv := buildTwiddlesBN254(n, domain.Generator, domain.GeneratorInv)
	return FFTDomainSpec{
		Curve:           CurveBN254,
		Size:            n,
		ForwardTwiddles: cloneRaw(rawBN254(fwd)),
		InverseTwiddles: cloneRaw(rawBN254(inv)),
		CardinalityInv:  cloneRaw(rawBN254([]bnfr.Element{domain.CardinalityInv})),
	}
}

func fftSpecBLS12377(n int) FFTDomainSpec {
	domain := blsfft.NewDomain(uint64(n))
	fwd, inv := buildTwiddlesBLS12377(n, domain.Generator, domain.GeneratorInv)
	return FFTDomainSpec{
		Curve:           CurveBLS12377,
		Size:            n,
		ForwardTwiddles: cloneRaw(rawBLS12377(fwd)),
		InverseTwiddles: cloneRaw(rawBLS12377(inv)),
		CardinalityInv:  cloneRaw(rawBLS12377([]blsfr.Element{domain.CardinalityInv})),
	}
}

func fftSpecBW6761(n int) FFTDomainSpec {
	domain := bwfft.NewDomain(uint64(n))
	fwd, inv := buildTwiddlesBW6761(n, domain.Generator, domain.GeneratorInv)
	return FFTDomainSpec{
		Curve:           CurveBW6761,
		Size:            n,
		ForwardTwiddles: cloneRaw(rawBW6761(fwd)),
		InverseTwiddles: cloneRaw(rawBW6761(inv)),
		CardinalityInv:  cloneRaw(rawBW6761([]bwfr.Element{domain.CardinalityInv})),
	}
}

func buildTwiddlesBN254(n int, generator, generatorInv bnfr.Element) ([]bnfr.Element, []bnfr.Element) {
	fwd := make([]bnfr.Element, n/2)
	inv := make([]bnfr.Element, n/2)
	if len(fwd) == 0 {
		return fwd, inv
	}
	fwd[0].SetOne()
	inv[0].SetOne()
	for i := 1; i < len(fwd); i++ {
		fwd[i].Mul(&fwd[i-1], &generator)
		inv[i].Mul(&inv[i-1], &generatorInv)
	}
	return fwd, inv
}

func buildTwiddlesBLS12377(n int, generator, generatorInv blsfr.Element) ([]blsfr.Element, []blsfr.Element) {
	fwd := make([]blsfr.Element, n/2)
	inv := make([]blsfr.Element, n/2)
	if len(fwd) == 0 {
		return fwd, inv
	}
	fwd[0].SetOne()
	inv[0].SetOne()
	for i := 1; i < len(fwd); i++ {
		fwd[i].Mul(&fwd[i-1], &generator)
		inv[i].Mul(&inv[i-1], &generatorInv)
	}
	return fwd, inv
}

func buildTwiddlesBW6761(n int, generator, generatorInv bwfr.Element) ([]bwfr.Element, []bwfr.Element) {
	fwd := make([]bwfr.Element, n/2)
	inv := make([]bwfr.Element, n/2)
	if len(fwd) == 0 {
		return fwd, inv
	}
	fwd[0].SetOne()
	inv[0].SetOne()
	for i := 1; i < len(fwd); i++ {
		fwd[i].Mul(&fwd[i-1], &generator)
		inv[i].Mul(&inv[i-1], &generatorInv)
	}
	return fwd, inv
}

func rawBN254(v []bnfr.Element) []uint64 {
	if len(v) == 0 {
		return nil
	}
	return unsafe.Slice((*uint64)(unsafe.Pointer(&v[0])), len(v)*4)
}

func rawBLS12377(v []blsfr.Element) []uint64 {
	if len(v) == 0 {
		return nil
	}
	return unsafe.Slice((*uint64)(unsafe.Pointer(&v[0])), len(v)*4)
}

func rawBW6761(v []bwfr.Element) []uint64 {
	if len(v) == 0 {
		return nil
	}
	return unsafe.Slice((*uint64)(unsafe.Pointer(&v[0])), len(v)*6)
}

func cloneRaw(v []uint64) []uint64 {
	return append([]uint64(nil), v...)
}

func bitReverseSlice[T any](v []T) {
	n := len(v)
	logN := 0
	for (1 << logN) < n {
		logN++
	}
	for i := range v {
		j := reverseBits(i, logN)
		if j > i {
			v[i], v[j] = v[j], v[i]
		}
	}
}

func reverseBits(x, width int) int {
	var y int
	for i := 0; i < width; i++ {
		y = (y << 1) | (x & 1)
		x >>= 1
	}
	return y
}

func requireBN254Equal(t *testing.T, want, got []bnfr.Element, label string) {
	t.Helper()
	require.Len(t, got, len(want), "%s output length should match", label)
	for i := range want {
		require.True(t, want[i].Equal(&got[i]), "%s mismatch at index %d", label, i)
	}
}

func requireBLS12377Equal(t *testing.T, want, got []blsfr.Element, label string) {
	t.Helper()
	require.Len(t, got, len(want), "%s output length should match", label)
	for i := range want {
		require.True(t, want[i].Equal(&got[i]), "%s mismatch at index %d", label, i)
	}
}

func requireBW6761Equal(t *testing.T, want, got []bwfr.Element, label string) {
	t.Helper()
	require.Len(t, got, len(want), "%s output length should match", label)
	for i := range want {
		require.True(t, want[i].Equal(&got[i]), "%s mismatch at index %d", label, i)
	}
}
