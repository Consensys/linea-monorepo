package ecpair

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/plonk"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/csvtraces"
)

type pairingDataTestCase struct {
	InputFName, ModuleFName                      string
	NbMillerLoops, NbFinalExps, NbSubgroupChecks int
}

var pairingDataTestCases = []pairingDataTestCase{
	{
		InputFName:    "testdata/ecpair_double_pair_input.csv",
		ModuleFName:   "testdata/ecpair_double_pair_module.csv",
		NbMillerLoops: 1,
		NbFinalExps:   1,
	},
	{
		InputFName:    "testdata/ecpair_two_pairings_input.csv",
		ModuleFName:   "testdata/ecpair_two_pairings_module.csv",
		NbMillerLoops: 2,
		NbFinalExps:   2,
	},
	{
		InputFName:    "testdata/ecpair_failing_double_pair_input.csv",
		ModuleFName:   "testdata/ecpair_failing_double_pair_module.csv",
		NbMillerLoops: 1,
		NbFinalExps:   1,
	},
	{
		InputFName:    "testdata/ecpair_triple_pair_input.csv",
		ModuleFName:   "testdata/ecpair_triple_pair_module.csv",
		NbMillerLoops: 2,
		NbFinalExps:   1,
	},
	{
		// test case with bigger limits than inputs. Tests that input fillers work correctly
		InputFName:    "testdata/ecpair_double_pair_input.csv",
		ModuleFName:   "testdata/ecpair_double_pair_module.csv",
		NbMillerLoops: 2,
		NbFinalExps:   2,
	},
	{
		// empty input to test edge case and input fillers
		InputFName:    "testdata/ecpair_empty.csv",
		ModuleFName:   "",
		NbMillerLoops: 2,
		NbFinalExps:   2,
	},
	{
		// trace test
		InputFName:    "testdata/ecpair_trace_input.csv",
		ModuleFName:   "",
		NbMillerLoops: 2,
		NbFinalExps:   1,
	},
}

func testModule(t *testing.T, tc pairingDataTestCase, withPairingCircuit, withG2MembershipCircuit bool, checkPairingModule, checkSubgroupModule bool) {
	t.Run(tc.InputFName, func(t *testing.T) {

		var (
			inp   *ECPairSource
			mod   *ECPair
			modCt *csvtraces.CsvTrace
			inpCt = csvtraces.MustOpenCsvFile(tc.InputFName)
		)
		if tc.ModuleFName == "" {
			checkPairingModule = false
			checkSubgroupModule = false
		}
		if checkPairingModule || checkSubgroupModule {
			modCt = csvtraces.MustOpenCsvFile(tc.ModuleFName)
		}

		limits := &Limits{
			NbMillerLoopInputInstances:   tc.NbMillerLoops,
			NbFinalExpInputInstances:     tc.NbFinalExps,
			NbG2MembershipInputInstances: tc.NbSubgroupChecks,
		}
		if tc.NbMillerLoops > 0 {
			limits.NbMillerLoopCircuits = 1
		}
		if tc.NbFinalExps > 0 {
			limits.NbFinalExpCircuits = 1
		}
		if tc.NbSubgroupChecks > 0 {
			limits.NbG2MembershipCircuits = 1
		}

		cmp := wizard.Compile(func(build *wizard.Builder) {
			inp = &ECPairSource{
				ID:                inpCt.GetCommit(build, "ECDATA_ID"),
				Index:             inpCt.GetCommit(build, "ECDATA_INDEX"),
				CsEcpairing:       inpCt.GetCommit(build, "ECDATA_CS_PAIRING"),
				CsG2Membership:    inpCt.GetCommit(build, "ECDATA_CS_G2_MEMBERSHIP"),
				IsEcPairingData:   inpCt.GetCommit(build, "ECDATA_IS_DATA"),
				IsEcPairingResult: inpCt.GetCommit(build, "ECDATA_IS_RES"),
				Limb:              inpCt.GetCommit(build, "ECDATA_LIMB"),
				SuccessBit:        inpCt.GetCommit(build, "ECDATA_SUCCESS_BIT"),
				AccPairings:       inpCt.GetCommit(build, "ECDATA_ACC_PAIRINGS"),
				TotalPairings:     inpCt.GetCommit(build, "ECDATA_TOTAL_PAIRINGS"),
			}

			mod = newECPair(build.CompiledIOP, limits, inp)
			if withPairingCircuit {
				mod.WithPairingCircuit(build.CompiledIOP, plonk.WithRangecheck(16, 6, false))
			}
			if withG2MembershipCircuit {
				mod.WithG2MembershipCircuit(build.CompiledIOP, plonk.WithRangecheck(16, 6, false))
			}
		}, dummy.Compile)

		proof := wizard.Prove(cmp, func(run *wizard.ProverRuntime) {
			inpCt.Assign(run,
				"ECDATA_ID",
				"ECDATA_INDEX",
				"ECDATA_CS_PAIRING",
				"ECDATA_CS_G2_MEMBERSHIP",
				"ECDATA_IS_DATA",
				"ECDATA_IS_RES",
				"ECDATA_LIMB",
				"ECDATA_SUCCESS_BIT",
				"ECDATA_ACC_PAIRINGS",
				"ECDATA_TOTAL_PAIRINGS",
			)

			mod.Assign(run)

			if checkPairingModule {
				modCt.CheckAssignment(run,
					"ECPAIR_IS_ACTIVE",
					"ECPAIR_UNALIGNED_PAIRING_DATA_IS_ACTIVE",
					"ECPAIR_UNALIGNED_PAIRING_DATA_INDEX",
					"ECPAIR_UNALIGNED_PAIRING_DATA_INSTANCE_ID",
					"ECPAIR_UNALIGNED_PAIRING_DATA_PAIR_ID",
					"ECPAIR_UNALIGNED_PAIRING_DATA_TOTAL_PAIRS",
					"ECPAIR_UNALIGNED_PAIRING_DATA_IS_COMPUTED",
					"ECPAIR_UNALIGNED_PAIRING_DATA_IS_PULLING",
					"ECPAIR_UNALIGNED_PAIRING_DATA_IS_FIRST_LINE_OF_INSTANCE",
					"ECPAIR_UNALIGNED_PAIRING_DATA_IS_ACCUMULATOR_INIT",
					"ECPAIR_UNALIGNED_PAIRING_DATA_IS_FIRST_LINE_OF_PREV_ACC",
					"ECPAIR_UNALIGNED_PAIRING_DATA_IS_ACCUMULATOR_PREV",
					"ECPAIR_UNALIGNED_PAIRING_DATA_IS_FIRST_LINE_OF_CURR_ACC",
					"ECPAIR_UNALIGNED_PAIRING_DATA_IS_ACCUMULATOR_CURR",
					"ECPAIR_UNALIGNED_PAIRING_DATA_LIMB",
					"ECPAIR_UNALIGNED_PAIRING_DATA_TO_MILLER_LOOP_CIRCUIT",
					"ECPAIR_UNALIGNED_PAIRING_DATA_TO_FINAL_EXP_CIRCUIT",
				)
			}
			if checkSubgroupModule {
				modCt.CheckAssignment(run,
					"ECPAIR_IS_ACTIVE",
					"ECPAIR_UNALIGNED_G2_DATA_IS_PULLING",
					"ECPAIR_UNALIGNED_G2_DATA_IS_COMPUTED",
					"ECPAIR_UNALIGNED_G2_DATA_LIMB",
					"ECPAIR_UNALIGNED_G2_DATA_TO_G2_MEMBERSHIP_CIRCUIT",
					"ECPAIR_UNALIGNED_G2_DATA_SUCCESS_BIT",
				)
			}
		})

		if err := wizard.Verify(cmp, proof); err != nil {
			t.Fatal("proof failed", err)
		}

		t.Log("proof succeeded")
	})
}

func TestPairingData(t *testing.T) {
	for _, tc := range pairingDataTestCases {
		testModule(t, tc, false, false, true, false)
	}
}

var membershipTestCases = []pairingDataTestCase{
	{
		InputFName:       "testdata/ecpair_g2_both_cases_input.csv",
		ModuleFName:      "testdata/ecpair_g2_both_cases_module.csv",
		NbSubgroupChecks: 3,
	},
	{
		// test case with bigger limits than inputs. Tests that input fillers work correctly.
		InputFName:       "testdata/ecpair_g2_both_cases_input.csv",
		ModuleFName:      "testdata/ecpair_g2_both_cases_module.csv",
		NbSubgroupChecks: 5,
	},
	{
		// empty input to test edge case and fillers
		InputFName:       "testdata/ecpair_empty.csv",
		ModuleFName:      "",
		NbSubgroupChecks: 2,
	},
}

func TestMembership(t *testing.T) {
	for _, tc := range membershipTestCases {
		testModule(t, tc, false, false, false, true)
	}
}
