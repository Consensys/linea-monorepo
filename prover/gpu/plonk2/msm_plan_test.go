package plonk2

import (
	"testing"

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
