package service

import (
	"context"
	"time"
)

// ReadinessChecker checks whether a required dependency is available.
type ReadinessChecker interface {
	Ping(context.Context) error
}

// Health reports process liveness and dependency readiness.
type Health struct {
	checker ReadinessChecker
	timeout time.Duration
}

// NewHealth creates a health service with a bounded readiness check.
func NewHealth(checker ReadinessChecker, timeout time.Duration) *Health {
	return &Health{checker: checker, timeout: timeout}
}

// Live reports whether this process is running.
func (h *Health) Live() bool {
	return true
}

// Ready reports whether required dependencies are reachable before timeout.
func (h *Health) Ready(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()
	return h.checker.Ping(ctx) == nil
}
