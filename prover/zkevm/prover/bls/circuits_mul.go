package bls

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/emulated/sw_bls12381"
	"github.com/consensys/gnark/std/algebra/emulated/sw_emulated"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/plonk"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

const (
	NAME_BLS_MSM = "BLS_MSM"
)

const (
	nbRowsPerG1Mul = nbFpLimbs + 3*nbG1Limbs // 1 scalar, 1 point, 1 previous accumulator, 1 next accumulator
	nbRowsPerG2Mul = nbFpLimbs + 3*nbG2Limbs // 1 scalar, 1 point, 1 previous accumulator, 1 next accumulator
)

type BlsMsmDataSource struct {
	CsMul  ifaces.Column
	Limb   ifaces.Column
	Index  ifaces.Column
	IsData ifaces.Column
	IsRes  ifaces.Column
}

func newMsmDataSource(comp *wizard.CompiledIOP, g group) *BlsMsmDataSource {
	var selectorCs ifaces.Column
	switch g {
	case G1:
		selectorCs = comp.Columns.GetHandle("bls.CIRCUIT_SELECTOR_BLS_G1_MSM")
	case G2:
		selectorCs = comp.Columns.GetHandle("bls.CIRCUIT_SELECTOR_BLS_G2_MSM")
	default:
		panic("unknown group for bls msm data source")
	}
	return &BlsMsmDataSource{
		CsMul:  selectorCs,
		Limb:   comp.Columns.GetHandle("bls.LIMB"),
		Index:  comp.Columns.GetHandle("bls.INDEX"),
		IsData: comp.Columns.GetHandle("bls.IS_BLS_MUL_DATA"),
		IsRes:  comp.Columns.GetHandle("bls.IS_BLS_MUL_RESULT"),
	}
}

type BlsMsm struct {
	*BlsMsmDataSource
	AlignedGnarkData *plonk.Alignment
	size             int
	*Limits
	group
}

func newMsm(comp *wizard.CompiledIOP, g group, limits *Limits, src *BlsMsmDataSource, plonkOptions []plonk.Option) *BlsMsm {
	size := limits.sizeMulIntegration(g)

	toAlign := &plonk.CircuitAlignmentInput{
		Name:               fmt.Sprintf("%s_%s", NAME_BLS_MSM, g.String()),
		Round:              ROUND_NR,
		DataToCircuitMask:  src.CsMul,
		DataToCircuit:      src.Limb,
		Circuit:            NewMulCircuit(g, limits),
		NbCircuitInstances: limits.nbMulCircuitInstances(g),
		PlonkOptions:       plonkOptions,
	}

	return &BlsMsm{
		BlsMsmDataSource: src,
		AlignedGnarkData: plonk.DefineAlignment(comp, toAlign),
		size:             size,
		Limits:           limits,
		group:            g,
	}
}

func NewG1MsmZkEvm(comp *wizard.CompiledIOP, limits *Limits) *BlsMsm {
	return newMsm(
		comp,
		G1,
		limits,
		newMsmDataSource(comp, G1),
		[]plonk.Option{plonk.WithRangecheck(16, 6, true)},
	)
}

func NewG2MsmZkEvm(comp *wizard.CompiledIOP, limits *Limits) *BlsMsm {
	return newMsm(
		comp,
		G2,
		limits,
		newMsmDataSource(comp, G2),
		[]plonk.Option{plonk.WithRangecheck(16, 6, true)},
	)
}

type MultiMulCircuit[C convertable[T], T element] struct {
	Instances []MulInstance[C, T]
}

func NewMulCircuit(g group, limits *Limits) frontend.Circuit {
	switch g {
	case G1:
		return &MultiMulCircuit[g1ElementWizard, sw_bls12381.G1Affine]{Instances: make([]MulInstance[g1ElementWizard, sw_bls12381.G1Affine], limits.NbG1MulInputInstances)}
	case G2:
		return &MultiMulCircuit[g2ElementWizard, sw_bls12381.G2Affine]{Instances: make([]MulInstance[g2ElementWizard, sw_bls12381.G2Affine], limits.NbG2MulInputInstances)}
	default:
		panic(fmt.Sprintf("unknown group %s for bls msm circuit", g.String()))
	}
}

type MulInstance[C convertable[T], T element] struct {
	CurrentAccumulator C                  `gnark:",public"`
	Scalar             sw_bls12381.Scalar `gnark:",public"`
	Point              C                  `gnark:",public"`
	NextAccumulator    C                  `gnark:",public"`
}

func (c *MultiMulCircuit[C, T]) Define(api frontend.API) error {
	f, err := emulated.NewField[sw_bls12381.BaseField](api)
	if err != nil {
		return fmt.Errorf("new field: %w", err)
	}
	nbInstances := len(c.Instances)
	Cs, Ps, Ns := make([]T, nbInstances), make([]T, nbInstances), make([]T, nbInstances)
	// TODO: move this part into the main loop below. Then we don't need to type assert
	for i := range c.Instances {
		Cs[i] = c.Instances[i].CurrentAccumulator.ToElement(api, f)
		Ps[i] = c.Instances[i].Point.ToElement(api, f)
		Ns[i] = c.Instances[i].NextAccumulator.ToElement(api, f)
	}
	var t T
	switch any(t).(type) {
	case sw_bls12381.G1Affine:
		g1, err := sw_bls12381.NewG1(api)
		if err != nil {
			return fmt.Errorf("new G1: %w", err)
		}
		curve, err := sw_emulated.New[sw_bls12381.BaseField, sw_bls12381.ScalarField](api, sw_emulated.GetBLS12381Params())
		if err != nil {
			return fmt.Errorf("new curve curve: %w", err)
		}
		for i := 0; i < nbInstances; i++ {
			tCsi := any(&Cs[i]).(*sw_bls12381.G1Affine)
			tPsi := any(&Ps[i]).(*sw_bls12381.G1Affine)
			tNsi := any(&Ns[i]).(*sw_bls12381.G1Affine)
			// we only assert that the given inputs is on the curve. As we work
			// in the MSM context, then the Current and Next accumulators are always on G1.
			g1.AssertIsOnG1(tPsi)
			res := curve.ScalarMul(tPsi, &c.Instances[i].Scalar)
			accumulated := curve.AddUnified(tCsi, res)
			curve.AssertIsEqual(accumulated, tNsi)
		}
	case sw_bls12381.G2Affine:
		g2 := sw_bls12381.NewG2(api)
		for i := 0; i < nbInstances; i++ {
			tCsi := any(&Cs[i]).(*sw_bls12381.G2Affine)
			tPsi := any(&Ps[i]).(*sw_bls12381.G2Affine)
			tNsi := any(&Ns[i]).(*sw_bls12381.G2Affine)
			// we only assert that the given inputs is on the twist. As we work
			// in the MSM context, then the Current and Next accumulators are always on G2.
			g2.AssertIsOnG2(tPsi)
			res := g2.ScalarMul(tPsi, &c.Instances[i].Scalar)
			accumulated := g2.AddUnified(tCsi, res)
			g2.AssertIsEqual(accumulated, tNsi)
		}
	default:
		return fmt.Errorf("unknown group %T for bls msm circuit", t)
	}
	return nil
}
