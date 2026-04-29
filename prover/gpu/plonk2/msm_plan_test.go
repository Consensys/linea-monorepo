package plonk2

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/stretchr/testify/require"
)

func TestPlanMSMMemory_BW6761DefaultsAvoidLargeBuckets(t *testing.T) {
	cfg, err := DefaultMSMPlanConfig(CurveBW6761, 1<<20)
	require.NoError(t, err)

	plan, err := PlanMSMMemory(cfg)
	require.NoError(t, err)

	require.Equal(t, 16, plan.WindowBits, "BW6-761 default window should balance sort and bucket memory at 1M points")
	require.Equal(t, 24, plan.Windows, "BW6-761 signed window count should include carry room")
	require.Equal(t, 192, plan.PointBytes, "BW6-761 affine points should be 2*12 limbs")
	require.Equal(t, 48, plan.ScalarBytes, "BW6-761 scalar should be 6 limbs")
	require.Equal(t, uint64(1<<20)*48, plan.ScalarStagingBytes, "scalar staging should match CUDA d_scalars")
	require.Greater(t, plan.PairBufferBytes, plan.AssignmentBytes, "key/value pair buffers should include in/out arrays")
	require.Greater(t, plan.PartialReductionBytes, uint64(0), "partial reduction buffers should be accounted")
	require.Greater(t, plan.BucketAccumulatorBytes, uint64(0), "bucket memory should be accounted")
	require.Greater(t, plan.EstimatedTotalBytes, plan.ResidentBytes, "scratch memory should be included")
}

func TestPlanMSMMemory_BLS12377DefaultIsGenericAffine(t *testing.T) {
	cfg, err := DefaultMSMPlanConfig(CurveBLS12377, 1<<16)
	require.NoError(t, err)

	plan, err := PlanMSMMemory(cfg)
	require.NoError(t, err)

	require.Equal(t, PointLayoutAffineSW, plan.Layout, "plonk2 should default to one generic affine input layout")
	require.Equal(t, 96, plan.PointBytes, "BLS12-377 affine points should be 2*6 limbs")
	require.Equal(t, 32, plan.ScalarBytes, "BLS12-377 scalar should be 4 limbs")
}

func TestPlanMSMMemory_BLS12377CompactTEOptionalPath(t *testing.T) {
	cfg, err := BLS12377CompactTEPlanConfig(1 << 16)
	require.NoError(t, err)

	plan, err := PlanMSMMemory(cfg)
	require.NoError(t, err)

	require.Equal(t, PointLayoutTwistedEdwardsXY, plan.Layout, "TE should be explicit rather than the plonk2 default")
	require.Equal(t, 96, plan.PointBytes, "compact TE points should be 2*6 limbs")
	require.Equal(t, 192, plan.BucketBytes, "extended TE buckets should use four base-field elements")
}

func TestPlanMSMMemory_ChunkEstimate(t *testing.T) {
	cfg, err := DefaultMSMPlanConfig(CurveBN254, 1<<20)
	require.NoError(t, err)
	cfg.ChunkPoints = 1 << 16

	plan, err := PlanMSMMemory(cfg)
	require.NoError(t, err)

	require.NotZero(t, plan.ChunkEstimatedBytes, "chunk estimate should be populated")
	require.Less(t, plan.ChunkEstimatedBytes, plan.EstimatedTotalBytes, "chunking should reduce peak estimate")
}

func TestPlanProverMemory_TargetCurves(t *testing.T) {
	for _, curve := range []Curve{CurveBN254, CurveBLS12377, CurveBW6761} {
		t.Run(curve.String(), func(t *testing.T) {
			plan, err := PlanProverMemory(nil, ProverMemoryPlanConfig{
				Curve:       curve,
				DomainSize:  1 << 16,
				Commitments: 8,
				PointCount:  1<<16 + 3,
			})
			require.NoError(t, err)

			require.Greater(t, plan.PersistentBytes(), uint64(0), "persistent bytes should be planned")
			require.Greater(t, plan.ScratchBytes(), uint64(0), "quotient scratch should be planned")
			require.Greater(t, plan.PerWaveBytes(), uint64(0), "MSM wave bytes should be planned")
			require.Greater(t, plan.EstimatedPeakBytes(), plan.PersistentBytes(), "peak should include scratch")
			require.Contains(t, plan.Summary(), curve.String(), "summary should include curve metadata")
		})
	}
}

func TestPlanProverMemory_Limits(t *testing.T) {
	_, err := PlanProverMemory(nil, ProverMemoryPlanConfig{
		Curve:       CurveBN254,
		DomainSize:  1 << 16,
		Commitments: 8,
		PointCount:  1<<16 + 3,
		MemoryLimit: 1,
	})
	require.Error(t, err, "tiny GPU memory limit should fail")
	require.Contains(t, err.Error(), "configured limit", "error should explain memory limit")

	_, err = PlanProverMemory(nil, ProverMemoryPlanConfig{
		Curve:           CurveBN254,
		DomainSize:      1 << 16,
		Commitments:     8,
		PointCount:      1<<16 + 3,
		PinnedHostLimit: 1,
	})
	require.Error(t, err, "tiny pinned host limit should fail")
	require.Contains(t, err.Error(), "pinned host memory", "error should explain pinned host limit")
}

func TestProverMemoryPlan_WiredIntoNewProver(t *testing.T) {
	fixture := newProverAPIFixture(t, ecc.BN254)

	prover, err := NewProver(nil, fixture.ccs, fixture.pk)
	require.NoError(t, err)
	require.NotZero(t, prover.MemoryPlan().EstimatedPeakBytes(), "prepared prover should retain selected plan")

	_, err = NewProver(nil, fixture.ccs, fixture.pk, WithMemoryLimit(1))
	require.Error(t, err, "constructor should enforce configured memory limit")
}

func TestPlanMSMMemory_BW6761WindowBoundaries(t *testing.T) {
	for _, tc := range []struct {
		points     int
		windowBits int
	}{
		{points: (1 << 18) - 1, windowBits: 13},
		{points: 1 << 18, windowBits: 16},
		{points: (1 << 22) - 1, windowBits: 16},
		{points: 1 << 22, windowBits: 18},
	} {
		cfg, err := DefaultMSMPlanConfig(CurveBW6761, tc.points)
		require.NoError(t, err)
		require.Equal(t, tc.windowBits, cfg.WindowBits, "BW6-761 window policy should be size-aware")
	}
}

func TestMSMRunPlan_TargetCurves(t *testing.T) {
	for _, tc := range []struct {
		curve      Curve
		points     int
		scalarBits int
		windowBits int
	}{
		{curve: CurveBN254, points: 1 << 16, scalarBits: 254, windowBits: 16},
		{curve: CurveBLS12377, points: 1 << 16, scalarBits: 253, windowBits: 16},
		{curve: CurveBW6761, points: 1 << 16, scalarBits: 377, windowBits: 13},
	} {
		t.Run(tc.curve.String(), func(t *testing.T) {
			plan, err := defaultMSMRunPlan(MSMRunPlanConfig{
				Curve:  tc.curve,
				Points: tc.points,
			})
			require.NoError(t, err)

			require.Equal(t, tc.scalarBits, plan.ScalarBits, "scalar bit size should follow the curve")
			require.Equal(t, tc.windowBits, plan.WindowBits, "default window policy should be selected")
			require.Equal(t, 1, plan.BatchSize, "single commitment remains the default")
			require.Equal(t, 1, plan.PrecomputeFactor, "precomputation is not enabled yet")
			require.Equal(t, 4, plan.LargeBucketFactor, "large-bucket threshold should be explicit")
			require.Equal(t, tc.points, plan.ChunkPoints, "unlimited plans should not chunk")
			require.Greater(t, plan.MemoryPlan.EstimatedTotalBytes, uint64(0), "memory plan should be attached")
		})
	}
}

func TestMSMRunPlan_ChunkPolicy(t *testing.T) {
	unlimited, err := defaultMSMRunPlan(MSMRunPlanConfig{
		Curve:  CurveBW6761,
		Points: 1 << 20,
	})
	require.NoError(t, err)

	limited, err := defaultMSMRunPlan(MSMRunPlanConfig{
		Curve:       CurveBW6761,
		Points:      1 << 20,
		MemoryLimit: unlimited.MemoryPlan.EstimatedTotalBytes / 2,
	})
	require.NoError(t, err)

	require.Less(t, limited.ChunkPoints, unlimited.Points, "memory-limited plan should choose chunking")
	require.NotZero(t, limited.MemoryPlan.ChunkEstimatedBytes, "chunk memory estimate should be retained")
	require.LessOrEqual(
		t,
		limited.MemoryPlan.ChunkEstimatedBytes,
		unlimited.MemoryPlan.EstimatedTotalBytes/2,
		"selected chunk should fit under the configured limit",
	)
}

func TestMSMRunPlan_ExplicitWindowAndBatch(t *testing.T) {
	plan, err := defaultMSMRunPlan(MSMRunPlanConfig{
		Curve:      CurveBW6761,
		Points:     1 << 20,
		WindowBits: 18,
		BatchSize:  3,
		SharedBase: true,
	})
	require.NoError(t, err)

	require.Equal(t, 18, plan.WindowBits, "explicit internal window override should be honored")
	require.Equal(t, 3, plan.BatchSize, "internal batch size should be retained")
	require.Equal(t, 3, plan.MemoryPlan.BatchSize, "memory plan should retain batch size metadata")
	require.True(t, plan.SharedBase, "shared-base flag should be retained")
}
