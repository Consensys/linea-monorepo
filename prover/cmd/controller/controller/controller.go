package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/cmd/controller/controller/metrics"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/config/assets"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

type ControllerState struct {
	activeJob   *Job
	activeJobMu sync.Mutex

	jobDoneChan   chan struct{}
	jobDoneChanMu sync.Mutex

	// per-job cancel so SIGUSR2 cancells only the running child process
	currentJobCancel context.CancelFunc
	cancelJobMu      sync.Mutex
}

var (
	// Atomic coordination flags
	spotReclaimDetected       atomic.Bool
	gracefulShutdownRequested atomic.Bool
	peerAbortDetected         atomic.Bool
)

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

		// Channel to receive signals
		signalChan = make(chan os.Signal, 2)
	)

	// Start the metric server
	if cfg.Controller.Prometheus.Enabled {
		metrics.StartServer(
			cfg.Controller.LocalID,
			cfg.Controller.Prometheus.Route,
			cfg.Controller.Prometheus.Port,
		)
	}

	// Derive the controller context with a ctrlCancel function from the parent ctx
	ctrlCtx, ctrlCancel := context.WithCancel(ctx)
	defer ctrlCancel()

	// Graceful shutdown => Only SIGTERM
	// Spot-reclaim => SIGUSR1 and SIGTERM
	// Peer Abort   => SIGUSR2 and SIGTERM
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)
	defer signal.Stop(signalChan)

	// Spin up the signal handler go routine to handle SIGTERM and SIGUSR1 signals
	go state.handleSignals(ctrlCancel, cLog, cfg, signalChan)

	// Spin up the log on context done go routine for prompt logging and diagnostics
	go state.logOnCtxDone(ctrlCtx, cLog, cfg)

	// If any files are locked to RAM unlock it before exiting
	defer assets.UnlockAllLockedFiles()

	// Before entering the main loop, check for any previous dangling `.inprogress` files belonging to this
	// controller (LocalID). This usually happens if the prover pod OOMs or crashes hard abruptly without a
	// chance to cleanup. We rename these to `.large` so they are picked up by a larger prover instance or retried
	// with higher limits.
	rmPrevDanglingFiles(cfg, cLog)
	for {
		select {
		case <-ctrlCtx.Done():
			cLog.Infoln("Context cancelled by caller or externally triggered. EXITING!!!")
			metrics.ShutdownServer(ctrlCtx)

			if spotReclaimDetected.Load() {
				cLog.Infof("Waiting up to %ds for spot reclaim...", cfg.Controller.SpotInstanceReclaimTime)
				time.Sleep(time.Duration(cfg.Controller.SpotInstanceReclaimTime) * time.Second)
			}

			// For graceful shutdown, do not sleep here and requeue only if
			// the job is still active (timer already slept in signal handler) - See signal handler go routine
			if !spotReclaimDetected.Load() && gracefulShutdownRequested.Load() {
				if job := state.getActiveJob(); job != nil {
					cLog.Infof("Job %v did not finish before termination grace period. REQUEUING..", job.OriginalFile)
					// Write transient failures for limitless jobs
					if err := state.writeTransientFailFile(cfg, cLog, CodeKilledByExtSig); err != nil {
						utils.Panic("error while writing transient failure files:%v", err)
					}
					state.requeueJob(cfg, cLog)
				}
			}

			return

			// Processing a new job
		case <-retryDelay(cfg.Controller.RetryDelays, numRetrySoFar):
			// Prevent starting new jobs if shutdown started
			if ctrlCtx.Err() != nil {
				continue
			}

			// Skip fetching new jobs if shutdown is in progress
			if gracefulShutdownRequested.Load() || spotReclaimDetected.Load() {
				numRetrySoFar++
				logRetryMessage(cLog, "Shutdown in progress, skipping new jobs", numRetrySoFar)
				continue
			}

			// Fetch the best job available
			job := fsWatcher.GetBest()
			if job == nil {
				numRetrySoFar++
				logRetryMessage(cLog, "Found no jobs in the queue", numRetrySoFar)
				continue
			}

			// Create a per-job context derived from controller context
			jobCtx, jobCancel := context.WithCancel(ctrlCtx)
			state.setCurrentJobCancel(jobCancel)

			// Reset retry counter and claim the job
			numRetrySoFar = 0

			// Claim active job for safe requeue mechanism
			state.setActiveJob(job, cLog)

			// Rm any prev. transient failure files and other intermediate files - relevant only for Limitless jobs
			if err := state.preCleanForLimitlessJob(cfg, cLog); err != nil {
				cLog.Error(err)
				state.clearActiveJob()
				jobCancel()
				state.clearCurrentJobCancel()
				continue
			}

			if err := state.preCheckForLimitlessJob(cfg, cLog); err != nil {
				cLog.Error(err)
				state.clearActiveJob()
				jobCancel()
				state.clearCurrentJobCancel()
				continue
			}

			// If a shared-failure marker exists for this job, skip it and cleanup
			// Relevant only for limitless controller
			if state.shouldSkipDueToSharedFail(cfg, cLog, job) {
				state.clearActiveJob()
				jobCancel()
				state.clearCurrentJobCancel()
				continue
			}

			// Reset peer-abort flag now that we are about to start this job
			// This ensures stale abort signals from earlier jobs do not misclassify this one.
			peerAbortDetected.Store(false)

			// Run the command (potentially retrying in large mode) using jobCtx so it can be
			// cancelled independently by SIGUSR2/peer abort.
			status := executor.Run(jobCtx, job)

			// Job finished (clear per-job cancel)
			state.clearCurrentJobCancel()

			// Continue with existing result handling
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

func (st *ControllerState) notifyJobDone() {
	st.jobDoneChanMu.Lock()
	if st.jobDoneChan != nil {
		close(st.jobDoneChan)
		st.jobDoneChan = nil
	}
	st.jobDoneChanMu.Unlock()
}

func (st *ControllerState) requeueJob(cfg *config.Config, cLog *logrus.Entry) {
	job := st.getActiveJob()
	if job == nil {
		return
	}

	_ = os.Remove(job.TmpResponseFile(cfg))
	if err := os.Rename(job.InProgressPath(), job.OriginalPath()); err != nil {
		cLog.Errorf("Failed to requeue job %v: %v", job.InProgressPath(), err)
	} else {
		cLog.Infof("REQUEUED job: %v", job.OriginalFile)
	}
	// Edge case during conglomeration requeueing:
	// Any leftover file in the execution request directory must also be handled.
	if job.Def.Name == jobNameConglomeration {
		// Also, Rename the *inprogress file in execution request dir back to partial bootstrap suffix marker
		// so that the next conglomerator can correctly pick it up appending its local id to it
		baseFile, oldFile, err := files.LocateBaseBySuffix(job.Start, job.End, cfg.Execution.DirFrom(), config.InProgressSuffix)
		if err != nil {
			cLog.Errorf("error locating base file during requeue for conglomeration job: %v", err)
		}

		newFile := baseFile + "." + config.BootstrapPartialSucessSuffix
		if err := os.Rename(oldFile, newFile); err != nil {
			cLog.Errorf("Failed to rename %v → %v: %v", oldFile, newFile, err)
		} else {
			cLog.Infof("Successfully renamed %v to %v", oldFile, newFile)
		}
	}
	st.clearActiveJob()
	st.notifyJobDone()
}

func (st *ControllerState) handleSignals(ctrlCancel context.CancelFunc, cLog *logrus.Entry, cfg *config.Config, signalChan <-chan os.Signal) {

	for sig := range signalChan {
		switch sig {
		case syscall.SIGUSR1:
			spotReclaimDetected.Store(true)
			if !gracefulShutdownRequested.Load() {
				cLog.Infof("Received SIGUSR1: marking spot reclaim detected, cancelling context ASAP (max %ds)...", cfg.Controller.SpotInstanceReclaimTime)
				ctrlCancel()
				if cfg.Controller.Prometheus.Enabled {
					metrics.IncSpotInterruption(cfg.Controller.LocalID)
				}
			} else {
				cLog.Info("Received SIGUSR1 after SIGTERM: marking spot reclaim detected (shutdown already in progress)")
			}

			// Requeue active job immediately
			st.requeueJob(cfg, cLog)

		case syscall.SIGTERM:
			gracefulShutdownRequested.Store(true)
			if !spotReclaimDetected.Load() {
				cLog.Info("Received SIGTERM: graceful shutdown requested")
			}

			if job := st.getActiveJob(); job != nil {
				st.jobDoneChanMu.Lock()
				if st.jobDoneChan == nil {
					st.jobDoneChan = make(chan struct{})
				}
				localJobDoneChan := st.jobDoneChan
				st.jobDoneChanMu.Unlock()

				jobName := job.OriginalFile

				go func() {
					cLog.Infof("Allowing in-flight job %s to finish (max %ds)...", jobName, cfg.Controller.TerminationGracePeriod)
					select {
					case <-time.After(time.Duration(cfg.Controller.TerminationGracePeriod) * time.Second):
						cLog.Info("Termination grace period expired. Cancelling context to force shutdown...")
						ctrlCancel()
					case <-localJobDoneChan:
						cLog.Info("Job finished before grace period expired. Cancelling context to exit immediately...")
						ctrlCancel()
					}
				}()
			} else {
				cLog.Info("No active job during SIGTERM, cancelling context to exit immediately")
				ctrlCancel()
			}

		case syscall.SIGUSR2:
			cLog.Warn("Received SIGUSR2 (Abort-By-Peer): cancelling the current child job only")
			peerAbortDetected.Store(true)

			// Cancel only the current job's context so executor.Run returns a CodePeerAbortReceived.
			// IMPORTANT: DONT call the parent ctrlCancel() here; that would stop the whole controller.
			st.cancelCurrentJob()
		}
	}
}

// handleJobResult processes a job according to its exit status.
func (st *ControllerState) handleJobResult(cfg *config.Config, cLog *logrus.Entry, job *Job, status Status) {

	switch {
	case status.ExitCode == CodeSuccess:
		st.handleJobSuccess(cfg, cLog, job, status)

	case job.Def.Name == jobNameExecution && isIn(status.ExitCode, cfg.Controller.DeferToOtherLargeCodes):
		st.handleDeferToLarge(cLog, job, status)

	case status.ExitCode == CodeKilledByExtSig:
		st.handleJobTerminatedByExtSig(cfg, cLog, job)

	default:
		st.handleJobFailure(cfg, cLog, job, status)
	}
}

// handleJobSuccess moves response and in-progress files to their final locations.
func (st *ControllerState) handleJobSuccess(cfg *config.Config, cLog *logrus.Entry, job *Job, status Status) {

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

	// ADD THIS BLOCK: Delete trace file after successful execution proof
	if job.Def.Name == jobNameExecution && job.End > cfg.Execution.KeepTracesUntilBlock {
		tryDeleteExecution(cLog, job, cfg)
	}

	// --- After conglomeration finishes, trim the boostrap suffix marker to indicate full success ---
	if job.Def.Name == jobNameConglomeration {
		finalizeExecLimitlessStatus(cfg, cLog, job, config.SuccessSuffix)
		defer rmTmpArtificats(cLog, job)

		if job.End > cfg.Execution.KeepTracesUntilBlock {
			tryDeleteExecution(cLog, job, cfg)
		}
	}

	// Limitless-specific renaming
	if job.Def.Name == jobNameBootstrap {
		// For bootstrap: append `bootstrap.partial.success` suffix to mark partial success.
		// This is because it is technically incorrect to consider moving the entire file move to
		// `requests-done` folder with success prefix. So we append `bootstrap.partial.success` suffix to indicate that
		// at this moment only bootstrap is successful. The entire file should be moved only after
		// conglomeration is successful.
		var (
			current = job.InProgressPath()
			target  = filepath.Join(filepath.Dir(job.InProgressPath()),
				job.OriginalFile+"."+config.BootstrapPartialSucessSuffix,
			)
		)

		if err := os.Rename(current, target); err != nil {
			cLog.Errorf("Error renaming bootstrap job %v → %v: %v", current, target, err)
		}
		cLog.Infof("Bootstrap completed. Staying in requests and marked partial success: %v → %v", current, target)
	} else {
		// Move the input (in-progress) file to requests-done with a .success suffix.
		cLog.Infof("Moving %v to %v with the success prefix", job.OriginalFile, job.Def.dirDone())
		jobDone := job.DoneFile(status)
		if err := os.Rename(job.InProgressPath(), jobDone); err != nil {
			cLog.Errorf("Error renaming %v to %v: %v", job.InProgressPath(), jobDone, err)
		}
	}
	// Set active job to nil once the job is successful and notify the job is done
	st.clearActiveJob()
	st.notifyJobDone()
}

// handleDeferToLarge defers job execution to the large prover.
func (st *ControllerState) handleDeferToLarge(cLog *logrus.Entry, job *Job, status Status) {
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
	st.clearActiveJob()
	st.notifyJobDone()
}

// handleJobTerminatedByExtSig puts the job back in the request folder and cleans up tmp files.
func (st *ControllerState) handleJobTerminatedByExtSig(cfg *config.Config, cLog *logrus.Entry, job *Job) {
	cLog.Infof("Job %v was terminated by user signal", job.OriginalFile)

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
func (st *ControllerState) handleJobFailure(cfg *config.Config, cLog *logrus.Entry, job *Job, status Status) {
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
	if strings.HasPrefix(job.Def.Name, jobNameGL) || strings.HasPrefix(job.Def.Name, jobNameLPP) || job.Def.Name == jobNameConglomeration {
		failSuffix := fmt.Sprintf("failure.%v_%v", config.FailSuffix, status.ExitCode)
		finalizeExecLimitlessStatus(cfg, cLog, job, failSuffix)
	}

	// Write transient failure for limitless jobs - only genuine failures are written here
	// and not due to PeerAbortion
	if status.ExitCode != CodePeerAbortReceived {
		if err := st.writeTransientFailFile(cfg, cLog, status.ExitCode); err != nil {
			cLog.Errorf("error writing failure marker: %v", err)
		}
	} else {
		cLog.Infof("Safely ignoring to write transient failure file for job:%s with status exit code:%d", job.OriginalFile, CodePeerAbortReceived)
	}

	// Set active job to nil as there is nothing more to retry and notify the job is done
	st.clearActiveJob()
	st.notifyJobDone()
}

// writeTransientFailFile: writes a failure marker to the shared failure directory for distributed
// error propogation in the context of external signal or generic job failure.
// This is relevant only for limitless jobs
func (st *ControllerState) writeTransientFailFile(cfg *config.Config, cLog *logrus.Entry, exitCode int) error {

	job := st.getActiveJob()
	if job == nil {
		return nil
	}

	// Relevant only for limitless jobs (excl. conglomeration - final stage)
	if !isExecLimitlessJob(job) || job.Def.Name == jobNameConglomeration {
		return nil
	}

	// For worker jobs - interrupted through SIGUSR1/SIGTERM, do not write peer abort files
	// This is because, if one of the worker was spot-reclaimed/ did not terminate within grace period
	// there is no reason to abort other workers. We propogate peer-abort only if there is a genuine failure
	// in processing any of the worker sub-proofs
	if (strings.HasPrefix(job.Def.Name, jobNameGL) || strings.HasPrefix(job.Def.Name, jobNameLPP)) &&
		exitCode == CodeKilledByExtSig {
		logrus.Infof("Safely ignoring to write peer-abort failure failure for worker:%s terminated via external signal(exitcode:%d)", fLocalID, CodeKilledByExtSig)
		return nil
	}

	var (
		sharedFailFileName = fmt.Sprintf("%d-%d-%s-failure_code%d", job.Start, job.End, job.Def.Name, exitCode)
		sharedFailPath     = filepath.Join(cfg.ExecutionLimitless.SharedFailureDir, sharedFailFileName)
	)

	if err := os.WriteFile(sharedFailPath, []byte{}, 0600); err != nil {
		cLog.Errorf("%s could not create failure marker %s: %v", fLocalID, sharedFailPath, err)
		return err
	} else {
		cLog.Warnf("%s wrote abort-by-peer failure marker: %s", fLocalID, sharedFailPath)
	}

	return nil
}

// preCleanForLimitlessJob: clean up any transient failure files for limitless jobs
// Only the Bootstrap job can preCleanForLimitlessJob
func (st *ControllerState) preCleanForLimitlessJob(cfg *config.Config, cLog *logrus.Entry) error {

	job := st.getActiveJob()
	if job == nil {
		return nil
	}

	// Relevant only for bootstrap job
	if !(job.Def.Name == jobNameBootstrap) {
		return nil
	}

	// Remove any transient failure files if present at the start before executing
	// the job to remove stale failure files and prevent false-negatives

	var (
		sharedFilePattern = fmt.Sprintf("%d-%d-*-failure_code*", job.Start, job.End)
		exists, err       = files.RemoveMatchingFiles(filepath.Join(cfg.ExecutionLimitless.SharedFailureDir, sharedFilePattern), true)
	)

	// If failure files indeed exist, we need to enforce limitless specific job logic to
	// clean up any  dangling tmp artifact files removed here (not inside the child process)
	// and, before the start of the next job
	// If not, then for example, dangling witness files could have already been picked
	// up by other active workers.
	if exists {
		// Get all possible dir paths where all tmp artificats for the job can be created
		var (
			rmPattern = fmt.Sprintf("%d-%d-*", job.Start, job.End)
		)
		for _, dir := range limitlessDirs {
			_, err := files.RemoveMatchingFiles(filepath.Join(dir, rmPattern), true)
			if err != nil {
				return fmt.Errorf("error removing dangling witness pattern in prev. interrupted bootstrap job:%v", err)
			}
		}
	}

	cLog.Infof("pre-cleaned all tmp artifacts for job:%s conflation:%v-%v", job.Def.Name, job.Start, job.End)
	return err
}

func (st *ControllerState) preCheckForLimitlessJob(cfg *config.Config, cLog *logrus.Entry) error {
	job := st.getActiveJob()
	if job == nil {
		return nil
	}

	if !isExecLimitlessJob(job) || job.Def.Name == jobNameBootstrap {
		return nil
	}

	pattern := fmt.Sprintf("%d-%d-*", job.Start, job.End)
	// For worker or conglomeration jobs - check if the original execution request file still exists in the request/ dir
	exists, err := files.HasWildcardMatch(cfg.Execution.DirFrom(), pattern)
	if !exists {
		return fmt.Errorf("could not find original execution request file in request dir %s for conflation:%d-%d during job:%s", cfg.Execution.DirFrom(), job.Start, job.End, job.Def.Name)
	}

	cLog.Infof("Validated preCheck conditions for limitless job: %s are met before starting", job.Def.Name)
	return err
}

// Only for limitless job: Returns true if a shared-failure marker exists for this job and so we should skip picking it up.
// Allows job of its own type to proceed to clean-up and futher execution.
func (st *ControllerState) shouldSkipDueToSharedFail(cfg *config.Config, cLog *logrus.Entry, job *Job) bool {
	// Only relevant for limitless jobs
	if !isExecLimitlessJob(job) {
		return false
	}

	// Match any job type for the same range (not just this job)
	pattern := fmt.Sprintf("%d-%d-*-failure_code*", job.Start, job.End)
	globPattern := filepath.Join(cfg.ExecutionLimitless.SharedFailureDir, pattern)

	matches, err := filepath.Glob(globPattern)
	if err != nil {
		cLog.Errorf("could not glob shared failure marker for pattern %s: %v", globPattern, err)
		// safer to skip this job so we don't execute when filesystem is weird
		return true
	}

	if len(matches) == 0 {
		return false
	}

	// Check if any marker belongs to this job’s own type (e.g. bootstrap)
	// Meaning we dont want an instance to write the transient failure file (to
	// propogate peer-abort (SIGUSR2) for other peer workers)
	// and then log receiving SIGUSR2 on the same source instance
	for _, m := range matches {
		base := filepath.Base(m)
		if strings.Contains(base, fmt.Sprintf("-%s-failure_code", job.Def.Name)) {
			cLog.Infof("Job %s matches its own failure marker (%s); skipping", job.OriginalFile, base)
			return true
		}
	}

	// Otherwise skip since another job already marked this range failed
	cLog.Warnf("Skipping job %s because shared failure marker exists: %v", job.OriginalFile, matches)
	return true
}

func (st *ControllerState) logOnCtxDone(cmdCtx context.Context, cLog *logrus.Entry, cfg *config.Config) {
	<-cmdCtx.Done()
	switch {
	case spotReclaimDetected.Load():
		cLog.Infof("Context done due to spot reclaim. Aborting ASAP (max %ds)...", cfg.Controller.SpotInstanceReclaimTime)

	case gracefulShutdownRequested.Load():
		if st.getActiveJob() != nil {
			cLog.Infof("Context done: grace period expired, forcing shutdown")
		} else {
			cLog.Info("Context done: graceful shutdown complete")
		}

	case peerAbortDetected.Load():
		cLog.Infoln("Context cancellation triggered due to peer-aborted service")

	default:
		cLog.Infoln("Context cancelled due to unknown reason.")
	}
}

// finalizeExecLimitlessStatus replaces the suffix of the execution limitless job file with the new suffix
// relevant only for limitless job
func finalizeExecLimitlessStatus(cfg *config.Config, cLog *logrus.Entry, job *Job, newSuffix string) {

	base, oldFile, err := files.LocateBaseBySuffix(job.Start, job.End, cfg.Execution.DirFrom(), config.InProgressSuffix)
	if err != nil {
		cLog.Errorf("failed to find base file: %v", err)
		return
	}
	var (
		newFileName = filepath.Base(base) + "." + newSuffix
		dir         = filepath.Dir(base)
		doneDir     = strings.Replace(dir, cfg.Execution.DirFrom(), cfg.Execution.DirDone(), 1) // requests -> requests-done
		newFile     = filepath.Join(doneDir, newFileName)
	)
	if err := os.Rename(oldFile, newFile); err != nil {
		cLog.Errorf("Failed to rename %v → %v: %v", oldFile, newFile, err)
	} else {
		cLog.Infof("Successfully replaced %s suffix with status:%s for conflation %d-%d", config.InProgressSuffix, newSuffix, job.Start, job.End)
	}
}

func logRetryMessage(cLog *logrus.Entry, msg string, numRetry int) {
	if numRetry > 5 {
		cLog.Debug(msg)
	} else {
		cLog.Info(msg)
	}
}

// tryDeleteExecution attemps to delete the execution trace file to free up some
// space. It will log a warning if it fails to delete and keep going if it fails
// to do so.
//
// The function expects to be provided an execution job or it will panic.
func tryDeleteExecution(cLog *logrus.Entry, job *Job, cfg *config.Config) {

	if job.Def.Name != jobNameExecution {
		panic("expected an execution job")
	}

	req := &execution.Request{}
	f, err := os.Open(job.InProgressPath())
	if err != nil {
		// This is sort of unexpected. It could happen if the underlying prover binary
		// could delete or move the request folder. If it fails open, the file then
		// we cannot delete the trace file and we simply continue as if nothing happened
		// on the normal flow.
		cLog.Errorf("could not open file: %s", err.Error())
		return
	}

	defer f.Close()

	if err := json.NewDecoder(f).Decode(req); err != nil {
		cLog.Errorf("could not decode input file: %s", err.Error())
		return
	}

	// Get the trace file and delete it. It is gauranteed that trace file will exist
	// If otherwise, the prover would panic at the beginning.
	traceFile := req.ConflatedExecTraceFilepath(cfg.Execution.ConflatedTracesDir)
	if err := os.Remove(traceFile); err != nil {
		cLog.Errorf("could not remove trace file: %s", err.Error())
		return
	}

	cLog.Infof("Deleted trace file: %s after successful execution proof request", traceFile)
}

// rmPrevDanglingFiles scans the execution request directory for any prev `.inprogress` files
// matching the current controller's LocalID. If found (meaning the previous
// run crashed), it renames them to .large to allow reprocessing.
func rmPrevDanglingFiles(cfg *config.Config, cLog *logrus.Entry) {
	logrus.Infof("Checking for any dangling in-progress jobs from previous run...")
	reqDir := cfg.Execution.DirFrom()
	if reqDir == "" {
		cLog.Error("Empty request directory for stale jobs cleanup")
		return
	}

	entries, err := os.ReadDir(reqDir)
	if err != nil {
		cLog.Errorf("Failed to read directory for stale jobs cleanup: %v", err)
		return
	}

	// Suffix format: .inprogress.<LocalID>
	suffix := "." + config.InProgressSuffix + "." + cfg.Controller.LocalID

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(name, suffix) {
			oldPath := filepath.Join(reqDir, name)

			// Remove the .inprogress.<ID> suffix and append .large
			// Example: req.json.inprogress.ID -> req.json.large
			baseName := strings.TrimSuffix(name, suffix)
			newName := baseName + "." + config.LargeSuffix
			newPath := filepath.Join(reqDir, newName)

			cLog.Warnf("Found dangling in-progress job under this controller's local-id from previous run (possibly OOM/Crash): %s. Recovering by renaming to %s", name, newName)
			if err := os.Rename(oldPath, newPath); err != nil {
				cLog.Errorf("Failed to recover stale job %s: %v", name, err)
			}
		}
	}
}
