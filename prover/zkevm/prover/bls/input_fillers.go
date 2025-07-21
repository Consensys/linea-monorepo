package bls

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/plonk"
)

func init() {
	plonk.RegisterInputFiller(membershipInputFillerKey(G1, CURVE), newMembershipInputFiller(G1, CURVE))
	plonk.RegisterInputFiller(membershipInputFillerKey(G2, CURVE), newMembershipInputFiller(G2, CURVE))
}

func membershipInputFillerKey(g group, m membership) string {
	base := "bls12381-%s-membership-input-filler"
	switch m {
	case CURVE:
		return fmt.Sprintf(base, g.StringCurve())
	case GROUP:
		return fmt.Sprintf(base, g.String())
	default:
		panic(fmt.Sprintf("unknown membership type %v for group %v", m, g))
	}
}

func newMembershipInputFiller(g group, m membership) plonk.InputFiller {
	switch m {
	case CURVE:
		return func(circuitInstance, inputIndex int) field.Element {
			nbL := nbLimbs(g)
			if inputIndex%(nbL+1) == 0 {
				return field.One() // first input is the success bit
			} else {
				return field.Zero() // other inputs are zero
			}
		}
	}
	return nil
}
