package ports

import (
	"context"

	"github.com/z1j1e/porthog/internal/core/domain"
)

// Enumerator lists network port bindings from the operating system.
type Enumerator interface {
	List(ctx context.Context, filter *domain.Filter) (*domain.PartialResult[[]domain.PortBinding], error)
}
