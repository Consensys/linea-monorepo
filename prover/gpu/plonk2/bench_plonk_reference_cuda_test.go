//go:build cuda

package plonk2

import (
	"fmt"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377"
	blsfr "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	blsfft "github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	bnfr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	bnfft "github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
	bw6761 "github.com/consensys/gnark-crypto/ecc/bw6-761"
	bwfr "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	bwfft "github.com/consensys/gnark-crypto/ecc/bw6-761/fr/fft"
	gnarkplonk "github.com/consensys/gnark/backend/plonk"
	blsplonk "github.com/consensys/gnark/backend/plonk/bls12-377"
	bnplonk "github.com/consensys/gnark/backend/plonk/bn254"
	bwplonk "github.com/consensys/gnark/backend/plonk/bw6-761"
	blscs "github.com/consensys/gnark/constraint/bls12-377"
	bncs "github.com/consensys/gnark/constraint/bn254"
	bwcs "github.com/consensys/gnark/constraint/bw6-761"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/stretchr/testify/require"

	"github.com/consensys/linea-monorepo/prover/gpu"
)

func BenchmarkPlonkReferenceSetupCommitmentsCPUvsGPU_CUDA(b *testing.B) {
	dev, err := gpu.New()
	require.NoError(b, err, "creating CUDA device should succeed")
	defer dev.Close()

	for _, constraints := range plonkBenchConstraintCounts(b) {
		b.Run(benchPlonkCaseName("bn254", constraints), func(b *testing.B) {
			benchSetupCommitmentsBN254(b, dev, constraints)
		})
		b.Run(benchPlonkCaseName("bls12-377", constraints), func(b *testing.B) {
			benchSetupCommitmentsBLS12377(b, dev, constraints)
		})
		b.Run(benchPlonkCaseName("bw6-761", constraints), func(b *testing.B) {
			benchSetupCommitmentsBW6761(b, dev, constraints)
		})
	}
}

type setupCommitmentsBN254 struct {
	points      []bn254.G1Affine
	coeffs      [][]bnfr.Element
	wants       []bn254.G1Affine
	labels      []string
	constraints int
}

type setupCommitmentsBLS12377 struct {
	points      []bls12377.G1Affine
	coeffs      [][]blsfr.Element
	wants       []bls12377.G1Affine
	labels      []string
	constraints int
}

type setupCommitmentsBW6761 struct {
	points      []bw6761.G1Affine
	coeffs      [][]bwfr.Element
	wants       []bw6761.G1Affine
	labels      []string
	constraints int
}

func benchSetupCommitmentsBN254(b *testing.B, dev *gpu.Device, constraints int) {
	if testing.Short() && constraints > 16 {
		b.Skip("skipping larger PlonK setup-commitment benchmark in short mode")
	}
	setup := newSetupCommitmentsBN254(b, constraints)
	rawScalars := rawSetupScalarsBN254(setup.coeffs)

	b.Run("cpu", func(b *testing.B) {
		requireCPUSetupCommitmentsBN254(b, setup)
		benchSetupCommitmentsMetadata(b, setup.constraints, len(setup.points), totalBN254Scalars(setup.coeffs))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			require.NoError(b, commitSetupBN254CPU(setup.points, setup.coeffs), "CPU setup commitments should succeed")
		}
	})

	b.Run("gpu", func(b *testing.B) {
		msm, err := NewG1MSM(dev, CurveBN254, rawBN254G1Slice(setup.points))
		require.NoError(b, err, "creating BN254 resident MSM should succeed")
		defer msm.Close()
		require.NoError(b, msm.PinWorkBuffers(), "pinning BN254 MSM work buffers should succeed")
		defer func() { require.NoError(b, msm.ReleaseWorkBuffers()) }()
		requireGPUSetupCommitmentsBN254(b, msm, setup)
		require.NoError(b, dev.Sync(), "device sync before benchmark should succeed")
		benchSetupCommitmentsMetadata(b, setup.constraints, len(setup.points), totalBN254Scalars(setup.coeffs))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			require.NoError(b, commitSetupBN254GPU(msm, rawScalars), "GPU setup commitments should succeed")
		}
	})
}

func benchSetupCommitmentsBLS12377(b *testing.B, dev *gpu.Device, constraints int) {
	if testing.Short() && constraints > 16 {
		b.Skip("skipping larger PlonK setup-commitment benchmark in short mode")
	}
	setup := newSetupCommitmentsBLS12377(b, constraints)
	rawScalars := rawSetupScalarsBLS12377(setup.coeffs)

	b.Run("cpu", func(b *testing.B) {
		requireCPUSetupCommitmentsBLS12377(b, setup)
		benchSetupCommitmentsMetadata(b, setup.constraints, len(setup.points), totalBLS12377Scalars(setup.coeffs))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			require.NoError(b, commitSetupBLS12377CPU(setup.points, setup.coeffs), "CPU setup commitments should succeed")
		}
	})

	b.Run("gpu", func(b *testing.B) {
		msm, err := NewG1MSM(dev, CurveBLS12377, rawBLS12377G1Slice(setup.points))
		require.NoError(b, err, "creating BLS12-377 resident MSM should succeed")
		defer msm.Close()
		require.NoError(b, msm.PinWorkBuffers(), "pinning BLS12-377 MSM work buffers should succeed")
		defer func() { require.NoError(b, msm.ReleaseWorkBuffers()) }()
		requireGPUSetupCommitmentsBLS12377(b, msm, setup)
		require.NoError(b, dev.Sync(), "device sync before benchmark should succeed")
		benchSetupCommitmentsMetadata(b, setup.constraints, len(setup.points), totalBLS12377Scalars(setup.coeffs))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			require.NoError(b, commitSetupBLS12377GPU(msm, rawScalars), "GPU setup commitments should succeed")
		}
	})
}

func benchSetupCommitmentsBW6761(b *testing.B, dev *gpu.Device, constraints int) {
	if testing.Short() && constraints > 16 {
		b.Skip("skipping larger PlonK setup-commitment benchmark in short mode")
	}
	setup := newSetupCommitmentsBW6761(b, constraints)
	rawScalars := rawSetupScalarsBW6761(setup.coeffs)

	b.Run("cpu", func(b *testing.B) {
		requireCPUSetupCommitmentsBW6761(b, setup)
		benchSetupCommitmentsMetadata(b, setup.constraints, len(setup.points), totalBW6761Scalars(setup.coeffs))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			require.NoError(b, commitSetupBW6761CPU(setup.points, setup.coeffs), "CPU setup commitments should succeed")
		}
	})

	b.Run("gpu", func(b *testing.B) {
		msm, err := NewG1MSM(dev, CurveBW6761, rawBW6761G1Slice(setup.points))
		require.NoError(b, err, "creating BW6-761 resident MSM should succeed")
		defer msm.Close()
		require.NoError(b, msm.PinWorkBuffers(), "pinning BW6-761 MSM work buffers should succeed")
		defer func() { require.NoError(b, msm.ReleaseWorkBuffers()) }()
		requireGPUSetupCommitmentsBW6761(b, msm, setup)
		require.NoError(b, dev.Sync(), "device sync before benchmark should succeed")
		benchSetupCommitmentsMetadata(b, setup.constraints, len(setup.points), totalBW6761Scalars(setup.coeffs))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			require.NoError(b, commitSetupBW6761GPU(msm, rawScalars), "GPU setup commitments should succeed")
		}
	})
}

func newSetupCommitmentsBN254(tb testing.TB, constraints int) setupCommitmentsBN254 {
	tb.Helper()
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), scs.NewBuilder, newBenchChainCircuit(constraints))
	require.NoError(tb, err, "compiling BN254 benchmark circuit should succeed")
	srs, srsLagrange := testSRSAssets(tb).loadForCCS(tb, ccs)
	pkI, vkI, err := gnarkplonk.Setup(ccs, srs, srsLagrange)
	require.NoError(tb, err, "BN254 PlonK setup should succeed")

	spr := ccs.(*bncs.SparseR1CS)
	pk := pkI.(*bnplonk.ProvingKey)
	vk := vkI.(*bnplonk.VerifyingKey)
	domain := bnfft.NewDomain(uint64(spr.GetNbConstraints()+len(spr.Public)), bnfft.WithoutPrecompute())
	trace := bnplonk.NewTrace(spr, domain)
	coeffs := [][]bnfr.Element{
		trace.Ql.Coefficients(),
		trace.Qr.Coefficients(),
		trace.Qm.Coefficients(),
		trace.Qo.Coefficients(),
		trace.Qk.Coefficients(),
		trace.S1.Coefficients(),
		trace.S2.Coefficients(),
		trace.S3.Coefficients(),
	}
	wants := []bn254.G1Affine{vk.Ql, vk.Qr, vk.Qm, vk.Qo, vk.Qk, vk.S[0], vk.S[1], vk.S[2]}
	labels := []string{"Ql", "Qr", "Qm", "Qo", "Qk", "S1", "S2", "S3"}
	for i := range trace.Qcp {
		coeffs = append(coeffs, trace.Qcp[i].Coefficients())
		wants = append(wants, vk.Qcp[i])
		labels = append(labels, fmt.Sprintf("Qcp[%d]", i))
	}
	return setupCommitmentsBN254{
		points:      pk.KzgLagrange.G1,
		coeffs:      coeffs,
		wants:       wants,
		labels:      labels,
		constraints: spr.GetNbConstraints(),
	}
}

func newSetupCommitmentsBLS12377(tb testing.TB, constraints int) setupCommitmentsBLS12377 {
	tb.Helper()
	ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, newBenchChainCircuit(constraints))
	require.NoError(tb, err, "compiling BLS12-377 benchmark circuit should succeed")
	srs, srsLagrange := testSRSAssets(tb).loadForCCS(tb, ccs)
	pkI, vkI, err := gnarkplonk.Setup(ccs, srs, srsLagrange)
	require.NoError(tb, err, "BLS12-377 PlonK setup should succeed")

	spr := ccs.(*blscs.SparseR1CS)
	pk := pkI.(*blsplonk.ProvingKey)
	vk := vkI.(*blsplonk.VerifyingKey)
	domain := blsfft.NewDomain(uint64(spr.GetNbConstraints()+len(spr.Public)), blsfft.WithoutPrecompute())
	trace := blsplonk.NewTrace(spr, domain)
	coeffs := [][]blsfr.Element{
		trace.Ql.Coefficients(),
		trace.Qr.Coefficients(),
		trace.Qm.Coefficients(),
		trace.Qo.Coefficients(),
		trace.Qk.Coefficients(),
		trace.S1.Coefficients(),
		trace.S2.Coefficients(),
		trace.S3.Coefficients(),
	}
	wants := []bls12377.G1Affine{vk.Ql, vk.Qr, vk.Qm, vk.Qo, vk.Qk, vk.S[0], vk.S[1], vk.S[2]}
	labels := []string{"Ql", "Qr", "Qm", "Qo", "Qk", "S1", "S2", "S3"}
	for i := range trace.Qcp {
		coeffs = append(coeffs, trace.Qcp[i].Coefficients())
		wants = append(wants, vk.Qcp[i])
		labels = append(labels, fmt.Sprintf("Qcp[%d]", i))
	}
	return setupCommitmentsBLS12377{
		points:      pk.KzgLagrange.G1,
		coeffs:      coeffs,
		wants:       wants,
		labels:      labels,
		constraints: spr.GetNbConstraints(),
	}
}

func newSetupCommitmentsBW6761(tb testing.TB, constraints int) setupCommitmentsBW6761 {
	tb.Helper()
	ccs, err := frontend.Compile(ecc.BW6_761.ScalarField(), scs.NewBuilder, newBenchChainCircuit(constraints))
	require.NoError(tb, err, "compiling BW6-761 benchmark circuit should succeed")
	srs, srsLagrange := testSRSAssets(tb).loadForCCS(tb, ccs)
	pkI, vkI, err := gnarkplonk.Setup(ccs, srs, srsLagrange)
	require.NoError(tb, err, "BW6-761 PlonK setup should succeed")

	spr := ccs.(*bwcs.SparseR1CS)
	pk := pkI.(*bwplonk.ProvingKey)
	vk := vkI.(*bwplonk.VerifyingKey)
	domain := bwfft.NewDomain(uint64(spr.GetNbConstraints()+len(spr.Public)), bwfft.WithoutPrecompute())
	trace := bwplonk.NewTrace(spr, domain)
	coeffs := [][]bwfr.Element{
		trace.Ql.Coefficients(),
		trace.Qr.Coefficients(),
		trace.Qm.Coefficients(),
		trace.Qo.Coefficients(),
		trace.Qk.Coefficients(),
		trace.S1.Coefficients(),
		trace.S2.Coefficients(),
		trace.S3.Coefficients(),
	}
	wants := []bw6761.G1Affine{vk.Ql, vk.Qr, vk.Qm, vk.Qo, vk.Qk, vk.S[0], vk.S[1], vk.S[2]}
	labels := []string{"Ql", "Qr", "Qm", "Qo", "Qk", "S1", "S2", "S3"}
	for i := range trace.Qcp {
		coeffs = append(coeffs, trace.Qcp[i].Coefficients())
		wants = append(wants, vk.Qcp[i])
		labels = append(labels, fmt.Sprintf("Qcp[%d]", i))
	}
	return setupCommitmentsBW6761{
		points:      pk.KzgLagrange.G1,
		coeffs:      coeffs,
		wants:       wants,
		labels:      labels,
		constraints: spr.GetNbConstraints(),
	}
}

func requireCPUSetupCommitmentsBN254(tb testing.TB, setup setupCommitmentsBN254) {
	tb.Helper()
	for i := range setup.coeffs {
		var gotJac bn254.G1Jac
		_, err := gotJac.MultiExp(setup.points[:len(setup.coeffs[i])], setup.coeffs[i], ecc.MultiExpConfig{})
		require.NoError(tb, err, "CPU BN254 setup commitment %s should succeed", setup.labels[i])
		var got bn254.G1Affine
		got.FromJacobian(&gotJac)
		requireBN254AffineMatches(tb, setup.wants[i], got, "setup "+setup.labels[i])
	}
}

func requireCPUSetupCommitmentsBLS12377(tb testing.TB, setup setupCommitmentsBLS12377) {
	tb.Helper()
	for i := range setup.coeffs {
		var gotJac bls12377.G1Jac
		_, err := gotJac.MultiExp(setup.points[:len(setup.coeffs[i])], setup.coeffs[i], ecc.MultiExpConfig{})
		require.NoError(tb, err, "CPU BLS12-377 setup commitment %s should succeed", setup.labels[i])
		var got bls12377.G1Affine
		got.FromJacobian(&gotJac)
		requireBLS12377AffineMatches(tb, setup.wants[i], got, "setup "+setup.labels[i])
	}
}

func requireCPUSetupCommitmentsBW6761(tb testing.TB, setup setupCommitmentsBW6761) {
	tb.Helper()
	for i := range setup.coeffs {
		var gotJac bw6761.G1Jac
		_, err := gotJac.MultiExp(setup.points[:len(setup.coeffs[i])], setup.coeffs[i], ecc.MultiExpConfig{})
		require.NoError(tb, err, "CPU BW6-761 setup commitment %s should succeed", setup.labels[i])
		var got bw6761.G1Affine
		got.FromJacobian(&gotJac)
		requireBW6761AffineMatches(tb, setup.wants[i], got, "setup "+setup.labels[i])
	}
}

func requireGPUSetupCommitmentsBN254(tb testing.TB, msm *G1MSM, setup setupCommitmentsBN254) {
	tb.Helper()
	outputs, err := msm.commitRawBatch(rawSetupScalarsBN254(setup.coeffs))
	require.NoError(tb, err, "batched BN254 setup commitments should succeed")
	for i := range outputs {
		requireBN254ProjectiveMatches(tb, setup.wants[i], outputs[i], "setup "+setup.labels[i])
	}
}

func requireGPUSetupCommitmentsBLS12377(tb testing.TB, msm *G1MSM, setup setupCommitmentsBLS12377) {
	tb.Helper()
	outputs, err := msm.commitRawBatch(rawSetupScalarsBLS12377(setup.coeffs))
	require.NoError(tb, err, "batched BLS12-377 setup commitments should succeed")
	for i := range outputs {
		requireBLS12377ProjectiveMatches(tb, setup.wants[i], outputs[i], "setup "+setup.labels[i])
	}
}

func requireGPUSetupCommitmentsBW6761(tb testing.TB, msm *G1MSM, setup setupCommitmentsBW6761) {
	tb.Helper()
	outputs, err := msm.commitRawBatch(rawSetupScalarsBW6761(setup.coeffs))
	require.NoError(tb, err, "batched BW6-761 setup commitments should succeed")
	for i := range outputs {
		requireBW6761ProjectiveMatches(tb, setup.wants[i], outputs[i], "setup "+setup.labels[i])
	}
}

func commitSetupBN254CPU(points []bn254.G1Affine, coeffs [][]bnfr.Element) error {
	for i := range coeffs {
		var got bn254.G1Jac
		if _, err := got.MultiExp(points[:len(coeffs[i])], coeffs[i], ecc.MultiExpConfig{}); err != nil {
			return err
		}
	}
	return nil
}

func commitSetupBLS12377CPU(points []bls12377.G1Affine, coeffs [][]blsfr.Element) error {
	for i := range coeffs {
		var got bls12377.G1Jac
		if _, err := got.MultiExp(points[:len(coeffs[i])], coeffs[i], ecc.MultiExpConfig{}); err != nil {
			return err
		}
	}
	return nil
}

func commitSetupBW6761CPU(points []bw6761.G1Affine, coeffs [][]bwfr.Element) error {
	for i := range coeffs {
		var got bw6761.G1Jac
		if _, err := got.MultiExp(points[:len(coeffs[i])], coeffs[i], ecc.MultiExpConfig{}); err != nil {
			return err
		}
	}
	return nil
}

func commitSetupBN254GPU(msm *G1MSM, scalars [][]uint64) error {
	_, err := msm.commitRawBatch(scalars)
	return err
}

func commitSetupBLS12377GPU(msm *G1MSM, scalars [][]uint64) error {
	_, err := msm.commitRawBatch(scalars)
	return err
}

func commitSetupBW6761GPU(msm *G1MSM, scalars [][]uint64) error {
	_, err := msm.commitRawBatch(scalars)
	return err
}

func rawSetupScalarsBN254(coeffs [][]bnfr.Element) [][]uint64 {
	raw := make([][]uint64, len(coeffs))
	for i := range coeffs {
		raw[i] = cloneRaw(rawBN254(coeffs[i]))
	}
	return raw
}

func rawSetupScalarsBLS12377(coeffs [][]blsfr.Element) [][]uint64 {
	raw := make([][]uint64, len(coeffs))
	for i := range coeffs {
		raw[i] = cloneRaw(rawBLS12377(coeffs[i]))
	}
	return raw
}

func rawSetupScalarsBW6761(coeffs [][]bwfr.Element) [][]uint64 {
	raw := make([][]uint64, len(coeffs))
	for i := range coeffs {
		raw[i] = cloneRaw(rawBW6761(coeffs[i]))
	}
	return raw
}

func benchSetupCommitmentsMetadata(b *testing.B, constraints, domain, scalarWords int) {
	b.Helper()
	b.ReportMetric(float64(constraints), "constraints")
	b.ReportMetric(float64(domain), "domain_points")
	b.ReportMetric(float64(scalarWords), "scalar_words")
	b.SetBytes(int64(scalarWords) * 8)
}

func totalBN254Scalars(coeffs [][]bnfr.Element) int {
	var total int
	for i := range coeffs {
		total += len(coeffs[i]) * bnfr.Limbs
	}
	return total
}

func totalBLS12377Scalars(coeffs [][]blsfr.Element) int {
	var total int
	for i := range coeffs {
		total += len(coeffs[i]) * blsfr.Limbs
	}
	return total
}

func totalBW6761Scalars(coeffs [][]bwfr.Element) int {
	var total int
	for i := range coeffs {
		total += len(coeffs[i]) * bwfr.Limbs
	}
	return total
}

func requireBN254AffineMatches(tb testing.TB, want, got bn254.G1Affine, label string) {
	tb.Helper()
	if want.IsInfinity() || got.IsInfinity() {
		require.Equal(tb, want.IsInfinity(), got.IsInfinity(), "%s infinity flag should match", label)
		return
	}
	require.True(tb, want.X.Equal(&got.X), "%s affine X should match", label)
	require.True(tb, want.Y.Equal(&got.Y), "%s affine Y should match", label)
}

func requireBLS12377AffineMatches(tb testing.TB, want, got bls12377.G1Affine, label string) {
	tb.Helper()
	if want.IsInfinity() || got.IsInfinity() {
		require.Equal(tb, want.IsInfinity(), got.IsInfinity(), "%s infinity flag should match", label)
		return
	}
	require.True(tb, want.X.Equal(&got.X), "%s affine X should match", label)
	require.True(tb, want.Y.Equal(&got.Y), "%s affine Y should match", label)
}

func requireBW6761AffineMatches(tb testing.TB, want, got bw6761.G1Affine, label string) {
	tb.Helper()
	if want.IsInfinity() || got.IsInfinity() {
		require.Equal(tb, want.IsInfinity(), got.IsInfinity(), "%s infinity flag should match", label)
		return
	}
	require.True(tb, want.X.Equal(&got.X), "%s affine X should match", label)
	require.True(tb, want.Y.Equal(&got.Y), "%s affine Y should match", label)
}
