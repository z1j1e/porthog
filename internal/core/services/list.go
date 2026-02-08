package services

import (
	"context"
	"sort"

	"github.com/porthog/porthog/internal/core/domain"
	"github.com/porthog/porthog/internal/core/ports"
)

// SortField specifies which field to sort port bindings by.
type SortField int

const (
	SortByPort SortField = iota
	SortByPID
	SortByName
	SortByProtocol
)

// ListPortsService enumerates and enriches port bindings.
type ListPortsService struct {
	enumerator ports.Enumerator
	resolver   ports.ProcessResolver
}

// NewListPortsService creates a new ListPortsService.
func NewListPortsService(e ports.Enumerator, r ports.ProcessResolver) *ListPortsService {
	return &ListPortsService{enumerator: e, resolver: r}
}

// List enumerates ports, enriches with process info, and applies sorting.
func (s *ListPortsService) List(ctx context.Context, filter *domain.Filter, sortBy SortField) (*domain.PartialResult[[]domain.PortBinding], error) {
	result, err := s.enumerator.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	enriched, enrichErr := s.resolver.Enrich(ctx, result.Data)
	if enrichErr != nil {
		result.Warnings = append(result.Warnings, "process enrichment partially failed: "+enrichErr.Error())
	} else {
		result.Data = enriched
	}

	sortBindings(result.Data, sortBy)
	return result, nil
}

func sortBindings(bindings []domain.PortBinding, by SortField) {
	sort.Slice(bindings, func(i, j int) bool {
		switch by {
		case SortByPID:
			return bindings[i].PID < bindings[j].PID
		case SortByName:
			ni, nj := processName(&bindings[i]), processName(&bindings[j])
			return ni < nj
		case SortByProtocol:
			if bindings[i].Protocol != bindings[j].Protocol {
				return bindings[i].Protocol < bindings[j].Protocol
			}
			return bindings[i].LocalPort < bindings[j].LocalPort
		default: // SortByPort
			return bindings[i].LocalPort < bindings[j].LocalPort
		}
	})
}

func processName(pb *domain.PortBinding) string {
	if pb.Process != nil {
		return pb.Process.Name
	}
	return ""
}
