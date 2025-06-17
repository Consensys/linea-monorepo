package bls

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/emulated/sw_bls12381"
	"github.com/consensys/gnark/std/evmprecompiles"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/plonk"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

type OnCurveInstance[C convertable[T], T element] struct {
	// IsOnCurve is 1 if the check is successful, 0 otherwise
	IsOnCurve frontend.Variable `gnark:",public"`
	// P is the purported element
	P C `gnark:",public"`
}

func (c OnCurveInstance[C, T]) Check(api frontend.API, fp *emulated.Field[sw_bls12381.BaseField], pairing *sw_bls12381.Pairing) error {
	switch v := any(c.P).(type) {
	case g1ElementWizard:
		P := v.ToElement(api, fp)
		res := pairing.IsOnCurve(&P)
		api.AssertIsEqual(res, c.IsOnCurve)
		return nil
	case g2ElementWizard:
		Q := v.ToElement(api, fp)
		res := pairing.IsOnTwist(&Q)
		api.AssertIsEqual(res, c.IsOnCurve)
		return nil
	default:
		return fmt.Errorf("unsupported element type %T for on-curve check", c.P)
	}
}

type NonGroupMembershipInstance[C convertable[T], T element] struct {
	// P is the purported element
	P C `gnark:",public"`
}

func (c NonGroupMembershipInstance[C, T]) Check(api frontend.API, fp *emulated.Field[sw_bls12381.BaseField], pairing *sw_bls12381.Pairing) error {
	switch v := any(c.P).(type) {
	case g1ElementWizard:
		P := v.ToElement(api, fp)
		// We expect the element to be not on G1
		return evmprecompiles.ECPairBLSIsOnG1(api, &P, 0) // 0 means we expect it to be not on G1
	case g2ElementWizard:
		Q := v.ToElement(api, fp)
		// We expect the element to be not on G2
		return evmprecompiles.ECPairBLSIsOnG2(api, &Q, 0) // 0 means we expect it to be not on G2
	default:
		return fmt.Errorf("unsupported element type %T for non-group membership check", c.P)
	}
}

// -- circuit which performs multiple checks

type checkableInstance interface {
	OnCurveInstance[g1ElementWizard, sw_bls12381.G1Affine] | OnCurveInstance[g2ElementWizard, sw_bls12381.G2Affine] |
		NonGroupMembershipInstance[g1ElementWizard, sw_bls12381.G1Affine] | NonGroupMembershipInstance[g2ElementWizard, sw_bls12381.G2Affine]
	Check(api frontend.API, fp *emulated.Field[sw_bls12381.BaseField], pairing *sw_bls12381.Pairing) error
}

type multiCheckableCircuit[T checkableInstance] struct {
	Instances []T
}

func newMultiCheckableCircuit[T checkableInstance](nbInstances int) *multiCheckableCircuit[T] {
	return &multiCheckableCircuit[T]{
		Instances: make([]T, nbInstances),
	}
}

func (c *multiCheckableCircuit[T]) Define(api frontend.API) error {
	fp, err := emulated.NewField[sw_bls12381.BaseField](api)
	if err != nil {
		return fmt.Errorf("new field emulation: %w", err)
	}
	pairing, err := sw_bls12381.NewPairing(api)
	if err != nil {
		return fmt.Errorf("new pairing: %w", err)
	}
	for i := range c.Instances {
		if err := c.Instances[i].Check(api, fp, pairing); err != nil {
			return fmt.Errorf("instance %d check: %w", i, err)
		}
	}
	return nil
}

func NewCheckCircuit(g group, membership membership, limits *Limits) frontend.Circuit {
	switch g {
	case G1:
		switch membership {
		case CURVE:
			return newMultiCheckableCircuit[OnCurveInstance[g1ElementWizard, sw_bls12381.G1Affine]](limits.NbC1MembershipInputInstances)
		case GROUP:
			return newMultiCheckableCircuit[NonGroupMembershipInstance[g1ElementWizard, sw_bls12381.G1Affine]](limits.NbG1MembershipInputInstances)
		default:
			panic(fmt.Sprintf("unknown membership type for G1: %v", membership))
		}
	case G2:
		switch membership {
		case CURVE:
			return newMultiCheckableCircuit[OnCurveInstance[g2ElementWizard, sw_bls12381.G2Affine]](limits.NbC2MembershipInputInstances)
		case GROUP:
			return newMultiCheckableCircuit[NonGroupMembershipInstance[g2ElementWizard, sw_bls12381.G2Affine]](limits.NbG2MembershipInputInstances)
		default:
			panic(fmt.Sprintf("unknown membership type for G2: %v", membership))
		}
	default:
		panic(fmt.Sprintf("unknown group for bls curve membership circuit: %v", g))
	}
}

type UnalignedCurveMembershipData struct {
	*UnalignedCurveMembershipDataSource

	// IsActive is a constructed column which indicates if the circuit is active. Set when selector is on or when we provide the input data.
	IsActive ifaces.Column
	// IsFirstLineOfInput is a constructed column which indicates if the row is
	// the first line of the input.
	IsFirstLineOfInput    ifaces.Column
	IsFirstLineOfInputAct wizard.ProverAction

	GnarkData ifaces.Column
}

type UnalignedCurveMembershipDataSource struct {
	Limb              ifaces.Column
	Counter           ifaces.Column
	CsCurveMembership ifaces.Column
}

func newUnalignedCurveMembershipData(comp *wizard.CompiledIOP, g group, size int, src *UnalignedCurveMembershipDataSource) *UnalignedCurveMembershipData {
	createCol := createColFn(comp, fmt.Sprintf("UNALIGNED_%s_CURVE_MEMBERSHIP", g.StringCurve()), size)
	res := &UnalignedCurveMembershipData{
		UnalignedCurveMembershipDataSource: src,
		IsActive:                           createCol("IS_ACTIVE"),
		GnarkData:                          createCol("GNARK_DATA"),
	}
	res.IsFirstLineOfInput, res.IsFirstLineOfInputAct = dedicated.IsZero(comp, src.Counter)
	return res
}

func (d *UnalignedCurveMembershipData) Assign(run *wizard.ProverRuntime) {
	d.IsFirstLineOfInputAct.Run(run)

	var (
		srcLimb    = d.Limb.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcCounter = d.Counter.GetColAssignment(run).IntoRegVecSaveAlloc()
		srcCs      = d.CsCurveMembership.GetColAssignment(run).IntoRegVecSaveAlloc()
	)
	var (
		dstIsActive  = common.NewVectorBuilder(d.IsActive)
		dstGnarkData = common.NewVectorBuilder(d.GnarkData)
	)

	for i := 0; i < len(srcLimb); i++ {
		if srcCs[i].IsZero() {
			continue
		}
		dstIsActive.PushBoolean(true)
		// for the first line of input, we push the expected success bit
		if srcCounter[i].IsZero() {
			dstGnarkData.PushBoolean(false) // we push additional input to gnark input to indicate curve non-membership
		}
		dstGnarkData.PushField(srcLimb[i])
	}
	dstIsActive.PadAndAssign(run, field.Zero())
	dstGnarkData.PadAndAssign(run, field.Zero())
}

const (
	inputFillerC1MembershipKey = "bls12381-c1-membership-input-filler"
	inputFillerC2MembershipKey = "bls12381-c2-membership-input-filler"
)

func init() {
	plonk.RegisterInputFiller(inputFillerC1MembershipKey, newMembershipInputFiller(G1, CURVE))
	plonk.RegisterInputFiller(inputFillerC2MembershipKey, newMembershipInputFiller(G2, CURVE))
}

func newMembershipInputFiller(g group, m membership) plonk.InputFiller {
	switch m {
	case CURVE:
		return func(circuitInstance, inputIndex int) field.Element {
			var nbLimbs int
			switch g {
			case G1:
				nbLimbs = nbG1Limbs
			case G2:
				nbLimbs = nbG2Limbs
			}
			if inputIndex%(nbLimbs+1) == 0 {
				return field.One() // first input is the success bit
			} else {
				return field.Zero() // other inputs are zero
			}
		}
	}
	return nil
}
