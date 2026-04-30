package plonk2

import (
	"fmt"
	"strings"

	"github.com/consensys/linea-monorepo/prover/gpu"
)

// ProverMemoryPlanConfig describes the high-level dimensions of a prepared
// plonk2 prover instance.
type ProverMemoryPlanConfig struct {
	Curve              Curve
	DomainSize         int
	Commitments        int
	PointCount         int
	LagrangePointCount int
	MemoryLimit        uint64
	PinnedHostLimit    uint64
}

// PreparedKeyMemoryPlan estimates persistent GPU memory owned by a prepared
// proving key, excluding MSM resident SRS points and NTT twiddles, which have
// dedicated subplans.
type PreparedKeyMemoryPlan struct {
	FixedPolynomialBytes uint64
	PermutationBytes     uint64
	TotalBytes           uint64
}

// NTTDomainMemoryPlan estimates persistent twiddle memory for one FFT domain.
type NTTDomainMemoryPlan struct {
	TwiddleBytes      uint64
	CardinalityInv    uint64
	TotalBytes        uint64
	ScalarElementSize uint64
}

// SRSResidencyMemoryPlan estimates persistent GPU memory for resident
// canonical and Lagrange SRS point tables.
type SRSResidencyMemoryPlan struct {
	CanonicalPointCount int
	LagrangePointCount  int
	PointBytes          int
	CanonicalBytes      uint64
	LagrangeBytes       uint64
	TotalBytes          uint64
}

// QuotientMemoryPlan estimates per-proof GPU working vectors used by the
// quotient and permutation phases.
type QuotientMemoryPlan struct {
	WorkingVectorCount int
	WorkingVectorBytes uint64
	TotalBytes         uint64
}

// HostPinnedMemoryPlan estimates pinned system RAM used for hot host buffers.
type HostPinnedMemoryPlan struct {
	HotBufferBytes uint64
	LimitBytes     uint64
}

// ProverMemoryPlan is the top-level conservative memory contract for a
// prepared plonk2 prover.
type ProverMemoryPlan struct {
	Curve            Curve
	DomainSize       int
	Commitments      int
	PointCount       int
	MemoryLimit      uint64
	PinnedHostLimit  uint64
	DeviceFreeBytes  uint64
	DeviceTotalBytes uint64

	PreparedKey PreparedKeyMemoryPlan
	SRS         SRSResidencyMemoryPlan
	NTTDomain   NTTDomainMemoryPlan
	Quotient    QuotientMemoryPlan
	MSM         MSMMemoryPlan
	HostPinned  HostPinnedMemoryPlan
}

// PlanProverMemory builds the memory plan and enforces configured hard limits.
func PlanProverMemory(dev *gpu.Device, cfg ProverMemoryPlanConfig) (ProverMemoryPlan, error) {
	info, err := cfg.Curve.validate()
	if err != nil {
		return ProverMemoryPlan{}, err
	}
	if !isPowerOfTwo(cfg.DomainSize) {
		return ProverMemoryPlan{}, fmt.Errorf("plonk2: domain size must be a positive power of two")
	}
	if cfg.Commitments <= 0 {
		return ProverMemoryPlan{}, fmt.Errorf("plonk2: commitment count must be positive")
	}
	if cfg.PointCount <= 0 {
		return ProverMemoryPlan{}, fmt.Errorf("plonk2: point count must be positive")
	}

	msmCfg, err := DefaultMSMPlanConfig(cfg.Curve, cfg.PointCount)
	if err != nil {
		return ProverMemoryPlan{}, err
	}
	msmPlan, err := PlanMSMMemory(msmCfg)
	if err != nil {
		return ProverMemoryPlan{}, err
	}
	lagrangePointCount := cfg.LagrangePointCount
	if lagrangePointCount == 0 {
		lagrangePointCount = cfg.DomainSize
	}
	if lagrangePointCount < 0 {
		return ProverMemoryPlan{}, fmt.Errorf("plonk2: lagrange point count must be non-negative")
	}

	scalarBytes := uint64(info.ScalarLimbs * 8)
	fixedPolynomialBytes := uint64(cfg.Commitments) * uint64(cfg.DomainSize) * scalarBytes
	permutationBytes := uint64(3 * cfg.DomainSize * 8)
	prepared := PreparedKeyMemoryPlan{
		FixedPolynomialBytes: fixedPolynomialBytes,
		PermutationBytes:     permutationBytes,
		TotalBytes:           fixedPolynomialBytes + permutationBytes,
	}
	srs := SRSResidencyMemoryPlan{
		CanonicalPointCount: cfg.PointCount,
		LagrangePointCount:  lagrangePointCount,
		PointBytes:          msmPlan.PointBytes,
		CanonicalBytes:      uint64(cfg.PointCount) * uint64(msmPlan.PointBytes),
		LagrangeBytes:       uint64(lagrangePointCount) * uint64(msmPlan.PointBytes),
	}
	srs.TotalBytes = srs.CanonicalBytes + srs.LagrangeBytes

	ntt := NTTDomainMemoryPlan{
		TwiddleBytes:      uint64(cfg.DomainSize) * scalarBytes,
		CardinalityInv:    scalarBytes,
		ScalarElementSize: scalarBytes,
	}
	ntt.TotalBytes = ntt.TwiddleBytes + ntt.CardinalityInv

	quotient := QuotientMemoryPlan{
		WorkingVectorCount: 12,
		WorkingVectorBytes: uint64(cfg.DomainSize) * scalarBytes,
	}
	quotient.TotalBytes = uint64(quotient.WorkingVectorCount) * quotient.WorkingVectorBytes

	hostPinned := HostPinnedMemoryPlan{
		HotBufferBytes: estimatePinnedHostBytes(cfg.DomainSize, scalarBytes),
		LimitBytes:     cfg.PinnedHostLimit,
	}

	plan := ProverMemoryPlan{
		Curve:           cfg.Curve,
		DomainSize:      cfg.DomainSize,
		Commitments:     cfg.Commitments,
		PointCount:      cfg.PointCount,
		MemoryLimit:     cfg.MemoryLimit,
		PinnedHostLimit: cfg.PinnedHostLimit,
		PreparedKey:     prepared,
		SRS:             srs,
		NTTDomain:       ntt,
		Quotient:        quotient,
		MSM:             msmPlan,
		HostPinned:      hostPinned,
	}

	if gpu.Enabled && dev != nil {
		free, total, err := dev.MemGetInfo()
		if err != nil {
			return ProverMemoryPlan{}, fmt.Errorf("plonk2: query device memory: %w", err)
		}
		plan.DeviceFreeBytes = free
		plan.DeviceTotalBytes = total
	}

	if cfg.MemoryLimit > 0 && plan.EstimatedPeakBytes() > cfg.MemoryLimit {
		return ProverMemoryPlan{}, fmt.Errorf(
			"plonk2: estimated GPU memory %d exceeds configured limit %d",
			plan.EstimatedPeakBytes(),
			cfg.MemoryLimit,
		)
	}
	if cfg.PinnedHostLimit > 0 && plan.HostPinnedBytes() > cfg.PinnedHostLimit {
		return ProverMemoryPlan{}, fmt.Errorf(
			"plonk2: estimated pinned host memory %d exceeds configured limit %d",
			plan.HostPinnedBytes(),
			cfg.PinnedHostLimit,
		)
	}

	return plan, nil
}

// PersistentBytes returns memory expected to remain resident for the prepared key.
func (p ProverMemoryPlan) PersistentBytes() uint64 {
	return p.PreparedKey.TotalBytes + p.SRS.TotalBytes + p.NTTDomain.TotalBytes
}

// ScratchBytes returns non-MSM per-proof GPU working memory.
func (p ProverMemoryPlan) ScratchBytes() uint64 {
	return p.Quotient.TotalBytes
}

// PerWaveBytes returns scratch memory for an MSM commitment wave.
func (p ProverMemoryPlan) PerWaveBytes() uint64 {
	return p.MSM.ScratchBytes
}

// HostPinnedBytes returns estimated pinned system RAM for host-side hot buffers.
func (p ProverMemoryPlan) HostPinnedBytes() uint64 {
	return p.HostPinned.HotBufferBytes
}

// EstimatedPeakBytes returns a conservative peak GPU estimate.
func (p ProverMemoryPlan) EstimatedPeakBytes() uint64 {
	return p.PersistentBytes() + p.ScratchBytes() + p.PerWaveBytes()
}

// Summary formats the plan for benchmark and test metadata.
func (p ProverMemoryPlan) Summary() string {
	var b strings.Builder
	fmt.Fprintf(&b, "curve=%s domain=%d commitments=%d points=%d\n", p.Curve, p.DomainSize, p.Commitments, p.PointCount)
	fmt.Fprintf(&b, "srs canonical=%d lagrange=%d bytes=%d\n",
		p.SRS.CanonicalPointCount,
		p.SRS.LagrangePointCount,
		p.SRS.TotalBytes,
	)
	fmt.Fprintf(&b, "persistent=%d scratch=%d per_wave=%d peak=%d host_pinned=%d\n",
		p.PersistentBytes(),
		p.ScratchBytes(),
		p.PerWaveBytes(),
		p.EstimatedPeakBytes(),
		p.HostPinnedBytes(),
	)
	fmt.Fprintf(&b, "msm window_bits=%d windows=%d assignments=%d\n",
		p.MSM.WindowBits,
		p.MSM.Windows,
		p.MSM.AssignmentCount,
	)
	return b.String()
}

func estimatePinnedHostBytes(domainSize int, scalarBytes uint64) uint64 {
	hSize := 4 * domainSize
	if needed := 3 * (domainSize + 2); needed > hSize {
		hSize = needed
	}
	elements := uint64(domainSize+2)*3 + uint64(domainSize+3) + uint64(hSize)
	return elements * scalarBytes
}
