package finalwrap

import (
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// Builder implements the circuit builder interface used by the setup command.
// It compiles the FinalWrapCircuit into a BN254 constraint system.
type Builder struct {
	rootComp *wizard.CompiledIOP
}

// NewBuilder creates a builder for the final wrap circuit.
// rootComp is the CompiledIOP of the tree aggregation root level.
func NewBuilder(rootComp *wizard.CompiledIOP) *Builder {
	return &Builder{rootComp: rootComp}
}

// Compile compiles the FinalWrapCircuit into a BN254 constraint system.
func (b *Builder) Compile() (constraint.ConstraintSystem, error) {
	return MakeCS(b.rootComp)
}
