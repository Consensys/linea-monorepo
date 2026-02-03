package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Labels for the metrics
const (
	metricNamespace = "prover"
	metricSubsystem = "controller"
	labelExitCode   = "status"
	labelJobType    = "job_type"
	labelWorkerID   = "worker_id"
)

// global registry of metrics
var globalRegistry *Registry
var onceRegistry sync.Once

// Initialize the global registry of metrics
func initRegistry(worker_id string) {

	onceRegistry.Do(func() {
		globalRegistry = &Registry{

			NumProcessed: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace:   metricNamespace,
					Subsystem:   metricSubsystem,
					ConstLabels: map[string]string{labelWorkerID: worker_id},
					Name:        "processed_jobs_count",
					Help: "Count the number of processed jobs by the prover." +
						" Broken down by job types and status",
				},
				[]string{labelExitCode, labelJobType},
			),

			IsActive: promauto.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace:   metricNamespace,
					Subsystem:   metricSubsystem,
					ConstLabels: map[string]string{labelWorkerID: worker_id},
					Name:        "processing_jobs_count",
					Help: "One when the prover is processing jobs, and zero" +
						" when the prover is idle.",
				},
				[]string{labelJobType},
			),

			ProcessingTime: promauto.NewSummaryVec(
				prometheus.SummaryOpts{
					Namespace:   metricNamespace,
					Subsystem:   metricSubsystem,
					ConstLabels: map[string]string{labelWorkerID: worker_id},
					Name:        "processing_job_time_seconds",
					Help: "Returns the processing time of the jobs that we measure" +
						" over time.",
				},
				// The segmentation per exit code is important to not fix failed
				// jobs with successful ones.
				[]string{labelExitCode, labelJobType},
			),

			JobRangeStartAt: promauto.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace:   metricNamespace,
					Subsystem:   metricSubsystem,
					ConstLabels: map[string]string{labelWorkerID: worker_id},
					Name:        "job_range_start_at_l2_block",
					Help:        "Returns the beginning of the range of the job that we process",
				},
				[]string{labelJobType},
			),

			JobRangeEndAt: promauto.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace:   metricNamespace,
					Subsystem:   metricSubsystem,
					ConstLabels: map[string]string{labelWorkerID: worker_id},
					Name:        "job_range_end_at_l2_block",
					Help:        "Returns the end of the range of the job that we process",
				},
				[]string{labelJobType},
			),

			NumFilesInQueue: promauto.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: metricNamespace,
					Subsystem: metricSubsystem,
					// NB: The worker ID is important because it will allow us to
					// separate the large provers and the medium ones. This will
					// both return different results because they use different
					// filters to identify jobs.
					ConstLabels: map[string]string{labelWorkerID: worker_id},
					Name:        "num_files_in_queue",
					Help:        "Number of files in the queue. Segmented by type of job",
				},
				[]string{labelJobType},
			),

			NumEntriesInDirectory: promauto.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace:   metricNamespace,
					Subsystem:   metricSubsystem,
					ConstLabels: map[string]string{labelWorkerID: worker_id},
					Name:        "num_entries_in_request_directory",
					Help:        "Number of files in the request directories. Segmented by type of job",
				},
				[]string{labelJobType},
			),
		}
	})
}

// Registry maintains the metrics of the controller service of the prover. It
// exposes the following metrics:
type Registry struct {

	// Total number of processed (finished) requests. Labeled by type of jobs
	// and
	NumProcessed *prometheus.CounterVec

	// Total number of processing (including the current one) requests. This
	// metric can be used to derive the activity of the worker.
	IsActive *prometheus.GaugeVec

	// The time it took to complete the job
	ProcessingTime *prometheus.SummaryVec

	// The height at which the jobs are read. For instance, for a conflation
	// these will give respectively the initial and the final l2 block of the
	// conflation. They can be used to compute the range of the jobs also so
	// for instance : the number of conflated blocks etc...
	JobRangeStartAt *prometheus.GaugeVec
	JobRangeEndAt   *prometheus.GaugeVec

	// The span of the job (i.e)
	NumFilesInQueue       *prometheus.GaugeVec
	NumEntriesInDirectory *prometheus.GaugeVec
}
