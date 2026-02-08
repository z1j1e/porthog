package domain

// ProcessIdentity holds metadata about a process that owns a port binding.
type ProcessIdentity struct {
	PID             int32
	CreateTimeMs    int64
	Name            string
	Exe             string
	Cmdline         string
	Username        string
	PermissionDenied bool
}

// IsEnriched returns true if process metadata was successfully resolved.
func (p *ProcessIdentity) IsEnriched() bool {
	return p.Name != "" || p.Exe != ""
}

// MatchesIdentity checks if another process identity refers to the same process
// by comparing PID and create time (TOCTOU protection).
func (p *ProcessIdentity) MatchesIdentity(other *ProcessIdentity) bool {
	if other == nil {
		return false
	}
	return p.PID == other.PID && p.CreateTimeMs == other.CreateTimeMs
}
