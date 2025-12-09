//go:build !fuzzlight

package ecarith

import (
	"os"
	"testing"

	"fmt"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/plonkinternal"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/csvtraces"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

var (

	// 0x137b759c1c26a9f611b252ad6e9d9487d4fafa3f4aa3fa8cb891900213415e8f
	dummyPx = [16]frontend.Variable{
		0x137b, 0x759c, 0x1c26, 0xa9f6,
		0x11b2, 0x52ad, 0x6e9d, 0x9487,
		0xd4fa, 0xfa3f, 0x4aa3, 0xfa8c,
		0xb891, 0x9002, 0x1341, 0x5e8f,
	}

	// 0x2adb6aea4401d66038a35692495e45cd53a522ddb587f867ae86f46405a6c38d
	dummyPy = [16]frontend.Variable{
		0x2adb, 0x6aea, 0x4401, 0xd660,
		0x38a3, 0x5692, 0x495e, 0x45cd,
		0x53a5, 0x22dd, 0xb587, 0xf867,
		0xae86, 0xf464, 0x05a6, 0xc38d,
	}

	// 0xe21e1d12a5a4cdb4931ff0aebdfdeeab8e64ce2e903157afb81a4bd251322c1
	dummyN = [16]frontend.Variable{
		0x0e21, 0xe1d1, 0x2a5a, 0x4cdb,
		0x4931, 0xff0a, 0xebdf, 0xdeea,
		0xb8e6, 0x4ce2, 0xe903, 0x157a,
		0xfb81, 0xa4bd, 0x2513, 0x22c1,
	}

	// 0x286e0002a6554d662fe74c51f52e84ed57c801b7097dafdcd7596af888ba5260
	dummyRx = [16]frontend.Variable{
		0x286e, 0x0002, 0xa655, 0x4d66,
		0x2fe7, 0x4c51, 0xf52e, 0x84ed,
		0x57c8, 0x01b7, 0x097d, 0xafdc,
		0xd759, 0x6af8, 0x88ba, 0x5260,
	}

	// 0x15e41b76cac06da32fc48274f825d5b9214ca3b3ec3950b92e3c2ff1e6449695
	dummyRy = [16]frontend.Variable{
		0x15e4, 0x1b76, 0xcac0, 0x6da3,
		0x2fc4, 0x8274, 0xf825, 0xd5b9,
		0x214c, 0xa3b3, 0xec39, 0x50b9,
		0x2e3c, 0x2ff1, 0xe644, 0x9695,
	}
)

func TestMultiEcMulCircuit(t *testing.T) {
	limits := &Limits{
		NbInputInstances:   2,
		NbCircuitInstances: 0, // not used in this test
	}
	circuit := NewECMulCircuit(limits)
	assignment := NewECMulCircuit(limits)
	instanceAssignment := ECMulInstance{
		P_X: dummyPx,
		P_Y: dummyPy,
		R_X: dummyRx,
		R_Y: dummyRy,
		N:   dummyN,
	}
	for i := 0; i < limits.NbInputInstances; i++ {
		assignment.Instances[i] = instanceAssignment
	}

	// 403569 constraints
	builder, rcGetter := plonkinternal.NewExternalRangeCheckerBuilder(true)
	ccs, err := frontend.CompileU32(koalabear.Modulus(), builder, circuit)
	if err != nil {
		t.Fatalf("compiling circuit: %v", err)
	}

	_ = rcGetter() // to not leak the channel

	wit, err := frontend.NewWitness(circuit, koalabear.Modulus())
	if err != nil {
		t.Fatalf("assigning witness: %v", err)
	}

	if err := ccs.IsSolved(wit); err != nil {
		t.Fatalf("solving circuit: %v", err)
	}

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

			for i := 0; i < common.NbLimbU128; i++ {
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
