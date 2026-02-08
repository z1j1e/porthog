package integration_test

import (
	"context"
	"net"
	"strconv"
	"testing"

	"github.com/porthog/porthog/internal/adapters/platform"
	"github.com/porthog/porthog/internal/adapters/process"
	"github.com/porthog/porthog/internal/core/domain"
	"github.com/porthog/porthog/internal/core/services"
)

func TestListFindsEphemeralTCPListener(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	_, portStr, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.ParseUint(portStr, 10, 16)

	enum := platform.NewEnumerator()
	resolver := process.NewResolver()
	svc := services.NewListPortsService(enum, resolver)

	filter := &domain.Filter{Ports: []uint16{uint16(port)}, Protocols: []domain.Protocol{domain.TCP}}
	result, err := svc.List(context.Background(), filter, services.SortByPort)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Data) == 0 {
		t.Fatalf("expected to find TCP listener on port %d", port)
	}
	if result.Data[0].LocalPort != uint16(port) {
		t.Errorf("expected port %d, got %d", port, result.Data[0].LocalPort)
	}
	if result.Data[0].PID <= 0 {
		t.Error("expected positive PID")
	}
}

func TestFreePortIsActuallyFree(t *testing.T) {
	svc := services.NewFindFreePortService()
	ports, err := svc.FindFree(context.Background(), domain.TCP, &domain.PortRange{Start: 10000, End: 20000}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(ports) != 1 {
		t.Fatalf("expected 1 port, got %d", len(ports))
	}

	// Verify we can actually bind to it
	ln, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(int(ports[0])))
	if err != nil {
		t.Fatalf("returned port %d is not actually free: %v", ports[0], err)
	}
	ln.Close()
}
