package services

import (
	"context"
	"fmt"
	"net"

	"github.com/z1j1e/porthog/internal/core/domain"
)

const (
	defaultRangeStart uint16 = 1024
	defaultRangeEnd   uint16 = 65535
)

// FindFreePortService discovers available ports by attempting to bind.
type FindFreePortService struct{}

// NewFindFreePortService creates a new FindFreePortService.
func NewFindFreePortService() *FindFreePortService {
	return &FindFreePortService{}
}

// FindFree finds available ports in the given range by bind-checking.
func (s *FindFreePortService) FindFree(ctx context.Context, proto domain.Protocol, portRange *domain.PortRange, count int) ([]uint16, error) {
	if count <= 0 {
		count = 1
	}

	r := domain.PortRange{Start: defaultRangeStart, End: defaultRangeEnd}
	if portRange != nil {
		if !portRange.Valid() {
			return nil, domain.ErrInvalidRange
		}
		r = *portRange
	}

	var found []uint16
	for port := r.Start; port <= r.End && len(found) < count; port++ {
		select {
		case <-ctx.Done():
			return found, ctx.Err()
		default:
		}
		if isPortFree(proto, port) {
			found = append(found, port)
		}
	}

	if len(found) == 0 {
		return nil, fmt.Errorf("%w: range %d-%d", domain.ErrNoFreePort, r.Start, r.End)
	}
	return found, nil
}

func isPortFree(proto domain.Protocol, port uint16) bool {
	network := "tcp"
	if proto == domain.UDP {
		network = "udp"
	}
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen(network, addr)
	if err != nil {
		return false
	}
	ln.Close()
	return true
}
