package process

import (
	"context"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/process"

	"github.com/z1j1e/porthog/internal/core/domain"
)

type cacheEntry struct {
	identity  domain.ProcessIdentity
	expiresAt time.Time
}

// Resolver enriches port bindings with process metadata via gopsutil.
type Resolver struct {
	mu    sync.RWMutex
	cache map[int32]*cacheEntry
	ttl   time.Duration
}

func NewResolver() *Resolver {
	return &Resolver{
		cache: make(map[int32]*cacheEntry),
		ttl:   5 * time.Second,
	}
}

func (r *Resolver) Enrich(ctx context.Context, bindings []domain.PortBinding) ([]domain.PortBinding, error) {
	pids := uniquePIDs(bindings)
	identities := r.batchResolve(ctx, pids)

	for i := range bindings {
		if id, ok := identities[bindings[i].PID]; ok {
			cp := id
			bindings[i].Process = &cp
		}
	}
	return bindings, nil
}

func (r *Resolver) batchResolve(ctx context.Context, pids []int32) map[int32]domain.ProcessIdentity {
	result := make(map[int32]domain.ProcessIdentity, len(pids))
	for _, pid := range pids {
		if cached := r.getCache(pid); cached != nil {
			result[pid] = *cached
			continue
		}

		select {
		case <-ctx.Done():
			return result
		default:
		}

		id := resolveOne(pid)
		result[pid] = id
		r.setCache(pid, id)
	}
	return result
}

func resolveOne(pid int32) domain.ProcessIdentity {
	id := domain.ProcessIdentity{PID: pid}
	p, err := process.NewProcess(pid)
	if err != nil {
		id.PermissionDenied = true
		return id
	}

	if ct, err := p.CreateTime(); err == nil {
		id.CreateTimeMs = ct
	}
	if name, err := p.Name(); err == nil {
		id.Name = name
	}
	if exe, err := p.Exe(); err == nil {
		id.Exe = exe
	}
	if user, err := p.Username(); err == nil {
		id.Username = user
	}
	// Cmdline is expensive — lazy enrichment: only populate on explicit request
	return id
}

func (r *Resolver) getCache(pid int32) *domain.ProcessIdentity {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if e, ok := r.cache[pid]; ok && time.Now().Before(e.expiresAt) {
		return &e.identity
	}
	return nil
}

func (r *Resolver) setCache(pid int32, id domain.ProcessIdentity) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache[pid] = &cacheEntry{identity: id, expiresAt: time.Now().Add(r.ttl)}
}

// InvalidatePID removes a PID from the cache, forcing a fresh lookup on next Enrich.
func (r *Resolver) InvalidatePID(pid int32) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.cache, pid)
}

func uniquePIDs(bindings []domain.PortBinding) []int32 {
	seen := make(map[int32]bool)
	var pids []int32
	for _, b := range bindings {
		if b.PID > 0 && !seen[b.PID] {
			seen[b.PID] = true
			pids = append(pids, b.PID)
		}
	}
	return pids
}
