package ports

import (
	"context"

	"github.com/z1j1e/porthog/internal/core/domain"
)

// Snapshot represents a point-in-time view of all port bindings.
type Snapshot struct {
	Bindings []domain.PortBinding
	Partial  bool
}

// Watcher provides periodic snapshots of port bindings for watch mode.
type Watcher interface {
	Snapshot(ctx context.Context, filter *domain.Filter) (*Snapshot, error)
}
