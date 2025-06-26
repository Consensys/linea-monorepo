//go:build !fuzzlight

package test_cases_test

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/stitchsplit"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/require"
)

func defineLocalOpening(builder *wizard.Builder) {
	P1 := builder.RegisterCommit(P1, 8)
	P1Next := column.Shift(P1, 1)
	P2Prev := column.Shift(P1, -1)
	builder.LocalOpening("Inclusion1", P1Next)
	builder.LocalOpening("Inclusion2", P2Prev)
}

func proverLocalOpening(run *wizard.ProverRuntime) {
	run.AssignColumn(P1, smartvectors.ForTest(0, 1, 2, 3, 4, 5, 6, 7))
	run.AssignLocalPoint("Inclusion1", fext.NewElement(1, 0, 0, 0))
	run.AssignLocalPoint("Inclusion2", fext.NewElement(7, 0, 0, 0))
}

func TestGnarkCompile(t *testing.T) {
	comp := wizard.Compile(defineLocalOpening, stitchsplit.Splitter(32))
	proof := wizard.Prove(comp, proverLocalOpening)

	circ := wizard.AllocateWizardCircuit(comp, comp.NumRounds())

	scs, err := frontend.Compile(field.Modulus(), scs.NewBuilder, &SimpleTestGnarkCircuit{C: *circ})
	if err != nil {
		panic(err)
	}

	assignment := GetAssignment(comp, proof)
	witness, err := frontend.NewWitness(assignment, ecc.BLS12_377.ScalarField())
	require.NoError(t, err)

	err = scs.IsSolved(witness)
	if err != nil {
		// When the error string is too large `require.NoError` does not print
		// the error.
		t.Fatalf("circuit solving failed : %v\n", err)
	}
}
