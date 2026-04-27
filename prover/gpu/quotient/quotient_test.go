//go:build cuda

package quotient

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/linea-monorepo/prover/gpu"
	gpuvortex "github.com/consensys/linea-monorepo/prover/gpu/vortex"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// TestGPUNTTCosetEval verifies the GPU NTT pipeline matches the CPU
// for the full IFFT → coset FFT sequence used in quotient computation.
//
// GPU convention:
//   BitReverse → FFTInverse → Scale(1/n) → coefficients
//   CopyFromDevice → CosetFFT(shift) → BitReverse → evaluations
//
// Note: GPU FFTInverse does NOT include 1/n normalization (unlike gnark-crypto).
func TestGPUNTTCosetEval(t *testing.T) {
	dev := gpu.GetDevice()
	if dev == nil {
		t.Skip("no GPU")
	}

	const n = 1024
	domain0 := fft.NewDomain(uint64(n), fft.WithCache())
	nttDom, _ := gpuvortex.NewGPUFFTDomain(dev, n)
	defer nttDom.Free()

	witness := make([]field.Element, n)
	for i := range witness {
		witness[i].SetUint64(uint64(i + 1))
	}

	shift := computeShift(uint64(n), 2, 0)
	cosetDomain := fft.NewDomain(uint64(n), fft.WithShift(shift), fft.WithCache())

	// CPU reference: IFFT(DIF) → FFT(DIT, OnCoset)
	cpuResult := make([]field.Element, n)
	copy(cpuResult, witness)
	domain0.FFTInverse(cpuResult, fft.DIF, fft.WithNbTasks(1))
	cosetDomain.FFT(cpuResult, fft.DIT, fft.OnCoset(), fft.WithNbTasks(1))

	// GPU: BitRev → IFFT → Scale(1/n) → D2D → CosetFFT → BitRev
	dVec, _ := gpuvortex.NewKBVector(dev, n)
	defer dVec.Free()
	dVec.CopyFromHost(witness)
	dVec.BitReverse()
	nttDom.FFTInverse(dVec)
	var nInv field.Element
	nInv.SetUint64(uint64(n))
	nInv.Inverse(&nInv)
	dVec.Scale(nInv)

	dEval, _ := gpuvortex.NewKBVector(dev, n)
	defer dEval.Free()
	dEval.CopyFromDevice(dVec)
	nttDom.CosetFFT(dEval, shift)
	dEval.BitReverse()

	gpuResult := make([]field.Element, n)
	dEval.CopyToHost(gpuResult)

	mismatches := 0
	for i := range cpuResult {
		if cpuResult[i] != gpuResult[i] {
			mismatches++
			if mismatches <= 3 {
				t.Errorf("[%d] cpu=%v gpu=%v", i, cpuResult[i], gpuResult[i])
			}
		}
	}
	if mismatches > 0 {
		t.Errorf("total mismatches: %d/%d", mismatches, n)
	} else {
		t.Log("GPU coset evaluation matches CPU: PASS")
	}
}
