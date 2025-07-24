package bls

import (
	"errors"
	"os"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/csvtraces"
)

func testBlsG1Add(t *testing.T, withCircuit bool) {
	limits := &Limits{
		NbG1AddInputInstances:          16,
		NbG1AddCircuitInstances:        1,
		NbC1MembershipInputInstances:   16,
		NbC1MembershipCircuitInstances: 1,
	}
	f, err := os.Open("testdata/bls_g1_add_input.csv")
	if errors.Is(err, os.ErrNotExist) {
		t.Fatal("csv file does not exist, please run `go generate` to generate the test data")
	}
	if err != nil {
		t.Fatal("failed to open csv file", err)
	}
	defer f.Close()
	ct, err := csvtraces.NewCsvTrace(f)
	if err != nil {
		t.Fatal("failed to create csv trace", err)
	}
	var blsAdd *BlsAdd
	var blsAddSource *blsAddDataSource
	cmp := wizard.Compile(
		func(b *wizard.Builder) {
			blsAddSource = &blsAddDataSource{
				CsAdd:             ct.GetCommit(b, "CIRCUIT_SELECTOR_G1_ADD"),
				Limb:              ct.GetCommit(b, "LIMB"),
				Index:             ct.GetCommit(b, "INDEX"),
				Counter:           ct.GetCommit(b, "CT"),
				CsCurveMembership: ct.GetCommit(b, "CIRCUIT_SELECTOR_C1_MEMBERSHIP"),
				IsData:            ct.GetCommit(b, "DATA_G1_ADD"),
				IsRes:             ct.GetCommit(b, "RSLT_G1_ADD"),
			}
			blsAdd = newAdd(b.CompiledIOP, G1, limits, blsAddSource)
			if withCircuit {
				blsAdd = blsAdd.
					WithAddCircuit(b.CompiledIOP, query.PlonkRangeCheckOption(16, 6, true)).
					WithCurveMembershipCircuit(b.CompiledIOP, query.PlonkRangeCheckOption(16, 6, true))
			}
		},
		dummy.Compile,
	)

	proof := wizard.Prove(cmp,
		func(run *wizard.ProverRuntime) {
			ct.Assign(run, "CIRCUIT_SELECTOR_G1_ADD", "LIMB", "INDEX", "CT", "CIRCUIT_SELECTOR_C1_MEMBERSHIP", "DATA_G1_ADD", "RSLT_G1_ADD")
			blsAdd.Assign(run)
		})

	if err := wizard.Verify(cmp, proof); err != nil {
		t.Fatal("proof failed", err)
	}
	t.Log("proof succeeded")
}

func testBlsG2Add(t *testing.T, withCircuit bool) {
	limits := &Limits{
		NbG2AddInputInstances:          16,
		NbG2AddCircuitInstances:        1,
		NbC2MembershipInputInstances:   16,
		NbC2MembershipCircuitInstances: 1,
	}
	f, err := os.Open("testdata/bls_g2_add_input.csv")
	if errors.Is(err, os.ErrNotExist) {
		t.Fatal("csv file does not exist, please run `go generate` to generate the test data")
	}
	if err != nil {
		t.Fatal("failed to open csv file", err)
	}
	defer f.Close()
	ct, err := csvtraces.NewCsvTrace(f)
	if err != nil {
		t.Fatal("failed to create csv trace", err)
	}
	var blsAdd *BlsAdd
	var blsAddSource *blsAddDataSource
	cmp := wizard.Compile(
		func(b *wizard.Builder) {
			blsAddSource = &blsAddDataSource{
				CsAdd:             ct.GetCommit(b, "CIRCUIT_SELECTOR_G2_ADD"),
				Limb:              ct.GetCommit(b, "LIMB"),
				Index:             ct.GetCommit(b, "INDEX"),
				Counter:           ct.GetCommit(b, "CT"),
				CsCurveMembership: ct.GetCommit(b, "CIRCUIT_SELECTOR_C2_MEMBERSHIP"),
				IsData:            ct.GetCommit(b, "DATA_G2_ADD"),
				IsRes:             ct.GetCommit(b, "RSLT_G2_ADD"),
			}
			blsAdd = newAdd(b.CompiledIOP, G2, limits, blsAddSource)
			if withCircuit {
				blsAdd = blsAdd.
					WithAddCircuit(b.CompiledIOP, query.PlonkRangeCheckOption(16, 6, true)).
					WithCurveMembershipCircuit(b.CompiledIOP, query.PlonkRangeCheckOption(16, 6, true))
			}
		},
		dummy.Compile,
	)

	proof := wizard.Prove(cmp,
		func(run *wizard.ProverRuntime) {
			ct.Assign(run, "CIRCUIT_SELECTOR_G2_ADD", "LIMB", "INDEX", "CT", "CIRCUIT_SELECTOR_C2_MEMBERSHIP", "DATA_G2_ADD", "RSLT_G2_ADD")
			blsAdd.Assign(run)
		})

	if err := wizard.Verify(cmp, proof); err != nil {
		t.Fatal("proof failed", err)
	}
	t.Log("proof succeeded")
}

func TestBlsG1AddNoCircuit(t *testing.T) {
	testBlsG1Add(t, false)
}

func TestBlsG1AddWithCircuit(t *testing.T) {
	testBlsG1Add(t, true)
}

func TestBlsG2AddNoCircuit(t *testing.T) {
	testBlsG2Add(t, false)
}

func TestBlsG2AddWithCircuit(t *testing.T) {
	testBlsG2Add(t, true)
}
