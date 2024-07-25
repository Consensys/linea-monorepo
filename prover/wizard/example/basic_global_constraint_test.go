package example

import (
	"strings"
	"testing"

	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	sym "github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/wizard"
	"github.com/stretchr/testify/assert"
)

func TestExampleBasicGlobalConstraint(t *testing.T) {

	var (
		a, b *wizard.ColNatural
	)

	define := func(api *wizard.API) {

		subAPI := api.ChildScope("letters").AddTags("alphabet", "letters", "example")

		a = subAPI.NewCommit(0, 32).
			WithDoc("A stores the values contained in `A`. A is the first letter of the alphabet").
			WithName("A").
			WithTags("first letter", "encoded:string")

		b = subAPI.NewCommit(0, 32).
			WithDoc("B stores the values contained in `B`. B is the second letter of the alphabet").
			WithName("B").
			WithTags("alphabet", "letter")

		subAPI.NewQueryGlobal(sym.Mul(a, b)).
			WithDoc("Either A or B zero must be zero at every row").
			WithName("a-or-b-cancel").
			WithTags("alphabet", "second letter")

		subAPI.NewLocalOpening(a, 1).
			WithDoc("Returns the first position of A").
			WithName("a-first-position").
			WithTags("first-position")
	}

	prover := func(run *wizard.RuntimeProver) {
		run.Logger.Info("aaa")
		subRun := run.ChildScope("col-assignment")
		a.AssignConstant(subRun, field.Zero())
		b.AssignConstant(subRun, field.Zero())
	}

	var (
		comp        = wizard.NewAPI(define).Compile().CompiledIOP
		stats       = comp.Stats()
		csvStringer = &strings.Builder{}
		csvErr      = stats.SaveAsCsv(csvStringer)
		proverRT    = comp.NewRuntimeProver(prover).Run()
		proof       = proverRT.Proof()
		sanityErr   = proverRT.SanityCheck()
		verifierErr = comp.Verify(proof)
	)

	assert.NoError(t, sanityErr)
	assert.NoError(t, verifierErr)
	assert.NoError(t, csvErr)
}
