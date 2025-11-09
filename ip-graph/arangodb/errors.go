package arangodb

import "errors"

var (
	// ErrProcessorNotStarted indicates the processor has not been started
	ErrProcessorNotStarted = errors.New("processor not started")

	// ErrProcessorStopped indicates the processor has been stopped
	ErrProcessorStopped = errors.New("processor stopped")

	// ErrQueueFull indicates the processing queue is full
	ErrQueueFull = errors.New("processing queue is full")

	// ErrInvalidMessage indicates an invalid message was received
	ErrInvalidMessage = errors.New("invalid message")

	// ErrDatabaseConnection indicates a database connection error
	ErrDatabaseConnection = errors.New("database connection error")

	// ErrCollectionNotFound indicates a required collection was not found
	ErrCollectionNotFound = errors.New("collection not found")

	// ErrGraphNotFound indicates a required graph was not found
	ErrGraphNotFound = errors.New("graph not found")
)
