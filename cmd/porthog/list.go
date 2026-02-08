package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/z1j1e/porthog/internal/adapters/output"
	"github.com/z1j1e/porthog/internal/adapters/platform"
	"github.com/z1j1e/porthog/internal/adapters/process"
	"github.com/z1j1e/porthog/internal/core/domain"
	"github.com/z1j1e/porthog/internal/core/services"
)

var (
	listJSON bool
	listTCP  bool
	listUDP  bool
	listSort string
)

var listCmd = &cobra.Command{
	Use:   "list [port]",
	Short: "List listening ports with process info",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filter := buildListFilter(args)
		sortBy := parseSortField(listSort)

		enum := platform.NewEnumerator()
		resolver := process.NewResolver()
		svc := services.NewListPortsService(enum, resolver)

		result, err := svc.List(cmd.Context(), filter, sortBy)
		if err != nil {
			return err
		}

		format := output.FormatAuto
		if listJSON {
			format = output.FormatJSON
		}
		renderer := output.NewRenderer(os.Stdout, format)
		return renderer.Render(result, "list")
	},
}

func init() {
	listCmd.Flags().BoolVarP(&listJSON, "json", "j", false, "Output in JSON format")
	listCmd.Flags().BoolVar(&listTCP, "tcp", false, "Show only TCP ports")
	listCmd.Flags().BoolVar(&listUDP, "udp", false, "Show only UDP ports")
	listCmd.Flags().StringVar(&listSort, "sort", "port", "Sort by: port, pid, name, protocol")
}

func buildListFilter(args []string) *domain.Filter {
	f := &domain.Filter{}
	if listTCP && !listUDP {
		f.Protocols = []domain.Protocol{domain.TCP}
	} else if listUDP && !listTCP {
		f.Protocols = []domain.Protocol{domain.UDP}
	}
	if len(args) > 0 {
		if port, err := strconv.ParseUint(args[0], 10, 16); err == nil {
			f.Ports = []uint16{uint16(port)}
		} else {
			fmt.Fprintf(os.Stderr, "Warning: invalid port %q, ignoring\n", args[0])
		}
	}
	return f
}

func parseSortField(s string) services.SortField {
	switch s {
	case "pid":
		return services.SortByPID
	case "name":
		return services.SortByName
	case "protocol":
		return services.SortByProtocol
	default:
		return services.SortByPort
	}
}
