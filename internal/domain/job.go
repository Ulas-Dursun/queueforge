package domain

import (
	"time"

	"github.com/google/uuid"
)

// JobStatus represents the lifecycle state of a job.
type JobStatus string

const (
	StatusQueued     JobStatus = "QUEUED"
	StatusProcessing JobStatus = "PROCESSING"
	StatusCompleted  JobStatus = "COMPLETED"
	StatusFailed     JobStatus = "FAILED"
)

// Job is the core domain entity.
type Job struct {
	ID         uuid.UUID `json:"id"`
	Type       string    `json:"type"`
	Payload    string    `json:"payload"` // JSON string: {"url": "https://..."}
	Status     JobStatus `json:"status"`
	Result     *string   `json:"result"` // JSON string, null until completed
	ErrorMsg   *string   `json:"error"`  // null unless failed
	Attempts   int       `json:"attempts"`
	MaxRetries int       `json:"max_retries"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CanRetry returns true if the job has remaining retry attempts.
func (j *Job) CanRetry() bool {
	return j.Attempts < j.MaxRetries
}
