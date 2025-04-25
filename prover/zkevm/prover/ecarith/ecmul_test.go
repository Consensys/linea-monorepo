//go:build !fuzzlight

package ecarith

import (
	"os"
	"testing"

	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/test"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/csvtraces"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

var (

	// 8812078025088706800239420563438698135256154119486538953196282662928678739599
	dummyPxLo = [8]frontend.Variable{54522, 64063, 19107, 64140, 47249, 36866, 4929, 24207}
	dummyPxHi = [8]frontend.Variable{4987, 30108, 7206, 43510, 4530, 21165, 28317, 38023}

	// 19384817048585926091202754628019327561583315513931387211365745455438290666381
	dummyPyLo = [8]frontend.Variable{21413, 8925, 46471, 63591, 44678, 62564, 1446, 50061}
	dummyPyHi = [8]frontend.Variable{10971, 27370, 17409, 54880, 14499, 22162, 18782, 17869}

	// 6392244365320561037829193157981140850021222049710818941199853670416625377985
	dummyScalarLo = [8]frontend.Variable{47334, 19682, 59651, 5498, 64385, 42173, 9491, 8897}
	dummyScalarHi = [8]frontend.Variable{3617, 57809, 10842, 19675, 18737, 65290, 60383, 57066}

	// 18286867191893102909980352899825705524248643499324573950097278735467138404960
	dummyRxLo = [8]frontend.Variable{22472, 439, 2429, 45020, 55129, 27384, 35002, 21088}
	dummyRxHi = [8]frontend.Variable{10350, 2, 42581, 19814, 12263, 19537, 62766, 34029}

	// 9901600500795379168368190931083365122151061434285910708538887599132533626517
	dummyRyLo = [8]frontend.Variable{8524, 41907, 60473, 20665, 11836, 12273, 58948, 38549}
	dummyRyHi = [8]frontend.Variable{5604, 7030, 51904, 28067, 12228, 33396, 63525, 54713}
)

func TestMultiEcMulCircuit(t *testing.T) {
	limits := &Limits{
		NbInputInstances:   2,
		NbCircuitInstances: 0, // not used in this test
	}
	circuit := NewECMulCircuit(limits)
	assignment := NewECMulCircuit(limits)
	instanceAssignment := ECMulInstance{
		P_X_hi: dummyPxHi,
		P_X_lo: dummyPxLo,
		P_Y_hi: dummyPyHi,
		P_Y_lo: dummyPyLo,

		R_X_hi: dummyRxHi,
		R_X_lo: dummyRxLo,
		R_Y_hi: dummyRyHi,
		R_Y_lo: dummyRyLo,

		N_hi: dummyScalarHi,
		N_lo: dummyScalarLo,
	}
	for i := 0; i < limits.NbInputInstances; i++ {
		assignment.Instances[i] = instanceAssignment
	}

	// 403569 constraints
	assert := test.NewAssert(t)
	assert.SolvingSucceeded(
		circuit,
		assignment,
		test.NoFuzzing(),
		test.NoSerializationChecks(),
		test.WithBackends(backend.PLONK),
		test.WithCurves(ecc.BLS12_377))

}

func TestEcMulIntegration(t *testing.T) {
	f, err := os.Open("testdata/ecmul_test.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	limits := &Limits{
		NbInputInstances:   3,
		NbCircuitInstances: 2,
	}
	ct, err := csvtraces.NewCsvTrace(f, csvtraces.WithNbRows(limits.sizeEcMulIntegration()))
	if err != nil {
		t.Fatal(err)
	}
	var ecMul *EcMul
	var ecMulSource *EcDataMulSource
	cmp := wizard.Compile(
		func(b *wizard.Builder) {
			ecMulSource = &EcDataMulSource{
				CsEcMul: ct.GetCommit(b, "CS_MUL"),
				Index:   ct.GetCommit(b, "INDEX"),
				IsData:  ct.GetCommit(b, "IS_DATA"),
				IsRes:   ct.GetCommit(b, "IS_RES"),
			}

			for i := 0; i < common.NbFlattenColLimbs; i++ {
				ecMulSource.Limbs[i] = ct.GetCommit(b, fmt.Sprintf("LIMB_%d", i))
			}

			ecMul = newEcMul(b.CompiledIOP, limits, ecMulSource, []query.PlonkOption{query.PlonkRangeCheckOption(16, 6, true)})
		},
		dummy.Compile,
	)

	proof := wizard.Prove(cmp,
		func(run *wizard.ProverRuntime) {
			ct.Assign(run, "CS_MUL", "LIMB_0", "LIMB_1", "LIMB_2", "LIMB_3", "LIMB_4", "LIMB_5", "LIMB_6", "LIMB_7", "INDEX", "IS_DATA", "IS_RES")
			ecMul.Assign(run)
		})

	if err := wizard.Verify(cmp, proof); err != nil {
		t.Fatal("proof failed", err)
	}

	t.Log("proof succeeded")
}
