package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/z1j1e/porthog/internal/adapters/platform"
	"github.com/z1j1e/porthog/internal/adapters/process"
	"github.com/z1j1e/porthog/internal/core/domain"
	"github.com/z1j1e/porthog/internal/core/ports"
	"github.com/z1j1e/porthog/internal/core/services"
)

var (
	killForce       bool
	killDryRun      bool
	killForceSystem bool
)

var killCmd = &cobra.Command{
	Use:   "kill <port>",
	Short: "Kill the process occupying a port",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		port, err := strconv.ParseUint(args[0], 10, 16)
		if err != nil {
			return fmt.Errorf("invalid port: %s", args[0])
		}

		enum := platform.NewEnumerator()
		resolver := process.NewResolver()
		term := platform.NewTerminator()
		svc := services.NewKillByPortService(enum, resolver, term)

		policy := ports.SignalPolicy{
			Force:       killForce,
			ForceSystem: killForceSystem,
			DryRun:      killDryRun,
		}

		result, err := svc.Kill(cmd.Context(), uint16(port), domain.TCP, policy)
		if err != nil {
			return err
		}

		if result.DryRun {
			name := "unknown"
			if result.Process != nil && result.Process.Name != "" {
				name = result.Process.Name
			}
			fmt.Fprintf(os.Stdout, "[dry-run] Would kill PID %d (%s) on port %d\n", result.PID, name, result.Port)
			return nil
		}

		if result.Killed {
			fmt.Fprintf(os.Stdout, "Killed PID %d on port %d\n", result.PID, result.Port)
		}
		return nil
	},
}

func init() {
	killCmd.Flags().BoolVarP(&killForce, "force", "f", false, "Force kill (SIGKILL/TerminateProcess)")
	killCmd.Flags().BoolVar(&killDryRun, "dry-run", false, "Show what would be killed without acting")
	killCmd.Flags().BoolVar(&killForceSystem, "force-system", false, "Allow killing critical system processes")
}
