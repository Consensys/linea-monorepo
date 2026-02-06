package distributed_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

func TestManualShifter(t *testing.T) {
	testCases := []struct {
		Define     func(*wizard.Builder)
		Prove      func(*wizard.ProverRuntime)
		Title      string
		ShouldPass bool
	}{
		{
			Define: func(builder *wizard.Builder) {
				a := builder.RegisterCommit("A", 16)
				b := builder.RegisterCommit("B", 16)
				b_1 := column.Shift(b, 1)
				builder.Permutation("PERM", []ifaces.Column{a}, []ifaces.Column{b_1})
			},

			Prove: func(run *wizard.ProverRuntime) {
				run.AssignColumn("A", smartvectors.ForTest(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15))
				run.AssignColumn("B", smartvectors.ForTest(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 0))
			},

			Title:      "permutation",
			ShouldPass: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Title, func(t *testing.T) {
			comp := wizard.Compile(testCase.Define, distributed.CompileManualShifter)
			err := distributed.AuditInitialWizard(comp)
			if err != nil {
				t.Fatalf("audit failed: %v", err.Error())
			}
			proof := wizard.Prove(comp, testCase.Prove)
			if err := wizard.Verify(comp, proof); err != nil && testCase.ShouldPass {
				t.Fatalf("verifier did not pass: %v", err.Error())
			}
			if err := wizard.Verify(comp, proof); err == nil && !testCase.ShouldPass {
				t.Fatalf("verifier is passing for a false claim")
			}
		})
	}
}
