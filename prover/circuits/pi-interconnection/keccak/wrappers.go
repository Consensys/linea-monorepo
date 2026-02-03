package keccak

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/circuits/pi-interconnection/keccak"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/wizard"
)

type (
	CompilationParams    []func(iop *wizard.CompiledIOP)
	BlockHasher          keccak.BlockHasher
	StrictHasherCircuit  struct{ *keccak.StrictHasherCircuit }
	StrictHasherCompiler struct{ *keccak.StrictHasherCompiler }
	StrictHasherSnark    struct{ *keccak.StrictHasherSnark }
	CompiledStrictHasher struct{ *keccak.CompiledStrictHasher }
)

func WizardCompilationParameters() CompilationParams {
	return keccak.WizardCompilationParameters()
}

func NewStrictHasherCompiler(lengthsOfLengths ...int) StrictHasherCompiler {
	c := keccak.NewStrictHasherCompiler(lengthsOfLengths...)
	return StrictHasherCompiler{&c}
}

func (c StrictHasherCircuit) NewHasher(api frontend.API) StrictHasherSnark {
	h := c.StrictHasherCircuit.NewHasher(api)
	return StrictHasherSnark{&h}
}

func (c StrictHasherCompiler) Compile(params CompilationParams) CompiledStrictHasher {
	h := c.StrictHasherCompiler.Compile(([]func(iop *wizard.CompiledIOP))(params)...)
	return CompiledStrictHasher{&h}
}

func (h CompiledStrictHasher) GetCircuit() (StrictHasherCircuit, error) {
	c, err := h.CompiledStrictHasher.GetCircuit()
	return StrictHasherCircuit{&c}, err
}

func (h CompiledStrictHasher) GetHasher() {
	hsh := h.CompiledStrictHasher.GetHasher()
	return
}

func RegisterHints() {
	keccak.RegisterHints()
}
