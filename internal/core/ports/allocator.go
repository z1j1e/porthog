package ports

import (
	"context"

	"github.com/z1j1e/porthog/internal/core/domain"
)

// PortAllocator finds available (unbound) ports.
type PortAllocator interface {
	FindFree(ctx context.Context, protocol domain.Protocol, portRange *domain.PortRange, count int) ([]uint16, error)
}
