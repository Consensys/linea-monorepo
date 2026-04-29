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
	BatchSize   int
}

// MSMRunPlanConfig is the private policy input for one resident MSM handle.
type MSMRunPlanConfig struct {
	Curve       Curve
	Points      int
	WindowBits  int
	MemoryLimit uint64
	BatchSize   int
	SharedBase  bool
}

// MSMRunPlan records internal execution policy for one MSM shape.
type MSMRunPlan struct {
	Curve             Curve
	Points            int
	ScalarBits        int
	WindowBits        int
	Windows           int
	BatchSize         int
	ChunkPoints       int
	SharedBase        bool
	PrecomputeFactor  int
	LargeBucketFactor int
	MemoryPlan        MSMMemoryPlan
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
	BatchSize              int
	WindowBits             int
	Windows                int
	Layout                 PointLayout
	PointBytes             int
	ScalarBytes            int
	BucketBytes            int
	AssignmentCount        uint64
	ResidentBytes          uint64
	ScalarStagingBytes     uint64
	AssignmentBytes        uint64
	PairBufferBytes        uint64
	SortBytes              uint64
	BucketMetadataBytes    uint64
	BucketAccumulatorBytes uint64
	WindowResultBytes      uint64
	PartialReductionBytes  uint64
	OutputBytes            uint64
	ScratchBytes           uint64
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
		Curve:     curve,
		Points:    points,
		Layout:    PointLayoutAffineSW,
		BatchSize: 1,
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
		BatchSize:  1,
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
	batchSize := cfg.BatchSize
	if batchSize <= 0 {
		batchSize = 1
	}

	pointBytes, bucketBytes, err := msmPointAndBucketBytes(info, cfg.Layout)
	if err != nil {
		return MSMMemoryPlan{}, err
	}
	scalarBytes := info.ScalarLimbs * 8
	outputBytes := info.BaseFieldLimbs * 3 * 8
	windows := msmSignedWindowCount(info.ScalarBits, cfg.WindowBits)
	assignments := uint64(cfg.Points) * uint64(windows)
	assignmentBytes := assignments * 8
	pairBufferBytes := assignments * 4 * 4
	sortBytes := pairBufferBytes
	buckets := uint64(1) << (cfg.WindowBits - 1)
	totalBuckets := buckets * uint64(windows)
	reduceBPW := uint64(reduceBlocksPerWindow(windows, int(buckets)))
	totalPartials := uint64(windows) * reduceBPW
	bucketMetadataBytes := totalBuckets * 2 * 4
	bucketAccumulatorBytes := buckets * uint64(windows) * uint64(bucketBytes)
	windowResultBytes := uint64(windows) * uint64(bucketBytes)
	partialReductionBytes := totalPartials * uint64(bucketBytes) * 2
	scalarStagingBytes := uint64(cfg.Points) * uint64(scalarBytes)
	residentBytes := uint64(cfg.Points) * uint64(pointBytes)
	scratchBytes := scalarStagingBytes +
		uint64(outputBytes) +
		pairBufferBytes +
		sortBytes +
		bucketMetadataBytes +
		bucketAccumulatorBytes +
		windowResultBytes +
		partialReductionBytes

	plan := MSMMemoryPlan{
		Curve:                  cfg.Curve,
		Points:                 cfg.Points,
		BatchSize:              batchSize,
		WindowBits:             cfg.WindowBits,
		Windows:                windows,
		Layout:                 cfg.Layout,
		PointBytes:             pointBytes,
		ScalarBytes:            scalarBytes,
		BucketBytes:            bucketBytes,
		AssignmentCount:        assignments,
		ResidentBytes:          residentBytes,
		ScalarStagingBytes:     scalarStagingBytes,
		AssignmentBytes:        assignmentBytes,
		PairBufferBytes:        pairBufferBytes,
		SortBytes:              sortBytes,
		BucketMetadataBytes:    bucketMetadataBytes,
		BucketAccumulatorBytes: bucketAccumulatorBytes,
		WindowResultBytes:      windowResultBytes,
		PartialReductionBytes:  partialReductionBytes,
		OutputBytes:            uint64(outputBytes),
		ScratchBytes:           scratchBytes,
		EstimatedTotalBytes:    residentBytes + scratchBytes,
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

func defaultMSMRunPlan(cfg MSMRunPlanConfig) (MSMRunPlan, error) {
	info, err := cfg.Curve.validate()
	if err != nil {
		return MSMRunPlan{}, err
	}
	if cfg.Points <= 0 {
		return MSMRunPlan{}, fmt.Errorf("plonk2: point count must be positive")
	}
	windowBits := cfg.WindowBits
	if windowBits == 0 {
		windowBits = defaultMSMWindowBits(info, cfg.Points)
	}
	msmCfg := MSMPlanConfig{
		Curve:      cfg.Curve,
		Points:     cfg.Points,
		WindowBits: windowBits,
		Layout:     PointLayoutAffineSW,
		BatchSize:  cfg.BatchSize,
	}
	mem, err := PlanMSMMemory(msmCfg)
	if err != nil {
		return MSMRunPlan{}, err
	}
	chunkPoints := cfg.Points
	if cfg.MemoryLimit > 0 && mem.EstimatedTotalBytes > cfg.MemoryLimit {
		chunkPoints, mem, err = planMSMChunk(cfg, msmCfg)
		if err != nil {
			return MSMRunPlan{}, err
		}
	}
	batchSize := cfg.BatchSize
	if batchSize <= 0 {
		batchSize = 1
	}
	return MSMRunPlan{
		Curve:             cfg.Curve,
		Points:            cfg.Points,
		ScalarBits:        info.ScalarBits,
		WindowBits:        windowBits,
		Windows:           mem.Windows,
		BatchSize:         batchSize,
		ChunkPoints:       chunkPoints,
		SharedBase:        cfg.SharedBase,
		PrecomputeFactor:  1,
		LargeBucketFactor: 4,
		MemoryPlan:        mem,
	}, nil
}

func planMSMChunk(cfg MSMRunPlanConfig, msmCfg MSMPlanConfig) (int, MSMMemoryPlan, error) {
	for chunk := cfg.Points / 2; chunk > 0; chunk /= 2 {
		msmCfg.ChunkPoints = chunk
		plan, err := PlanMSMMemory(msmCfg)
		if err != nil {
			return 0, MSMMemoryPlan{}, err
		}
		if plan.ChunkEstimatedBytes <= cfg.MemoryLimit {
			return chunk, plan, nil
		}
	}
	return 0, MSMMemoryPlan{}, fmt.Errorf("plonk2: MSM memory exceeds configured limit %d and no chunk fits", cfg.MemoryLimit)
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

func reduceBlocksPerWindow(numWindows, numBuckets int) int {
	if numWindows >= 24 || numBuckets >= 1<<17 {
		return 16
	}
	if numBuckets >= 1<<15 {
		return 8
	}
	return 4
}
