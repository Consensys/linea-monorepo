package ecdsa

import (
	"os"
	"testing"

	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils/csvtraces"
)

func TestEcDataAssignData(t *testing.T) {
	f, err := os.Open("testdata/ecdata_test.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	ct, err := csvtraces.NewCsvTrace(f)
	if err != nil {
		t.Fatal(err)
	}
	limits := &Settings{
		MaxNbEcRecover: 3, // data has two entires, test if can align when have bigger limit
	}
	var ecRec *EcRecover
	var ecSrc *ecDataSource
	cmp := wizard.Compile(
		func(b *wizard.Builder) {
			ecSrc = &ecDataSource{
				CsEcrecover: ct.GetCommit(b, "EC_DATA_CS_ECRECOVER"),
				ID:          ct.GetCommit(b, "EC_DATA_ID"),
				Limb:        ct.GetCommit(b, "EC_DATA_LIMB"),
				SuccessBit:  ct.GetCommit(b, "EC_DATA_SUCCESS_BIT"),
				Index:       ct.GetCommit(b, "EC_DATA_INDEX"),
				IsData:      ct.GetCommit(b, "EC_DATA_IS_DATA"),
				IsRes:       ct.GetCommit(b, "EC_DATA_IS_RES"),
			}
			ecRec = newEcRecover(b.CompiledIOP, limits, ecSrc)
		},
		dummy.Compile,
	)

	proof := wizard.Prove(cmp,
		func(run *wizard.ProverRuntime) {
			ct.Assign(run, "EC_DATA_CS_ECRECOVER", "EC_DATA_ID", "EC_DATA_LIMB", "EC_DATA_SUCCESS_BIT", "EC_DATA_INDEX", "EC_DATA_IS_DATA", "EC_DATA_IS_RES")
			ecRec.Assign(run, ecSrc)
		})

	if err := wizard.Verify(cmp, proof); err != nil {
		t.Fatal("proof failed", err)
	}

	t.Log("proof succeeded")
}
