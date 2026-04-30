//go:build cuda

package plonk2

import (
	"fmt"

	"github.com/consensys/gnark/backend"
	gnarkplonk "github.com/consensys/gnark/backend/plonk"
	plonkbls12377 "github.com/consensys/gnark/backend/plonk/bls12-377"
	"github.com/consensys/gnark/backend/witness"
	csbls12377 "github.com/consensys/gnark/constraint/bls12-377"

	"github.com/consensys/linea-monorepo/prover/gpu"
	legacyplonk "github.com/consensys/linea-monorepo/prover/gpu/plonk"
)

func newLegacyProverGPUBackend(
	dev *gpu.Device,
	ccs any,
	pk gnarkplonk.ProvingKey,
) (proverGPUBackend, error) {
	if dev == nil || dev.Handle() == nil {
		return nil, gpu.ErrDeviceClosed
	}

	spr, ok := ccs.(*csbls12377.SparseR1CS)
	if !ok {
		return nil, errGPUProverNotWired
	}
	typedPK, ok := pk.(*plonkbls12377.ProvingKey)
	if !ok {
		return nil, fmt.Errorf("plonk2: unsupported BLS12-377 proving key type %T", pk)
	}

	gpk := legacyplonk.NewGPUProvingKey(
		legacyplonk.ConvertG1AffineToTE(typedPK.Kzg.G1),
		typedPK.Vk,
	)
	if err := gpk.Prepare(dev, spr); err != nil {
		gpk.Close()
		return nil, err
	}
	return &legacyBLS12377Backend{
		spr: spr,
		gpk: gpk,
	}, nil
}

type legacyBLS12377Backend struct {
	spr *csbls12377.SparseR1CS
	gpk *legacyplonk.GPUProvingKey
}

func (b *legacyBLS12377Backend) Label() string {
	return "legacy_bls12_377_gpu"
}

func (b *legacyBLS12377Backend) Prove(
	dev *gpu.Device,
	fullWitness witness.Witness,
	opts ...backend.ProverOption,
) (gnarkplonk.Proof, error) {
	if len(opts) != 0 {
		return nil, fmt.Errorf("plonk2: legacy BLS12-377 GPU backend does not support prover options")
	}
	return legacyplonk.GPUProve(dev, b.gpk, b.spr, fullWitness)
}

func (b *legacyBLS12377Backend) Close() error {
	if b == nil || b.gpk == nil {
		return nil
	}
	b.gpk.Close()
	b.gpk = nil
	return nil
}
