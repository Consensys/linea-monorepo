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
	noShutdown shutdownSignal = iota

	// SIGTERM
	gracefulShutdown

	// SIGUSR1
	spotReclaim
)

func runController(ctx context.Context, cfg *config.Config) {
	var (
		cLog      = cfg.Logger().WithField("component", "main-loop")
		fsWatcher = NewFsWatcher(cfg)
		executor  = NewExecutor(cfg)

		signalChan = make(chan os.Signal, 2)

		// Default shutdown type
		shutdownType = noShutdown

		// Track currently active job for safe requeue
		activeJob *Job

		numRetrySoFar int
	)

	// Start the metric server
	if cfg.Controller.Prometheus.Enabled {
		metrics.StartServer(
			cfg.Controller.LocalID,
			cfg.Controller.Prometheus.Route,
			cfg.Controller.Prometheus.Port,
		)
	}

	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGUSR1)
	defer signal.Stop(signalChan)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

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

	// cmdContext is the context we provide for the command execution. In
	// spot-instance mode, the context is subordinated to the parent ctx so
	// that the cancellation of parent ctx is propagated to the cmdContext.
	cmdContext := context.Background()
	if cfg.Controller.SpotInstanceMode {
		cmdContext = ctx
	}

	go func() {
		<-ctx.Done()
		switch shutdownType {
		case spotReclaim:
			cLog.Infof("Received spot reclaim (SIGUSR1). Will abort ASAP (max %s)...", config.SpotInstanceReclaimTime)
		case gracefulShutdown:
			cLog.Infof("Received SIGTERM. Finishing in-flight proof (if any) and then shutting down gracefully (max %s)...", cfg.Controller.TerminationGracePeriod)
		default:
			cLog.Infoln("Received cancellation request.")
		}
	}()

	for {
		select {
		case <-ctx.Done():
			// Graceful shutdown.
			// This case captures both cancellations initiated by the caller
			// through ctx and SIGTERM signals. Even if the cancellation
			// request is first intercepted by the goroutine at line 34, Go
			// allows the ctx.Done channel to be read multiple times, which, in
			// our scenario, ensures cancellation requests are effectively
			// detected and handled.
			cLog.Infoln("Context cancelled by caller or externally triggered (SIGTERM/SIGUSR1). Exiting")
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

			// Processing a new job
		case <-retryDelay(cfg.Controller.RetryDelays, numRetrySoFar):
			// Prevent starting new jobs if shutdown started
			if ctx.Err() != nil {
				continue
			}

			// Fetch the best block we can fetch
			job := fsWatcher.GetBest()

			// No jobs, waiting a little before we retry
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

			// Reset the retry counter
			numRetrySoFar = 0

			// Set active job to current job
			activeJob = job

			// Run the command (potentially retrying in large mode)
			status := executor.Run(cmdContext, job)

			// Clear active job on completion
			activeJob = nil

			switch {
			case status.ExitCode == CodeSuccess:
				// NB: we already check that the response filename can be
				// generated prior to running the command. So this actually
				// will not panic.
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
					// @Alex: it is unclear how the rename operation could fail
					// here. If this happens, we prefer removing the tmp file.
					// Note that the operation is an `rm -f`.
					os.Remove(tmpRespFile)
					cLog.Errorf("Error renaming %v to %v: %v, removed the tmp file", tmpRespFile, respFile, err)
				}
				cLog.Infof("Moving %v to %v with the success prefix", job.OriginalFile, job.Def.dirDone())
				jobDone := job.DoneFile(status)
				if err := os.Rename(job.InProgressPath(), jobDone); err != nil {
					// When that happens, the only thing left to do is to log
					// the error and let the inprogress file where it is. It
					// will likely require a human intervention.
					//
					// Note: this is assumedly an unreachable code path.
					cLog.Errorf("Error renaming %v to %v: %v", job.InProgressPath(), jobDone, err)
				}

			case job.Def.Name == jobNameExecution && isIn(status.ExitCode, cfg.Controller.DeferToOtherLargeCodes):
				cLog.Infof("Renaming %v for the large prover", job.OriginalFile)
				// Move the inprogress file back in the from directory with the new suffix
				toLargePath, err := job.DeferToLargeFile(status)
				if err != nil {
					// There are two possibilities of errors. (1), the status
					// we success but the above cases prevents that. The other
					// case is that the suffix was not provided. But, during
					// the config validation, we check already that the suffix
					// must be provided if the size of the list of
					// deferToOtherLargeCodes is non-zero. If the size of the
					// list was zero, then there would be no way to reach this
					// portion of the code given that the current exit code
					// cannot be part of the empty list. Thus, this section is
					// unreachable.
					cLog.Errorf("error deriving the to-large-name of %v: %v", job.InProgressPath(), err)
				}
				if err := os.Rename(job.InProgressPath(), toLargePath); err != nil {
					// When that happens, the only thing left to do is to log
					// the error and let the inprogress file where it is. It
					// will likely require a human intervention.
					cLog.Errorf("error renaming %v to %v: %v", job.InProgressPath(), toLargePath, err)
				}

			case status.ExitCode == CodeKilledByUs:
				cLog.Infof("Job %v was killed externally. Requeuing the request back to the request folder", job.OriginalFile)
				if err := os.Rename(job.InProgressPath(), job.OriginalPath()); err != nil {
					// When that happens, the only thing left to do is to log
					// the error and let the inprogress file where it is. It
					// will likely require a human intervention.
					//
					// Note: this is assumedly an unreachable code path.
					cLog.Errorf("Error renaming %v to %v: %v", job.InProgressPath(), job.OriginalPath(), err)
				}

				// As an edge-case, it's possible (in theory) that the process
				// completes exactly when we receive the kill signal. So we
				// could end up in a situation where the tmp-response file
				// exists. In that case, we simply delete it before exiting to
				// keep the FS clean.
				os.Remove(job.TmpResponseFile(cfg))

			// Failure case
			default:
				cLog.Infof("Moving %v with in %v with a failure suffix for code %v", job.OriginalFile, job.Def.dirDone(), status.ExitCode)
				jobFailed := job.DoneFile(status)
				if err := os.Rename(job.InProgressPath(), jobFailed); err != nil {
					// When that happens, the only thing left to do is to log
					// the error and let the inprogress file where it is. It
					// will likely require a human intervention.
					//
					// Note: this is assumedly an unreachable code path.

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
