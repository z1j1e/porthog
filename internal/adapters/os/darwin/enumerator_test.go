//go:build darwin

package darwin

import (
	"context"
	"net"
	"strconv"
	"testing"

	"github.com/z1j1e/porthog/internal/core/domain"
)

func TestDarwinEnumerator_ListReturnsValidBindings(t *testing.T) {
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
		t.Error("expected positive PID from lsof-based adapter")
	}
	if b.State == domain.StateUnknown {
		t.Error("State should not be unknown")
	}
}

func TestDarwinEnumerator_ProtocolFilter(t *testing.T) {
	enum := NewEnumerator()
	result, err := enum.List(context.Background(), &domain.Filter{
		Protocols: []domain.Protocol{domain.TCP},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range result.Data {
		if b.Protocol != domain.TCP {
			t.Errorf("expected only TCP, got %s", b.Protocol)
		}
	}
}
