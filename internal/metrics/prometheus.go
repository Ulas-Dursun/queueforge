package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	JobsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "queueforge_jobs_total",
			Help: "Total number of jobs by status",
		},
		[]string{"status"},
	)

	JobDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "queueforge_job_duration_seconds",
			Help:    "Job processing duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	QueueSize = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "queueforge_queue_size",
			Help: "Current number of jobs waiting in the queue",
		},
	)

	ActiveWorkers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "queueforge_active_workers",
			Help: "Number of active worker goroutines",
		},
	)
)

func Register() {
	prometheus.MustRegister(JobsTotal)
	prometheus.MustRegister(JobDuration)
	prometheus.MustRegister(QueueSize)
	prometheus.MustRegister(ActiveWorkers)
}
