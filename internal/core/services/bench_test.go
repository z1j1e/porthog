package services_test

import (
	"context"
	"testing"

	"github.com/z1j1e/porthog/internal/adapters/platform"
	"github.com/z1j1e/porthog/internal/adapters/process"
	"github.com/z1j1e/porthog/internal/core/services"
)

func BenchmarkListPorts(b *testing.B) {
	enum := platform.NewEnumerator()
	resolver := process.NewResolver()
	svc := services.NewListPortsService(enum, resolver)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := svc.List(context.Background(), nil, services.SortByPort)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkListPortsNoEnrich(b *testing.B) {
	enum := platform.NewEnumerator()
	noopResolver := &fakeResolver{}
	svc := services.NewListPortsService(enum, noopResolver)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := svc.List(context.Background(), nil, services.SortByPort)
		if err != nil {
			b.Fatal(err)
		}
	}
}
