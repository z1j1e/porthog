package ports

import (
	"context"

	"github.com/z1j1e/porthog/internal/core/domain"
)

// SignalPolicy controls how a process should be terminated.
type SignalPolicy struct {
	Force       bool
	ForceSystem bool
	DryRun      bool
}

// TerminateResult holds the outcome of a termination attempt.
type TerminateResult struct {
	PID       int32
	Port      uint16
	Protocol  domain.Protocol
	Process   *domain.ProcessIdentity
	Killed    bool
	DryRun    bool
	Blocked   bool
	BlockedBy string
}

// Terminator sends termination signals to processes.
type Terminator interface {
	Terminate(ctx context.Context, pid int32, policy SignalPolicy) error
}
