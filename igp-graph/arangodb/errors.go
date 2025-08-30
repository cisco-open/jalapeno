package arangodb

import "errors"

// Common errors used throughout the IGP graph processor
var (
	ErrProcessorNotStarted = errors.New("processor not started")
	ErrProcessorStopped    = errors.New("processor stopped")
	ErrQueueFull           = errors.New("operation queue is full")
	ErrInvalidOperation    = errors.New("invalid operation")
	ErrNodeNotFound        = errors.New("node not found")
	ErrLinkNotFound        = errors.New("link not found")
	ErrGraphNotFound       = errors.New("graph not found")
)
