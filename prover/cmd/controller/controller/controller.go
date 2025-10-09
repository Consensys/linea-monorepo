package controller

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/consensys/linea-monorepo/prover/cmd/controller/controller/metrics"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/config/assets"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

// runController runs the controller main loop.
func runController(ctx context.Context, cfg *config.Config) {
	cLog := cfg.Logger().WithField("component", "main-loop")

	fsWatcher := NewFsWatcher(cfg)
	executor := NewExecutor(cfg)
	numRetrySoFar := 0

	startMetricsServer(cfg)
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM)
	defer stop()

	// cmdContext is the context we provide for the command execution. In
	// spot-instance mode, the context is subordinated to the ctx.
	cmdContext := context.Background()
	if cfg.Controller.SpotInstanceMode {
		cmdContext = ctx
	}

	// log cancellation requests immediately when received
	// This goroutine's raison d'etre is to log a message immediately when a
	// cancellation request (e.g., ctx expiration/cancellation, SIGTERM, etc.)
	// is received. It ensures timely logging of the request's reception,
	// which may be important for diagnostics. Without this
	// goroutine, if the prover is busy with a proof when, for example, a
	// SIGTERM is received, there would be no log entry about the signal
	// until the proof completes.
	go logCancellationOnSignal(ctx, cfg, cLog)

	// If any files are locked to RAM unlock it before exiting
	defer assets.UnlockAllLockedFiles()

	// Main loop
	for {
		select {
		case <-ctx.Done():
			// This case captures both cancellations initiated by the caller
			// through ctx and SIGTERM signals. Even if the cancellation
			// request is first intercepted by the goroutine at line 34, Go
			// allows the ctx.Done channel to be read multiple times, which, in
			// our scenario, ensures cancellation requests are effectively
			// detected and handled.
			cLog.Infoln("Context canceled by caller or SIGTERM. Exiting")
			metrics.ShutdownServer(ctx)
			return

			// Processing a new job
		case <-retryDelay(cfg.Controller.RetryDelays, numRetrySoFar):
			job := fsWatcher.GetBest()

			// No jobs, waiting a little before we retry
			if job == nil {
				numRetrySoFar++
				logNoJobFound(cLog, numRetrySoFar)
				continue
			}

			numRetrySoFar = 0
			status := executor.Run(cmdContext, job)
			handleJobResult(cfg, cLog, job, status)
		}
	}
}

// handleJobResult processes a job according to its exit status.
func handleJobResult(cfg *config.Config, cLog *logrus.Entry, job *Job, status Status) {
	switch {
	case status.ExitCode == CodeSuccess:
		handleJobSuccess(cfg, cLog, job, status)

	case job.Def.Name == jobNameExecution && isIn(status.ExitCode, cfg.Controller.DeferToOtherLargeCodes):
		handleDeferToLarge(cLog, job, status)

	case status.ExitCode == CodeKilledByUs:
		handleJobKilledByUs(cfg, cLog, job)

	default:
		handleJobFailure(cfg, cLog, job, status)
	}
}

// handleJobSuccess moves response and in-progress files to their final locations.

func handleJobSuccess(cfg *config.Config, cLog *logrus.Entry, job *Job, status Status) {

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
		jobDone = strings.Replace(jobDone, "-getZkProof.json.success", "-getZkProof.json.success.bootstrap", 1)
		cLog.Info("Added .bootstrap suffix to indicate partial success (bootstrap phase only).")
	}

	if err := os.Rename(job.InProgressPath(), jobDone); err != nil {
		cLog.Errorf("Error renaming %v to %v: %v", job.InProgressPath(), jobDone, err)
	}

	// --- After conglomeration finishes, trim the boostrap suffix marker to indicate full success ---
	if job.Def.Name == jobNameConglomeration {
		trimBootstrapSuffix(cfg, cLog, job)
	}
}

func trimBootstrapSuffix(cfg *config.Config, cLog *logrus.Entry, job *Job) {
	pattern := filepath.Join(cfg.Execution.DirDone(), fmt.Sprintf("%d-%d-*.success.bootstrap", job.Start, job.End))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		cLog.Errorf("Glob pattern failed for %v: %v", pattern, err)
		return
	}

	if len(matches) == 0 {
		cLog.Warnf("No bootstrap success file found matching %v (maybe already moved?)", pattern)
		return
	}
	if len(matches) > 1 {
		cLog.Warnf("Multiple bootstrap success files match %v, using first: %v", pattern, matches)
	}

	bootstrapFile := matches[0]
	finalFile := strings.TrimSuffix(bootstrapFile, ".bootstrap")

	cLog.Infof("Renaming bootstrap success file → final success file: %v → %v", bootstrapFile, finalFile)

	if err := os.Rename(bootstrapFile, finalFile); err != nil {
		cLog.Errorf("Failed to rename %v → %v: %v", bootstrapFile, finalFile, err)
	} else {
		cLog.Infof("Successfully finalized conglomeration for %d-%d", job.Start, job.End)
	}
}

// handleDeferToLarge defers job execution to the large prover.
func handleDeferToLarge(cLog *logrus.Entry, job *Job, status Status) {
	cLog.Infof("Renaming %v for the large prover", job.OriginalFile)

	toLargePath, err := job.DeferToLargeFile(status)
	if err != nil {
		cLog.Errorf("error deriving the to-large-name of %v: %v", job.InProgressPath(), err)
	}

	if err := os.Rename(job.InProgressPath(), toLargePath); err != nil {
		cLog.Errorf("error renaming %v to %v: %v", job.InProgressPath(), toLargePath, err)
	}
}

// handleJobKilledByUs puts the job back in the request folder and cleans up tmp files.
func handleJobKilledByUs(cfg *config.Config, cLog *logrus.Entry, job *Job) {
	cLog.Infof("Job %v was killed by us. Reputting it in the request folder", job.OriginalFile)

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
func handleJobFailure(cfg *config.Config, cLog *logrus.Entry, job *Job, status Status) {
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

// startMetricsServer starts Prometheus metrics server if enabled.
func startMetricsServer(cfg *config.Config) {
	if cfg.Controller.Prometheus.Enabled {
		metrics.StartServer(
			cfg.Controller.LocalID,
			cfg.Controller.Prometheus.Route,
			cfg.Controller.Prometheus.Port,
		)
	}
}

// logCancellationOnSignal logs immediately when a cancellation request is received.
func logCancellationOnSignal(ctx context.Context, cfg *config.Config, cLog *logrus.Entry) {
	<-ctx.Done()
	if cfg.Controller.SpotInstanceMode {
		cLog.Infoln("Received cancellation request. Killing the ongoing process and exiting immediately after.")
	} else {
		cLog.Infoln("Received cancellation request, will exit as soon as possible or once current proof task is complete.")
	}
}

// logNoJobFound logs when no job is found, with reduced verbosity after 5 retries.
func logNoJobFound(cLog *logrus.Entry, numRetrySoFar int) {
	msg := "found no jobs in the queue"
	if numRetrySoFar > 5 {
		cLog.Debug(msg)
	} else {
		cLog.Info(msg)
	}
}
