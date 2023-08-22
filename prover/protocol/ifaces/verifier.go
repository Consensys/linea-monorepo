package ifaces

import (
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/gnark/frontend"
)

// Interface implemented by the ProverRuntime and the VerifierRuntime
type Runtime interface {
	GetColumn(ColID) ColAssignment
	GetColumnAt(ColID, int) field.Element
	GetRandomCoinField(name coin.Name) field.Element
	GetRandomCoinIntegerVec(name coin.Name) []int
	GetParams(id QueryID) QueryParams
}

// Interface implemented by the wizard.GnarkVerifierCircuit
type GnarkRuntime interface {
	GetColumn(ColID) []frontend.Variable
	GetColumnAt(ColID, int) frontend.Variable
	GetRandomCoinField(name coin.Name) frontend.Variable
	GetRandomCoinIntegerVec(name coin.Name) []frontend.Variable
	GetParams(id QueryID) GnarkQueryParams
}
