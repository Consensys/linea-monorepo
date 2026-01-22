package distributed_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
	"github.com/consensys/linea-monorepo/prover/zkevm"
)

// Global cache for compiled distributed wizard (in-memory cache for test session)
var (
	cachedDistWizardAfterCompile *distributed.DistributedWizard
	cacheCompileMutex            sync.Mutex
	cacheKey                     string
)

// TestConglomerationBasicCached is a cached version of TestConglomerationBasic
// It caches intermediate compilation results to speed up iterative testing
func TestConglomerationBasicCached(t *testing.T) {
	var (
		numRow = 1 << 5
		tc     = LookupTestCase{numRow: numRow}
		disc   = &distributed.StandardModuleDiscoverer{
			TargetWeight: 3 * numRow / 2,
			Predivision:  1,
		}

		// Custom compilation params for this test
		testCompilationParams = distributed.CompilationParams{
			FixedNbRowPlonkCircuit:       1 << 26,
			FixedNbRowExternalHasher:     1 << 23,
			FixedNbPublicInput:           1 << 10,
			InitialCompilerSize:          1 << 18,
			InitialCompilerSizeConglo:    1 << 18,
			ColumnProfileMPTS:            nil,
			ColumnProfileMPTSPrecomputed: 0,
		}
	)

	// Cache key based on test parameters
	currentCacheKey := fmt.Sprintf("basic_numrow_%d_target_%d", numRow, disc.TargetWeight)

	// Force clear cache if requested
	if os.Getenv("LINEA_CLEAR_CACHE") != "" {
		t.Logf("[CACHE] Clearing cache as requested...")
		cacheCompileMutex.Lock()
		cachedDistWizardAfterCompile = nil
		cacheKey = ""
		cacheCompileMutex.Unlock()
	}

	// Stage 1: Compile wizard
	t.Logf("[%v] Stage 1: Compiling wizard...", time.Now())
	comp := wizard.Compile(func(build *wizard.Builder) {
		tc.Define(build.CompiledIOP)
	})

	// Stage 2: Distribute and compile segments (EXPENSIVE - cache this!)
	t.Logf("[%v] Stage 2: Distributing wizard and compiling segments...", time.Now())

	var distWizardAfterCompile *distributed.DistributedWizard

	// Check if we have it cached in memory
	cacheCompileMutex.Lock()
	if cachedDistWizardAfterCompile != nil && cacheKey == currentCacheKey {
		t.Logf("[CACHE HIT] Using cached compiled wizard from previous run!")
		distWizardAfterCompile = cachedDistWizardAfterCompile
		cacheCompileMutex.Unlock()
	} else {
		cacheCompileMutex.Unlock()

		t.Logf("[CACHE MISS] Compiling from scratch - this will take ~6-7 minutes...")
		startCompile := time.Now()

		// This is the expensive part - CompileSegments
		distWizardAfterCompile = distributed.DistributeWizard(comp, disc).
			CompileSegments(testCompilationParams)

		compileTime := time.Since(startCompile)
		t.Logf("[CACHE] Compilation took %v, caching for next run...", compileTime)

		// Cache it in memory for next test run
		cacheCompileMutex.Lock()
		cachedDistWizardAfterCompile = distWizardAfterCompile
		cacheKey = currentCacheKey
		cacheCompileMutex.Unlock()
	}

	// Stage 3: Conglomerate (also expensive but done after cache point)
	t.Logf("[%v] Stage 3: Running conglomerate...", time.Now())
	startConglo := time.Now()
	distWizard := distWizardAfterCompile.Conglomerate(testCompilationParams)
	t.Logf("[%v] Conglomerate took %v", time.Now(), time.Since(startConglo))

	t.Logf("[%v] Stage 3: Running bootstrapper...", time.Now())
	runtimeBoot := wizard.RunProver(distWizard.Bootstrapper, tc.Assign, false)

	t.Logf("[%v] Stage 4: Segmenting runtime...", time.Now())
	witnessGLs, witnessLPPs := distributed.SegmentRuntime(
		runtimeBoot,
		distWizard.Disc,
		distWizard.BlueprintGLs,
		distWizard.BlueprintLPPs,
		distWizard.VerificationKeyMerkleTree.GetRoot(),
	)

	t.Logf("[%v] Stage 5: Running GL provers...", time.Now())
	glProofs := runProverGLs(t, distWizard, witnessGLs)

	t.Logf("[%v] Stage 6: Getting shared randomness...", time.Now())
	sharedRandomness := distributed.GetSharedRandomnessFromSegmentProofs(glProofs)

	t.Logf("[%v] Stage 7: Running LPP provers...", time.Now())
	runLPPs := runProverLPPs(t, distWizard, sharedRandomness, witnessLPPs)

	t.Logf("[%v] Stage 8: Running conglomeration prover...", time.Now())
	runConglomerationProver(
		&distWizard.VerificationKeyMerkleTree,
		distWizard.CompiledConglomeration,
		glProofs,
		runLPPs,
	)

	t.Logf("[%v] Test completed successfully!", time.Now())
}

// TestConglomerationProverFileCached is a cached version for file-based testing
func TestConglomerationProverFileCached(t *testing.T) {
	t.Skipf("the test is a development/debug/integration test. It is not needed for CI")

	var (
		reqFile      = files.MustRead("/home/ubuntu/mainnet-beta-v2-5.1.3/prover-execution/requests/20106872-20106937-etv0.2.0-stv2.3.0-getZkProof.json")
		cfgFilePath  = "/home/ubuntu/zkevm-monorepo/prover/config/config-mainnet-limitless.toml"
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

	// Initialize cache
	cacheDir := os.Getenv("LINEA_TEST_CACHE_DIR")
	if cacheDir == "" {
		cacheDir = "./debug/test-cache"
	}
	cache := NewTestCache(cacheDir)

	if os.Getenv("LINEA_CLEAR_CACHE") != "" {
		t.Logf("Clearing test cache as requested...")
		cache.ClearAll()
	}

	// Generate cache key based on config file hash
	cfgHash := hashFile(cfgFilePath)
	cacheKey := fmt.Sprintf("limitless_zkevm_%s", cfgHash[:8])

	t.Logf("loaded config: %++v", cfg)

	// Try to load zkEVM from cache (if possible)
	t.Logf("[%v] Initializing zkEVM...\n", time.Now())
	var limitlessZkEVM *zkevm.LimitlessZkEVM

	// Note: LimitlessZkEVM is likely not directly cacheable
	// You would need to cache the compilation outputs separately
	limitlessZkEVM = zkevm.NewLimitlessZkEVM(cfg)
	z := limitlessZkEVM.Zkevm
	distWizard := limitlessZkEVM.DistWizard

	t.Logf("[%v] running the bootstrapper\n", time.Now())
	var (
		_, witness  = test_utils.GetZkevmWitness(req, cfg)
		runtimeBoot = wizard.RunProver(distWizard.Bootstrapper, z.GetMainProverStep(witness), false)
	)

	t.Logf("[%v] done running the bootstrapper\n", time.Now())

	var (
		witnessGLs, witnessLPPs = distributed.SegmentRuntime(
			runtimeBoot,
			distWizard.Disc,
			distWizard.BlueprintGLs,
			distWizard.BlueprintLPPs,
			distWizard.VerificationKeyMerkleTree.GetRoot(),
		)
		glProofs         = runProverGLs(t, distWizard, witnessGLs)
		sharedRandomness = distributed.GetSharedRandomnessFromSegmentProofs(glProofs)
		runLPPs          = runProverLPPs(t, distWizard, sharedRandomness, witnessLPPs)
	)

	runConglomerationProver(
		&distWizard.VerificationKeyMerkleTree,
		distWizard.CompiledConglomeration,
		glProofs,
		runLPPs,
	)

	t.Logf("[CACHE INFO] For caching key used: %s", cacheKey)
}

// hashFile generates a SHA256 hash of a file
func hashFile(filepath string) string {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "unknown"
	}
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}
