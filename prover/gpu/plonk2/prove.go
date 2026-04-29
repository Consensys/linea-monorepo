package plonk2

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/consensys/gnark/backend"
	gnarkplonk "github.com/consensys/gnark/backend/plonk"
	plonkbls12377 "github.com/consensys/gnark/backend/plonk/bls12-377"
	plonkbn254 "github.com/consensys/gnark/backend/plonk/bn254"
	plonkbw6761 "github.com/consensys/gnark/backend/plonk/bw6-761"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/constraint"
	csbls12377 "github.com/consensys/gnark/constraint/bls12-377"
	csbn254 "github.com/consensys/gnark/constraint/bn254"
	csbw6761 "github.com/consensys/gnark/constraint/bw6-761"

	"github.com/consensys/linea-monorepo/prover/gpu"
)

var (
	errGPUProverNotWired = errors.New("plonk2: GPU prover not wired yet")
	errProverClosed      = errors.New("plonk2: prover is closed")
)

// Prover owns prepared state for the curve-generic plonk2 prover.
//
// This milestone keeps proof generation on gnark's CPU implementation through
// the CPU fallback policy. Later milestones will populate GPU-resident proving
// key state behind this API.
type Prover struct {
	dev *gpu.Device
	ccs constraint.ConstraintSystem
	pk  gnarkplonk.ProvingKey

	curve      Curve
	cfg        proverConfig
	memoryPlan ProverMemoryPlan

	closed bool
}

// NewProver validates the circuit/proving-key pair and prepares a plonk2
// prover handle. CPU fallback is enabled by default.
func NewProver(
	dev *gpu.Device,
	ccs constraint.ConstraintSystem,
	pk gnarkplonk.ProvingKey,
	opts ...Option,
) (*Prover, error) {
	ccsCurve, err := curveFromConstraintSystem(ccs)
	if err != nil {
		return nil, err
	}
	pkCurve, err := curveFromProvingKey(pk)
	if err != nil {
		return nil, err
	}
	if ccsCurve != pkCurve {
		return nil, fmt.Errorf(
			"plonk2: constraint system curve %s does not match proving key curve %s",
			ccsCurve,
			pkCurve,
		)
	}

	cfg := defaultProverConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	keyInfo, err := provingKeyInfo(pk)
	if err != nil {
		return nil, err
	}
	memoryPlan, err := PlanProverMemory(dev, ProverMemoryPlanConfig{
		Curve:           ccsCurve,
		DomainSize:      keyInfo.domainSize,
		Commitments:     keyInfo.commitments,
		PointCount:      keyInfo.pointCount,
		MemoryLimit:     cfg.memoryLimit,
		PinnedHostLimit: cfg.pinnedHostLimit,
	})
	if err != nil {
		traceProverEvent(cfg.tracePath, "prepare_error", ccsCurve, ProverMemoryPlan{}, map[string]any{
			"error": err.Error(),
		})
		return nil, err
	}
	traceProverEvent(cfg.tracePath, "prepare", ccsCurve, memoryPlan, map[string]any{
		"enabled":      cfg.enabled,
		"cpu_fallback": cfg.cpuFallback,
		"strict":       cfg.strict,
	})

	return &Prover{
		dev:        dev,
		ccs:        ccs,
		pk:         pk,
		curve:      ccsCurve,
		cfg:        cfg,
		memoryPlan: memoryPlan,
	}, nil
}

// Prove returns a PlonK proof for fullWitness.
//
// Until the GPU prover is wired, this delegates to gnark's CPU prover when CPU
// fallback is enabled. If fallback is disabled, it returns a clear readiness
// error instead of silently taking the CPU path.
func (p *Prover) Prove(
	fullWitness witness.Witness,
	opts ...backend.ProverOption,
) (gnarkplonk.Proof, error) {
	if p == nil || p.closed {
		return nil, errProverClosed
	}
	if !p.cfg.enabled {
		if p.cfg.strict || !p.cfg.cpuFallback {
			traceProverEvent(p.cfg.tracePath, "fallback_rejected", p.curve, p.memoryPlan, map[string]any{
				"reason": "gpu_disabled",
			})
			return nil, errGPUProverNotWired
		}
		traceProverEvent(p.cfg.tracePath, "fallback", p.curve, p.memoryPlan, map[string]any{
			"reason": "gpu_disabled",
		})
		return gnarkplonk.Prove(p.ccs, p.pk, fullWitness, opts...)
	}
	if p.cfg.strict || !p.cfg.cpuFallback {
		traceProverEvent(p.cfg.tracePath, "fallback_rejected", p.curve, p.memoryPlan, map[string]any{
			"reason": "gpu_not_wired",
		})
		return nil, errGPUProverNotWired
	}
	traceProverEvent(p.cfg.tracePath, "fallback", p.curve, p.memoryPlan, map[string]any{
		"reason": "gpu_not_wired",
	})
	return gnarkplonk.Prove(p.ccs, p.pk, fullWitness, opts...)
}

// Close releases prepared prover resources. It is safe to call multiple times.
func (p *Prover) Close() error {
	if p == nil || p.closed {
		return nil
	}
	p.closed = true
	return nil
}

// MemoryPlan returns the prepared prover's memory plan.
func (p *Prover) MemoryPlan() ProverMemoryPlan {
	if p == nil {
		return ProverMemoryPlan{}
	}
	return p.memoryPlan
}

// Prove prepares a temporary Prover, proves fullWitness, and closes the
// prepared state before returning.
func Prove(
	dev *gpu.Device,
	ccs constraint.ConstraintSystem,
	pk gnarkplonk.ProvingKey,
	fullWitness witness.Witness,
	opts ...backend.ProverOption,
) (gnarkplonk.Proof, error) {
	p, err := NewProver(dev, ccs, pk)
	if err != nil {
		return nil, err
	}
	defer p.Close()
	return p.Prove(fullWitness, opts...)
}

func curveFromConstraintSystem(ccs constraint.ConstraintSystem) (Curve, error) {
	switch ccs.(type) {
	case *csbn254.SparseR1CS:
		return CurveBN254, nil
	case *csbls12377.SparseR1CS:
		return CurveBLS12377, nil
	case *csbw6761.SparseR1CS:
		return CurveBW6761, nil
	default:
		return 0, fmt.Errorf("plonk2: unsupported constraint system type %T", ccs)
	}
}

func curveFromProvingKey(pk gnarkplonk.ProvingKey) (Curve, error) {
	switch pk.(type) {
	case *plonkbn254.ProvingKey:
		return CurveBN254, nil
	case *plonkbls12377.ProvingKey:
		return CurveBLS12377, nil
	case *plonkbw6761.ProvingKey:
		return CurveBW6761, nil
	default:
		return 0, fmt.Errorf("plonk2: unsupported proving key type %T", pk)
	}
}

type plonkKeyInfo struct {
	domainSize  int
	commitments int
	pointCount  int
}

func provingKeyInfo(pk gnarkplonk.ProvingKey) (plonkKeyInfo, error) {
	switch p := pk.(type) {
	case *plonkbn254.ProvingKey:
		return plonkKeyInfo{
			domainSize:  int(p.Vk.Size),
			commitments: 8 + len(p.Vk.Qcp),
			pointCount:  len(p.Kzg.G1),
		}, nil
	case *plonkbls12377.ProvingKey:
		return plonkKeyInfo{
			domainSize:  int(p.Vk.Size),
			commitments: 8 + len(p.Vk.Qcp),
			pointCount:  len(p.Kzg.G1),
		}, nil
	case *plonkbw6761.ProvingKey:
		return plonkKeyInfo{
			domainSize:  int(p.Vk.Size),
			commitments: 8 + len(p.Vk.Qcp),
			pointCount:  len(p.Kzg.G1),
		}, nil
	default:
		return plonkKeyInfo{}, fmt.Errorf("plonk2: unsupported proving key type %T", pk)
	}
}

func traceProverEvent(path, phase string, curve Curve, plan ProverMemoryPlan, extra map[string]any) {
	if path == "" {
		return
	}
	rec := map[string]any{
		"ts":    time.Now().UTC().Format(time.RFC3339Nano),
		"event": "plonk2_prover",
		"phase": phase,
		"curve": curve.String(),
	}
	if plan.DomainSize > 0 {
		rec["domain_size"] = plan.DomainSize
		rec["commitments"] = plan.Commitments
		rec["point_count"] = plan.PointCount
		rec["peak_bytes"] = plan.EstimatedPeakBytes()
		rec["host_pinned_bytes"] = plan.HostPinnedBytes()
		rec["msm_window_bits"] = plan.MSM.WindowBits
		rec["msm_chunk_points"] = plan.MSM.ChunkPoints
	}
	for k, v := range extra {
		rec[k] = v
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	_ = json.NewEncoder(f).Encode(rec)
}
