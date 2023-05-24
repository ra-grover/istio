package queue

import "istio.io/pkg/monitoring"

// Const strings for label value.
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
		"queue",
		"number of items in the queue.",
		monitoring.WithLabels(nameTag),
	)
)

func init() {
	monitoring.MustRegister(
		queue,
	)
}
