package ports

import (
	"context"

	"github.com/porthog/porthog/internal/core/domain"
)

// ProcessResolver enriches port bindings with process metadata.
type ProcessResolver interface {
	Enrich(ctx context.Context, bindings []domain.PortBinding) ([]domain.PortBinding, error)
	// InvalidatePID removes a PID from the cache, forcing fresh lookup on next Enrich.
	InvalidatePID(pid int32)
}
