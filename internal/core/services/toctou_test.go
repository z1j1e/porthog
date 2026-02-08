package services_test

import (
	"context"
	"testing"

	"github.com/porthog/porthog/internal/core/domain"
	"github.com/porthog/porthog/internal/core/ports"
	"github.com/porthog/porthog/internal/core/services"
)

// mutatingEnumerator changes PID between first and second List call
// to simulate a TOCTOU race condition.
type mutatingEnumerator struct {
	calls    int
	firstPID int32
	secondPID int32
	port     uint16
}

func (m *mutatingEnumerator) List(_ context.Context, _ *domain.Filter) (*domain.PartialResult[[]domain.PortBinding], error) {
	m.calls++
	pid := m.firstPID
	if m.calls > 1 {
		pid = m.secondPID
	}
	return &domain.PartialResult[[]domain.PortBinding]{
		Data: []domain.PortBinding{
			{Protocol: domain.TCP, LocalPort: m.port, PID: pid, State: domain.StateListen},
		},
	}, nil
}

type identityResolver struct{}

func (r *identityResolver) Enrich(_ context.Context, bindings []domain.PortBinding) ([]domain.PortBinding, error) {
	for i := range bindings {
		bindings[i].Process = &domain.ProcessIdentity{
			PID: bindings[i].PID, Name: "test", CreateTimeMs: int64(bindings[i].PID) * 1000,
		}
	}
	return bindings, nil
}

func (r *identityResolver) InvalidatePID(_ int32) {}

type noopTerminator struct{}

func (t *noopTerminator) Terminate(_ context.Context, _ int32, _ ports.SignalPolicy) error {
	return nil
}

func TestKill_TOCTOU_PIDChanged(t *testing.T) {
	enum := &mutatingEnumerator{firstPID: 100, secondPID: 200, port: 8080}
	svc := services.NewKillByPortService(enum, &identityResolver{}, &noopTerminator{})

	_, err := svc.Kill(context.Background(), 8080, domain.TCP, ports.SignalPolicy{})
	if err == nil {
		t.Fatal("expected ownership conflict error when PID changes between phases")
	}
	t.Logf("correctly detected TOCTOU: %v", err)
}

// mutatingIdentityResolver returns different create_time on second call
// to simulate PID reuse.
type mutatingIdentityResolver struct {
	calls int
}

func (r *mutatingIdentityResolver) Enrich(_ context.Context, bindings []domain.PortBinding) ([]domain.PortBinding, error) {
	r.calls++
	ct := int64(1000)
	if r.calls > 1 {
		ct = int64(9999) // different create_time = PID reuse
	}
	for i := range bindings {
		bindings[i].Process = &domain.ProcessIdentity{
			PID: bindings[i].PID, Name: "test", CreateTimeMs: ct,
		}
	}
	return bindings, nil
}

func (r *mutatingIdentityResolver) InvalidatePID(_ int32) {}

func TestKill_TOCTOU_CreateTimeMismatch(t *testing.T) {
	enum := &fakeEnumerator{bindings: []domain.PortBinding{
		{Protocol: domain.TCP, LocalPort: 8080, PID: 100, State: domain.StateListen},
	}}
	svc := services.NewKillByPortService(enum, &mutatingIdentityResolver{}, &noopTerminator{})

	_, err := svc.Kill(context.Background(), 8080, domain.TCP, ports.SignalPolicy{})
	if err == nil {
		t.Fatal("expected ownership conflict error when create_time changes (PID reuse)")
	}
	t.Logf("correctly detected PID reuse: %v", err)
}

func TestKill_TOCTOU_ProcessExited(t *testing.T) {
	// First call returns binding, second call returns empty (process exited)
	enum := &disappearingEnumerator{port: 8080}
	svc := services.NewKillByPortService(enum, &identityResolver{}, &noopTerminator{})

	_, err := svc.Kill(context.Background(), 8080, domain.TCP, ports.SignalPolicy{})
	if err == nil {
		t.Fatal("expected process exited error")
	}
	t.Logf("correctly detected exit: %v", err)
}

type disappearingEnumerator struct {
	calls int
	port  uint16
}

func (d *disappearingEnumerator) List(_ context.Context, _ *domain.Filter) (*domain.PartialResult[[]domain.PortBinding], error) {
	d.calls++
	if d.calls > 1 {
		return &domain.PartialResult[[]domain.PortBinding]{Data: nil}, nil
	}
	return &domain.PartialResult[[]domain.PortBinding]{
		Data: []domain.PortBinding{
			{Protocol: domain.TCP, LocalPort: d.port, PID: 100, State: domain.StateListen},
		},
	}, nil
}
