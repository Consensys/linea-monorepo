package experiment

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/logdata"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/sirupsen/logrus"
)

// TestDistributeWizard attempts to run and compile the distributed protocol.
func TestDistributedWizard(t *testing.T) {

	var (
		rng              = rand.New(utils.NewRandSource(0))
		sharedRandomness = field.PseudoRand(rng)
		zkevm            = GetZkEVM()
		disc             = &StandardModuleDiscoverer{
			TargetWeight: 1 << 28,
		}

		// This tests the compilation of the compiled-IOP
		distWizard = DistributeWizard(zkevm.WizardIOP, disc)
	)

	// This applies the dummy.Compiler to all parts of the distributed wizard.
	for i := range distWizard.GLs {
		dummy.CompileAtProverLvl(distWizard.GLs[i].Wiop)
		dummy.CompileAtProverLvl(distWizard.LPPs[i].Wiop)
	}

	var (
		reqFile      = files.MustRead("/home/ubuntu/beta-v2.1-rc1-trace/16303874-16303874-etv0.2.0-stv2.2.2-getZkProof.json")
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
		prevHornerN1Hash    = field.Element{}
		prevGlobalSent      = field.Element{}
	)

	witnessGLs, witnessLPPs := SegmentRuntime(runtimeBoot, &distWizard)

	for i := range witnessGLs {

		var (
			witnessGL   = witnessGLs[i]
			witnessLPP  = witnessLPPs[i]
			moduleIndex = witnessGLs[i].ModuleIndex
			moduleName  = witnessGLs[i].ModuleName
		)

		witnessLPP.InitialFiatShamirState = sharedRandomness

		t.Logf("segment(total)=%v module=%v segment.index=%v", i, witnessGL.ModuleName, witnessGL.ModuleIndex)

		if witnessLPP.ModuleName != moduleName || witnessLPP.ModuleIndex != moduleIndex {
			t.Fatalf(
				"GL and LPP are not aligned LPP[name=%v index=%v] GL[name=%v index=%v]",
				witnessLPPs[i].ModuleName, witnessLPPs[i].ModuleIndex,
				witnessGLs[i].ModuleName, witnessGLs[i].ModuleIndex,
			)
		}

		var (
			moduleGL  *ModuleGL
			moduleLPP *ModuleLPP
		)

		for k := range distWizard.ModuleNames {

			if distWizard.ModuleNames[k] != witnessGLs[i].ModuleName {
				continue
			}

			moduleGL = distWizard.GLs[k]
			moduleLPP = distWizard.LPPs[k]
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
			proverRunLPP         = wizard.RunProver(moduleLPP.Wiop, moduleLPP.GetMainProverStep(witnessLPPs[i]))
			proofLPP             = proverRunLPP.ExtractProof()
			verLPPRun, verLPPErr = wizard.VerifyWithRuntime(moduleLPP.Wiop, proofLPP)
		)

		if verLPPErr != nil {
			t.Errorf("verifier failed for segment %v, reason=%v", i, verLPPErr)
		}

		var (
			errMsg           = fmt.Sprintf("segment=%v, moduleName=%v, segment-index=%v", i, moduleName, moduleIndex)
			globalReceived   = verGLRun.GetPublicInput(globalReceiverPublicInput)
			globalSent       = verGLRun.GetPublicInput(globalSenderPublicInput)
			isFirst          = verGLRun.GetPublicInput(isFirstPublicInput)
			isLast           = verGLRun.GetPublicInput(isLastPublicInput)
			logDerivativeSum = verLPPRun.GetPublicInput(logDerivativeSumPublicInput)
			grandProduct     = verLPPRun.GetPublicInput(grandProductPublicInput)
			hornerSum        = verLPPRun.GetPublicInput(hornerPublicInput)
			hornerN0Hash     = verLPPRun.GetPublicInput(hornerN0HashPublicInput)
			hornerN1Hash     = verLPPRun.GetPublicInput(hornerN1HashPublicInput)
			shouldBeFirst    = i == 0 || witnessGLs[i].ModuleName != witnessGLs[i-1].ModuleName
			shouldBeLast     = i == len(witnessGLs)-1 || witnessGLs[i].ModuleName != witnessGLs[i+1].ModuleName
			colLPPFirstRound = proverRunLPP.Spec.Columns.AllKeysCommittedAt(0)
			colGLFirstRound  = proverRunGL.Spec.Columns.AllKeysCommittedAt(0)
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

		if !shouldBeFirst && hornerN0Hash != prevHornerN1Hash {
			t.Error("horner-n0-hash mismatch: " + errMsg)
		}

		if reflect.DeepEqual(colGLFirstRound, colLPPFirstRound) {
			t.Errorf("Module=%v Segment=%v: the LPP columns and the GL columns are not the same at round 0.colGLFirstRound=%v, colLPPFirstRound=%v", moduleName, i, colGLFirstRound, colLPPFirstRound)
		}

		prevGlobalSent = globalSent
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

	logrus.SetLevel(logrus.FatalLevel)

	go func() {
		ticker := time.Tick(time.Second * 10)
		for range ticker {
			memstats := &runtime.MemStats{}
			runtime.ReadMemStats(memstats)
			fmt.Printf("[memstats] heap-in-use=%vGiB stack-in-use=%vGiB\n", memstats.HeapInuse/(1<<30), memstats.StackInuse/(1<<30))
		}
	}()

	logrus.SetLevel(logrus.FatalLevel)

	var (
		rng              = rand.New(utils.NewRandSource(0))
		sharedRandomness = field.PseudoRand(rng)
		zkevm            = GetZkEVM()
		disc             = &StandardModuleDiscoverer{
			TargetWeight: 1 << 28,
		}

		// This tests the compilation of the compiled-IOP
		distWizard = DistributeWizard(zkevm.WizardIOP, disc)

		// Minimal witness size to compile
		minCompilationSize = 1 << 10

		compiledGLs  = make([]RecursedSegmentCompilation, len(distWizard.GLs))
		compiledLPPs = make([]RecursedSegmentCompilation, len(distWizard.LPPs))
	)

	// This applies the dummy.Compiler to all parts of the distributed wizard.
	for i := range distWizard.GLs {

		if cells := logdata.CountCells(distWizard.GLs[i].Wiop); cells.TotalCells() > minCompilationSize {
			fmt.Printf("[%v] Starting to compile module GL for %v\n", time.Now(), distWizard.ModuleNames[i])
			compiledGLs[i] = CompileSegmentGL(distWizard.GLs[i])
			fmt.Printf("[%v] Done compiling module GL for %v\n", time.Now(), distWizard.ModuleNames[i])
		}

		if cells := logdata.CountCells(distWizard.LPPs[i].Wiop); cells.TotalCells() > minCompilationSize {
			fmt.Printf("[%v] Starting to compile module LPP for %v\n", time.Now(), distWizard.ModuleNames[i])
			compiledLPPs[i] = CompileSegmentLPP(distWizard.LPPs[i])
			fmt.Printf("[%v] Done compiling module LPP for %v\n", time.Now(), distWizard.ModuleNames[i])
		}
	}

	var (
		reqFile      = files.MustRead("/home/ubuntu/beta-v2.1-rc1-trace/16303874-16303874-etv0.2.0-stv2.2.2-getZkProof.json")
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
		proof       = runtimeBoot.ExtractProof()
		verBootErr  = wizard.Verify(distWizard.Bootstrapper, proof)
	)

	t.Logf("[%v] done running the bootstrapper\n", time.Now())

	if verBootErr != nil {
		t.Fatalf("")
	}

	witnessGLs, witnessLPPs := SegmentRuntime(runtimeBoot, &distWizard)

	for i := range witnessGLs {

		var (
			witnessGL   = witnessGLs[i]
			witnessLPP  = witnessLPPs[i]
			moduleIndex = witnessGLs[i].ModuleIndex
			moduleName  = witnessGLs[i].ModuleName
		)

		witnessLPP.InitialFiatShamirState = sharedRandomness

		t.Logf("segment(total)=%v module=%v segment.index=%v", i, witnessGL.ModuleName, witnessGL.ModuleIndex)

		if witnessLPP.ModuleName != moduleName || witnessLPP.ModuleIndex != moduleIndex {
			t.Fatalf(
				"GL and LPP are not aligned LPP[name=%v index=%v] GL[name=%v index=%v]",
				witnessLPPs[i].ModuleName, witnessLPPs[i].ModuleIndex,
				witnessGLs[i].ModuleName, witnessGLs[i].ModuleIndex,
			)
		}

		var (
			moduleGL  *RecursedSegmentCompilation
			moduleLPP *RecursedSegmentCompilation
		)

		for k := range distWizard.ModuleNames {

			if distWizard.ModuleNames[k] != witnessGLs[i].ModuleName {
				continue
			}

			moduleGL = &compiledGLs[k]
			moduleLPP = &compiledLPPs[k]
		}

		if moduleGL == nil {
			t.Fatalf("module does not exists")
		}

		t.Logf("RUNNING THE GL PROVER: %v", time.Now())

		_ = moduleGL.ProveSegment(witnessGL)

		t.Logf("RUNNING THE GL PROVER - DONE: %v", time.Now())

		t.Logf("RUNNING THE LPP PROVER: %v", time.Now())

		_ = moduleLPP.ProveSegment(witnessLPP)

		t.Logf("RUNNING THE LPP PROVER - DONE: %v", time.Now())
	}
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
		Add:                                  1 << 19,
		Bin:                                  1 << 18,
		Blake2Fmodexpdata:                    1 << 14,
		Blockdata:                            1 << 12,
		Blockhash:                            1 << 12,
		Ecdata:                               1 << 18,
		Euc:                                  1 << 16,
		Exp:                                  1 << 14,
		Ext:                                  1 << 20,
		Gas:                                  1 << 16,
		Hub:                                  1 << 21,
		Logdata:                              1 << 16,
		Loginfo:                              1 << 12,
		Mmio:                                 1 << 21,
		Mmu:                                  1 << 21,
		Mod:                                  1 << 17,
		Mul:                                  1 << 16,
		Mxp:                                  1 << 19,
		Oob:                                  1 << 18,
		Rlpaddr:                              1 << 12,
		Rlptxn:                               1 << 17,
		Rlptxrcpt:                            1 << 17,
		Rom:                                  1 << 22,
		Romlex:                               1 << 12,
		Shakiradata:                          1 << 15,
		Shf:                                  1 << 16,
		Stp:                                  1 << 14,
		Trm:                                  1 << 15,
		Txndata:                              1 << 14,
		Wcp:                                  1 << 18,
		Binreftable:                          1 << 20,
		Shfreftable:                          4096,
		Instdecoder:                          512,
		PrecompileEcrecoverEffectiveCalls:    500,
		PrecompileSha2Blocks:                 600,
		PrecompileRipemdBlocks:               0,
		PrecompileModexpEffectiveCalls:       64,
		PrecompileEcaddEffectiveCalls:        1 << 14,
		PrecompileEcmulEffectiveCalls:        32,
		PrecompileEcpairingEffectiveCalls:    32,
		PrecompileEcpairingMillerLoops:       64,
		PrecompileEcpairingG2MembershipCalls: 64,
		PrecompileBlakeEffectiveCalls:        0,
		PrecompileBlakeRounds:                0,
		BlockKeccak:                          1 << 13,
		BlockL1Size:                          100_000,
		BlockL2L1Logs:                        16,
		BlockTransactions:                    400,
		ShomeiMerkleProofs:                   1 << 14,
	}

	return zkevm.FullZKEVMWithSuite(&traceLimits, []func(*wizard.CompiledIOP){})
}
