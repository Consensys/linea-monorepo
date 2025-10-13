package controller

import (
	"sync/atomic"
	"time"
)

// ControllerState tracks the current state of the controller for
// graceful shutdown and health/readiness probes
type ControllerState struct {
	// isProcessing indicates if a job is currently being processed
	isProcessing atomic.Bool

	// shutdownRequested indicates if SIGTERM has been received
	shutdownRequested atomic.Bool

	// isSpotReclaim indicates if shutdown is due to spot instance reclamation
	isSpotReclaim atomic.Bool

	// shutdownDeadline is when we must forcefully exit
	shutdownDeadline atomic.Value // time.Time

	// currentJob tracks the current job being processed
	currentJob atomic.Value // *Job
}

// NewControllerState creates a new ControllerState instance
func NewControllerState() *ControllerState {
	return &ControllerState{}
}

// StartProcessing marks that a job has started processing
func (s *ControllerState) StartProcessing(job *Job) {
	s.isProcessing.Store(true)
	s.currentJob.Store(job)
}

// StopProcessing marks that job processing has finished
func (s *ControllerState) StopProcessing() {
	s.isProcessing.Store(false)
	s.currentJob.Store((*Job)(nil))
}

// IsProcessing returns true if currently processing a job
func (s *ControllerState) IsProcessing() bool {
	return s.isProcessing.Load()
}

// RequestShutdown marks that shutdown has been requested and sets the deadline
func (s *ControllerState) RequestShutdown(deadline time.Time) {
	s.shutdownRequested.Store(true)
	s.shutdownDeadline.Store(deadline)
}

// RequestSpotReclaimShutdown marks that shutdown is due to spot instance reclamation
func (s *ControllerState) RequestSpotReclaimShutdown(deadline time.Time) {
	s.shutdownRequested.Store(true)
	s.isSpotReclaim.Store(true)
	s.shutdownDeadline.Store(deadline)
}

// IsShutdownRequested returns true if shutdown has been requested
func (s *ControllerState) IsShutdownRequested() bool {
	return s.shutdownRequested.Load()
}

// IsSpotReclaim returns true if shutdown is due to spot instance reclamation
func (s *ControllerState) IsSpotReclaim() bool {
	return s.isSpotReclaim.Load()
}

// GetShutdownDeadline returns the shutdown deadline
func (s *ControllerState) GetShutdownDeadline() time.Time {
	v := s.shutdownDeadline.Load()
	if v == nil {
		return time.Time{}
	}
	return v.(time.Time)
}

// GetCurrentJob returns the currently processing job, or nil if no job is processing
func (s *ControllerState) GetCurrentJob() *Job {
	v := s.currentJob.Load()
	if v == nil {
		return nil
	}
	return v.(*Job)
}
