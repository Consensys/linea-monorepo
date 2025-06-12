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
	return &BlsAddDataSource{
		CsAdd:  comp.Columns.GetHandle(ifaces.ColIDf("bls.CIRCUIT_SELECTOR_%s_ADD", g.String())),
		Limb:   comp.Columns.GetHandle("bls.LIMB"),
		Index:  comp.Columns.GetHandle("bls.INDEX"),
		IsData: comp.Columns.GetHandle(ifaces.ColIDf("bls.DATA_%s_ADD", g.String())),
		IsRes:  comp.Columns.GetHandle(ifaces.ColIDf("bls.RSLT_%s_ADD", g.String())),
	}
}

type BlsAdd struct {
	*BlsAddDataSource
	AlignedGnarkData *plonk.Alignment

	size int
	*Limits
	group
}

func newAdd(comp *wizard.CompiledIOP, g group, limits *Limits, src *BlsAddDataSource, plonkOptions []query.PlonkOption) *BlsAdd {
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

func (ba *BlsAdd) Assign(run *wizard.ProverRuntime) {
	ba.AlignedGnarkData.Assign(run)
}

func NewG1AddZkEvm(comp *wizard.CompiledIOP, limits *Limits) *BlsAdd {
	return newAdd(
		comp,
		G1,
		limits,
		newAddDataSource(comp, G1),
		[]query.PlonkOption{query.PlonkRangeCheckOption(16, 6, true)},
	)
}

func NewG2AddZkEvm(comp *wizard.CompiledIOP, limits *Limits) *BlsAdd {
	return newAdd(
		comp,
		G2,
		limits,
		newAddDataSource(comp, G2),
		[]query.PlonkOption{query.PlonkRangeCheckOption(16, 6, true)},
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
	switch vv := any(c.Instances).(type) {
	case []AddInstance[g1ElementWizard, sw_bls12381.G1Affine]:
		for i := 0; i < nbInstances; i++ {
			left := vv[i].InputLeft.ToElement(api, f)
			right := vv[i].InputRight.ToElement(api, f)
			expected := vv[i].Res.ToElement(api, f)
			evmprecompiles.ECAddG1BLS(api, &left, &right, &expected)
		}
	case []AddInstance[g2ElementWizard, sw_bls12381.G2Affine]:
		for i := 0; i < nbInstances; i++ {
			left := vv[i].InputLeft.ToElement(api, f)
			right := vv[i].InputRight.ToElement(api, f)
			expected := vv[i].Res.ToElement(api, f)
			evmprecompiles.ECAddG2BLS(api, &left, &right, &expected)
		}
	default:
		return fmt.Errorf("unknown element type %T for bls add circuit", vv)
	}

	return nil
}
