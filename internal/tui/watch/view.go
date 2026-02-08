package watch

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).MarginBottom(1)
	rowStyle    = lipgloss.NewStyle()
	selStyle    = lipgloss.NewStyle().Background(lipgloss.Color("236")).Bold(true)
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("🐷 porthog watch"))
	b.WriteString("\n")

	if m.searching {
		b.WriteString("/ " + m.search.View() + "\n")
	}

	// Header
	hdr := fmt.Sprintf("%-5s %-22s %-8s %-20s %-12s", "PROTO", "LOCAL ADDRESS", "PID", "PROCESS", "STATE")
	b.WriteString(lipgloss.NewStyle().Bold(true).Render(hdr))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 70))
	b.WriteString("\n")

	// Rows
	visible := m.height - 6
	if visible < 5 {
		visible = 20
	}
	start := 0
	if m.cursor >= visible {
		start = m.cursor - visible + 1
	}

	for i := start; i < len(m.bindings) && i < start+visible; i++ {
		pb := m.bindings[i]
		name := "-"
		if pb.Process != nil && pb.Process.Name != "" {
			name = pb.Process.Name
		}
		row := fmt.Sprintf("%-5s %-22s %-8d %-20s %-12s",
			pb.Protocol, fmt.Sprintf("%s:%d", pb.LocalIP, pb.LocalPort),
			pb.PID, name, pb.State)

		if i == m.cursor {
			b.WriteString(selStyle.Render(row))
		} else {
			b.WriteString(rowStyle.Render(row))
		}
		b.WriteString("\n")
	}

	if m.err != nil {
		b.WriteString(errStyle.Render("Error: " + m.err.Error()))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓ navigate • / search • q quit"))

	return b.String()
}
