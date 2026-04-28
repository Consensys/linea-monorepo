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

func BenchmarkBW6761MSMCommitRawSizes_CUDA(b *testing.B) {
	dev, err := gpu.New()
	require.NoError(b, err, "creating CUDA device should succeed")
	defer dev.Close()

	for _, n := range bw6761MSMBenchSizes(b) {
		b.Run(fmt.Sprintf("n=%s", benchFormatCount(n)), func(b *testing.B) {
			benchBW6761MSMCommitRawSize(b, dev, n)
		})
	}
}

func TestPlanMSMMemory_BW6761ThirtyMillionPoints(t *testing.T) {
	cfg, err := DefaultMSMPlanConfig(CurveBW6761, 30_000_000)
	require.NoError(t, err, "creating default BW6-761 MSM config should succeed")
	cfg.WindowBits = defaultMSMWindowBits(mustCurveInfo(t, CurveBW6761), cfg.Points)
	plan, err := PlanMSMMemory(cfg)
	require.NoError(t, err, "planning 30M point BW6-761 MSM should succeed")

	require.Equal(t, 18, plan.WindowBits, "30M point BW6-761 MSM should use the large-window policy")
	require.Less(t, plan.EstimatedTotalBytes, uint64(40<<30), "30M point BW6-761 MSM should fit comfortably below 40 GiB")
}

func benchBW6761MSMCommitRawSize(b *testing.B, dev *gpu.Device, n int) {
	info := mustCurveInfo(b, CurveBW6761)
	windowBits := defaultMSMWindowBits(info, n)
	if override := bw6761MSMBenchWindowBits(b); override > 0 {
		windowBits = override
	}
	cfg := MSMPlanConfig{
		Curve:      CurveBW6761,
		Points:     n,
		WindowBits: windowBits,
		Layout:     PointLayoutAffineSW,
	}
	plan, err := PlanMSMMemory(cfg)
	require.NoError(b, err, "planning BW6-761 MSM memory should succeed")

	b.ReportMetric(float64(windowBits), "window_bits")
	b.ReportMetric(float64(plan.Windows), "windows")
	b.ReportMetric(bytesToGiB(plan.EstimatedTotalBytes), "estimated_GiB")

	points := repeatedBW6761PointRaw(n)
	scalars := deterministicBW6761ScalarRaw(n)
	b.Cleanup(func() {
		points = nil
		scalars = nil
		runtime.GC()
	})

	msm, err := newG1MSMWithWindowBits(dev, CurveBW6761, points, windowBits)
	require.NoError(b, err, "creating BW6-761 resident MSM should succeed")
	defer msm.Close()
	require.NoError(b, msm.PinWorkBuffers(), "pinning BW6-761 MSM work buffers should succeed")
	defer func() { require.NoError(b, msm.ReleaseWorkBuffers()) }()

	_, err = msm.CommitRaw(scalars)
	require.NoError(b, err, "warmup BW6-761 MSM should succeed")
	require.NoError(b, dev.Sync(), "device sync before benchmark should succeed")

	b.SetBytes(int64(n) * int64(info.ScalarLimbs*8))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := msm.CommitRaw(scalars)
		require.NoError(b, err, "BW6-761 MSM should succeed")
	}
}

func bw6761MSMBenchSizes(tb testing.TB) []int {
	tb.Helper()
	raw := strings.TrimSpace(os.Getenv("PLONK2_BW6_MSM_BENCH_SIZES"))
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

func bw6761MSMBenchWindowBits(tb testing.TB) int {
	tb.Helper()
	raw := strings.TrimSpace(os.Getenv("PLONK2_BW6_MSM_WINDOW_BITS"))
	if raw == "" {
		return 0
	}
	windowBits, err := strconv.Atoi(raw)
	require.NoError(tb, err, "parsing PLONK2_BW6_MSM_WINDOW_BITS should succeed")
	require.GreaterOrEqual(tb, windowBits, 2, "window bits should be at least 2")
	require.LessOrEqual(tb, windowBits, 24, "window bits should be at most 24")
	return windowBits
}

func parseBenchSize(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("empty size")
	}
	multiplier := 1
	last := s[len(s)-1]
	switch last {
	case 'k', 'K':
		multiplier = 1 << 10
		s = s[:len(s)-1]
	case 'm', 'M':
		multiplier = 1 << 20
		s = s[:len(s)-1]
	}
	base, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	if base > int(^uint(0)>>1)/multiplier {
		return 0, fmt.Errorf("size overflows int")
	}
	return base * multiplier, nil
}

func repeatedBW6761PointRaw(n int) []uint64 {
	p := bw6761Point(1)
	point := rawBW6761G1(&p)
	wordsPerPoint := len(point)
	out := make([]uint64, n*wordsPerPoint)
	for i := 0; i < n; i++ {
		copy(out[i*wordsPerPoint:], point)
	}
	return out
}

func deterministicBW6761ScalarRaw(n int) []uint64 {
	const limbs = 6
	out := make([]uint64, n*limbs)
	x := uint64(0x9e3779b97f4a7c15)
	for i := range out {
		x += 0x9e3779b97f4a7c15
		x ^= x >> 30
		x *= 0xbf58476d1ce4e5b9
		x ^= x >> 27
		x *= 0x94d049bb133111eb
		x ^= x >> 31
		out[i] = x
	}
	return out
}

func bytesToGiB(v uint64) float64 {
	return float64(v) / float64(uint64(1)<<30)
}

func benchFormatCount(n int) string {
	if n >= 1_000_000 && n%1_000_000 == 0 {
		return fmt.Sprintf("%dM", n/1_000_000)
	}
	if n >= 1<<20 && n%(1<<20) == 0 {
		return fmt.Sprintf("%dMi", n>>20)
	}
	if n >= 1<<10 && n%(1<<10) == 0 {
		return fmt.Sprintf("%dKi", n>>10)
	}
	return strconv.Itoa(n)
}

func mustCurveInfo(tb testing.TB, curve Curve) CurveInfo {
	tb.Helper()
	info, ok := curve.Info()
	require.True(tb, ok, "curve info should exist")
	return info
}
