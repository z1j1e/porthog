package services

import (
	"context"

	"github.com/porthog/porthog/internal/core/domain"
	"github.com/porthog/porthog/internal/core/ports"
)

// DiffEntry represents a change between two snapshots.
type DiffEntry struct {
	Type    DiffType
	Binding domain.PortBinding
}

// DiffType classifies a snapshot change.
type DiffType int

const (
	DiffAdded DiffType = iota
	DiffRemoved
)

// WatchPortsService provides snapshot and diff capabilities for watch mode.
type WatchPortsService struct {
	enumerator ports.Enumerator
	resolver   ports.ProcessResolver
}

// NewWatchPortsService creates a new WatchPortsService.
func NewWatchPortsService(e ports.Enumerator, r ports.ProcessResolver) *WatchPortsService {
	return &WatchPortsService{enumerator: e, resolver: r}
}

// Snapshot returns the current state of all port bindings.
func (s *WatchPortsService) Snapshot(ctx context.Context, filter *domain.Filter) (*ports.Snapshot, error) {
	result, err := s.enumerator.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	enriched, _ := s.resolver.Enrich(ctx, result.Data)
	if enriched != nil {
		result.Data = enriched
	}
	return &ports.Snapshot{Bindings: result.Data, Partial: result.Partial}, nil
}

// Diff computes the difference between two snapshots.
func Diff(prev, curr *ports.Snapshot) []DiffEntry {
	prevMap := bindingKey(prev.Bindings)
	currMap := bindingKey(curr.Bindings)

	var diffs []DiffEntry
	for key, b := range currMap {
		if _, exists := prevMap[key]; !exists {
			diffs = append(diffs, DiffEntry{Type: DiffAdded, Binding: b})
		}
	}
	for key, b := range prevMap {
		if _, exists := currMap[key]; !exists {
			diffs = append(diffs, DiffEntry{Type: DiffRemoved, Binding: b})
		}
	}
	return diffs
}

type portKey struct {
	proto domain.Protocol
	port  uint16
	pid   int32
}

func bindingKey(bindings []domain.PortBinding) map[portKey]domain.PortBinding {
	m := make(map[portKey]domain.PortBinding, len(bindings))
	for _, b := range bindings {
		k := portKey{proto: b.Protocol, port: b.LocalPort, pid: b.PID}
		m[k] = b
	}
	return m
}
