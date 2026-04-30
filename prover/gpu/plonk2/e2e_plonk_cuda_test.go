//go:build cuda

package plonk2

import (
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

func TestPlonkE2EGPUSetupCommitments_AllTargetCurves_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	t.Run("bn254", func(t *testing.T) { testPlonkE2EGPUSetupBN254(t, dev) })
	t.Run("bls12-377", func(t *testing.T) { testPlonkE2EGPUSetupBLS12377(t, dev) })
	t.Run("bw6-761", func(t *testing.T) { testPlonkE2EGPUSetupBW6761(t, dev) })
}

func testPlonkE2EGPUSetupBN254(t *testing.T, dev *gpu.Device) {
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), scs.NewBuilder, newBenchChainCircuit(128))
	require.NoError(t, err, "compiling BN254 circuit should succeed")

	srs, srsLagrange := testSRSAssets(t).loadForCCS(t, ccs)
	pkI, vkI, err := gnarkplonk.Setup(ccs, srs, srsLagrange)
	require.NoError(t, err, "BN254 PlonK setup should succeed")

	fullWitness, err := frontend.NewWitness(newBenchChainAssignment(ecc.BN254, 128), ecc.BN254.ScalarField())
	require.NoError(t, err, "creating BN254 witness should succeed")
	publicWitness, err := fullWitness.Public()
	require.NoError(t, err, "extracting BN254 public witness should succeed")
	proof, err := gnarkplonk.Prove(ccs, pkI, fullWitness)
	require.NoError(t, err, "BN254 PlonK prove should succeed")
	require.NoError(t, gnarkplonk.Verify(proof, vkI, publicWitness), "BN254 PlonK verify should succeed")

	spr := ccs.(*bncs.SparseR1CS)
	pk := pkI.(*bnplonk.ProvingKey)
	vk := vkI.(*bnplonk.VerifyingKey)
	domain := bnfft.NewDomain(uint64(spr.GetNbConstraints()+len(spr.Public)), bnfft.WithoutPrecompute())
	trace := bnplonk.NewTrace(spr, domain)

	msm, err := NewG1MSM(dev, CurveBN254, rawBN254G1Slice(pk.KzgLagrange.G1))
	require.NoError(t, err, "creating BN254 resident setup MSM should succeed")
	defer msm.Close()

	requireGPUCommitBN254(t, msm, trace.Ql.Coefficients(), vk.Ql, "Ql")
	requireGPUCommitBN254(t, msm, trace.Qr.Coefficients(), vk.Qr, "Qr")
	requireGPUCommitBN254(t, msm, trace.Qm.Coefficients(), vk.Qm, "Qm")
	requireGPUCommitBN254(t, msm, trace.Qo.Coefficients(), vk.Qo, "Qo")
	requireGPUCommitBN254(t, msm, trace.Qk.Coefficients(), vk.Qk, "Qk")
	requireGPUCommitBN254(t, msm, trace.S1.Coefficients(), vk.S[0], "S1")
	requireGPUCommitBN254(t, msm, trace.S2.Coefficients(), vk.S[1], "S2")
	requireGPUCommitBN254(t, msm, trace.S3.Coefficients(), vk.S[2], "S3")
	for i := range trace.Qcp {
		requireGPUCommitBN254(t, msm, trace.Qcp[i].Coefficients(), vk.Qcp[i], "Qcp")
	}
}

func testPlonkE2EGPUSetupBLS12377(t *testing.T, dev *gpu.Device) {
	ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, newBenchChainCircuit(128))
	require.NoError(t, err, "compiling BLS12-377 circuit should succeed")

	srs, srsLagrange := testSRSAssets(t).loadForCCS(t, ccs)
	pkI, vkI, err := gnarkplonk.Setup(ccs, srs, srsLagrange)
	require.NoError(t, err, "BLS12-377 PlonK setup should succeed")

	fullWitness, err := frontend.NewWitness(newBenchChainAssignment(ecc.BLS12_377, 128), ecc.BLS12_377.ScalarField())
	require.NoError(t, err, "creating BLS12-377 witness should succeed")
	publicWitness, err := fullWitness.Public()
	require.NoError(t, err, "extracting BLS12-377 public witness should succeed")
	proof, err := gnarkplonk.Prove(ccs, pkI, fullWitness)
	require.NoError(t, err, "BLS12-377 PlonK prove should succeed")
	require.NoError(t, gnarkplonk.Verify(proof, vkI, publicWitness), "BLS12-377 PlonK verify should succeed")

	spr := ccs.(*blscs.SparseR1CS)
	pk := pkI.(*blsplonk.ProvingKey)
	vk := vkI.(*blsplonk.VerifyingKey)
	domain := blsfft.NewDomain(uint64(spr.GetNbConstraints()+len(spr.Public)), blsfft.WithoutPrecompute())
	trace := blsplonk.NewTrace(spr, domain)

	msm, err := NewG1MSM(dev, CurveBLS12377, rawBLS12377G1Slice(pk.KzgLagrange.G1))
	require.NoError(t, err, "creating BLS12-377 resident setup MSM should succeed")
	defer msm.Close()

	requireGPUCommitBLS12377(t, msm, trace.Ql.Coefficients(), vk.Ql, "Ql")
	requireGPUCommitBLS12377(t, msm, trace.Qr.Coefficients(), vk.Qr, "Qr")
	requireGPUCommitBLS12377(t, msm, trace.Qm.Coefficients(), vk.Qm, "Qm")
	requireGPUCommitBLS12377(t, msm, trace.Qo.Coefficients(), vk.Qo, "Qo")
	requireGPUCommitBLS12377(t, msm, trace.Qk.Coefficients(), vk.Qk, "Qk")
	requireGPUCommitBLS12377(t, msm, trace.S1.Coefficients(), vk.S[0], "S1")
	requireGPUCommitBLS12377(t, msm, trace.S2.Coefficients(), vk.S[1], "S2")
	requireGPUCommitBLS12377(t, msm, trace.S3.Coefficients(), vk.S[2], "S3")
	for i := range trace.Qcp {
		requireGPUCommitBLS12377(t, msm, trace.Qcp[i].Coefficients(), vk.Qcp[i], "Qcp")
	}
}

func testPlonkE2EGPUSetupBW6761(t *testing.T, dev *gpu.Device) {
	ccs, err := frontend.Compile(ecc.BW6_761.ScalarField(), scs.NewBuilder, newBenchChainCircuit(128))
	require.NoError(t, err, "compiling BW6-761 circuit should succeed")

	srs, srsLagrange := testSRSAssets(t).loadForCCS(t, ccs)
	pkI, vkI, err := gnarkplonk.Setup(ccs, srs, srsLagrange)
	require.NoError(t, err, "BW6-761 PlonK setup should succeed")

	fullWitness, err := frontend.NewWitness(newBenchChainAssignment(ecc.BW6_761, 128), ecc.BW6_761.ScalarField())
	require.NoError(t, err, "creating BW6-761 witness should succeed")
	publicWitness, err := fullWitness.Public()
	require.NoError(t, err, "extracting BW6-761 public witness should succeed")
	proof, err := gnarkplonk.Prove(ccs, pkI, fullWitness)
	require.NoError(t, err, "BW6-761 PlonK prove should succeed")
	require.NoError(t, gnarkplonk.Verify(proof, vkI, publicWitness), "BW6-761 PlonK verify should succeed")

	spr := ccs.(*bwcs.SparseR1CS)
	pk := pkI.(*bwplonk.ProvingKey)
	vk := vkI.(*bwplonk.VerifyingKey)
	domain := bwfft.NewDomain(uint64(spr.GetNbConstraints()+len(spr.Public)), bwfft.WithoutPrecompute())
	trace := bwplonk.NewTrace(spr, domain)

	msm, err := NewG1MSM(dev, CurveBW6761, rawBW6761G1Slice(pk.KzgLagrange.G1))
	require.NoError(t, err, "creating BW6-761 resident setup MSM should succeed")
	defer msm.Close()

	requireGPUCommitBW6761(t, msm, trace.Ql.Coefficients(), vk.Ql, "Ql")
	requireGPUCommitBW6761(t, msm, trace.Qr.Coefficients(), vk.Qr, "Qr")
	requireGPUCommitBW6761(t, msm, trace.Qm.Coefficients(), vk.Qm, "Qm")
	requireGPUCommitBW6761(t, msm, trace.Qo.Coefficients(), vk.Qo, "Qo")
	requireGPUCommitBW6761(t, msm, trace.Qk.Coefficients(), vk.Qk, "Qk")
	requireGPUCommitBW6761(t, msm, trace.S1.Coefficients(), vk.S[0], "S1")
	requireGPUCommitBW6761(t, msm, trace.S2.Coefficients(), vk.S[1], "S2")
	requireGPUCommitBW6761(t, msm, trace.S3.Coefficients(), vk.S[2], "S3")
	for i := range trace.Qcp {
		requireGPUCommitBW6761(t, msm, trace.Qcp[i].Coefficients(), vk.Qcp[i], "Qcp")
	}
}

func requireGPUCommitBN254(tb testing.TB, msm *G1MSM, coeffs []bnfr.Element, want bn254.G1Affine, label string) {
	tb.Helper()
	out, err := msm.CommitRaw(cloneRaw(rawBN254(coeffs)))
	require.NoError(tb, err, "GPU BN254 setup commitment %s should succeed", label)
	requireBN254ProjectiveMatches(tb, want, out, "setup "+label)
}

func requireGPUCommitBLS12377(tb testing.TB, msm *G1MSM, coeffs []blsfr.Element, want bls12377.G1Affine, label string) {
	tb.Helper()
	out, err := msm.CommitRaw(cloneRaw(rawBLS12377(coeffs)))
	require.NoError(tb, err, "GPU BLS12-377 setup commitment %s should succeed", label)
	requireBLS12377ProjectiveMatches(tb, want, out, "setup "+label)
}

func requireGPUCommitBW6761(tb testing.TB, msm *G1MSM, coeffs []bwfr.Element, want bw6761.G1Affine, label string) {
	tb.Helper()
	out, err := msm.CommitRaw(cloneRaw(rawBW6761(coeffs)))
	require.NoError(tb, err, "GPU BW6-761 setup commitment %s should succeed", label)
	requireBW6761ProjectiveMatches(tb, want, out, "setup "+label)
}
