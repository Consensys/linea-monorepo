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
	NAME_BLS_ADD = "BLS_ADD"
)

const (
	nbRowsPerG1Add = 3 * nbG1Limbs // 2 for the inputs, 1 for the output
	nbRowsPerG2Add = 3 * nbG2Limbs // 2 for the inputs, 1 for the output
)

type BlsAddDataSource struct {
	CsAdd  ifaces.Column
	Limb   ifaces.Column
	Index  ifaces.Column
	IsData ifaces.Column
	IsRes  ifaces.Column
}

func newAddDataSource(comp *wizard.CompiledIOP, g group) *BlsAddDataSource {
	var selectorCs ifaces.Column
	switch g {
	case G1:
		selectorCs = comp.Columns.GetHandle("bls.CIRCUIT_SELECTOR_BLS_G1_ADD")
	case G2:
		selectorCs = comp.Columns.GetHandle("bls.CIRCUIT_SELECTOR_BLS_G2_ADD")
	default:
		panic("unknown group for bls add data source")
	}
	return &BlsAddDataSource{
		CsAdd:  selectorCs,
		Limb:   comp.Columns.GetHandle("bls.LIMB"),
		Index:  comp.Columns.GetHandle("bls.INDEX"),
		IsData: comp.Columns.GetHandle("bls.IS_BLS_ADD_DATA"),
		IsRes:  comp.Columns.GetHandle("bls.IS_BLS_ADD_RESULT"),
	}
}

type BlsAdd struct {
	*BlsAddDataSource
	AlignedGnarkData *plonk.Alignment

	size int
	*Limits
	group
}

func newAdd(comp *wizard.CompiledIOP, g group, limits *Limits, src *BlsAddDataSource, plonkOptions []plonk.Option) *BlsAdd {
	size := limits.sizeAddIntegration(g)

	toAlign := &plonk.CircuitAlignmentInput{
		Name:               fmt.Sprintf("%s_%s_ALIGNMENT", NAME_BLS_ADD, g.String()),
		Round:              ROUND_NR,
		DataToCircuitMask:  src.CsAdd,
		DataToCircuit:      src.Limb,
		Circuit:            NewAddCircuit(g, limits),
		NbCircuitInstances: limits.nbAddCircuitInstances(g),
		PlonkOptions:       plonkOptions,
	}

	return &BlsAdd{
		BlsAddDataSource: src,
		AlignedGnarkData: plonk.DefineAlignment(comp, toAlign),
		size:             size,
		Limits:           limits,
		group:            g,
	}
}

func NewG1AddZkEvm(comp *wizard.CompiledIOP, limits *Limits) *BlsAdd {
	return newAdd(
		comp,
		G1,
		limits,
		newAddDataSource(comp, G1),
		[]plonk.Option{plonk.WithRangecheck(16, 6, true)},
	)
}

func NewG2AddZkEvm(comp *wizard.CompiledIOP, limits *Limits) *BlsAdd {
	return newAdd(
		comp,
		G2,
		limits,
		newAddDataSource(comp, G2),
		[]plonk.Option{plonk.WithRangecheck(16, 6, true)},
	)
}

type MultiAddCircuit[C convertable[T], T element] struct {
	Instances []AddInstance[C, T]
}

func NewAddCircuit(g group, limits *Limits) frontend.Circuit {
	switch g {
	case G1:
		return &MultiAddCircuit[g1ElementWizard, sw_bls12381.G1Affine]{Instances: make([]AddInstance[g1ElementWizard, sw_bls12381.G1Affine], limits.NbG1AddInputInstances)}
	case G2:
		return &MultiAddCircuit[g2ElementWizard, sw_bls12381.G2Affine]{Instances: make([]AddInstance[g2ElementWizard, sw_bls12381.G2Affine], limits.NbG2AddInputInstances)}
	default:
		panic(fmt.Sprintf("unknown group for bls add circuit: %v", g))
	}
}

type AddInstance[C convertable[T], T element] struct {
	InputLeft, InputRight C `gnark:",public"`
	Res                   C `gnark:",public"`
}

func (c *MultiAddCircuit[C, T]) Define(api frontend.API) error {
	f, err := emulated.NewField[sw_bls12381.BaseField](api)
	if err != nil {
		return fmt.Errorf("new field: %w", err)
	}
	nbInstances := len(c.Instances)
	As, Bs, Rs := make([]T, nbInstances), make([]T, nbInstances), make([]T, nbInstances)
	// TODO: move this part into the main loop below. Then we don't need to type assert
	for i := range c.Instances {
		As[i] = c.Instances[i].InputLeft.ToElement(api, f)
		Bs[i] = c.Instances[i].InputRight.ToElement(api, f)
		Rs[i] = c.Instances[i].Res.ToElement(api, f)
	}
	var t T
	switch any(t).(type) {
	case sw_bls12381.G1Affine:
		// TODO: update evmprecompiles to take the result as input and then use that
		curve, err := sw_emulated.New[emulated.BLS12381Fp, emulated.BLS12381Fr](api, sw_emulated.GetBLS12381Params())
		if err != nil {
			return fmt.Errorf("get curve: %w", err)
		}
		for i := 0; i < nbInstances; i++ {
			tAsi := any(&As[i]).(*sw_bls12381.G1Affine)
			tBsi := any(&Bs[i]).(*sw_bls12381.G1Affine)
			tRsi := any(&Rs[i]).(*sw_bls12381.G1Affine)
			curve.AssertIsOnCurve(tAsi)
			curve.AssertIsOnCurve(tBsi)
			res := curve.AddUnified(any(&As[i]).(*sw_bls12381.G1Affine), any(&Bs[i]).(*sw_bls12381.G1Affine))
			curve.AssertIsEqual(res, tRsi)
		}
	case sw_bls12381.G2Affine:
		// TODO: update evmprecompiles to take the result as input and then use that
		g2 := sw_bls12381.NewG2(api)
		for i := 0; i < nbInstances; i++ {
			tAsi := any(&As[i]).(*sw_bls12381.G2Affine)
			tBsi := any(&Bs[i]).(*sw_bls12381.G2Affine)
			tRsi := any(&Rs[i]).(*sw_bls12381.G2Affine)
			g2.AssertIsOnTwist(tAsi)
			g2.AssertIsOnTwist(tBsi)
			res := g2.AddUnified(tAsi, tBsi)
			g2.AssertIsEqual(res, tRsi)
		}
	default:
		return fmt.Errorf("unknown element type %T for bls add circuit", t)
	}

	return nil
}
