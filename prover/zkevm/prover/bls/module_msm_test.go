package bls

import (
	"errors"
	"os"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/csvtraces"
)

func testBlsG1Msm(t *testing.T, withCircuit bool) {
	limits := &Limits{
		NbG1MulInputInstances:          6,
		NbG1MulCircuitInstances:        200,
		NbG1MembershipInputInstances:   6,
		NbG1MembershipCircuitInstances: 300,
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
			blsMsm = newMsm(b.CompiledIOP, G1, limits, blsMsmSource)
			if withCircuit {
				blsMsm = blsMsm.
					WithGroupMembershipCircuit(b.CompiledIOP, query.PlonkRangeCheckOption(16, 6, true)).
					WithMsmCircuit(b.CompiledIOP, query.PlonkRangeCheckOption(16, 6, true))
			}
		},
		dummy.Compile,
	)

	fout, err := os.Create("testdata/bls_g1_msm_output.csv")
	if err != nil {
		t.Fatal("failed to create output csv file", err)
	}
	defer fout.Close()
	fout2, err := os.Create("testdata/bls_g1_msm_output2.csv")
	if err != nil {
		t.Fatal("failed to create output csv file", err)
	}
	defer fout2.Close()
	proof := wizard.Prove(cmp,
		func(run *wizard.ProverRuntime) {
			ct.Assign(run, "ID", "CIRCUIT_SELECTOR_G1_MSM", "CIRCUIT_SELECTOR_G1_MEMBERSHIP", "LIMB", "INDEX", "CT", "DATA_G1_MSM", "RSLT_G1_MSM")
			blsMsm.Assign(run)
			csvtraces.FmtCsv(fout, run,
				[]ifaces.Column{
					blsMsm.unalignedMsmData.IsActive,
					blsMsm.unalignedMsmData.IsFirstLine,
					blsMsm.unalignedMsmData.IsLastLine,
					blsMsm.unalignedMsmData.Scalar[0],
					blsMsm.unalignedMsmData.Scalar[1],
					blsMsm.unalignedMsmData.Point[0],
					blsMsm.unalignedMsmData.Point[1],
					blsMsm.unalignedMsmData.Point[2],
					blsMsm.unalignedMsmData.Point[3],
					blsMsm.unalignedMsmData.Point[4],
					blsMsm.unalignedMsmData.Point[5],
					blsMsm.unalignedMsmData.Point[6],
					blsMsm.unalignedMsmData.Point[7],
					blsMsm.unalignedMsmData.CurrentAccumulator[0],
					blsMsm.unalignedMsmData.CurrentAccumulator[1],
					blsMsm.unalignedMsmData.CurrentAccumulator[2],
					blsMsm.unalignedMsmData.CurrentAccumulator[3],
					blsMsm.unalignedMsmData.CurrentAccumulator[4],
					blsMsm.unalignedMsmData.CurrentAccumulator[5],
					blsMsm.unalignedMsmData.CurrentAccumulator[6],
					blsMsm.unalignedMsmData.CurrentAccumulator[7],
					blsMsm.unalignedMsmData.NextAccumulator[0],
					blsMsm.unalignedMsmData.NextAccumulator[1],
					blsMsm.unalignedMsmData.NextAccumulator[2],
					blsMsm.unalignedMsmData.NextAccumulator[3],
					blsMsm.unalignedMsmData.NextAccumulator[4],
					blsMsm.unalignedMsmData.NextAccumulator[5],
					blsMsm.unalignedMsmData.NextAccumulator[6],
					blsMsm.unalignedMsmData.NextAccumulator[7],
				},
				[]csvtraces.Option{csvtraces.SkipPrepaddingZero},
			)
			csvtraces.FmtCsv(fout2, run,
				[]ifaces.Column{
					blsMsm.unalignedMsmData.GnarkIsActiveMsm,
					blsMsm.unalignedMsmData.GnarkDataMsm,
				},
				[]csvtraces.Option{csvtraces.SkipPrepaddingZero},
			)
		})
	if err := wizard.Verify(cmp, proof); err != nil {
		t.Fatal("proof failed", err)
	}
	t.Log("proof succeeded")
}

func testBlsG2Msm(t *testing.T, withCircuit bool) {
	limits := &Limits{
		NbG2MulInputInstances:          6,
		NbG2MulCircuitInstances:        200,
		NbG2MembershipInputInstances:   6,
		NbG2MembershipCircuitInstances: 300,
	}
	f, err := os.Open("testdata/bls_g2_msm_inputs.csv")
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
				CsMul:        ct.GetCommit(b, "CIRCUIT_SELECTOR_G2_MSM"),
				CsMembership: ct.GetCommit(b, "CIRCUIT_SELECTOR_G2_MEMBERSHIP"),
				Limb:         ct.GetCommit(b, "LIMB"),
				Index:        ct.GetCommit(b, "INDEX"),
				Counter:      ct.GetCommit(b, "CT"),
				IsData:       ct.GetCommit(b, "DATA_G2_MSM"),
				IsRes:        ct.GetCommit(b, "RSLT_G2_MSM"),
			}
			blsMsm = newMsm(b.CompiledIOP, G2, limits, blsMsmSource)
			if withCircuit {
				blsMsm = blsMsm.
					WithGroupMembershipCircuit(b.CompiledIOP, query.PlonkRangeCheckOption(16, 6, true)).
					WithMsmCircuit(b.CompiledIOP, query.PlonkRangeCheckOption(16, 6, true))
			}
		},
		dummy.Compile,
	)

	fout, err := os.Create("testdata/bls_g2_msm_output.csv")
	if err != nil {
		t.Fatal("failed to create output csv file", err)
	}
	defer fout.Close()
	fout2, err := os.Create("testdata/bls_g2_msm_output2.csv")
	if err != nil {
		t.Fatal("failed to create output csv file", err)
	}
	defer fout2.Close()
	proof := wizard.Prove(cmp,
		func(run *wizard.ProverRuntime) {
			ct.Assign(run, "ID", "CIRCUIT_SELECTOR_G2_MSM", "CIRCUIT_SELECTOR_G2_MEMBERSHIP", "LIMB", "INDEX", "CT", "DATA_G2_MSM", "RSLT_G2_MSM")
			blsMsm.Assign(run)
			csvtraces.FmtCsv(fout, run,
				[]ifaces.Column{
					blsMsm.unalignedMsmData.IsActive,
					blsMsm.unalignedMsmData.IsFirstLine,
					blsMsm.unalignedMsmData.IsLastLine,
					blsMsm.unalignedMsmData.Scalar[0],
					blsMsm.unalignedMsmData.Scalar[1],
					blsMsm.unalignedMsmData.Point[0],
					blsMsm.unalignedMsmData.Point[1],
					blsMsm.unalignedMsmData.Point[2],
					blsMsm.unalignedMsmData.Point[3],
					blsMsm.unalignedMsmData.Point[4],
					blsMsm.unalignedMsmData.Point[5],
					blsMsm.unalignedMsmData.Point[6],
					blsMsm.unalignedMsmData.Point[7],
					blsMsm.unalignedMsmData.CurrentAccumulator[0],
					blsMsm.unalignedMsmData.CurrentAccumulator[1],
					blsMsm.unalignedMsmData.CurrentAccumulator[2],
					blsMsm.unalignedMsmData.CurrentAccumulator[3],
					blsMsm.unalignedMsmData.CurrentAccumulator[4],
					blsMsm.unalignedMsmData.CurrentAccumulator[5],
					blsMsm.unalignedMsmData.CurrentAccumulator[6],
					blsMsm.unalignedMsmData.CurrentAccumulator[7],
					blsMsm.unalignedMsmData.NextAccumulator[0],
					blsMsm.unalignedMsmData.NextAccumulator[1],
					blsMsm.unalignedMsmData.NextAccumulator[2],
					blsMsm.unalignedMsmData.NextAccumulator[3],
					blsMsm.unalignedMsmData.NextAccumulator[4],
					blsMsm.unalignedMsmData.NextAccumulator[5],
					blsMsm.unalignedMsmData.NextAccumulator[6],
					blsMsm.unalignedMsmData.NextAccumulator[7],
				},
				[]csvtraces.Option{csvtraces.SkipPrepaddingZero},
			)
			csvtraces.FmtCsv(fout2, run,
				[]ifaces.Column{
					blsMsm.unalignedMsmData.GnarkIsActiveMsm,
					blsMsm.unalignedMsmData.GnarkDataMsm,
				},
				[]csvtraces.Option{csvtraces.SkipPrepaddingZero},
			)
		})
	if err := wizard.Verify(cmp, proof); err != nil {
		t.Fatal("proof failed", err)
	}
	t.Log("proof succeeded")
}

func TestBlsG1MsmNoCircuit(t *testing.T) {
	testBlsG1Msm(t, false)
}

func TestBlsG1MsmWithCircuit(t *testing.T) {
	testBlsG1Msm(t, true)
}

func TestBlsG2MsmNoCircuit(t *testing.T) {
	testBlsG2Msm(t, false)
}

func TestBlsG2MsmWithCircuit(t *testing.T) {
	testBlsG2Msm(t, true)
}
