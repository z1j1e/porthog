package watch

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"

	"github.com/z1j1e/porthog/internal/core/domain"
	"github.com/z1j1e/porthog/internal/core/ports"
	"github.com/z1j1e/porthog/internal/core/services"
)

type tickMsg time.Time

type snapshotMsg struct {
	snapshot *ports.Snapshot
	err      error
}

// Model is the Bubbletea model for watch mode.
type Model struct {
	svc      *services.WatchPortsService
	filter   *domain.Filter
	interval time.Duration

	bindings []domain.PortBinding
	cursor   int
	search   textinput.Model
	searching bool
	width    int
	height   int
	err      error
	quitting bool
}

// New creates a new watch TUI model.
func New(svc *services.WatchPortsService, filter *domain.Filter, interval time.Duration) Model {
	ti := textinput.New()
	ti.Placeholder = "filter..."
	ti.CharLimit = 64
	return Model{
		svc: svc, filter: filter, interval: interval,
		search: ti,
	}
}
