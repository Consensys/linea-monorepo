//go:build cuda

// Package gpu provides comprehensive GPU benchmarks with H2D/compute/D2H breakdown.
//
// Run:
//
//	go test -tags cuda -bench=. -benchtime=3s -timeout=0 ./gpu/ -v 2>&1 | tee gpu_bench.txt
package gpu

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// BenchmarkDeviceInit measures GPU context creation overhead.
func BenchmarkDeviceInit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dev, err := New()
		if err != nil {
			b.Fatal(err)
		}
		dev.Close()
	}
}

// BenchmarkDeviceSync measures round-trip sync latency (no-op sync).
func BenchmarkDeviceSync(b *testing.B) {
	dev, err := New()
	require.NoError(b, err)
	defer dev.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dev.Sync()
	}
}

// BenchmarkMemGetInfo measures VRAM query latency.
func BenchmarkMemGetInfo(b *testing.B) {
	dev, err := New()
	require.NoError(b, err)
	defer dev.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := dev.MemGetInfo()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkStreamCreate measures stream creation overhead.
func BenchmarkStreamCreate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dev, err := New()
		if err != nil {
			b.Fatal(err)
		}
		dev.InitMultiStream()
		dev.Close()
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Timing helpers
// ─────────────────────────────────────────────────────────────────────────────

type TimingResult struct {
	Name    string
	H2D     time.Duration
	Compute time.Duration
	D2H     time.Duration
	Total   time.Duration
	N       int
	Bytes   int64 // total data moved
}

func (t TimingResult) Report(b *testing.B) {
	b.ReportMetric(float64(t.H2D.Microseconds()), "h2d_µs")
	b.ReportMetric(float64(t.Compute.Microseconds()), "compute_µs")
	b.ReportMetric(float64(t.D2H.Microseconds()), "d2h_µs")
	b.ReportMetric(float64(t.Total.Microseconds()), "total_µs")
	if t.Bytes > 0 {
		gbps := float64(t.Bytes) / t.Total.Seconds() / 1e9
		b.ReportMetric(gbps, "GB/s")
	}
}

func fmtSize(n int) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%dM", n>>20)
	case n >= 1<<10:
		return fmt.Sprintf("%dK", n>>10)
	default:
		return fmt.Sprintf("%d", n)
	}
}
