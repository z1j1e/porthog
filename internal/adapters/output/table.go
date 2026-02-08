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
	headers := []string{"PROTO", "LOCAL ADDRESS", "PID", "PROCESS", "USER", "STATE"}
	widths := []int{5, 22, 8, 20, 15, 12}

	// Header
	var hdr strings.Builder
	for i, h := range headers {
		hdr.WriteString(headerStyle.Width(widths[i]).Render(h))
		hdr.WriteString("  ")
	}
	fmt.Fprintln(r.w, hdr.String())
	fmt.Fprintln(r.w, strings.Repeat("─", 90))

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

		fmt.Fprintf(r.w, "%-7s %-24s %-8s %-22s %-17s %s\n",
			proto, addr, pid, cellStyle.Render(truncate(name, 20)),
			cellStyle.Render(truncate(user, 15)), b.State)
	}
	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
