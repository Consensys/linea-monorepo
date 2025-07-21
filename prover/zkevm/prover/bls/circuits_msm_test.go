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

func TestBlsG1Msm(t *testing.T) {
	limits := &Limits{
		NbG1MulInputInstances:          16,
		NbG1MulCircuitInstances:        1,
		NbG1MembershipInputInstances:   16,
		NbG1MembershipCircuitInstances: 1,
	}
	f, err := os.Open("testdata/bls_g1_msm_inputs.csv")
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
	var blsMsm *BlsMsm
	var blsMsmSource *BlsMsmDataSource
	cmp := wizard.Compile(
		func(b *wizard.Builder) {
			blsMsmSource = &BlsMsmDataSource{
				ID:           ct.GetCommit(b, "ID"),
				CsMul:        ct.GetCommit(b, "CIRCUIT_SELECTOR_G1_MSM"),
				CsMembership: ct.GetCommit(b, "CIRCUIT_SELECTOR_G1_MEMBERSHIP"),
				Limb:         ct.GetCommit(b, "LIMB"),
				Index:        ct.GetCommit(b, "INDEX"),
				Counter:      ct.GetCommit(b, "CT"),
				IsData:       ct.GetCommit(b, "DATA_G1_MSM"),
				IsRes:        ct.GetCommit(b, "RSLT_G1_MSM"),
			}
			blsMsm = newMsm(b.CompiledIOP, G1, limits, blsMsmSource, []query.PlonkOption{query.PlonkRangeCheckOption(16, 6, true)})
		},
		dummy.Compile,
	)

	proof := wizard.Prove(cmp,
		func(run *wizard.ProverRuntime) {
			ct.Assign(run, "ID", "CIRCUIT_SELECTOR_G1_MSM", "CIRCUIT_SELECTOR_G1_MEMBERSHIP", "LIMB", "INDEX", "CT", "DATA_G1_MSM", "RSLT_G1_MSM")
			blsMsm.Assign(run)
		})

	if err := wizard.Verify(cmp, proof); err != nil {
		t.Fatal("proof failed", err)
	}
	t.Log("proof succeeded")
}
