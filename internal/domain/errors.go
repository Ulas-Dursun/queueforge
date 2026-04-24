package domain

import "errors"

var (
	ErrJobNotFound    = errors.New("job not found")
	ErrQueuePush      = errors.New("failed to push job to queue")
	ErrInvalidPayload = errors.New("invalid job payload")
	ErrInvalidJobType = errors.New("invalid job type")
)
