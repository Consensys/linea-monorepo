package ecdsa

import (
	"fmt"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	"os"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/csvtraces"
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
			SuccessBit: ct.GetCommit(build, "SUCCESS_BIT"),
			IsData:     ct.GetCommit(build, "IS_DATA"),
			IsRes:      ct.GetCommit(build, "IS_RES"),
		}

		for i := 0; i < common.NbLimbU128; i++ {
			uagSrc.Limb[i] = ct.GetCommit(build, fmt.Sprintf("LIMB_%d", i))
		}

		for i := 0; i < common.NbLimbU256; i++ {
			uagSrc.TxHash[i] = ct.GetCommit(build, fmt.Sprintf("TX_HASH_%d", i))
		}

		uag = newUnalignedGnarkData(build.CompiledIOP, ct.LenPadded(), uagSrc)
	}, dummy.Compile)
	proof := wizard.Prove(cmp, func(run *wizard.ProverRuntime) {
		var names = []string{"SOURCE", "IS_ACTIVE", "IS_PUSHING", "IS_FETCHING"}
		for i := 0; i < common.NbLimbU128; i++ {
			names = append(names, fmt.Sprintf("LIMB_%d", i))
		}

		names = append(names, "SUCCESS_BIT", "IS_DATA", "IS_RES")
		for i := 0; i < common.NbLimbU256; i++ {
			names = append(names, fmt.Sprintf("TX_HASH_%d", i))
		}

		ct.Assign(run, names...)
		uag.Assign(run, uagSrc, dummyTxSignatureGetter)

		assignementNames := []string{string(uag.IsPublicKey.GetColID()), string(uag.GnarkIndex.GetColID())}

		for i := 0; i < common.NbLimbU128; i++ {
			assignementNames = append(assignementNames, string(uag.GnarkData[i].GetColID()))
		}

		ct.CheckAssignment(run, assignementNames...)
	})
	if err := wizard.Verify(cmp, proof); err != nil {
		t.Fatal("proof failed", err)
	}
	t.Log("proof succeeded")
}
