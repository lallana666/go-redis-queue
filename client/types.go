package client

import "time"

type Job struct {
	ID     string
	Status JobStatus
	Spec   JobSpec
}

type JobStatus struct {
	Phase          JobPhase
	StartTime      *time.Time
	CompletionTime *time.Time
}

type JobSpec struct {
	Payload   string
	QueueName string
	Batch     string
}

// JobPhase is a label for the condition of a job at the current time.
type JobPhase string

// These are the valid statuses of jobs.
const (
	// JobPending means the job has been accepted by the system, but the command has not been started.
	JobPending JobPhase = "pending"
	// JobRunning means the job has been bound to a worker and the command have been started.
	JobRunning JobPhase = "running"
	// JobSucceeded means that the command have voluntarily terminated with exit code of 0.
	JobSucceeded JobPhase = "done"
	// JobFailed means that the command have terminated, and was terminated in a failure (exited with
	// a non-zero exit code or was stopped by the system).
	JobFailed JobPhase = "failed"
)

// redis 任务计数器
type BatchCount struct {
	Pending *string `json:"pending"`
	Running *string `json:"running"`
	Done    *string `json:"done"`
	Failed  *string `json:"failed"`
}
