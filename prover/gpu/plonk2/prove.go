package plonk2

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
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
// With the CUDA build and WithEnabled(true), Prover prepares curve-generic
// GPU-resident state for the proving key and uses it for BN254, BLS12-377, and
// BW6-761 proof generation. Unsupported build or configuration paths use the
// configured fallback policy.
type Prover struct {
	dev *gpu.Device
	ccs constraint.ConstraintSystem
	pk  gnarkplonk.ProvingKey

	curve         Curve
	cfg           proverConfig
	memoryPlan    ProverMemoryPlan
	gpuBackend    proverGPUBackend
	gpuPrepareErr error
	genericState  *genericProverState

	closed bool
}

type proverGPUBackend interface {
	Label() string
	Prove(*gpu.Device, witness.Witness, ...backend.ProverOption) (gnarkplonk.Proof, error)
	Close() error
}

type genericGPUBackend struct {
	state *genericProverState
	ccs   constraint.ConstraintSystem
	pk    gnarkplonk.ProvingKey
}

func (b *genericGPUBackend) Label() string {
	return "generic_gpu"
}

func (b *genericGPUBackend) Prove(
	_ *gpu.Device,
	fullWitness witness.Witness,
	opts ...backend.ProverOption,
) (gnarkplonk.Proof, error) {
	if b == nil || b.state == nil {
		return nil, errGPUProverNotWired
	}
	return proveGenericGPUBackend(b, fullWitness, opts...)
}

func (b *genericGPUBackend) Close() error {
	return nil
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
	unbind, err := bindProverThread(dev)
	if err != nil {
		return nil, err
	}
	defer unbind()
	memoryPlan, err := PlanProverMemory(dev, ProverMemoryPlanConfig{
		Curve:              ccsCurve,
		DomainSize:         keyInfo.domainSize,
		Commitments:        keyInfo.commitments,
		PointCount:         keyInfo.pointCount,
		LagrangePointCount: keyInfo.lagrangePointCount,
		MemoryLimit:        cfg.memoryLimit,
		PinnedHostLimit:    cfg.pinnedHostLimit,
	})
	if err != nil {
		traceProverEvent(cfg.tracePath, "prepare_error", ccsCurve, ProverMemoryPlan{}, map[string]any{
			"error": err.Error(),
		})
		return nil, err
	}
	traceProverEvent(cfg.tracePath, "prepare", ccsCurve, memoryPlan, map[string]any{
		"enabled":        cfg.enabled,
		"cpu_fallback":   cfg.cpuFallback,
		"strict":         cfg.strict,
		"legacy_bls_gpu": cfg.legacyBLSGPU,
	})

	var gpuBackend proverGPUBackend
	var gpuPrepareErr error
	var genericState *genericProverState
	if cfg.enabled && cfg.legacyBLSGPU {
		gpuBackend, gpuPrepareErr = newLegacyProverGPUBackend(dev, ccs, pk)
		if gpuPrepareErr != nil {
			traceProverEvent(cfg.tracePath, "legacy_gpu_prepare_error", ccsCurve, memoryPlan, map[string]any{
				"error": gpuPrepareErr.Error(),
			})
			if cfg.strict || !cfg.cpuFallback {
				return nil, gpuPrepareErr
			}
		} else {
			traceProverEvent(cfg.tracePath, "legacy_gpu_prepare", ccsCurve, memoryPlan, map[string]any{
				"backend": gpuBackend.Label(),
			})
		}
	} else if cfg.enabled {
		genericState, gpuPrepareErr = newGenericProverState(dev, ccs, pk)
		if gpuPrepareErr != nil {
			traceProverEvent(cfg.tracePath, "generic_gpu_prepare_error", ccsCurve, memoryPlan, map[string]any{
				"error": gpuPrepareErr.Error(),
			})
			if cfg.strict || !cfg.cpuFallback {
				return nil, gpuPrepareErr
			}
		} else {
			gpuBackend = &genericGPUBackend{state: genericState, ccs: ccs, pk: pk}
			traceProverEvent(cfg.tracePath, "generic_gpu_prepare", ccsCurve, memoryPlan, map[string]any{
				"backend":           gpuBackend.Label(),
				"fixed_polynomials": genericState.fixedPolynomialCount(),
			})
		}
	}

	return &Prover{
		dev:           dev,
		ccs:           ccs,
		pk:            pk,
		curve:         ccsCurve,
		cfg:           cfg,
		memoryPlan:    memoryPlan,
		gpuBackend:    gpuBackend,
		gpuPrepareErr: gpuPrepareErr,
		genericState:  genericState,
	}, nil
}

// Prove returns a PlonK proof for fullWitness.
//
// When the CUDA generic GPU backend is enabled, Prove returns a typed gnark
// PlonK proof assembled from GPU commitments and quotient work. When no GPU
// backend is available, CPU fallback is used according to the prover
// configuration; strict mode returns the GPU error instead.
func (p *Prover) Prove(
	fullWitness witness.Witness,
	opts ...backend.ProverOption,
) (gnarkplonk.Proof, error) {
	if p == nil || p.closed {
		return nil, errProverClosed
	}
	unbind, err := bindProverThread(p.dev)
	if err != nil {
		return nil, err
	}
	defer unbind()
	if !p.cfg.enabled {
		if p.cfg.strict || !p.cfg.cpuFallback {
			traceProverEvent(p.cfg.tracePath, "cpu_fallback_rejected", p.curve, p.memoryPlan, map[string]any{
				"reason": "gpu_disabled",
			})
			return nil, errGPUProverNotWired
		}
		traceProverEvent(p.cfg.tracePath, "cpu_fallback", p.curve, p.memoryPlan, map[string]any{
			"reason": "gpu_disabled",
		})
		return gnarkplonk.Prove(p.ccs, p.pk, fullWitness, opts...)
	}

	if p.gpuBackend != nil {
		backendLabel := p.gpuBackend.Label()
		traceProverEvent(p.cfg.tracePath, backendLabel+"_prove", p.curve, p.memoryPlan, map[string]any{
			"backend": backendLabel,
		})
		proof, err := p.gpuBackend.Prove(p.dev, fullWitness, opts...)
		if err == nil {
			traceProverEvent(p.cfg.tracePath, backendLabel+"_prove_done", p.curve, p.memoryPlan, map[string]any{
				"backend": backendLabel,
			})
			return proof, nil
		}
		traceProverEvent(p.cfg.tracePath, backendLabel+"_error", p.curve, p.memoryPlan, map[string]any{
			"backend": backendLabel,
			"error":   err.Error(),
		})
		if p.cfg.strict || !p.cfg.cpuFallback {
			return nil, err
		}
		traceProverEvent(p.cfg.tracePath, "cpu_fallback", p.curve, p.memoryPlan, map[string]any{
			"reason": "gpu_error",
			"error":  err.Error(),
		})
		return gnarkplonk.Prove(p.ccs, p.pk, fullWitness, opts...)
	}

	if p.cfg.strict || !p.cfg.cpuFallback {
		if p.gpuPrepareErr != nil {
			traceProverEvent(p.cfg.tracePath, "cpu_fallback_rejected", p.curve, p.memoryPlan, map[string]any{
				"reason": "gpu_prepare_error",
				"error":  p.gpuPrepareErr.Error(),
			})
			return nil, p.gpuPrepareErr
		}
		traceProverEvent(p.cfg.tracePath, "cpu_fallback_rejected", p.curve, p.memoryPlan, map[string]any{
			"reason": "gpu_not_wired",
		})
		return nil, errGPUProverNotWired
	}
	if p.gpuPrepareErr != nil {
		traceProverEvent(p.cfg.tracePath, "cpu_fallback", p.curve, p.memoryPlan, map[string]any{
			"reason": "gpu_prepare_error",
			"error":  p.gpuPrepareErr.Error(),
		})
		return gnarkplonk.Prove(p.ccs, p.pk, fullWitness, opts...)
	}
	traceProverEvent(p.cfg.tracePath, "cpu_fallback", p.curve, p.memoryPlan, map[string]any{
		"reason": "gpu_not_wired",
	})
	return gnarkplonk.Prove(p.ccs, p.pk, fullWitness, opts...)
}

// Close releases prepared prover resources. It is safe to call multiple times.
func (p *Prover) Close() error {
	if p == nil || p.closed {
		return nil
	}
	if p.gpuBackend != nil {
		if err := p.gpuBackend.Close(); err != nil {
			return err
		}
		p.gpuBackend = nil
	}
	if p.genericState != nil {
		p.genericState.Close()
		p.genericState = nil
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
	domainSize         int
	commitments        int
	pointCount         int
	lagrangePointCount int
}

func provingKeyInfo(pk gnarkplonk.ProvingKey) (plonkKeyInfo, error) {
	switch p := pk.(type) {
	case *plonkbn254.ProvingKey:
		return plonkKeyInfo{
			domainSize:         int(p.Vk.Size),
			commitments:        8 + len(p.Vk.Qcp),
			pointCount:         len(p.Kzg.G1),
			lagrangePointCount: len(p.KzgLagrange.G1),
		}, nil
	case *plonkbls12377.ProvingKey:
		return plonkKeyInfo{
			domainSize:         int(p.Vk.Size),
			commitments:        8 + len(p.Vk.Qcp),
			pointCount:         len(p.Kzg.G1),
			lagrangePointCount: len(p.KzgLagrange.G1),
		}, nil
	case *plonkbw6761.ProvingKey:
		return plonkKeyInfo{
			domainSize:         int(p.Vk.Size),
			commitments:        8 + len(p.Vk.Qcp),
			pointCount:         len(p.Kzg.G1),
			lagrangePointCount: len(p.KzgLagrange.G1),
		}, nil
	default:
		return plonkKeyInfo{}, fmt.Errorf("plonk2: unsupported proving key type %T", pk)
	}
}

func bindProverThread(dev *gpu.Device) (func(), error) {
	if dev == nil || dev.Handle() == nil {
		return func() {}, nil
	}
	runtime.LockOSThread()
	if err := dev.Bind(); err != nil {
		runtime.UnlockOSThread()
		return nil, fmt.Errorf("plonk2: bind CUDA device %d: %w", dev.DeviceID(), err)
	}
	return runtime.UnlockOSThread, nil
}

func genericPlainSolvePreflight(ccs constraint.ConstraintSystem, fullWitness witness.Witness) error {
	if hasPlonkCommitments(ccs) {
		return nil
	}
	switch spr := ccs.(type) {
	case *csbn254.SparseR1CS:
		_, err := spr.Solve(fullWitness)
		return err
	case *csbls12377.SparseR1CS:
		_, err := spr.Solve(fullWitness)
		return err
	case *csbw6761.SparseR1CS:
		_, err := spr.Solve(fullWitness)
		return err
	default:
		return fmt.Errorf("plonk2: unsupported constraint system type %T", ccs)
	}
}

func hasPlonkCommitments(ccs constraint.ConstraintSystem) bool {
	switch spr := ccs.(type) {
	case *csbn254.SparseR1CS:
		return plonkCommitmentCount(spr.CommitmentInfo) > 0
	case *csbls12377.SparseR1CS:
		return plonkCommitmentCount(spr.CommitmentInfo) > 0
	case *csbw6761.SparseR1CS:
		return plonkCommitmentCount(spr.CommitmentInfo) > 0
	default:
		return false
	}
}

func plonkCommitmentCount(info constraint.Commitments) int {
	commitments, ok := info.(constraint.PlonkCommitments)
	if !ok {
		return 0
	}
	return len(commitments)
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
		rec["srs_bytes"] = plan.SRS.TotalBytes
		rec["srs_canonical_points"] = plan.SRS.CanonicalPointCount
		rec["srs_lagrange_points"] = plan.SRS.LagrangePointCount
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
