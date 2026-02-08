package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/z1j1e/porthog/internal/adapters/platform"
	"github.com/z1j1e/porthog/internal/adapters/process"
	"github.com/z1j1e/porthog/internal/core/services"
	"github.com/z1j1e/porthog/internal/tui/watch"
)

var watchInterval time.Duration

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Real-time port monitoring TUI",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !isatty.IsTerminal(os.Stdout.Fd()) {
			return fmt.Errorf("watch requires a TTY; use --ci-snapshot for non-interactive mode")
		}

		enum := platform.NewEnumerator()
		resolver := process.NewResolver()
		svc := services.NewWatchPortsService(enum, resolver)

		model := watch.New(svc, nil, watchInterval)
		p := tea.NewProgram(model, tea.WithAltScreen())
		_, err := p.Run()
		return err
	},
}

func init() {
	watchCmd.Flags().DurationVar(&watchInterval, "interval", 1*time.Second, "Refresh interval")
}
