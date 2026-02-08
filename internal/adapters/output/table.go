package output

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/z1j1e/porthog/internal/core/domain"
)

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	cellStyle   = lipgloss.NewStyle()
	pidStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	protoTCP    = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("TCP")
	protoUDP    = lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Render("UDP")
)

func (r *Renderer) renderTable(bindings []domain.PortBinding) error {
	tw := r.termWidth()

	// Column definitions: name, minWidth, weight (for distributing extra space)
	type col struct {
		header string
		min    int
		weight int
	}
	cols := []col{
		{"PROTO", 5, 0},
		{"LOCAL ADDRESS", 15, 2},
		{"PID", 7, 0},
		{"PROCESS", 8, 3},
		{"USER", 8, 2},
		{"STATE", 6, 1},
	}

	// Calculate adaptive widths
	gaps := (len(cols) - 1) * 2 // 2-char gap between columns
	fixedMin := gaps
	totalWeight := 0
	for _, c := range cols {
		fixedMin += c.min
		totalWeight += c.weight
	}

	widths := make([]int, len(cols))
	extra := tw - fixedMin
	if extra < 0 {
		extra = 0
	}
	for i, c := range cols {
		widths[i] = c.min
		if totalWeight > 0 && c.weight > 0 {
			widths[i] += (extra * c.weight) / totalWeight
		}
	}

	// Narrow mode: hide USER column if terminal < 80
	showUser := tw >= 80

	// Header
	var hdr strings.Builder
	for i, c := range cols {
		if i == 4 && !showUser {
			continue
		}
		hdr.WriteString(headerStyle.Width(widths[i]).Render(c.header))
		if i < len(cols)-1 {
			hdr.WriteString("  ")
		}
	}
	fmt.Fprintln(r.w, hdr.String())

	sepWidth := tw
	if sepWidth > 120 {
		sepWidth = 120
	}
	fmt.Fprintln(r.w, strings.Repeat("─", sepWidth))

	// Rows
	for _, b := range bindings {
		proto := protoTCP
		if b.Protocol == domain.UDP {
			proto = protoUDP
		}
		addr := fmt.Sprintf("%s:%d", b.LocalIP, b.LocalPort)
		pid := pidStyle.Render(fmt.Sprintf("%d", b.PID))
		name, user := "-", "-"
		if b.Process != nil {
			if b.Process.Name != "" {
				name = b.Process.Name
			}
			if b.Process.Username != "" {
				user = b.Process.Username
			}
		}

		if showUser {
			fmt.Fprintf(r.w, "%-*s  %-*s  %-*s  %-*s  %-*s  %s\n",
				widths[0], proto,
				widths[1], truncate(addr, widths[1]),
				widths[2], pid,
				widths[3], cellStyle.Render(truncate(name, widths[3])),
				widths[4], cellStyle.Render(truncate(user, widths[4])),
				truncate(string(b.State), widths[5]))
		} else {
			fmt.Fprintf(r.w, "%-*s  %-*s  %-*s  %-*s  %s\n",
				widths[0], proto,
				widths[1], truncate(addr, widths[1]),
				widths[2], pid,
				widths[3], cellStyle.Render(truncate(name, widths[3])),
				truncate(string(b.State), widths[5]))
		}
	}
	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
