package plonk2

import "fmt"

// PointLayout describes the host point representation uploaded for MSM.
type PointLayout uint8

const (
	// PointLayoutAffineSW is gnark-crypto's short-Weierstrass affine layout.
	PointLayoutAffineSW PointLayout = iota + 1
	// PointLayoutTwistedEdwardsXY is the compact BLS12-377 layout used today.
	PointLayoutTwistedEdwardsXY
	// PointLayoutTwistedEdwardsPrecomputed is the larger BLS12-377
	// precomputed layout used by the current gpu/plonk experiments.
	PointLayoutTwistedEdwardsPrecomputed
)

func (l PointLayout) String() string {
	switch l {
	case PointLayoutAffineSW:
		return "affine-sw"
	case PointLayoutTwistedEdwardsXY:
		return "twisted-edwards-xy"
	case PointLayoutTwistedEdwardsPrecomputed:
		return "twisted-edwards-precomputed"
	default:
		return fmt.Sprintf("unknown point layout %d", l)
	}
}

// MSMPlanConfig describes the high-memory objects needed by a Pippenger MSM.
type MSMPlanConfig struct {
	Curve       Curve
	Points      int
	WindowBits  int
	Layout      PointLayout
	ChunkPoints int
}

// MSMMemoryPlan estimates resident and scratch memory for one MSM run.
//
// The estimate intentionally separates point/scalar residency from per-run
// sort and bucket memory. This is the key tradeoff for BW6-761: affine points
// are already 192 bytes each, and projective buckets are 288 bytes each before
// any sort buffers are counted.
type MSMMemoryPlan struct {
	Curve                  Curve
	Points                 int
	WindowBits             int
	Windows                int
	Layout                 PointLayout
	PointBytes             int
	ScalarBytes            int
	BucketBytes            int
	AssignmentCount        uint64
	ResidentBytes          uint64
	AssignmentBytes        uint64
	SortBytes              uint64
	BucketAccumulatorBytes uint64
	EstimatedTotalBytes    uint64
	ChunkPoints            int
	ChunkEstimatedBytes    uint64
}

// DefaultMSMPlanConfig returns conservative starting parameters for one curve.
func DefaultMSMPlanConfig(curve Curve, points int) (MSMPlanConfig, error) {
	info, err := curve.validate()
	if err != nil {
		return MSMPlanConfig{}, err
	}
	if points <= 0 {
		return MSMPlanConfig{}, fmt.Errorf("plonk2: point count must be positive")
	}

	cfg := MSMPlanConfig{
		Curve:  curve,
		Points: points,
		Layout: PointLayoutAffineSW,
	}
	cfg.WindowBits = defaultMSMWindowBits(info, points)
	return cfg, nil
}

// BLS12377CompactTEPlanConfig estimates the current gpu/plonk compact
// twisted-Edwards path. It is kept separate from the plonk2 default because the
// generic backend should use one affine short-Weierstrass input contract for
// every curve.
func BLS12377CompactTEPlanConfig(points int) (MSMPlanConfig, error) {
	if points <= 0 {
		return MSMPlanConfig{}, fmt.Errorf("plonk2: point count must be positive")
	}
	return MSMPlanConfig{
		Curve:      CurveBLS12377,
		Points:     points,
		WindowBits: 16,
		Layout:     PointLayoutTwistedEdwardsXY,
	}, nil
}

// PlanMSMMemory estimates the dominant MSM memory terms for cfg.
func PlanMSMMemory(cfg MSMPlanConfig) (MSMMemoryPlan, error) {
	info, err := cfg.Curve.validate()
	if err != nil {
		return MSMMemoryPlan{}, err
	}
	if cfg.Points <= 0 {
		return MSMMemoryPlan{}, fmt.Errorf("plonk2: point count must be positive")
	}
	if cfg.WindowBits <= 1 || cfg.WindowBits > 24 {
		return MSMMemoryPlan{}, fmt.Errorf("plonk2: window bits must be in [2,24]")
	}

	pointBytes, bucketBytes, err := msmPointAndBucketBytes(info, cfg.Layout)
	if err != nil {
		return MSMMemoryPlan{}, err
	}
	scalarBytes := info.ScalarLimbs * 8
	windows := msmSignedWindowCount(info.ScalarBits, cfg.WindowBits)
	assignments := uint64(cfg.Points) * uint64(windows)
	assignmentBytes := assignments * 8
	sortBytes := assignmentBytes * 2
	buckets := uint64(1) << (cfg.WindowBits - 1)
	bucketAccumulatorBytes := buckets * uint64(windows) * uint64(bucketBytes)
	residentBytes := uint64(cfg.Points) * uint64(pointBytes+scalarBytes)

	plan := MSMMemoryPlan{
		Curve:                  cfg.Curve,
		Points:                 cfg.Points,
		WindowBits:             cfg.WindowBits,
		Windows:                windows,
		Layout:                 cfg.Layout,
		PointBytes:             pointBytes,
		ScalarBytes:            scalarBytes,
		BucketBytes:            bucketBytes,
		AssignmentCount:        assignments,
		ResidentBytes:          residentBytes,
		AssignmentBytes:        assignmentBytes,
		SortBytes:              sortBytes,
		BucketAccumulatorBytes: bucketAccumulatorBytes,
		EstimatedTotalBytes:    residentBytes + assignmentBytes + sortBytes + bucketAccumulatorBytes,
		ChunkPoints:            cfg.ChunkPoints,
	}

	if cfg.ChunkPoints > 0 {
		chunkCfg := cfg
		chunkCfg.Points = cfg.ChunkPoints
		chunkCfg.ChunkPoints = 0
		chunkPlan, err := PlanMSMMemory(chunkCfg)
		if err != nil {
			return MSMMemoryPlan{}, err
		}
		plan.ChunkEstimatedBytes = chunkPlan.EstimatedTotalBytes
	}

	return plan, nil
}

func msmSignedWindowCount(scalarBits, windowBits int) int {
	// Signed window recoding can carry one bit into the next window.
	return (scalarBits + 1 + windowBits - 1) / windowBits
}

func msmPointAndBucketBytes(info CurveInfo, layout PointLayout) (pointBytes, bucketBytes int, err error) {
	baseBytes := info.BaseFieldLimbs * 8
	switch layout {
	case PointLayoutAffineSW:
		return 2 * baseBytes, 3 * baseBytes, nil
	case PointLayoutTwistedEdwardsXY:
		return 2 * baseBytes, 4 * baseBytes, nil
	case PointLayoutTwistedEdwardsPrecomputed:
		return 3 * baseBytes, 4 * baseBytes, nil
	default:
		return 0, 0, fmt.Errorf("plonk2: unsupported point layout %d", layout)
	}
}
