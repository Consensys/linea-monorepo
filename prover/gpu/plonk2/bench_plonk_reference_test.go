package plonk2

import (
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	gnarkplonk "github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/test/unsafekzg"
	"github.com/stretchr/testify/require"
)

type benchChainCircuit struct {
	X []frontend.Variable
	Y frontend.Variable `gnark:",public"`
}

func (c *benchChainCircuit) Define(api frontend.API) error {
	acc := frontend.Variable(1)
	for i := range c.X {
		acc = api.Mul(acc, c.X[i])
		acc = api.Add(acc, benchChainAddend(i))
	}
	api.AssertIsEqual(acc, c.Y)
	return nil
}

type benchPlonkCurve struct {
	name string
	id   ecc.ID
}

var benchPlonkCurves = []benchPlonkCurve{
	{name: "bn254", id: ecc.BN254},
	{name: "bls12-377", id: ecc.BLS12_377},
	{name: "bw6-761", id: ecc.BW6_761},
}

var benchPlonkConstraintCounts = []int{128, 1024}

func BenchmarkPlonkReferenceCPUSetup(b *testing.B) {
	for _, curve := range benchPlonkCurves {
		for _, constraints := range plonkBenchConstraintCounts(b) {
			b.Run(benchPlonkCaseName(curve.name, constraints), func(b *testing.B) {
				benchPlonkReferenceCPUSetup(b, curve.id, constraints)
			})
		}
	}
}

func BenchmarkPlonkReferenceCPUProve(b *testing.B) {
	for _, curve := range benchPlonkCurves {
		for _, constraints := range plonkBenchConstraintCounts(b) {
			b.Run(benchPlonkCaseName(curve.name, constraints), func(b *testing.B) {
				benchPlonkReferenceCPUProve(b, curve.id, constraints)
			})
		}
	}
}

func benchPlonkReferenceCPUSetup(b *testing.B, curveID ecc.ID, constraints int) {
	if testing.Short() && constraints > 16 {
		b.Skip("skipping larger PlonK setup benchmark in short mode")
	}

	ccs, err := frontend.Compile(
		curveID.ScalarField(),
		scs.NewBuilder,
		newBenchChainCircuit(constraints),
	)
	require.NoError(b, err, "compiling benchmark circuit should succeed")
	srs, srsLagrange, err := unsafekzg.NewSRS(ccs)
	require.NoError(b, err, "creating unsafe test SRS should succeed")

	pk, vk, err := gnarkplonk.Setup(ccs, srs, srsLagrange)
	require.NoError(b, err, "warmup PlonK setup should succeed")
	require.NotNil(b, pk, "warmup proving key should not be nil")
	require.NotNil(b, vk, "warmup verifying key should not be nil")

	b.ReportMetric(float64(ccs.GetNbConstraints()), "constraints")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pk, vk, err := gnarkplonk.Setup(ccs, srs, srsLagrange)
		require.NoError(b, err, "PlonK setup should succeed")
		require.NotNil(b, pk, "proving key should not be nil")
		require.NotNil(b, vk, "verifying key should not be nil")
	}
}

func benchPlonkReferenceCPUProve(b *testing.B, curveID ecc.ID, constraints int) {
	if testing.Short() && constraints > 16 {
		b.Skip("skipping larger PlonK prove benchmark in short mode")
	}

	ccs, err := frontend.Compile(
		curveID.ScalarField(),
		scs.NewBuilder,
		newBenchChainCircuit(constraints),
	)
	require.NoError(b, err, "compiling benchmark circuit should succeed")
	srs, srsLagrange, err := unsafekzg.NewSRS(ccs)
	require.NoError(b, err, "creating unsafe test SRS should succeed")
	pk, vk, err := gnarkplonk.Setup(ccs, srs, srsLagrange)
	require.NoError(b, err, "PlonK setup should succeed")

	witness, err := frontend.NewWitness(
		newBenchChainAssignment(curveID, constraints),
		curveID.ScalarField(),
	)
	require.NoError(b, err, "creating witness should succeed")
	publicWitness, err := witness.Public()
	require.NoError(b, err, "extracting public witness should succeed")

	proof, err := gnarkplonk.Prove(ccs, pk, witness)
	require.NoError(b, err, "warmup PlonK prove should succeed")
	require.NoError(b, gnarkplonk.Verify(proof, vk, publicWitness), "warmup proof should verify")

	b.ReportMetric(float64(ccs.GetNbConstraints()), "constraints")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		proof, err := gnarkplonk.Prove(ccs, pk, witness)
		require.NoError(b, err, "PlonK prove should succeed")
		if proof == nil {
			b.Fatal("PlonK prove returned nil proof")
		}
	}
}

func newBenchChainCircuit(constraints int) *benchChainCircuit {
	return &benchChainCircuit{
		X: make([]frontend.Variable, constraints),
	}
}

func newBenchChainAssignment(curveID ecc.ID, constraints int) *benchChainCircuit {
	modulus := curveID.ScalarField()
	xs := make([]frontend.Variable, constraints)
	acc := big.NewInt(1)
	for i := range xs {
		x := big.NewInt(int64(2 + i%13))
		xs[i] = new(big.Int).Set(x)
		acc.Mul(acc, x)
		acc.Add(acc, new(big.Int).SetUint64(benchChainAddend(i)))
		acc.Mod(acc, modulus)
	}
	return &benchChainCircuit{
		X: xs,
		Y: new(big.Int).Set(acc),
	}
}

func benchChainAddend(i int) uint64 {
	return uint64(3 + i%7)
}

func benchPlonkCaseName(curve string, constraints int) string {
	return fmt.Sprintf("%s/constraints=%s", curve, benchPlonkSizeLabel(constraints))
}

func TestParsePlonkBenchConstraintCount(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want int
	}{
		{in: "17M", want: 17_000_000},
		{in: "17Mi", want: 17 << 20},
		{in: "1024", want: 1024},
		{in: "1K", want: 1_000},
		{in: "1Ki", want: 1 << 10},
	} {
		t.Run(tc.in, func(t *testing.T) {
			got, err := parsePlonkBenchConstraintCount(tc.in)
			require.NoError(t, err, "valid count should parse")
			require.Equal(t, tc.want, got, "parsed count should match")
		})
	}
}

func benchPlonkSizeLabel(n int) string {
	if n >= 1<<20 && n%(1<<20) == 0 {
		return fmt.Sprintf("%dM", n>>20)
	}
	if n >= 1<<10 && n%(1<<10) == 0 {
		return fmt.Sprintf("%dK", n>>10)
	}
	return fmt.Sprintf("%d", n)
}

func plonkBenchConstraintCounts(tb testing.TB) []int {
	tb.Helper()
	raw := strings.TrimSpace(os.Getenv("PLONK2_PLONK_BENCH_CONSTRAINTS"))
	if raw == "" {
		return benchPlonkConstraintCounts
	}

	parts := strings.Split(raw, ",")
	counts := make([]int, 0, len(parts))
	for _, part := range parts {
		n, err := parsePlonkBenchConstraintCount(strings.TrimSpace(part))
		require.NoError(tb, err, "parsing benchmark constraint count %q should succeed", part)
		require.Positive(tb, n, "benchmark constraint count should be positive")
		counts = append(counts, n)
	}
	return counts
}

func parsePlonkBenchConstraintCount(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("empty constraint count")
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
	case strings.HasSuffix(lower, "k"):
		multiplier = 1_000
		s = s[:len(s)-1]
	case strings.HasSuffix(lower, "m"):
		multiplier = 1_000_000
		s = s[:len(s)-1]
	}
	if s == "" {
		return 0, fmt.Errorf("empty constraint count")
	}
	base, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	if base > int(^uint(0)>>1)/multiplier {
		return 0, fmt.Errorf("constraint count overflows int")
	}
	return base * multiplier, nil
}
