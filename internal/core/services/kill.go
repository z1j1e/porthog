package services

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/porthog/porthog/internal/core/domain"
	"github.com/porthog/porthog/internal/core/ports"
)

var criticalPIDs = map[int32]bool{0: true, 1: true}

var criticalNames = map[string]bool{
	"systemd":   true,
	"launchd":   true,
	"init":      true,
	"csrss.exe": true,
	"smss.exe":  true,
	"wininit.exe": true,
}

func init() {
	if runtime.GOOS == "windows" {
		criticalPIDs[4] = true // System process on Windows
	}
}

// KillByPortService terminates the process occupying a given port
// with TOCTOU protection and critical process safeguards.
type KillByPortService struct {
	enumerator ports.Enumerator
	resolver   ports.ProcessResolver
	terminator ports.Terminator
}

// NewKillByPortService creates a new KillByPortService.
func NewKillByPortService(e ports.Enumerator, r ports.ProcessResolver, t ports.Terminator) *KillByPortService {
	return &KillByPortService{enumerator: e, resolver: r, terminator: t}
}

// Kill terminates the process on the specified port with two-phase TOCTOU validation.
func (s *KillByPortService) Kill(ctx context.Context, port uint16, proto domain.Protocol, policy ports.SignalPolicy) (*ports.TerminateResult, error) {
	// Phase 1: Enumerate and identify target
	filter := &domain.Filter{
		Ports:     []uint16{port},
		Protocols: []domain.Protocol{proto},
		States:    []domain.SocketState{domain.StateListen},
	}
	result, err := s.enumerator.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("%w: no process found on %s port %d", domain.ErrNotFound, proto, port)
	}

	target := result.Data[0]
	enriched, enrichErr := s.resolver.Enrich(ctx, []domain.PortBinding{target})
	if enrichErr != nil {
		return nil, fmt.Errorf("cannot safely identify target process: %w", enrichErr)
	}
	if len(enriched) > 0 {
		target = enriched[0]
	}

	res := &ports.TerminateResult{
		PID:      target.PID,
		Port:     port,
		Protocol: proto,
		Process:  target.Process,
	}

	// Check critical process protection
	if isCritical(target.PID, target.Process) && !policy.ForceSystem {
		res.Blocked = true
		res.BlockedBy = "critical system process"
		return res, domain.ErrCriticalProcess
	}

	if policy.DryRun {
		res.DryRun = true
		return res, nil
	}

	// Phase 2: Revalidate before kill (TOCTOU protection)
	// Invalidate cache to force fresh process identity lookup
	s.resolver.InvalidatePID(target.PID)

	recheck, err := s.enumerator.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(recheck.Data) == 0 {
		return nil, domain.ErrProcessExited
	}
	if recheck.Data[0].PID != target.PID {
		return nil, fmt.Errorf("%w: PID changed from %d to %d", domain.ErrOwnershipConflict, target.PID, recheck.Data[0].PID)
	}

	// Validate create_time if available (guards against PID reuse)
	if target.Process != nil && target.Process.CreateTimeMs > 0 {
		recheckEnriched, err := s.resolver.Enrich(ctx, []domain.PortBinding{recheck.Data[0]})
		if err != nil || len(recheckEnriched) == 0 || recheckEnriched[0].Process == nil {
			return nil, fmt.Errorf("cannot revalidate process identity before termination: %w", domain.ErrOwnershipConflict)
		}
		if !target.Process.MatchesIdentity(recheckEnriched[0].Process) {
			return nil, fmt.Errorf("%w: process identity changed (PID reuse detected)", domain.ErrOwnershipConflict)
		}
	}

	// Execute termination
	if err := s.terminator.Terminate(ctx, target.PID, policy); err != nil {
		return nil, err
	}

	res.Killed = true
	return res, nil
}

func isCritical(pid int32, proc *domain.ProcessIdentity) bool {
	if criticalPIDs[pid] {
		return true
	}
	if proc != nil && criticalNames[strings.ToLower(proc.Name)] {
		return true
	}
	return false
}
