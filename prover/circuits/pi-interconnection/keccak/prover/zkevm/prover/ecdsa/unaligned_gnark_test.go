package ecdsa

import (
	"os"
	"testing"

	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils/csvtraces"
)

func TestUnalignedGnarkDataAssign(t *testing.T) {
	f, err := os.Open("testdata/unaligned_gnark_test.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	ct, err := csvtraces.NewCsvTrace(f)
	if err != nil {
		t.Fatal(err)
	}
	var uag *UnalignedGnarkData
	var uagSrc *unalignedGnarkDataSource
	cmp := wizard.Compile(func(build *wizard.Builder) {
		uagSrc = &unalignedGnarkDataSource{
			Source:     ct.GetCommit(build, "SOURCE"),
			IsPushing:  ct.GetCommit(build, "IS_PUSHING"),
			IsFetching: ct.GetCommit(build, "IS_FETCHING"),
			IsActive:   ct.GetCommit(build, "IS_ACTIVE"),
			Limb:       ct.GetCommit(build, "LIMB"),
			SuccessBit: ct.GetCommit(build, "SUCCESS_BIT"),
			IsData:     ct.GetCommit(build, "IS_DATA"),
			IsRes:      ct.GetCommit(build, "IS_RES"),
			TxHashHi:   ct.GetCommit(build, "TX_HASH_HI"),
			TxHashLo:   ct.GetCommit(build, "TX_HASH_LO"),
		}
		uag = newUnalignedGnarkData(build.CompiledIOP, ct.LenPadded(), uagSrc)
	}, dummy.Compile)
	proof := wizard.Prove(cmp, func(run *wizard.ProverRuntime) {
		ct.Assign(run,
			"SOURCE",
			"IS_ACTIVE",
			"IS_PUSHING",
			"IS_FETCHING",
			"LIMB",
			"SUCCESS_BIT",
			"IS_DATA",
			"IS_RES",
			"TX_HASH_HI",
			"TX_HASH_LO")
		uag.Assign(run, uagSrc, dummyTxSignatureGetter)
		ct.CheckAssignment(run,
			// TODO: add also auxiliary columns
			string(uag.IsPublicKey.GetColID()),
			string(uag.GnarkIndex.GetColID()),
			string(uag.GnarkData.GetColID()))
	})
	if err := wizard.Verify(cmp, proof); err != nil {
		t.Fatal("proof failed", err)
	}
	t.Log("proof succeeded")
}
