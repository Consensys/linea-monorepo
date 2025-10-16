package controller

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/consensys/linea-monorepo/prover/cmd/controller/controller/metrics"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/config/assets"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

type ControllerState struct {
	activeJob     *Job
	jobDoneChan   chan struct{}
	jobDoneChanMu sync.Mutex
}

// runController runs the controller main loop.
func runController(ctx context.Context, cfg *config.Config) {
	var (
		cLog      = cfg.Logger().WithField("component", "main-loop")
		fsWatcher = NewFsWatcher(cfg)
		executor  = NewExecutor(cfg)

		// Tracks the controller for active job and job done channel
		state = &ControllerState{}

		// Track the number of retries so far
		numRetrySoFar int

		// Atomic coordination flags
		spotReclaimDetected       atomic.Bool
		gracefulShutdownRequested atomic.Bool

		// Channel to receive signals
		signalChan = make(chan os.Signal, 2)
	)

	// Helper to requeue a job
	requeueJob := func(job *Job) {
		_ = os.Remove(job.TmpResponseFile(cfg))
		if err := os.Rename(job.InProgressPath(), job.OriginalPath()); err != nil {
			cLog.Errorf("Failed to requeue job %v: %v", job.InProgressPath(), err)
		}
		cLog.Infof("REQUEUED job: %v", job.OriginalFile)
	}

	// Start the metric server
	if cfg.Controller.Prometheus.Enabled {
		metrics.StartServer(
			cfg.Controller.LocalID,
			cfg.Controller.Prometheus.Route,
			cfg.Controller.Prometheus.Port,
		)
	}

	// Derive the command context with a cancel function
	cmdCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGUSR1)
	defer signal.Stop(signalChan)

	// Signal handler goroutine — race-safe with atomic flags
	// We never stop listening since it is possible to receive more than one signals
	// For ex: spot reclaim => SIGUSR1 -> SIGTERM
	go func() {
		for sig := range signalChan {
			switch sig {
			case syscall.SIGUSR1:
				spotReclaimDetected.Store(true)
				if !gracefulShutdownRequested.Load() {
					cLog.Info("Received SIGUSR1: marking spot reclaim detected, cancelling context ASAP...")
					cancel()
					if cfg.Controller.Prometheus.Enabled {
						metrics.IncSpotInterruption(cfg.Controller.LocalID)
					}
				} else {
					cLog.Info("Received SIGUSR1 after SIGTERM: marking spot reclaim detected (shutdown already in progress)")
				}

				// Requeue active job immediately for spot reclaim
				if state.activeJob != nil {
					cLog.Infof("Spot reclaim detected during job: %v", state.activeJob.OriginalFile)
					requeueJob(state.activeJob)
					state.activeJob = nil
					state.notifyJobDone()
				}
			case syscall.SIGTERM:
				gracefulShutdownRequested.Store(true)
				if !spotReclaimDetected.Load() {
					cLog.Info("Received SIGTERM: graceful shutdown requested")
				}
				// NOTE: DO NOT call cancel() here to respect the termination grace period
				// If there's an active job, start a timer to enforce grace period
				if state.activeJob != nil {
					state.jobDoneChanMu.Lock()
					if state.jobDoneChan == nil {
						state.jobDoneChan = make(chan struct{})
					}
					// Using a local copy prevents subtle bugs and race conditions that can happen if the global jobDoneChan
					// is modified from elsewhere
					localJobDoneChan := state.jobDoneChan
					state.jobDoneChanMu.Unlock()
					go func() {
						cLog.Infof("Allowing in-flight job to finish (max %ds)...", cfg.Controller.TerminationGracePeriod)
						select {
						case <-time.After(time.Duration(cfg.Controller.TerminationGracePeriod) * time.Second):
							cLog.Info("Termination grace period expired. Cancelling context to force shutdown...")
							cancel()

						// Closing the done channel immediately  succeeds in reading (returning zero value)
						// This case is triggered when the running job is done and the channel is closed—letting the
						// shutdown happen promptly instead of waiting out the whole grace period.
						case <-localJobDoneChan:
							cLog.Info("Job finished before grace period expired. Cancelling context to exit immediately...")
							cancel()
						}
					}()
				} else {
					cLog.Info("No active job during SIGTERM, cancelling context to exit immediately")
					cancel()
				}
			}
		}
	}()

	// This goroutine's purpose is to log a message immediately when a
	// cancellation request (e.g., ctx expiration/cancellation, SIGTERM, etc.)
	// is received. It ensures timely logging of the request's reception,
	// which may be important for diagnostics. Without this
	// goroutine, if the prover is busy with a proof when, for example, a
	// SIGTERM is received, there would be no log entry about the signal
	// until the proof completes.
	go func() {
		<-cmdCtx.Done()
		if spotReclaimDetected.Load() {
			cLog.Infof("Context done due to spot reclaim. Aborting ASAP (max %ds)...", cfg.Controller.SpotInstanceReclaimTime)
		} else if gracefulShutdownRequested.Load() {
			if state.activeJob != nil {
				cLog.Infof("Context done: grace period expired, forcing shutdown")
			} else {
				cLog.Info("Context done: graceful shutdown complete")
			}
		} else {
			cLog.Infoln("Context done due to cancellation request.")
		}
	}()

	// If any files are locked to RAM unlock it before exiting
	defer assets.UnlockAllLockedFiles()

	// Main loop
	for {
		select {
		case <-ctx.Done():
			cLog.Infoln("Context cancelled by caller or externally triggered (SIGTERM/SIGUSR1). Exiting...")
			metrics.ShutdownServer(cmdCtx)

			if spotReclaimDetected.Load() {
				cLog.Infof("Waiting up to %ds for spot reclaim...", cfg.Controller.SpotInstanceReclaimTime)
				time.Sleep(time.Duration(cfg.Controller.SpotInstanceReclaimTime) * time.Second)
			}

			// For graceful shutdown, do not sleep here and requeue only if
			// the job is still active (timer already slept in signal handler) - See signal handler go routine
			if !spotReclaimDetected.Load() && gracefulShutdownRequested.Load() && state.activeJob != nil {
				cLog.Infof("Job %v did not finish before termination grace period. Requeuing...", state.activeJob.OriginalFile)
				requeueJob(state.activeJob)
				state.activeJob = nil
				state.notifyJobDone()
			}
			return

			// Processing a new job
		case <-retryDelay(cfg.Controller.RetryDelays, numRetrySoFar):
			// Prevent starting new jobs if shutdown started
			if cmdCtx.Err() != nil {
				continue
			}

			//  Skip fetching new jobs if graceful shutdown or spot reclaim is in progress
			if gracefulShutdownRequested.Load() || spotReclaimDetected.Load() {
				numRetrySoFar++
				noJobFoundMsg := "Shutdown in progress, skipping new jobs"
				if numRetrySoFar > 5 {
					cLog.Debug(noJobFoundMsg)
				} else {
					cLog.Info(noJobFoundMsg)
				}
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

			// Important: Set the active job to the current job for safe requeue mechanism
			state.activeJob = job

			// Run the command (potentially retrying in large mode)
			status := executor.Run(cmdCtx, job)

			state.handleJobResult(cfg, cLog, job, status)
		}
	}
}

// retryDelay returns the duration to wait before retrying to find a job in the queue.
// This avoids spamming the FS with LS queries.
func retryDelay(retryDelaysSec []int, numRetrySoFar int) <-chan time.Time {
	retryDurations := make([]time.Duration, len(retryDelaysSec))
	for i := range retryDelaysSec {
		retryDurations[i] = time.Duration(retryDelaysSec[i]) * time.Second
	}

	// By default, take the last value
	ttw := retryDurations[len(retryDurations)-1]

	// If it does not overflow, take the value at `numRetrySoFar`
	if numRetrySoFar < len(retryDurations) {
		ttw = retryDurations[numRetrySoFar]
	}

	return time.After(ttw)
}

func (s *ControllerState) notifyJobDone() {
	s.jobDoneChanMu.Lock()
	if s.jobDoneChan != nil {
		close(s.jobDoneChan)
		s.jobDoneChan = nil
	}
	s.jobDoneChanMu.Unlock()
}

// handleJobResult processes a job according to its exit status.
func (state *ControllerState) handleJobResult(cfg *config.Config, cLog *logrus.Entry, job *Job, status Status) {
	switch {
	case status.ExitCode == CodeSuccess:
		state.handleJobSuccess(cfg, cLog, job, status)

	case job.Def.Name == jobNameExecution && isIn(status.ExitCode, cfg.Controller.DeferToOtherLargeCodes):
		state.handleDeferToLarge(cLog, job, status)

	case status.ExitCode == CodeKilledByExtSig:
		state.handleJobKilledByExtSig(cfg, cLog, job)

	default:
		state.handleJobFailure(cfg, cLog, job, status)
	}
}

// handleJobSuccess moves response and in-progress files to their final locations.
func (state *ControllerState) handleJobSuccess(cfg *config.Config, cLog *logrus.Entry, job *Job, status Status) {

	if job.Def.WritesToDevNull() {
		// Jobs that target /dev/null do not produce a response file.
		cLog.Debugf("Job %s has ResponseRootDir='/dev/null', skipping response file move", job.OriginalFile)
	} else {
		respFile, err := job.ResponseFile()
		tmpRespFile := job.TmpResponseFile(cfg)
		if err != nil {
			utils.Panic("Could not generate the response file: %v (original request file: %v)", err, job.OriginalFile)
		}

		logrus.Infof("Moving the response file from the tmp response file `%v`, to the final response file: `%v`", tmpRespFile, respFile)

		if err := os.Rename(tmpRespFile, respFile); err != nil {
			// If rename fails, remove tmp file.
			_ = os.Remove(tmpRespFile)
			cLog.Errorf("Error renaming %v to %v: %v, removed the tmp file", tmpRespFile, respFile, err)
		}
	}

	// Move the input (in-progress) file to requests-done with a .success suffix.
	cLog.Infof("Moving %v to %v with the success prefix", job.OriginalFile, job.Def.dirDone())

	jobDone := job.DoneFile(status)

	// Limitless-specific renaming
	if job.Def.Name == jobNameBootstrap {
		// For bootstrap: append `-bootstrap` suffix to mark partial success.
		// This is because it is technically incorrect to consider moving the entire file move to
		// `requests-done` folder with success prefix. So we prepend`.bootstrap` suffix to indicate that
		// at this moment only bootstrap is successful. The entire file
		// should be moved only after conglomeration is successful. At this point, we should just trim the
		// bootstrap suffix to indicate the entire request is successful.
		jobDone = strings.Replace(jobDone, "-getZkProof.json."+config.SuccessSuffix, "-getZkProof.json."+config.BootstrapPartialSucessSuffix, 1)
		cLog.Info("Added suffix to indicate partial success (bootstrap phase only).")
	}

	if err := os.Rename(job.InProgressPath(), jobDone); err != nil {
		cLog.Errorf("Error renaming %v to %v: %v", job.InProgressPath(), jobDone, err)
	}

	// --- After conglomeration finishes, trim the boostrap suffix marker to indicate full success ---
	if job.Def.Name == jobNameConglomeration {
		replaceExecDoneSuffix(cfg, cLog, job, config.BootstrapPartialSucessSuffix, config.SuccessSuffix)
	}

	// Set active job to nil once the job is successful
	state.activeJob = nil
	state.notifyJobDone()
}

func replaceExecDoneSuffix(cfg *config.Config, cLog *logrus.Entry, job *Job, oldSuffix, newSuffix string) {
	pattern := filepath.Join(cfg.Execution.DirDone(), fmt.Sprintf("%d-%d-*.%s", job.Start, job.End, oldSuffix))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		cLog.Errorf("Glob pattern failed for %v: %v", pattern, err)
		return
	}

	if len(matches) == 0 {
		cLog.Warnf("No file found matching %v (maybe already moved?)", pattern)
		return
	}
	if len(matches) > 1 {
		cLog.Warnf("Multiple files match pattern %v, using first: %v", pattern, matches)
	}

	oldFile := matches[0]
	newFile := strings.TrimSuffix(oldFile, "."+oldSuffix) + "." + newSuffix

	cLog.Infof("Renaming file: %v → %v", oldFile, newFile)

	if err := os.Rename(oldFile, newFile); err != nil {
		cLog.Errorf("Failed to rename %v → %v: %v", oldFile, newFile, err)
	} else {
		cLog.Infof("Successfully replaced suffix %q → %q for %d-%d", oldSuffix, newSuffix, job.Start, job.End)
	}
}

// handleDeferToLarge defers job execution to the large prover.
func (state *ControllerState) handleDeferToLarge(cLog *logrus.Entry, job *Job, status Status) {
	cLog.Infof("Renaming %v for the large prover", job.OriginalFile)

	toLargePath, err := job.DeferToLargeFile(status)
	if err != nil {
		cLog.Errorf("error deriving the to-large-name of %v: %v", job.InProgressPath(), err)
	}

	if err := os.Rename(job.InProgressPath(), toLargePath); err != nil {
		cLog.Errorf("error renaming %v to %v: %v", job.InProgressPath(), toLargePath, err)
	}

	// From the controller perspective, it has completed its job by renaming and
	// moving it to the large prover’s queue. Setting activeJob to nil prevents
	// incorrect requeuing during shutdown, avoiding filesystem errors
	state.activeJob = nil
	state.notifyJobDone()
}

// handleJobKilledByExtSig puts the job back in the request folder and cleans up tmp files.
func (state *ControllerState) handleJobKilledByExtSig(cfg *config.Config, cLog *logrus.Entry, job *Job) {
	cLog.Infof("Job %v was killed by user signal", job.OriginalFile)
	cLog.Infof("Re-queuing the request:%s back to the request folder", job.OriginalFile)

	if err := os.Rename(job.InProgressPath(), job.OriginalPath()); err != nil {
		cLog.Errorf("Error renaming %v to %v: %v", job.InProgressPath(), job.OriginalPath(), err)
	}

	// Remove tmp-response only if we produce a response file for this job.
	if !job.Def.WritesToDevNull() {
		// Edge-case: delete tmp-response file if it exists.
		if tmp := job.TmpResponseFile(cfg); tmp != "" {
			_ = os.Remove(tmp)
		}
	} else {
		cLog.Debugf("Job %s writes to /dev/null, no tmp response to clean", job.OriginalFile)
	}
}

// handleJobFailure moves the failed job to the done directory with failure suffix.
// TODO: handle failure of the conglomeration. Rename bootstrap suffix.
func (state *ControllerState) handleJobFailure(cfg *config.Config, cLog *logrus.Entry, job *Job, status Status) {
	cLog.Infof("Moving %v with in %v with a failure suffix for code %v", job.OriginalFile, job.Def.dirDone(), status.ExitCode)

	jobFailed := job.DoneFile(status)
	if err := os.Rename(job.InProgressPath(), jobFailed); err != nil {
		cLog.Errorf("Error renaming %v to %v: %v", job.InProgressPath(), jobFailed, err)
	}

	if !job.Def.WritesToDevNull() {
		if tmp := job.TmpResponseFile(cfg); tmp != "" {
			_ = os.Remove(tmp)
		}
	} else {
		cLog.Debugf("Job %s writes to /dev/null, skipping tmp-response cleanup", job.OriginalFile)
	}

	// Upon failure for GL/LPP/Conglomeration jobs - replace the partial success bootstrap suffix
	// with the failure suffix in the cfg.Execution.DirDone path
	if job.Def.Name == jobNameGL || job.Def.Name == jobNameLPP || job.Def.Name == jobNameConglomeration {
		failSuffix := fmt.Sprintf("failure.%v_%v", config.FailSuffix, status.ExitCode)
		replaceExecDoneSuffix(cfg, cLog, job, config.BootstrapPartialSucessSuffix, failSuffix)
	}

	// Set active job to nil as there is nothing more to retry
	state.activeJob = nil
	state.notifyJobDone()
}
