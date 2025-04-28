package distributed

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"reflect"
	"testing"
	"time"

	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/logdata"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm"
)

// TestDistributedWizard attempts to compiler the wizard distribution.
func TestDistributedWizard(t *testing.T) {

	var (
		zkevm      = GetZkEVM()
		affinities = GetAffinities(zkevm)
		discoverer = &StandardModuleDiscoverer{
			TargetWeight: 1 << 28,
			Affinities:   affinities,
			Predivision:  16,
		}
	)

	// This dumps a CSV with the list of all the columns in the original wizard
	logdata.GenCSV(
		files.MustOverwrite("./base-data/main.csv"),
		logdata.IncludeColumnCSVFilter,
	)(zkevm.WizardIOP)

	distributed := DistributeWizard(zkevm.WizardIOP, discoverer)

	for i, modGL := range distributed.GLs {

		logdata.GenCSV(
			files.MustOverwrite(fmt.Sprintf("./module-data/module-%v-gl.csv", i)),
			logdata.IncludeColumnCSVFilter,
		)(modGL.Wiop)
	}

	for i, modLPP := range distributed.LPPs {

		logdata.GenCSV(
			files.MustOverwrite(fmt.Sprintf("./module-data/module-%v-lpp.csv", i)),
			logdata.IncludeColumnCSVFilter,
		)(modLPP.Wiop)
	}
}

// TestDistributeWizard attempts to run and compile the distributed protocol in
// dummy mode. Meaning without actual compilation.
func TestDistributedWizardLogic(t *testing.T) {

	t.Skipf("the test is a development/debug/integration test. It is not needed for CI")

	var (
		// #nosec G404 --we don't need a cryptographic RNG for testing purpose
		// rng              = rand.New(utils.NewRandSource(0))
		// sharedRandomness = field.PseudoRand(rng)
		zkevm = GetZkEVM()
		disc  = &StandardModuleDiscoverer{
			TargetWeight: 1 << 28,
			Affinities:   GetAffinities(zkevm),
			Predivision:  16,
		}

		// This tests the compilation of the compiled-IOP
		distWizard = DistributeWizard(zkevm.WizardIOP, disc)
	)

	// This applies the dummy.Compiler to all parts of the distributed wizard.
	for i := range distWizard.GLs {
		dummy.CompileAtProverLvl()(distWizard.GLs[i].Wiop)
	}

	for i := range distWizard.LPPs {
		dummy.CompileAtProverLvl()(distWizard.LPPs[i].Wiop)
	}

	var (
		reqFile      = files.MustRead("/home/ubuntu/beta-v2-rc11/10556002-10556002-etv0.2.0-stv2.2.2-getZkProof.json")
		cfgFilePath  = "/home/ubuntu/zkevm-monorepo/prover/config/config-sepolia-full.toml"
		req          = &execution.Request{}
		reqDecodeErr = json.NewDecoder(reqFile).Decode(req)
		cfg, cfgErr  = config.NewConfigFromFileUnchecked(cfgFilePath)
	)

	if reqDecodeErr != nil {
		t.Fatalf("could not read the request file: %v", reqDecodeErr)
	}

	if cfgErr != nil {
		t.Fatalf("could not read the config file: err=%v, cfg=%++v", cfgErr, cfg)
	}

	t.Logf("loaded config: %++v", cfg)

	t.Logf("Checking the initial bootstrapper - wizard")
	var (
		witness     = GetZkevmWitness(req, cfg)
		runtimeBoot = wizard.RunProver(distWizard.Bootstrapper, zkevm.GetMainProverStep(witness))
		proof       = runtimeBoot.ExtractProof()
		verBootErr  = wizard.Verify(distWizard.Bootstrapper, proof)
	)
	t.Logf("Checking the initial bootstrapper - wizard")

	if verBootErr != nil {
		t.Fatalf("Bootstrapper failed because: %v", verBootErr)
	}

	var (
		allGrandProduct     = field.NewElement(1)
		allLogDerivativeSum = field.Element{}
		allHornerSum        = field.Element{}
		prevGlobalSent      = field.Element{}
		prevHornerN1Hash    = field.Element{}
	)

	witnessGLs, witnessLPPs := SegmentRuntime(runtimeBoot, distWizard)

	for i := range witnessGLs {

		var (
			witnessGL   = witnessGLs[i]
			moduleIndex = witnessGLs[i].ModuleIndex
			moduleName  = witnessGLs[i].ModuleName
			moduleGL    *ModuleGL
		)

		t.Logf("segment(total)=%v module=%v segment.index=%v", i, witnessGL.ModuleName, witnessGL.ModuleIndex)

		for k := range distWizard.ModuleNames {
			if distWizard.ModuleNames[k] != witnessGLs[i].ModuleName {
				continue
			}

			moduleGL = distWizard.GLs[k]
			break
		}

		if moduleGL == nil {
			t.Fatalf("module does not exists")
		}

		var (
			proverRunGL        = wizard.RunProver(moduleGL.Wiop, moduleGL.GetMainProverStep(witnessGLs[i]))
			proofGL            = proverRunGL.ExtractProof()
			verGLRun, verGLErr = wizard.VerifyWithRuntime(moduleGL.Wiop, proofGL)
		)

		if verGLErr != nil {
			t.Errorf("verifier failed for segment %v, reason=%v", i, verGLErr)
		}

		var (
			errMsg         = fmt.Sprintf("segment=%v, moduleName=%v, segment-index=%v", i, moduleName, moduleIndex)
			globalReceived = verGLRun.GetPublicInput(globalReceiverPublicInput)
			globalSent     = verGLRun.GetPublicInput(globalSenderPublicInput)
			isFirst        = verGLRun.GetPublicInput(isFirstPublicInput)
			isLast         = verGLRun.GetPublicInput(isLastPublicInput)
			shouldBeFirst  = i == 0 || witnessGLs[i].ModuleName != witnessGLs[i-1].ModuleName
			shouldBeLast   = i == len(witnessGLs)-1 || witnessGLs[i].ModuleName != witnessGLs[i+1].ModuleName
		)

		if isFirst.IsOne() != shouldBeFirst {
			t.Error("isFirst has unexpected values: " + errMsg)
		}

		if isLast.IsOne() != shouldBeLast {
			t.Error("isLast has unexpected values: " + errMsg)
		}

		if !shouldBeFirst && globalReceived != prevGlobalSent {
			t.Error("global-received does not match: " + errMsg)
		}

		prevGlobalSent = globalSent
	}

	for i := range witnessLPPs {

		var (
			witnessLPP  = witnessLPPs[i]
			moduleIndex = witnessLPPs[i].ModuleIndex
			moduleNames = witnessLPPs[i].ModuleNames
			moduleLPP   *ModuleLPP
		)

		for k := range distWizard.LPPs {
			if !reflect.DeepEqual(distWizard.LPPs[k].ModuleNames(), moduleNames) {
				continue
			}

			moduleLPP = distWizard.LPPs[k]
			break
		}

		if moduleLPP == nil {
			t.Fatalf("module does not exists")
		}

		var (
			proverRunLPP         = wizard.RunProver(moduleLPP.Wiop, moduleLPP.GetMainProverStep(witnessLPP))
			proofLPP             = proverRunLPP.ExtractProof()
			verLPPRun, verLPPErr = wizard.VerifyWithRuntime(moduleLPP.Wiop, proofLPP)
		)

		if verLPPErr != nil {
			t.Errorf("verifier failed for segment %v, reason=%v", i, verLPPErr)
		}

		var (
			errMsg           = fmt.Sprintf("segment=%v, moduleName=%v, segment-index=%v", i, moduleNames, moduleIndex)
			logDerivativeSum = verLPPRun.GetPublicInput(logDerivativeSumPublicInput)
			grandProduct     = verLPPRun.GetPublicInput(grandProductPublicInput)
			hornerSum        = verLPPRun.GetPublicInput(hornerPublicInput)
			hornerN0Hash     = verLPPRun.GetPublicInput(hornerN0HashPublicInput)
			hornerN1Hash     = verLPPRun.GetPublicInput(hornerN1HashPublicInput)
			shouldBeFirst    = i == 0 || !reflect.DeepEqual(witnessLPPs[i].ModuleNames, witnessLPPs[i-1].ModuleNames)
		)

		if !shouldBeFirst && hornerN0Hash != prevHornerN1Hash {
			t.Error("horner-n0-hash mismatch: " + errMsg)
		}

		prevHornerN1Hash = hornerN1Hash
		allGrandProduct.Mul(&allGrandProduct, &grandProduct)
		allHornerSum.Add(&allHornerSum, &hornerSum)
		allLogDerivativeSum.Add(&allLogDerivativeSum, &logDerivativeSum)
	}

	if !allGrandProduct.IsOne() {
		t.Errorf("grand-product does not cancel")
	}

	if !allHornerSum.IsZero() {
		t.Errorf("horner does not cancel")
	}

	if !allLogDerivativeSum.IsZero() {
		t.Errorf("log-derivative-sum does not cancel. Has %v", allLogDerivativeSum.String())
	}
}

// TestBenchDistributedWizard runs the distributed wizard will all the compilations
func TestBenchDistributedWizard(t *testing.T) {

	t.Skipf("the test is a development/debug/integration test. It is not needed for CI")

	var (
		zkevm = GetZkEVM()
		disc  = &StandardModuleDiscoverer{
			TargetWeight: 1 << 28,
			Affinities:   GetAffinities(zkevm),
			Predivision:  16,
		}

		// This tests the compilation of the compiled-IOP
		distWizard = DistributeWizard(zkevm.WizardIOP, disc).CompileSegments()
	)

	var (
		reqFile      = files.MustRead("/home/ubuntu/beta-v2-rc11/10556002-10556002-etv0.2.0-stv2.2.2-getZkProof.json")
		cfgFilePath  = "/home/ubuntu/zkevm-monorepo/prover/config/config-sepolia-full.toml"
		req          = &execution.Request{}
		reqDecodeErr = json.NewDecoder(reqFile).Decode(req)
		cfg, cfgErr  = config.NewConfigFromFileUnchecked(cfgFilePath)
	)

	if reqDecodeErr != nil {
		t.Fatalf("could not read the request file: %v", reqDecodeErr)
	}

	if cfgErr != nil {
		t.Fatalf("could not read the config file: err=%v, cfg=%++v", cfgErr, cfg)
	}

	t.Logf("loaded config: %++v", cfg)

	t.Logf("[%v] running the bootstrapper\n", time.Now())

	var (
		witness     = GetZkevmWitness(req, cfg)
		runtimeBoot = wizard.RunProver(distWizard.Bootstrapper, zkevm.GetMainProverStep(witness))
	)

	t.Logf("[%v] done running the bootstrapper\n", time.Now())

	witnessGLs, witnessLPPs := SegmentRuntime(runtimeBoot, distWizard)
	runProverGLs(t, distWizard, witnessGLs)
	runProverLPPs(t, distWizard, witnessLPPs)
}

// GetZkevmWitness returns a [zkevm.Witness]
func GetZkevmWitness(req *execution.Request, cfg *config.Config) *zkevm.Witness {
	out := execution.CraftProverOutput(cfg, req)
	witness := execution.NewWitness(cfg, req, &out)
	return witness.ZkEVM
}

// GetZKEVM returns a [zkevm.ZkEvm] with its trace limits inflated so that it
// can be used as input for the package functions. The zkevm is returned
// without any compilation.
func GetZkEVM() *zkevm.ZkEvm {

	// This are the config trace-limits from sepolia. All multiplied by 16.
	traceLimits := config.TracesLimits{
		Add:                                  1 << 23,
		Bin:                                  1 << 22,
		Blake2Fmodexpdata:                    1 << 18,
		Blockdata:                            1 << 16,
		Blockhash:                            1 << 16,
		Ecdata:                               1 << 22,
		Euc:                                  1 << 20,
		Exp:                                  1 << 18,
		Ext:                                  1 << 24,
		Gas:                                  1 << 20,
		Hub:                                  1 << 25,
		Logdata:                              1 << 20,
		Loginfo:                              1 << 16,
		Mmio:                                 1 << 25,
		Mmu:                                  1 << 25,
		Mod:                                  1 << 21,
		Mul:                                  1 << 20,
		Mxp:                                  1 << 23,
		Oob:                                  1 << 22,
		Rlpaddr:                              1 << 16,
		Rlptxn:                               1 << 21,
		Rlptxrcpt:                            1 << 21,
		Rom:                                  1 << 26,
		Romlex:                               1 << 16,
		Shakiradata:                          1 << 19,
		Shf:                                  1 << 20,
		Stp:                                  1 << 18,
		Trm:                                  1 << 19,
		Txndata:                              1 << 18,
		Wcp:                                  1 << 22,
		Binreftable:                          1 << 24,
		Shfreftable:                          1 << 16,
		Instdecoder:                          1 << 13,
		PrecompileEcrecoverEffectiveCalls:    1 << 13,
		PrecompileSha2Blocks:                 1 << 13,
		PrecompileRipemdBlocks:               0,
		PrecompileModexpEffectiveCalls:       1 << 10,
		PrecompileModexpEffectiveCalls4096:   1 << 4,
		PrecompileEcaddEffectiveCalls:        1 << 12,
		PrecompileEcmulEffectiveCalls:        1 << 9,
		PrecompileEcpairingEffectiveCalls:    1 << 9,
		PrecompileEcpairingMillerLoops:       1 << 10,
		PrecompileEcpairingG2MembershipCalls: 1 << 10,
		PrecompileBlakeEffectiveCalls:        0,
		PrecompileBlakeRounds:                0,
		BlockKeccak:                          1 << 17,
		BlockL1Size:                          100_000,
		BlockL2L1Logs:                        16,
		BlockTransactions:                    1 << 12,
		ShomeiMerkleProofs:                   1 << 18,
	}

	return zkevm.FullZKEVMWithSuite(&traceLimits, zkevm.CompilationSuite{}, &config.Config{})
}

// GetAffinities returns a list of affinities for the following modules. This
// affinities regroup how the modules are grouped.
//
//	ecadd / ecmul / ecpairing
//	hub / hub.scp / hub.acp
//	everything related to keccak
func GetAffinities(z *zkevm.ZkEvm) [][]column.Natural {

	return [][]column.Natural{
		{
			z.Ecmul.AlignedGnarkData.IsActive.(column.Natural),
			z.Ecadd.AlignedGnarkData.IsActive.(column.Natural),
			z.Ecpair.AlignedFinalExpCircuit.IsActive.(column.Natural),
			z.Ecpair.AlignedG2MembershipData.IsActive.(column.Natural),
			z.Ecpair.AlignedMillerLoopCircuit.IsActive.(column.Natural),
		},
		{
			z.WizardIOP.Columns.GetHandle("hub.HUB_STAMP").(column.Natural),
			z.WizardIOP.Columns.GetHandle("hub.scp_ADDRESS_HI").(column.Natural),
			z.WizardIOP.Columns.GetHandle("hub.acp_ADDRESS_HI").(column.Natural),
			z.WizardIOP.Columns.GetHandle("hub.ccp_HUB_STAMP").(column.Natural),
			z.WizardIOP.Columns.GetHandle("hub.envcp_HUB_STAMP").(column.Natural),
			z.WizardIOP.Columns.GetHandle("hub.stkcp_PEEK_AT_STACK_POW_4").(column.Natural),
		},
		{
			z.WizardIOP.Columns.GetHandle("KECCAK_IMPORT_PAD_HASH_NUM").(column.Natural),
			z.WizardIOP.Columns.GetHandle("CLEANING_KECCAK_CleanLimb").(column.Natural),
			z.WizardIOP.Columns.GetHandle("DECOMPOSITION_KECCAK_Decomposed_Len_0").(column.Natural),
			z.WizardIOP.Columns.GetHandle("KECCAK_FILTERS_SPAGHETTI").(column.Natural),
			z.WizardIOP.Columns.GetHandle("LANE_KECCAK_Lane").(column.Natural),
			z.WizardIOP.Columns.GetHandle("KECCAKF_IS_ACTIVE_").(column.Natural),
			z.WizardIOP.Columns.GetHandle("KECCAKF_BLOCK_BASE_2_0").(column.Natural),
			z.WizardIOP.Columns.GetHandle("KECCAK_OVER_BLOCKS_TAGS_0").(column.Natural),
			z.WizardIOP.Columns.GetHandle("HASH_OUTPUT_Hash_Lo").(column.Natural),
		},
		{
			z.WizardIOP.Columns.GetHandle("SHA2_IMPORT_PAD_HASH_NUM").(column.Natural),
			z.WizardIOP.Columns.GetHandle("DECOMPOSITION_SHA2_Decomposed_Len_0").(column.Natural),
			z.WizardIOP.Columns.GetHandle("LENGTH_CONSISTENCY_SHA2_BYTE_LEN_0_0").(column.Natural),
			z.WizardIOP.Columns.GetHandle("SHA2_FILTERS_SPAGHETTI").(column.Natural),
			z.WizardIOP.Columns.GetHandle("LANE_SHA2_Lane").(column.Natural),
			z.WizardIOP.Columns.GetHandle("Coefficient_SHA2").(column.Natural),
			z.WizardIOP.Columns.GetHandle("SHA2_OVER_BLOCK_IS_ACTIVE").(column.Natural),
			z.WizardIOP.Columns.GetHandle("SHA2_OVER_BLOCK_SHA2_COMPRESSION_CIRCUIT_IS_ACTIVE").(column.Natural),
		},
		{
			z.WizardIOP.Columns.GetHandle("mmio.CN_ABC").(column.Natural),
			z.WizardIOP.Columns.GetHandle("mmio.MMIO_STAMP").(column.Natural),
			z.WizardIOP.Columns.GetHandle("mmu.STAMP").(column.Natural),
		},
	}
}

// DistributeTestCase is an implementation of the testcase interface. The
// testcase generates 2 triplets of columns a, b, c such that a + b = c
// and the two modules are joined by a lookup.
type DistributeTestCase struct {
	numRow int
}

// Define defines the structure of the distributed wizard. The structure is
// composed of 2 modules that are connected by a lookup. The two modules are
// identical and are defined as a + b = c. The a, b and c are each defined as
// a commit in the wizard. The lookup is defined as a global constraint that
// enforces the equality of the two modules.
func (d DistributeTestCase) Define(comp *wizard.CompiledIOP) {

	// Define the first module
	a0 := comp.InsertCommit(0, "a0", d.numRow)
	b0 := comp.InsertCommit(0, "b0", d.numRow)
	c0 := comp.InsertCommit(0, "c0", d.numRow)

	a1 := comp.InsertCommit(0, "a1", d.numRow)
	b1 := comp.InsertCommit(0, "b1", d.numRow)
	c1 := comp.InsertCommit(0, "c1", d.numRow)

	comp.InsertGlobal(0, "global-0", symbolic.Sub(c0, b0, a0))
	comp.InsertGlobal(0, "global-1", symbolic.Sub(c1, b1, a1))

	comp.InsertInclusion(0, "inclusion-0", []ifaces.Column{c0, b0, a0}, []ifaces.Column{c1, b1, a1})
}

// Assign sets up the column assignments for the DistributeTestCase
// within the ProverRuntime. It assigns constant values to six columns
// ('a0', 'b0', 'c0', 'a1', 'b1', 'c1') where each column is assigned
// a smart vector with a constant field element value and a specified
// number of rows (d.numRow). This function helps initialize the columns
// with predetermined values for the testing setup.
func (d DistributeTestCase) Assign(run *wizard.ProverRuntime) {
	run.AssignColumn("a0", smartvectors.RightZeroPadded(vector.Repeat(field.NewElement(1), d.numRow-2), d.numRow))
	run.AssignColumn("b0", smartvectors.RightZeroPadded(vector.Repeat(field.NewElement(2), d.numRow-2), d.numRow))
	run.AssignColumn("c0", smartvectors.RightZeroPadded(vector.Repeat(field.NewElement(3), d.numRow-2), d.numRow))
	run.AssignColumn("a1", smartvectors.RightZeroPadded(vector.Repeat(field.NewElement(1), d.numRow-2), d.numRow))
	run.AssignColumn("b1", smartvectors.RightZeroPadded(vector.Repeat(field.NewElement(2), d.numRow-2), d.numRow))
	run.AssignColumn("c1", smartvectors.RightZeroPadded(vector.Repeat(field.NewElement(3), d.numRow-2), d.numRow))
}

// runProverGLs executes the prover for each GL module segment. It takes in a list of
// compiled GL segments and corresponding witnesses, then runs the prover for each
// segment. The function logs the start and end times of the prover execution for each
// segment. It returns a slice of ProverRuntime instances, each representing the
// result of the prover execution for a segment.
func runProverGLs(
	t *testing.T,
	distWizard *DistributedWizard,
	witnessGLs []*ModuleWitnessGL,
) []*wizard.ProverRuntime {

	var (
		compiledGLs = distWizard.CompiledGLs
		runs        = make([]*wizard.ProverRuntime, len(witnessGLs))
	)

	for i := range witnessGLs {

		var (
			witnessGL = witnessGLs[i]
			moduleGL  *RecursedSegmentCompilation
		)

		t.Logf("segment(total)=%v module=%v segment.index=%v", i, witnessGL.ModuleName, witnessGL.ModuleIndex)
		for k := range distWizard.ModuleNames {
			if distWizard.ModuleNames[k] != witnessGLs[i].ModuleName {
				continue
			}
			moduleGL = compiledGLs[k]
		}

		if moduleGL == nil {
			t.Fatalf("module does not exists")
		}

		t.Logf("RUNNING THE GL PROVER: %v", time.Now())
		runs[i] = moduleGL.ProveSegment(witnessGL)
		t.Logf("RUNNING THE GL PROVER - DONE: %v", time.Now())
	}

	return runs
}

// runProverLPPs runs a prover for a LPP segment. It takes in a DistributedWizard
// object, a slice of RecursedSegmentCompilation objects, and a slice of
// ModuleWitnessLPP objects. It runs the prover for each segment and logs the
// time at which the prover starts and ends. It returns a slice of ProverRuntime
// instances, each representing the result of the prover execution for a segment.
func runProverLPPs(
	t *testing.T,
	distWizard *DistributedWizard,
	witnessLPPs []*ModuleWitnessLPP,
) []*wizard.ProverRuntime {

	var (
		// #nosec G404 --we don't need a cryptographic RNG for testing purpose
		rng              = rand.New(utils.NewRandSource(0))
		sharedRandomness = field.PseudoRand(rng)
		runs             = make([]*wizard.ProverRuntime, len(witnessLPPs))
		compiledLPPs     = distWizard.CompiledLPPs
	)

	for i := range witnessLPPs {

		var (
			witnessLPP = witnessLPPs[i]
			moduleLPP  *RecursedSegmentCompilation
		)

		witnessLPP.InitialFiatShamirState = sharedRandomness

		t.Logf("segment(total)=%v module=%v segment.index=%v", i, witnessLPP.ModuleNames, witnessLPP.ModuleIndex)
		for k := range distWizard.LPPs {

			if !reflect.DeepEqual(distWizard.LPPs[k].ModuleNames(), witnessLPPs[i].ModuleNames) {
				continue
			}

			moduleLPP = compiledLPPs[k]
		}

		if moduleLPP == nil {
			t.Fatalf("module does not exists")
		}

		t.Logf("RUNNING THE LPP PROVER: %v", time.Now())
		runs[i] = moduleLPP.ProveSegment(witnessLPP)
		t.Logf("RUNNING THE LPP PROVER - DONE: %v", time.Now())
	}

	return runs
}
