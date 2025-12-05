package emulated

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/require"
)

func TestEmulatedMultiplication(t *testing.T) {
	nbEntries := 1
	var pa *EmulatedProverAction
	define := func(b *wizard.Builder) {
		P := Limbs{
			Columns: []ifaces.Column{
				b.RegisterCommit("P_0", nbEntries),
				b.RegisterCommit("P_1", nbEntries),
				b.RegisterCommit("P_2", nbEntries),
				b.RegisterCommit("P_3", nbEntries),
				b.RegisterCommit("P_4", nbEntries),
				b.RegisterCommit("P_5", nbEntries),
			},
		}
		A := Limbs{
			Columns: []ifaces.Column{
				b.RegisterCommit("A_0", nbEntries),
				b.RegisterCommit("A_1", nbEntries),
				b.RegisterCommit("A_2", nbEntries),
				b.RegisterCommit("A_3", nbEntries),
				b.RegisterCommit("A_4", nbEntries),
				b.RegisterCommit("A_5", nbEntries),
			},
		}
		B := Limbs{
			Columns: []ifaces.Column{
				b.RegisterCommit("B_0", nbEntries),
				b.RegisterCommit("B_1", nbEntries),
				b.RegisterCommit("B_2", nbEntries),
				b.RegisterCommit("B_3", nbEntries),
				b.RegisterCommit("B_4", nbEntries),
				b.RegisterCommit("B_5", nbEntries),
			},
		}
		pa = EmulatedMultiplication(b.CompiledIOP, A, B, P, 64)
	}

	P := []uint64{13402431016077863595, 2210141511517208575, 7435674573564081700, 7239337960414712511, 5412103778470702295, 1873798617647539866}
	A := []uint64{4511170697608804288, 10450013966173050091, 15052883910335077615, 7458991181107567583, 14001554864696251020, 62079380433102230}
	B := []uint64{17807246233085667035, 4885506624874997090, 7320308865577758342, 3348888139601105932, 627243233652392778, 530851350610086122}

	assignmentP := make([]smartvectors.SmartVector, len(P))
	assignmentA := make([]smartvectors.SmartVector, len(A))
	assignmentB := make([]smartvectors.SmartVector, len(B))
	for i := range P {
		assignmentP[i] = smartvectors.NewRegular([]field.Element{field.NewElement(P[i])})
		assignmentA[i] = smartvectors.NewRegular([]field.Element{field.NewElement(A[i])})
		assignmentB[i] = smartvectors.NewRegular([]field.Element{field.NewElement(B[i])})
	}

	prover := func(run *wizard.ProverRuntime) {
		for i := range P {
			run.AssignColumn(ifaces.ColIDf("P_%d", i), assignmentP[i])
			run.AssignColumn(ifaces.ColIDf("A_%d", i), assignmentA[i])
			run.AssignColumn(ifaces.ColIDf("B_%d", i), assignmentB[i])
		}
		pa.Assign(run)

	}

	comp := wizard.Compile(define, dummy.Compile)
	proof := wizard.Prove(comp, prover)
	err := wizard.Verify(comp, proof)
	require.NoError(t, err)
}
