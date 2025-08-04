package bls

import (
	"os"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/csvtraces"
)

func testBlsG1Map(t *testing.T, withCircuit bool) {
	limits := &Limits{
		NbG1MapToInputInstances:   16,
		NbG1MapToCircuitInstances: 1,
	}
	f, err := os.Open("testdata/bls_g1_map_inputs.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	ct, err := csvtraces.NewCsvTrace(f)
	if err != nil {
		t.Fatal("failed to create csv trace", err)
	}
	var blsMap *BlsMap
	cmp := wizard.Compile(
		func(b *wizard.Builder) {
			blsMapSource := &blsMapDataSource{
				ID:      ct.GetCommit(b, "ID"),
				CsMap:   ct.GetCommit(b, "CIRCUIT_SELECTOR_MAP_FP_TO_G1"),
				Index:   ct.GetCommit(b, "INDEX"),
				Counter: ct.GetCommit(b, "CT"),
				Limb:    ct.GetCommit(b, "LIMB"),
				IsData:  ct.GetCommit(b, "DATA_MAP_FP_TO_G1"),
				IsRes:   ct.GetCommit(b, "RSLT_MAP_FP_TO_G1"),
			}
			blsMap = newMap(b.CompiledIOP, G1, limits, blsMapSource)
			if withCircuit {
				blsMap = blsMap.WithMapCircuit(b.CompiledIOP, query.PlonkRangeCheckOption(16, 6, true))
			}
		},
		dummy.Compile,
	)

	proof := wizard.Prove(cmp,
		func(run *wizard.ProverRuntime) {
			ct.Assign(run, "ID", "CIRCUIT_SELECTOR_MAP_FP_TO_G1", "INDEX", "CT", "LIMB", "DATA_MAP_FP_TO_G1", "RSLT_MAP_FP_TO_G1")
			blsMap.Assign(run)
		})

	if err := wizard.Verify(cmp, proof); err != nil {
		t.Fatal("proof failed", err)
	}
	t.Log("proof succeeded")
}

func TestBlsMapG1NoCircuit(t *testing.T) {
	testBlsG1Map(t, false)
}

func TestBlsMapG1WithCircuit(t *testing.T) {
	testBlsG1Map(t, true)
}

func testBlsG2Map(t *testing.T, withCircuit bool) {
	limits := &Limits{
		NbG2MapToInputInstances:   4,
		NbG2MapToCircuitInstances: 1,
	}
	f, err := os.Open("testdata/bls_g2_map_inputs.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	ct, err := csvtraces.NewCsvTrace(f)
	if err != nil {
		t.Fatal("failed to create csv trace", err)
	}
	var blsMap *BlsMap
	cmp := wizard.Compile(
		func(b *wizard.Builder) {
			blsMapSource := &blsMapDataSource{
				ID:      ct.GetCommit(b, "ID"),
				CsMap:   ct.GetCommit(b, "CIRCUIT_SELECTOR_MAP_FP2_TO_G2"),
				Index:   ct.GetCommit(b, "INDEX"),
				Counter: ct.GetCommit(b, "CT"),
				Limb:    ct.GetCommit(b, "LIMB"),
				IsData:  ct.GetCommit(b, "DATA_MAP_FP2_TO_G2"),
				IsRes:   ct.GetCommit(b, "RSLT_MAP_FP2_TO_G2"),
			}
			blsMap = newMap(b.CompiledIOP, G2, limits, blsMapSource)
			if withCircuit {
				blsMap = blsMap.WithMapCircuit(b.CompiledIOP, query.PlonkRangeCheckOption(16, 6, true))
			}
		},
		dummy.Compile,
	)

	proof := wizard.Prove(cmp,
		func(run *wizard.ProverRuntime) {
			ct.Assign(run, "ID", "CIRCUIT_SELECTOR_MAP_FP2_TO_G2", "INDEX", "CT", "LIMB", "DATA_MAP_FP2_TO_G2", "RSLT_MAP_FP2_TO_G2")
			blsMap.Assign(run)
		})

	if err := wizard.Verify(cmp, proof); err != nil {
		t.Fatal("proof failed", err)
	}
	t.Log("proof succeeded")
}

func TestBlsMapG2NoCircuit(t *testing.T) {
	testBlsG2Map(t, false)
}

func TestBlsMapG2WithCircuit(t *testing.T) {
	testBlsG2Map(t, true)
}
