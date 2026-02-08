//go:build linux || darwin

package platform

import (
	"context"
	"os"
	"syscall"
	"time"

	"github.com/z1j1e/porthog/internal/core/ports"
)

type unixTerminator struct{}

func NewTerminator() ports.Terminator { return &unixTerminator{} }

func (t *unixTerminator) Terminate(ctx context.Context, pid int32, policy ports.SignalPolicy) error {
	p, err := os.FindProcess(int(pid))
	if err != nil {
		return err
	}

	if policy.Force {
		return p.Signal(syscall.SIGKILL)
	}

	// Graceful: SIGTERM, wait 2s, then SIGKILL
	if err := p.Signal(syscall.SIGTERM); err != nil {
		return err
	}

	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-ctx.Done():
		return ctx.Err()
	}

	return p.Signal(syscall.SIGKILL)
}
