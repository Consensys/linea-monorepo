package distributed

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"

	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/exit"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var (

	// numConcurrentWritingGoroutines governs the goroutine serializing,
	// compressing and writing the  witness. The writing part is also controlled
	// by a semaphore on top of this.
	numConcurrentWritingGoroutines = 12
)

type Metadata struct {
	BootstrapRequestDoneFile string `json:"bootstrapRequestDoneFile"`
	StartBlock               string `json:"startBlock"`
	EndBlock                 string `json:"endBlock"`
	NumGL                    int    `json:"numGL"`
	NumLPP                   int    `json:"numLPP"`

	GLProofFiles  []string `json:"glProofFiles"`
	GLCommitFiles []string `json:"glCommitFiles"`

	SharedRndFile string `json:"sharedRndFile"`

	LPPProofFiles []string `json:"lppProofFiles"`
}

func RunBootstrapper(cfg *config.Config, req *execution.Request, metadata *Metadata) (*Metadata, error) {

	// Set MonitorParams before any bootstrapping happens
	profiling.SetMonitorParams(cfg)

	// Recover wrapper for panics
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("[PANIC] Bootstrapper crashed for conflation request %s-%s:", metadata.StartBlock, metadata.EndBlock)
			debug.PrintStack()
			os.Exit(2)
		}
	}()

	// Setting the issue handler to exit on unsatisfied constraint but not limit overflow.
	exit.SetIssueHandlingMode(exit.ExitOnUnsatisfiedConstraint)

	// Setup execution witness and output response
	var (
		out     = execution.CraftProverOutput(cfg, req)
		witness = execution.NewWitness(cfg, req, &out)
	)

	if cfg.Execution.LimitlessWithDebug {
		limitlessZkEVM := zkevm.NewLimitlessDebugZkEVM(cfg)
		limitlessZkEVM.RunDebug(cfg, witness.ZkEVM)
		return nil, nil
	}

	logrus.Info("Starting to run the bootstrapper")

	err := initBootstrap(cfg, witness.ZkEVM, metadata)
	if err != nil {
		return nil, fmt.Errorf("error during bootstrap:%s", err.Error())
	}

	logrus.Infof("Bootstrapper finished successfully and generated %d GL modules and %d LPP modules", metadata.NumGL, metadata.NumLPP)
	return metadata, nil
}

func initBootstrap(cfg *config.Config, zkevmWitness *zkevm.Witness, metadata *Metadata) error {

	assets := &zkevm.LimitlessZkEVM{}
	loadStaticProverAssetsFromDisk(cfg, assets)

	// Load blueprints in background, we only need them after bootstrapping succeeds.
	eg := &errgroup.Group{}
	eg.Go(func() error {
		return assets.LoadBlueprints(cfg)
	})

	// The function initially attempt to run the bootstrapper directly and will
	// catch "limit-overflow" panic msgs. When they happen, we retry running
	// the bootstrapper with higher and higher limits until it works.
	var (
		scalingFactor = 1
		runtimeBoot   *wizard.ProverRuntime
	)

	for runtimeBoot == nil {

		logrus.Infof("Trying to bootstrap with a scaling of %v\n", scalingFactor)

		func() {

			// Since the [exit] package is configured to only send panic messages
			// on overflow. The overflows are catchable.
			defer func() {
				if err := recover(); err != nil {
					oFReport, isOF := err.(exit.LimitOverflowReport)
					if isOF {
						extra := utils.DivCeil(oFReport.RequestedSize, oFReport.Limit)
						scalingFactor *= utils.NextPowerOfTwo(extra)
						return
					}

					panic(err)
				}
			}()

			if scalingFactor == 1 {
				logrus.Infof("Running bootstrapper")
				runtimeBoot = wizard.RunProver(
					assets.DistWizard.Bootstrapper,
					assets.Zkevm.GetMainProverStep(zkevmWitness),
				)
				return
			}

			scaledUpBootstrapper, scaledUpZkEVM := zkevm.GetScaledUpBootstrapper(
				cfg, assets.DistWizard.Disc, scalingFactor,
			)

			runtimeBoot = wizard.RunProver(
				scaledUpBootstrapper,
				scaledUpZkEVM.GetMainProverStep(zkevmWitness),
			)
		}()
	}

	// Wait for blueprints to finish before freeing assets
	if err := eg.Wait(); err != nil {
		utils.Panic("could not load GL and LPP blueprint modules: %v", err)
	}

	// This frees the memory from the assets that are no longer needed. We don't
	// need to do that for the module GLs and LPPs because they are thrown away
	// when the current function returns.
	assets.Zkevm = nil
	assets.DistWizard.Bootstrapper = nil

	logrus.Infof("Segmenting the runtime")
	witnessGLs, witnessLPPs := distributed.SegmentRuntime(
		runtimeBoot,
		assets.DistWizard.Disc,
		assets.DistWizard.BlueprintGLs,
		assets.DistWizard.BlueprintLPPs,
	)

	// Populate the metadata fields
	metadata.NumGL = len(witnessGLs)
	metadata.NumLPP = len(witnessLPPs)
	metadata.GLProofFiles = make([]string, len(witnessGLs))
	metadata.GLCommitFiles = make([]string, len(witnessGLs))
	metadata.LPPProofFiles = make([]string, len(witnessLPPs))

	logrus.Info("Saving the witnesses")
	wg, ctx := errgroup.WithContext(context.Background())
	wg.SetLimit(numConcurrentWritingGoroutines)

	for i, witnessGL := range witnessGLs {
		i := i
		wg.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:

				var (
					witnessGLFileName = fmt.Sprintf("%s-%s-seg-%d-mod-%d-gl-wit.bin", metadata.StartBlock, metadata.EndBlock, i, witnessGL.ModuleIndex)
					witnessGLDirFrom  = path.Join(cfg.ExecutionLimitless.WitnessDir, "GL", string(witnessGL.ModuleName), config.RequestsFromSubDir)
					witnessGLFile     = path.Join(witnessGLDirFrom, witnessGLFileName)

					witnessGLFilePattern = fmt.Sprintf("%s-%s-seg-%d-mod-%d-gl-wit.bin*", metadata.StartBlock, metadata.EndBlock, i, witnessGL.ModuleIndex)
					witnessGLDirDone     = path.Join(cfg.ExecutionLimitless.WitnessDir, "GL", string(witnessGL.ModuleName), config.RequestsDoneSubDir)
					WitnessGLPatternFrom = filepath.Join(witnessGLDirFrom, witnessGLFilePattern)
					WitnessGLPatternDone = filepath.Join(witnessGLDirDone, witnessGLFilePattern)
				)

				// Clean up any prev. witness file (and temp variants) before starting. This helps addressing the situation
				// where a previous process have been interrupted.
				_ = files.RemoveMatchingFiles(WitnessGLPatternFrom, false)
				_ = files.RemoveMatchingFiles(WitnessGLPatternDone, false)

				if err := serialization.StoreToDisk(witnessGLFile, *witnessGL, true); err != nil {
					return fmt.Errorf("could not save witnessGL: %w", err)
				}

				glProofFileName := fmt.Sprintf("%s-%s-seg-%d-mod-%d-gl-proof.bin", metadata.StartBlock, metadata.EndBlock, i, witnessGLs[i].ModuleIndex)
				glProofFile := path.Join(cfg.ExecutionLimitless.SubproofsDir, "GL", string(witnessGL.ModuleName), config.RequestsFromSubDir, glProofFileName)
				glCommitFileName := fmt.Sprintf("%s-%s-seg-%d-mod-%d-gl-lpp-commit.bin", metadata.StartBlock, metadata.EndBlock, i, witnessGLs[i].ModuleIndex)
				glCommitFile := path.Join(cfg.ExecutionLimitless.CommitsDir, string(witnessGL.ModuleName), config.RequestsFromSubDir, glCommitFileName)

				metadata.GLProofFiles[i] = glProofFile
				metadata.GLCommitFiles[i] = glCommitFile
				witnessGL = nil
				return nil
			}
		})
	}

	for i, witnessLPP := range witnessLPPs {
		i := i
		wg.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:

				var (
					witnessLPPFileName = fmt.Sprintf("%s-%s-seg-%d-mod-%d-lpp-wit.bin", metadata.StartBlock, metadata.EndBlock, i, witnessLPP.ModuleIndex)
					witnessLPPDirFrom  = path.Join(cfg.ExecutionLimitless.WitnessDir, "LPP", string(witnessLPP.ModuleName[0]), config.RequestsFromSubDir)
					witnessLPPFile     = path.Join(witnessLPPDirFrom, witnessLPPFileName)

					witnessLPPFilePattern = fmt.Sprintf("%s-%s-seg-%d-mod-%d-lpp-wit.bin*", metadata.StartBlock, metadata.EndBlock, i, witnessLPP.ModuleIndex)
					witnessLPPDirDone     = path.Join(cfg.ExecutionLimitless.WitnessDir, "LPP", string(witnessLPP.ModuleName[0]), config.RequestsDoneSubDir)
					WitnessLPPPatternFrom = filepath.Join(witnessLPPDirFrom, witnessLPPFilePattern)
					WitnessLPPPatternDone = filepath.Join(witnessLPPDirDone, witnessLPPFilePattern)
				)

				// Clean up any prev. witness file (and temp variants) before starting. This helps addressing the situation
				// where a previous process have been interrupted.
				_ = files.RemoveMatchingFiles(WitnessLPPPatternFrom, false)
				_ = files.RemoveMatchingFiles(WitnessLPPPatternDone, false)

				if err := serialization.StoreToDisk(witnessLPPFile, *witnessLPP, true); err != nil {
					return fmt.Errorf("could not save witnessLPP: %w", err)
				}

				lppProofFileName := fmt.Sprintf("%s-%s-seg-%d-mod-%d-lpp-proof.bin", metadata.StartBlock, metadata.EndBlock, i, witnessLPPs[i].ModuleIndex)
				lppProofFile := path.Join(cfg.ExecutionLimitless.SubproofsDir, "LPP", string(witnessLPP.ModuleName[0]), config.RequestsFromSubDir, lppProofFileName)
				metadata.LPPProofFiles[i] = lppProofFile
				witnessLPP = nil
				return nil
			}
		})
	}

	if err := wg.Wait(); err != nil {
		return fmt.Errorf("could not save witnesses: %v", err)
	}

	sharedRandomnessFileName := fmt.Sprintf("%s-%s-commit.bin", metadata.StartBlock, metadata.EndBlock)
	sharedRandomnessPath := path.Join(cfg.ExecutionLimitless.SharedRandomnessDir, config.RequestsFromSubDir, sharedRandomnessFileName)
	metadata.SharedRndFile = sharedRandomnessPath

	return nil
}

// loadStaticProverAssetsFromDisk: Loads static prover assets from disk into memory for every proof request
func loadStaticProverAssetsFromDisk(cfg *config.Config, assets *zkevm.LimitlessZkEVM) {
	logrus.Infof("Loading bootstrapper and zkevm")

	if err := assets.LoadBootstrapper(cfg); err != nil || assets.DistWizard.Bootstrapper == nil {
		utils.Panic("could not load bootstrapper: %v", err)
	}

	if err := assets.LoadZkEVM(cfg); err != nil || assets.Zkevm == nil {
		utils.Panic("could not load zkevm: %v", err)
	}

	if err := assets.LoadDisc(cfg); err != nil || assets.DistWizard.Disc == nil {
		utils.Panic("could not load disc: %v", err)
	}
}
