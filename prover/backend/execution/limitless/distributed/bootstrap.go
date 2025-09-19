package distributed

import (
	"context"
	"fmt"
	"os"
	"path"

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

type Metadata struct {
	StartBlock string `json:"startBlock"`
	EndBlock   string `json:"endBlock"`
	NumGL      int    `json:"numGL"`
	NumLPP     int    `json:"numLPP"`
}

func Bootstrap(cfg *config.Config, req *execution.Request, metadata Metadata) error {

	// Set MonitorParams before any proving happens
	profiling.SetMonitorParams(cfg)

	// Setting the issue handler to exit on unsatisfied constraint but not limit
	// overflow.
	exit.SetIssueHandlingMode(exit.ExitOnUnsatisfiedConstraint)

	// Setup execution witness and output response
	var (
		out     = execution.CraftProverOutput(cfg, req)
		witness = execution.NewWitness(cfg, req, &out)
	)

	if cfg.Execution.LimitlessWithDebug {
		limitlessZkEVM := zkevm.NewLimitlessDebugZkEVM(cfg)
		limitlessZkEVM.RunDebug(cfg, witness.ZkEVM)
		return nil
	}

	logrus.Info("Starting to run the bootstrapper")

	numGL, numLPP, err := bootstrap(cfg, witness.ZkEVM, metadata)
	if err != nil {
		return fmt.Errorf("error during bootstrap:%s", err.Error())
	}

	logrus.Infof("Bootstrapper generated %d GLs and %d LPPs", numGL, numLPP)

	metadata.NumGL = numGL
	metadata.NumLPP = numLPP

	// Publish metadata.json atomically to --out
	metadataPath := path.Join(config.MetadataDirPrefix, fmt.Sprintf("%s-%s-metadata-getZkProof.json", metadata.StartBlock, metadata.EndBlock))
	if err := files.MustWriteFileAtomic(metadataPath, metadata); err != nil {
		return fmt.Errorf("error during writing metadata: %w", err)
	}

	logrus.Infof("Bootstrapper finished successfully and generated %s file", metadataPath)
	return nil
}

func bootstrap(cfg *config.Config, zkevmWitness *zkevm.Witness, metadata Metadata) (numGL, numLPP int, err error) {

	// TODO: This call will goaway once we implement `mmap` optimization in the controller process
	// By this way, we ensure the static prover assets are loaded into memory only once and it is passed as `memfd`
	// to the prover process, thereby saving disk load time.
	assets := &zkevm.LimitlessZkEVM{}
	loadStaticProverAssetsFromDisk(cfg, assets)

	// The GL and LPP modules are loaded in the background immediately but we
	// only need them for the [distributed.SegmentRuntime] call.
	distDone := make(chan error)
	go func() {
		err := assets.LoadBlueprints(cfg)
		distDone <- err
	}()

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

	// This frees the memory from the assets that are no longer needed. We don't
	// need to do that for the module GLs and LPPs because they are thrown away
	// when the current function returns.
	assets.Zkevm = nil
	assets.DistWizard.Bootstrapper = nil

	if err := <-distDone; err != nil {
		utils.Panic("could not load GL and LPP blueprint modules: %v", err)
	}

	logrus.Infof("Segmenting the runtime")
	witnessGLs, witnessLPPs := distributed.SegmentRuntime(
		runtimeBoot,
		assets.DistWizard.Disc,
		assets.DistWizard.BlueprintGLs,
		assets.DistWizard.BlueprintLPPs,
	)

	logrus.Info("Saving the witnesses")
	eg, ctx := errgroup.WithContext(context.Background())

	for i, witnessGL := range witnessGLs {
		i := i
		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			fileName := fmt.Sprintf("%s-%s-seg-%d-mod-%d-witness.bin", metadata.StartBlock, metadata.EndBlock, i, witnessGL.ModuleIndex)
			filePath := path.Join(config.WitnessGLDirPrefix, string(witnessGL.ModuleName), fileName)

			// Clean up anuy prev. witness file before starting. This helps addressing the situation
			// where a previous process have been interrupted.
			os.Remove(filePath)

			if err := serialization.StoreToDisk(filePath, *witnessGL, true); err != nil {
				return fmt.Errorf("could not save witnessGL: %w", err)
			}
			witnessGL = nil
			return nil
		})
	}

	for i, witnessLPP := range witnessLPPs {
		i := i
		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			fileName := fmt.Sprintf("%s-%s-seg-%d-mod-%d-witness.bin", metadata.StartBlock, metadata.EndBlock, i, witnessLPP.ModuleIndex)
			filePath := path.Join(config.WitnessLPPDirPrefix, string(witnessLPP.ModuleName[0]), fileName)

			// Clean up anuy prev. witness file before starting. This helps addressing the situation
			// where a previous process have been interrupted.
			os.Remove(filePath)
			if err := serialization.StoreToDisk(filePath, *witnessLPP, true); err != nil {
				return fmt.Errorf("could not save witnessLPP: %w", err)
			}
			witnessLPP = nil
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return 0, 0, fmt.Errorf("could not save witnesses: %v", err)
	}

	return len(witnessGLs), len(witnessLPPs), nil
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
