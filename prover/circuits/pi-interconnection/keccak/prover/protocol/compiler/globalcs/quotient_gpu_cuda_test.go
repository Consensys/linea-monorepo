//go:build cuda

package globalcs

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/gpu"
	"github.com/stretchr/testify/require"
)

func TestGPUQuotientReevalCosetMatchesCPU(t *testing.T) {
	if gpu.PhysicalDeviceCount() == 0 {
		t.Skip("requires a visible CUDA device")
	}
	t.Setenv(envPIQuotientGPUReeval, "1")

	const (
		domainSize = gpuQuotientReevalMinDomain
		numRoots   = 5
	)
	shift := computeShift(domainSize, 4, 1)
	inputs := make([][]field.Element, numRoots)
	outputs := make([][]field.Element, numRoots)
	expected := make([][]field.Element, numRoots)

	for i := range inputs {
		inputs[i] = make([]field.Element, domainSize)
		outputs[i] = make([]field.Element, domainSize)
		expected[i] = make([]field.Element, domainSize)
		for j := range inputs[i] {
			inputs[i][j].SetUint64(uint64(17 + i*domainSize + j*3))
		}
		copy(expected[i], inputs[i])
	}

	require.True(
		t,
		tryGPUQuotientReevalCoset(domainSize, shift, inputs, outputs),
		"GPU quotient reevaluation should run when enabled",
	)

	cpuDomain := fft.NewDomain(domainSize, fft.WithCache(), fft.WithShift(shift))
	for i := range expected {
		cpuDomain.FFT(expected[i], fft.DIT, fft.OnCoset(), fft.WithNbTasks(2))
		require.Equal(t, expected[i], outputs[i], "GPU coset reevaluation should match CPU for root %d", i)
	}
}
