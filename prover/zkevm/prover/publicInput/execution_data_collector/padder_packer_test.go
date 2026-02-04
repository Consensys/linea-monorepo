package execution_data_collector

import (
	"fmt"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	util "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/utilities"
	"testing"
)

func TestDefineAndAssignmentPadderPacker(t *testing.T) {
	testCaseBytes := [][]string{
		{"0x00001234", "0x00005678", "0x00003333", "0x00004567", "0x00004444", "0x00003456", "0x00007891", "0x00002345"},
		{"0x00001111", "0x00005432", "0x00001987", "0x00006543", "0x00002198", "0x00000000", "0x00000000", "0x00000000"},
		{"0x00001929", "0x00003949", "0x00005969", "0x00007989", "0x00001213", "0x00001415", "0x00000000", "0x00000000"},
	}

	testCaseNBytes := []int{
		16, 10, 12, 0,
	}

	t.Run(fmt.Sprintf("testcase"), func(t *testing.T) {

		size := 4
		inputNoBytes := make([]field.Element, size)
		inputIsActive := make([]field.Element, size)
		inputLimbs := make([][]field.Element, common.NbLimbU128)
		for j := 0; j < common.NbLimbU128; j++ {
			inputLimbs[j] = make([]field.Element, size)
		}

		for index := 0; index < size; index++ {
			if testCaseNBytes[index] > 0 {
				inputNoBytes[index].SetUint64(uint64(testCaseNBytes[index]))
				inputIsActive[index].SetOne()
				for j := 0; j < common.NbLimbU128; j++ {
					inputLimbs[j][index] = field.NewFromString(testCaseBytes[index][j])
					bytes := inputLimbs[j][index].Bytes()
					fmt.Println(bytes)
					fmt.Println(utils.HexEncodeToString(bytes[:]))
				}
			}

		}

		fmt.Println("SEPARATION")
		testLimbs := make([]ifaces.Column, 8)
		var (
			testNoBytes, testIsActive ifaces.Column
			ppp                       PadderPacker
		)

		define := func(b *wizard.Builder) {
			for j := 0; j < common.NbLimbU128; j++ {
				testLimbs[j] = util.CreateCol("TEST_PADDER", fmt.Sprintf("PACKER_LIMBS_%d", j), size, b.CompiledIOP)
			}
			testNoBytes = util.CreateCol("TEST_PADDER_PACKER", "NO_BYTES", size, b.CompiledIOP)
			testIsActive = util.CreateCol("TEST_PADDER_PACKER", "IS_ACTIVE", size, b.CompiledIOP)
			ppp = NewPadderPacker(b.CompiledIOP, [8]ifaces.Column(testLimbs), testNoBytes, testIsActive, "TEST_PADDER_PACKER")
			DefinePadderPacker(b.CompiledIOP, ppp, "TEST_PADDER_PACKER")
		}

		prove := func(run *wizard.ProverRuntime) {
			for j := 0; j < common.NbLimbU128; j++ {
				run.AssignColumn(testLimbs[j].GetColID(), smartvectors.NewRegular(inputLimbs[j]))
			}
			run.AssignColumn(testNoBytes.GetColID(), smartvectors.NewRegular(inputNoBytes))
			run.AssignColumn(testIsActive.GetColID(), smartvectors.NewRegular(inputIsActive))
			AssignPadderPacker(run, ppp)

		}

		comp := wizard.Compile(define, dummy.Compile)
		proof := wizard.Prove(comp, prove)
		err := wizard.Verify(comp, proof)

		if err != nil {
			t.Fatalf("verification failed: %v", err)
		}
	})

}
