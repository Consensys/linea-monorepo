package ecarith

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra"
	"github.com/consensys/gnark/std/algebra/emulated/sw_bn254"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/plonk"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/arithmetization"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

const (
	NAME_ECADD = "ECADD_INTEGRATION"
)

const (
	nbRowsPerEcAdd = 12
)

// EcAdd integrated EC_ADD precompile call verification inside a
// gnark circuit.
type EcAdd struct {
	*EcDataAddSource
	AlignedGnarkData *plonk.Alignment

	FlattenLimbs *common.FlattenColumn

	Size int
	*Limits
}

func NewEcAddZkEvm(comp *wizard.CompiledIOP, limits *Limits, arith *arithmetization.Arithmetization) *EcAdd {

	src := &EcDataAddSource{
		CsEcAdd: arith.ColumnOf(comp, "ecdata", "CIRCUIT_SELECTOR_ECADD"),
		Index:   arith.ColumnOf(comp, "ecdata", "INDEX"),
		IsData:  arith.ColumnOf(comp, "ecdata", "IS_ECADD_DATA"),
		IsRes:   arith.ColumnOf(comp, "ecdata", "IS_ECADD_RESULT"),
		Limbs:   arith.LimbColumnsOfArr8(comp, "ecdata", "LIMB"),
	}

	return newEcAdd(
		comp,
		limits,
		src,
		[]query.PlonkOption{query.PlonkRangeCheckOption(16, 6, true)},
	)
}

// newEcAdd creates a new EC_ADD integration.
func newEcAdd(comp *wizard.CompiledIOP, limits *Limits, src *EcDataAddSource, plonkOptions []query.PlonkOption) *EcAdd {
	size := limits.sizeEcAddIntegration()

	flattenLimbs := common.NewFlattenColumn(comp, common.NbLimbU128, src.Limbs[:], src.CsEcAdd)

	toAlign := &plonk.CircuitAlignmentInput{
		Name:               NAME_ECADD + "_ALIGNMENT",
		Round:              ROUND_NR,
		DataToCircuitMask:  flattenLimbs.Mask(),
		DataToCircuit:      flattenLimbs.Limbs(),
		Circuit:            NewECAddCircuit(limits),
		NbCircuitInstances: limits.NbCircuitInstances,
		PlonkOptions:       plonkOptions,
		// This resolves to the statement (0, 0) + (0, 0) = (0, 0), which is
		// correct as (0, 0) encodes the point at infinity.
		InputFillerKey: "",
	}

	res := &EcAdd{
		EcDataAddSource:  src,
		AlignedGnarkData: plonk.DefineAlignment(comp, toAlign),
		FlattenLimbs:     flattenLimbs,
		Size:             size,
	}

	flattenLimbs.CsFlattenProjection(comp)

	return res
}

// Assign assigns the data from the trace to the gnark inputs.
func (em *EcAdd) Assign(run *wizard.ProverRuntime) {
	em.FlattenLimbs.Run(run)
	em.AlignedGnarkData.Assign(run)
}

// EcDataAddSource is a struct that holds the columns that are used to
// fetch data from the EC_DATA module from the arithmetization.
type EcDataAddSource struct {
	CsEcAdd ifaces.Column
	Limbs   [common.NbLimbU128]ifaces.Column
	Index   ifaces.Column
	IsData  ifaces.Column
	IsRes   ifaces.Column
}

// MultiECAddCircuit is a circuit that can handle multiple EC_ADD instances. The
// length of the slice Instances should corresponds to the one defined in the
// Limits struct.
type MultiECAddCircuit struct {
	Instances []ECAddInstance
}

type ECAddInstance struct {
	// First input to addition
	P_X [common.NbLimbU256]frontend.Variable `gnark:",public"`
	P_Y [common.NbLimbU256]frontend.Variable `gnark:",public"`

	// Second input to addition
	Q_X [common.NbLimbU256]frontend.Variable `gnark:",public"`
	Q_Y [common.NbLimbU256]frontend.Variable `gnark:",public"`

	// The result of the addition. Is provided non-deterministically by the
	// caller, we have to ensure that the result is correct.
	R_X [common.NbLimbU256]frontend.Variable `gnark:",public"`
	R_Y [common.NbLimbU256]frontend.Variable `gnark:",public"`
}

// NewECAddCircuit creates a new circuit for verifying the EC_MUL precompile
// based on the defined number of inputs.
func NewECAddCircuit(limits *Limits) *MultiECAddCircuit {
	return &MultiECAddCircuit{
		Instances: make([]ECAddInstance, limits.NbInputInstances),
	}
}

func (c *MultiECAddCircuit) Define(api frontend.API) error {

	f, err := emulated.NewField[sw_bn254.BaseField](api)
	if err != nil {
		return fmt.Errorf("field emulation: %w", err)
	}

	// gnark circuit works with 64 bits values, we need to split the 128 bits
	// values into high and low parts.
	nbInstances := len(c.Instances)
	Ps := make([]sw_bn254.G1Affine, nbInstances)
	Qs := make([]sw_bn254.G1Affine, nbInstances)
	Rs := make([]sw_bn254.G1Affine, nbInstances)
	for i := range c.Instances {

		var (
			PX = f.NewElement(c.Instances[i].P_X[:])
			PY = f.NewElement(c.Instances[i].P_Y[:])
			QX = f.NewElement(c.Instances[i].Q_X[:])
			QY = f.NewElement(c.Instances[i].Q_Y[:])
			RX = f.NewElement(c.Instances[i].R_X[:])
			RY = f.NewElement(c.Instances[i].R_Y[:])
			P  = sw_bn254.G1Affine{X: *PX, Y: *PY}
			Q  = sw_bn254.G1Affine{X: *QX, Y: *QY}
			R  = sw_bn254.G1Affine{X: *RX, Y: *RY}
		)

		Ps[i] = P
		Qs[i] = Q
		Rs[i] = R
	}

	curve, err := algebra.GetCurve[sw_bn254.ScalarField, sw_bn254.G1Affine](api)
	if err != nil {
		panic(err)
	}
	for i := range Rs {
		res := curve.AddUnified(&Ps[i], &Qs[i])
		curve.AssertIsEqual(&Rs[i], res)
	}
	return nil
}
