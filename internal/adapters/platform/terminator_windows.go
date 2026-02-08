//go:build windows

package platform

import (
	"context"
	"os"
	"time"

	"github.com/z1j1e/porthog/internal/core/ports"
)

type winTerminator struct{}

func NewTerminator() ports.Terminator { return &winTerminator{} }

func (t *winTerminator) Terminate(ctx context.Context, pid int32, policy ports.SignalPolicy) error {
	p, err := os.FindProcess(int(pid))
	if err != nil {
		return err
	}

	if !policy.Force {
		// Graceful: send interrupt, wait 2s
		_ = p.Signal(os.Interrupt)
		timer := time.NewTimer(2 * time.Second)
		defer timer.Stop()
		select {
		case <-timer.C:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return p.Kill()
}
