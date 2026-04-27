//go:build cuda

package plonk_test

import (
	"fmt"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/linea-monorepo/prover/gpu"
	"github.com/consensys/linea-monorepo/prover/gpu/plonk"
	"github.com/stretchr/testify/require"
)

// TestDeviceLifecycle tests device creation, sync, and close.
func TestDeviceLifecycle(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	assert.NoError(dev.Sync(), "Sync failed")
}

// TestDeviceWithID tests creating a device with a specific GPU ID.
func TestDeviceWithID(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New(gpu.WithDeviceID(0))
	assert.NoError(err, "New(WithDeviceID(0)) failed")
	defer dev.Close()

	assert.NoError(dev.Sync(), "Sync failed")
}

// TestFrVectorRoundtrip tests copying data to GPU and back.
func TestFrVectorRoundtrip(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	const n = 1024
	src := randomFrVector(n)

	v, err := plonk.NewFrVector(dev, n)
	assert.NoError(err, "NewFrVector failed")
	defer v.Free()

	v.CopyFromHost(src)

	dst := make(fr.Vector, n)
	v.CopyToHost(dst)

	for i := 0; i < n; i++ {
		if !src[i].Equal(&dst[i]) {
			t.Errorf("mismatch at %d: got %v, want %v", i, dst[i], src[i])
		}
	}
}

// TestFrVectorMul tests element-wise multiplication against gnark-crypto.
func TestFrVectorMul(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	const n = 1024
	a := randomFrVector(n)
	b := randomFrVector(n)
	expected := make(fr.Vector, n)
	expected.Mul(a, b)

	gpuA, _ := plonk.NewFrVector(dev, n)
	gpuB, _ := plonk.NewFrVector(dev, n)
	gpuResult, _ := plonk.NewFrVector(dev, n)
	defer gpuA.Free()
	defer gpuB.Free()
	defer gpuResult.Free()

	gpuA.CopyFromHost(a)
	gpuB.CopyFromHost(b)
	gpuResult.Mul(gpuA, gpuB)
	dev.Sync()

	result := make(fr.Vector, n)
	gpuResult.CopyToHost(result)

	for i := 0; i < n; i++ {
		if !result[i].Equal(&expected[i]) {
			t.Errorf("mismatch at %d:\n  got:  %v\n  want: %v", i, result[i], expected[i])
		}
	}
}

// TestFrVectorAdd tests element-wise addition against gnark-crypto.
func TestFrVectorAdd(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	const n = 1024
	a := randomFrVector(n)
	b := randomFrVector(n)
	expected := make(fr.Vector, n)
	expected.Add(a, b)

	gpuA, _ := plonk.NewFrVector(dev, n)
	gpuB, _ := plonk.NewFrVector(dev, n)
	gpuResult, _ := plonk.NewFrVector(dev, n)
	defer gpuA.Free()
	defer gpuB.Free()
	defer gpuResult.Free()

	gpuA.CopyFromHost(a)
	gpuB.CopyFromHost(b)
	gpuResult.Add(gpuA, gpuB)
	dev.Sync()

	result := make(fr.Vector, n)
	gpuResult.CopyToHost(result)

	for i := 0; i < n; i++ {
		if !result[i].Equal(&expected[i]) {
			t.Errorf("mismatch at %d:\n  got:  %v\n  want: %v", i, result[i], expected[i])
		}
	}
}

// TestFrVectorSub tests element-wise subtraction against gnark-crypto.
func TestFrVectorSub(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	const n = 1024
	a := randomFrVector(n)
	b := randomFrVector(n)
	expected := make(fr.Vector, n)
	expected.Sub(a, b)

	gpuA, _ := plonk.NewFrVector(dev, n)
	gpuB, _ := plonk.NewFrVector(dev, n)
	gpuResult, _ := plonk.NewFrVector(dev, n)
	defer gpuA.Free()
	defer gpuB.Free()
	defer gpuResult.Free()

	gpuA.CopyFromHost(a)
	gpuB.CopyFromHost(b)
	gpuResult.Sub(gpuA, gpuB)
	dev.Sync()

	result := make(fr.Vector, n)
	gpuResult.CopyToHost(result)

	for i := 0; i < n; i++ {
		if !result[i].Equal(&expected[i]) {
			t.Errorf("mismatch at %d:\n  got:  %v\n  want: %v", i, result[i], expected[i])
		}
	}
}

// TestFrVectorLargeScale tests with 64K elements to stress the GPU.
func TestFrVectorLargeScale(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	const n = 1 << 16 // 65536 elements
	a := randomFrVector(n)
	b := randomFrVector(n)
	expected := make(fr.Vector, n)
	for i := 0; i < n; i++ {
		expected[i].Mul(&a[i], &b[i])
	}

	gpuA, _ := plonk.NewFrVector(dev, n)
	gpuB, _ := plonk.NewFrVector(dev, n)
	gpuResult, _ := plonk.NewFrVector(dev, n)
	defer gpuA.Free()
	defer gpuB.Free()
	defer gpuResult.Free()

	gpuA.CopyFromHost(a)
	gpuB.CopyFromHost(b)
	gpuResult.Mul(gpuA, gpuB)
	dev.Sync()

	result := make(fr.Vector, n)
	gpuResult.CopyToHost(result)

	for i := 0; i < 100; i++ {
		if !result[i].Equal(&expected[i]) {
			t.Errorf("mismatch at %d", i)
		}
	}
	for i := n - 100; i < n; i++ {
		if !result[i].Equal(&expected[i]) {
			t.Errorf("mismatch at %d", i)
		}
	}
}

// TestFrVectorEdgeCases tests with n=17 (non-aligned to block size).
func TestFrVectorEdgeCases(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	const n = 17
	a := randomFrVector(n)
	b := randomFrVector(n)
	expected := make(fr.Vector, n)
	for i := 0; i < n; i++ {
		expected[i].Add(&a[i], &b[i])
	}

	gpuA, _ := plonk.NewFrVector(dev, n)
	gpuB, _ := plonk.NewFrVector(dev, n)
	gpuResult, _ := plonk.NewFrVector(dev, n)
	defer gpuA.Free()
	defer gpuB.Free()
	defer gpuResult.Free()

	gpuA.CopyFromHost(a)
	gpuB.CopyFromHost(b)
	gpuResult.Add(gpuA, gpuB)
	dev.Sync()

	result := make(fr.Vector, n)
	gpuResult.CopyToHost(result)

	for i := 0; i < n; i++ {
		if !result[i].Equal(&expected[i]) {
			t.Errorf("mismatch at %d", i)
		}
	}
}

// TestFrVectorZero tests zero identity: 0 + b == b.
func TestFrVectorZero(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	const n = 100
	a := make(fr.Vector, n) // zero
	b := randomFrVector(n)

	gpuA, _ := plonk.NewFrVector(dev, n)
	gpuB, _ := plonk.NewFrVector(dev, n)
	gpuResult, _ := plonk.NewFrVector(dev, n)
	defer gpuA.Free()
	defer gpuB.Free()
	defer gpuResult.Free()

	gpuA.CopyFromHost(a)
	gpuB.CopyFromHost(b)
	gpuResult.Add(gpuA, gpuB)
	dev.Sync()

	result := make(fr.Vector, n)
	gpuResult.CopyToHost(result)

	for i := 0; i < n; i++ {
		if !result[i].Equal(&b[i]) {
			t.Errorf("0 + b != b at index %d", i)
		}
	}
}

// TestFrVectorSizeMismatchPanics tests that mismatched sizes panic.
func TestFrVectorSizeMismatchPanics(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	a, _ := plonk.NewFrVector(dev, 10)
	b, _ := plonk.NewFrVector(dev, 20)
	result, _ := plonk.NewFrVector(dev, 10)
	defer a.Free()
	defer b.Free()
	defer result.Free()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic on size mismatch, got none")
		}
	}()

	result.Mul(a, b)
}

// TestFrVectorCopyFromPanics tests that CopyFrom panics on size mismatch.
func TestFrVectorCopyFromPanics(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	v, _ := plonk.NewFrVector(dev, 10)
	defer v.Free()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic on CopyFrom size mismatch, got none")
		}
	}()

	src := make(fr.Vector, 20) // wrong size
	v.CopyFromHost(src)
}

// BenchmarkFrVectorMul benchmarks GPU multiplication at various sizes.
func BenchmarkFrVectorMul(b *testing.B) {
	assert := require.New(b)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	sizes := []int{1 << 10, 1 << 14, 1 << 18, 1 << 20}

	for _, n := range sizes {
		b.Run(formatSize(n), func(b *testing.B) {
			a := randomFrVector(n)
			bVec := randomFrVector(n)

			gpuA, _ := plonk.NewFrVector(dev, n)
			gpuB, _ := plonk.NewFrVector(dev, n)
			gpuResult, _ := plonk.NewFrVector(dev, n)
			defer gpuA.Free()
			defer gpuB.Free()
			defer gpuResult.Free()

			gpuA.CopyFromHost(a)
			gpuB.CopyFromHost(bVec)
			dev.Sync()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				gpuResult.Mul(gpuA, gpuB)
				dev.Sync()
			}
		})
	}
}

// BenchmarkFrVectorMulCPU benchmarks CPU multiplication using gnark-crypto.
func BenchmarkFrVectorMulCPU(b *testing.B) {
	sizes := []int{1 << 10, 1 << 14, 1 << 18, 1 << 20}

	for _, n := range sizes {
		b.Run(formatSize(n), func(b *testing.B) {
			a := randomFrVector(n)
			bVec := randomFrVector(n)
			result := make(fr.Vector, n)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result.Mul(a, bVec)
			}
		})
	}
}

// =============================================================================
// FrVector Primitive Operations
// =============================================================================

func TestScaleByPowers(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	sizes := []int{1, 17, 256, 1024, 1 << 15}
	for _, n := range sizes {
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			assert := require.New(t)

			data := randomFrVector(n)

			// CPU reference: v[i] *= g^i
			var g fr.Element
			g.SetRandom()
			cpuData := make(fr.Vector, n)
			copy(cpuData, data)
			var power fr.Element
			power.SetOne()
			for i := 0; i < n; i++ {
				cpuData[i].Mul(&cpuData[i], &power)
				power.Mul(&power, &g)
			}

			// GPU
			gpuVec, err := plonk.NewFrVector(dev, n)
			assert.NoError(err, "NewFrVector failed")
			defer gpuVec.Free()

			gpuVec.CopyFromHost(data)
			gpuVec.ScaleByPowers(g)
			dev.Sync()

			result := make(fr.Vector, n)
			gpuVec.CopyToHost(result)

			for i := 0; i < n; i++ {
				if !result[i].Equal(&cpuData[i]) {
					t.Errorf("mismatch at %d:\n  gpu: %v\n  cpu: %v", i, result[i], cpuData[i])
					if i > 5 {
						t.Fatalf("too many mismatches")
					}
				}
			}
		})
	}
}

// TestScaleByPowersIdentity validates that g=1 leaves vector unchanged.
func TestScaleByPowersIdentity(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	const n = 256
	data := randomFrVector(n)

	gpuVec, _ := plonk.NewFrVector(dev, n)
	defer gpuVec.Free()

	gpuVec.CopyFromHost(data)
	var one fr.Element
	one.SetOne()
	gpuVec.ScaleByPowers(one)
	dev.Sync()

	result := make(fr.Vector, n)
	gpuVec.CopyToHost(result)

	for i := 0; i < n; i++ {
		if !result[i].Equal(&data[i]) {
			t.Errorf("g=1 changed element at %d", i)
		}
	}
}

// TestScalarMul validates v[i] *= c against CPU reference.
func TestScalarMul(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	sizes := []int{1, 17, 1024, 1 << 15}
	for _, n := range sizes {
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			data := randomFrVector(n)

			var c fr.Element
			c.SetRandom()

			// CPU reference
			cpuData := make(fr.Vector, n)
			for i := 0; i < n; i++ {
				cpuData[i].Mul(&data[i], &c)
			}

			// GPU
			gpuVec, _ := plonk.NewFrVector(dev, n)
			defer gpuVec.Free()

			gpuVec.CopyFromHost(data)
			gpuVec.ScalarMul(c)
			dev.Sync()

			result := make(fr.Vector, n)
			gpuVec.CopyToHost(result)

			for i := 0; i < n; i++ {
				if !result[i].Equal(&cpuData[i]) {
					t.Errorf("mismatch at %d:\n  gpu: %v\n  cpu: %v", i, result[i], cpuData[i])
					if i > 5 {
						t.Fatalf("too many mismatches")
					}
				}
			}
		})
	}
}

// TestCopyFromDevice validates GPU-to-GPU copy.
func TestCopyFromDevice(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	const n = 1024
	data := randomFrVector(n)

	src, _ := plonk.NewFrVector(dev, n)
	dst, _ := plonk.NewFrVector(dev, n)
	defer src.Free()
	defer dst.Free()

	src.CopyFromHost(data)
	dst.CopyFromDevice(src)
	dev.Sync()

	result := make(fr.Vector, n)
	dst.CopyToHost(result)

	for i := 0; i < n; i++ {
		if !result[i].Equal(&data[i]) {
			t.Errorf("D2D copy mismatch at %d", i)
		}
	}
}

// TestSetZero validates that SetZero produces all-zero elements.
func TestSetZero(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	const n = 1024
	// Fill with random data first
	data := randomFrVector(n)
	gpuVec, _ := plonk.NewFrVector(dev, n)
	defer gpuVec.Free()

	gpuVec.CopyFromHost(data)
	gpuVec.SetZero()
	dev.Sync()

	result := make(fr.Vector, n)
	gpuVec.CopyToHost(result)

	var zero fr.Element
	for i := 0; i < n; i++ {
		if !result[i].Equal(&zero) {
			t.Errorf("non-zero at %d: %v", i, result[i])
		}
	}
}

// TestAddMul validates v[i] += a[i] * b[i] against CPU reference.
func TestAddMul(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	sizes := []int{1, 17, 1024, 1 << 15}
	for _, n := range sizes {
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			v := randomFrVector(n)
			a := randomFrVector(n)
			b := randomFrVector(n)

			// CPU reference: v[i] += a[i] * b[i]
			cpuResult := make(fr.Vector, n)
			copy(cpuResult, v)
			for i := 0; i < n; i++ {
				var tmp fr.Element
				tmp.Mul(&a[i], &b[i])
				cpuResult[i].Add(&cpuResult[i], &tmp)
			}

			// GPU
			gpuV, _ := plonk.NewFrVector(dev, n)
			gpuA, _ := plonk.NewFrVector(dev, n)
			gpuB, _ := plonk.NewFrVector(dev, n)
			defer gpuV.Free()
			defer gpuA.Free()
			defer gpuB.Free()

			gpuV.CopyFromHost(v)
			gpuA.CopyFromHost(a)
			gpuB.CopyFromHost(b)
			gpuV.AddMul(gpuA, gpuB)
			dev.Sync()

			result := make(fr.Vector, n)
			gpuV.CopyToHost(result)

			for i := 0; i < n; i++ {
				if !result[i].Equal(&cpuResult[i]) {
					t.Errorf("mismatch at %d:\n  gpu: %v\n  cpu: %v", i, result[i], cpuResult[i])
					if i > 5 {
						t.Fatalf("too many mismatches")
					}
				}
			}
		})
	}
}

// TestAddMulAccumulate validates that multiple AddMul calls accumulate correctly.
func TestAddMulAccumulate(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	const n = 512
	a1 := randomFrVector(n)
	b1 := randomFrVector(n)
	a2 := randomFrVector(n)
	b2 := randomFrVector(n)

	// CPU: v = 0; v += a1*b1; v += a2*b2
	cpuResult := make(fr.Vector, n)
	for i := 0; i < n; i++ {
		var tmp fr.Element
		tmp.Mul(&a1[i], &b1[i])
		cpuResult[i].Add(&cpuResult[i], &tmp)
		tmp.Mul(&a2[i], &b2[i])
		cpuResult[i].Add(&cpuResult[i], &tmp)
	}

	// GPU
	gpuV, _ := plonk.NewFrVector(dev, n)
	gpuA1, _ := plonk.NewFrVector(dev, n)
	gpuB1, _ := plonk.NewFrVector(dev, n)
	gpuA2, _ := plonk.NewFrVector(dev, n)
	gpuB2, _ := plonk.NewFrVector(dev, n)
	defer gpuV.Free()
	defer gpuA1.Free()
	defer gpuB1.Free()
	defer gpuA2.Free()
	defer gpuB2.Free()

	gpuV.SetZero()
	gpuA1.CopyFromHost(a1)
	gpuB1.CopyFromHost(b1)
	gpuA2.CopyFromHost(a2)
	gpuB2.CopyFromHost(b2)

	gpuV.AddMul(gpuA1, gpuB1)
	gpuV.AddMul(gpuA2, gpuB2)
	dev.Sync()

	result := make(fr.Vector, n)
	gpuV.CopyToHost(result)

	for i := 0; i < n; i++ {
		if !result[i].Equal(&cpuResult[i]) {
			t.Errorf("accumulate mismatch at %d", i)
		}
	}
}

// TestAddScalarMulAccumulate verifies that multiple AddScalarMul calls accumulate.
func TestAddScalarMulAccumulate(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err)
	defer dev.Close()

	const n = 512
	a1 := randomFrVector(n)
	a2 := randomFrVector(n)
	var s1, s2 fr.Element
	s1.SetRandom()
	s2.SetRandom()

	// CPU: v = 0; v += a1*s1; v += a2*s2
	cpuResult := make(fr.Vector, n)
	for i := 0; i < n; i++ {
		var t1, t2 fr.Element
		t1.Mul(&a1[i], &s1)
		t2.Mul(&a2[i], &s2)
		cpuResult[i].Add(&t1, &t2)
	}

	// GPU
	gpuV, _ := plonk.NewFrVector(dev, n)
	gpuA1, _ := plonk.NewFrVector(dev, n)
	gpuA2, _ := plonk.NewFrVector(dev, n)
	defer gpuV.Free()
	defer gpuA1.Free()
	defer gpuA2.Free()

	gpuV.SetZero()
	gpuA1.CopyFromHost(a1)
	gpuA2.CopyFromHost(a2)
	gpuV.AddScalarMul(gpuA1, s1)
	gpuV.AddScalarMul(gpuA2, s2)
	dev.Sync()

	result := make(fr.Vector, n)
	gpuV.CopyToHost(result)

	for i := 0; i < n; i++ {
		if !result[i].Equal(&cpuResult[i]) {
			t.Errorf("mismatch at %d", i)
			if i > 5 {
				t.Fatalf("too many mismatches")
			}
		}
	}
}

// =============================================================================
// Batch Inversion Tests
// =============================================================================

// TestBatchInvert validates v[i] = 1/v[i] against element-wise CPU inversion.
func TestBatchInvert(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	sizes := []int{1, 2, 17, 256, 1024, 1 << 15}
	for _, n := range sizes {
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			data := randomFrVector(n)

			// CPU reference: element-wise inversion
			cpuResult := make(fr.Vector, n)
			for i := 0; i < n; i++ {
				cpuResult[i].Inverse(&data[i])
			}

			// GPU batch inversion
			gpuV, _ := plonk.NewFrVector(dev, n)
			gpuTemp, _ := plonk.NewFrVector(dev, n)
			defer gpuV.Free()
			defer gpuTemp.Free()

			gpuV.CopyFromHost(data)
			gpuV.BatchInvert(gpuTemp)
			dev.Sync()

			result := make(fr.Vector, n)
			gpuV.CopyToHost(result)

			for i := 0; i < n; i++ {
				if !result[i].Equal(&cpuResult[i]) {
					t.Errorf("mismatch at %d:\n  gpu: %v\n  cpu: %v", i, result[i], cpuResult[i])
					if i > 5 {
						t.Fatalf("too many mismatches")
					}
				}
			}
		})
	}
}

// TestBatchInvertVerify validates that v[i] * (1/v[i]) == 1 for all i.
func TestBatchInvertVerify(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	const n = 4096
	data := randomFrVector(n)

	gpuV, _ := plonk.NewFrVector(dev, n)
	gpuOrig, _ := plonk.NewFrVector(dev, n)
	gpuTemp, _ := plonk.NewFrVector(dev, n)
	gpuProd, _ := plonk.NewFrVector(dev, n)
	defer gpuV.Free()
	defer gpuOrig.Free()
	defer gpuTemp.Free()
	defer gpuProd.Free()

	gpuV.CopyFromHost(data)
	gpuOrig.CopyFromHost(data)
	gpuV.BatchInvert(gpuTemp)
	// Compute v * orig = should be all 1s
	gpuProd.Mul(gpuV, gpuOrig)
	dev.Sync()

	result := make(fr.Vector, n)
	gpuProd.CopyToHost(result)

	var one fr.Element
	one.SetOne()
	for i := 0; i < n; i++ {
		if !result[i].Equal(&one) {
			t.Errorf("v[%d] * inv(v[%d]) != 1: got %v", i, i, result[i])
			if i > 5 {
				t.Fatalf("too many mismatches")
			}
		}
	}
}

// TestBatchInvertDoubleInverse validates that inv(inv(v[i])) == v[i].
func TestBatchInvertDoubleInverse(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	const n = 1024
	data := randomFrVector(n)

	gpuV, _ := plonk.NewFrVector(dev, n)
	gpuTemp, _ := plonk.NewFrVector(dev, n)
	defer gpuV.Free()
	defer gpuTemp.Free()

	gpuV.CopyFromHost(data)
	gpuV.BatchInvert(gpuTemp) // v = 1/v
	gpuV.BatchInvert(gpuTemp) // v = 1/(1/v) = v
	dev.Sync()

	result := make(fr.Vector, n)
	gpuV.CopyToHost(result)

	for i := 0; i < n; i++ {
		if !result[i].Equal(&data[i]) {
			t.Errorf("double inverse mismatch at %d", i)
			if i > 5 {
				t.Fatalf("too many mismatches")
			}
		}
	}
}
