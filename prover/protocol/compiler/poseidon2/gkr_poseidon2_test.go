package poseidon2

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/compiler/logdata"
	"github.com/consensys/linea-monorepo/prover/protocol/internal/testtools"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/require"
)

func TestGKRPoseidon2(t *testing.T) {
	for _, tc := range testtools.ListOfPoseidon2Testcase {
		t.Run(tc.NameStr, func(t *testing.T) {
			comp := wizard.Compile(func(b *wizard.Builder) { tc.Define(b.CompiledIOP) }, CompileGKRPoseidon2)
			stats := logdata.GetWizardStats(comp)
			fmt.Printf("GKR stats=%+v\n", stats)

			proof := wizard.Prove(comp, tc.Assign)
			err := wizard.Verify(comp, proof)
			require.NoError(t, err)
		})
	}
}
