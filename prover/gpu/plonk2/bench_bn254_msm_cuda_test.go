//go:build cuda

package plonk2

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/consensys/linea-monorepo/prover/gpu"
)

func BenchmarkBN254MSMCommitRawSizes_CUDA(b *testing.B) {
	dev, err := gpu.New()
	require.NoError(b, err, "creating CUDA device should succeed")
	defer dev.Close()

	for _, n := range bn254MSMBenchSizes(b) {
		b.Run(fmt.Sprintf("n=%s", benchFormatCount(n)), func(b *testing.B) {
			benchBN254MSMCommitRawSize(b, dev, n)
		})
	}
}

func benchBN254MSMCommitRawSize(b *testing.B, dev *gpu.Device, n int) {
	info := mustCurveInfo(b, CurveBN254)
	windowBits := defaultMSMWindowBits(info, n)
	if override := bn254MSMBenchWindowBits(b); override > 0 {
		windowBits = override
	}
	cfg := MSMPlanConfig{
		Curve:      CurveBN254,
		Points:     n,
		WindowBits: windowBits,
		Layout:     PointLayoutAffineSW,
	}
	plan, err := PlanMSMMemory(cfg)
	require.NoError(b, err, "planning BN254 MSM memory should succeed")

	b.ReportMetric(float64(windowBits), "window_bits")
	b.ReportMetric(float64(plan.Windows), "windows")
	b.ReportMetric(bytesToGiB(plan.EstimatedTotalBytes), "estimated_GiB")

	points := rawBN254G1Slice(testSRSAssets(b).loadBN254(b, n, true).Pk.G1)
	scalars := deterministicBN254ScalarRaw(n, bn254MSMScalarMode(b))
	b.Cleanup(func() {
		points = nil
		scalars = nil
		runtime.GC()
	})

	msm, err := newG1MSMWithWindowBits(dev, CurveBN254, points, windowBits)
	require.NoError(b, err, "creating BN254 resident MSM should succeed")
	defer msm.Close()
	require.NoError(b, msm.PinWorkBuffers(), "pinning BN254 MSM work buffers should succeed")
	defer func() { require.NoError(b, msm.ReleaseWorkBuffers()) }()

	_, err = msm.CommitRaw(scalars)
	require.NoError(b, err, "warmup BN254 MSM should succeed")
	require.NoError(b, dev.Sync(), "device sync before benchmark should succeed")

	b.SetBytes(int64(n) * int64(info.ScalarLimbs*8))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := msm.CommitRaw(scalars)
		require.NoError(b, err, "BN254 MSM should succeed")
	}
	timings := msm.LastPhaseTimings()
	for phase := MSMPhase(0); phase < MSMPhaseCount; phase++ {
		b.ReportMetric(float64(timings[phase]*1000), phase.String()+"_us")
	}
}

func bn254MSMBenchSizes(tb testing.TB) []int {
	tb.Helper()
	raw := strings.TrimSpace(os.Getenv("PLONK2_BN254_MSM_BENCH_SIZES"))
	if raw == "" {
		return []int{1 << 14, 1 << 16, 1 << 18, 1 << 20}
	}

	parts := strings.Split(raw, ",")
	sizes := make([]int, 0, len(parts))
	for _, part := range parts {
		n, err := parseBenchSize(strings.TrimSpace(part))
		require.NoError(tb, err, "parsing benchmark size %q should succeed", part)
		require.Positive(tb, n, "benchmark size should be positive")
		sizes = append(sizes, n)
	}
	return sizes
}

func bn254MSMBenchWindowBits(tb testing.TB) int {
	tb.Helper()
	raw := strings.TrimSpace(os.Getenv("PLONK2_BN254_MSM_WINDOW_BITS"))
	if raw == "" {
		return 0
	}
	windowBits, err := strconv.Atoi(raw)
	require.NoError(tb, err, "parsing PLONK2_BN254_MSM_WINDOW_BITS should succeed")
	require.GreaterOrEqual(tb, windowBits, 2, "window bits should be at least 2")
	require.LessOrEqual(tb, windowBits, 24, "window bits should be at most 24")
	return windowBits
}

func deterministicBN254ScalarRaw(n int, mode string) []uint64 {
	return rawBN254(deterministicBN254Scalars(n, mode))
}
