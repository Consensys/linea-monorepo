//go:build cuda

package plonk2

import "testing"

func TestShouldUseCPUFallback_Cutoffs_CUDA(t *testing.T) {
	t.Setenv("PLONK2_BN254_MSM_DISABLE_CPU_FALLBACK", "")
	t.Setenv("PLONK2_BW6_MSM_DISABLE_CPU_FALLBACK", "")

	if !shouldUseCPUFallback(CurveBN254, bn254CPUFallbackPointLimit-1) {
		t.Fatal("BN254 should use CPU fallback below the cutoff")
	}
	if shouldUseCPUFallback(CurveBN254, bn254CPUFallbackPointLimit) {
		t.Fatal("BN254 should use GPU at the cutoff")
	}
	if !shouldUseCPUFallback(CurveBW6761, bw6761CPUFallbackPointLimit-1) {
		t.Fatal("BW6-761 should use CPU fallback below the cutoff")
	}
	if shouldUseCPUFallback(CurveBW6761, bw6761CPUFallbackPointLimit) {
		t.Fatal("BW6-761 should use GPU at the cutoff")
	}
}
