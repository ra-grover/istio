package queue

import "istio.io/pkg/monitoring"

// TODO: this is how workqueue from k8s project does it. We should do the same
const (
	WorkQueueSubsystem         = "workqueue"
	DepthKey                   = "depth"
	AddsKey                    = "adds_total"
	QueueLatencyKey            = "queue_duration_seconds"
	WorkDurationKey            = "work_duration_seconds"
	UnfinishedWorkKey          = "unfinished_work_seconds"
	LongestRunningProcessorKey = "longest_running_processor_seconds"
	RetriesKey                 = "retries_total"
)

var (
	nameTag = monitoring.MustCreateLabel("name")

	queue = monitoring.NewGauge(
		"workqueue_depth",
		"number of items in the queue.",
		monitoring.WithLabels(nameTag),
	)
)

func init() {
	monitoring.MustRegister(
		queue,
	)
}
