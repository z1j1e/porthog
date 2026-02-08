package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/z1j1e/porthog/internal/core/domain"
	"github.com/z1j1e/porthog/internal/core/services"
)

var (
	freeCount    int
	freeRange    string
	freeJSON     bool
)

var freeCmd = &cobra.Command{
	Use:   "free",
	Short: "Find available ports",
	RunE: func(cmd *cobra.Command, args []string) error {
		var portRange *domain.PortRange
		if freeRange != "" {
			r, err := parseRange(freeRange)
			if err != nil {
				return err
			}
			portRange = &r
		}

		svc := services.NewFindFreePortService()
		ports, err := svc.FindFree(cmd.Context(), domain.TCP, portRange, freeCount)
		if err != nil {
			return err
		}

		if freeJSON {
			return json.NewEncoder(os.Stdout).Encode(map[string]any{"ports": ports})
		}

		for _, p := range ports {
			fmt.Println(p)
		}
		return nil
	},
}

func init() {
	freeCmd.Flags().IntVarP(&freeCount, "count", "c", 1, "Number of free ports to find")
	freeCmd.Flags().StringVarP(&freeRange, "range", "r", "", "Port range (e.g., 8000-9000)")
	freeCmd.Flags().BoolVarP(&freeJSON, "json", "j", false, "Output in JSON format")
}

func parseRange(s string) (domain.PortRange, error) {
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return domain.PortRange{}, fmt.Errorf("invalid range format: %q (expected START-END)", s)
	}
	start, err := strconv.ParseUint(parts[0], 10, 16)
	if err != nil {
		return domain.PortRange{}, fmt.Errorf("invalid range start: %s", parts[0])
	}
	end, err := strconv.ParseUint(parts[1], 10, 16)
	if err != nil {
		return domain.PortRange{}, fmt.Errorf("invalid range end: %s", parts[1])
	}
	return domain.PortRange{Start: uint16(start), End: uint16(end)}, nil
}
