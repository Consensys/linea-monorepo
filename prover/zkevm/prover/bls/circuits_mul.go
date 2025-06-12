package bls

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/emulated/sw_bls12381"
	"github.com/consensys/gnark/std/evmprecompiles"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/plonk"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
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

func newMsm(comp *wizard.CompiledIOP, g group, limits *Limits, src *BlsMsmDataSource, plonkOptions []query.PlonkOption) *BlsMsm {
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

func (bm *BlsMsm) Assign(run *wizard.ProverRuntime) {
	bm.AlignedGnarkData.Assign(run)
}

func NewG1MsmZkEvm(comp *wizard.CompiledIOP, limits *Limits) *BlsMsm {
	return newMsm(
		comp,
		G1,
		limits,
		newMsmDataSource(comp, G1),
		[]query.PlonkOption{query.PlonkRangeCheckOption(16, 6, true)},
	)
}

func NewG2MsmZkEvm(comp *wizard.CompiledIOP, limits *Limits) *BlsMsm {
	return newMsm(
		comp,
		G2,
		limits,
		newMsmDataSource(comp, G2),
		[]query.PlonkOption{query.PlonkRangeCheckOption(16, 6, true)},
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
	switch vv := any(c.Instances).(type) {
	case []MulInstance[g1ElementWizard, sw_bls12381.G1Affine]:
		for i := 0; i < nbInstances; i++ {
			tCurrent := vv[i].CurrentAccumulator.ToElement(api, f)
			tScalar := &c.Instances[i].Scalar
			tPoint := vv[i].Point.ToElement(api, f)
			tNext := vv[i].NextAccumulator.ToElement(api, f)
			if err := evmprecompiles.ECG1ScalarMulSumBLS(api, &tCurrent, &tPoint, tScalar, &tNext); err != nil {
				return fmt.Errorf("instance %d scalar mul sum: %w", i, err)
			}
		}
	case []MulInstance[g2ElementWizard, sw_bls12381.G2Affine]:
		for i := 0; i < nbInstances; i++ {
			tCurrent := vv[i].CurrentAccumulator.ToElement(api, f)
			tScalar := &c.Instances[i].Scalar
			tPoint := vv[i].Point.ToElement(api, f)
			tNext := vv[i].NextAccumulator.ToElement(api, f)
			if err := evmprecompiles.ECG2ScalarMulSumBLS(api, &tCurrent, &tPoint, tScalar, &tNext); err != nil {
				return fmt.Errorf("instance %d scalar mul sum: %w", i, err)
			}
		}
	default:
		return fmt.Errorf("unknown group %T for bls msm circuit", vv)
	}
	return nil
}
