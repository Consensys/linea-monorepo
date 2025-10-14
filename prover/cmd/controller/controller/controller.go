package controller

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/consensys/linea-monorepo/prover/cmd/controller/controller/metrics"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

type shutdownSignal int

const (
	noShutdown       shutdownSignal = iota
	gracefulShutdown                // SIGTERM
	spotReclaim                     // SIGUSR1
)

func runController(ctx context.Context, cfg *config.Config) {
	var (
		cLog          = cfg.Logger().WithField("component", "main-loop")
		fsWatcher     = NewFsWatcher(cfg)
		executor      = NewExecutor(cfg)
		numRetrySoFar int
	)

	if cfg.Controller.Prometheus.Enabled {
		metrics.StartServer(
			cfg.Controller.LocalID,
			cfg.Controller.Prometheus.Route,
			cfg.Controller.Prometheus.Port,
		)
	}

	signalChan := make(chan os.Signal, 2)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGUSR1)
	defer signal.Stop(signalChan)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	shutdownType := noShutdown

	// Track currently active job for safe requeue
	var activeJob *Job

	go func() {
		for sig := range signalChan {
			switch sig {
			case syscall.SIGUSR1:
				shutdownType = spotReclaim
				cancel()
				if cfg.Controller.Prometheus.Enabled {
					metrics.IncSpotInterruption(cfg.Controller.LocalID) // or another label like instance_type
				}
				return
			case syscall.SIGTERM:
				shutdownType = gracefulShutdown
				cancel()
				return
			}
		}
	}()

	cmdContext := context.Background()
	if cfg.Controller.SpotInstanceMode {
		cmdContext = ctx
	}

	go func() {
		<-ctx.Done()
		switch shutdownType {
		case spotReclaim:
			cLog.Infof("Received spot reclaim (SIGUSR1). Will abort and requeue ASAP (max %s)...", config.SpotInstanceReclaimTime)
		case gracefulShutdown:
			cLog.Infof("Received SIGTERM. Finishing in-flight proof (if any) and then shutting down gracefully (max %s)...", cfg.Controller.TerminationGracePeriod)
		default:
			cLog.Infoln("Received cancellation request.")
		}
	}()

	for {
		select {
		case <-ctx.Done():
			metrics.ShutdownServer(ctx)

			// If spot reclaim and activeJob exists, requeue it safely exactly once
			if shutdownType == spotReclaim && activeJob != nil {
				cLog.Infof("Spot reclaim: requeuing active job %v", activeJob.OriginalFile)
				// Remove temp response file if exists
				_ = os.Remove(activeJob.TmpResponseFile(cfg))

				// Move .inprogress file back to original request path
				err := os.Rename(activeJob.InProgressPath(), activeJob.OriginalPath())
				if err != nil {
					cLog.Errorf("Failed to requeue job %v: %v", activeJob.InProgressPath(), err)
				}
				// Prevent double requeue
				activeJob = nil
			}

			switch shutdownType {
			case spotReclaim:
				<-time.After(config.SpotInstanceReclaimTime)
			case gracefulShutdown:
				<-time.After(cfg.Controller.TerminationGracePeriod.Abs())
			default:
				// immediate exit
			}
			return

		case <-retryDelay(cfg.Controller.RetryDelays, numRetrySoFar):
			// Prevent starting new jobs if shutdown started
			if ctx.Err() != nil {
				continue
			}

			job := fsWatcher.GetBest()
			if job == nil {
				numRetrySoFar++
				noJobFoundMsg := "found no jobs in the queue"
				if numRetrySoFar > 5 {
					cLog.Debug(noJobFoundMsg)
				} else {
					cLog.Info(noJobFoundMsg)
				}
				continue
			}
			numRetrySoFar = 0

			// Set active job to current job
			activeJob = job

			status := executor.Run(cmdContext, job)

			// Clear active job on completion
			activeJob = nil

			switch {
			case status.ExitCode == CodeSuccess:
				respFile, err := job.ResponseFile()
				tmpRespFile := job.TmpResponseFile(cfg)
				if err != nil {
					utils.Panic("Could not generate the response file: %v (original request file: %v)", err, job.OriginalFile)
				}
				logrus.Infof(
					"Moving the response file from the tmp response file `%v`, to the final response file: `%v`",
					tmpRespFile, respFile,
				)
				if err := os.Rename(tmpRespFile, respFile); err != nil {
					os.Remove(tmpRespFile)
					cLog.Errorf("Error renaming %v to %v: %v, removed the tmp file", tmpRespFile, respFile, err)
				}
				cLog.Infof("Moving %v to %v with the success prefix", job.OriginalFile, job.Def.dirDone())
				jobDone := job.DoneFile(status)
				if err := os.Rename(job.InProgressPath(), jobDone); err != nil {
					cLog.Errorf("Error renaming %v to %v: %v", job.InProgressPath(), jobDone, err)
				}

			case job.Def.Name == jobNameExecution && isIn(status.ExitCode, cfg.Controller.DeferToOtherLargeCodes):
				cLog.Infof("Renaming %v for the large prover", job.OriginalFile)
				toLargePath, err := job.DeferToLargeFile(status)
				if err != nil {
					cLog.Errorf("error deriving the to-large-name of %v: %v", job.InProgressPath(), err)
				}
				if err := os.Rename(job.InProgressPath(), toLargePath); err != nil {
					cLog.Errorf("error renaming %v to %v: %v", job.InProgressPath(), toLargePath, err)
				}

			case status.ExitCode == CodeKilledByUs:
				cLog.Infof("Job %v was killed by us. Requeuing it back to the request folder", job.OriginalFile)
				if err := os.Rename(job.InProgressPath(), job.OriginalPath()); err != nil {
					cLog.Errorf("Error renaming %v to %v: %v", job.InProgressPath(), job.OriginalPath(), err)
				}
				os.Remove(job.TmpResponseFile(cfg))

			default:
				cLog.Infof("Moving %v with in %v with a failure suffix for code %v", job.OriginalFile, job.Def.dirDone(), status.ExitCode)
				jobFailed := job.DoneFile(status)
				if err := os.Rename(job.InProgressPath(), jobFailed); err != nil {
					cLog.Errorf("Error renaming %v to %v: %v", job.InProgressPath(), jobFailed, err)
				}
			}
		}
	}
}

// Returns the duration to wait before retrying to find a job in the queue. This
// is to avoid spamming the FS with LS queries.
func retryDelay(retryDelaysSec []int, numRetrySoFar int) <-chan time.Time {

	// Hard coded values
	retryDelaysDur := make([]time.Duration, len(retryDelaysSec))
	for i := range retryDelaysSec {
		retryDelaysDur[i] = time.Duration(retryDelaysSec[i]) * time.Second
	}

	// By default, take the last value
	ttw := retryDelaysDur[len(retryDelaysDur)-1]

	// If it does not overflows, take the last value at `numSoFar`
	if numRetrySoFar < len(retryDelaysDur) {
		ttw = retryDelaysDur[numRetrySoFar]
	}

	return time.After(ttw)
}
