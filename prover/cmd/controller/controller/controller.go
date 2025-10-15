package controller

import (
	"context"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/consensys/linea-monorepo/prover/cmd/controller/controller/metrics"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

func runController(ctx context.Context, cfg *config.Config) {
	var (
		cLog      = cfg.Logger().WithField("component", "main-loop")
		fsWatcher = NewFsWatcher(cfg)
		executor  = NewExecutor(cfg)

		signalChan = make(chan os.Signal, 2)

		// Track currently active job for safe requeue
		activeJob *Job

		// Track the number of retries so far
		numRetrySoFar int

		// Atomic coordination flags
		spotReclaimDetected       atomic.Bool
		gracefulShutdownRequested atomic.Bool
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

	// cmdContext is the context we provide for the command execution. In
	// spot-instance mode, the context is subordinated to the ctx.
	cmdContext := context.Background()
	if cfg.Controller.SpotInstanceMode {
		cmdContext = ctx
	}

	// Signal handler goroutine — race-safe with atomic flags
	// We never stop listening since it is possible to receive more than one signals
	// For ex: spot reclaim => SIGUSR1 -> SIGTERM
	go func() {
		for sig := range signalChan {
			switch sig {
			case syscall.SIGUSR1:
				spotReclaimDetected.Store(true)
				if !gracefulShutdownRequested.Load() {
					cLog.Info("Received SIGUSR1: marking spot reclaim detected, cancelling context")
					cancel()
					if cfg.Controller.Prometheus.Enabled {
						metrics.IncSpotInterruption(cfg.Controller.LocalID)
					}
				} else {
					cLog.Info("Received SIGUSR1 after SIGTERM: marking spot reclaim detected (shutdown already in progress)")
				}

				if activeJob != nil {
					cLog.Infof("Spot reclaim detected. REQUEUE active job: %v", activeJob.OriginalFile)
					_ = os.Remove(activeJob.TmpResponseFile(cfg))
					if err := os.Rename(activeJob.InProgressPath(), activeJob.OriginalPath()); err != nil {
						cLog.Errorf("Failed to requeue job %v: %v", activeJob.InProgressPath(), err)
					}
					activeJob = nil
				}
			case syscall.SIGTERM:
				gracefulShutdownRequested.Store(true)
				if !spotReclaimDetected.Load() {
					cLog.Info("Received SIGTERM: graceful shutdown requested")
				} else {
					cLog.Info("Received SIGTERM after SIGUSR1: will use spot reclaim timeout")
				}
				cancel()
			}
		}
	}()

	// This goroutine's raison d'etre is to log a message immediately when a
	// cancellation request (e.g., ctx expiration/cancellation, SIGTERM, etc.)
	// is received. It ensures timely logging of the request's reception,
	// which may be important for diagnostics. Without this
	// goroutine, if the prover is busy with a proof when, for example, a
	// SIGTERM is received, there would be no log entry about the signal
	// until the proof completes.
	go func() {
		<-ctx.Done()
		if spotReclaimDetected.Load() {
			cLog.Infof("Received SIGUSR1 (spot reclaim). Will abort ASAP (max %s)...", cfg.Controller.SpotInstanceReclaimTime)
		} else if gracefulShutdownRequested.Load() {
			cLog.Infof("Received SIGTERM. Finishing in-flight proof (if any) and then shutting down gracefully (max %s)...", cfg.Controller.TerminationGracePeriod)
		} else {
			cLog.Infoln("Received cancellation request.")
		}
	}()

	for {
		select {
		case <-ctx.Done():
			cLog.Infoln("Context cancelled by caller or externally triggered (SIGTERM/SIGUSR1). Exiting...")
			metrics.ShutdownServer(ctx)

			// Defensive: if controller is killed while a job is somehow still active!
			if spotReclaimDetected.Load() && activeJob != nil {
				cLog.Infof("Spot reclaim (shutdown branch): requeuing active job %v", activeJob.OriginalFile)
				_ = os.Remove(activeJob.TmpResponseFile(cfg))
				if err := os.Rename(activeJob.InProgressPath(), activeJob.OriginalPath()); err != nil {
					cLog.Errorf("Failed to requeue job %v: %v", activeJob.InProgressPath(), err)
				}
				activeJob = nil
			}

			// Apply appropriate grace period
			switch {
			case spotReclaimDetected.Load():
				time.Sleep(cfg.Controller.SpotInstanceReclaimTime)
			case gracefulShutdownRequested.Load():
				time.Sleep(cfg.Controller.TerminationGracePeriod.Abs())
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

			numRetrySoFar = 0

			// Important: Set the active job to the current job for safe requeue in case of spot-reclaim
			activeJob = job

			// Run the command (potentially retrying in large mode)
			status := executor.Run(cmdContext, job)

			// CreateColumns the job according to the status we got
			switch {
			case status.ExitCode == CodeSuccess:
				// NB: we already check that the response filename can be
				// generated prior to running the command. So this actually
				// will not panic.
				respFile, err := job.ResponseFile()
				tmpRespFile := job.TmpResponseFile(cfg)
				if err != nil {
					formatStr := "Could not generate the response file: %v (original request file: %v)"
					utils.Panic(formatStr, err, job.OriginalFile)
				}
				logrus.Infof("Moving tmp response file %v to final response file: %v", tmpRespFile, respFile)
				if err := os.Rename(tmpRespFile, respFile); err != nil {
					// It is unclear how the rename operation could fail
					// here. If this happens, we prefer removing the tmp file.
					// Note that the operation is an `rm -f`.
					os.Remove(tmpRespFile)
					cLog.Errorf("Error renaming %v to %v: %v, removed tmp file", tmpRespFile, respFile, err)
				}

				// Move the inprogress to the done directory
				cLog.Infof("Moving %v to %v with success prefix", job.OriginalFile, job.Def.dirDone())
				jobDone := job.DoneFile(status)
				if err := os.Rename(job.InProgressPath(), jobDone); err != nil {
					// When that happens, the only thing left to do is to log
					// the error and let the inprogress file where it is. It
					// will likely require a human intervention.
					//
					// Note: this is assumedly an unreachable code path
					cLog.Errorf("Error renaming %v to %v: %v", job.InProgressPath(), jobDone, err)
				}

				// Set active job to nil once the job is successful
				activeJob = nil

			// Defer to the large prover: We do not set active job to nil, because the job is not complete yet
			case job.Def.Name == jobNameExecution && isIn(status.ExitCode, cfg.Controller.DeferToOtherLargeCodes):
				cLog.Infof("Renaming %v for the large prover", job.OriginalFile)
				// Move the inprogress file back in the from directory with
				// the new suffix
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
					cLog.Errorf("Error deriving large name for %v: %v", job.InProgressPath(), err)
				}
				if err := os.Rename(job.InProgressPath(), toLargePath); err != nil {
					cLog.Errorf("Error renaming %v to %v: %v", job.InProgressPath(), toLargePath, err)
				}

			// Controller killed the job via external signal handlers
			case status.ExitCode == CodeKilledByUs:
				// When receiving the killed-by-us code, the prover will put back the file in the request queue
				// only spot reclaim is detected. For graceful shutdown (only SIGTERM),
				// the proof will be continue until terminationGracePeriod is reached and will not be requeued.
				// REMARK: In case the proof fails to complete before terminationGracePeriod is reached, the file
				// will have to be manually requeued.
				cLog.Infof("Active job %v killed by us. Requeue the job only if spot reclaim detected.", job.OriginalFile)

				// As an edge-case, it's possible (in theory) that the process
				// completes exactly when we receive the kill signal. So we
				// could end up in a situation where the tmp-response file
				// exists. In that case, we simply delete it before exiting to
				// keep the FS clean.
				os.Remove(job.TmpResponseFile(cfg))

			// Failure case
			default:
				// Move the inprogress to the done directory
				cLog.Infof("Moving %v with failure suffix (code %v)", job.OriginalFile, status.ExitCode)
				jobFailed := job.DoneFile(status)
				if err := os.Rename(job.InProgressPath(), jobFailed); err != nil {
					// When that happens, the only thing left to do is to log and will require human intervention
					cLog.Errorf("Error renaming %v to %v: %v", job.InProgressPath(), jobFailed, err)
				}

				// Set active job to nil as there is nothing more to retry
				activeJob = nil
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
