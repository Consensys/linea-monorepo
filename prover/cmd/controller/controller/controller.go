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

// runController runs the controller main loop.
func runController(ctx context.Context, cfg *config.Config) {
	cLog := cfg.Logger().WithField("component", "main-loop")

	fsWatcher := NewFsWatcher(cfg)
	executor := NewExecutor(cfg)
	numRetrySoFar := 0

	startMetricsServer(cfg)
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM)
	defer stop()

	cmdContext := commandContext(ctx, cfg)

	// log cancellation requests immediately when received
	go logCancellationOnSignal(ctx, cfg, cLog)

	// Main loop
	for {
		select {
		case <-ctx.Done():
			handleShutdown(ctx, cLog)
			return

		case <-retryDelay(cfg.Controller.RetryDelays, numRetrySoFar):
			job := fsWatcher.GetBest()
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

// commandContext returns the execution context for commands.
// In spot-instance mode, the command context is subordinated to the main ctx.
func commandContext(ctx context.Context, cfg *config.Config) context.Context {
	if cfg.Controller.SpotInstanceMode {
		return ctx
	}
	return context.Background()
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

// handleShutdown handles graceful shutdown of the controller.
func handleShutdown(ctx context.Context, cLog *logrus.Entry) {
	cLog.Infoln("Context canceled by caller or SIGTERM. Exiting")
	metrics.ShutdownServer(ctx)
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
		handleJobFailure(cLog, job, status)
	}
}

// handleJobSuccess moves response and in-progress files to their final locations.
func handleJobSuccess(cfg *config.Config, cLog *logrus.Entry, job *Job, status Status) {
	respFile, err := job.ResponseFile()
	tmpRespFile := job.TmpResponseFile(cfg)
	if err != nil {
		utils.Panic("Could not generate the response file: %v (original request file: %v)", err, job.OriginalFile)
	}

	logrus.Infof("Moving the response file from the tmp response file `%v`, to the final response file: `%v`", tmpRespFile, respFile)

	if err := os.Rename(tmpRespFile, respFile); err != nil {
		// If rename fails, remove tmp file.
		os.Remove(tmpRespFile)
		cLog.Errorf("Error renaming %v to %v: %v, removed the tmp file", tmpRespFile, respFile, err)
	}

	cLog.Infof("Moving %v to %v with the success prefix", job.OriginalFile, job.Def.dirDone())

	jobDone := job.DoneFile(status)
	if err := os.Rename(job.InProgressPath(), jobDone); err != nil {
		cLog.Errorf("Error renaming %v to %v: %v", job.InProgressPath(), jobDone, err)
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

	// Edge-case: delete tmp-response file if it exists.
	os.Remove(job.TmpResponseFile(cfg))
}

// handleJobFailure moves the failed job to the done directory with failure suffix.
func handleJobFailure(cLog *logrus.Entry, job *Job, status Status) {
	cLog.Infof("Moving %v with in %v with a failure suffix for code %v", job.OriginalFile, job.Def.dirDone(), status.ExitCode)

	jobFailed := job.DoneFile(status)
	if err := os.Rename(job.InProgressPath(), jobFailed); err != nil {
		cLog.Errorf("Error renaming %v to %v: %v", job.InProgressPath(), jobFailed, err)
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
