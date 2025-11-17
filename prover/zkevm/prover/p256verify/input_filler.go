package p256verify

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/plonk"
)

const (
	input_filler_key = "p256-verify-input-filler"
)

func init() {
	plonk.RegisterInputFiller(input_filler_key, inputFiller)
}

func inputFiller(circuitInstance, inputIndex int) field.Element {
	// TODO
	return field.NewElement(0)
}
