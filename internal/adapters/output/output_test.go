package output_test

import (
	"bytes"
	"encoding/json"
	"net"
	"testing"

	"github.com/z1j1e/porthog/internal/adapters/output"
	"github.com/z1j1e/porthog/internal/core/domain"
)

func TestJSONOutput_Schema(t *testing.T) {
	bindings := []domain.PortBinding{
		{
			Protocol: domain.TCP, LocalIP: net.IPv4(127, 0, 0, 1), LocalPort: 8080,
			RemoteIP: net.IPv4zero, State: domain.StateListen, PID: 1234,
			Process: &domain.ProcessIdentity{PID: 1234, Name: "myapp", Username: "user"},
		},
	}
	result := &domain.PartialResult[[]domain.PortBinding]{Data: bindings}

	var buf bytes.Buffer
	r := output.NewRenderer(&buf, output.FormatJSON)
	if err := r.Render(result, "list"); err != nil {
		t.Fatal(err)
	}

	var envelope map[string]any
	if err := json.Unmarshal(buf.Bytes(), &envelope); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if v, ok := envelope["schema_version"]; !ok || v != "1.0" {
		t.Errorf("expected schema_version 1.0, got %v", v)
	}
	if v, ok := envelope["command"]; !ok || v != "list" {
		t.Errorf("expected command list, got %v", v)
	}
	if _, ok := envelope["timestamp"]; !ok {
		t.Error("missing timestamp field")
	}

	data, ok := envelope["data"].([]any)
	if !ok || len(data) != 1 {
		t.Fatalf("expected 1 data entry, got %v", envelope["data"])
	}

	entry := data[0].(map[string]any)
	if entry["protocol"] != "tcp" {
		t.Errorf("expected tcp, got %v", entry["protocol"])
	}
	if entry["local_port"].(float64) != 8080 {
		t.Errorf("expected port 8080, got %v", entry["local_port"])
	}
	if entry["pid"].(float64) != 1234 {
		t.Errorf("expected pid 1234, got %v", entry["pid"])
	}

	proc := entry["process"].(map[string]any)
	if proc["name"] != "myapp" {
		t.Errorf("expected process name myapp, got %v", proc["name"])
	}
}

func TestPlainOutput(t *testing.T) {
	bindings := []domain.PortBinding{
		{
			Protocol: domain.TCP, LocalIP: net.IPv4(0, 0, 0, 0), LocalPort: 3000,
			State: domain.StateListen, PID: 42,
			Process: &domain.ProcessIdentity{PID: 42, Name: "node"},
		},
	}
	result := &domain.PartialResult[[]domain.PortBinding]{Data: bindings}

	var buf bytes.Buffer
	r := output.NewRenderer(&buf, output.FormatPlain)
	if err := r.Render(result, "list"); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !bytes.Contains([]byte(out), []byte("3000")) {
		t.Errorf("expected port 3000 in output: %s", out)
	}
	if !bytes.Contains([]byte(out), []byte("node")) {
		t.Errorf("expected process name in output: %s", out)
	}
}
