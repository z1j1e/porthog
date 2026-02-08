package services_test

import (
	"context"
	"testing"

	"github.com/z1j1e/porthog/internal/core/domain"
	"github.com/z1j1e/porthog/internal/core/ports"
	"github.com/z1j1e/porthog/internal/core/services"
)

// --- Fake adapters ---

type fakeEnumerator struct {
	bindings []domain.PortBinding
	err      error
}

func (f *fakeEnumerator) List(_ context.Context, filter *domain.Filter) (*domain.PartialResult[[]domain.PortBinding], error) {
	if f.err != nil {
		return nil, f.err
	}
	data := f.bindings
	if filter != nil {
		var filtered []domain.PortBinding
		for i := range data {
			if filter.Matches(&data[i]) {
				filtered = append(filtered, data[i])
			}
		}
		data = filtered
	}
	return &domain.PartialResult[[]domain.PortBinding]{Data: data}, nil
}

type fakeResolver struct{}

func (f *fakeResolver) Enrich(_ context.Context, bindings []domain.PortBinding) ([]domain.PortBinding, error) {
	for i := range bindings {
		bindings[i].Process = &domain.ProcessIdentity{
			PID: bindings[i].PID, Name: "fake-process",
		}
	}
	return bindings, nil
}

func (f *fakeResolver) InvalidatePID(_ int32) {}

// PLACEHOLDER_TESTS_CONT

func sampleBindings() []domain.PortBinding {
	return []domain.PortBinding{
		{Protocol: domain.TCP, LocalPort: 8080, PID: 100, State: domain.StateListen},
		{Protocol: domain.TCP, LocalPort: 3000, PID: 200, State: domain.StateListen},
		{Protocol: domain.UDP, LocalPort: 5353, PID: 300, State: domain.StateListen},
	}
}

func TestListPorts_SortByPort(t *testing.T) {
	enum := &fakeEnumerator{bindings: sampleBindings()}
	svc := services.NewListPortsService(enum, &fakeResolver{})

	result, err := svc.List(context.Background(), nil, services.SortByPort)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Data) != 3 {
		t.Fatalf("expected 3 bindings, got %d", len(result.Data))
	}
	if result.Data[0].LocalPort != 3000 {
		t.Errorf("expected first port 3000, got %d", result.Data[0].LocalPort)
	}
	if result.Data[2].LocalPort != 8080 {
		t.Errorf("expected last port 8080, got %d", result.Data[2].LocalPort)
	}
}

func TestListPorts_FilterByProtocol(t *testing.T) {
	enum := &fakeEnumerator{bindings: sampleBindings()}
	svc := services.NewListPortsService(enum, &fakeResolver{})

	filter := &domain.Filter{Protocols: []domain.Protocol{domain.UDP}}
	result, err := svc.List(context.Background(), filter, services.SortByPort)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("expected 1 UDP binding, got %d", len(result.Data))
	}
	if result.Data[0].LocalPort != 5353 {
		t.Errorf("expected port 5353, got %d", result.Data[0].LocalPort)
	}
}

func TestListPorts_Enrichment(t *testing.T) {
	enum := &fakeEnumerator{bindings: sampleBindings()}
	svc := services.NewListPortsService(enum, &fakeResolver{})

	result, err := svc.List(context.Background(), nil, services.SortByPort)
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range result.Data {
		if b.Process == nil {
			t.Error("expected process to be enriched")
		}
		if b.Process.Name != "fake-process" {
			t.Errorf("expected fake-process, got %s", b.Process.Name)
		}
	}
}

// --- Kill service tests ---

type fakeTerminator struct {
	terminated []int32
}

func (f *fakeTerminator) Terminate(_ context.Context, pid int32, _ ports.SignalPolicy) error {
	f.terminated = append(f.terminated, pid)
	return nil
}

func TestKillByPort_Success(t *testing.T) {
	enum := &fakeEnumerator{bindings: []domain.PortBinding{
		{Protocol: domain.TCP, LocalPort: 8080, PID: 100, State: domain.StateListen},
	}}
	term := &fakeTerminator{}
	svc := services.NewKillByPortService(enum, &fakeResolver{}, term)

	result, err := svc.Kill(context.Background(), 8080, domain.TCP, ports.SignalPolicy{})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Killed {
		t.Error("expected process to be killed")
	}
	if len(term.terminated) != 1 || term.terminated[0] != 100 {
		t.Errorf("expected PID 100 terminated, got %v", term.terminated)
	}
}

func TestKillByPort_CriticalProcess(t *testing.T) {
	enum := &fakeEnumerator{bindings: []domain.PortBinding{
		{Protocol: domain.TCP, LocalPort: 445, PID: 1, State: domain.StateListen},
	}}
	svc := services.NewKillByPortService(enum, &fakeResolver{}, &fakeTerminator{})

	_, err := svc.Kill(context.Background(), 445, domain.TCP, ports.SignalPolicy{})
	if err == nil {
		t.Error("expected error for critical process")
	}
}

func TestKillByPort_DryRun(t *testing.T) {
	enum := &fakeEnumerator{bindings: []domain.PortBinding{
		{Protocol: domain.TCP, LocalPort: 8080, PID: 100, State: domain.StateListen},
	}}
	term := &fakeTerminator{}
	svc := services.NewKillByPortService(enum, &fakeResolver{}, term)

	result, err := svc.Kill(context.Background(), 8080, domain.TCP, ports.SignalPolicy{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if !result.DryRun {
		t.Error("expected dry run")
	}
	if len(term.terminated) != 0 {
		t.Error("expected no termination in dry run")
	}
}

func TestKillByPort_NotFound(t *testing.T) {
	enum := &fakeEnumerator{bindings: nil}
	svc := services.NewKillByPortService(enum, &fakeResolver{}, &fakeTerminator{})

	_, err := svc.Kill(context.Background(), 9999, domain.TCP, ports.SignalPolicy{})
	if err == nil {
		t.Error("expected not found error")
	}
}
