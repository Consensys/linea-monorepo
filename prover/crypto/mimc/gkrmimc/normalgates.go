package gkrmimc

import (
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/gnark/frontend"
)

// RoundGate represents a normal round of gkr (i.e. any
// round except for the first one)
type RoundGate struct {
	Ark frontend.Variable
}

// NewRoundGateGnark creates a new MiMCCipherGate with the given parameters
func NewRoundGateGnark(ark field.Element) *RoundGate {
	return &RoundGate{
		Ark: ark,
	}
}

func (m RoundGate) Evaluate(api frontend.API, input ...frontend.Variable) frontend.Variable {

	if len(input) != 2 {
		panic("mimc has fan-in 2")
	}

	initialState := input[0]
	curr := input[1]

	// Compute the s-box (curr + init + ark)^5
	sum := api.Add(curr, initialState, m.Ark)

	sumPow4 := api.Mul(sum, sum)        // sum^2
	sumPow4 = api.Mul(sumPow4, sumPow4) // sum^4
	return api.Mul(sumPow4, sum)
}

func (m RoundGate) Degree() int {
	return 5
}

// CryptoRoundgate implements the gate for the GKR prover's side
type RoundGateCrypto struct {
	Ark field.Element
}

func NewRoundGateCrypto(ark field.Element) *RoundGateCrypto {
	return &RoundGateCrypto{Ark: ark}
}

func (m RoundGateCrypto) Evaluate(inputs ...field.Element) field.Element {

	if len(inputs) != 2 {
		panic("mimc has fan-in 2")
	}

	initialState := inputs[0]
	curr := inputs[1]

	var sum, sumPow4 field.Element
	sum.Add(&initialState, &curr).Add(&sum, &m.Ark)
	sumPow4.Mul(&sum, &sum)
	sumPow4.Mul(&sumPow4, &sumPow4)
	sum.Mul(&sumPow4, &sum)

	return sum
}

func (m RoundGateCrypto) Degree() int {
	return 5
}
