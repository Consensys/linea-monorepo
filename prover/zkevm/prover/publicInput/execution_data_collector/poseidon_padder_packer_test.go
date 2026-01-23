package execution_data_collector

import (
	"fmt"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	util "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/utilities"
	"testing"
)

func TestDefineAndAssignmentPoseidonPadderPacker(t *testing.T) {

	testCase := [][]smartvectors.SmartVector{
		/*		{
					smartvectors.ForTest(1, 2),
					smartvectors.ForTest(1, 1),
				},
				{
					smartvectors.ForTest(1, 2, 3, 4),
					smartvectors.ForTest(1, 1, 1, 1),
				},
				{
					smartvectors.ForTest(1, 2, 3, 4, 5, 6, 7, 8),
					smartvectors.ForTest(1, 1, 1, 1, 1, 1, 1, 1),
				},*/
		{
			smartvectors.ForTest(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16),
			smartvectors.ForTest(1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0),
		},
	}

	for i, tc := range testCase {

		t.Run(fmt.Sprintf("testcase-%v", i), func(t *testing.T) {

			var (
				testCol, testFilter ifaces.Column
				ppp                 PoseidonPadderPacker
			)

			define := func(b *wizard.Builder) {
				testCol = util.CreateCol("TEST_POSEIDON_PADDER", "PACKER", tc[0].Len(), b.CompiledIOP)
				testFilter = util.CreateCol("TEST_POSEIDON_PADDER_PACKER", "FILTER", tc[0].Len(), b.CompiledIOP)
				ppp = NewPoseidonPadderPacker(b.CompiledIOP, testCol, testFilter, "TEST_POSEIDON_PADDER_PACKER")
				DefinePoseidonPadderPacker(b.CompiledIOP, ppp, "TEST_POSEIDON_PADDER_PACKER")
			}

			prove := func(run *wizard.ProverRuntime) {
				run.AssignColumn("TEST_POSEIDON_PADDER_PACKER", tc[0])
				run.AssignColumn("TEST_POSEIDON_PADDER_PACKER_FILTER", tc[1])
				AssignPoseidonPadderPacker(run, ppp)

			}

			comp := wizard.Compile(define, dummy.Compile)
			proof := wizard.Prove(comp, prove)
			err := wizard.Verify(comp, proof)

			if err != nil {
				t.Fatalf("verification failed: %v", err)
			}
		})
	}

}
