//go:build windows

package windows

import (
	"context"
	"net"
	"strconv"
	"testing"

	"github.com/z1j1e/porthog/internal/core/domain"
)

func TestWindowsEnumerator_ListReturnsValidBindings(t *testing.T) {
	// Spawn an ephemeral TCP listener to guarantee at least one result
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	_, portStr, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.ParseUint(portStr, 10, 16)

	enum := NewEnumerator()
	result, err := enum.List(context.Background(), &domain.Filter{
		Ports:     []uint16{uint16(port)},
		Protocols: []domain.Protocol{domain.TCP},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Data) == 0 {
		t.Fatalf("expected at least one binding on port %d", port)
	}

	b := result.Data[0]
	// Validate normalized PortBinding shape
	if b.Protocol != domain.TCP {
		t.Errorf("expected TCP, got %s", b.Protocol)
	}
	if b.LocalPort != uint16(port) {
		t.Errorf("expected port %d, got %d", port, b.LocalPort)
	}
	if b.LocalIP == nil {
		t.Error("LocalIP should not be nil")
	}
	if b.PID <= 0 {
		t.Error("expected positive PID from Windows adapter")
	}
	if b.State == domain.StateUnknown {
		t.Error("State should not be unknown")
	}
}

func TestWindowsEnumerator_ProtocolFilter(t *testing.T) {
	enum := NewEnumerator()

	// TCP only
	result, err := enum.List(context.Background(), &domain.Filter{
		Protocols: []domain.Protocol{domain.TCP},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range result.Data {
		if b.Protocol != domain.TCP {
			t.Errorf("expected only TCP bindings, got %s", b.Protocol)
		}
	}

	// UDP only
	result, err = enum.List(context.Background(), &domain.Filter{
		Protocols: []domain.Protocol{domain.UDP},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range result.Data {
		if b.Protocol != domain.UDP {
			t.Errorf("expected only UDP bindings, got %s", b.Protocol)
		}
	}
}
