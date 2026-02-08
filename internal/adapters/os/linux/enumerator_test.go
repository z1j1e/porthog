//go:build linux

package linux

import (
	"context"
	"net"
	"strconv"
	"testing"

	"github.com/z1j1e/porthog/internal/core/domain"
)

func TestLinuxEnumerator_ListReturnsValidBindings(t *testing.T) {
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
		t.Errorf("expected positive PID, got %d (inode-to-PID mapping may have failed)", b.PID)
	}
	if b.State == domain.StateUnknown {
		t.Error("State should not be unknown")
	}
}

func TestLinuxEnumerator_NetlinkAndProcFallback(t *testing.T) {
	enum := NewEnumerator()
	// List all TCP — should succeed via netlink or /proc/net/tcp fallback
	result, err := enum.List(context.Background(), &domain.Filter{
		Protocols: []domain.Protocol{domain.TCP},
	})
	if err != nil {
		t.Fatal(err)
	}
	// On any Linux system there should be at least one TCP socket
	if len(result.Data) == 0 {
		t.Log("warning: no TCP bindings found, this may be expected in minimal containers")
	}
	for _, b := range result.Data {
		if b.Protocol != domain.TCP {
			t.Errorf("expected TCP, got %s", b.Protocol)
		}
	}
}
