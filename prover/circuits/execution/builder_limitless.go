package execution

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/profile"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

type limitlessBuilder struct {
	congWIOP    *wizard.CompiledIOP
	traceLimits *config.TracesLimits
}

func NewLimitlessBuilder(congWIOP *wizard.CompiledIOP, traceLimits *config.TracesLimits) *limitlessBuilder {
	return &limitlessBuilder{congWIOP: congWIOP, traceLimits: traceLimits}
}

func (b *limitlessBuilder) Compile() (constraint.ConstraintSystem, error) {
	return makeLimitlessCS(b), nil
}

// Makes the constraint system for the execution-limitless circuit
func makeLimitlessCS(b *limitlessBuilder) constraint.ConstraintSystem {
	circuit := AllocateLimitless(b.congWIOP, b.traceLimits)

	// To be removed as it slows down the compilation
	gnarkPprof := profile.Start(profile.WithPath("execution_limitless_gnark.pprof"))
	defer gnarkPprof.Stop()

	scs, err := frontend.Compile(fr.Modulus(), scs.NewBuilder, &circuit, frontend.WithCapacity(1<<24))
	if err != nil {
		panic(err)
	}
	return scs
}
