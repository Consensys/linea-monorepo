// Multi-GPU correctness + throughput benchmark for the device-resident
// commit path (gpu.CommitSIS / GPUVortex.Commit).
//
// Why this exists
// ───────────────
// The drop-in CommitMerkleWithSIS is slower than the AVX-512 CPU path at
// large sizes — its full D2H of the encoded matrix and Go-side
// reconstruction of []SmartVector wipes out the GPU win (see the breakdown
// in worklog_gpu_prover.md "Step 3 — vortex bottleneck" section).
//
// gv.Commit (used here) keeps the encoded matrix on device and returns a
// *CommitState; downstream LinComb / ExtractColumns operate on-device. This
// is the right path for prover integration.
//
// What this measures
// ──────────────────
//  1. BenchmarkCommitGPUResident_Large — single-GPU baseline at production
//     size (1<<19 cols × 1<<11 rows ≈ 1 B input cells).
//  2. BenchmarkCommitGPUResident_2GPU — two segment goroutines, one per
//     GPU, each committing an independent matrix concurrently. Validates
//     that DeviceCount=2 + per-thread CurrentDevice is correctly threaded
//     through getOrCreateGPUVortex (cache key must include deviceID).
//
// Read with the older BenchmarkCommitMerkle_Large to compare:
//   - drop-in:           1854 ms/op (slower than CPU)
//   - resident single:    258 ms/op (4.7× over CPU)
//   - resident 2-GPU:    expect ~258/2 = 130 ms/op of wall clock
//
//go:build cuda

package vortex

import (
	"math/rand/v2"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/consensys/gnark-crypto/field/koalabear/sis"
	"github.com/consensys/linea-monorepo/prover/crypto/vortex/vortex_koalabear"
	"github.com/consensys/linea-monorepo/prover/gpu"
)

// vortexKoalaParamsForBench returns the CPU-side params used by the
// protocol's actual vortex_koalabear path, matching the GPU-side params.
func vortexKoalaParamsForBench(nCols, nRows, rate int) vortex_koalabear.Params {
	return vortex_koalabear.NewParams(rate, nCols, nRows, 9, 16)
}

// BenchmarkCommitGPUResident_Large reproduces BenchmarkCommitMerkle_Large
// but lives next to the multi-GPU bench for direct comparison.
//
// Production size: 1<<19 cols × 1<<11 rows, rate=2 ≈ 4 GB encoded matrix.
func BenchmarkCommitGPUResident_Large(b *testing.B) {
	benchCommitResidentSingle(b, 1<<19, 1<<11, 2)
}

// BenchmarkCommitGPUResident_MedLarge — 65k × 1024, rate=2.
func BenchmarkCommitGPUResident_MedLarge(b *testing.B) {
	benchCommitResidentSingle(b, 1<<16, 1<<10, 2)
}

func benchCommitResidentSingle(b *testing.B, nCols, nRows, rate int) {
	rng := rand.New(rand.NewChaCha8([32]byte{0x1A}))
	dev := gpu.GetDeviceN(0)
	if dev == nil {
		b.Skip("no GPU 0")
	}

	sisP, _ := sis.NewRSis(0, 9, 16, nRows)
	params, _ := NewParams(nCols, nRows, sisP, rate, min(256, nCols*rate/4))

	gv, err := NewGPUVortex(dev, params, nRows)
	if err != nil {
		b.Fatal(err)
	}
	defer gv.Free()

	m := randMatrixMixed(rng, nRows, nCols)

	// Warmup
	if _, _, err := gv.Commit(m); err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(nCols) * int64(nRows) * 4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := gv.Commit(m); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCommitGPUResident_2GPU runs two concurrent commits, one per
// GPU. With LIMITLESS_GPU_COUNT≥2 set, expects ~2× throughput vs the
// single-GPU bench at the same per-GPU input size.
//
// Each goroutine pins its OS thread, sets gpu.CurrentDevice, and uses
// initGPU() (which now consults CurrentDevice + per-deviceID cache).
func BenchmarkCommitGPUResident_2GPU(b *testing.B) {
	const (
		nCols = 1 << 19
		nRows = 1 << 11
		rate  = 2
	)

	if gpu.DeviceCount() < 2 {
		b.Skip("need LIMITLESS_GPU_COUNT>=2")
	}
	dev0 := gpu.GetDeviceN(0)
	dev1 := gpu.GetDeviceN(1)
	if dev0 == nil || dev1 == nil {
		b.Skip("could not init both GPUs")
	}

	rng0 := rand.New(rand.NewChaCha8([32]byte{0x2A}))
	rng1 := rand.New(rand.NewChaCha8([32]byte{0x2B}))

	sisP, _ := sis.NewRSis(0, 9, 16, nRows)
	params0, _ := NewParams(nCols, nRows, sisP, rate, 256)
	params1, _ := NewParams(nCols, nRows, sisP, rate, 256)

	gv0, err := NewGPUVortex(dev0, params0, nRows)
	if err != nil {
		b.Fatal(err)
	}
	defer gv0.Free()
	gv1, err := NewGPUVortex(dev1, params1, nRows)
	if err != nil {
		b.Fatal(err)
	}
	defer gv1.Free()

	m0 := randMatrixMixed(rng0, nRows, nCols)
	m1 := randMatrixMixed(rng1, nRows, nCols)

	// Warmup both pipelines.
	gv0.Commit(m0)
	gv1.Commit(m1)

	b.SetBytes(2 * int64(nCols) * int64(nRows) * 4) // two matrices per op
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			gpu.SetCurrentDevice(dev0)
			gpu.SetCurrentDeviceID(0)
			defer gpu.SetCurrentDevice(nil)
			if _, _, err := gv0.Commit(m0); err != nil {
				b.Error(err)
			}
		}()
		go func() {
			defer wg.Done()
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			gpu.SetCurrentDevice(dev1)
			gpu.SetCurrentDeviceID(1)
			defer gpu.SetCurrentDevice(nil)
			if _, _, err := gv1.Commit(m1); err != nil {
				b.Error(err)
			}
		}()
		wg.Wait()
	}
}

// BenchmarkCommitGPUResidentVsCPU compares the device-resident GPU commit
// against the protocol's actual AVX-512 vortex_koalabear path
// (params.CommitMerkleWithSIS), at MedLarge and Large.
//
// Reports a real speedup_x metric tied to the production CPU baseline.
func BenchmarkCommitGPUResidentVsCPU_MedLarge(b *testing.B) {
	benchCommitResidentVsCPU(b, 1<<16, 1<<10, 2)
}

func BenchmarkCommitGPUResidentVsCPU_Large(b *testing.B) {
	benchCommitResidentVsCPU(b, 1<<19, 1<<11, 2)
}

func benchCommitResidentVsCPU(b *testing.B, nCols, nRows, rate int) {
	dev := gpu.GetDeviceN(0)
	if dev == nil {
		b.Skip("no GPU 0")
	}
	rng := rand.New(rand.NewChaCha8([32]byte{0x3A}))

	// GPU side
	sisP, _ := sis.NewRSis(0, 9, 16, nRows)
	gpuParams, _ := NewParams(nCols, nRows, sisP, rate, 256)
	gv, err := NewGPUVortex(dev, gpuParams, nRows)
	if err != nil {
		b.Fatal(err)
	}
	defer gv.Free()
	mGPU := randMatrixMixed(rng, nRows, nCols)
	gv.Commit(mGPU) // warmup

	// CPU side: protocol's vortex_koalabear with same parameters.
	cpuParamsPkg := vortexKoalaParamsForBench(nCols, nRows, rate)
	mCPU := randSmartVectorMatrix(rand.New(rand.NewChaCha8([32]byte{0x3A})), nRows, nCols)

	var gpuTotal, cpuTotal time.Duration
	for i := 0; i < b.N; i++ {
		t := time.Now()
		gv.Commit(mGPU)
		gpuTotal += time.Since(t)

		t = time.Now()
		cpuParamsPkg.CommitMerkleWithSIS(mCPU)
		cpuTotal += time.Since(t)
	}
	if b.N > 0 {
		gpuAvg := gpuTotal / time.Duration(b.N)
		cpuAvg := cpuTotal / time.Duration(b.N)
		b.ReportMetric(float64(gpuAvg.Milliseconds()), "gpu_ms")
		b.ReportMetric(float64(cpuAvg.Milliseconds()), "cpu_ms")
		b.ReportMetric(float64(cpuAvg)/float64(gpuAvg), "speedup_x")
	}
}
