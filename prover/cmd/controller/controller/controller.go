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

// runController: Runs the controller with the given config
func runController(ctx context.Context, cfg *config.Config) {
	var (
		cLog          = cfg.Logger().WithField("component", "main-loop")
		fsWatcher     = NewFsWatcher(cfg)
		executor      = NewExecutor(cfg)
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

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM)
	defer stop()

	// This goroutine's raison d'etre is to log a message immediately when a
	// cancellation request (e.g., ctx expiration/cancellation, SIGTERM, etc.)
	// is received. It ensures timely logging of the request's reception,
	// which may be important for diagnostics. Without this
	// goroutine, if the prover is busy with a proof when, for example, a
	// SIGTERM is received, there would be no log entry about the signal
	// until the proof completes.
	go func() {
		<-ctx.Done()
		cLog.Infoln("Received cancellation request, will exit as soon as possible or once current proof task is complete.")
	}()

	for {
		select {
		// Graceful shutdown.
		// This case captures both cancellations initiated by the caller
		// through ctx and SIGTERM signals. Even if the cancellation
		// request is first intercepted by the goroutine at line 34, Go
		// allows the ctx.Done channel to be read multiple times, which, in
		// our scenario, ensures cancellation requests are effectively
		// detected and handled.
		case <-ctx.Done():
			cLog.Infoln("Context canceled by caller or SIGTERM. Exiting")
			metrics.ShutdownServer(ctx)
			return

			// Processing a new job
		case <-retryDelay(cfg.Controller.RetryDelays, numRetrySoFar):
			job := fsWatcher.GetBest()
			if job == nil {
				numRetrySoFar++
				noJobFoundMsg := "found no jobs in the queue"
				if numRetrySoFar > 5 {
					cLog.Debugf("%v", noJobFoundMsg)
				} else {
					cLog.Infof("%v", noJobFoundMsg)
				}
				continue
			}

			numRetrySoFar = 0
			status := executor.Run(job)
			switch {
			// Success
			case status.ExitCode == CodeSuccess:
				handleSuccess(job, cfg, status, cLog)
			// Defer to the large prover
			case job.Def.Name == jobNameExecution && isIn(status.ExitCode, cfg.Controller.DeferToOtherLargeCodes):
				handleDeferToLarge(job, status, cLog)
			// Failure case
			default:
				handleFailure(job, status, cLog)
			}
		}
	}
}

func handleSuccess(job *Job, cfg *config.Config, status Status, cLog *logrus.Entry) {
	// NB: we already check that the response filename can be
	// generated prior to running the command. So this actually
	// will not panic.
	for opIdx := range job.Def.ResponsesRootDir {
		respFile, err := job.ResponseFile(opIdx)
		tmpRespFile := job.TmpResponseFile(cfg, opIdx)
		if err != nil {
			utils.Panic("Could not generate the response file: %v (original request file: %v)", err, job.OriginalFile[opIdx])
		}

		logrus.Infof("Moving the response file from the tmp response file `%v`, to the final response file: `%v`", tmpRespFile, respFile)

		// @Alex: it is unclear how the rename operation could fail
		// here. If this happens, we prefer removing the tmp file.
		// Note that the operation is an `mv -f`
		if err := os.Rename(tmpRespFile, respFile); err != nil {
			os.Remove(tmpRespFile)
			cLog.Errorf("Error renaming %v to %v: %v, removed the tmp file", tmpRespFile, respFile, err)
		}
	}

	// Move the inprogress to the done directory
	for ipIdx := range job.OriginalFile {
		cLog.Infof("Moving %v to %v with the success prefix", job.OriginalFile[ipIdx], job.Def.dirDone(ipIdx))
		jobDone := job.DoneFile(status, ipIdx)
		if err := os.Rename(job.InProgressPath(ipIdx), jobDone); err != nil {
			cLog.Errorf("Error renaming %v to %v: %v", job.InProgressPath(ipIdx), jobDone, err)
		}
	}
}

func handleDeferToLarge(job *Job, status Status, cLog *logrus.Entry) {
	for ipIdx := range job.OriginalFile {
		cLog.Infof("Renaming %v for the large prover", job.OriginalFile[ipIdx])
		// Move the inprogress file back in the from directory with the new suffix
		toLargePath, err := job.DeferToLargeFile(status, ipIdx)
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
			cLog.Errorf("error deriving the to-large-name of %v: %v", job.InProgressPath(ipIdx), err)
		}

		if err := os.Rename(job.InProgressPath(ipIdx), toLargePath); err != nil {
			// When that happens, the only thing left to do is to log
			// the error and let the inprogress file where it is. It
			// will likely require a human intervention.
			cLog.Errorf("error renaming %v to %v: %v", job.InProgressPath(ipIdx), toLargePath, err)
		}
	}
}

func handleFailure(job *Job, status Status, cLog *logrus.Entry) {
	for ipIdx := range job.OriginalFile {
		// Move the inprogress to the done directory
		cLog.Infof("Moving %v with in %v with a failure suffix for code %v", job.OriginalFile[ipIdx], job.Def.dirDone(ipIdx), status.ExitCode)
		jobFailed := job.DoneFile(status, ipIdx)
		if err := os.Rename(job.InProgressPath(ipIdx), jobFailed); err != nil {
			// When that happens, the only thing left to do is to log
			// the error and let the inprogress file where it is. It
			// will likely require a human intervention.
			cLog.Errorf("Error renaming %v to %v: %v", job.InProgressPath(ipIdx), jobFailed, err)
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
