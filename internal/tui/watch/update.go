package watch

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/porthog/porthog/internal/core/domain"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tickMsg:
		return m, m.fetchSnapshot
	case snapshotMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.bindings = filterSearch(msg.snapshot.Bindings, m.search.Value())
		}
		return m, tickCmd(m.interval)
	}

	if m.searching {
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.searching {
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			m.searching = false
			m.search.Blur()
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			m.searching = false
			m.search.Blur()
			return m, nil
		}
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		return m, cmd
	}

	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
		m.quitting = true
		return m, tea.Quit
	case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
		if m.cursor < len(m.bindings)-1 {
			m.cursor++
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("/"))):
		m.searching = true
		m.search.Focus()
		return m, nil
	}
	return m, nil
}

func filterSearch(bindings []domain.PortBinding, query string) []domain.PortBinding {
	if query == "" {
		return bindings
	}
	q := strings.ToLower(query)
	var filtered []domain.PortBinding
	for _, b := range bindings {
		name := ""
		if b.Process != nil {
			name = b.Process.Name
		}
		if strings.Contains(strings.ToLower(name), q) ||
			strings.Contains(b.LocalIP.String(), q) {
			filtered = append(filtered, b)
		}
	}
	return filtered
}
