package compilers_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/wiop"
	"github.com/consensys/linea-monorepo/prover/wiop/compilers"
	"github.com/consensys/linea-monorepo/prover/wiop/wioptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompile_Completeness verifies that for every vanishing scenario, an
// honest witness satisfies all verifier actions that Compile registers.
func TestCompile_Completeness(t *testing.T) {
	for _, build := range wioptest.VanishingScenarios() {
		sc := build()
		t.Run(sc.Name, func(t *testing.T) {
			compilers.Compile(sc.Sys)
			rt := wiop.NewRuntime(sc.Sys)
			sc.AssignHonest(&rt)
			require.NoError(t, wioptest.RunAndVerify(&rt),
				"compiled verifier must accept an honest witness")
		})
	}
}

// TestCompile_Soundness verifies that for every vanishing scenario, an invalid
// witness (one that violates at least one constraint) is rejected by the
// compiled verifier. This is the core soundness property of the compiler: a
// cheating prover cannot produce a quotient that satisfies the identity check.
func TestCompile_Soundness(t *testing.T) {
	for _, build := range wioptest.VanishingScenarios() {
		sc := build()
		t.Run(sc.Name, func(t *testing.T) {
			compilers.Compile(sc.Sys)
			rt := wiop.NewRuntime(sc.Sys)
			sc.AssignInvalid(&rt)
			assert.Error(t, wioptest.RunAndVerify(&rt),
				"compiled verifier must reject an invalid witness")
		})
	}
}
