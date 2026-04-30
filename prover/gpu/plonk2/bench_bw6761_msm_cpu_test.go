package plonk2

import (
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	bw6761 "github.com/consensys/gnark-crypto/ecc/bw6-761"
	bwfr "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	"github.com/stretchr/testify/require"
)

func BenchmarkBW6761MSMCommitRawSizesCPU(b *testing.B) {
	for _, n := range bw6761MSMCPUBenchSizes(b) {
		b.Run(fmt.Sprintf("n=%s", benchCPUFormatCount(n)), func(b *testing.B) {
			benchBW6761MSMCPUSize(b, n, bw6761MSMScalarMode(b))
		})
	}
}

func benchBW6761MSMCPUSize(b *testing.B, n int, scalarMode string) {
	points := repeatedBW6761Point(n)
	scalars := deterministicBW6761Scalars(n, scalarMode)
	b.Cleanup(func() {
		points = nil
		scalars = nil
		runtime.GC()
	})

	var warmup bw6761.G1Jac
	_, err := warmup.MultiExp(points, scalars, ecc.MultiExpConfig{})
	require.NoError(b, err, "warmup BW6-761 CPU MSM should succeed")

	b.SetBytes(int64(n) * int64(bwfr.Bytes))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var got bw6761.G1Jac
		_, err := got.MultiExp(points, scalars, ecc.MultiExpConfig{})
		require.NoError(b, err, "BW6-761 CPU MSM should succeed")
	}
}

func repeatedBW6761Point(n int) []bw6761.G1Affine {
	var p bw6761.G1Affine
	p.ScalarMultiplicationBase(big.NewInt(1))

	points := make([]bw6761.G1Affine, n)
	for i := range points {
		points[i] = p
	}
	return points
}

func deterministicBW6761Scalars(n int, mode string) []bwfr.Element {
	switch mode {
	case "full":
		return deterministicFullBW6761Scalars(n)
	case "low64":
		return deterministicLow64BW6761Scalars(n)
	default:
		panic(fmt.Sprintf("unknown BW6-761 scalar mode %q", mode))
	}
}

func deterministicLow64BW6761Scalars(n int) []bwfr.Element {
	scalars := make([]bwfr.Element, n)
	x := uint64(0x9e3779b97f4a7c15)
	for i := range scalars {
		x += 0x9e3779b97f4a7c15
		x ^= x >> 30
		x *= 0xbf58476d1ce4e5b9
		x ^= x >> 27
		x *= 0x94d049bb133111eb
		x ^= x >> 31
		scalars[i].SetUint64(x)
	}
	return scalars
}

func deterministicFullBW6761Scalars(n int) []bwfr.Element {
	// #nosec G404 - deterministic benchmark vectors are required for reproducibility.
	rng := rand.New(rand.NewSource(761))
	modulus := bwfr.Modulus()
	scalars := make([]bwfr.Element, n)
	for i := range scalars {
		scalars[i].SetBigInt(new(big.Int).Rand(rng, modulus))
	}
	if n > 0 {
		scalars[0].SetZero()
	}
	if n > 1 {
		scalars[1].SetOne()
	}
	if n > 2 {
		var minusOne bwfr.Element
		minusOne.SetBigInt(new(big.Int).Sub(modulus, big.NewInt(1)))
		scalars[2] = minusOne
	}
	return scalars
}

func bw6761MSMScalarMode(tb testing.TB) string {
	tb.Helper()
	mode := strings.TrimSpace(os.Getenv("PLONK2_BW6_MSM_SCALAR_MODE"))
	if mode == "" {
		return "full"
	}
	require.Contains(tb, []string{"low64", "full"}, mode, "scalar mode should be supported")
	return mode
}

func bw6761MSMCPUBenchSizes(tb testing.TB) []int {
	tb.Helper()
	raw := strings.TrimSpace(os.Getenv("PLONK2_BW6_MSM_BENCH_SIZES"))
	if raw == "" {
		return []int{1 << 14, 1 << 16, 1 << 18, 1 << 20}
	}

	parts := strings.Split(raw, ",")
	sizes := make([]int, 0, len(parts))
	for _, part := range parts {
		n, err := parseCPUBenchSize(strings.TrimSpace(part))
		require.NoError(tb, err, "parsing benchmark size %q should succeed", part)
		require.Positive(tb, n, "benchmark size should be positive")
		sizes = append(sizes, n)
	}
	return sizes
}

func parseCPUBenchSize(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("empty size")
	}
	multiplier := 1
	lower := strings.ToLower(s)
	switch {
	case strings.HasSuffix(lower, "ki"):
		multiplier = 1 << 10
		s = s[:len(s)-2]
	case strings.HasSuffix(lower, "mi"):
		multiplier = 1 << 20
		s = s[:len(s)-2]
	}
	if s == "" {
		return 0, fmt.Errorf("empty size")
	}
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

func benchCPUFormatCount(n int) string {
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
