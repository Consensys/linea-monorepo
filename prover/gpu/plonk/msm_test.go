//go:build cuda

package plonk_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/consensys/gnark-crypto/ecc"
	bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fp"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/utils/unsafe"

	"github.com/consensys/linea-monorepo/prover/gpu"
	"github.com/consensys/linea-monorepo/prover/gpu/plonk"
	"github.com/stretchr/testify/require"
)

var msmReferenceResults = map[int]bls12377.G1Jac{
	1024: {
		X: fp.Element{0xb548b0258f4aed81, 0x9007340b11bdaf84, 0xec91fc52c273b824, 0xc094a534469024ee, 0x3267512fb7895482, 0x00484aed0655f301},
		Y: fp.Element{0xe98f91fcb12b22d9, 0x1fdf5435f552530c, 0xc07975b0fcd03f76, 0x939878e75d360289, 0x265d659246075aab, 0x00a8b9764003cd5b},
		Z: fp.Element{0xaa502685cafc1149, 0xfef06e6409f563c2, 0x10696b5db794f06a, 0x89bbad145b6841d2, 0xdcc408f6997e9d86, 0x012a40fdce8f713b},
	},
	32768: {
		X: fp.Element{0x87f27aa08714b2ba, 0x25d05ccd4d25bcdf, 0xddb2d2f140f180e7, 0x9145c3b9d800a5ce, 0xcbb4c06a953aef71, 0x00337025b7b8a431},
		Y: fp.Element{0x78d7b85e489bd358, 0x97d23e0dd308875a, 0xe4e6713351d93eaf, 0x0eeb8c64f4e093a3, 0x56f9d62a7970a211, 0x01a42c3381cea30b},
		Z: fp.Element{0xdc877f0061ca3782, 0x42f70df0de1e847d, 0x3efbf1560b51d205, 0x39465b1a8469f2e8, 0x0f9b170369c96eab, 0x011332edf7f3556b},
	},
	1048576: {
		X: fp.Element{0x2a1cc8a507e26268, 0x3b3355be018268ac, 0x736fd056ffc0de3a, 0xe11a880508f5422a, 0xf95a714df79abd7c, 0x015cb57d02cd8154},
		Y: fp.Element{0x1d4507ecbe03389d, 0x1f3a18ab25e1048d, 0x58da48dddca9640c, 0x4cc0e97d95d84c89, 0x43f4edfbf0ea07d6, 0x015d83f3019feef7},
		Z: fp.Element{0x71258d46109f7f3f, 0xb18c8afd64cf075e, 0x35ffd171da0f5ad0, 0xc037a5d5beaad748, 0x0628dd597dda6229, 0x0110e14c9a306ad1},
	},
	134217728: {
		X: fp.Element{0x9f59bc3cc30058a2, 0x844c6f662ba2d986, 0x0c119c366e75232a, 0x5400a11eec1ac2e5, 0x867690b7d4dfe755, 0x015b28b02786c3a2},
		Y: fp.Element{0x410652062082d943, 0xc7b59821e41597ef, 0x4a0aa8140cf1ecb7, 0x918573ffece506f2, 0xd0400d44410275d7, 0x00ceb8f23fc6a420},
		Z: fp.Element{0x492312ce1ee2e562, 0xb17e8946884e6c6f, 0x1d0535e3a91d9970, 0x5af07566db8a0b8d, 0x4527f723c3c81042, 0x017eba325c996b02},
	},
}

const srsRootDir = "/home/ubuntu/dev/go/src/github.com/consensys/linea-monorepo/prover/prover-assets/kzgsrs"

var testSRSStore *plonk.SRSStore

func getTestSRSStore(tb testing.TB) *plonk.SRSStore {
	tb.Helper()
	if testSRSStore != nil {
		return testSRSStore
	}
	store, err := plonk.NewSRSStore(srsRootDir)
	if err != nil {
		tb.Fatalf("NewSRSStore(%s) failed: %v", srsRootDir, err)
	}
	testSRSStore = store
	return store
}

func loadSRSTEPoints(tb testing.TB, n int) []plonk.G1TEPoint {
	tb.Helper()
	store := getTestSRSStore(tb)
	pts, err := store.LoadTEPoints(n, true)
	if err != nil {
		tb.Fatalf("LoadTEPoints(%d) failed: %v", n, err)
	}
	return pts
}

func loadSRSAffinePoints(tb testing.TB, n int) []bls12377.G1Affine {
	tb.Helper()
	store := getTestSRSStore(tb)
	pts, err := store.LoadPointsAffine(n, true)
	if err != nil {
		tb.Fatalf("LoadTEPointsAffine(%d) failed: %v", n, err)
	}
	return pts
}

func loadTestScalars(n int) []fr.Element {
	path := "/home/ubuntu/dev/go/src/github.com/consensys/linea-monorepo/prover/prover-assets/kzgsrs/random_scalars_134217728_bls12377.memdump"
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := unsafe.ReadMarker(f); err != nil {
		panic(err)
	}
	s, _, err := unsafe.ReadSlice[[]fr.Element](f, n)
	if err != nil {
		panic(err)
	}
	return s
}

func randomFrVector(n int) fr.Vector {
	v := make(fr.Vector, n)
	for i := range v {
		v[i].SetRandom()
	}
	return v
}

func formatSize(n int) string {
	if n >= 1<<20 {
		return fmt.Sprintf("%dM", n>>20)
	}
	return fmt.Sprintf("%dK", n>>10)
}

// TestSRSTEPointsRoundtrip verifies TE dump points match the original SRS points.
func TestSRSTEPointsRoundtrip(t *testing.T) {
	assert := require.New(t)

	store := getTestSRSStore(t)

	const n = 256
	tePoints, err := store.LoadTEPoints(n, false)
	if err != nil {
		t.Skipf("lagrange TE dump not found for n=%d (run TestSRSConvertAllToTE first): %v", n, err)
	}

	origAffine, err := store.LoadPointsAffine(n, false)
	assert.NoError(err, "LoadTEPointsAffine failed")

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	scalars := make([]fr.Element, n)
	for i := range scalars {
		scalars[i].SetRandom()
	}

	pts, err := plonk.PinG1TEPoints(tePoints)
	assert.NoError(err, "PinG1TEPoints failed")

	msm, err := plonk.NewG1MSM(dev, pts)
	assert.NoError(err, "NewG1MSM failed")
	defer msm.Close()

	results := msm.MultiExp(scalars)
	gpuResult := results[0]

	cpuResult := cpuMSM(origAffine, scalars)
	if !gpuResult.Equal(&cpuResult) {
		t.Error("TE dump MSM does not match CPU MSM from original SRS")
	}
}

// =============================================================================
// MSM Tests
// =============================================================================

// cpuMSM computes MSM using gnark-crypto's CPU implementation for reference.
// Used only for edge case tests with non-deterministic scalars.
func cpuMSM(points []bls12377.G1Affine, scalars []fr.Element) bls12377.G1Jac {
	var result bls12377.G1Jac
	result.MultiExp(points, scalars, ecc.MultiExpConfig{})
	return result
}

// TestMSMSmall tests GPU MSM at n=1024 against precomputed reference.
func TestMSMSmall(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	const n = 1024
	tePoints := loadSRSTEPoints(t, n)
	scalars := loadTestScalars(n)

	ref, ok := msmReferenceResults[n]
	if !ok {
		t.Fatalf("no precomputed reference for n=%d", n)
	}

	pts, err := plonk.PinG1TEPoints(tePoints)
	assert.NoError(err, "PinG1TEPoints failed")

	msm, err := plonk.NewG1MSM(dev, pts)
	assert.NoError(err, "NewG1MSM failed")
	defer msm.Close()

	results := msm.MultiExp(scalars)
	gpuResult := results[0]

	if !gpuResult.Equal(&ref) {
		t.Errorf("GPU MSM != precomputed reference for n=%d", n)
	}
}

// TestMSMMedium tests GPU MSM at n=32768 against precomputed reference.
func TestMSMMedium(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	const n = 1 << 15
	tePoints := loadSRSTEPoints(t, n)
	scalars := loadTestScalars(n)

	ref, ok := msmReferenceResults[n]
	if !ok {
		t.Fatalf("no precomputed reference for n=%d", n)
	}

	pts, err := plonk.PinG1TEPoints(tePoints)
	assert.NoError(err, "PinG1TEPoints failed")

	msm, err := plonk.NewG1MSM(dev, pts)
	assert.NoError(err, "NewG1MSM failed")
	defer msm.Close()

	results := msm.MultiExp(scalars)
	gpuResult := results[0]

	if !gpuResult.Equal(&ref) {
		t.Errorf("GPU MSM != precomputed reference for n=%d", n)
	}
}

// TestMSMEdgeCases tests GPU MSM with adversarial inputs (zero, single, same scalar).
// These use random scalars + CPU MSM since hardcoded references don't make sense.
func TestMSMEdgeCases(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	allTE := loadSRSTEPoints(t, 1024)
	allAffine := loadSRSAffinePoints(t, 1024)

	t.Run("ZeroScalars", func(t *testing.T) {
		assert := require.New(t)

		const n = 64
		scalars := make([]fr.Element, n)

		pts, err := plonk.PinG1TEPoints(allTE[:n])
		assert.NoError(err, "PinG1TEPoints failed")

		msm, err := plonk.NewG1MSM(dev, pts)
		assert.NoError(err, "NewG1MSM failed")
		defer msm.Close()

		results := msm.MultiExp(scalars)
		gpuResult := results[0]

		cpuResult := cpuMSM(allAffine[:n], scalars)
		if !gpuResult.Equal(&cpuResult) {
			t.Errorf("GPU MSM != CPU MSM for zero scalars")
		}
	})

	t.Run("SinglePoint", func(t *testing.T) {
		assert := require.New(t)

		scalars := make([]fr.Element, 1)
		scalars[0].SetRandom()

		pts, err := plonk.PinG1TEPoints(allTE[:1])
		assert.NoError(err, "PinG1TEPoints failed")

		msm, err := plonk.NewG1MSM(dev, pts)
		assert.NoError(err, "NewG1MSM failed")
		defer msm.Close()

		results := msm.MultiExp(scalars)
		gpuResult := results[0]

		cpuResult := cpuMSM(allAffine[:1], scalars)
		if !gpuResult.Equal(&cpuResult) {
			t.Errorf("GPU MSM != CPU MSM for single point")
		}
	})

	t.Run("AllSameScalar", func(t *testing.T) {
		assert := require.New(t)

		const n = 512
		scalars := make([]fr.Element, n)
		scalars[0].SetRandom()
		for i := 1; i < n; i++ {
			scalars[i] = scalars[0]
		}

		pts, err := plonk.PinG1TEPoints(allTE[:n])
		assert.NoError(err, "PinG1TEPoints failed")

		msm, err := plonk.NewG1MSM(dev, pts)
		assert.NoError(err, "NewG1MSM failed")
		defer msm.Close()

		results := msm.MultiExp(scalars)
		gpuResult := results[0]

		cpuResult := cpuMSM(allAffine[:n], scalars)
		if !gpuResult.Equal(&cpuResult) {
			t.Errorf("GPU MSM != CPU MSM for same scalars")
		}
	})
}

// TestMSMLarge tests GPU MSM at n=1M against precomputed reference.
func TestMSMLarge(t *testing.T) {
	assert := require.New(t)

	if testing.Short() {
		t.Skip("skipping large MSM test in short mode")
	}

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	const n = 1 << 20
	tePoints := loadSRSTEPoints(t, n)
	scalars := loadTestScalars(n)

	ref, ok := msmReferenceResults[n]
	if !ok {
		t.Fatalf("no precomputed reference for n=%d", n)
	}

	pts, err := plonk.PinG1TEPoints(tePoints)
	assert.NoError(err, "PinG1TEPoints failed")

	msm, err := plonk.NewG1MSM(dev, pts)
	assert.NoError(err, "NewG1MSM failed")
	defer msm.Close()

	results := msm.MultiExp(scalars)
	gpuResult := results[0]

	if !gpuResult.Equal(&ref) {
		t.Errorf("GPU MSM != precomputed reference for n=%d", n)
	}
}

// TestMSMSmokeFullGPUOnly runs a full-size MSM at 2^27 against precomputed reference.
// No CPU MSM needed — this is the test that was previously slow.
func TestMSMSmokeFullGPUOnly(t *testing.T) {
	assert := require.New(t)

	if testing.Short() {
		t.Skip("skipping full-size MSM smoke test in short mode")
	}

	const n = 1 << 27
	ref, ok := msmReferenceResults[n]
	if !ok {
		t.Skip("no precomputed reference for n=2^27 (compute and add to msmReferenceResults)")
	}

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	tePoints := loadSRSTEPoints(t, n)
	scalars := loadTestScalars(n)

	pts, err := plonk.PinG1TEPoints(tePoints)
	assert.NoError(err, "PinG1TEPoints failed")

	msm, err := plonk.NewG1MSM(dev, pts)
	assert.NoError(err, "NewG1MSM failed")
	defer msm.Close()

	results := msm.MultiExp(scalars)
	gpuResult := results[0]
	if err := dev.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if !gpuResult.Equal(&ref) {
		t.Errorf("GPU MSM != precomputed reference for n=%d", n)
	}
}

func BenchmarkMSM(b *testing.B) {
	assert := require.New(b)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	sizes := []int{1 << 16, 1 << 20, 1 << 24, 1 << 26, 1 << 27}

	for _, n := range sizes {
		if testing.Short() && n > 1<<24 {
			continue
		}
		b.Run(formatSize(n), func(b *testing.B) {
			assert := require.New(b)

			tePoints := loadSRSTEPoints(b, n)
			scalars := loadTestScalars(n)

			pts, err := plonk.PinG1TEPoints(tePoints)
			assert.NoError(err, "PinG1TEPoints failed")

			msm, err := plonk.NewG1MSM(dev, pts)
			assert.NoError(err, "NewG1MSM failed")
			defer msm.Close()

			msm.MultiExp(scalars) // warmup

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				msm.MultiExp(scalars)
			}
		})
	}
}

// =============================================================================
// Batched MultiExp (variadic) Tests
// =============================================================================

// TestMSMBatchedMultiExp verifies that batched MultiExp (multiple scalar sets)
// returns the same results as individual MultiExp calls.
func TestMSMBatchedMultiExp(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	const n = 1024
	tePoints := loadSRSTEPoints(t, n)

	// Create 3 independent scalar sets (simulates L, R, O commits in plonk)
	scalars1 := loadTestScalars(n)
	scalars2 := randomFrVector(n)
	scalars3 := randomFrVector(n)

	// Batched path: single MultiExp call with 3 scalar sets
	pts1, err := plonk.PinG1TEPoints(tePoints)
	assert.NoError(err, "PinG1TEPoints failed")
	msmBatched, err := plonk.NewG1MSM(dev, pts1)
	assert.NoError(err, "NewG1MSM failed")
	defer msmBatched.Close()

	batched := msmBatched.MultiExp(scalars1, scalars2, scalars3)
	if len(batched) != 3 {
		t.Fatalf("batched MultiExp returned %d results, want 3", len(batched))
	}

	// Individual path: 3 separate MultiExp calls
	pts2, err := plonk.PinG1TEPoints(tePoints)
	assert.NoError(err, "PinG1TEPoints failed")
	msmIndiv, err := plonk.NewG1MSM(dev, pts2)
	assert.NoError(err, "NewG1MSM failed")
	defer msmIndiv.Close()

	indiv1 := msmIndiv.MultiExp(scalars1)
	indiv2 := msmIndiv.MultiExp(scalars2)
	indiv3 := msmIndiv.MultiExp(scalars3)

	if !batched[0].Equal(&indiv1[0]) {
		t.Error("batched[0] != individual[0]")
	}
	if !batched[1].Equal(&indiv2[0]) {
		t.Error("batched[1] != individual[1]")
	}
	if !batched[2].Equal(&indiv3[0]) {
		t.Error("batched[2] != individual[2]")
	}
}

// =============================================================================
// Convert and Pin Tests
// =============================================================================

func TestMSMConvertG1Points(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	const n = 1024
	tePoints := loadSRSTEPoints(t, n)
	scalars := loadTestScalars(n)

	// Direct TE path (pin + create MSM)
	pts1, err := plonk.PinG1TEPoints(tePoints)
	assert.NoError(err, "PinG1TEPoints failed")

	msm1, err := plonk.NewG1MSM(dev, pts1)
	assert.NoError(err, "NewG1MSM failed")
	defer msm1.Close()

	results1 := msm1.MultiExp(scalars)
	result1 := results1[0]

	// Second pinned path (pin again + create MSM)
	pts2, err := plonk.PinG1TEPoints(tePoints)
	assert.NoError(err, "PinG1TEPoints failed")

	if pts2.Len() != n {
		t.Fatalf("Len() = %d, want %d", pts2.Len(), n)
	}

	msm2, err := plonk.NewG1MSM(dev, pts2)
	assert.NoError(err, "NewG1MSM failed")
	defer msm2.Close()

	results2 := msm2.MultiExp(scalars)
	result2 := results2[0]

	if !result1.Equal(&result2) {
		t.Error("second NewG1MSM result does not match first NewG1MSM result")
	}

	ref, ok := msmReferenceResults[n]
	if ok && !result1.Equal(&ref) {
		t.Error("GPU MSM result does not match precomputed reference")
	}
}

// =============================================================================
// TE Point Type and Production Flow Tests
// =============================================================================

func TestConvertG1AffineToTE(t *testing.T) {
	assert := require.New(t)

	const n = 1024
	tePoints := loadSRSTEPoints(t, n)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	scalars := loadTestScalars(n)

	pts, err := plonk.PinG1TEPoints(tePoints)
	assert.NoError(err, "PinG1TEPoints failed")

	msm, err := plonk.NewG1MSM(dev, pts)
	assert.NoError(err, "NewG1MSM failed")
	defer msm.Close()

	results := msm.MultiExp(scalars)
	gpuResult := results[0]

	ref, ok := msmReferenceResults[n]
	if !ok {
		t.Fatal("no precomputed reference for n=1024")
	}
	if !gpuResult.Equal(&ref) {
		t.Error("NewG1MSM result does not match precomputed reference")
	}
}

func TestG1TEPointSerialization(t *testing.T) {
	const n = 256
	tePoints := loadSRSTEPoints(t, n)

	t.Run("Safe", func(t *testing.T) {
		assert := require.New(t)

		var buf bytes.Buffer
		if err := plonk.WriteG1TEPoints(&buf, tePoints); err != nil {
			t.Fatalf("WriteG1TEPoints failed: %v", err)
		}

		read, err := plonk.ReadG1TEPoints(&buf)
		assert.NoError(err, "ReadG1TEPoints failed")
		if len(read) != n {
			t.Fatalf("len(read) = %d, want %d", len(read), n)
		}
		for i := range tePoints {
			if read[i] != tePoints[i] {
				t.Fatalf("mismatch at point %d", i)
			}
		}
	})

	t.Run("Raw", func(t *testing.T) {
		assert := require.New(t)

		var buf bytes.Buffer
		if err := plonk.WriteG1TEPointsRaw(&buf, tePoints); err != nil {
			t.Fatalf("WriteG1TEPointsRaw failed: %v", err)
		}

		if buf.Len() != n*96 {
			t.Fatalf("raw size = %d, want %d", buf.Len(), n*96)
		}

		read, err := plonk.ReadG1TEPointsRaw(&buf, n)
		assert.NoError(err, "ReadG1TEPointsRaw failed")
		for i := range tePoints {
			if read[i] != tePoints[i] {
				t.Fatalf("mismatch at point %d", i)
			}
		}
	})

	t.Run("SafeInvalidMagic", func(t *testing.T) {
		buf := bytes.NewReader([]byte("BADMAGIC00000000"))
		_, err := plonk.ReadG1TEPoints(buf)
		if err == nil {
			t.Error("expected error on invalid magic")
		}
	})
}

func TestG1TEPointPinned(t *testing.T) {
	assert := require.New(t)

	const n = 512
	tePoints := loadSRSTEPoints(t, n)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	pts, err := plonk.PinG1TEPoints(tePoints)
	assert.NoError(err, "PinG1TEPoints failed")

	msm, err := plonk.NewG1MSM(dev, pts)
	assert.NoError(err, "NewG1MSM failed")
	defer msm.Close()

	scalars := make([]fr.Element, n)
	for i := range scalars {
		scalars[i].SetRandom()
	}

	results := msm.MultiExp(scalars)
	gpuResult := results[0]

	affinePoints := loadSRSAffinePoints(t, n)
	cpuResult := cpuMSM(affinePoints, scalars)
	if !gpuResult.Equal(&cpuResult) {
		t.Error("PinG1TEPoints path does not match CPU MSM")
	}
}

// =============================================================================
// Production Flow Profiling
// =============================================================================

func TestMSMProductionFlow(t *testing.T) {
	assert := require.New(t)

	if testing.Short() {
		t.Skip("skipping production flow test in short mode")
	}

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	sizes := []int{1 << 20, 1 << 24}

	for _, n := range sizes {
		t.Run(formatSize(n), func(t *testing.T) {
			assert := require.New(t)

			scalars := loadTestScalars(n)

			t0 := time.Now()
			tePoints := loadSRSTEPoints(t, n)
			diskLoad := time.Since(t0)

			t1 := time.Now()
			pts, err := plonk.PinG1TEPoints(tePoints)
			assert.NoError(err, "PinG1TEPoints failed")
			pinTime := time.Since(t1)

			t2 := time.Now()
			msm, err := plonk.NewG1MSM(dev, pts)
			assert.NoError(err, "NewG1MSM failed")
			defer msm.Close()
			dev.Sync()
			pinnedToGPU := time.Since(t2)

			t3 := time.Now()
			results := msm.MultiExp(scalars)
			gpuResult := results[0]
			msmRun := time.Since(t3)

			// Verify against reference if available
			ref, ok := msmReferenceResults[n]
			if ok && !gpuResult.Equal(&ref) {
				t.Error("MSM result does not match precomputed reference")
			}

			total := diskLoad + pinTime + pinnedToGPU + msmRun
			t.Logf("n=%d (%s) — Production flow (SRS TE dump → GPU → MSM):", n, formatSize(n))
			t.Logf("  1. TE dump load:        %v  (%.1f GB/s)", diskLoad, float64(n*96)/float64(diskLoad.Nanoseconds()))
			t.Logf("  2. Pin TE points:       %v", pinTime)
			t.Logf("  3. Pinned → GPU:        %v  (%.1f GB/s)", pinnedToGPU, float64(n*96)/float64(pinnedToGPU.Nanoseconds()))
			t.Logf("  4. MSM run:             %v", msmRun)
			t.Logf("  Total (load+pin+upload+MSM): %v", total)
		})
	}
}
