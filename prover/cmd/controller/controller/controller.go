package controller

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/consensys/linea-monorepo/prover/cmd/controller/controller/metrics"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

const terminationTypeFile = "/tmp/termination-type"

// isSpotInstanceReclaim checks if the termination is due to spot instance reclamation.
// It reads /tmp/termination-type file to determine the shutdown type:
//   - File missing or unreadable: ASSUME SPOT RECLAIM (safer default for spot instances)
//   - File contains "SPOT_RECLAIM": Spot instance reclamation
//   - File contains "NORMAL_SHUTDOWN": Normal K8s/manual shutdown
//   - File contains anything else: Assume spot reclaim
//
// To explicitly request normal shutdown, create the file with "NORMAL_SHUTDOWN":
//
//	echo "NORMAL_SHUTDOWN" > /tmp/termination-type
func isSpotInstanceReclaim() bool {
	data, err := os.ReadFile(terminationTypeFile)
	if err != nil {
		// File doesn't exist or can't be read - assume spot reclaim for safety
		// This prevents jobs from wasting time on doomed spot instances
		logrus.Warnf("File %s not found, defaulting to spot instance reclaim mode", terminationTypeFile)
		return true
	}

	content := strings.TrimSpace(string(data))

	// Only explicit "NORMAL_SHUTDOWN" triggers graceful shutdown
	if content == "NORMAL_SHUTDOWN" {
		logrus.Infof("File %s contains NORMAL_SHUTDOWN, using graceful shutdown", terminationTypeFile)
		return false
	}

	// Everything else (including "SPOT_RECLAIM" or empty) is treated as spot reclaim
	logrus.Warnf("File %s contains '%s', treating as spot instance reclaim", terminationTypeFile, content)
	return true
}

// function to run the controller
func runController(ctx context.Context, cfg *config.Config) {
	var (
		cLog          = cfg.Logger().WithField("component", "main-loop")
		fsWatcher     = NewFsWatcher(cfg)
		executor      = NewExecutor(cfg)
		numRetrySoFar int
		state         = NewControllerState() // Track controller state for graceful shutdown
	)

	// Start the metric server with readiness endpoint
	if cfg.Controller.Prometheus.Enabled {
		metrics.StartServerWithReadiness(
			cfg.Controller.LocalID,
			cfg.Controller.Prometheus.Route,
			cfg.Controller.Prometheus.Port,
			state,
		)
	}

	// Listen for both SIGTERM and SIGUSR1
	// SIGUSR1 is sent by external spot reclaim detection scripts
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGUSR1)
	defer stop()

	// cmdContext is the context we provide for the command execution.
	// We always create a cancellable context now for graceful shutdown
	cmdContext, cmdCancel := context.WithCancel(context.Background())
	defer cmdCancel()

	// Graceful shutdown handler
	go func() {
		<-ctx.Done()

		// Check if this is a spot instance reclamation
		isSpotReclaim := isSpotInstanceReclaim()

		if isSpotReclaim {
			cLog.Warnln("Spot instance reclamation detected (SIGUSR1 or /tmp/termination-type=SPOT_RECLAIM)")
			cLog.Warnln("Initiating immediate shutdown to requeue job before VM termination")

			// Set a very short grace period for spot reclaim (just enough to cleanup)
			deadline := time.Now().Add(10 * time.Second)
			state.RequestSpotReclaimShutdown(deadline)

			// Cancel immediately to trigger job requeuing
			cmdCancel()
			return
		}

		// Normal shutdown (K8s rolling update, manual shutdown, etc.)
		// Calculate shutdown deadline
		gracePeriod := cfg.Controller.TerminationGracePeriod
		if gracePeriod == 0 {
			gracePeriod = 3550 * time.Second // Default if not set
		}
		deadline := time.Now().Add(gracePeriod)
		state.RequestShutdown(deadline)

		cLog.Infof(
			"Received SIGTERM. Grace period: %v (until %v)",
			gracePeriod,
			deadline.Format(time.RFC3339),
		)

		if state.IsProcessing() {
			job := state.GetCurrentJob()
			if job != nil {
				cLog.Infof(
					"Job %v is currently processing. Will wait for completion or grace period expiry.",
					job.OriginalFile,
				)
			}
		} else {
			cLog.Infoln("No job currently processing. Will exit after cleanup.")
			// If no job is running, cancel immediately to exit fast
			cmdCancel()
			return
		}

		// If job is running, monitor it and only cancel when needed
		// Leave some buffer time for cleanup (e.g., 10 seconds before deadline)
		shutdownTimeout := cfg.Controller.ChildProcessShutdownTimeout
		if shutdownTimeout == 0 {
			shutdownTimeout = 10 * time.Second
		}

		// Cancel context when we're close to the deadline
		gracefulDeadline := deadline.Add(-shutdownTimeout)

		cLog.Infof(
			"Will allow job to run until %v before initiating shutdown sequence",
			gracefulDeadline.Format(time.RFC3339),
		)

		// Monitor job status and deadline
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Check if job finished
				if !state.IsProcessing() {
					cLog.Infoln("Job completed before grace period, cancelling context for cleanup")
					cmdCancel()
					return
				}

				// Check if we're approaching deadline
				if time.Now().After(gracefulDeadline) {
					cLog.Warnln("Grace period approaching end, initiating shutdown sequence")
					cmdCancel()
					return
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			// Graceful shutdown
			cLog.Infoln("Entering graceful shutdown phase")

			// The shutdown handler goroutine is already monitoring the job
			// and will call cmdCancel() when needed. We just need to wait
			// for the job to finish (state will become not processing).

			if state.IsProcessing() {
				// Wait for shutdown handler to complete its work
				// It will call cmdCancel() which will trigger job shutdown
				cLog.Infoln("Waiting for shutdown handler to complete job termination")

				// Simple wait loop - just check if processing stopped
				ticker := time.NewTicker(1 * time.Second)
				defer ticker.Stop()

				for state.IsProcessing() {
					<-ticker.C
				}

				cLog.Infoln("Job processing completed, proceeding with cleanup")
			}

			cLog.Infoln("Shutting down metrics server")
			metrics.ShutdownServer(context.Background())
			cLog.Infoln("Graceful shutdown complete")
			return

		// Processing a new job
		case <-retryDelay(cfg.Controller.RetryDelays, numRetrySoFar):
			// Check if shutdown was requested - don't pick up new jobs
			if state.IsShutdownRequested() {
				cLog.Infoln("Shutdown requested, not picking up new jobs")
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

			// Else, reset the retry counter
			numRetrySoFar = 0

			// Mark job as processing
			state.StartProcessing(job)

			// Run the command (potentially retrying in large mode)
			status := executor.Run(cmdContext, job, state)

			// Mark job as finished
			state.StopProcessing()

			// createColumns the job according to the status we got
			switch {

			// Success
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

				logrus.Infof(
					"Moving the response file from the tmp response file `%v`, to the final response file: `%v`",
					tmpRespFile, respFile,
				)

				if err := os.Rename(tmpRespFile, respFile); err != nil {
					// @Alex: it is unclear how the rename operation could fail
					// here. If this happens, we prefer removing the tmp file.
					// Note that the operation is an `rm -f`.
					os.Remove(tmpRespFile)

					cLog.Errorf(
						"Error renaming %v to %v: %v, removed the tmp file",
						tmpRespFile, respFile, err,
					)
				}

				// Move the inprogress to the done directory
				cLog.Infof(
					"Moving %v to %v with the success prefix",
					job.OriginalFile, job.Def.dirDone(),
				)

				jobDone := job.DoneFile(status)
				if err := os.Rename(job.InProgressPath(), jobDone); err != nil {
					// When that happens, the only thing left to do is to log
					// the error and let the inprogress file where it is. It
					// will likely require a human intervention.
					//
					// Note: this is assumedly an unreachable code path.
					cLog.Errorf(
						"Error renaming %v to %v: %v",
						job.InProgressPath(), jobDone, err,
					)
				}

			// Defer to the large prover
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
					cLog.Errorf(
						"error deriving the to-large-name of %v: %v",
						job.InProgressPath(), err,
					)
				}

				if err := os.Rename(job.InProgressPath(), toLargePath); err != nil {
					// When that happens, the only thing left to do is to log
					// the error and let the inprogress file where it is. It
					// will likely require a human intervention.
					cLog.Errorf(
						"error renaming %v to %v: %v",
						job.InProgressPath(), toLargePath, err,
					)
				}

			case status.ExitCode == CodeKilledByUs:
				// When receiving the killed-by-us code, the prover will put
				// back the file in the queue. We do not immediately exit the
				// loop as there is already a clause to exit the loop in case
				// of sigterm at the top of the loop. So it will immediately
				// exit as soon as we enter the next iteration.
				cLog.Infof("Job %v was killed by us. Reputting it in the request folder", job.OriginalFile)

				if err := os.Rename(job.InProgressPath(), job.OriginalPath()); err != nil {
					// When that happens, the only thing left to do is to log
					// the error and let the inprogress file where it is. It
					// will likely require a human intervention.
					//
					// Note: this is assumedly an unreachable code path.
					cLog.Errorf(
						"Error renaming %v to %v: %v",
						job.InProgressPath(), job.OriginalPath(), err,
					)
				}

				// As an edge-case, it's possible (in theory) that the process
				// completes exactly when we receive the kill signal. So we
				// could end up in a situation where the tmp-response file
				// exists. In that case, we simply delete it before exiting to
				// keep the FS clean.
				os.Remove(job.TmpResponseFile(cfg))

			// Failure case
			default:
				// Move the inprogress to the done directory
				cLog.Infof(
					"Moving %v with in %v with a failure suffix for code %v",
					job.OriginalFile, job.Def.dirDone(), status.ExitCode,
				)

				jobFailed := job.DoneFile(status)
				if err := os.Rename(job.InProgressPath(), jobFailed); err != nil {
					// When that happens, the only thing left to do is to log
					// the error and let the inprogress file where it is. It
					// will likely require a human intervention.
					//
					// Note: this is assumedly an unreachable code path.
					cLog.Errorf(
						"Error renaming %v to %v: %v",
						job.InProgressPath(), jobFailed, err,
					)
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
