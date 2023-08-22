package gkrmimc

import (
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/gnark/frontend"
)

// FinalRoundGate represents the last round in the circuit
// It performs all the actions required to complete the
// compression function of MiMC.
type FinalRoundGate struct {
	Ark frontend.Variable
}

// NewFinalRoundGateGnark creates a new MiMCCipherGate with the given parameters
func NewFinalRoundGateGnark(ark field.Element) FinalRoundGate {
	return FinalRoundGate{
		Ark: ark,
	}
}

func (m FinalRoundGate) Evaluate(api frontend.API, input ...frontend.Variable) frontend.Variable {

	if len(input) != 3 {
		utils.Panic("expected fan-in of 3, got %v", len(input))
	}

	// Parse the inputs
	initialState := input[0]
	block := input[1]
	currentState := input[2]

	// Compute the S-box function
	sum := api.Add(currentState, initialState, m.Ark)
	sumPow4 := api.Mul(sum, sum)        // sum^2
	sumPow4 = api.Mul(sumPow4, sumPow4) // sum^4
	sum = api.Mul(sumPow4, sum)

	// And add back the last values, following the Miyaguchi-Preneel
	// construction.
	return api.Add(sum, initialState, initialState, block)
}

func (m FinalRoundGate) Degree() int {
	return 5
}

// FinalRoundGateCrypto represents the last round in the circuit
// It performs all the actions required to complete the
// compression function of MiMC.
type FinalRoundGateCrypto struct {
	Ark field.Element
}

// NewFinalRoundGateCrypto creates a new MiMCCipherGate with the given parameters
func NewFinalRoundGateCrypto(ark field.Element) FinalRoundGateCrypto {
	return FinalRoundGateCrypto{
		Ark: ark,
	}
}

func (m FinalRoundGateCrypto) Evaluate(input ...field.Element) field.Element {

	if len(input) != 3 {
		utils.Panic("expected fan-in of 3, got %v", len(input))
	}

	// Parse the inputs
	initialState := input[0]
	block := input[1]
	curr := input[2]

	// Compute the S-box function
	var sum, sumPow4 field.Element
	sum.Add(&initialState, &curr).Add(&sum, &m.Ark)
	sumPow4.Mul(&sum, &sum)
	sumPow4.Mul(&sumPow4, &sumPow4)
	sum.Mul(&sumPow4, &sum)

	// And add back the last values, following the Miyaguchi-Preneel
	// construction.
	sum.Add(&sum, &initialState)
	sum.Add(&sum, &initialState)
	sum.Add(&sum, &block)
	return sum
}

func (m FinalRoundGateCrypto) Degree() int {
	return 5
}
