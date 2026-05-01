package plonk2

import (
	"errors"
	"fmt"

	curve_bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377"
	curve_bn254 "github.com/consensys/gnark-crypto/ecc/bn254"
	curve_bw6761 "github.com/consensys/gnark-crypto/ecc/bw6-761"

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
		pts := make([]curve_bn254.G1Affine, len(gpk.Kzg.G1))
		copy(pts, gpk.Kzg.G1)
		p.bn254PK = bn254.NewGPUProvingKey(pts, p.vk.(*plonk_bn254.VerifyingKey))
	case *plonk_bls12377.ProvingKey:
		pts := make([]curve_bls12377.G1Affine, len(gpk.Kzg.G1))
		copy(pts, gpk.Kzg.G1)
		p.bls12377PK = bls12377.NewGPUProvingKey(pts, p.vk.(*plonk_bls12377.VerifyingKey))
	case *plonk_bw6761.ProvingKey:
		pts := make([]curve_bw6761.G1Affine, len(gpk.Kzg.G1))
		copy(pts, gpk.Kzg.G1)
		p.bw6761PK = bw6761.NewGPUProvingKey(pts, p.vk.(*plonk_bw6761.VerifyingKey))
	default:
		return fmt.Errorf("plonk2: unsupported proving key type %T", p.pk)
	}
	return nil
}

// Prove generates a PlonK proof for the given witness.
func (p *Prover) Prove(w witness.Witness, opts ...backend.ProverOption) (gnarkplonk.Proof, error) {
	if p.closed {
		return nil, errors.New("plonk2: prover is closed")
	}
	if p.cfg.enabled && (p.bn254PK != nil || p.bls12377PK != nil || p.bw6761PK != nil) {
		proof, err := p.proveGPU(w)
		if err == nil {
			return proof, nil
		}
		if !p.cfg.cpuFallback {
			return nil, err
		}
	}
	if !p.cfg.cpuFallback {
		return nil, errors.New("plonk2: GPU disabled and CPU fallback disabled")
	}
	return gnarkplonk.Prove(p.ccs, p.pk, w, opts...)
}

func (p *Prover) proveGPU(w witness.Witness) (gnarkplonk.Proof, error) {
	switch {
	case p.bn254PK != nil:
		spr, ok := p.ccs.(*cs_bn254.SparseR1CS)
		if !ok {
			return nil, fmt.Errorf("plonk2: BN254 CCS type mismatch: got %T", p.ccs)
		}
		return bn254.GPUProve(p.dev, p.bn254PK, spr, w)
	case p.bls12377PK != nil:
		spr, ok := p.ccs.(*cs_bls12377.SparseR1CS)
		if !ok {
			return nil, fmt.Errorf("plonk2: BLS12-377 CCS type mismatch: got %T", p.ccs)
		}
		return bls12377.GPUProve(p.dev, p.bls12377PK, spr, w)
	case p.bw6761PK != nil:
		spr, ok := p.ccs.(*cs_bw6761.SparseR1CS)
		if !ok {
			return nil, fmt.Errorf("plonk2: BW6-761 CCS type mismatch: got %T", p.ccs)
		}
		return bw6761.GPUProve(p.dev, p.bw6761PK, spr, w)
	default:
		return nil, errors.New("plonk2: no GPU proving key initialized")
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
