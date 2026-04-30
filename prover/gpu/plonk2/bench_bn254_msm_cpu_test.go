package plonk2

import (
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	bnfr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/stretchr/testify/require"
)

func BenchmarkBN254MSMCommitRawSizesCPU(b *testing.B) {
	for _, n := range bn254MSMCPUBenchSizes(b) {
		b.Run(fmt.Sprintf("n=%s", benchCPUFormatCount(n)), func(b *testing.B) {
			benchBN254MSMCPUSize(b, n, bn254MSMScalarMode(b))
		})
	}
}

func benchBN254MSMCPUSize(b *testing.B, n int, scalarMode string) {
	points := repeatedBN254Point(n)
	scalars := deterministicBN254Scalars(n, scalarMode)
	b.Cleanup(func() {
		points = nil
		scalars = nil
		runtime.GC()
	})

	var warmup bn254.G1Jac
	_, err := warmup.MultiExp(points, scalars, ecc.MultiExpConfig{})
	require.NoError(b, err, "warmup BN254 CPU MSM should succeed")

	b.SetBytes(int64(n) * int64(bnfr.Bytes))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var got bn254.G1Jac
		_, err := got.MultiExp(points, scalars, ecc.MultiExpConfig{})
		require.NoError(b, err, "BN254 CPU MSM should succeed")
	}
}

func repeatedBN254Point(n int) []bn254.G1Affine {
	var p bn254.G1Affine
	p.ScalarMultiplicationBase(big.NewInt(1))

	points := make([]bn254.G1Affine, n)
	for i := range points {
		points[i] = p
	}
	return points
}

func deterministicBN254Scalars(n int, mode string) []bnfr.Element {
	switch mode {
	case "full":
		return deterministicFullBN254Scalars(n)
	case "low64":
		return deterministicLow64BN254Scalars(n)
	default:
		panic(fmt.Sprintf("unknown BN254 scalar mode %q", mode))
	}
}

func deterministicLow64BN254Scalars(n int) []bnfr.Element {
	scalars := make([]bnfr.Element, n)
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

func deterministicFullBN254Scalars(n int) []bnfr.Element {
	// #nosec G404 - deterministic benchmark vectors are required for reproducibility.
	rng := rand.New(rand.NewSource(254))
	modulus := bnfr.Modulus()
	scalars := make([]bnfr.Element, n)
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
		var minusOne bnfr.Element
		minusOne.SetBigInt(new(big.Int).Sub(modulus, big.NewInt(1)))
		scalars[2] = minusOne
	}
	return scalars
}

func bn254MSMScalarMode(tb testing.TB) string {
	tb.Helper()
	mode := strings.TrimSpace(os.Getenv("PLONK2_BN254_MSM_SCALAR_MODE"))
	if mode == "" {
		return "full"
	}
	require.Contains(tb, []string{"low64", "full"}, mode, "scalar mode should be supported")
	return mode
}

func bn254MSMCPUBenchSizes(tb testing.TB) []int {
	tb.Helper()
	raw := strings.TrimSpace(os.Getenv("PLONK2_BN254_MSM_BENCH_SIZES"))
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
