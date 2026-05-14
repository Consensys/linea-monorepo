package plonk2

import (
	"errors"
	"fmt"
	"log"
	"reflect"

	"github.com/consensys/gnark/backend"
	gnarkplonk "github.com/consensys/gnark/backend/plonk"
	plonk_bls12377 "github.com/consensys/gnark/backend/plonk/bls12-377"
	plonk_bn254 "github.com/consensys/gnark/backend/plonk/bn254"
	plonk_bw6761 "github.com/consensys/gnark/backend/plonk/bw6-761"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/constraint"
	cs_bls12377 "github.com/consensys/gnark/constraint/bls12-377"
	cs_bn254 "github.com/consensys/gnark/constraint/bn254"
	cs_bw6761 "github.com/consensys/gnark/constraint/bw6-761"

	"github.com/consensys/linea-monorepo/prover/gpu"
	"github.com/consensys/linea-monorepo/prover/gpu/plonk2/bls12377"
	"github.com/consensys/linea-monorepo/prover/gpu/plonk2/bn254"
	"github.com/consensys/linea-monorepo/prover/gpu/plonk2/bw6761"
)

// Prover is a multi-curve GPU PlonK prover dispatcher.
//
// CPU fallback is applied when:
//   - The GPU is disabled (WithEnabled(false))
//   - The GPU path returns an error with WithCPUFallback(true) (default)
//   - The constraint system curve is not supported
type Prover struct {
	dev *gpu.Device
	cfg proverConfig
	ccs constraint.ConstraintSystem
	pk  gnarkplonk.ProvingKey
	vk  gnarkplonk.VerifyingKey

	// per-curve GPU proving keys — at most one is non-nil
	bn254PK    *bn254.GPUProvingKey
	bls12377PK *bls12377.GPUProvingKey
	bw6761PK   *bw6761.GPUProvingKey

	closed bool
}

// NewProver creates a GPU prover for the given constraint system and proving key.
// The prover inspects the curve ID and instantiates the appropriate per-curve GPU PK.
func NewProver(dev *gpu.Device, ccs constraint.ConstraintSystem, pk gnarkplonk.ProvingKey, vk gnarkplonk.VerifyingKey, opts ...Option) (*Prover, error) {
	cfg := defaultProverConfig()
	for _, o := range opts {
		o(&cfg)
	}
	p := &Prover{dev: dev, cfg: cfg, ccs: ccs, pk: pk, vk: vk}
	if cfg.enabled {
		if err := p.initGPU(); err != nil {
			if !cfg.cpuFallback {
				return nil, fmt.Errorf("plonk2: init GPU: %w", err)
			}
			// GPU init failed but CPU fallback is enabled — continue with CPU only.
		}
	}
	return p, nil
}

func (p *Prover) initGPU() error {
	switch gpk := p.pk.(type) {
	case *plonk_bn254.ProvingKey:
		p.bn254PK = bn254.NewGPUProvingKey(gpk.Kzg.G1, p.vk.(*plonk_bn254.VerifyingKey))
	case *plonk_bls12377.ProvingKey:
		p.bls12377PK = bls12377.NewGPUProvingKey(gpk.Kzg.G1, p.vk.(*plonk_bls12377.VerifyingKey))
	case *plonk_bw6761.ProvingKey:
		p.bw6761PK = bw6761.NewGPUProvingKey(gpk.Kzg.G1, p.vk.(*plonk_bw6761.VerifyingKey))
	default:
		return fmt.Errorf("plonk2: unsupported proving key type %T", p.pk)
	}
	return nil
}

// Prove generates a PlonK proof for the given witness. When the GPU path
// is enabled and the curve is supported, dispatches to the per-curve
// gpu/plonk2/<curve>.GPUProve. On error, falls back to gnark's CPU prover
// unless WithStrictMode/WithCPUFallback(false) was set.
func (p *Prover) Prove(w witness.Witness, opts ...backend.ProverOption) (gnarkplonk.Proof, error) {
	if p.closed {
		return nil, errors.New("plonk2: prover is closed")
	}
	if p.cfg.enabled && (p.bn254PK != nil || p.bls12377PK != nil || p.bw6761PK != nil) {
		proof, err := p.proveGPU(w, opts...)
		if err == nil {
			return proof, nil
		}
		if !p.cfg.cpuFallback {
			return nil, err
		}
		log.Printf("plonk2: GPU prove failed, falling back to CPU: %v", err)
	}
	if !p.cfg.cpuFallback {
		return nil, errors.New("plonk2: GPU disabled and CPU fallback disabled")
	}
	return gnarkplonk.Prove(p.ccs, p.pk, w, opts...)
}

func (p *Prover) proveGPU(w witness.Witness, opts ...backend.ProverOption) (gnarkplonk.Proof, error) {
	switch {
	case p.bn254PK != nil:
		spr, ok := p.ccs.(*cs_bn254.SparseR1CS)
		if !ok {
			return nil, fmt.Errorf("plonk2: BN254 CCS type mismatch: got %T", p.ccs)
		}
		normalizeGkrScheduleLevels(spr.Blueprints)
		return bn254.GPUProve(p.dev, p.bn254PK, spr, w, opts...)
	case p.bls12377PK != nil:
		spr, ok := p.ccs.(*cs_bls12377.SparseR1CS)
		if !ok {
			return nil, fmt.Errorf("plonk2: BLS12-377 CCS type mismatch: got %T", p.ccs)
		}
		normalizeGkrScheduleLevels(spr.Blueprints)
		return bls12377.GPUProve(p.dev, p.bls12377PK, spr, w, opts...)
	case p.bw6761PK != nil:
		spr, ok := p.ccs.(*cs_bw6761.SparseR1CS)
		if !ok {
			return nil, fmt.Errorf("plonk2: BW6-761 CCS type mismatch: got %T", p.ccs)
		}
		normalizeGkrScheduleLevels(spr.Blueprints)
		return bw6761.GPUProve(p.dev, p.bw6761PK, spr, w, opts...)
	default:
		return nil, errors.New("plonk2: no GPU proving key initialized")
	}
}

// normalizeGkrScheduleLevels rewrites pointer-typed GKR schedule levels into
// their value-typed equivalents in-place. gnark's solver hands us blueprints
// where the GKR schedule is a slice of interface values; some of those values
// are *constraint.GkrSkipLevel etc., others are constraint.GkrSkipLevel. The
// per-curve GPU prover assumes the value form (we never need to mutate them
// after the solver hands them over), so we deref each pointer once at the
// start of Prove. This keeps the per-curve switch in proveGPU stable.

func normalizeGkrScheduleLevels(blueprints []constraint.Blueprint) {
	scheduleType := reflect.TypeFor[constraint.GkrProvingSchedule]()

	for _, blueprint := range blueprints {
		value := reflect.ValueOf(blueprint)
		if value.Kind() != reflect.Pointer || value.IsNil() {
			continue
		}

		value = value.Elem()
		if value.Kind() != reflect.Struct {
			continue
		}

		schedule := value.FieldByName("Schedule")
		if !schedule.IsValid() || !schedule.CanSet() || schedule.Type() != scheduleType {
			continue
		}

		for i := range schedule.Len() {
			level := schedule.Index(i)
			switch typed := level.Interface().(type) {
			case *constraint.GkrSkipLevel:
				if typed != nil {
					level.Set(reflect.ValueOf(*typed))
				}
			case *constraint.GkrSingleSourceZeroCheckLevel:
				if typed != nil {
					level.Set(reflect.ValueOf(*typed))
				}
			case *constraint.GkrSumcheckLevel:
				if typed != nil {
					level.Set(reflect.ValueOf(*typed))
				}
			}
		}
	}
}

// Close releases all GPU resources.
func (p *Prover) Close() {
	if p.closed {
		return
	}
	p.closed = true
	if p.bn254PK != nil {
		p.bn254PK.Close()
		p.bn254PK = nil
	}
	if p.bls12377PK != nil {
		p.bls12377PK.Close()
		p.bls12377PK = nil
	}
	if p.bw6761PK != nil {
		p.bw6761PK.Close()
		p.bw6761PK = nil
	}
}
