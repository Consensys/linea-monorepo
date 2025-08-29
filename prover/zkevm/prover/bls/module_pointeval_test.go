package bls

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/csvtraces"
)

func testPointEval(t *testing.T, withCircuit bool) {
	limits := &Limits{
		NbPointEvalInputInstances:          2,
		NbPointEvalCircuitInstances:        25,
		NbPointEvalFailureInputInstances:   2,
		NbPointEvalFailureCircuitInstances: 25,
	}
	files, err := filepath.Glob("testdata/bls_pointeval_inputs-[0-9]*.csv")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no csv files found, please run `go generate` to generate the test data")
	}
	// we test all files found
	var cmp *wizard.CompiledIOP
	var bp *BlsPointEval
	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			f, err := os.Open(file)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			ct, err := csvtraces.NewCsvTrace(f)
			if err != nil {
				t.Fatal("failed to create csv trace", err)
			}
			if cmp == nil {
				cmp = wizard.Compile(
					func(b *wizard.Builder) {
						pointEvalSource := &blsPointEvalDataSource{
							ID:                 ct.GetCommit(b, "ID"),
							CsPointEval:        ct.GetCommit(b, "CIRCUIT_SELECTOR_POINT_EVALUATION"),
							CsPointEvalInvalid: ct.GetCommit(b, "CIRCUIT_SELECTOR_POINT_EVALUATION_FAILURE"),
							Limb:               ct.GetCommit(b, "LIMB"),
							Index:              ct.GetCommit(b, "INDEX"),
							Counter:            ct.GetCommit(b, "CT"),
							IsData:             ct.GetCommit(b, "DATA_POINT_EVALUATION"),
							IsRes:              ct.GetCommit(b, "RSLT_POINT_EVALUATION"),
						}
						bp = newPointEval(b.CompiledIOP, limits, pointEvalSource)
						if withCircuit {
							bp = bp.WithPointEvalCircuit(b.CompiledIOP, query.PlonkRangeCheckOption(16, 6, true))
							bp = bp.WithPointEvalFailureCircuit(b.CompiledIOP, query.PlonkRangeCheckOption(16, 6, true))
						}
					},
					dummy.Compile,
				)
			}

			proof := wizard.Prove(cmp,
				func(run *wizard.ProverRuntime) {
					ct.Assign(run, "ID", "CIRCUIT_SELECTOR_POINT_EVALUATION", "CIRCUIT_SELECTOR_POINT_EVALUATION_FAILURE", "INDEX", "CT", "LIMB", "DATA_POINT_EVALUATION", "RSLT_POINT_EVALUATION")
					bp.Assign(run)
				})

			if err := wizard.Verify(cmp, proof); err != nil {
				t.Fatal("proof failed", err)
			}
			t.Log("proof succeeded")
		})
	}
}

func TestPointEvalNoCircuit(t *testing.T) {
	testPointEval(t, false)
}

func TestPointEvalWithCircuit(t *testing.T) {
	testPointEval(t, true)
}
