package bls

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/csvtraces"
)

func TestBlsG1Add(t *testing.T) {
	limits := &Limits{
		NbG1AddInputInstances:          16,
		NbG1AddCircuitInstances:        1,
		NbC1MembershipInputInstances:   16,
		NbC1MembershipCircuitInstances: 1,
	}
	ct := csvtraces.MustOpenCsvFile("testdata/bls_g1_add_input.csv")
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
			blsAdd = newAdd(b.CompiledIOP, G1, limits, blsAddSource, []query.PlonkOption{query.PlonkRangeCheckOption(16, 6, true)})
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

func TestBlsG2Add(t *testing.T) {
	limits := &Limits{
		NbG2AddInputInstances:          16,
		NbG2AddCircuitInstances:        1,
		NbC2MembershipInputInstances:   16,
		NbC2MembershipCircuitInstances: 1,
	}
	ct := csvtraces.MustOpenCsvFile("testdata/bls_g2_add_input.csv")
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
			blsAdd = newAdd(b.CompiledIOP, G2, limits, blsAddSource, []query.PlonkOption{query.PlonkRangeCheckOption(16, 6, true)})
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
