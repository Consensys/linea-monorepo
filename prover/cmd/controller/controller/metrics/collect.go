package metrics

import (
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// Collect metrics relative to the queue
func CollectFS(jobType string, numDirEntries, numMatched int) {

	if globalRegistry == nil {
		logrus.Tracef("No global registry found, not collecting")
		return
	}

	globalRegistry.NumEntriesInDirectory.
		With(jobLab(jobType)).
		Set(float64(numDirEntries))

	globalRegistry.NumFilesInQueue.
		With(jobLab(jobType)).
		Set(float64(numMatched))
}

// Collect metrics relative to a job we are about to run. Retry means that
// this is a file that we are retrying locally.
func CollectPreProcess(jobType string, start, end int, retry bool) {

	if globalRegistry == nil {
		logrus.Tracef("No global registry found, not collecting")
		return
	}

	if retry {
		// alter the job type so that we do not "pollute" non retried metric
		jobType = strings.Join(
			[]string{jobType, "retry", "large", "locally"},
			"_",
		)
	}

	globalRegistry.JobRangeStartAt.
		With(jobLab(jobType)).
		Set(float64(start))

	globalRegistry.JobRangeEndAt.
		With(jobLab(jobType)).
		Set(float64(end))

	globalRegistry.IsActive.
		With(jobLab(jobType)).
		Set(1)
}

// Collect metrics from jobs for which the processing has completed
func CollectPostProcess(jobType string, code int, t time.Duration, retry bool) {

	if globalRegistry == nil {
		logrus.Tracef("No global registry found, not collecting")
		return
	}

	if retry {
		// alter the job type so that we do not "pollute" non retried metric
		jobType = strings.Join(
			[]string{jobType, "retry", "large", "locally"},
			"_",
		)
	}

	globalRegistry.IsActive.
		With(jobLab(jobType)).
		Set(0)

	globalRegistry.NumProcessed.
		With(jobAndStatusLabs(jobType, code)).
		Inc()

	globalRegistry.ProcessingTime.
		With(jobAndStatusLabs(jobType, code)).
		Observe(t.Seconds())
}

// helper function that returns a label map for some job type
func jobLab(jobType string) prometheus.Labels {
	return prometheus.Labels{
		labelJobType: jobType,
	}
}

// helper function that returns a label map for some job type and exit code
func jobAndStatusLabs(jobType string, code int) prometheus.Labels {
	return prometheus.Labels{
		labelExitCode: strconv.Itoa(code),
		labelJobType:  jobType,
	}
}
